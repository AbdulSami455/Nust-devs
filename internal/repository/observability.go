package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/abdulsami/nust-devs/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ObservabilityRepo struct {
	db *pgxpool.Pool
}

func NewObservabilityRepo(db *pgxpool.Pool) *ObservabilityRepo {
	return &ObservabilityRepo{db: db}
}

type AuditLogInput struct {
	ActorType    string
	ActorID      string
	Action       string
	ResourceType string
	ResourceID   string
	Method       string
	Path         string
	StatusCode   int
	IP           string
	UserAgent    string
	Metadata     map[string]any
}

type AgentRunInput struct {
	SessionID   string
	AgentName   string
	UserMessage string
	InputHash   string
	IP          string
	UserAgent   string
}

type AgentRunFinishInput struct {
	Status        string
	LatencyMS     int
	ErrorMessage  string
	ResponseChars int
	ToolCalls     int
}

type AgentRunEventInput struct {
	RunID     string
	EventType string
	ToolName  string
	Payload   map[string]any
	LatencyMS int
	Success   bool
}

func (r *ObservabilityRepo) InsertAuditLog(ctx context.Context, in AuditLogInput) error {
	metaJSON, _ := json.Marshal(in.Metadata)
	_, err := r.db.Exec(ctx, `
		INSERT INTO audit_logs (
			actor_type, actor_id, action, resource_type, resource_id,
			method, path, status_code, ip, user_agent, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		in.ActorType, in.ActorID, in.Action, in.ResourceType, in.ResourceID,
		in.Method, in.Path, in.StatusCode, in.IP, in.UserAgent, metaJSON,
	)
	return err
}

func (r *ObservabilityRepo) StartAgentRun(ctx context.Context, in AgentRunInput) (string, error) {
	var id string
	err := r.db.QueryRow(ctx, `
		INSERT INTO agent_runs (session_id, agent_name, user_message, input_hash, ip, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		in.SessionID, in.AgentName, in.UserMessage, in.InputHash, in.IP, in.UserAgent,
	).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *ObservabilityRepo) FinishAgentRun(ctx context.Context, runID string, in AgentRunFinishInput) error {
	_, err := r.db.Exec(ctx, `
		UPDATE agent_runs
		SET status = $2,
		    latency_ms = $3,
		    error_message = $4,
		    response_chars = $5,
		    tool_calls = $6,
		    finished_at = NOW()
		WHERE id = $1`,
		runID, in.Status, in.LatencyMS, in.ErrorMessage, in.ResponseChars, in.ToolCalls,
	)
	return err
}

func (r *ObservabilityRepo) InsertAgentRunEvent(ctx context.Context, in AgentRunEventInput) error {
	payloadJSON, _ := json.Marshal(in.Payload)
	_, err := r.db.Exec(ctx, `
		INSERT INTO agent_run_events (run_id, event_type, tool_name, payload, latency_ms, success)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		in.RunID, in.EventType, in.ToolName, payloadJSON, in.LatencyMS, in.Success,
	)
	return err
}

func (r *ObservabilityRepo) ListAuditLogs(ctx context.Context, limit int) ([]models.AuditLog, error) {
	if limit < 1 || limit > 200 {
		limit = 50
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, actor_type, actor_id, action, resource_type, resource_id,
		       method, path, status_code, ip, user_agent, metadata, created_at
		FROM audit_logs
		ORDER BY created_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.AuditLog
	for rows.Next() {
		var item models.AuditLog
		var raw []byte
		if err := rows.Scan(
			&item.ID, &item.ActorType, &item.ActorID, &item.Action, &item.ResourceType, &item.ResourceID,
			&item.Method, &item.Path, &item.StatusCode, &item.IP, &item.UserAgent, &raw, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		item.Metadata = decodeJSONMap(raw)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *ObservabilityRepo) ListAgentRuns(ctx context.Context, limit int) ([]models.AgentRun, error) {
	if limit < 1 || limit > 100 {
		limit = 25
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, session_id, agent_name, user_message, input_hash, status, ip, user_agent,
		       tool_calls, latency_ms, error_message, response_chars, created_at, finished_at
		FROM agent_runs
		ORDER BY created_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.AgentRun
	for rows.Next() {
		var item models.AgentRun
		if err := rows.Scan(
			&item.ID, &item.SessionID, &item.AgentName, &item.UserMessage, &item.InputHash,
			&item.Status, &item.IP, &item.UserAgent, &item.ToolCalls, &item.LatencyMS,
			&item.ErrorMessage, &item.ResponseChars, &item.CreatedAt, &item.FinishedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *ObservabilityRepo) ListRecentAgentEvents(ctx context.Context, limit int) ([]models.AgentRunEvent, error) {
	if limit < 1 || limit > 200 {
		limit = 75
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, run_id, event_type, tool_name, payload, latency_ms, success, created_at
		FROM agent_run_events
		ORDER BY created_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.AgentRunEvent
	for rows.Next() {
		var item models.AgentRunEvent
		var raw []byte
		if err := rows.Scan(
			&item.ID, &item.RunID, &item.EventType, &item.ToolName, &raw,
			&item.LatencyMS, &item.Success, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		item.Payload = decodeJSONMap(raw)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *ObservabilityRepo) GetObservabilityOverview(ctx context.Context) (*models.ObservabilityOverview, error) {
	var out models.ObservabilityOverview
	var lastRunAt *time.Time
	err := r.db.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*)::int FROM audit_logs),
			(SELECT COUNT(*)::int FROM agent_runs WHERE created_at >= NOW() - INTERVAL '24 hours'),
			(SELECT COALESCE(AVG(CASE WHEN success THEN 100.0 ELSE 0.0 END), 0)
			   FROM ai_eval_logs WHERE created_at >= NOW() - INTERVAL '24 hours'),
			(SELECT COALESCE(AVG(latency_ms), 0)::int
			   FROM agent_runs WHERE finished_at IS NOT NULL AND created_at >= NOW() - INTERVAL '24 hours'),
			(SELECT COUNT(*)::int FROM agent_runs WHERE status = 'running'),
			(SELECT MAX(created_at) FROM agent_runs)`).
		Scan(&out.TotalAuditLogs, &out.AgentRuns24h, &out.AgentSuccessRate24h, &out.AvgAgentLatencyMS, &out.ActiveAgentRuns, &lastRunAt)
	if err != nil {
		return nil, fmt.Errorf("overview: %w", err)
	}
	out.LastAgentRunAt = lastRunAt
	return &out, nil
}

func decodeJSONMap(raw []byte) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil || out == nil {
		return map[string]any{}
	}
	return out
}
