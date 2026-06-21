package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/abdulsami/nust-devs/internal/models"
	"github.com/abdulsami/nust-devs/internal/repository"
)

const shareTextMaxAge = 24 * time.Hour

const shareTextSystemPrompt = `You are generating a short shareable social post for a NUST profile or project.

Return ONLY a JSON object in this exact format (no markdown fences):
{
  "headline": "short title",
  "share_text": "1-2 sentences that can be copied into a post"
}

Rules:
- Use only the snapshot below.
- Keep it natural, factual, and concise.
- Avoid hashtags unless they are clearly supported.
- Output ONLY the JSON object, nothing else.`

type shareTextJSON struct {
	Headline  string `json:"headline"`
	ShareText string `json:"share_text"`
}

type ShareTextService struct {
	chat  *ChatService
	stats *repository.StatsRepo
	model string
}

func NewShareTextService(chat *ChatService, stats *repository.StatsRepo, model string) *ShareTextService {
	return &ShareTextService{chat: chat, stats: stats, model: model}
}

func (s *ShareTextService) GetDeveloper(ctx context.Context, username string) (*models.ShareTextInsight, error) {
	dev, err := s.stats.GetDeveloperByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	repos, _ := s.stats.GetDeveloperRepos(ctx, dev.ID)
	contribs, _ := s.stats.GetContributions(ctx, dev.ID)
	totalContribs, activeDays, _ := contributionTotals(contribs)
	topRepos, topLanguage, _ := summarizeRepos(repos)
	rank, _ := s.rankForDeveloper(ctx, dev.GithubUsername)

	snapshot := map[string]any{
		"developer": map[string]any{
			"username":       dev.GithubUsername,
			"display_name":   dev.DisplayName,
			"bio":            dev.Bio,
			"activity_score": dev.ActivityScore,
			"total_stars":    dev.TotalStars,
			"public_repos":   dev.PublicRepos,
			"current_streak": dev.CurrentStreak,
			"power_level":    dev.PowerLevel,
			"rank":           rank,
		},
		"repositories": map[string]any{
			"top_repos":    topRepos,
			"top_language": topLanguage,
		},
		"contributions": map[string]any{
			"total":       totalContribs,
			"active_days": activeDays,
		},
	}

	raw, err := s.chat.RunSync(ctx, fmt.Sprintf(
		"Generate a share text for this developer profile.\n\n%s\n\nSnapshot:\n%s",
		shareTextSystemPrompt, mustJSON(snapshot),
	))
	if err != nil {
		slog.Warn("share text generation failed", "entity_type", "developer", "username", username, "err", err)
		return fallbackDeveloperShareText(dev, rank, totalContribs, activeDays, topRepos, topLanguage), nil
	}
	parsed, err := parseShareTextJSON(raw)
	if err != nil {
		slog.Warn("share text parse failed", "entity_type", "developer", "username", username, "err", err)
		return fallbackDeveloperShareText(dev, rank, totalContribs, activeDays, topRepos, topLanguage), nil
	}
	return &models.ShareTextInsight{
		EntityType:   "developer",
		EntityID:     dev.ID,
		Headline:     parsed.Headline,
		ShareText:    parsed.ShareText,
		ModelVersion: s.model,
		GeneratedAt:  time.Now().UTC(),
	}, nil
}

func (s *ShareTextService) GetProject(ctx context.Context, repoID string) (*models.ShareTextInsight, error) {
	repo, err := s.stats.GetProjectByID(ctx, repoID)
	if err != nil {
		return nil, err
	}
	raw, err := s.chat.RunSync(ctx, fmt.Sprintf(
		"Generate a share text for this featured project.\n\n%s\n\nSnapshot:\n%s",
		shareTextSystemPrompt, mustJSON(map[string]any{
			"repo": repo,
		}),
	))
	if err != nil {
		slog.Warn("share text generation failed", "entity_type", "project", "repo", repo.FullName, "err", err)
		return fallbackProjectShareText(repo), nil
	}
	parsed, err := parseShareTextJSON(raw)
	if err != nil {
		slog.Warn("share text parse failed", "entity_type", "project", "repo", repo.FullName, "err", err)
		return fallbackProjectShareText(repo), nil
	}
	return &models.ShareTextInsight{
		EntityType:   "project",
		EntityID:     repo.ID,
		Headline:     parsed.Headline,
		ShareText:    parsed.ShareText,
		ModelVersion: s.model,
		GeneratedAt:  time.Now().UTC(),
	}, nil
}

func parseShareTextJSON(raw string) (*shareTextJSON, error) {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "```") {
		parts := strings.SplitN(raw, "```", 3)
		if len(parts) >= 2 {
			raw = parts[1]
			raw = strings.TrimPrefix(raw, "json")
			raw = strings.TrimSpace(raw)
		}
	}
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no JSON object found in response")
	}
	raw = raw[start : end+1]

	var out shareTextJSON
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	if out.Headline == "" || out.ShareText == "" {
		return nil, fmt.Errorf("incomplete share text JSON")
	}
	return &out, nil
}

func fallbackDeveloperShareText(dev *models.Developer, rank *int, totalContribs, activeDays int, topRepos []compareRepo, topLanguage string) *models.ShareTextInsight {
	headline := "Share this profile"
	if dev.DisplayName != nil && strings.TrimSpace(*dev.DisplayName) != "" {
		headline = fmt.Sprintf("Share %s", *dev.DisplayName)
	}
	parts := []string{
		fmt.Sprintf("%s is building on GitHub with %d stars across %d repos.", ptrString(dev.DisplayName), dev.TotalStars, dev.PublicRepos),
		fmt.Sprintf("The profile shows %d contributions across %d active days.", totalContribs, activeDays),
	}
	if rank != nil {
		parts = append(parts, fmt.Sprintf("Current rank: #%d.", *rank))
	}
	if topLanguage != "" {
		parts = append(parts, fmt.Sprintf("Top language: %s.", topLanguage))
	}
	if len(topRepos) > 0 {
		parts = append(parts, fmt.Sprintf("Featured repo: %s.", topRepos[0].FullName))
	}
	return &models.ShareTextInsight{
		EntityType:   "developer",
		EntityID:     dev.ID,
		Headline:     headline,
		ShareText:    strings.Join(parts, " "),
		ModelVersion: "fallback",
		GeneratedAt:  time.Now().UTC(),
	}
}

func fallbackProjectShareText(repo *models.PublicRepo) *models.ShareTextInsight {
	headline := "Share this project"
	if repo.FullName != "" {
		headline = fmt.Sprintf("Share %s", repo.FullName)
	}
	desc := strings.TrimSpace(repo.Description)
	if desc == "" {
		desc = fmt.Sprintf("%s is a tracked NUST project.", repo.FullName)
	}
	language := "mixed"
	if repo.Language != nil && strings.TrimSpace(*repo.Language) != "" {
		language = *repo.Language
	}
	text := fmt.Sprintf("%s It uses %s, has %d stars and %d forks, and is listed on NUST Devs.", desc, language, repo.Stars, repo.Forks)
	return &models.ShareTextInsight{
		EntityType:   "project",
		EntityID:     repo.ID,
		Headline:     headline,
		ShareText:    text,
		ModelVersion: "fallback",
		GeneratedAt:  time.Now().UTC(),
	}
}

func (s *ShareTextService) rankForDeveloper(ctx context.Context, username string) (*int, error) {
	leaderboard, err := s.stats.GetLeaderboardWithTrends(ctx, "activity_score", "", 0, 1, 1000)
	if err != nil {
		return nil, err
	}
	for _, entry := range leaderboard {
		if strings.EqualFold(entry.GithubUsername, username) {
			r := entry.Rank
			return &r, nil
		}
	}
	return nil, nil
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
