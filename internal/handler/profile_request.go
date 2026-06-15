package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/abdulsami/nust-devs/internal/githubutil"
	"github.com/abdulsami/nust-devs/internal/middleware"
	"github.com/abdulsami/nust-devs/internal/models"
	"github.com/abdulsami/nust-devs/internal/repository"
	"github.com/abdulsami/nust-devs/internal/worker"
	"github.com/hibiken/asynq"
)

type ProfileRequestHandler struct {
	requests *repository.RequestRepo
	devs     *repository.DeveloperRepo
	asynq    *asynq.Client
	obs      *repository.ObservabilityRepo
}

func NewProfileRequestHandler(
	requests *repository.RequestRepo,
	devs *repository.DeveloperRepo,
	asynqClient *asynq.Client,
	obs *repository.ObservabilityRepo,
) *ProfileRequestHandler {
	return &ProfileRequestHandler{requests: requests, devs: devs, asynq: asynqClient, obs: obs}
}

func (h *ProfileRequestHandler) Submit(w http.ResponseWriter, r *http.Request) {
	var in models.SubmitProfileRequestInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	username, err := githubutil.NormalizeUsername(in.GithubUsername)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid github username")
		return
	}
	in.GithubUsername = username
	in.Email = optionalString(in.Email)
	in.DisplayName = optionalString(in.DisplayName)
	in.Batch = optionalString(in.Batch)
	in.Course = optionalString(in.Course)
	in.Message = optionalString(in.Message)

	req, err := h.requests.Create(r.Context(), in)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDeveloperExists):
			writeError(w, http.StatusConflict, "this github profile is already on NUST Devs")
		case errors.Is(err, repository.ErrRequestPending):
			writeError(w, http.StatusConflict, "a pending request already exists for this username")
		default:
			writeError(w, http.StatusInternalServerError, "could not submit request")
		}
		return
	}
	h.logRequestAction(r, "profile_request.submit", req.ID, map[string]any{"github_username": req.GithubUsername})
	writeJSON(w, http.StatusCreated, req)
}

func (h *ProfileRequestHandler) CheckUsername(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	normalized, err := githubutil.NormalizeUsername(username)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"available": false,
			"reason":    "invalid username",
		})
		return
	}

	ctx := r.Context()
	if exists, _ := h.requests.DeveloperExists(ctx, normalized); exists {
		writeJSON(w, http.StatusOK, map[string]any{
			"available": false,
			"reason":    "already registered",
			"username":  normalized,
		})
		return
	}
	if pending, _ := h.requests.PendingRequestExists(ctx, normalized); pending {
		writeJSON(w, http.StatusOK, map[string]any{
			"available": false,
			"reason":    "request pending",
			"username":  normalized,
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"available": true,
		"username":  normalized,
	})
}

func (h *ProfileRequestHandler) List(w http.ResponseWriter, r *http.Request) {
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	reqs, err := h.requests.List(r.Context(), status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list requests")
		return
	}
	if reqs == nil {
		reqs = []models.DeveloperRequest{}
	}
	writeJSON(w, http.StatusOK, reqs)
}

func (h *ProfileRequestHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in models.ReviewProfileRequestInput
	_ = json.NewDecoder(r.Body).Decode(&in)
	in.AdminNotes = optionalString(in.AdminNotes)

	req, err := h.requests.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "request not found")
		return
	}
	if req.Status != "pending" {
		writeError(w, http.StatusConflict, "request already reviewed")
		return
	}

	notes := "Approved from profile request"
	if in.AdminNotes != nil {
		notes = *in.AdminNotes
	}
	notes = appendRequestDetails(notes, req.Batch, req.Course)
	dev, err := h.devs.Create(r.Context(), models.CreateDeveloperInput{
		GithubUsername: req.GithubUsername,
		Email:          req.Email,
		DisplayName:    req.DisplayName,
		Notes:          &notes,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDeveloperDuplicate) {
			_, _ = h.requests.SetStatus(r.Context(), id, "rejected", strPtr("duplicate: already registered"))
			writeError(w, http.StatusConflict, "developer already registered")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not create developer")
		return
	}

	updated, err := h.requests.SetStatus(r.Context(), id, "approved", in.AdminNotes)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "developer created but request update failed")
		return
	}

	if h.asynq != nil {
		if task, err := worker.NewSyncDeveloperTask(dev.ID, dev.GithubUsername); err == nil {
			_, _ = h.asynq.Enqueue(task)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"request":   updated,
		"developer": dev,
	})
	h.logAdminAction(r, "profile_request.approve", updated.ID, map[string]any{
		"github_username": updated.GithubUsername,
		"developer_id":    dev.ID,
	})
}

func (h *ProfileRequestHandler) Reject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in models.ReviewProfileRequestInput
	_ = json.NewDecoder(r.Body).Decode(&in)
	in.AdminNotes = optionalString(in.AdminNotes)

	req, err := h.requests.SetStatus(r.Context(), id, "rejected", in.AdminNotes)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "request not found or already reviewed")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not reject request")
		return
	}
	h.logAdminAction(r, "profile_request.reject", req.ID, map[string]any{"github_username": req.GithubUsername})
	writeJSON(w, http.StatusOK, req)
}

func strPtr(s string) *string {
	return &s
}

func appendRequestDetails(notes string, batch, course *string) string {
	var details []string
	if batch != nil {
		details = append(details, "Batch: "+*batch)
	}
	if course != nil {
		details = append(details, "Course: "+*course)
	}
	if len(details) == 0 {
		return notes
	}
	return notes + " (" + strings.Join(details, ", ") + ")"
}

func (h *ProfileRequestHandler) logRequestAction(r *http.Request, action, resourceID string, metadata map[string]any) {
	if h.obs == nil {
		return
	}
	go h.obs.InsertAuditLog(context.Background(), repository.AuditLogInput{
		ActorType:    "public",
		Action:       action,
		ResourceType: "profile_request",
		ResourceID:   resourceID,
		Method:       r.Method,
		Path:         r.URL.Path,
		StatusCode:   http.StatusOK,
		IP:           r.RemoteAddr,
		UserAgent:    r.UserAgent(),
		Metadata:     metadata,
	})
}

func (h *ProfileRequestHandler) logAdminAction(r *http.Request, action, resourceID string, metadata map[string]any) {
	if h.obs == nil {
		return
	}
	adminID, _ := middleware.AdminIDFromContext(r.Context())
	go h.obs.InsertAuditLog(context.Background(), repository.AuditLogInput{
		ActorType:    "admin",
		ActorID:      adminID,
		Action:       action,
		ResourceType: "profile_request",
		ResourceID:   resourceID,
		Method:       r.Method,
		Path:         r.URL.Path,
		StatusCode:   http.StatusOK,
		IP:           r.RemoteAddr,
		UserAgent:    r.UserAgent(),
		Metadata:     metadata,
	})
}
