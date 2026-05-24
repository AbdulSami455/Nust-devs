package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/abdulsami/nust-devs/internal/config"
	"github.com/abdulsami/nust-devs/internal/db"
	"github.com/abdulsami/nust-devs/internal/handler"
	"github.com/abdulsami/nust-devs/internal/middleware"
	"github.com/abdulsami/nust-devs/internal/repository"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := db.RunMigrations(cfg.DBUrl); err != nil {
		slog.Error("migrations failed", "err", err)
		os.Exit(1)
	}

	pool, err := db.Connect(context.Background(), cfg.DBUrl)
	if err != nil {
		slog.Error("db connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	adminRepo := repository.NewAdminRepo(pool)
	seedAdmin(context.Background(), adminRepo, cfg)

	devRepo := repository.NewDeveloperRepo(pool)
	authH := handler.NewAuthHandler(adminRepo, cfg.JWTSecret)
	devH := handler.NewDeveloperHandler(devRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("POST /api/v1/admin/auth/login", authH.Login)

	protected := http.NewServeMux()
	protected.HandleFunc("POST /api/v1/admin/developers", devH.Create)
	protected.HandleFunc("GET /api/v1/admin/developers", devH.List)
	protected.HandleFunc("PATCH /api/v1/admin/developers/{id}", devH.Update)
	protected.HandleFunc("DELETE /api/v1/admin/developers/{id}", devH.Delete)

	mux.Handle("/api/v1/admin/developers", middleware.Auth(cfg.JWTSecret)(protected))
	mux.Handle("/api/v1/admin/developers/", middleware.Auth(cfg.JWTSecret)(protected))

	slog.Info("server starting", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, middleware.CORS(mux)); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func seedAdmin(ctx context.Context, repo *repository.AdminRepo, cfg *config.Config) {
	n, err := repo.Count(ctx)
	if err != nil || n > 0 {
		return
	}
	email := "admin@nust.edu.pk"
	password := "admin123"
	if v := os.Getenv("ADMIN_EMAIL"); v != "" {
		email = v
	}
	if v := os.Getenv("ADMIN_PASSWORD"); v != "" {
		password = v
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("seed admin: bcrypt failed", "err", err)
		return
	}
	if _, err := repo.Create(ctx, email, string(hash)); err != nil {
		slog.Error("seed admin: create failed", "err", err)
		return
	}
	slog.Info("seeded admin user", "email", email)
}
