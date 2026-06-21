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

const homeBannerMaxAge = 2 * time.Hour

const homeBannerSystemPrompt = `You are generating a short "what changed today" banner for the NUST Devs homepage.

Return ONLY a JSON object in this exact format (no markdown fences):
{
  "headline": "short title",
  "summary": "2 concise sentences summarizing the latest movement on the platform",
  "highlights": ["highlight 1", "highlight 2", "highlight 3"]
}

Rules:
- Use only the snapshot below.
- Focus on fresh activity, new projects, leaderboard movement, and active developers.
- Keep it concise and specific.
- Output ONLY the JSON object, nothing else.`

type homeBannerJSON struct {
	Headline   string   `json:"headline"`
	Summary    string   `json:"summary"`
	Highlights []string `json:"highlights"`
}

type HomeBannerService struct {
	chat  *ChatService
	stats *repository.StatsRepo
	model string
}

func NewHomeBannerService(chat *ChatService, stats *repository.StatsRepo, model string) *HomeBannerService {
	return &HomeBannerService{chat: chat, stats: stats, model: model}
}

func (s *HomeBannerService) Get(ctx context.Context) (*models.HomeBannerInsight, error) {
	overview, err := s.stats.GetOverview(ctx)
	if err != nil {
		return nil, err
	}
	activity, _ := s.stats.GetRecentActivity(ctx, 6)
	topProjects, _ := s.stats.GetTopProjects(ctx, 3)
	spotlight, _ := s.stats.GetSpotlightDeveloper(ctx)

	snapshot := map[string]any{
		"overview": map[string]any{
			"total_developers":    overview.TotalDevelopers,
			"total_repos":         overview.TotalRepos,
			"total_stars":         overview.TotalStars,
			"total_contributions": overview.TotalContributions,
		},
		"recent_activity": activity,
		"top_projects":    topProjects,
		"spotlight":       spotlight,
	}

	raw, err := s.chat.RunSync(ctx, fmt.Sprintf(
		"Generate a homepage banner from this snapshot.\n\n%s\n\nSnapshot:\n%s",
		homeBannerSystemPrompt, mustJSON(snapshot),
	))
	if err != nil {
		slog.Warn("home banner generation failed", "err", err)
		return fallbackHomeBanner(overview, activity, topProjects, spotlight), nil
	}

	parsed, err := parseHomeBannerJSON(raw)
	if err != nil {
		slog.Warn("home banner parse failed", "err", err)
		return fallbackHomeBanner(overview, activity, topProjects, spotlight), nil
	}

	return &models.HomeBannerInsight{
		Headline:     parsed.Headline,
		Summary:      parsed.Summary,
		Highlights:   uniqStrings(parsed.Highlights),
		ModelVersion: s.model,
		GeneratedAt:  time.Now().UTC(),
	}, nil
}

func parseHomeBannerJSON(raw string) (*homeBannerJSON, error) {
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

	var out homeBannerJSON
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	if out.Headline == "" || out.Summary == "" {
		return nil, fmt.Errorf("incomplete home banner JSON")
	}
	if out.Highlights == nil {
		out.Highlights = []string{}
	}
	return &out, nil
}

func fallbackHomeBanner(overview *models.Overview, activity []models.ActivityEvent, topProjects []models.PublicRepo, spotlight *models.Developer) *models.HomeBannerInsight {
	highlights := []string{
		fmt.Sprintf("%d developers tracked", overview.TotalDevelopers),
		fmt.Sprintf("%d repos tracked", overview.TotalRepos),
		fmt.Sprintf("%d total stars", overview.TotalStars),
	}
	if len(activity) > 0 {
		highlights = append(highlights, fmt.Sprintf("Latest update: %s", activity[0].Message))
	}
	if len(topProjects) > 0 {
		highlights = append(highlights, fmt.Sprintf("Top project: %s", topProjects[0].FullName))
	}
	if spotlight != nil {
		highlights = append(highlights, fmt.Sprintf("Spotlight: %s", spotlight.GithubUsername))
	}
	return &models.HomeBannerInsight{
		Headline:     "What changed today",
		Summary:      fmt.Sprintf("The platform currently tracks %d developers, %d repositories, and %d total contributions.", overview.TotalDevelopers, overview.TotalRepos, overview.TotalContributions),
		Highlights:   uniqStrings(highlights),
		ModelVersion: "fallback",
		GeneratedAt:  time.Now().UTC(),
	}
}
