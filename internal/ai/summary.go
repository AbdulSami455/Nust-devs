package ai

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DeveloperSummary is the structured output stored in ai_developer_summaries.
type DeveloperSummary struct {
	DeveloperID  string    `json:"developer_id"`
	Headline     string    `json:"headline"`
	Summary      string    `json:"summary"`
	Strengths    []string  `json:"strengths"`
	ModelVersion string    `json:"model_version"`
	GeneratedAt  time.Time `json:"generated_at"`
}

const summaryMaxAge = 24 * time.Hour

const summarySystemPrompt = `You are generating a professional developer summary card for a NUST developer profile page.

After gathering data with your tools, respond with ONLY a JSON object in this exact format (no markdown fences):
{
  "headline": "one short line describing the developer's focus (max 10 words, no name)",
  "summary": "2-3 sentences covering their activity, strengths, and notable stats (be specific with numbers)",
  "strengths": ["strength with number", "strength with number", "strength with number"]
}

Rules:
- Only reference numbers you actually retrieved from tools
- Do NOT include the developer's name in the headline
- Strengths must be specific: "120+ code reviews" not "good contributor"
- Output ONLY the JSON object, nothing else`

// SummaryService generates and persists developer summaries.
type SummaryService struct {
	chat  *ChatService
	db    *pgxpool.Pool
	model string
}

func NewSummaryService(chat *ChatService, db *pgxpool.Pool, model string) *SummaryService {
	return &SummaryService{chat: chat, db: db, model: model}
}

// Get returns a fresh or cached summary for the developer.
// Returns nil, nil if the developer doesn't exist or generation fails.
func (s *SummaryService) Get(ctx context.Context, developerID, username string) (*DeveloperSummary, error) {
	// Check cache in DB
	cached, err := s.load(ctx, developerID)
	if err == nil && cached != nil && time.Since(cached.GeneratedAt) < summaryMaxAge {
		return cached, nil
	}

	// Generate fresh
	summary, err := s.generate(ctx, developerID, username)
	if err != nil {
		slog.Warn("summary generation failed", "username", username, "err", err)
		// Return stale cache if available rather than nothing
		if cached != nil {
			return cached, nil
		}
		return nil, err
	}

	if err := s.save(ctx, summary); err != nil {
		slog.Warn("summary save failed", "username", username, "err", err)
	}

	return summary, nil
}

func (s *SummaryService) generate(ctx context.Context, developerID, username string) (*DeveloperSummary, error) {
	prompt := fmt.Sprintf(
		"Generate a summary card for NUST developer with GitHub username: %s. "+
			"Prefer get_developer_snapshot first, and only fall back to narrower tools if needed.\n\n%s",
		username, summarySystemPrompt,
	)

	raw, err := s.chat.RunSync(ctx, prompt)
	if err != nil {
		return nil, err
	}

	parsed, err := parseSummaryJSON(raw)
	if err != nil {
		return nil, fmt.Errorf("parse summary json: %w", err)
	}

	return &DeveloperSummary{
		DeveloperID:  developerID,
		Headline:     parsed.Headline,
		Summary:      parsed.Summary,
		Strengths:    parsed.Strengths,
		ModelVersion: s.model,
		GeneratedAt:  time.Now(),
	}, nil
}

type summaryJSON struct {
	Headline  string   `json:"headline"`
	Summary   string   `json:"summary"`
	Strengths []string `json:"strengths"`
}

func parseSummaryJSON(raw string) (*summaryJSON, error) {
	raw = strings.TrimSpace(raw)
	// Strip markdown fences if model wrapped anyway
	if strings.HasPrefix(raw, "```") {
		parts := strings.SplitN(raw, "```", 3)
		if len(parts) >= 2 {
			raw = parts[1]
			raw = strings.TrimPrefix(raw, "json")
			raw = strings.TrimSpace(raw)
		}
	}
	// Find JSON object boundaries
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no JSON object found in response")
	}
	raw = raw[start : end+1]

	var out summaryJSON
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	if out.Headline == "" || out.Summary == "" {
		return nil, fmt.Errorf("incomplete summary JSON")
	}
	if len(out.Strengths) == 0 {
		out.Strengths = []string{}
	}
	return &out, nil
}

func (s *SummaryService) load(ctx context.Context, developerID string) (*DeveloperSummary, error) {
	row := s.db.QueryRow(ctx, `
		SELECT developer_id, headline, summary, strengths, model_version, generated_at
		FROM ai_developer_summaries
		WHERE developer_id = $1`, developerID)

	var ds DeveloperSummary
	if err := row.Scan(&ds.DeveloperID, &ds.Headline, &ds.Summary, &ds.Strengths, &ds.ModelVersion, &ds.GeneratedAt); err != nil {
		return nil, err
	}
	return &ds, nil
}

func (s *SummaryService) save(ctx context.Context, ds *DeveloperSummary) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO ai_developer_summaries (developer_id, headline, summary, strengths, model_version, generated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (developer_id) DO UPDATE SET
			headline     = EXCLUDED.headline,
			summary      = EXCLUDED.summary,
			strengths    = EXCLUDED.strengths,
			model_version = EXCLUDED.model_version,
			generated_at = EXCLUDED.generated_at`,
		ds.DeveloperID, ds.Headline, ds.Summary, ds.Strengths, ds.ModelVersion, ds.GeneratedAt,
	)
	return err
}

// LogEval persists an eval log entry asynchronously (best-effort).
func LogEval(ctx context.Context, db *pgxpool.Pool, agentName, inputHash string, output map[string]any, latencyMS int, success bool) {
	outJSON, _ := json.Marshal(output)
	_, err := db.Exec(ctx, `
		INSERT INTO ai_eval_logs (agent_name, input_hash, output, latency_ms, success)
		VALUES ($1, $2, $3, $4, $5)`,
		agentName, inputHash, outJSON, latencyMS, success,
	)
	if err != nil {
		slog.Warn("eval log write failed", "err", err)
	}
}

// HashInput produces a short hash for dedup/tracking in eval logs.
func HashInput(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h[:8])
}
