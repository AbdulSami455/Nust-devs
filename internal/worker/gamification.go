package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/abdulsami/nust-devs/internal/repository"
	"github.com/hibiken/asynq"
)

type GamificationProcessor struct {
	stats *repository.StatsRepo
}

func NewGamificationProcessor(stats *repository.StatsRepo) *GamificationProcessor {
	return &GamificationProcessor{stats: stats}
}

func (p *GamificationProcessor) ProcessDevOfMonth(ctx context.Context, _ *asynq.Task) error {
	prev := time.Now().UTC().AddDate(0, -1, 0)
	year := prev.Year()
	month := int(prev.Month())

	if err := p.stats.AwardDevOfMonth(ctx, year, month); err != nil {
		return err
	}
	slog.Info("dev of month awarded", "year", year, "month", month)
	return nil
}
