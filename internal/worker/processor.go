package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/abdulsami/nust-devs/internal/models"
	"github.com/abdulsami/nust-devs/internal/service"
	"github.com/hibiken/asynq"
)

type SyncProcessor struct {
	svc *service.SyncService
}

func NewSyncProcessor(svc *service.SyncService) *SyncProcessor {
	return &SyncProcessor{svc: svc}
}

func (p *SyncProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload SyncDeveloperPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	slog.Info("processing sync task", "developer", payload.GithubUsername)

	dev := &models.Developer{
		ID:             payload.DeveloperID,
		GithubUsername: payload.GithubUsername,
	}
	return p.svc.SyncDeveloper(ctx, dev)
}
