package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

// HistoryMessage is one turn in chat history sent from the frontend.
type HistoryMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ParseHistoryJSON parses the JSON history array sent from the HTTP handler.
func ParseHistoryJSON(raw string) ([]HistoryMessage, error) {
	if raw == "" || raw == "[]" {
		return nil, nil
	}
	var items []HistoryMessage
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, err
	}
	msgs := make([]HistoryMessage, 0, len(items))
	for _, item := range items {
		if item.Role != "user" && item.Role != "assistant" {
			continue
		}
		content := strings.TrimSpace(item.Content)
		if content == "" {
			continue
		}
		if len(content) > MaxMessageLength {
			content = content[:MaxMessageLength]
		}
		msgs = append(msgs, HistoryMessage{Role: item.Role, Content: content})
	}
	if len(msgs) > 20 {
		msgs = msgs[len(msgs)-20:]
	}
	return msgs, nil
}

func formatMessageWithHistory(message string, history []HistoryMessage) string {
	if len(history) == 0 {
		return message
	}
	if len(history) > 6 {
		history = history[len(history)-6:]
	}
	var b strings.Builder
	b.WriteString("Conversation so far:\n")
	for _, h := range history {
		fmt.Fprintf(&b, "%s: %s\n", h.Role, h.Content)
	}
	b.WriteString("\nLatest user message: ")
	b.WriteString(message)
	return b.String()
}
