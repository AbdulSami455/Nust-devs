package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/abdulsami/nust-devs/internal/models"
	"github.com/abdulsami/nust-devs/internal/repository"
)

const platformInsightsMaxAge = 6 * time.Hour
const weeklyReportMaxAge = 24 * time.Hour

const platformInsightsSystemPrompt = `You are generating one-line insights for a NUST Devs platform page.

Return ONLY a JSON object in this exact format (no markdown fences):
{
  "headline": "short title",
  "project_insights": ["one line about top projects", "one line about another project", "one line about project momentum"],
  "community_trends": ["one line about community activity", "one line about language or trend", "one line about developer momentum"]
}

Rules:
- Use only the snapshot below.
- Keep each line concise and useful.
- Do not invent data not supported by the snapshot.
- Output ONLY the JSON object, nothing else.`

const weeklyReportSystemPrompt = `You are generating a weekly community report for NUST Devs.

Return ONLY a JSON object in this exact format (no markdown fences):
{
  "headline": "short title",
  "summary": "2 concise sentences about the week",
  "highlights": ["highlight 1", "highlight 2", "highlight 3"]
}

Rules:
- Use only the snapshot below.
- Mention platform-wide movement, top projects, and active developers.
- Keep it factual and concise.
- Output ONLY the JSON object, nothing else.`

type platformInsightsJSON struct {
	Headline        string   `json:"headline"`
	ProjectInsights []string `json:"project_insights"`
	CommunityTrends []string `json:"community_trends"`
}

type weeklyReportJSON struct {
	Headline   string   `json:"headline"`
	Summary    string   `json:"summary"`
	Highlights []string `json:"highlights"`
}

type PlatformInsightsService struct {
	chat  *ChatService
	stats *repository.StatsRepo
	model string
}

func NewPlatformInsightsService(chat *ChatService, stats *repository.StatsRepo, model string) *PlatformInsightsService {
	return &PlatformInsightsService{chat: chat, stats: stats, model: model}
}

type WeeklyCommunityReportService struct {
	chat  *ChatService
	stats *repository.StatsRepo
	model string
}

func NewWeeklyCommunityReportService(chat *ChatService, stats *repository.StatsRepo, model string) *WeeklyCommunityReportService {
	return &WeeklyCommunityReportService{chat: chat, stats: stats, model: model}
}

func (s *PlatformInsightsService) Get(ctx context.Context) (*models.PlatformInsights, error) {
	overview, err := s.stats.GetOverview(ctx)
	if err != nil {
		return nil, err
	}
	projects, _ := s.stats.GetTopProjects(ctx, 5)
	languages, _ := s.stats.GetLanguageStats(ctx)
	activity, _ := s.stats.GetCommunityActivity(ctx, 7)
	board, _ := s.stats.GetLeaderboardWithTrends(ctx, "activity_score", "", 0, 1, 5)

	snapshot := map[string]any{
		"overview":  overview,
		"projects":  projects,
		"languages": languages,
		"activity":  activity,
		"leaders":   board,
	}

	raw, err := s.chat.RunSync(ctx, fmt.Sprintf(
		"Generate one-line platform insights.\n\n%s\n\nSnapshot:\n%s",
		platformInsightsSystemPrompt, mustJSON(snapshot),
	))
	if err != nil {
		slog.Warn("platform insights generation failed", "err", err)
		return fallbackPlatformInsights(overview, projects, languages, activity, board), nil
	}
	parsed, err := parsePlatformInsightsJSON(raw)
	if err != nil {
		slog.Warn("platform insights parse failed", "err", err)
		return fallbackPlatformInsights(overview, projects, languages, activity, board), nil
	}
	return &models.PlatformInsights{
		Headline:        parsed.Headline,
		ProjectInsights: uniqStrings(parsed.ProjectInsights),
		CommunityTrends: uniqStrings(parsed.CommunityTrends),
		ModelVersion:    s.model,
		GeneratedAt:     time.Now().UTC(),
	}, nil
}

func (s *WeeklyCommunityReportService) Get(ctx context.Context) (*models.WeeklyCommunityReport, error) {
	overview, err := s.stats.GetOverview(ctx)
	if err != nil {
		return nil, err
	}
	activity, _ := s.stats.GetCommunityActivity(ctx, 30)
	board, _ := s.stats.GetLeaderboardWithTrends(ctx, "activity_score", "", 0, 1, 10)
	projects, _ := s.stats.GetTopProjects(ctx, 5)
	recent, _ := s.stats.GetRecentActivity(ctx, 8)
	recentSync, _ := s.stats.GetRecentSyncedDevelopers(ctx, 8)

	snapshot := map[string]any{
		"overview":    overview,
		"activity":    activity,
		"leaderboard": board,
		"projects":    projects,
		"recent":      recent,
		"recent_sync": recentSync,
	}

	raw, err := s.chat.RunSync(ctx, fmt.Sprintf(
		"Generate a weekly community report.\n\n%s\n\nSnapshot:\n%s",
		weeklyReportSystemPrompt, mustJSON(snapshot),
	))
	if err != nil {
		slog.Warn("weekly report generation failed", "err", err)
		return fallbackWeeklyReport(overview, activity, board, projects, recent, recentSync), nil
	}
	parsed, err := parseWeeklyReportJSON(raw)
	if err != nil {
		slog.Warn("weekly report parse failed", "err", err)
		return fallbackWeeklyReport(overview, activity, board, projects, recent, recentSync), nil
	}
	return &models.WeeklyCommunityReport{
		Headline:     parsed.Headline,
		Summary:      parsed.Summary,
		Highlights:   uniqStrings(parsed.Highlights),
		ModelVersion: s.model,
		GeneratedAt:  time.Now().UTC(),
	}, nil
}

func parsePlatformInsightsJSON(raw string) (*platformInsightsJSON, error) {
	raw = stripJSON(raw)
	var out platformInsightsJSON
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	if out.Headline == "" {
		return nil, fmt.Errorf("missing headline")
	}
	if out.ProjectInsights == nil {
		out.ProjectInsights = []string{}
	}
	if out.CommunityTrends == nil {
		out.CommunityTrends = []string{}
	}
	return &out, nil
}

func parseWeeklyReportJSON(raw string) (*weeklyReportJSON, error) {
	raw = stripJSON(raw)
	var out weeklyReportJSON
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	if out.Headline == "" || out.Summary == "" {
		return nil, fmt.Errorf("missing weekly report fields")
	}
	if out.Highlights == nil {
		out.Highlights = []string{}
	}
	return &out, nil
}

func stripJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "```") {
		parts := strings.SplitN(raw, "```", 3)
		if len(parts) >= 2 {
			raw = strings.TrimPrefix(parts[1], "json")
			raw = strings.TrimSpace(raw)
		}
	}
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		return raw[start : end+1]
	}
	return raw
}

func fallbackPlatformInsights(overview *models.Overview, projects []models.PublicRepo, languages []models.LanguageStat, activity []models.CommunityActivityDay, board []models.LeaderboardEntry) *models.PlatformInsights {
	projectLines := []string{}
	if len(projects) > 0 {
		projectLines = append(projectLines, fmt.Sprintf("%s leads with %d stars", projects[0].FullName, projects[0].Stars))
	}
	if len(projects) > 1 {
		projectLines = append(projectLines, fmt.Sprintf("%s keeps momentum with %d forks", projects[1].FullName, projects[1].Forks))
	}
	if len(projects) > 2 {
		projectLines = append(projectLines, fmt.Sprintf("%s rounds out the top projects list", projects[2].FullName))
	}
	communityLines := []string{
		fmt.Sprintf("%d developers and %d repositories are tracked", overview.TotalDevelopers, overview.TotalRepos),
		fmt.Sprintf("%d total contributions are reflected across the platform", overview.TotalContributions),
	}
	if len(activity) > 0 {
		communityLines = append(communityLines, fmt.Sprintf("Recent activity landed on %s", activity[len(activity)-1].Date))
	}
	if len(languages) > 0 {
		communityLines = append(communityLines, fmt.Sprintf("Top language by volume is %s", languages[0].Language))
	}
	if len(board) > 0 {
		communityLines = append(communityLines, fmt.Sprintf("Current leader is @%s", board[0].GithubUsername))
	}
	return &models.PlatformInsights{
		Headline:        "Platform snapshots",
		ProjectInsights: uniqStrings(projectLines),
		CommunityTrends: uniqStrings(communityLines),
		ModelVersion:    "fallback",
		GeneratedAt:     time.Now().UTC(),
	}
}

func fallbackWeeklyReport(overview *models.Overview, activity []models.CommunityActivityDay, board []models.LeaderboardEntry, projects []models.PublicRepo, recent []models.ActivityEvent, recentSync []models.Developer) *models.WeeklyCommunityReport {
	highlights := []string{
		fmt.Sprintf("%d developers tracked", overview.TotalDevelopers),
		fmt.Sprintf("%d repositories tracked", overview.TotalRepos),
		fmt.Sprintf("%d total contributions", overview.TotalContributions),
	}
	if len(board) > 0 {
		highlights = append(highlights, fmt.Sprintf("Top developer: @%s", board[0].GithubUsername))
	}
	if len(projects) > 0 {
		highlights = append(highlights, fmt.Sprintf("Top project: %s", projects[0].FullName))
	}
	if len(recent) > 0 {
		highlights = append(highlights, fmt.Sprintf("Recent activity: %s", recent[0].Message))
	}
	if len(recentSync) > 0 {
		highlights = append(highlights, fmt.Sprintf("Recently synced: @%s", recentSync[0].GithubUsername))
	}
	if len(activity) > 0 {
		highlights = append(highlights, fmt.Sprintf("Most active day: %s", activity[len(activity)-1].Date))
	}
	return &models.WeeklyCommunityReport{
		Headline:     "Weekly community report",
		Summary:      fmt.Sprintf("The platform tracks %d developers and %d repositories with %d contributions across the latest reporting window.", overview.TotalDevelopers, overview.TotalRepos, overview.TotalContributions),
		Highlights:   uniqStrings(highlights),
		ModelVersion: "fallback",
		GeneratedAt:  time.Now().UTC(),
	}
}

func sortSyncedDevelopers(devs []models.Developer) {
	sort.SliceStable(devs, func(i, j int) bool {
		if devs[i].LastSyncedAt == nil && devs[j].LastSyncedAt == nil {
			return devs[i].GithubUsername < devs[j].GithubUsername
		}
		if devs[i].LastSyncedAt == nil {
			return false
		}
		if devs[j].LastSyncedAt == nil {
			return true
		}
		return devs[i].LastSyncedAt.After(*devs[j].LastSyncedAt)
	})
}
