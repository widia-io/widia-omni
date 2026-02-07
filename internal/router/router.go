package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/widia-io/widia-omni/internal/config"
	"github.com/widia-io/widia-omni/internal/handler"
	"github.com/widia-io/widia-omni/internal/middleware"
	"github.com/widia-io/widia-omni/internal/observability"
	"github.com/widia-io/widia-omni/internal/service"
)

func New(cfg *config.Config, logger zerolog.Logger, db *pgxpool.Pool, rdb *redis.Client) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(logger))
	r.Use(middleware.CORS(cfg.AllowedOrigins))
	r.Use(observability.MetricsMiddleware)

	// Services
	authSvc := service.NewAuthService(cfg.SupabaseURL, cfg.SupabaseServiceKey)
	userSvc := service.NewUserService(db)
	wsSvc := service.NewWorkspaceService(db)

	// Handlers
	healthH := handler.NewHealthHandler(db, rdb)
	authH := handler.NewAuthHandler(authSvc)
	userH := handler.NewUserHandler(userSvc)
	wsH := handler.NewWorkspaceHandler(wsSvc)

	// Public routes
	r.Get("/health", healthH.Health)
	r.Get("/ready", healthH.Ready)
	r.Handle("/metrics", promhttp.Handler())

	// Auth routes (no auth required)
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authH.Register)
		r.Post("/login", authH.Login)
		r.Post("/refresh", authH.Refresh)
		r.Post("/logout", authH.Logout)
		r.Post("/forgot-password", authH.ForgotPassword)
		r.Post("/reset-password", authH.ResetPassword)
		r.Post("/verify-email", authH.VerifyEmail)
	})

	// Protected API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.Auth(cfg.SupabaseJWTSecret))
		r.Use(middleware.Tenant(db, rdb))
		r.Use(middleware.RateLimit(rdb, 60))

		// User
		r.Get("/me", userH.GetProfile)
		r.Put("/me", userH.UpdateProfile)
		r.Delete("/me", userH.DeleteAccount)
		r.Get("/me/preferences", userH.GetPreferences)
		r.Put("/me/preferences", userH.UpdatePreferences)

		// Workspace
		r.Get("/workspace", wsH.GetWorkspace)
		r.Put("/workspace", wsH.UpdateWorkspace)
		r.Get("/workspace/usage", wsH.GetUsage)
	})

	return r
}
