package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/abdulsami/nust-devs/internal/config"
	"github.com/abdulsami/nust-devs/internal/handler"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.Health)

	slog.Info("server starting", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}
