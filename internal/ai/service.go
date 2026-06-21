package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/genai"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"

	"github.com/abdulsami/nust-devs/internal/config"
	"github.com/abdulsami/nust-devs/internal/repository"
)

const appName = "nust-devs"

// ChatService runs the Google ADK agent with in-process repository tools.
type ChatService struct {
	runner *runner.Runner
	obs    *repository.ObservabilityRepo
}

// NewChatService wires the ADK llmagent, tools, and runner.
func NewChatService(ctx context.Context, cfg *config.Config, stats *repository.StatsRepo, obs *repository.ObservabilityRepo) (*ChatService, error) {
	llm, err := NewLLM(ctx, cfg)
	if err != nil {
		return nil, err
	}

	tools, err := buildTools(stats)
	if err != nil {
		return nil, err
	}

	ag, err := llmagent.New(llmagent.Config{
		Name:        "nust_devs_assistant",
		Model:       llm,
		Description: "NUST Devs community assistant for developer stats and recruiting.",
		Instruction: systemPrompt,
		Tools:       tools,
	})
	if err != nil {
		return nil, fmt.Errorf("adk agent: %w", err)
	}

	sessionSvc := session.InMemoryService()
	r, err := runner.New(runner.Config{
		AppName:           appName,
		Agent:             ag,
		SessionService:    sessionSvc,
		AutoCreateSession: true,
	})
	if err != nil {
		return nil, fmt.Errorf("adk runner: %w", err)
	}

	return &ChatService{runner: r, obs: obs}, nil
}

// RunStreaming executes the agent and streams the final answer tokens to ch.
func (s *ChatService) RunStreaming(ctx context.Context, meta RunMetadata, history []HistoryMessage, ch chan<- string, emit func(StreamEvent)) error {
	sessionID := uuid.NewString()
	startedAt := meta.StartedAt
	if startedAt.IsZero() {
		startedAt = time.Now()
	}
	content := genai.NewContentFromText(formatMessageWithHistory(meta.UserMessage, history), genai.RoleUser)
	runID, err := s.startRun(ctx, RunMetadata{
		SessionID:   sessionID,
		StartedAt:   startedAt,
		UserMessage: meta.UserMessage,
		IP:          meta.IP,
		UserAgent:   meta.UserAgent,
	})
	if err != nil {
		slog.Warn("agent run log start failed", "err", err)
	}
	slog.Info("ai chat run started", "session_id", sessionID, "ip", meta.IP, "message_len", len(meta.UserMessage))
	emitSafe(emit, StreamEvent{Type: StreamEventStatus, Message: "Thinking"})

	// Non-streaming LLM mode so tool-call rounds complete correctly with OpenRouter.
	var finalText string
	toolCalls := 0
	callStartedAt := map[string]time.Time{}
	for event, err := range s.runner.Run(ctx, "chat", sessionID, content, agent.RunConfig{
		StreamingMode: agent.StreamingModeNone,
	}) {
		if err != nil {
			slog.Warn("adk runner error", "err", err)
			s.finishRun(context.Background(), runID, buildRunFinish(err, finalText, toolCalls, startedAt))
			return fmt.Errorf("agent error: %w", err)
		}
		if event == nil {
			continue
		}
		if err := s.inspectEvent(context.Background(), runID, event, emit, callStartedAt, &toolCalls); err != nil {
			s.finishRun(context.Background(), runID, buildRunFinish(err, finalText, toolCalls, startedAt))
			return err
		}
		if eventHasToolParts(event) {
			continue
		}
		if event.IsFinalResponse() {
			if text := strings.TrimSpace(eventText(event)); text != "" {
				finalText = text
			}
		}
	}

	if finalText == "" {
		finalText = "I couldn't retrieve developer data right now. Please try again."
	}
	finalText = SanitizeOutput(finalText)
	for _, word := range strings.SplitAfter(finalText, " ") {
		select {
		case ch <- word:
		case <-ctx.Done():
			s.finishRun(context.Background(), runID, buildRunFinish(ctx.Err(), finalText, toolCalls, startedAt))
			return ctx.Err()
		}
	}
	slog.Info("ai chat run finished", "session_id", sessionID, "tool_calls", toolCalls, "response_chars", len(finalText), "status", "completed")
	s.finishRun(context.Background(), runID, buildRunFinish(nil, finalText, toolCalls, startedAt))
	return nil
}

// RunSync runs the agent to completion and returns the final text response.
func (s *ChatService) RunSync(ctx context.Context, prompt string) (string, error) {
	sessionID := uuid.NewString()
	content := genai.NewContentFromText(prompt, genai.RoleUser)
	slog.Info("ai sync run started", "session_id", sessionID, "prompt_len", len(prompt))

	var finalText string
	for event, err := range s.runner.Run(ctx, "summary", sessionID, content, agent.RunConfig{
		StreamingMode: agent.StreamingModeNone,
	}) {
		if err != nil {
			return "", err
		}
		if event != nil && event.IsFinalResponse() && !eventHasToolParts(event) {
			if text := strings.TrimSpace(eventText(event)); text != "" {
				finalText = text
			}
		}
	}
	if finalText == "" {
		return "", fmt.Errorf("empty response")
	}
	slog.Info("ai sync run finished", "session_id", sessionID, "response_chars", len(finalText))
	return SanitizeOutput(finalText), nil
}

func eventText(e *session.Event) string {
	if e == nil || e.Content == nil {
		return ""
	}
	var b strings.Builder
	for _, p := range e.Content.Parts {
		if p == nil || p.Thought {
			continue
		}
		if p.Text != "" {
			b.WriteString(p.Text)
		}
	}
	return b.String()
}

func eventHasToolParts(e *session.Event) bool {
	if e == nil || e.Content == nil {
		return false
	}
	for _, p := range e.Content.Parts {
		if p == nil {
			continue
		}
		if p.FunctionCall != nil || p.FunctionResponse != nil {
			return true
		}
	}
	return false
}

func (s *ChatService) inspectEvent(
	ctx context.Context,
	runID string,
	event *session.Event,
	emit func(StreamEvent),
	callStartedAt map[string]time.Time,
	toolCalls *int,
) error {
	if event == nil || event.Content == nil {
		return nil
	}
	for _, p := range event.Content.Parts {
		if p == nil {
			continue
		}
		if p.ExecutableCode != nil || p.CodeExecutionResult != nil || p.ToolCall != nil || p.ToolResponse != nil {
			return fmt.Errorf("unsupported agent capability requested")
		}
		if p.FunctionCall != nil {
			name := strings.TrimSpace(p.FunctionCall.Name)
			if _, ok := allowedToolNames[name]; !ok {
				return fmt.Errorf("unexpected tool requested")
			}
			*toolCalls++
			if *toolCalls > 12 {
				return fmt.Errorf("tool call limit exceeded")
			}
			key := p.FunctionCall.ID
			if key == "" {
				key = name
			}
			callStartedAt[key] = time.Now()
			slog.Info("ai tool call", "run_id", runID, "tool", name, "args", sanitizeArgs(p.FunctionCall.Args))
			emitSafe(emit, StreamEvent{Type: StreamEventToolCall, ToolName: name, Message: toolStatusText(name)})
			s.logAgentEvent(ctx, runID, repository.AgentRunEventInput{
				RunID:     runID,
				EventType: "tool_call",
				ToolName:  name,
				Payload: map[string]any{
					"args": sanitizeArgs(p.FunctionCall.Args),
				},
				Success: true,
			})
			continue
		}
		if p.FunctionResponse != nil {
			name := strings.TrimSpace(p.FunctionResponse.Name)
			key := p.FunctionResponse.ID
			if key == "" {
				key = name
			}
			latency := 0
			if start, ok := callStartedAt[key]; ok {
				latency = int(time.Since(start).Milliseconds())
				delete(callStartedAt, key)
			}
			success := responseSucceeded(p.FunctionResponse.Response)
			slog.Info("ai tool done", "run_id", runID, "tool", name, "success", success, "latency_ms", latency)
			emitSafe(emit, StreamEvent{Type: StreamEventToolDone, ToolName: name, Success: success, LatencyMS: latency})
			s.logAgentEvent(ctx, runID, repository.AgentRunEventInput{
				RunID:     runID,
				EventType: "tool_done",
				ToolName:  name,
				Payload:   map[string]any{"response": sanitizeArgs(p.FunctionResponse.Response)},
				LatencyMS: latency,
				Success:   success,
			})
			continue
		}
	}
	return nil
}

func (s *ChatService) startRun(ctx context.Context, meta RunMetadata) (string, error) {
	if s.obs == nil {
		return "", nil
	}
	slog.Info("ai run log start", "session_id", meta.SessionID, "ip", meta.IP, "agent", "chat")
	return s.obs.StartAgentRun(ctx, repository.AgentRunInput{
		SessionID:   meta.SessionID,
		AgentName:   "chat",
		UserMessage: meta.UserMessage,
		InputHash:   HashInput(meta.UserMessage),
		IP:          meta.IP,
		UserAgent:   meta.UserAgent,
	})
}

func (s *ChatService) finishRun(ctx context.Context, runID string, in repository.AgentRunFinishInput) {
	if s.obs == nil || runID == "" {
		return
	}
	slog.Info("ai run log finish", "run_id", runID, "status", in.Status, "latency_ms", in.LatencyMS, "tool_calls", in.ToolCalls)
	if err := s.obs.FinishAgentRun(ctx, runID, in); err != nil {
		slog.Warn("agent run finish log failed", "err", err)
	}
}

func buildRunFinish(err error, finalText string, toolCalls int, startedAt time.Time) repository.AgentRunFinishInput {
	status := "completed"
	msg := ""
	if err != nil {
		status = "failed"
		msg = err.Error()
	}
	return repository.AgentRunFinishInput{
		Status:        status,
		LatencyMS:     int(time.Since(startedAt).Milliseconds()),
		ErrorMessage:  msg,
		ResponseChars: len(finalText),
		ToolCalls:     toolCalls,
	}
}

func (s *ChatService) logAgentEvent(ctx context.Context, runID string, in repository.AgentRunEventInput) {
	if s.obs == nil || runID == "" {
		return
	}
	if err := s.obs.InsertAgentRunEvent(ctx, in); err != nil {
		slog.Warn("agent event log failed", "err", err)
	}
}

func emitSafe(emit func(StreamEvent), ev StreamEvent) {
	if emit != nil {
		emit(ev)
	}
}

func toolStatusText(name string) string {
	switch name {
	case "get_developer_profile":
		return "Fetching developer profile"
	case "get_developer_repos":
		return "Loading repositories"
	case "get_developer_contribution_stats", "get_developer_contributions":
		return "Analyzing contributions"
	case "get_top_developers":
		return "Ranking developers"
	case "get_top_projects", "get_fastest_growing_projects":
		return "Inspecting projects"
	case "get_stats_overview", "get_language_stats", "get_community_activity", "get_innovation_graph":
		return "Collecting community stats"
	default:
		return "Running " + name
	}
}

func responseSucceeded(resp map[string]any) bool {
	if resp == nil {
		return true
	}
	_, hasError := resp["error"]
	return !hasError
}

func sanitizeArgs(v any) any {
	raw, err := json.Marshal(v)
	if err != nil {
		return map[string]any{}
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{}
	}
	return out
}
