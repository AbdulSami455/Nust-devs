package github

import (
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type rateLimiter struct {
	mu        sync.Mutex
	remaining int
	resetAt   time.Time
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{remaining: 5000}
}

func (r *rateLimiter) update(resp *http.Response) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if v := resp.Header.Get("x-ratelimit-remaining"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			r.remaining = n
		}
	}
	if v := resp.Header.Get("x-ratelimit-reset"); v != "" {
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			r.resetAt = time.Unix(ts, 0)
		}
	}
}

// wait blocks if remaining quota is dangerously low.
func (r *rateLimiter) wait() {
	r.mu.Lock()
	remaining := r.remaining
	resetAt := r.resetAt
	r.mu.Unlock()

	if remaining < 100 {
		pause := time.Until(resetAt)
		if pause > 0 {
			slog.Warn("github rate limit low — pausing", "remaining", remaining, "resume_in", pause.Round(time.Second))
			time.Sleep(pause)
		}
	}
}

func (r *rateLimiter) State() (remaining int, resetAt time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.remaining, r.resetAt
}
