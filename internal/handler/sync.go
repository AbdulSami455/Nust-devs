package handler

import (
	"net/http"
	"time"

	"github.com/abdulsami/nust-devs/internal/repository"
	"github.com/abdulsami/nust-devs/internal/worker"
	gh "github.com/abdulsami/nust-devs/internal/github"
	"github.com/hibiken/asynq"
)

type SyncHandler struct {
	devRepo  *repository.DeveloperRepo
	client   *asynq.Client
	ghClient *gh.Client
}

func NewSyncHandler(devRepo *repository.DeveloperRepo, client *asynq.Client, ghClient *gh.Client) *SyncHandler {
	return &SyncHandler{devRepo: devRepo, client: client, ghClient: ghClient}
}

// TriggerSync enqueues a sync job for one developer (?id=<uuid>) or all developers.
func (h *SyncHandler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if id := r.URL.Query().Get("id"); id != "" {
		dev, err := h.devRepo.GetByID(ctx, id)
		if err != nil {
			writeError(w, http.StatusNotFound, "developer not found")
			return
		}
		task, _ := worker.NewSyncDeveloperTask(dev.ID, dev.GithubUsername)
		if _, err := h.client.Enqueue(task); err != nil {
			writeError(w, http.StatusInternalServerError, "could not enqueue sync")
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]string{"queued": dev.GithubUsername})
		return
	}

	devs, err := h.devRepo.List(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list developers")
		return
	}
	queued := 0
	for _, dev := range devs {
		task, _ := worker.NewSyncDeveloperTask(dev.ID, dev.GithubUsername)
		if _, err := h.client.Enqueue(task); err == nil {
			queued++
		}
	}
	writeJSON(w, http.StatusAccepted, map[string]int{"queued": queued})
}

// SyncStatus returns queue depth and GitHub rate-limit state.
func (h *SyncHandler) SyncStatus(w http.ResponseWriter, r *http.Request) {
	remaining, resetAt := h.ghClient.RateLimitState()
	writeJSON(w, http.StatusOK, map[string]any{
		"github_rate_limit": map[string]any{
			"remaining": remaining,
			"reset_at":  resetAt.UTC().Format(time.RFC3339),
		},
	})
}
