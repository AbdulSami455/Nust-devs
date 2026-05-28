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

	redisOpt := asynq.RedisClientOpt{Addr: cfg.RedisAddr()}
	ghClient := gh.NewClient(ghToken)
	syncRepo := repository.NewSyncRepo(pool)
	devRepo := repository.NewDeveloperRepo(pool)
	svc := service.NewSyncService(ghClient, syncRepo, devRepo)

	asynqClient := asynq.NewClient(redisOpt)
	defer asynqClient.Close()

	syncProcessor := worker.NewSyncProcessor(svc)
	bulkProcessor := worker.NewBulkSyncProcessor(devRepo, asynqClient)

	// Asynq server — 3 concurrent sync slots to respect rate limits
	srv := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: 3,
		Queues:      map[string]int{"default": 1},
	})

	mux := asynq.NewServeMux()
	mux.HandleFunc(worker.TaskSyncDeveloper, syncProcessor.ProcessTask)
	mux.HandleFunc(worker.TaskSyncAll, bulkProcessor.ProcessSyncAll)
	mux.HandleFunc(worker.TaskSyncActive, bulkProcessor.ProcessSyncActive)

	// Staggered scheduling: full sync nightly, incremental every 6h
	scheduler := asynq.NewScheduler(redisOpt, nil)
	scheduler.Register("@midnight", asynq.NewTask(worker.TaskSyncAll, nil))
	scheduler.Register("0 */6 * * *", asynq.NewTask(worker.TaskSyncActive, nil))

	if err := scheduler.Start(); err != nil {
		slog.Error("scheduler failed to start", "err", err)
		os.Exit(1)
	}
	defer scheduler.Shutdown()

	slog.Info("worker starting")
	if err := srv.Run(mux); err != nil {
		slog.Error("worker failed", "err", err)
		os.Exit(1)
	}
}
