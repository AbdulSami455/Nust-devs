package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/abdulsami/nust-devs/internal/config"
	"github.com/abdulsami/nust-devs/internal/db"
	gh "github.com/abdulsami/nust-devs/internal/github"
	"github.com/abdulsami/nust-devs/internal/repository"
	"github.com/abdulsami/nust-devs/internal/service"
	"github.com/abdulsami/nust-devs/internal/worker"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	pool, err := db.Connect(context.Background(), cfg.DBUrl)
	if err != nil {
		slog.Error("db connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	ghToken := os.Getenv("GITHUB_TOKEN")
	if ghToken == "" {
		slog.Error("GITHUB_TOKEN is required")
		os.Exit(1)
	}

	ghClient := gh.NewClient(ghToken)
	syncRepo := repository.NewSyncRepo(pool)
	devRepo := repository.NewDeveloperRepo(pool)
	svc := service.NewSyncService(ghClient, syncRepo, devRepo)
	processor := worker.NewSyncProcessor(svc)

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisAddr()},
		asynq.Config{
			Concurrency: 3,
			Queues:      map[string]int{"default": 1},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(worker.TaskSyncDeveloper, processor.ProcessTask)

	slog.Info("worker starting")
	if err := srv.Run(mux); err != nil {
		slog.Error("worker failed", "err", err)
		os.Exit(1)
	}
}
