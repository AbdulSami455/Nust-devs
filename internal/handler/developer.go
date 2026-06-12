package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/abdulsami/nust-devs/internal/githubutil"
	"github.com/abdulsami/nust-devs/internal/models"
	"github.com/abdulsami/nust-devs/internal/repository"
	"github.com/abdulsami/nust-devs/internal/worker"
	"github.com/hibiken/asynq"
)

type DeveloperHandler struct {
	devs  *repository.DeveloperRepo
	asynq *asynq.Client
}

func NewDeveloperHandler(devs *repository.DeveloperRepo, asynqClient *asynq.Client) *DeveloperHandler {
	return &DeveloperHandler{devs: devs, asynq: asynqClient}
}

func (h *DeveloperHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in models.CreateDeveloperInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(in.GithubUsername) == "" {
		writeError(w, http.StatusBadRequest, "github_username is required")
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
	in.Notes = optionalString(in.Notes)

	dev, err := h.devs.Create(r.Context(), in)
	if err != nil {
		if errors.Is(err, repository.ErrDeveloperDuplicate) {
			writeError(w, http.StatusConflict, "this github profile is already registered")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not create developer")
		return
	}
	if h.asynq != nil {
		task, err := worker.NewSyncDeveloperTask(dev.ID, dev.GithubUsername)
		if err != nil {
			slog.Warn("failed to build developer sync task", "developer", dev.GithubUsername, "err", err)
		} else if _, err := h.asynq.Enqueue(task); err != nil {
			slog.Warn("failed to enqueue initial developer sync", "developer", dev.GithubUsername, "err", err)
		}
	}
	writeJSON(w, http.StatusCreated, dev)
}

func (h *DeveloperHandler) List(w http.ResponseWriter, r *http.Request) {
	devs, err := h.devs.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list developers")
		return
	}
	if devs == nil {
		devs = []models.Developer{}
	}
	writeJSON(w, http.StatusOK, devs)
}

func (h *DeveloperHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in models.UpdateDeveloperInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	dev, err := h.devs.Update(r.Context(), id, in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update developer")
		return
	}
	writeJSON(w, http.StatusOK, dev)
}

func (h *DeveloperHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.devs.Delete(r.Context(), id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "developer not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not delete developer")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func optionalString(s *string) *string {
	if s == nil {
		return nil
	}
	t := strings.TrimSpace(*s)
	if t == "" {
		return nil
	}
	return &t
}
