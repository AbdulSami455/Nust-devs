package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/abdulsami/nust-devs/internal/gamification"
	"github.com/abdulsami/nust-devs/internal/models"
	"github.com/abdulsami/nust-devs/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

const profileInsightMaxAge = 24 * time.Hour

const profileInsightSystemPrompt = `You are generating a practical profile insight card for a NUST developer.

Return ONLY a JSON object in this exact format (no markdown fences):
{
  "headline": "short summary line",
  "recent_activity_recap": "1 concise paragraph summarizing recent activity and profile shape",
  "top_achievements": ["achievement 1", "achievement 2", "achievement 3"],
  "completion_tips": ["tip 1", "tip 2", "tip 3"]
}

Rules:
- Use only the JSON snapshot below. Do not call tools.
- Keep achievements grounded in measurable activity, rank, stars, repos, streaks, or contributions.
- Completion tips should be actionable and tailored to the missing or weak profile signals.
- Output ONLY the JSON object, nothing else.`

type profileInsightJSON struct {
	Headline            string   `json:"headline"`
	RecentActivityRecap string   `json:"recent_activity_recap"`
	TopAchievements     []string `json:"top_achievements"`
	CompletionTips      []string `json:"completion_tips"`
}

type ProfileInsightsService struct {
	chat  *ChatService
	db    *pgxpool.Pool
	stats *repository.StatsRepo
	model string
}

func NewProfileInsightsService(chat *ChatService, db *pgxpool.Pool, stats *repository.StatsRepo, model string) *ProfileInsightsService {
	return &ProfileInsightsService{chat: chat, db: db, stats: stats, model: model}
}

func (s *ProfileInsightsService) Get(ctx context.Context, username string) (*models.ProfileInsights, error) {
	dev, err := s.stats.GetDeveloperByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	cached, err := s.load(ctx, dev.ID)
	if err == nil && cached != nil && time.Since(cached.GeneratedAt) < profileInsightMaxAge {
		slog.Info("profile insight cache hit", "username", username, "developer_id", dev.ID)
		return cached, nil
	}

	slog.Info("profile insight cache miss", "username", username, "developer_id", dev.ID)
	insight, err := s.generate(ctx, dev)
	if err != nil {
		slog.Warn("profile insight generation failed", "username", username, "developer_id", dev.ID, "err", err)
		if cached != nil {
			return cached, nil
		}
		return fallbackProfileInsights(*dev), nil
	}

	if err := s.save(ctx, insight); err != nil {
		slog.Warn("profile insight save failed", "username", username, "developer_id", dev.ID, "err", err)
	}
	return insight, nil
}

func (s *ProfileInsightsService) generate(ctx context.Context, dev *models.Developer) (*models.ProfileInsights, error) {
	rank, repos, contribs, contribStats, topRepos, topLanguage, languages, repoTips, missing := s.buildSnapshot(ctx, dev)

	totalContribs, activeDays, peakDay := contributionTotals(contribs)
	badgeTitle := gamification.PowerTitle(dev.PowerLevel)
	snapshot := map[string]any{
		"developer": map[string]any{
			"username":             dev.GithubUsername,
			"display_name":         dev.DisplayName,
			"bio":                  dev.Bio,
			"website":              dev.Website,
			"email":                dev.Email,
			"verification_status":  dev.VerificationStatus,
			"power_level":          dev.PowerLevel,
			"power_title":          badgeTitle,
			"xp":                   dev.XP,
			"activity_score":       dev.ActivityScore,
			"builder_score":        dev.BuilderScore,
			"contributor_score":    dev.ContributorScore,
			"reviewer_score":       dev.ReviewerScore,
			"community_score":      dev.CommunityScore,
			"current_streak":       dev.CurrentStreak,
			"longest_streak":       dev.LongestStreak,
			"streak_multiplier":    dev.StreakMultiplier,
			"public_repos":         dev.PublicRepos,
			"readme_repos":         dev.ReadmeRepos,
			"total_stars":          dev.TotalStars,
			"followers":            dev.Followers,
			"following":            dev.Following,
			"pr_contributions":     dev.PRContributions,
			"issue_contributions":  dev.IssueContributions,
			"review_contributions": dev.ReviewContributions,
			"rank":                 rank,
		},
		"repositories": map[string]any{
			"top_repos":    topRepos,
			"top_language": topLanguage,
			"languages":    languages,
			"repo_tips":    repoTips,
		},
		"contributions": map[string]any{
			"total":       totalContribs,
			"active_days": activeDays,
			"peak_day":    peakDay,
			"stats":       contribStats,
		},
		"missing_signals": missing,
	}
	snapshotJSON, _ := json.Marshal(snapshot)

	prompt := fmt.Sprintf(
		"Generate a developer profile insight card.\n\n%s\n\nSnapshot:\n%s",
		profileInsightSystemPrompt, string(snapshotJSON),
	)

	raw, err := s.chat.RunSync(ctx, prompt)
	if err != nil {
		return nil, err
	}
	parsed, err := parseProfileInsightJSON(raw)
	if err != nil {
		return nil, fmt.Errorf("parse profile insight json: %w", err)
	}

	topAchievements := uniqStrings(parsed.TopAchievements)
	if len(topAchievements) == 0 {
		topAchievements = fallbackAchievements(dev, rank, totalContribs, activeDays, peakDay, topRepos, topLanguage)
	}
	completionTips := uniqStrings(parsed.CompletionTips)
	if len(completionTips) == 0 {
		completionTips = missingProfileTips(dev, repos, missing)
	}

	return &models.ProfileInsights{
		DeveloperID:         dev.ID,
		Headline:            parsed.Headline,
		RecentActivityRecap: parsed.RecentActivityRecap,
		TopAchievements:     topAchievements,
		CompletionTips:      completionTips,
		ModelVersion:        s.model,
		GeneratedAt:         time.Now().UTC(),
	}, nil
}

func parseProfileInsightJSON(raw string) (*profileInsightJSON, error) {
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

	var out profileInsightJSON
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	if out.Headline == "" || out.RecentActivityRecap == "" {
		return nil, fmt.Errorf("incomplete profile insight JSON")
	}
	return &out, nil
}

func (s *ProfileInsightsService) buildSnapshot(ctx context.Context, dev *models.Developer) (rank *int, repos []models.PublicRepo, contribs []models.ContributionDay, contribStats *models.ContributionStats, topRepos []compareRepo, topLanguage string, languages []string, repoTips []string, missing []string) {
	leaderboard, err := s.stats.GetLeaderboardWithTrends(ctx, "activity_score", "", 0, 1, 1000)
	if err == nil {
		for _, entry := range leaderboard {
			if entry.GithubUsername == dev.GithubUsername {
				r := entry.Rank
				rank = &r
				break
			}
		}
	}

	repos, _ = s.stats.GetDeveloperRepos(ctx, dev.ID)
	contribs, _ = s.stats.GetContributions(ctx, dev.ID)
	contribStats, _ = s.stats.GetContributionStats(ctx, dev.ID)
	topRepos, topLanguage, languages = summarizeRepos(repos)
	repoTips = profileRepoTips(repos)
	missing = missingProfileSignals(dev, repos)
	return
}

func fallbackAchievements(dev *models.Developer, rank *int, totalContribs, activeDays, peakDay int, topRepos []compareRepo, topLanguage string) []string {
	out := []string{}
	if rank != nil {
		out = append(out, fmt.Sprintf("Ranked #%d on the leaderboard", *rank))
	}
	if dev.TotalStars > 0 {
		out = append(out, fmt.Sprintf("%d total stars across tracked repos", dev.TotalStars))
	}
	if totalContribs > 0 {
		out = append(out, fmt.Sprintf("%d contributions across %d active days", totalContribs, activeDays))
	}
	if dev.CurrentStreak > 0 {
		out = append(out, fmt.Sprintf("%d-day current streak", dev.CurrentStreak))
	}
	if topLanguage != "" {
		out = append(out, fmt.Sprintf("Most common language: %s", topLanguage))
	}
	if len(topRepos) > 0 {
		out = append(out, fmt.Sprintf("Featured repo: %s", topRepos[0].FullName))
	}
	if peakDay > 0 {
		out = append(out, fmt.Sprintf("Peak day with %d contributions", peakDay))
	}
	return uniqStrings(out)
}

func missingProfileSignals(dev *models.Developer, repos []models.PublicRepo) []string {
	out := []string{}
	if strings.TrimSpace(ptrString(dev.Bio)) == "" {
		out = append(out, "Add a short bio")
	}
	if strings.TrimSpace(ptrString(dev.Website)) == "" {
		out = append(out, "Add a portfolio or project website")
	}
	if strings.TrimSpace(ptrString(dev.Email)) == "" {
		out = append(out, "Add an email if you want admin contact")
	}
	if dev.ReadmeRepos == 0 {
		out = append(out, "Add READMEs to key repositories")
	}
	if dev.CurrentStreak == 0 {
		out = append(out, "Keep contributing regularly to build a streak")
	}
	activeOriginals := 0
	for _, repo := range repos {
		if !repo.IsFork {
			activeOriginals++
		}
	}
	if activeOriginals == 0 {
		out = append(out, "Publish or pin an original project")
	}
	return uniqStrings(out)
}

func profileRepoTips(repos []models.PublicRepo) []string {
	tips := []string{}
	for _, repo := range repos {
		if repo.IsFork {
			continue
		}
		if repo.Description == "" {
			tips = append(tips, fmt.Sprintf("Add a description to %s", repo.Name))
		}
		if repo.Language == nil || strings.TrimSpace(*repo.Language) == "" {
			tips = append(tips, fmt.Sprintf("Set a primary language for %s", repo.Name))
		}
	}
	return uniqStrings(tips)
}

func ptrString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func missingProfileTips(dev *models.Developer, repos []models.PublicRepo, missing []string) []string {
	if len(missing) > 0 {
		return missing
	}
	tips := []string{}
	if dev.ReadmeRepos > 0 {
		tips = append(tips, "Keep README coverage on your original repos")
	}
	if len(profileRepoTips(repos)) > 0 {
		tips = append(tips, profileRepoTips(repos)...)
	}
	if len(tips) == 0 {
		tips = append(tips, "Keep your profile updated as new projects land")
	}
	return uniqStrings(tips)
}

func (s *ProfileInsightsService) load(ctx context.Context, developerID string) (*models.ProfileInsights, error) {
	row := s.db.QueryRow(ctx, `
		SELECT developer_id, headline, recent_activity_recap, top_achievements, completion_tips, model_version, generated_at
		FROM ai_profile_insights
		WHERE developer_id = $1`, developerID)

	var out models.ProfileInsights
	if err := row.Scan(&out.DeveloperID, &out.Headline, &out.RecentActivityRecap, &out.TopAchievements, &out.CompletionTips, &out.ModelVersion, &out.GeneratedAt); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ProfileInsightsService) save(ctx context.Context, insight *models.ProfileInsights) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO ai_profile_insights (developer_id, headline, recent_activity_recap, top_achievements, completion_tips, model_version, generated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (developer_id) DO UPDATE SET
			headline = EXCLUDED.headline,
			recent_activity_recap = EXCLUDED.recent_activity_recap,
			top_achievements = EXCLUDED.top_achievements,
			completion_tips = EXCLUDED.completion_tips,
			model_version = EXCLUDED.model_version,
			generated_at = EXCLUDED.generated_at`,
		insight.DeveloperID, insight.Headline, insight.RecentActivityRecap, insight.TopAchievements, insight.CompletionTips, insight.ModelVersion, insight.GeneratedAt,
	)
	return err
}

func fallbackProfileInsights(dev models.Developer) *models.ProfileInsights {
	return &models.ProfileInsights{
		DeveloperID: dev.ID,
		Headline:    "Profile growth snapshot",
		RecentActivityRecap: fmt.Sprintf(
			"%s currently has %d stars across %d repos, with a %d-day streak and power level %d.",
			ptrString(dev.DisplayName), dev.TotalStars, dev.PublicRepos, dev.CurrentStreak, dev.PowerLevel,
		),
		TopAchievements: fallbackAchievements(&dev, nil, dev.PRContributions+dev.IssueContributions+dev.ReviewContributions, 0, 0, nil, ""),
		CompletionTips:  missingProfileSignals(&dev, nil),
		ModelVersion:    "fallback",
		GeneratedAt:     time.Now().UTC(),
	}
}
