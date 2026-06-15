package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type rateLimitBucket struct {
	count   int
	resetAt time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*rateLimitBucket
	limit   int
	window  time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[string]*rateLimitBucket),
		limit:   limit,
		window:  window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		remaining, retryAfter, ok := rl.allow(clientIP(r))
		w.Header().Set("RateLimit-Limit", strconv.Itoa(rl.limit))
		w.Header().Set("RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("RateLimit-Reset", strconv.Itoa(secondsUntil(retryAfter)))
		if !ok {
			w.Header().Set("Retry-After", strconv.Itoa(secondsUntil(retryAfter)))
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allow(key string) (remaining int, retryAfter time.Duration, ok bool) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b := rl.buckets[key]
	if b == nil || now.After(b.resetAt) {
		rl.buckets[key] = &rateLimitBucket{count: 1, resetAt: now.Add(rl.window)}
		return rl.limit - 1, rl.window, true
	}

	retryAfter = time.Until(b.resetAt)
	if b.count >= rl.limit {
		return 0, retryAfter, false
	}

	b.count++
	return rl.limit - b.count, retryAfter, true
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, bucket := range rl.buckets {
			if now.After(bucket.resetAt) {
				delete(rl.buckets, key)
			}
		}
		rl.mu.Unlock()
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	if i := strings.LastIndex(r.RemoteAddr, ":"); i != -1 {
		return r.RemoteAddr[:i]
	}
	return r.RemoteAddr
}

func secondsUntil(d time.Duration) int {
	seconds := int(d.Round(time.Second).Seconds())
	if seconds < 1 {
		return 1
	}
	return seconds
}
