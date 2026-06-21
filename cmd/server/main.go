package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

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
	obsRepo := repository.NewObservabilityRepo(pool)
	ghClient := gh.NewClient(os.Getenv("GITHUB_TOKEN"))
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.RedisAddr()})
	defer asynqClient.Close()

	if err := seedAdmin(context.Background(), adminRepo, cfg); err != nil {
		slog.Error("admin seed failed", "err", err)
		os.Exit(1)
	}

	authH := handler.NewAuthHandler(adminRepo, obsRepo, cfg.JWTSecret, cfg.SecureCookies)
	devH := handler.NewDeveloperHandler(devRepo, asynqClient, obsRepo)
	syncH := handler.NewSyncHandler(devRepo, asynqClient, ghClient, obsRepo)
	pubH := handler.NewPublicHandler(statsRepo, redisCach)
	reqH := handler.NewProfileRequestHandler(requestRepo, devRepo, asynqClient, obsRepo)
	obsH := handler.NewObservabilityHandler(obsRepo)

	aiChat, err := ai.NewChatService(context.Background(), cfg, statsRepo, obsRepo)
	if err != nil {
		slog.Error("ai setup failed", "err", err)
		os.Exit(1)
	}
	aiSummary := ai.NewSummaryService(aiChat, pool, cfg.AIModel)
	aiProjectSummary := ai.NewProjectSummaryService(aiChat, pool, cfg.AIModel)
	aiRankInsight := ai.NewRankInsightService(aiChat, pool, statsRepo, cfg.AIModel)
	aiTags := ai.NewNormalizedTagsService(aiChat, pool, statsRepo, cfg.AIModel)
	aiCompare := ai.NewCompareService(aiChat, statsRepo, cfg.AIModel)
	aiH := handler.NewAIHandler(aiChat, aiSummary, aiProjectSummary, aiRankInsight, aiTags, aiCompare, statsRepo, pool, redisCach)

	mux := http.NewServeMux()
	public := http.NewServeMux()
	publicLimiter := middleware.NewRateLimiter(
		cfg.PublicRateLimit,
		cfg.PublicRateWindow,
	).Middleware(public)

	// Health
	public.HandleFunc("GET /health", handler.Health)

	// AI routes (public, rate-limited internally)
	public.HandleFunc("POST /api/v1/ai/chat", aiH.Chat)
	public.HandleFunc("GET /api/v1/ai/compare", aiH.GetDeveloperComparison)
	public.HandleFunc("GET /api/v1/developers/{username}/summary", aiH.GetDeveloperSummary)
	public.HandleFunc("GET /api/v1/developers/{username}/rank-insight", aiH.GetRankInsight)
	public.HandleFunc("GET /api/v1/developers/{username}/normalized-tags", aiH.GetDeveloperNormalizedTags)
	public.HandleFunc("GET /api/v1/repos/{id}/summary", aiH.GetProjectSummary)
	public.HandleFunc("GET /api/v1/repos/{id}/normalized-tags", aiH.GetProjectNormalizedTags)

	// Public API
	public.HandleFunc("GET /api/v1/developers", pubH.ListDevelopers)
	public.HandleFunc("GET /api/v1/developers/spotlight", pubH.GetSpotlight)
	public.HandleFunc("GET /api/v1/developers/{username}/wrapped", pubH.GetWrapped)
	public.HandleFunc("GET /api/v1/developers/{username}/repos", pubH.GetDeveloperRepos)
	public.HandleFunc("GET /api/v1/developers/{username}/contributions", pubH.GetDeveloperContributions)
	public.HandleFunc("GET /api/v1/developers/{username}/contribution-stats", pubH.GetDeveloperContributionStats)
	public.HandleFunc("GET /api/v1/developers/{username}", pubH.GetDeveloper)
	public.HandleFunc("GET /api/v1/leaderboard", pubH.GetLeaderboard)
	public.HandleFunc("GET /api/v1/projects/top", pubH.GetTopProjects)
	public.HandleFunc("GET /api/v1/projects/fastest-growing", pubH.GetFastestGrowingProjects)
	public.HandleFunc("GET /api/v1/repos/{id}/growth", pubH.GetRepoGrowth)
	public.HandleFunc("GET /api/v1/activity/recent", pubH.GetRecentActivity)
	public.HandleFunc("GET /api/v1/stats/overview", pubH.GetOverview)
	public.HandleFunc("GET /api/v1/stats/languages", pubH.GetLanguages)
	public.HandleFunc("GET /api/v1/stats/community-activity", pubH.GetCommunityActivity)
	public.HandleFunc("GET /api/v1/stats/open-source", pubH.GetOSSStats)
	public.HandleFunc("GET /api/v1/stats/innovation-graph", pubH.GetInnovationGraph)
	public.HandleFunc("GET /api/v1/stats/streak-summary", pubH.GetStreakSummary)
	public.HandleFunc("GET /api/v1/dev-of-month", pubH.GetDevOfMonth)

	public.HandleFunc("POST /api/v1/profile-requests", reqH.Submit)
	public.HandleFunc("GET /api/v1/profile-requests/check", reqH.CheckUsername)
	public.HandleFunc("POST /api/v1/admin/auth/login", authH.Login)
	public.HandleFunc("POST /api/v1/admin/auth/logout", authH.Logout)

	mux.Handle("/health", publicLimiter)
	mux.Handle("/api/v1/ai/", publicLimiter)
	mux.Handle("/api/v1/developers", publicLimiter)
	mux.Handle("/api/v1/developers/", publicLimiter)
	mux.Handle("/api/v1/leaderboard", publicLimiter)
	mux.Handle("/api/v1/projects/", publicLimiter)
	mux.Handle("/api/v1/repos/", publicLimiter)
	mux.Handle("/api/v1/activity/", publicLimiter)
	mux.Handle("/api/v1/stats/", publicLimiter)
	mux.Handle("/api/v1/dev-of-month", publicLimiter)
	mux.Handle("/api/v1/profile-requests", publicLimiter)
	mux.Handle("/api/v1/profile-requests/", publicLimiter)
	mux.Handle("/api/v1/admin/auth/", publicLimiter)

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
	protected.HandleFunc("GET /api/v1/admin/observability", obsH.GetOverview)
	protected.HandleFunc("GET /api/v1/admin/observability/logs", obsH.ListAuditLogs)
	protected.HandleFunc("GET /api/v1/admin/observability/agent-runs", obsH.ListAgentRuns)
	protected.HandleFunc("GET /api/v1/admin/observability/agent-events", obsH.ListAgentEvents)

	auth := middleware.Auth(cfg.JWTSecret)
	mux.Handle("/api/v1/admin/developers", auth(protected))
	mux.Handle("/api/v1/admin/developers/", auth(protected))
	mux.Handle("/api/v1/admin/sync", auth(protected))
	mux.Handle("/api/v1/admin/sync/", auth(protected))
	mux.Handle("/api/v1/admin/profile-requests", auth(protected))
	mux.Handle("/api/v1/admin/profile-requests/", auth(protected))
	mux.Handle("/api/v1/admin/observability", auth(protected))
	mux.Handle("/api/v1/admin/observability/", auth(protected))

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           middleware.AuditLogs(obsRepo, cfg.JWTSecret)(middleware.CORS(mux, cfg.AllowedCORSOrigins)),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	slog.Info("server starting", "port", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
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
