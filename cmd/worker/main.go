package main

import (
	"log/slog"
	"os"

	"github.com/abdulsami/nust-devs/internal/config"
)

func main() {
	cfg := config.Load()
	_ = cfg

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("worker starting")
	// Asynq worker setup comes in M3
	select {}
}
