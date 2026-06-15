package handler

import (
	"net/http"
	"strconv"

	"github.com/abdulsami/nust-devs/internal/models"
	"github.com/abdulsami/nust-devs/internal/repository"
)

type ObservabilityHandler struct {
	repo *repository.ObservabilityRepo
}

func NewObservabilityHandler(repo *repository.ObservabilityRepo) *ObservabilityHandler {
	return &ObservabilityHandler{repo: repo}
}

func (h *ObservabilityHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	overview, err := h.repo.GetObservabilityOverview(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load observability overview")
		return
	}
	logs, err := h.repo.ListAuditLogs(r.Context(), 25)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load audit logs")
		return
	}
	runs, err := h.repo.ListAgentRuns(r.Context(), 20)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load agent runs")
		return
	}
	events, err := h.repo.ListRecentAgentEvents(r.Context(), 40)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load agent events")
		return
	}
	writeJSON(w, http.StatusOK, models.ObservabilityResponse{
		Overview:     *overview,
		RecentLogs:   logs,
		RecentRuns:   runs,
		RecentEvents: events,
	})
}

func (h *ObservabilityHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	limit := clampLimit(r.URL.Query().Get("limit"), 50, 1, 200)
	logs, err := h.repo.ListAuditLogs(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load audit logs")
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

func (h *ObservabilityHandler) ListAgentRuns(w http.ResponseWriter, r *http.Request) {
	limit := clampLimit(r.URL.Query().Get("limit"), 25, 1, 100)
	runs, err := h.repo.ListAgentRuns(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load agent runs")
		return
	}
	writeJSON(w, http.StatusOK, runs)
}

func (h *ObservabilityHandler) ListAgentEvents(w http.ResponseWriter, r *http.Request) {
	limit := clampLimit(r.URL.Query().Get("limit"), 40, 1, 200)
	events, err := h.repo.ListRecentAgentEvents(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load agent events")
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func clampLimit(raw string, fallback, min, max int) int {
	n, err := strconv.Atoi(raw)
	if err != nil || n == 0 {
		return fallback
	}
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}
