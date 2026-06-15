package ai

import "time"

type StreamEventType string

const (
	StreamEventStatus   StreamEventType = "status"
	StreamEventToolCall StreamEventType = "tool_call"
	StreamEventToolDone StreamEventType = "tool_done"
)

type StreamEvent struct {
	Type      StreamEventType `json:"type"`
	Message   string          `json:"message,omitempty"`
	ToolName  string          `json:"tool_name,omitempty"`
	Success   bool            `json:"success,omitempty"`
	LatencyMS int             `json:"latency_ms,omitempty"`
}

type RunMetadata struct {
	SessionID   string
	IP          string
	UserAgent   string
	StartedAt   time.Time
	UserMessage string
}
