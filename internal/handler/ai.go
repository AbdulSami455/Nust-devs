package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/abdulsami/nust-devs/internal/ai"
	"github.com/abdulsami/nust-devs/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ── Rate limiter ──────────────────────────────────────────────────────────────

type ipBucket struct {
	count   int
	resetAt time.Time
}

type rateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*ipBucket
	limit   int
	window  time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		buckets: make(map[string]*ipBucket),
		limit:   limit,
		window:  window,
	}
	// Background cleanup every 10 minutes
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		for range ticker.C {
			rl.mu.Lock()
			now := time.Now()
			for ip, b := range rl.buckets {
				if now.After(b.resetAt) {
					delete(rl.buckets, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()
	return rl
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.buckets[ip]
	if !ok || now.After(b.resetAt) {
		rl.buckets[ip] = &ipBucket{count: 1, resetAt: now.Add(rl.window)}
		return true
	}
	if b.count >= rl.limit {
		return false
	}
	b.count++
	return true
}

func clientIP(r *http.Request) string {
	// Trust X-Forwarded-For only from private/local network or if you control the proxy.
	// For simplicity we use RemoteAddr; swap for XFF if behind a known proxy.
	ip := r.RemoteAddr
	if i := strings.LastIndex(ip, ":"); i != -1 {
		ip = ip[:i]
	}
	return ip
}

// ── Handler ────────────────────────────────────────────────────────────────────

type AIHandler struct {
	chat      *ai.ChatService
	summaryS  *ai.SummaryService
	statsRepo *repository.StatsRepo
	db        *pgxpool.Pool
	chatRL    *rateLimiter
	summaryRL *rateLimiter
}

func NewAIHandler(
	chat *ai.ChatService,
	summaryS *ai.SummaryService,
	statsRepo *repository.StatsRepo,
	db *pgxpool.Pool,
) *AIHandler {
	return &AIHandler{
		chat:      chat,
		summaryS:  summaryS,
		statsRepo: statsRepo,
		db:        db,
		chatRL:    newRateLimiter(20, time.Hour),
		summaryRL: newRateLimiter(10, time.Hour),
	}
}

// ── Chat SSE endpoint: POST /api/v1/ai/chat ────────────────────────────────────

type chatRequest struct {
	Message string `json:"message"`
	History []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"history"`
}

// Chat handles POST /api/v1/ai/chat
// It validates the request, runs the agent, and streams the response via SSE.
func (h *AIHandler) Chat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Rate limit
	ip := clientIP(r)
	if !h.chatRL.allow(ip) {
		writeError(w, http.StatusTooManyRequests, "rate limit exceeded: 20 messages per hour")
		return
	}

	// Parse body
	r.Body = http.MaxBytesReader(w, r.Body, 8*1024) // 8 KB max body
	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Input guardrail
	cleanMsg, err := ai.ValidateMessage(req.Message)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	// Build validated history
	var historyJSON []byte
	if len(req.History) > 0 {
		historyJSON, _ = json.Marshal(req.History)
	}
	history, _ := ai.ParseHistoryJSON(string(historyJSON))

	// Setup SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // nginx: disable buffering

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()

	// Token channel
	ch := make(chan string, 32)
	t0 := time.Now()
	var agentErr error
	var writeMu sync.Mutex

	go func() {
		defer close(ch)
		agentErr = h.chat.RunStreaming(ctx, ai.RunMetadata{
			StartedAt:   t0,
			UserMessage: cleanMsg,
			IP:          ip,
			UserAgent:   r.UserAgent(),
		}, history, ch, func(ev ai.StreamEvent) {
			writeMu.Lock()
			defer writeMu.Unlock()
			_ = writeSSEEvent(w, flusher, string(ev.Type), ev)
		})
	}()

	var buf strings.Builder
	for token := range ch {
		writeMu.Lock()
		buf.WriteString(token)
		_ = writeSSEEvent(w, flusher, "token", token)
		writeMu.Unlock()
	}

	// Send done event
	writeMu.Lock()
	_ = writeSSEEvent(w, flusher, "done", map[string]any{})
	writeMu.Unlock()

	// Eval log (best-effort, async)
	latencyMS := int(time.Since(t0).Milliseconds())
	success := agentErr == nil
	if !success {
		slog.Warn("chat agent error", "err", agentErr, "ip", ip)
	}
	go ai.LogEval(context.Background(), h.db, "chat",
		ai.HashInput(cleanMsg),
		map[string]any{"response_len": buf.Len(), "rounds": "n/a"},
		latencyMS, success,
	)
}

func writeSSEEvent(w http.ResponseWriter, flusher http.Flusher, event string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

// ── Developer summary endpoint: GET /api/v1/developers/{username}/summary ─────

// GetDeveloperSummary handles GET /api/v1/developers/{username}/summary
func (h *AIHandler) GetDeveloperSummary(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if !ai.ValidateUsername(username) {
		writeError(w, http.StatusBadRequest, "invalid username")
		return
	}

	// Rate limit
	if !h.summaryRL.allow(clientIP(r)) {
		writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
		return
	}

	// Resolve developer ID from username
	dev, err := h.statsRepo.GetDeveloperByUsername(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusNotFound, "developer not found")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()

	summary, err := h.summaryS.Get(ctx, dev.ID, username)
	if err != nil || summary == nil {
		writeError(w, http.StatusServiceUnavailable, "summary not available")
		return
	}

	writeJSON(w, http.StatusOK, summary)
}
