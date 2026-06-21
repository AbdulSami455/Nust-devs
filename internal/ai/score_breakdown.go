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

const scoreBreakdownMaxAge = 12 * time.Hour

const scoreBreakdownSystemPrompt = `You are generating a short explanation of a NUST developer's score breakdown.

Return ONLY a JSON object in this exact format (no markdown fences):
{
  "headline": "short title",
  "summary": "2 concise sentences explaining what drives the profile score mix",
  "breakdown": ["bullet 1", "bullet 2", "bullet 3"]
}

Rules:
- Use only the snapshot below.
- Keep the explanation specific and factual.
- Ground the breakdown in the score components, streak, repos, stars, and contribution totals.
- Output ONLY the JSON object, nothing else.`

type scoreBreakdownJSON struct {
	Headline  string   `json:"headline"`
	Summary   string   `json:"summary"`
	Breakdown []string `json:"breakdown"`
}

type ScoreBreakdownService struct {
	chat  *ChatService
	db    *pgxpool.Pool
	stats *repository.StatsRepo
	model string
}

func NewScoreBreakdownService(chat *ChatService, db *pgxpool.Pool, stats *repository.StatsRepo, model string) *ScoreBreakdownService {
	return &ScoreBreakdownService{chat: chat, db: db, stats: stats, model: model}
}

func (s *ScoreBreakdownService) Get(ctx context.Context, username string) (*models.ScoreBreakdownInsight, error) {
	dev, err := s.stats.GetDeveloperByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	rank, repos, contribs, contribStats, topRepos, topLanguage, languages := s.snapshot(ctx, dev)
	totalContribs, activeDays, peakDay := contributionTotals(contribs)

	snapshot := map[string]any{
		"developer": map[string]any{
			"username":             dev.GithubUsername,
			"display_name":         dev.DisplayName,
			"verification_status":  dev.VerificationStatus,
			"power_level":          dev.PowerLevel,
			"power_title":          gamification.PowerTitle(dev.PowerLevel),
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
		},
		"contributions": map[string]any{
			"total":       totalContribs,
			"active_days": activeDays,
			"peak_day":    peakDay,
			"stats":       contribStats,
		},
	}

	raw, err := s.chat.RunSync(ctx, fmt.Sprintf(
		"Generate a profile score breakdown for this developer.\n\n%s\n\nSnapshot:\n%s",
		scoreBreakdownSystemPrompt, mustJSON(snapshot),
	))
	if err != nil {
		slog.Warn("score breakdown generation failed", "username", username, "err", err)
		return fallbackScoreBreakdown(dev, rank, repos, contribs, topRepos, topLanguage), nil
	}

	parsed, err := parseScoreBreakdownJSON(raw)
	if err != nil {
		slog.Warn("score breakdown parse failed", "username", username, "err", err)
		return fallbackScoreBreakdown(dev, rank, repos, contribs, topRepos, topLanguage), nil
	}

	breakdown := uniqStrings(parsed.Breakdown)
	if len(breakdown) == 0 {
		breakdown = fallbackScoreBreakdown(dev, rank, repos, contribs, topRepos, topLanguage).Breakdown
	}

	return &models.ScoreBreakdownInsight{
		DeveloperID:  dev.ID,
		Headline:     parsed.Headline,
		Summary:      parsed.Summary,
		Breakdown:    breakdown,
		ModelVersion: s.model,
		GeneratedAt:  time.Now().UTC(),
	}, nil
}

func (s *ScoreBreakdownService) snapshot(ctx context.Context, dev *models.Developer) (rank *int, repos []models.PublicRepo, contribs []models.ContributionDay, contribStats *models.ContributionStats, topRepos []compareRepo, topLanguage string, languages []string) {
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
	return
}

func parseScoreBreakdownJSON(raw string) (*scoreBreakdownJSON, error) {
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

	var out scoreBreakdownJSON
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	if out.Headline == "" || out.Summary == "" {
		return nil, fmt.Errorf("incomplete score breakdown JSON")
	}
	if out.Breakdown == nil {
		out.Breakdown = []string{}
	}
	return &out, nil
}

func fallbackScoreBreakdown(dev *models.Developer, rank *int, repos []models.PublicRepo, contribs []models.ContributionDay, topRepos []compareRepo, topLanguage string) *models.ScoreBreakdownInsight {
	totalContribs, activeDays, peakDay := contributionTotals(contribs)
	bullets := []string{
		fmt.Sprintf("Builder score: %.0f from %d public repos and %d readme-backed repos", dev.BuilderScore, dev.PublicRepos, dev.ReadmeRepos),
		fmt.Sprintf("Contributor score: %.0f from %d PRs, %d issues, and %d reviews", dev.ContributorScore, dev.PRContributions, dev.IssueContributions, dev.ReviewContributions),
		fmt.Sprintf("Community score: %.0f with %d followers and %d following", dev.CommunityScore, dev.Followers, dev.Following),
	}
	if rank != nil {
		bullets = append(bullets, fmt.Sprintf("Leaderboard rank: #%d", *rank))
	}
	if dev.CurrentStreak > 0 {
		bullets = append(bullets, fmt.Sprintf("Current streak: %d days", dev.CurrentStreak))
	}
	if topLanguage != "" {
		bullets = append(bullets, fmt.Sprintf("Most common language: %s", topLanguage))
	}
	if len(topRepos) > 0 {
		bullets = append(bullets, fmt.Sprintf("Top repo: %s", topRepos[0].FullName))
	}
	return &models.ScoreBreakdownInsight{
		DeveloperID: dev.ID,
		Headline:    "What drives this profile score",
		Summary: fmt.Sprintf(
			"%s's score mix is shaped by builder, contributor, reviewer, and community activity. The tracked activity totals show %d contributions across %d active days, with a peak day of %d contributions.",
			ptrString(dev.DisplayName), totalContribs, activeDays, peakDay,
		),
		Breakdown:    uniqStrings(bullets),
		ModelVersion: "fallback",
		GeneratedAt:  time.Now().UTC(),
	}
}
