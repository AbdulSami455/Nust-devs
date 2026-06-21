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
	"github.com/abdulsami/nust-devs/internal/cache"
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
	chat            *ai.ChatService
	summaryS        *ai.SummaryService
	projectSummaryS *ai.ProjectSummaryService
	rankInsightS    *ai.RankInsightService
	tagsS           *ai.NormalizedTagsService
	profileS        *ai.ProfileInsightsService
	compareS        *ai.CompareService
	statsRepo       *repository.StatsRepo
	db              *pgxpool.Pool
	cache           *cache.Cache
	chatRL          *rateLimiter
	summaryRL       *rateLimiter
	projectRL       *rateLimiter
	insightRL       *rateLimiter
	compareRL       *rateLimiter
}

func NewAIHandler(
	chat *ai.ChatService,
	summaryS *ai.SummaryService,
	projectSummaryS *ai.ProjectSummaryService,
	rankInsightS *ai.RankInsightService,
	tagsS *ai.NormalizedTagsService,
	profileS *ai.ProfileInsightsService,
	compareS *ai.CompareService,
	statsRepo *repository.StatsRepo,
	db *pgxpool.Pool,
	cache *cache.Cache,
) *AIHandler {
	return &AIHandler{
		chat:            chat,
		summaryS:        summaryS,
		projectSummaryS: projectSummaryS,
		rankInsightS:    rankInsightS,
		tagsS:           tagsS,
		profileS:        profileS,
		compareS:        compareS,
		statsRepo:       statsRepo,
		db:              db,
		cache:           cache,
		chatRL:          newRateLimiter(20, time.Hour),
		summaryRL:       newRateLimiter(10, time.Hour),
		projectRL:       newRateLimiter(12, time.Hour),
		insightRL:       newRateLimiter(12, time.Hour),
		compareRL:       newRateLimiter(10, time.Hour),
	}
}

func (h *AIHandler) cachedJSON(w http.ResponseWriter, r *http.Request, key string, ttl time.Duration, fetch func() (any, error)) {
	if h.cache == nil {
		val, err := fetch()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		writeJSON(w, http.StatusOK, val)
		return
	}

	var dest any
	if hit, _ := h.cache.GetJSON(r.Context(), key, &dest); hit {
		writeJSON(w, http.StatusOK, dest)
		return
	}
	val, err := fetch()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	_ = h.cache.SetJSON(r.Context(), key, val, ttl)
	writeJSON(w, http.StatusOK, val)
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
	slog.Info("ai chat request", "ip", ip, "message_len", len(cleanMsg), "user_agent", r.UserAgent())

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
	slog.Info("ai chat completed", "ip", ip, "success", success, "latency_ms", latencyMS)
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
	slog.Info("ai developer summary request", "username", strings.ToLower(username), "ip", clientIP(r))

	// Rate limit
	if !h.insightRL.allow(clientIP(r)) {
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
		slog.Warn("ai developer summary unavailable", "username", strings.ToLower(username), "err", err)
		writeError(w, http.StatusServiceUnavailable, "summary not available")
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

// ── Project summary endpoint: GET /api/v1/repos/{id}/summary ─────────────────

func (h *AIHandler) GetProjectSummary(w http.ResponseWriter, r *http.Request) {
	repoID := strings.TrimSpace(r.PathValue("id"))
	if repoID == "" {
		writeError(w, http.StatusBadRequest, "invalid repository id")
		return
	}

	ip := clientIP(r)
	slog.Info("ai project summary request", "repo_id", repoID, "ip", ip)

	if !h.projectRL.allow(ip) {
		writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
		return
	}

	if h.projectSummaryS == nil {
		writeError(w, http.StatusServiceUnavailable, "project summary service unavailable")
		return
	}

	repo, err := h.statsRepo.GetProjectByID(r.Context(), repoID)
	if err != nil {
		writeError(w, http.StatusNotFound, "repository not found")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()

	cacheKey := fmt.Sprintf("projects:%s:summary", repo.ID)
	h.cachedJSON(w, r, cacheKey, 24*time.Hour, func() (any, error) {
		return h.projectSummaryS.Get(ctx, *repo)
	})
}

// ── Rank/badge explanation endpoint: GET /api/v1/developers/{username}/rank-insight ─

func (h *AIHandler) GetRankInsight(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.PathValue("username"))
	if !ai.ValidateUsername(username) {
		writeError(w, http.StatusBadRequest, "invalid username")
		return
	}
	slog.Info("ai rank insight request", "username", strings.ToLower(username), "ip", clientIP(r))

	if !h.summaryRL.allow(clientIP(r)) {
		writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
		return
	}

	if h.rankInsightS == nil {
		writeError(w, http.StatusServiceUnavailable, "rank insight service unavailable")
		return
	}

	cacheKey := fmt.Sprintf("developers:%s:rank-insight", strings.ToLower(username))
	h.cachedJSON(w, r, cacheKey, 12*time.Hour, func() (any, error) {
		return h.rankInsightS.Get(r.Context(), username)
	})
}

// ── Normalized tags endpoints ─────────────────────────────────────────────────

func (h *AIHandler) GetDeveloperNormalizedTags(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.PathValue("username"))
	if !ai.ValidateUsername(username) {
		writeError(w, http.StatusBadRequest, "invalid username")
		return
	}
	if h.tagsS == nil {
		writeError(w, http.StatusServiceUnavailable, "normalized tags service unavailable")
		return
	}
	cacheKey := fmt.Sprintf("developers:%s:normalized-tags", strings.ToLower(username))
	h.cachedJSON(w, r, cacheKey, 24*time.Hour, func() (any, error) {
		return h.tagsS.GetDeveloper(r.Context(), username)
	})
}

func (h *AIHandler) GetProjectNormalizedTags(w http.ResponseWriter, r *http.Request) {
	repoID := strings.TrimSpace(r.PathValue("id"))
	if repoID == "" {
		writeError(w, http.StatusBadRequest, "invalid repository id")
		return
	}
	if h.tagsS == nil {
		writeError(w, http.StatusServiceUnavailable, "normalized tags service unavailable")
		return
	}
	cacheKey := fmt.Sprintf("projects:%s:normalized-tags", repoID)
	h.cachedJSON(w, r, cacheKey, 24*time.Hour, func() (any, error) {
		return h.tagsS.GetProject(r.Context(), repoID)
	})
}

// ── Developer profile insights endpoint: GET /api/v1/developers/{username}/profile-insights ─

func (h *AIHandler) GetProfileInsights(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.PathValue("username"))
	if !ai.ValidateUsername(username) {
		writeError(w, http.StatusBadRequest, "invalid username")
		return
	}
	if h.profileS == nil {
		writeError(w, http.StatusServiceUnavailable, "profile insights service unavailable")
		return
	}
	cacheKey := fmt.Sprintf("developers:%s:profile-insights", strings.ToLower(username))
	h.cachedJSON(w, r, cacheKey, 24*time.Hour, func() (any, error) {
		return h.profileS.Get(r.Context(), username)
	})
}

// ── Developer comparison endpoint: GET /api/v1/ai/compare ────────────────────

func (h *AIHandler) GetDeveloperComparison(w http.ResponseWriter, r *http.Request) {
	left := strings.TrimSpace(r.URL.Query().Get("left"))
	right := strings.TrimSpace(r.URL.Query().Get("right"))

	if !ai.ValidateUsername(left) || !ai.ValidateUsername(right) {
		writeError(w, http.StatusBadRequest, "invalid usernames")
		return
	}
	if strings.EqualFold(left, right) {
		writeError(w, http.StatusBadRequest, "pick two different developers")
		return
	}

	if !h.compareRL.allow(clientIP(r)) {
		writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
		return
	}

	if h.compareS == nil {
		writeError(w, http.StatusServiceUnavailable, "comparison service unavailable")
		return
	}

	slog.Info(
		"ai developer comparison request",
		"left", strings.ToLower(left),
		"right", strings.ToLower(right),
		"ip", clientIP(r),
	)

	cacheKey := fmt.Sprintf("developers:compare:%s:%s", strings.ToLower(left), strings.ToLower(right))
	h.cachedJSON(w, r, cacheKey, 15*time.Minute, func() (any, error) {
		return h.compareS.Get(r.Context(), left, right)
	})
}
