package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/abdulsami/nust-devs/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProjectSummary is the structured output stored in ai_project_summaries.
type ProjectSummary struct {
	RepoID       string    `json:"repo_id"`
	Headline     string    `json:"headline"`
	Summary      string    `json:"summary"`
	ModelVersion string    `json:"model_version"`
	GeneratedAt  time.Time `json:"generated_at"`
}

const projectSummaryMaxAge = 24 * time.Hour

const projectSummarySystemPrompt = `You are generating a one-paragraph impact summary for a featured open-source project on the NUST Devs platform.

After reviewing the JSON snapshot below, respond with ONLY a JSON object in this exact format (no markdown fences):
{
  "headline": "one short line describing the project's impact (max 10 words)",
  "summary": "1 concise paragraph of 2-3 sentences explaining what the project does, why it matters, and any recent momentum"
}

Rules:
- Use only the data in the snapshot. Do not invent stats or capabilities.
- Keep the summary grounded in the repo name, description, stars, forks, language, and growth signals if present.
- Do not mention tool names or internal system details.
- Output ONLY the JSON object, nothing else.`

// ProjectSummaryService generates and persists summaries for featured repos.
type ProjectSummaryService struct {
	chat  *ChatService
	db    *pgxpool.Pool
	model string
}

func NewProjectSummaryService(chat *ChatService, db *pgxpool.Pool, model string) *ProjectSummaryService {
	return &ProjectSummaryService{chat: chat, db: db, model: model}
}

// Get returns a fresh or cached summary for the project.
func (s *ProjectSummaryService) Get(ctx context.Context, repo models.PublicRepo) (*ProjectSummary, error) {
	cached, err := s.load(ctx, repo.ID)
	if err == nil && cached != nil && time.Since(cached.GeneratedAt) < projectSummaryMaxAge {
		slog.Info("project summary cache hit", "repo", repo.FullName, "repo_id", repo.ID)
		return cached, nil
	}

	slog.Info("project summary cache miss", "repo", repo.FullName, "repo_id", repo.ID)
	summary, err := s.generate(ctx, repo)
	if err != nil {
		slog.Warn("project summary generation failed", "repo", repo.FullName, "repo_id", repo.ID, "err", err)
		if cached != nil {
			return cached, nil
		}
		return fallbackProjectSummary(repo), nil
	}

	if err := s.save(ctx, summary); err != nil {
		slog.Warn("project summary save failed", "repo", repo.FullName, "repo_id", repo.ID, "err", err)
	}

	return summary, nil
}

func (s *ProjectSummaryService) generate(ctx context.Context, repo models.PublicRepo) (*ProjectSummary, error) {
	slog.Info("project summary generation started", "repo", repo.FullName, "repo_id", repo.ID)
	snapshot := map[string]any{
		"id":               repo.ID,
		"name":             repo.Name,
		"full_name":        repo.FullName,
		"owner":            repo.Owner,
		"description":      repo.Description,
		"url":              repo.URL,
		"language":         repo.Language,
		"stars":            repo.Stars,
		"forks":            repo.Forks,
		"is_fork":          repo.IsFork,
		"pushed_at":        repo.PushedAt,
		"stars_growth_30d": repo.StarsGrowth30d,
		"forks_growth_30d": repo.ForksGrowth30d,
		"sparkline":        repo.Sparkline,
	}
	snapshotJSON, _ := json.Marshal(snapshot)

	prompt := fmt.Sprintf(
		"Generate a project impact summary for this featured repo.\n\n%s\n\nRepo snapshot:\n%s",
		projectSummarySystemPrompt, string(snapshotJSON),
	)

	raw, err := s.chat.RunSync(ctx, prompt)
	if err != nil {
		return nil, err
	}

	parsed, err := parseProjectSummaryJSON(raw)
	if err != nil {
		return nil, fmt.Errorf("parse project summary json: %w", err)
	}

	return &ProjectSummary{
		RepoID:       repo.ID,
		Headline:     parsed.Headline,
		Summary:      parsed.Summary,
		ModelVersion: s.model,
		GeneratedAt:  time.Now().UTC(),
	}, nil
}

type projectSummaryJSON struct {
	Headline string `json:"headline"`
	Summary  string `json:"summary"`
}

func parseProjectSummaryJSON(raw string) (*projectSummaryJSON, error) {
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

	var out projectSummaryJSON
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	if out.Headline == "" || out.Summary == "" {
		return nil, fmt.Errorf("incomplete project summary JSON")
	}
	return &out, nil
}

func (s *ProjectSummaryService) load(ctx context.Context, repoID string) (*ProjectSummary, error) {
	row := s.db.QueryRow(ctx, `
		SELECT repo_id, headline, summary, model_version, generated_at
		FROM ai_project_summaries
		WHERE repo_id = $1`, repoID)

	var ps ProjectSummary
	if err := row.Scan(&ps.RepoID, &ps.Headline, &ps.Summary, &ps.ModelVersion, &ps.GeneratedAt); err != nil {
		return nil, err
	}
	return &ps, nil
}

func (s *ProjectSummaryService) save(ctx context.Context, ps *ProjectSummary) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO ai_project_summaries (repo_id, headline, summary, model_version, generated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (repo_id) DO UPDATE SET
			headline = EXCLUDED.headline,
			summary = EXCLUDED.summary,
			model_version = EXCLUDED.model_version,
			generated_at = EXCLUDED.generated_at`,
		ps.RepoID, ps.Headline, ps.Summary, ps.ModelVersion, ps.GeneratedAt,
	)
	return err
}

func fallbackProjectSummary(repo models.PublicRepo) *ProjectSummary {
	headline := "Featured community project"
	if repo.IsFork {
		headline = "Notable community fork"
	}
	if repo.StarsGrowth30d != nil && *repo.StarsGrowth30d > 0 {
		headline = "Growing community project"
	}

	desc := strings.TrimSpace(repo.Description)
	if desc == "" {
		desc = fmt.Sprintf("%s is a tracked NUST project.", repo.FullName)
	}

	language := "mixed"
	if repo.Language != nil && strings.TrimSpace(*repo.Language) != "" {
		language = *repo.Language
	}

	parts := []string{
		desc,
		fmt.Sprintf("It is a %s project with %d stars and %d forks.", language, repo.Stars, repo.Forks),
	}
	if repo.StarsGrowth30d != nil {
		parts = append(parts, fmt.Sprintf("It gained %d stars in the last 30 days.", *repo.StarsGrowth30d))
	}
	if repo.PushedAt != nil {
		parts = append(parts, fmt.Sprintf("The repo was last updated %s.", *repo.PushedAt))
	}

	return &ProjectSummary{
		RepoID:       repo.ID,
		Headline:     headline,
		Summary:      strings.Join(parts, " "),
		ModelVersion: "fallback",
		GeneratedAt:  time.Now().UTC(),
	}
}
