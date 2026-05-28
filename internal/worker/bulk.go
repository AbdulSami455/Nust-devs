package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/abdulsami/nust-devs/internal/repository"
	"github.com/hibiken/asynq"
)

// BulkSyncProcessor handles sync:all and sync:active tasks by enqueuing
// individual sync:developer tasks for each matching developer.
type BulkSyncProcessor struct {
	devRepo *repository.DeveloperRepo
	client  *asynq.Client
}

func NewBulkSyncProcessor(devRepo *repository.DeveloperRepo, client *asynq.Client) *BulkSyncProcessor {
	return &BulkSyncProcessor{devRepo: devRepo, client: client}
}

func (p *BulkSyncProcessor) ProcessSyncAll(ctx context.Context, _ *asynq.Task) error {
	return p.enqueue(ctx, 0)
}

func (p *BulkSyncProcessor) ProcessSyncActive(ctx context.Context, _ *asynq.Task) error {
	// Only enqueue developers synced within the last 7 days (active tier).
	return p.enqueue(ctx, 7*24*time.Hour)
}

func (p *BulkSyncProcessor) enqueue(ctx context.Context, maxAge time.Duration) error {
	devs, err := p.devRepo.List(ctx)
	if err != nil {
		return err
	}
	queued := 0
	for _, dev := range devs {
		if maxAge > 0 && dev.LastSyncedAt != nil {
			if time.Since(*dev.LastSyncedAt) > maxAge {
				continue
			}
		}
		task, _ := NewSyncDeveloperTask(dev.ID, dev.GithubUsername)
		if _, err := p.client.Enqueue(task, asynq.Queue("default")); err != nil {
			slog.Warn("failed to enqueue developer sync", "dev", dev.GithubUsername, "err", err)
			continue
		}
		queued++
	}
	slog.Info("bulk sync enqueued", "queued", queued, "total", len(devs))
	return nil
}
