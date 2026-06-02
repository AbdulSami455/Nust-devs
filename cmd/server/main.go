package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/abdulsami/nust-devs/internal/cache"
	"github.com/abdulsami/nust-devs/internal/config"
	"github.com/abdulsami/nust-devs/internal/db"
	gh "github.com/abdulsami/nust-devs/internal/github"
	"github.com/abdulsami/nust-devs/internal/handler"
	"github.com/abdulsami/nust-devs/internal/middleware"
	"github.com/abdulsami/nust-devs/internal/repository"
	"github.com/hibiken/asynq"
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

	redisCach := cache.New(cfg.RedisAddr())
	adminRepo := repository.NewAdminRepo(pool)
	devRepo := repository.NewDeveloperRepo(pool)
	statsRepo := repository.NewStatsRepo(pool)
	ghClient := gh.NewClient(os.Getenv("GITHUB_TOKEN"))
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.RedisAddr()})
	defer asynqClient.Close()

	seedAdmin(context.Background(), adminRepo, cfg)

	authH := handler.NewAuthHandler(adminRepo, cfg.JWTSecret)
	devH := handler.NewDeveloperHandler(devRepo)
	syncH := handler.NewSyncHandler(devRepo, asynqClient, ghClient)
	pubH := handler.NewPublicHandler(statsRepo, redisCach)

	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("GET /health", handler.Health)

	// Public API
	mux.HandleFunc("GET /api/v1/developers", pubH.ListDevelopers)
	mux.HandleFunc("GET /api/v1/developers/spotlight", pubH.GetSpotlight)
	mux.HandleFunc("GET /api/v1/developers/{username}", pubH.GetDeveloper)
	mux.HandleFunc("GET /api/v1/developers/{username}/repos", pubH.GetDeveloperRepos)
	mux.HandleFunc("GET /api/v1/developers/{username}/contributions", pubH.GetDeveloperContributions)
	mux.HandleFunc("GET /api/v1/leaderboard", pubH.GetLeaderboard)
	mux.HandleFunc("GET /api/v1/projects/top", pubH.GetTopProjects)
	mux.HandleFunc("GET /api/v1/activity/recent", pubH.GetRecentActivity)
	mux.HandleFunc("GET /api/v1/stats/overview", pubH.GetOverview)
	mux.HandleFunc("GET /api/v1/stats/languages", pubH.GetLanguages)
	mux.HandleFunc("GET /api/v1/stats/community-activity", pubH.GetCommunityActivity)
	mux.HandleFunc("GET /api/v1/stats/open-source", pubH.GetOSSStats)

	// Admin auth (public)
	mux.HandleFunc("POST /api/v1/admin/auth/login", authH.Login)

	// Admin protected
	protected := http.NewServeMux()
	protected.HandleFunc("POST /api/v1/admin/developers", devH.Create)
	protected.HandleFunc("GET /api/v1/admin/developers", devH.List)
	protected.HandleFunc("PATCH /api/v1/admin/developers/{id}", devH.Update)
	protected.HandleFunc("DELETE /api/v1/admin/developers/{id}", devH.Delete)
	protected.HandleFunc("POST /api/v1/admin/sync", syncH.TriggerSync)
	protected.HandleFunc("GET /api/v1/admin/sync/status", syncH.SyncStatus)

	auth := middleware.Auth(cfg.JWTSecret)
	mux.Handle("/api/v1/admin/developers", auth(protected))
	mux.Handle("/api/v1/admin/developers/", auth(protected))
	mux.Handle("/api/v1/admin/sync", auth(protected))
	mux.Handle("/api/v1/admin/sync/", auth(protected))

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
