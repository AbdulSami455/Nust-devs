package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/abdulsami/nust-devs/internal/models"
	"github.com/abdulsami/nust-devs/internal/repository"
)

type DeveloperHandler struct {
	devs *repository.DeveloperRepo
}

func NewDeveloperHandler(devs *repository.DeveloperRepo) *DeveloperHandler {
	return &DeveloperHandler{devs: devs}
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
	in.GithubUsername = strings.TrimSpace(in.GithubUsername)
	in.Email = optionalString(in.Email)
	in.DisplayName = optionalString(in.DisplayName)
	in.Notes = optionalString(in.Notes)

	dev, err := h.devs.Create(r.Context(), in)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			writeError(w, http.StatusConflict, "developer already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not create developer")
		return
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
