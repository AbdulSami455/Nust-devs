package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/abdulsami/nust-devs/internal/repository"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func AuditLogs(repo *repository.ObservabilityRepo, jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			start := time.Now()
			next.ServeHTTP(rec, r)

			actorType := "public"
			actorID := ""
			if adminID, ok := AdminIDFromContext(r.Context()); ok {
				actorType = "admin"
				actorID = adminID
			} else if adminID, ok := AdminIDFromRequest(r, jwtSecret); ok {
				actorType = "admin"
				actorID = adminID
			}

			go repo.InsertAuditLog(context.Background(), repository.AuditLogInput{
				ActorType:    actorType,
				ActorID:      actorID,
				Action:       "http.request",
				ResourceType: "request",
				ResourceID:   r.Pattern,
				Method:       r.Method,
				Path:         r.URL.Path,
				StatusCode:   rec.status,
				IP:           requestIP(r),
				UserAgent:    r.UserAgent(),
				Metadata: map[string]any{
					"duration_ms": time.Since(start).Milliseconds(),
					"query":       r.URL.RawQuery,
				},
			})
		})
	}
}

func requestIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	if i := strings.LastIndex(r.RemoteAddr, ":"); i != -1 {
		return r.RemoteAddr[:i]
	}
	return r.RemoteAddr
}
