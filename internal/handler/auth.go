package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/abdulsami/nust-devs/internal/middleware"
	"github.com/abdulsami/nust-devs/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	admins       *repository.AdminRepo
	jwtSecret    string
	secureCookie bool
	loginLimiter *loginRateLimiter
}

func NewAuthHandler(admins *repository.AdminRepo, jwtSecret string, secureCookie bool) *AuthHandler {
	return &AuthHandler{
		admins:       admins,
		jwtSecret:    jwtSecret,
		secureCookie: secureCookie,
		loginLimiter: newLoginRateLimiter(),
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ip := loginClientIP(r)
	if retryAfter, ok := h.loginLimiter.allowIP(ip); !ok {
		setRetryAfter(w, retryAfter)
		writeError(w, http.StatusTooManyRequests, "too many login attempts")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 4*1024)
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Email == "" || body.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password required")
		return
	}

	email := strings.ToLower(strings.TrimSpace(body.Email))
	if email == "" {
		writeError(w, http.StatusBadRequest, "email and password required")
		return
	}
	if retryAfter, ok := h.loginLimiter.allowCredential(ip, email); !ok {
		setRetryAfter(w, retryAfter)
		writeError(w, http.StatusTooManyRequests, "too many login attempts")
		return
	}

	admin, err := h.admins.GetByEmail(r.Context(), email)
	if err != nil {
		h.loginLimiter.recordFailure(ip, email)
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(body.Password)); err != nil {
		h.loginLimiter.recordFailure(ip, email)
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	h.loginLimiter.resetCredential(ip, email)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": admin.ID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not sign token")
		return
	}

	http.SetCookie(w, h.authCookie(signed, time.Now().Add(24*time.Hour)))
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, h.authCookie("", time.Unix(0, 0)))
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *AuthHandler) authCookie(value string, expires time.Time) *http.Cookie {
	maxAge := int(time.Until(expires).Seconds())
	if value == "" {
		maxAge = -1
	}
	sameSite := http.SameSiteLaxMode
	if h.secureCookie {
		sameSite = http.SameSiteNoneMode
	}
	return &http.Cookie{
		Name:     middleware.AdminTokenCookie,
		Value:    value,
		Path:     "/",
		Expires:  expires,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: sameSite,
	}
}

type loginBucket struct {
	count   int
	resetAt time.Time
}

type loginRateLimiter struct {
	mu                sync.Mutex
	ipBuckets         map[string]*loginBucket
	credentialBuckets map[string]*loginBucket
	ipLimit           int
	credentialLimit   int
	window            time.Duration
}

func newLoginRateLimiter() *loginRateLimiter {
	rl := &loginRateLimiter{
		ipBuckets:         make(map[string]*loginBucket),
		credentialBuckets: make(map[string]*loginBucket),
		ipLimit:           30,
		credentialLimit:   5,
		window:            15 * time.Minute,
	}
	go rl.cleanup()
	return rl
}

func (rl *loginRateLimiter) allowIP(ip string) (time.Duration, bool) {
	return rl.increment(rl.ipBuckets, ip, rl.ipLimit)
}

func (rl *loginRateLimiter) allowCredential(ip, email string) (time.Duration, bool) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	b := rl.credentialBuckets[credentialKey(ip, email)]
	if b == nil || time.Now().After(b.resetAt) {
		return 0, true
	}
	if b.count >= rl.credentialLimit {
		return time.Until(b.resetAt).Round(time.Second), false
	}
	return 0, true
}

func (rl *loginRateLimiter) recordFailure(ip, email string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	key := credentialKey(ip, email)
	now := time.Now()
	b := rl.credentialBuckets[key]
	if b == nil || now.After(b.resetAt) {
		rl.credentialBuckets[key] = &loginBucket{count: 1, resetAt: now.Add(rl.window)}
		return
	}
	b.count++
}

func (rl *loginRateLimiter) resetCredential(ip, email string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.credentialBuckets, credentialKey(ip, email))
}

func (rl *loginRateLimiter) increment(buckets map[string]*loginBucket, key string, limit int) (time.Duration, bool) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	b := buckets[key]
	if b == nil || now.After(b.resetAt) {
		buckets[key] = &loginBucket{count: 1, resetAt: now.Add(rl.window)}
		return 0, true
	}
	if b.count >= limit {
		return time.Until(b.resetAt).Round(time.Second), false
	}
	b.count++
	return 0, true
}

func (rl *loginRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, bucket := range rl.ipBuckets {
			if now.After(bucket.resetAt) {
				delete(rl.ipBuckets, key)
			}
		}
		for key, bucket := range rl.credentialBuckets {
			if now.After(bucket.resetAt) {
				delete(rl.credentialBuckets, key)
			}
		}
		rl.mu.Unlock()
	}
}

func credentialKey(ip, email string) string {
	return ip + "|" + email
}

func loginClientIP(r *http.Request) string {
	ip := r.RemoteAddr
	if i := strings.LastIndex(ip, ":"); i != -1 {
		ip = ip[:i]
	}
	return ip
}

func setRetryAfter(w http.ResponseWriter, retryAfter time.Duration) {
	seconds := int(retryAfter.Seconds())
	if seconds < 1 {
		seconds = 1
	}
	w.Header().Set("Retry-After", strconv.Itoa(seconds))
}
