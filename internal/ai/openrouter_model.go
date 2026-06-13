package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"strings"
	"time"

	"google.golang.org/genai"

	adkmodel "google.golang.org/adk/model"
)

const openRouterURL = "https://openrouter.ai/api/v1/chat/completions"

type openRouterModel struct {
	apiKey string
	name   string
	http   *http.Client
}

func newOpenRouterModel(apiKey, model string) adkmodel.LLM {
	return &openRouterModel{
		apiKey: apiKey,
		name:   model,
		http:   &http.Client{Timeout: 120 * time.Second},
	}
}

func (m *openRouterModel) Name() string { return m.name }

func (m *openRouterModel) GenerateContent(ctx context.Context, req *adkmodel.LLMRequest, stream bool) iter.Seq2[*adkmodel.LLMResponse, error] {
	msgs, tools := convertGenaiRequest(req)
	if len(tools) > 0 {
		stream = false
	}
	if stream {
		return m.generateStream(ctx, msgs, tools)
	}
	return func(yield func(*adkmodel.LLMResponse, error) bool) {
		resp, err := m.complete(ctx, msgs, tools)
		yield(resp, err)
	}
}

type orMessage struct {
	Role       string          `json:"role"`
	Content    string          `json:"content,omitempty"`
	ToolCallID string          `json:"tool_call_id,omitempty"`
	ToolCalls  []orToolCallMsg `json:"tool_calls,omitempty"`
}

type orToolCallMsg struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type orToolDef struct {
	Type     string `json:"type"`
	Function struct {
		Name        string         `json:"name"`
		Description string         `json:"description"`
		Parameters  map[string]any `json:"parameters"`
	} `json:"function"`
}

type orChatRequest struct {
	Model      string      `json:"model"`
	Messages   []orMessage `json:"messages"`
	Tools      []orToolDef `json:"tools,omitempty"`
	Stream     bool        `json:"stream"`
	MaxTokens  int         `json:"max_tokens"`
	ToolChoice string      `json:"tool_choice,omitempty"`
}

type orChatResponse struct {
	Choices []struct {
		Message struct {
			Role      string          `json:"role"`
			Content   string          `json:"content"`
			ToolCalls []orToolCallMsg `json:"tool_calls"`
		} `json:"message"`
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

func convertGenaiRequest(req *adkmodel.LLMRequest) ([]orMessage, []orToolDef) {
	var msgs []orMessage
	if req.Config != nil && req.Config.SystemInstruction != nil {
		if text := contentText(req.Config.SystemInstruction); text != "" {
			msgs = append(msgs, orMessage{Role: "system", Content: text})
		}
	}
	for _, c := range req.Contents {
		msgs = appendContentMessages(msgs, c)
	}
	return msgs, extractOpenAITools(req.Config)
}

func appendContentMessages(msgs []orMessage, c *genai.Content) []orMessage {
	if c == nil {
		return msgs
	}

	var textParts []string
	var fnCalls []orToolCallMsg
	var fnResponses []struct {
		id      string
		content string
	}

	for _, p := range c.Parts {
		if p == nil {
			continue
		}
		if p.Text != "" && !p.Thought {
			textParts = append(textParts, p.Text)
		}
		if p.FunctionCall != nil {
			args, _ := json.Marshal(p.FunctionCall.Args)
			var tc orToolCallMsg
			tc.ID = p.FunctionCall.ID
			if tc.ID == "" {
				tc.ID = "call_" + p.FunctionCall.Name
			}
			tc.Type = "function"
			tc.Function.Name = p.FunctionCall.Name
			tc.Function.Arguments = string(args)
			fnCalls = append(fnCalls, tc)
		}
		if p.FunctionResponse != nil {
			resp, _ := json.Marshal(p.FunctionResponse.Response)
			id := p.FunctionResponse.ID
			if id == "" {
				id = "call_" + p.FunctionResponse.Name
			}
			fnResponses = append(fnResponses, struct {
				id      string
				content string
			}{id: id, content: string(resp)})
		}
	}

	if len(fnResponses) > 0 {
		for _, fr := range fnResponses {
			msgs = append(msgs, orMessage{Role: "tool", ToolCallID: fr.id, Content: fr.content})
		}
		return msgs
	}

	if len(fnCalls) > 0 {
		msg := orMessage{Role: "assistant", ToolCalls: fnCalls}
		if len(textParts) > 0 {
			msg.Content = strings.Join(textParts, "")
		}
		return append(msgs, msg)
	}

	if len(textParts) == 0 {
		return msgs
	}
	return append(msgs, orMessage{Role: mapGenaiRole(c.Role), Content: strings.Join(textParts, "")})
}

func mapGenaiRole(role string) string {
	switch role {
	case genai.RoleModel, "assistant":
		return "assistant"
	case "system":
		return "system"
	default:
		return "user"
	}
}

func contentText(c *genai.Content) string {
	if c == nil {
		return ""
	}
	var parts []string
	for _, p := range c.Parts {
		if p != nil && p.Text != "" && !p.Thought {
			parts = append(parts, p.Text)
		}
	}
	return strings.Join(parts, "")
}

func extractOpenAITools(cfg *genai.GenerateContentConfig) []orToolDef {
	if cfg == nil {
		return nil
	}
	var out []orToolDef
	for _, t := range cfg.Tools {
		if t == nil {
			continue
		}
		for _, fd := range t.FunctionDeclarations {
			if fd == nil {
				continue
			}
			var def orToolDef
			def.Type = "function"
			def.Function.Name = fd.Name
			def.Function.Description = fd.Description
			def.Function.Parameters = schemaToMap(fd.Parameters)
			out = append(out, def)
		}
	}
	return out
}

func schemaToMap(s *genai.Schema) map[string]any {
	if s == nil {
		return map[string]any{"type": "object", "properties": map[string]any{}}
	}
	b, err := json.Marshal(s)
	if err != nil {
		return map[string]any{"type": "object", "properties": map[string]any{}}
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return map[string]any{"type": "object", "properties": map[string]any{}}
	}
	return m
}

func (m *openRouterModel) complete(ctx context.Context, msgs []orMessage, tools []orToolDef) (*adkmodel.LLMResponse, error) {
	body := orChatRequest{
		Model:     m.name,
		Messages:  msgs,
		Tools:     tools,
		Stream:    false,
		MaxTokens: 1024,
	}
	if len(tools) > 0 {
		body.ToolChoice = "auto"
	}
	data, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openRouterURL, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	m.setHeaders(req)

	resp, err := m.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openrouter %d: %s", resp.StatusCode, string(b))
	}

	var out orChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Choices) == 0 {
		return nil, fmt.Errorf("empty openrouter response")
	}
	return openAIChoiceToLLMResponse(out.Choices[0].Message, out.Choices[0].FinishReason), nil
}

func (m *openRouterModel) generateStream(ctx context.Context, msgs []orMessage, tools []orToolDef) iter.Seq2[*adkmodel.LLMResponse, error] {
	return func(yield func(*adkmodel.LLMResponse, error) bool) {
		body := orChatRequest{
			Model:     m.name,
			Messages:  msgs,
			Tools:     tools,
			Stream:    true,
			MaxTokens: 900,
		}
		data, _ := json.Marshal(body)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, openRouterURL, bytes.NewReader(data))
		if err != nil {
			yield(nil, err)
			return
		}
		m.setHeaders(req)

		resp, err := m.http.Do(req)
		if err != nil {
			yield(nil, err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			yield(nil, fmt.Errorf("openrouter %d: %s", resp.StatusCode, string(b)))
			return
		}

		var accumulated strings.Builder
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			payload := strings.TrimPrefix(line, "data: ")
			if payload == "[DONE]" {
				break
			}
			var chunk orChatResponse
			if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
				continue
			}
			if len(chunk.Choices) == 0 {
				continue
			}
			delta := chunk.Choices[0].Delta.Content
			if delta == "" {
				continue
			}
			accumulated.WriteString(delta)
			if !yield(&adkmodel.LLMResponse{
				Content: &genai.Content{
					Role:  genai.RoleModel,
					Parts: []*genai.Part{{Text: accumulated.String()}},
				},
				Partial: true,
			}, nil) {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			yield(nil, err)
			return
		}
		finalText := accumulated.String()
		yield(&adkmodel.LLMResponse{
			Content: &genai.Content{
				Role:  genai.RoleModel,
				Parts: []*genai.Part{{Text: finalText}},
			},
			Partial:      false,
			TurnComplete: true,
			FinishReason: genai.FinishReasonStop,
		}, nil)
	}
}

func openAIChoiceToLLMResponse(msg struct {
	Role      string          `json:"role"`
	Content   string          `json:"content"`
	ToolCalls []orToolCallMsg `json:"tool_calls"`
}, finishReason string) *adkmodel.LLMResponse {
	parts := make([]*genai.Part, 0, len(msg.ToolCalls)+1)
	if msg.Content != "" {
		parts = append(parts, &genai.Part{Text: msg.Content})
	}
	for _, tc := range msg.ToolCalls {
		var args map[string]any
		_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
		id := tc.ID
		if id == "" {
			id = "call_" + tc.Function.Name
		}
		parts = append(parts, &genai.Part{
			FunctionCall: &genai.FunctionCall{
				ID:   id,
				Name: tc.Function.Name,
				Args: args,
			},
		})
	}
	_ = finishReason
	return &adkmodel.LLMResponse{
		Content:      &genai.Content{Role: genai.RoleModel, Parts: parts},
		FinishReason: genai.FinishReasonStop,
		TurnComplete: true,
	}
}

func (m *openRouterModel) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://nustdevs.com")
	req.Header.Set("X-Title", "NUST Devs AI")
}
