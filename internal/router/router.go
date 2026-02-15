package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/widia-io/widia-omni/internal/config"
	"github.com/widia-io/widia-omni/internal/handler"
	"github.com/widia-io/widia-omni/internal/llm"
	"github.com/widia-io/widia-omni/internal/middleware"
	"github.com/widia-io/widia-omni/internal/observability"
	"github.com/widia-io/widia-omni/internal/service"
)

func New(cfg *config.Config, logger zerolog.Logger, db *pgxpool.Pool, rdb *redis.Client) *chi.Mux {
	r := chi.NewRouter()
	appURL := cfg.AppURL

	var queueClient *asynq.Client
	if opt, err := redis.ParseURL(cfg.RedisURL); err == nil {
		queueClient = asynq.NewClient(asynq.RedisClientOpt{
			Addr:     opt.Addr,
			Password: opt.Password,
			DB:       opt.DB,
		})
	} else {
		logger.Warn().Err(err).Msg("failed to parse redis url for asynq client")
	}

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(logger))
	r.Use(middleware.CORS(cfg.AllowedOrigins))
	r.Use(observability.MetricsMiddleware)

	// Services
	authSvc := service.NewAuthService(cfg.SupabaseURL, cfg.SupabaseServiceKey)
	userSvc := service.NewUserService(db)
	wsSvc := service.NewWorkspaceService(db, rdb, queueClient, appURL, logger)
	counterSvc := service.NewCounterService(db)
	entSvc := service.NewEntitlementService(db, rdb)
	referralSvc := service.NewReferralService(db, appURL)
	areaSvc := service.NewAreaService(db, counterSvc)
	goalSvc := service.NewGoalService(db, counterSvc)
	habitSvc := service.NewHabitService(db, counterSvc)
	taskSvc := service.NewTaskService(db, counterSvc)
	billingSvc := service.NewBillingService(db, entSvc, referralSvc, cfg.StripeSecretKey, cfg.StripeWebhookSecret,
		appURL+"/billing/success", appURL+"/billing/cancel")
	onboardingSvc := service.NewOnboardingService(db)
	dashSvc := service.NewDashboardService(db, rdb)
	journalSvc := service.NewJournalService(db)
	scoreSvc := service.NewScoreService(db, rdb)
	notifSvc := service.NewNotificationService(db)
	auditSvc := service.NewAuditService(db)
	_ = auditSvc // used by workers and future middleware
	exportSvc := service.NewExportService(db)
	financeSvc := service.NewFinanceService(db, counterSvc)
	llmClient := llm.NewClient(cfg.OpenRouterAPIKey, cfg.OpenRouterModel)
	insightSvc := service.NewInsightService(db, rdb, llmClient)
	projectSvc := service.NewProjectService(db, counterSvc)
	labelSvc := service.NewLabelService(db)
	sectionSvc := service.NewSectionService(db)
	apiKeySvc := service.NewAPIKeyService(db, rdb)
	adminSvc := service.NewAdminService(db, entSvc)

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
	financeH := handler.NewFinanceHandler(financeSvc)
	insightH := handler.NewInsightHandler(insightSvc)
	projectH := handler.NewProjectHandler(projectSvc)
	labelH := handler.NewLabelHandler(labelSvc)
	sectionH := handler.NewSectionHandler(sectionSvc)
	apiKeyH := handler.NewAPIKeyHandler(apiKeySvc)
	referralH := handler.NewReferralHandler(referralSvc)
	adminH := handler.NewAdminHandler(adminSvc)

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
		r.Get("/workspaces", wsH.ListWorkspaces)
		r.Get("/workspace", wsH.GetWorkspace)
		r.Put("/workspace", wsH.UpdateWorkspace)
		r.Post("/workspace/switch", wsH.SwitchWorkspace)
		r.Get("/workspace/members", wsH.ListMembers)
		r.Delete("/workspace/members/{userID}", wsH.RemoveMember)
		r.Post("/workspace/invites", wsH.CreateInvite)
		r.Get("/workspace/invites", wsH.ListInvites)
		r.Post("/workspace/invites/accept", wsH.AcceptInvite)
		r.Delete("/workspace/invites/{id}", wsH.RevokeInvite)
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

		// Labels
		r.Route("/labels", func(r chi.Router) {
			r.Get("/", labelH.List)
			r.Post("/", labelH.Create)
			r.Put("/{id}", labelH.Update)
			r.Delete("/{id}", labelH.Delete)
		})

		// Sections
		r.Route("/sections", func(r chi.Router) {
			r.Get("/", sectionH.List)
			r.Post("/", sectionH.Create)
			r.Put("/{id}", sectionH.Update)
			r.Delete("/{id}", sectionH.Delete)
			r.Patch("/{id}/reorder", sectionH.Reorder)
		})

		// Projects
		r.Route("/projects", func(r chi.Router) {
			r.Get("/", projectH.List)
			r.Post("/", projectH.Create)
			r.Get("/{id}", projectH.GetByID)
			r.Put("/{id}", projectH.Update)
			r.Delete("/{id}", projectH.Delete)
			r.Patch("/{id}/reorder", projectH.Reorder)
			r.Patch("/{id}/archive", projectH.Archive)
			r.Patch("/{id}/unarchive", projectH.Unarchive)
			r.Get("/{id}/sections", projectH.ListSections)
			r.Post("/{id}/sections", projectH.CreateSection)
			r.Put("/{id}/sections/{sectionId}", projectH.UpdateSection)
			r.Delete("/{id}/sections/{sectionId}", projectH.DeleteSection)
			r.Patch("/{id}/sections/{sectionId}/reorder", projectH.ReorderSection)
		})

		// Tasks
		r.Route("/tasks", func(r chi.Router) {
			r.Get("/", taskH.List)
			r.Post("/", taskH.Create)
			r.Put("/{id}", taskH.Update)
			r.Delete("/{id}", taskH.Delete)
			r.Patch("/{id}/complete", taskH.Complete)
			r.Patch("/{id}/reopen", taskH.Reopen)
			r.Patch("/{id}/focus", taskH.ToggleFocus)
			r.Patch("/{id}/reorder", taskH.Reorder)
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
			r.Get("/area-templates", onboardingH.GetAreaTemplates)
			r.Get("/goal-suggestions", onboardingH.GetGoalSuggestions)
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

		// Finances
		r.Route("/finances", func(r chi.Router) {
			r.Get("/summary", financeH.GetSummary)
			r.Get("/transactions", financeH.ListTransactions)
			r.Post("/transactions", financeH.CreateTransaction)
			r.Put("/transactions/{id}", financeH.UpdateTransaction)
			r.Delete("/transactions/{id}", financeH.DeleteTransaction)
			r.Get("/categories", financeH.ListCategories)
			r.Post("/categories", financeH.CreateCategory)
			r.Put("/categories/{id}", financeH.UpdateCategory)
			r.Delete("/categories/{id}", financeH.DeleteCategory)
			r.Get("/budgets", financeH.ListBudgets)
			r.Post("/budgets", financeH.UpsertBudget)
			r.Delete("/budgets/{id}", financeH.DeleteBudget)
		})

		// Insights
		r.Route("/insights", func(r chi.Router) {
			r.Get("/", insightH.List)
			r.Get("/latest", insightH.GetLatest)
			r.Post("/generate", insightH.Generate)
		})

		// API Keys
		r.Route("/api-keys", func(r chi.Router) {
			r.Get("/", apiKeyH.List)
			r.Post("/", apiKeyH.Create)
			r.Delete("/{id}", apiKeyH.Revoke)
		})

		// Referrals
		r.Route("/referrals", func(r chi.Router) {
			r.Get("/me", referralH.GetMe)
			r.Post("/regenerate", referralH.Regenerate)
			r.Get("/attributions", referralH.ListAttributions)
			r.Get("/credits", referralH.ListCredits)
		})

		// Dashboard
		r.Get("/dashboard", dashH.GetDashboard)
	})

	// Public API — API key auth, read-only, reuses existing handlers
	r.Route("/public/v1", func(r chi.Router) {
		r.Use(middleware.APIKeyAuth(apiKeySvc))
		r.Use(middleware.RateLimit(rdb, 60))

		r.Get("/areas", areaH.List)
		r.Get("/goals", goalH.List)
		r.Get("/goals/{id}", goalH.GetByID)
		r.Get("/habits", habitH.List)
		r.Get("/habits/entries", habitH.ListEntries)
		r.Get("/habits/streaks", habitH.GetStreaks)
		r.Get("/labels", labelH.List)
		r.Get("/sections", sectionH.List)
		r.Get("/projects", projectH.List)
		r.Get("/projects/{id}", projectH.GetByID)
		r.Get("/tasks", taskH.List)
		r.Get("/scores/current", scoreH.GetCurrent)
		r.Get("/scores/history", scoreH.GetHistory)
		r.Get("/journal", journalH.List)
		r.Get("/journal/{date}", journalH.Get)
		r.Get("/finances/summary", financeH.GetSummary)
		r.Get("/finances/transactions", financeH.ListTransactions)
		r.Get("/insights", insightH.List)
		r.Get("/insights/latest", insightH.GetLatest)
	})

	// Admin routes (service key auth)
	r.Route("/admin", func(r chi.Router) {
		r.Use(middleware.AdminAuth(cfg.SupabaseServiceKey))
		r.Get("/metrics", adminH.GetMetrics)
		r.Get("/users", adminH.ListUsers)
		r.Get("/users/{id}", adminH.GetUser)
		r.Get("/workspaces/{id}/usage", adminH.GetWorkspaceUsage)
		r.Post("/entitlements/override", adminH.OverrideEntitlement)
	})

	return r
}
