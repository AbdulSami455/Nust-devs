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

const rankInsightMaxAge = 12 * time.Hour

const rankInsightSystemPrompt = `You are generating a short AI explanation for a NUST developer profile or leaderboard row.

Return ONLY a JSON object in this exact format (no markdown fences):
{
  "headline": "one short line explaining the current rank or badge state",
  "summary": "2-3 concise sentences explaining the current rank, recent movement, and why the badges make sense",
  "highlights": ["short highlight 1", "short highlight 2", "short highlight 3"]
}

Rules:
- Use only the JSON snapshot below. Do not call tools.
- Ground the explanation in the snapshot's rank, rank changes, power level, streak, verification, stars, repos, and contributions.
- Keep it factual. If no new movement exists, say the profile is stable.
- Output ONLY the JSON object, nothing else.`

type rankInsightJSON struct {
	Headline   string   `json:"headline"`
	Summary    string   `json:"summary"`
	Highlights []string `json:"highlights"`
}

// RankInsightService generates and persists explanation cards for rank/badge changes.
type RankInsightService struct {
	chat  *ChatService
	db    *pgxpool.Pool
	stats *repository.StatsRepo
	model string
}

func NewRankInsightService(chat *ChatService, db *pgxpool.Pool, stats *repository.StatsRepo, model string) *RankInsightService {
	return &RankInsightService{chat: chat, db: db, stats: stats, model: model}
}

func (s *RankInsightService) Get(ctx context.Context, username string) (*models.RankInsight, error) {
	dev, err := s.stats.GetDeveloperByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	cached, err := s.load(ctx, dev.ID)
	if err == nil && cached != nil && time.Since(cached.GeneratedAt) < rankInsightMaxAge {
		slog.Info("rank insight cache hit", "username", username, "developer_id", dev.ID)
		return cached, nil
	}

	slog.Info("rank insight cache miss", "username", username, "developer_id", dev.ID)
	summary, err := s.generate(ctx, dev)
	if err != nil {
		slog.Warn("rank insight generation failed", "username", username, "developer_id", dev.ID, "err", err)
		if cached != nil {
			return cached, nil
		}
		return fallbackRankInsight(*dev), nil
	}

	if err := s.save(ctx, summary); err != nil {
		slog.Warn("rank insight save failed", "username", username, "developer_id", dev.ID, "err", err)
	}

	return summary, nil
}

func (s *RankInsightService) generate(ctx context.Context, dev *models.Developer) (*models.RankInsight, error) {
	ranks, err := s.stats.GetLeaderboardWithTrends(ctx, "activity_score", "", 0, 1, 1000)
	if err != nil {
		return nil, err
	}

	var rank *int
	var rankDelta7d *int
	var rankDelta30d *int
	var scoreDelta7d *float64
	var scoreDelta30d *float64
	for _, entry := range ranks {
		if entry.GithubUsername == dev.GithubUsername {
			r := entry.Rank
			rank = &r
			rankDelta7d = entry.RankDelta7d
			rankDelta30d = entry.RankDelta30d
			scoreDelta7d = entry.ScoreDelta7d
			scoreDelta30d = entry.ScoreDelta30d
			break
		}
	}

	repos, err := s.stats.GetDeveloperRepos(ctx, dev.ID)
	if err != nil {
		return nil, err
	}
	contributions, err := s.stats.GetContributions(ctx, dev.ID)
	if err != nil {
		return nil, err
	}
	contributionStats, err := s.stats.GetContributionStats(ctx, dev.ID)
	if err != nil {
		return nil, err
	}

	topRepos, topLanguage, languages := summarizeRepos(repos)
	totalContribs, activeDays, peakDay := contributionTotals(contributions)
	badgeContext := buildBadgeContext(dev, rank, rankDelta7d, rankDelta30d)

	snapshot := map[string]any{
		"developer": map[string]any{
			"username":             dev.GithubUsername,
			"display_name":         dev.DisplayName,
			"verification_status":  dev.VerificationStatus,
			"power_level":          dev.PowerLevel,
			"power_title":          gamification.PowerTitle(dev.PowerLevel),
			"current_streak":       dev.CurrentStreak,
			"longest_streak":       dev.LongestStreak,
			"streak_multiplier":    dev.StreakMultiplier,
			"xp":                   dev.XP,
			"activity_score":       dev.ActivityScore,
			"builder_score":        dev.BuilderScore,
			"contributor_score":    dev.ContributorScore,
			"reviewer_score":       dev.ReviewerScore,
			"community_score":      dev.CommunityScore,
			"followers":            dev.Followers,
			"following":            dev.Following,
			"public_repos":         dev.PublicRepos,
			"readme_repos":         dev.ReadmeRepos,
			"total_stars":          dev.TotalStars,
			"pr_contributions":     dev.PRContributions,
			"issue_contributions":  dev.IssueContributions,
			"review_contributions": dev.ReviewContributions,
		},
		"rank":                rank,
		"rank_delta_7d":       rankDelta7d,
		"rank_delta_30d":      rankDelta30d,
		"score_delta_7d":      scoreDelta7d,
		"score_delta_30d":     scoreDelta30d,
		"badge_context":       badgeContext,
		"languages":           languages,
		"top_language":        topLanguage,
		"top_repos":           topRepos,
		"contribution_stats":  contributionStats,
		"total_contributions": totalContribs,
		"active_days":         activeDays,
		"peak_contributions":  peakDay,
	}
	snapshotJSON, _ := json.Marshal(snapshot)

	prompt := fmt.Sprintf(
		"Generate a rank and badge explanation for this developer.\n\n%s\n\nSnapshot:\n%s",
		rankInsightSystemPrompt, string(snapshotJSON),
	)

	raw, err := s.chat.RunSync(ctx, prompt)
	if err != nil {
		return nil, err
	}

	parsed, err := parseRankInsightJSON(raw)
	if err != nil {
		return nil, fmt.Errorf("parse rank insight json: %w", err)
	}

	if len(parsed.Highlights) == 0 {
		parsed.Highlights = badgeContext
	}

	return &models.RankInsight{
		DeveloperID:  dev.ID,
		Headline:     parsed.Headline,
		Summary:      parsed.Summary,
		Highlights:   parsed.Highlights,
		ModelVersion: s.model,
		GeneratedAt:  time.Now().UTC(),
	}, nil
}

func parseRankInsightJSON(raw string) (*rankInsightJSON, error) {
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

	var out rankInsightJSON
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	if out.Headline == "" || out.Summary == "" {
		return nil, fmt.Errorf("incomplete rank insight JSON")
	}
	if out.Highlights == nil {
		out.Highlights = []string{}
	}
	return &out, nil
}

func buildBadgeContext(dev *models.Developer, rank, rankDelta7d, rankDelta30d *int) []string {
	out := []string{}
	if rank != nil {
		if *rank <= 10 {
			out = append(out, fmt.Sprintf("Top 10 rank badge: #%d", *rank))
		} else {
			out = append(out, fmt.Sprintf("Current rank: #%d", *rank))
		}
	}
	if dev.VerificationStatus == "email_verified" {
		out = append(out, "Verified developer badge")
	}
	out = append(out, fmt.Sprintf("Power level: Lv.%d %s", dev.PowerLevel, gamification.PowerTitle(dev.PowerLevel)))
	if dev.CurrentStreak > 0 {
		out = append(out, fmt.Sprintf("%d-day streak badge", dev.CurrentStreak))
	}
	if rankDelta7d != nil && *rankDelta7d != 0 {
		out = append(out, fmt.Sprintf("7-day rank change: %+d", *rankDelta7d))
	} else if rankDelta30d != nil && *rankDelta30d != 0 {
		out = append(out, fmt.Sprintf("30-day rank change: %+d", *rankDelta30d))
	}
	return out
}

func (s *RankInsightService) load(ctx context.Context, developerID string) (*models.RankInsight, error) {
	row := s.db.QueryRow(ctx, `
		SELECT developer_id, headline, summary, highlights, model_version, generated_at
		FROM ai_rank_insights
		WHERE developer_id = $1`, developerID)

	var out models.RankInsight
	if err := row.Scan(&out.DeveloperID, &out.Headline, &out.Summary, &out.Highlights, &out.ModelVersion, &out.GeneratedAt); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *RankInsightService) save(ctx context.Context, insight *models.RankInsight) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO ai_rank_insights (developer_id, headline, summary, highlights, model_version, generated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (developer_id) DO UPDATE SET
			headline = EXCLUDED.headline,
			summary = EXCLUDED.summary,
			highlights = EXCLUDED.highlights,
			model_version = EXCLUDED.model_version,
			generated_at = EXCLUDED.generated_at`,
		insight.DeveloperID, insight.Headline, insight.Summary, insight.Highlights, insight.ModelVersion, insight.GeneratedAt,
	)
	return err
}

func fallbackRankInsight(dev models.Developer) *models.RankInsight {
	levelTitle := gamification.PowerTitle(dev.PowerLevel)
	headline := fmt.Sprintf("Lv.%d %s with a stable profile", dev.PowerLevel, levelTitle)
	summary := fmt.Sprintf("%s currently sits at power level %d and has %d stars across %d public repos. The profile looks stable, with verification status %s and a current streak of %d days.",
		valueOrUsername(dev.DisplayName, dev.GithubUsername),
		dev.PowerLevel,
		dev.TotalStars,
		dev.PublicRepos,
		dev.VerificationStatus,
		dev.CurrentStreak,
	)
	highlights := []string{
		fmt.Sprintf("Lv.%d %s", dev.PowerLevel, levelTitle),
		fmt.Sprintf("%d stars across %d repos", dev.TotalStars, dev.PublicRepos),
	}
	if dev.VerificationStatus == "email_verified" {
		highlights = append(highlights, "Verified developer badge")
	}
	if dev.CurrentStreak > 0 {
		highlights = append(highlights, fmt.Sprintf("%d-day streak", dev.CurrentStreak))
	}
	return &models.RankInsight{
		DeveloperID:  dev.ID,
		Headline:     headline,
		Summary:      summary,
		Highlights:   highlights,
		ModelVersion: "fallback",
		GeneratedAt:  time.Now().UTC(),
	}
}

func valueOrUsername(displayName *string, username string) string {
	if displayName != nil && strings.TrimSpace(*displayName) != "" {
		return *displayName
	}
	return username
}
