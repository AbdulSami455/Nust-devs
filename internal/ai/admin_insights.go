package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/abdulsami/nust-devs/internal/models"
	"github.com/abdulsami/nust-devs/internal/repository"
)

const adminInsightsMaxAge = 12 * time.Hour
const syncSummaryMaxAge = 6 * time.Hour

const adminRequestSystemPrompt = `You are generating a concise admin review card for a pending NUST profile request.

Return ONLY a JSON object in this exact format (no markdown fences):
{
  "headline": "short title",
  "summary": "2 concise sentences summarizing the request and fit",
  "duplicate_warning": "one concise warning if the request looks similar to an existing profile, otherwise a short reassuring note"
}

Rules:
- Use only the snapshot below.
- Compare the request against existing profiles only from the data in the snapshot.
- Be factual and concise.
- Output ONLY the JSON object, nothing else.`

const syncSummarySystemPrompt = `You are generating a short admin summary of recent sync changes for NUST Devs.

Return ONLY a JSON object in this exact format (no markdown fences):
{
  "headline": "short title",
  "summary": "2 concise sentences describing recent sync activity",
  "highlights": ["highlight 1", "highlight 2", "highlight 3"]
}

Rules:
- Use only the snapshot below.
- Mention recent synced developers, activity, or repo movement.
- Keep it concise and grounded in the data.
- Output ONLY the JSON object, nothing else.`

type adminRequestJSON struct {
	Headline         string `json:"headline"`
	Summary          string `json:"summary"`
	DuplicateWarning string `json:"duplicate_warning"`
}

type syncSummaryJSON struct {
	Headline   string   `json:"headline"`
	Summary    string   `json:"summary"`
	Highlights []string `json:"highlights"`
}

type AdminInsightsService struct {
	chat     *ChatService
	requests *repository.RequestRepo
	devs     *repository.DeveloperRepo
	stats    *repository.StatsRepo
	model    string
}

func NewAdminInsightsService(chat *ChatService, requests *repository.RequestRepo, devs *repository.DeveloperRepo, stats *repository.StatsRepo, model string) *AdminInsightsService {
	return &AdminInsightsService{chat: chat, requests: requests, devs: devs, stats: stats, model: model}
}

func NewSyncSummaryService(chat *ChatService, stats *repository.StatsRepo, model string) *SyncSummaryService {
	return &SyncSummaryService{chat: chat, stats: stats, model: model}
}

type SyncSummaryService struct {
	chat  *ChatService
	stats *repository.StatsRepo
	model string
}

func (s *AdminInsightsService) GetRequest(ctx context.Context, requestID string) (*models.JoinRequestInsight, error) {
	req, err := s.requests.GetByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	existing, _ := s.devs.List(ctx)
	match, score := bestRequestMatch(req, existing)

	snapshot := map[string]any{
		"request":           req,
		"existing_profiles": topMatchedProfiles(req, existing),
		"best_match": map[string]any{
			"username": ptrString(matchUsername(match)),
			"score":    score,
		},
	}

	raw, err := s.chat.RunSync(ctx, fmt.Sprintf(
		"Generate an admin request review summary.\n\n%s\n\nSnapshot:\n%s",
		adminRequestSystemPrompt, mustJSON(snapshot),
	))
	if err != nil {
		slog.Warn("admin request insight generation failed", "request_id", requestID, "err", err)
		return fallbackRequestInsight(req, match, score), nil
	}
	parsed, err := parseAdminRequestJSON(raw)
	if err != nil {
		slog.Warn("admin request insight parse failed", "request_id", requestID, "err", err)
		return fallbackRequestInsight(req, match, score), nil
	}

	out := &models.JoinRequestInsight{
		RequestID:        req.ID,
		Headline:         parsed.Headline,
		Summary:          parsed.Summary,
		DuplicateWarning: parsed.DuplicateWarning,
		MatchedUsername:  matchUsername(match),
		ModelVersion:     s.model,
		GeneratedAt:      time.Now().UTC(),
	}
	if strings.TrimSpace(out.DuplicateWarning) == "" {
		out.DuplicateWarning = fallbackDuplicateWarning(match, score)
	}
	return out, nil
}

func (s *SyncSummaryService) Get(ctx context.Context) (*models.SyncSummary, error) {
	devs, err := s.stats.GetRecentSyncedDevelopers(ctx, 10)
	if err != nil {
		return nil, err
	}
	overview, _ := s.stats.GetOverview(ctx)
	recentActivity, _ := s.stats.GetRecentActivity(ctx, 8)
	projects, _ := s.stats.GetTopProjects(ctx, 5)

	snapshot := map[string]any{
		"developers":      devs,
		"overview":        overview,
		"recent_activity": recentActivity,
		"top_projects":    projects,
	}

	raw, err := s.chat.RunSync(ctx, fmt.Sprintf(
		"Generate a recent sync changes summary for admins.\n\n%s\n\nSnapshot:\n%s",
		syncSummarySystemPrompt, mustJSON(snapshot),
	))
	if err != nil {
		slog.Warn("sync summary generation failed", "err", err)
		return fallbackSyncSummary(devs, overview, recentActivity, projects), nil
	}
	parsed, err := parseSyncSummaryJSON(raw)
	if err != nil {
		slog.Warn("sync summary parse failed", "err", err)
		return fallbackSyncSummary(devs, overview, recentActivity, projects), nil
	}

	return &models.SyncSummary{
		Headline:     parsed.Headline,
		Summary:      parsed.Summary,
		Highlights:   uniqStrings(parsed.Highlights),
		ModelVersion: s.model,
		GeneratedAt:  time.Now().UTC(),
	}, nil
}

func parseAdminRequestJSON(raw string) (*adminRequestJSON, error) {
	raw = stripJSON(raw)
	var out adminRequestJSON
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	if out.Headline == "" || out.Summary == "" {
		return nil, fmt.Errorf("missing request fields")
	}
	return &out, nil
}

func parseSyncSummaryJSON(raw string) (*syncSummaryJSON, error) {
	raw = stripJSON(raw)
	var out syncSummaryJSON
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	if out.Headline == "" || out.Summary == "" {
		return nil, fmt.Errorf("missing sync summary fields")
	}
	if out.Highlights == nil {
		out.Highlights = []string{}
	}
	return &out, nil
}

func bestRequestMatch(req *models.DeveloperRequest, devs []models.Developer) (*models.Developer, float64) {
	if len(devs) == 0 {
		return nil, 0
	}
	reqTokens := requestTokens(req)
	var best *models.Developer
	bestScore := 0.0
	for i := range devs {
		score := tokenSimilarity(reqTokens, developerTokens(&devs[i]))
		if score > bestScore {
			bestScore = score
			best = &devs[i]
		}
	}
	return best, bestScore
}

func topMatchedProfiles(req *models.DeveloperRequest, devs []models.Developer) []map[string]any {
	reqTokens := requestTokens(req)
	type scored struct {
		dev   models.Developer
		score float64
	}
	items := make([]scored, 0, len(devs))
	for _, dev := range devs {
		items = append(items, scored{dev: dev, score: tokenSimilarity(reqTokens, developerTokens(&dev))})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].score != items[j].score {
			return items[i].score > items[j].score
		}
		return items[i].dev.GithubUsername < items[j].dev.GithubUsername
	})
	out := make([]map[string]any, 0, minInt(3, len(items)))
	for i := 0; i < len(items) && i < 3; i++ {
		out = append(out, map[string]any{
			"username":     items[i].dev.GithubUsername,
			"display_name": items[i].dev.DisplayName,
			"score":        items[i].score,
		})
	}
	return out
}

func requestTokens(req *models.DeveloperRequest) []string {
	return normalizeTokens(
		req.GithubUsername,
		ptrString(req.DisplayName),
		ptrString(req.Email),
		ptrString(req.Batch),
		ptrString(req.Course),
		ptrString(req.Message),
	)
}

func developerTokens(dev *models.Developer) []string {
	return normalizeTokens(
		dev.GithubUsername,
		ptrString(dev.DisplayName),
		ptrString(dev.Email),
		ptrString(dev.Location),
		ptrString(dev.Company),
		ptrString(dev.Bio),
		ptrString(dev.Website),
	)
}

func normalizeTokens(values ...string) []string {
	out := make([]string, 0, len(values)*2)
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		fields := strings.FieldsFunc(value, func(r rune) bool {
			return r == ' ' || r == '-' || r == '_' || r == '.' || r == '@' || r == '/' || r == ','
		})
		out = append(out, fields...)
		out = append(out, value)
	}
	return uniqStrings(out)
}

func tokenSimilarity(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	set := map[string]struct{}{}
	for _, item := range b {
		set[item] = struct{}{}
	}
	matches := 0
	for _, item := range a {
		if _, ok := set[item]; ok {
			matches++
		}
	}
	denom := math.Max(float64(len(a)), float64(len(b)))
	return float64(matches) / denom
}

func fallbackRequestInsight(req *models.DeveloperRequest, match *models.Developer, score float64) *models.JoinRequestInsight {
	return &models.JoinRequestInsight{
		RequestID: req.ID,
		Headline:  "Pending profile review",
		Summary: fmt.Sprintf(
			"@%s requested to join the platform%s.",
			req.GithubUsername, requestSummaryTail(req),
		),
		DuplicateWarning: fallbackDuplicateWarning(match, score),
		MatchedUsername:  matchUsername(match),
		ModelVersion:     "fallback",
		GeneratedAt:      time.Now().UTC(),
	}
}

func requestSummaryTail(req *models.DeveloperRequest) string {
	parts := []string{}
	if req.DisplayName != nil && strings.TrimSpace(*req.DisplayName) != "" {
		parts = append(parts, "display name "+*req.DisplayName)
	}
	if req.Course != nil && strings.TrimSpace(*req.Course) != "" {
		parts = append(parts, "course "+*req.Course)
	}
	if req.Batch != nil && strings.TrimSpace(*req.Batch) != "" {
		parts = append(parts, "batch "+*req.Batch)
	}
	if len(parts) == 0 {
		return ""
	}
	return " with " + strings.Join(parts, ", ")
}

func fallbackDuplicateWarning(match *models.Developer, score float64) string {
	if match == nil {
		return "No close duplicate profile found in the current directory."
	}
	if score >= 0.8 {
		return fmt.Sprintf("Possible duplicate of @%s with a high similarity score.", match.GithubUsername)
	}
	return fmt.Sprintf("Closest existing profile is @%s; review for overlap.", match.GithubUsername)
}

func matchUsername(dev *models.Developer) *string {
	if dev == nil {
		return nil
	}
	username := dev.GithubUsername
	return &username
}

func fallbackSyncSummary(devs []models.Developer, overview *models.Overview, recent []models.ActivityEvent, projects []models.PublicRepo) *models.SyncSummary {
	highlights := []string{}
	for i, dev := range devs {
		if i >= 3 {
			break
		}
		if dev.LastSyncedAt != nil {
			highlights = append(highlights, fmt.Sprintf("@%s synced at %s", dev.GithubUsername, dev.LastSyncedAt.Format(time.RFC822)))
		}
	}
	if len(recent) > 0 {
		highlights = append(highlights, recent[0].Message)
	}
	if len(projects) > 0 {
		highlights = append(highlights, fmt.Sprintf("Top repo currently: %s", projects[0].FullName))
	}
	return &models.SyncSummary{
		Headline:     "Recent sync changes",
		Summary:      fmt.Sprintf("%d developers and %d repositories are currently tracked; latest syncs are visible in the admin feed.", overview.TotalDevelopers, overview.TotalRepos),
		Highlights:   uniqStrings(highlights),
		ModelVersion: "fallback",
		GeneratedAt:  time.Now().UTC(),
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
