package ai

import (
	"context"
	"fmt"
	"strings"

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
}

// NewChatService wires the ADK llmagent, tools, and runner.
func NewChatService(ctx context.Context, cfg *config.Config, stats *repository.StatsRepo) (*ChatService, error) {
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

	return &ChatService{runner: r}, nil
}

// RunStreaming executes the agent and streams the final answer tokens to ch.
func (s *ChatService) RunStreaming(ctx context.Context, message string, history []HistoryMessage, ch chan<- string) error {
	sessionID := uuid.NewString()
	content := genai.NewContentFromText(formatMessageWithHistory(message, history), genai.RoleUser)

	var sent int
	for event, err := range s.runner.Run(ctx, "chat", sessionID, content, agent.RunConfig{
		StreamingMode: agent.StreamingModeSSE,
	}) {
		if err != nil {
			return err
		}
		if event == nil || eventHasToolParts(event) {
			continue
		}
		text := eventText(event)
		if len(text) <= sent {
			continue
		}
		delta := text[sent:]
		sent = len(text)
		for _, word := range strings.SplitAfter(delta, " ") {
			select {
			case ch <- word:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	if sent == 0 {
		fallback := SanitizeOutput("I don't have enough data to answer that right now.")
		for _, word := range strings.SplitAfter(fallback, " ") {
			select {
			case ch <- word:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return nil
}

// RunSync runs the agent to completion and returns the final text response.
func (s *ChatService) RunSync(ctx context.Context, prompt string) (string, error) {
	sessionID := uuid.NewString()
	content := genai.NewContentFromText(prompt, genai.RoleUser)

	var finalText string
	for event, err := range s.runner.Run(ctx, "summary", sessionID, content, agent.RunConfig{
		StreamingMode: agent.StreamingModeNone,
	}) {
		if err != nil {
			return "", err
		}
		if event != nil && event.IsFinalResponse() {
			finalText = eventText(event)
		}
	}
	if finalText == "" {
		return "", fmt.Errorf("empty response")
	}
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
