package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/abdulsami/nust-devs/internal/cache"
	"github.com/abdulsami/nust-devs/internal/models"
	"github.com/abdulsami/nust-devs/internal/service"
	"github.com/hibiken/asynq"
)

type SyncProcessor struct {
	svc   *service.SyncService
	cache *cache.Cache
}

func NewSyncProcessor(svc *service.SyncService, c *cache.Cache) *SyncProcessor {
	return &SyncProcessor{svc: svc, cache: c}
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
	if err := p.svc.SyncDeveloper(ctx, dev); err != nil {
		slog.Error("sync failed", "developer", payload.GithubUsername, "err", err)
		return err
	}
	if p.cache != nil {
		if err := p.cache.InvalidatePublic(ctx); err != nil {
			slog.Warn("cache invalidation failed", "err", err)
		}
	}
	return nil
}
