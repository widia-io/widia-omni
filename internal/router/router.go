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
	counterSvc := service.NewCounterService(db)
	entSvc := service.NewEntitlementService(db, rdb)
	areaSvc := service.NewAreaService(db, counterSvc)
	goalSvc := service.NewGoalService(db, counterSvc)
	habitSvc := service.NewHabitService(db, counterSvc)
	taskSvc := service.NewTaskService(db, counterSvc)
	billingSvc := service.NewBillingService(db, entSvc, cfg.StripeSecretKey, cfg.StripeWebhookSecret,
		cfg.AllowedOrigins[0]+"/billing/success", cfg.AllowedOrigins[0]+"/billing/cancel")
	onboardingSvc := service.NewOnboardingService(db)
	dashSvc := service.NewDashboardService(db)
	journalSvc := service.NewJournalService(db)
	scoreSvc := service.NewScoreService(db)
	notifSvc := service.NewNotificationService(db)
	auditSvc := service.NewAuditService(db)
	_ = auditSvc // used by workers and future middleware
	exportSvc := service.NewExportService(db)

	// Handlers
	healthH := handler.NewHealthHandler(db, rdb)
	authH := handler.NewAuthHandler(authSvc)
	userH := handler.NewUserHandler(userSvc)
	wsH := handler.NewWorkspaceHandler(wsSvc)
	areaH := handler.NewAreaHandler(areaSvc)
	goalH := handler.NewGoalHandler(goalSvc)
	habitH := handler.NewHabitHandler(habitSvc)
	taskH := handler.NewTaskHandler(taskSvc)
	billingH := handler.NewBillingHandler(billingSvc)
	webhookH := handler.NewStripeWebhookHandler(billingSvc)
	onboardingH := handler.NewOnboardingHandler(onboardingSvc)
	dashH := handler.NewDashboardHandler(dashSvc)
	journalH := handler.NewJournalHandler(journalSvc)
	scoreH := handler.NewScoreHandler(scoreSvc)
	notifH := handler.NewNotificationHandler(notifSvc)
	exportH := handler.NewExportHandler(exportSvc)

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

	// Stripe webhook (no auth, signature verified in handler)
	r.Post("/webhooks/stripe", webhookH.Handle)

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
		r.Post("/me/export", exportH.Export)

		// Workspace
		r.Get("/workspace", wsH.GetWorkspace)
		r.Put("/workspace", wsH.UpdateWorkspace)
		r.Get("/workspace/usage", wsH.GetUsage)

		// Areas
		r.Route("/areas", func(r chi.Router) {
			r.Get("/", areaH.List)
			r.Post("/", areaH.Create)
			r.Put("/{id}", areaH.Update)
			r.Delete("/{id}", areaH.Delete)
			r.Patch("/{id}/reorder", areaH.Reorder)
		})

		// Goals
		r.Route("/goals", func(r chi.Router) {
			r.Get("/", goalH.List)
			r.Post("/", goalH.Create)
			r.Get("/{id}", goalH.GetByID)
			r.Put("/{id}", goalH.Update)
			r.Delete("/{id}", goalH.Delete)
			r.Patch("/{id}/progress", goalH.UpdateProgress)
		})

		// Habits
		r.Route("/habits", func(r chi.Router) {
			r.Get("/", habitH.List)
			r.Post("/", habitH.Create)
			r.Put("/{id}", habitH.Update)
			r.Delete("/{id}", habitH.Delete)
			r.Post("/{id}/check-in", habitH.CheckIn)
			r.Delete("/{id}/check-in", habitH.DeleteCheckIn)
			r.Get("/entries", habitH.ListEntries)
			r.Get("/streaks", habitH.GetStreaks)
		})

		// Tasks
		r.Route("/tasks", func(r chi.Router) {
			r.Get("/", taskH.List)
			r.Post("/", taskH.Create)
			r.Put("/{id}", taskH.Update)
			r.Delete("/{id}", taskH.Delete)
			r.Patch("/{id}/complete", taskH.Complete)
			r.Patch("/{id}/focus", taskH.ToggleFocus)
		})

		// Billing
		r.Route("/billing", func(r chi.Router) {
			r.Get("/plans", billingH.ListPlans)
			r.Get("/subscription", billingH.GetSubscription)
			r.Post("/checkout", billingH.CreateCheckout)
			r.Post("/portal", billingH.CreatePortal)
		})

		// Onboarding
		r.Route("/onboarding", func(r chi.Router) {
			r.Get("/status", onboardingH.GetStatus)
			r.Post("/areas", onboardingH.SetupAreas)
			r.Post("/goals", onboardingH.SetupGoals)
			r.Post("/habits", onboardingH.SetupHabits)
			r.Post("/complete", onboardingH.Complete)
		})

		// Journal
		r.Route("/journal", func(r chi.Router) {
			r.Get("/", journalH.List)
			r.Get("/{date}", journalH.Get)
			r.Put("/{date}", journalH.Upsert)
			r.Delete("/{date}", journalH.Delete)
		})

		// Scores
		r.Route("/scores", func(r chi.Router) {
			r.Get("/history", scoreH.GetHistory)
			r.Get("/current", scoreH.GetCurrent)
		})

		// Notifications
		r.Route("/notifications", func(r chi.Router) {
			r.Get("/", notifH.List)
			r.Patch("/{id}/read", notifH.MarkRead)
			r.Patch("/read-all", notifH.MarkAllRead)
			r.Get("/unread-count", notifH.UnreadCount)
		})

		// Dashboard
		r.Get("/dashboard", dashH.GetDashboard)
	})

	return r
}
