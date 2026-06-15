package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/abdulsami/nust-devs/internal/ai"
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

	if err := cfg.ValidateServer(); err != nil {
		slog.Error("invalid server configuration", "err", err)
		os.Exit(1)
	}

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
	requestRepo := repository.NewRequestRepo(pool)
	adminRepo := repository.NewAdminRepo(pool)
	devRepo := repository.NewDeveloperRepo(pool)
	statsRepo := repository.NewStatsRepo(pool)
	ghClient := gh.NewClient(os.Getenv("GITHUB_TOKEN"))
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.RedisAddr()})
	defer asynqClient.Close()

	if err := seedAdmin(context.Background(), adminRepo, cfg); err != nil {
		slog.Error("admin seed failed", "err", err)
		os.Exit(1)
	}

	authH := handler.NewAuthHandler(adminRepo, cfg.JWTSecret)
	devH := handler.NewDeveloperHandler(devRepo, asynqClient)
	syncH := handler.NewSyncHandler(devRepo, asynqClient, ghClient)
	pubH := handler.NewPublicHandler(statsRepo, redisCach)
	reqH := handler.NewProfileRequestHandler(requestRepo, devRepo, asynqClient)

	aiChat, err := ai.NewChatService(context.Background(), cfg, statsRepo)
	if err != nil {
		slog.Error("ai setup failed", "err", err)
		os.Exit(1)
	}
	aiSummary := ai.NewSummaryService(aiChat, pool, cfg.AIModel)
	aiH := handler.NewAIHandler(aiChat, aiSummary, statsRepo, pool)

	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("GET /health", handler.Health)

	// AI routes (public, rate-limited internally)
	mux.HandleFunc("POST /api/v1/ai/chat", aiH.Chat)
	mux.HandleFunc("GET /api/v1/developers/{username}/summary", aiH.GetDeveloperSummary)

	// Public API
	mux.HandleFunc("GET /api/v1/developers", pubH.ListDevelopers)
	mux.HandleFunc("GET /api/v1/developers/spotlight", pubH.GetSpotlight)
	mux.HandleFunc("GET /api/v1/developers/{username}/wrapped", pubH.GetWrapped)
	mux.HandleFunc("GET /api/v1/developers/{username}/repos", pubH.GetDeveloperRepos)
	mux.HandleFunc("GET /api/v1/developers/{username}/contributions", pubH.GetDeveloperContributions)
	mux.HandleFunc("GET /api/v1/developers/{username}/contribution-stats", pubH.GetDeveloperContributionStats)
	mux.HandleFunc("GET /api/v1/developers/{username}", pubH.GetDeveloper)
	mux.HandleFunc("GET /api/v1/leaderboard", pubH.GetLeaderboard)
	mux.HandleFunc("GET /api/v1/projects/top", pubH.GetTopProjects)
	mux.HandleFunc("GET /api/v1/projects/fastest-growing", pubH.GetFastestGrowingProjects)
	mux.HandleFunc("GET /api/v1/repos/{id}/growth", pubH.GetRepoGrowth)
	mux.HandleFunc("GET /api/v1/activity/recent", pubH.GetRecentActivity)
	mux.HandleFunc("GET /api/v1/stats/overview", pubH.GetOverview)
	mux.HandleFunc("GET /api/v1/stats/languages", pubH.GetLanguages)
	mux.HandleFunc("GET /api/v1/stats/community-activity", pubH.GetCommunityActivity)
	mux.HandleFunc("GET /api/v1/stats/open-source", pubH.GetOSSStats)
	mux.HandleFunc("GET /api/v1/stats/innovation-graph", pubH.GetInnovationGraph)
	mux.HandleFunc("GET /api/v1/stats/streak-summary", pubH.GetStreakSummary)
	mux.HandleFunc("GET /api/v1/dev-of-month", pubH.GetDevOfMonth)

	mux.HandleFunc("POST /api/v1/profile-requests", reqH.Submit)
	mux.HandleFunc("GET /api/v1/profile-requests/check", reqH.CheckUsername)

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
	protected.HandleFunc("GET /api/v1/admin/profile-requests", reqH.List)
	protected.HandleFunc("POST /api/v1/admin/profile-requests/{id}/approve", reqH.Approve)
	protected.HandleFunc("POST /api/v1/admin/profile-requests/{id}/reject", reqH.Reject)

	auth := middleware.Auth(cfg.JWTSecret)
	mux.Handle("/api/v1/admin/developers", auth(protected))
	mux.Handle("/api/v1/admin/developers/", auth(protected))
	mux.Handle("/api/v1/admin/sync", auth(protected))
	mux.Handle("/api/v1/admin/sync/", auth(protected))
	mux.Handle("/api/v1/admin/profile-requests", auth(protected))
	mux.Handle("/api/v1/admin/profile-requests/", auth(protected))

	slog.Info("server starting", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, middleware.CORS(mux)); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func seedAdmin(ctx context.Context, repo *repository.AdminRepo, cfg *config.Config) error {
	n, err := repo.Count(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

	if cfg.AdminEmail == "" || cfg.AdminPassword == "" {
		return errors.New("no admin users exist; set ADMIN_EMAIL and ADMIN_PASSWORD to seed the first admin")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if _, err := repo.Create(ctx, cfg.AdminEmail, string(hash)); err != nil {
		return err
	}
	slog.Info("seeded admin user", "email", cfg.AdminEmail)
	return nil
}
