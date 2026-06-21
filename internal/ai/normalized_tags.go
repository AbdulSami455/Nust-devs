package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/abdulsami/nust-devs/internal/gamification"
	"github.com/abdulsami/nust-devs/internal/models"
	"github.com/abdulsami/nust-devs/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

const normalizedTagsMaxAge = 24 * time.Hour

const normalizedTagsSystemPrompt = `You are generating normalized tags for a NUST Devs profile or project.

Return ONLY a JSON object in this exact format (no markdown fences):
{
  "headline": "short plain-language heading",
  "summary": "1-2 concise sentences explaining the normalized skill and language tags",
  "languages": ["canonical language 1", "canonical language 2"],
  "skills": ["canonical skill 1", "canonical skill 2"],
  "tags": ["canonical tag 1", "canonical tag 2", "canonical tag 3"]
}

Rules:
- Normalize equivalent names to a single canonical label.
- Prefer common canonical names such as Go, TypeScript, JavaScript, Python, Rust, Java, C++, C, backend, frontend, api, devops, data, testing, open-source.
- Keep tags short and consistent. Do not invent skills that are not supported by the snapshot.
- Output ONLY the JSON object, nothing else.`

type normalizedTagsJSON struct {
	Headline  string   `json:"headline"`
	Summary   string   `json:"summary"`
	Languages []string `json:"languages"`
	Skills    []string `json:"skills"`
	Tags      []string `json:"tags"`
}

type NormalizedTagsService struct {
	chat  *ChatService
	db    *pgxpool.Pool
	stats *repository.StatsRepo
	model string
}

func NewNormalizedTagsService(chat *ChatService, db *pgxpool.Pool, stats *repository.StatsRepo, model string) *NormalizedTagsService {
	return &NormalizedTagsService{chat: chat, db: db, stats: stats, model: model}
}

func (s *NormalizedTagsService) GetDeveloper(ctx context.Context, username string) (*models.NormalizedTags, error) {
	dev, err := s.stats.GetDeveloperByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	cached, err := s.load(ctx, "developer", dev.ID)
	if err == nil && cached != nil && time.Since(cached.GeneratedAt) < normalizedTagsMaxAge {
		slog.Info("normalized tags cache hit", "entity_type", "developer", "username", username, "entity_id", dev.ID)
		return cached, nil
	}

	slog.Info("normalized tags cache miss", "entity_type", "developer", "username", username, "entity_id", dev.ID)
	summary, err := s.generateDeveloper(ctx, dev)
	if err != nil {
		slog.Warn("normalized tags generation failed", "entity_type", "developer", "username", username, "entity_id", dev.ID, "err", err)
		if cached != nil {
			return cached, nil
		}
		return fallbackDeveloperTags(*dev), nil
	}

	if err := s.save(ctx, summary); err != nil {
		slog.Warn("normalized tags save failed", "entity_type", "developer", "username", username, "entity_id", dev.ID, "err", err)
	}
	return summary, nil
}

func (s *NormalizedTagsService) GetProject(ctx context.Context, repoID string) (*models.NormalizedTags, error) {
	repo, err := s.stats.GetProjectByID(ctx, repoID)
	if err != nil {
		return nil, err
	}

	cached, err := s.load(ctx, "project", repo.ID)
	if err == nil && cached != nil && time.Since(cached.GeneratedAt) < normalizedTagsMaxAge {
		slog.Info("normalized tags cache hit", "entity_type", "project", "repo", repo.FullName, "entity_id", repo.ID)
		return cached, nil
	}

	slog.Info("normalized tags cache miss", "entity_type", "project", "repo", repo.FullName, "entity_id", repo.ID)
	summary, err := s.generateProject(ctx, repo)
	if err != nil {
		slog.Warn("normalized tags generation failed", "entity_type", "project", "repo", repo.FullName, "entity_id", repo.ID, "err", err)
		if cached != nil {
			return cached, nil
		}
		return fallbackProjectTags(*repo), nil
	}

	if err := s.save(ctx, summary); err != nil {
		slog.Warn("normalized tags save failed", "entity_type", "project", "repo", repo.FullName, "entity_id", repo.ID, "err", err)
	}
	return summary, nil
}

func (s *NormalizedTagsService) generateDeveloper(ctx context.Context, dev *models.Developer) (*models.NormalizedTags, error) {
	repos, err := s.stats.GetDeveloperRepos(ctx, dev.ID)
	if err != nil {
		return nil, err
	}
	contribs, err := s.stats.GetContributions(ctx, dev.ID)
	if err != nil {
		return nil, err
	}
	_, activeDays, peakDay := contributionTotals(contribs)
	topRepos, topLanguage, languages := summarizeRepos(repos)
	snapshot := map[string]any{
		"developer": map[string]any{
			"username":            dev.GithubUsername,
			"display_name":        dev.DisplayName,
			"bio":                 dev.Bio,
			"verification_status": dev.VerificationStatus,
			"power_level":         dev.PowerLevel,
			"power_title":         gamification.PowerTitle(dev.PowerLevel),
			"activity_score":      dev.ActivityScore,
			"builder_score":       dev.BuilderScore,
			"contributor_score":   dev.ContributorScore,
			"reviewer_score":      dev.ReviewerScore,
			"community_score":     dev.CommunityScore,
			"public_repos":        dev.PublicRepos,
			"total_stars":         dev.TotalStars,
			"current_streak":      dev.CurrentStreak,
			"longest_streak":      dev.LongestStreak,
			"streak_multiplier":   dev.StreakMultiplier,
			"xp":                  dev.XP,
		},
		"languages":    languages,
		"top_language": topLanguage,
		"top_repos":    topRepos,
		"active_days":  activeDays,
		"peak_day":     peakDay,
		"skills_hint":  developerSkillHints(*dev, repos),
	}
	return s.generate(ctx, "developer", dev.ID, snapshot)
}

func (s *NormalizedTagsService) generateProject(ctx context.Context, repo *models.PublicRepo) (*models.NormalizedTags, error) {
	snapshot := map[string]any{
		"repo": map[string]any{
			"id":               repo.ID,
			"name":             repo.Name,
			"full_name":        repo.FullName,
			"owner":            repo.Owner,
			"description":      repo.Description,
			"language":         repo.Language,
			"stars":            repo.Stars,
			"forks":            repo.Forks,
			"is_fork":          repo.IsFork,
			"stars_growth_30d": repo.StarsGrowth30d,
			"forks_growth_30d": repo.ForksGrowth30d,
		},
		"skills_hint": projectSkillHints(repo),
	}
	return s.generate(ctx, "project", repo.ID, snapshot)
}

func (s *NormalizedTagsService) generate(ctx context.Context, entityType, entityID string, snapshot map[string]any) (*models.NormalizedTags, error) {
	snapshotJSON, _ := json.Marshal(snapshot)
	prompt := fmt.Sprintf(
		"Generate normalized tags for this %s.\n\n%s\n\nSnapshot:\n%s",
		entityType, normalizedTagsSystemPrompt, string(snapshotJSON),
	)

	raw, err := s.chat.RunSync(ctx, prompt)
	if err != nil {
		return nil, err
	}

	parsed, err := parseNormalizedTagsJSON(raw)
	if err != nil {
		return nil, fmt.Errorf("parse normalized tags json: %w", err)
	}

	languages := canonicalizeList(parsed.Languages, "language")
	skills := canonicalizeList(parsed.Skills, "skill")
	tags := canonicalizeList(parsed.Tags, "tag")
	tags = uniqStrings(append(append([]string{}, languages...), append(skills, tags...)...))

	return &models.NormalizedTags{
		EntityType:   entityType,
		EntityID:     entityID,
		Headline:     parsed.Headline,
		Summary:      parsed.Summary,
		Languages:    languages,
		Skills:       skills,
		Tags:         tags,
		ModelVersion: s.model,
		GeneratedAt:  time.Now().UTC(),
	}, nil
}

func parseNormalizedTagsJSON(raw string) (*normalizedTagsJSON, error) {
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

	var out normalizedTagsJSON
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	if out.Headline == "" || out.Summary == "" {
		return nil, fmt.Errorf("incomplete normalized tags JSON")
	}
	return &out, nil
}

func canonicalizeList(items []string, kind string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if v := canonicalTag(item, kind); v != "" {
			out = append(out, v)
		}
	}
	return uniqStrings(out)
}

func canonicalTag(value, kind string) string {
	s := strings.TrimSpace(strings.ToLower(value))
	if s == "" {
		return ""
	}
	if kind == "language" {
		switch s {
		case "go", "golang":
			return "Go"
		case "typescript", "ts":
			return "TypeScript"
		case "javascript", "js":
			return "JavaScript"
		case "python", "py":
			return "Python"
		case "rust":
			return "Rust"
		case "java":
			return "Java"
		case "c++", "cpp", "c plus plus":
			return "C++"
		case "c":
			return "C"
		case "c#", "csharp", "c sharp":
			return "C#"
		case "php":
			return "PHP"
		case "ruby":
			return "Ruby"
		case "dart":
			return "Dart"
		case "kotlin":
			return "Kotlin"
		case "swift":
			return "Swift"
		}
		return strings.Title(s)
	}

	switch s {
	case "backend", "back-end", "server-side":
		return "backend"
	case "frontend", "front-end", "ui":
		return "frontend"
	case "full stack", "full-stack":
		return "full-stack"
	case "api", "apis":
		return "api"
	case "devops", "dev-ops":
		return "devops"
	case "testing", "qa":
		return "testing"
	case "data", "data-engineering":
		return "data"
	case "ml", "machine learning", "machine-learning", "ai":
		return "ml"
	case "open source", "open-source", "oss":
		return "open-source"
	case "cli", "command line":
		return "cli"
	case "library", "lib":
		return "library"
	case "tooling", "tools":
		return "tooling"
	default:
		return s
	}
}

func uniqStrings(items []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return strings.ToLower(out[i]) < strings.ToLower(out[j])
	})
	return out
}

func developerSkillHints(dev models.Developer, repos []models.PublicRepo) []string {
	hints := []string{}
	if dev.PRContributions > 0 {
		hints = append(hints, "collaboration")
	}
	if dev.ReviewContributions > 0 {
		hints = append(hints, "reviewing")
	}
	if dev.PublicRepos > 0 {
		hints = append(hints, "open-source")
	}
	if dev.CurrentStreak > 0 {
		hints = append(hints, "consistent")
	}
	for _, repo := range repos {
		if repo.IsFork {
			continue
		}
		if repo.Language != nil && *repo.Language != "" {
			hints = append(hints, strings.ToLower(canonicalTag(*repo.Language, "language")))
		}
	}
	return uniqStrings(hints)
}

func projectSkillHints(repo *models.PublicRepo) []string {
	hints := []string{}
	name := strings.ToLower(repo.Name + " " + repo.FullName + " " + repo.Description)
	switch {
	case strings.Contains(name, "api"):
		hints = append(hints, "api")
	case strings.Contains(name, "cli"):
		hints = append(hints, "cli")
	case strings.Contains(name, "dashboard") || strings.Contains(name, "web"):
		hints = append(hints, "frontend")
	}
	if repo.IsFork {
		hints = append(hints, "open-source")
	}
	if repo.Language != nil && *repo.Language != "" {
		hints = append(hints, strings.ToLower(canonicalTag(*repo.Language, "language")))
	}
	return uniqStrings(hints)
}

func fallbackDeveloperTags(dev models.Developer) *models.NormalizedTags {
	languages := []string{}
	skills := []string{}
	if dev.PublicRepos > 0 {
		skills = append(skills, "open-source")
	}
	if dev.PRContributions > 0 {
		skills = append(skills, "collaboration")
	}
	if dev.ReviewContributions > 0 {
		skills = append(skills, "reviewing")
	}
	if dev.CurrentStreak > 0 {
		skills = append(skills, "consistent")
	}
	tags := uniqStrings(append(append([]string{}, languages...), skills...))
	headline := "Normalized developer stack"
	summary := fmt.Sprintf("%s has a normalized tag set built from their repositories and activity patterns.", dev.GithubUsername)
	return &models.NormalizedTags{
		EntityType:   "developer",
		EntityID:     dev.ID,
		Headline:     headline,
		Summary:      summary,
		Languages:    uniqStrings(languages),
		Skills:       uniqStrings(skills),
		Tags:         tags,
		ModelVersion: "fallback",
		GeneratedAt:  time.Now().UTC(),
	}
}

func fallbackProjectTags(repo models.PublicRepo) *models.NormalizedTags {
	languages := []string{}
	skills := projectSkillHints(&repo)
	tags := []string{}
	if repo.Language != nil && *repo.Language != "" {
		languages = append(languages, canonicalTag(*repo.Language, "language"))
	}
	if repo.IsFork {
		tags = append(tags, "open-source")
	}
	tags = append(tags, skills...)
	if len(languages) > 0 {
		tags = append(tags, languages...)
	}
	tags = uniqStrings(tags)
	headline := "Normalized project tags"
	summary := fmt.Sprintf("%s has normalized tags based on its language, description, and repo metadata.", repo.FullName)
	return &models.NormalizedTags{
		EntityType:   "project",
		EntityID:     repo.ID,
		Headline:     headline,
		Summary:      summary,
		Languages:    uniqStrings(languages),
		Skills:       uniqStrings(skills),
		Tags:         tags,
		ModelVersion: "fallback",
		GeneratedAt:  time.Now().UTC(),
	}
}

func (s *NormalizedTagsService) load(ctx context.Context, entityType, entityID string) (*models.NormalizedTags, error) {
	row := s.db.QueryRow(ctx, `
		SELECT entity_type, entity_id, headline, summary, languages, skills, tags, model_version, generated_at
		FROM ai_normalized_tags
		WHERE entity_type = $1 AND entity_id = $2`, entityType, entityID)

	var out models.NormalizedTags
	if err := row.Scan(&out.EntityType, &out.EntityID, &out.Headline, &out.Summary, &out.Languages, &out.Skills, &out.Tags, &out.ModelVersion, &out.GeneratedAt); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *NormalizedTagsService) save(ctx context.Context, tags *models.NormalizedTags) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO ai_normalized_tags (entity_type, entity_id, headline, summary, languages, skills, tags, model_version, generated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (entity_type, entity_id) DO UPDATE SET
			headline = EXCLUDED.headline,
			summary = EXCLUDED.summary,
			languages = EXCLUDED.languages,
			skills = EXCLUDED.skills,
			tags = EXCLUDED.tags,
			model_version = EXCLUDED.model_version,
			generated_at = EXCLUDED.generated_at`,
		tags.EntityType, tags.EntityID, tags.Headline, tags.Summary, tags.Languages, tags.Skills, tags.Tags, tags.ModelVersion, tags.GeneratedAt,
	)
	return err
}
