package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/widia-io/widia-omni/internal/config"
	"github.com/widia-io/widia-omni/internal/email"
	"github.com/widia-io/widia-omni/internal/llm"
	"github.com/widia-io/widia-omni/internal/observability"
	"github.com/widia-io/widia-omni/internal/service"
	"github.com/widia-io/widia-omni/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger := observability.NewLogger(cfg)
	logger.Info().Str("env", cfg.Env).Msg("starting worker")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Database
	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to parse database url")
	}
	poolCfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "SET search_path TO widia_omni")
		return err
	}
	db, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		logger.Fatal().Err(err).Msg("failed to ping database")
	}
	logger.Info().Msg("database connected")

	// Redis
	redisOpt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to parse redis url")
	}

	// Redis client for caching
	rdb := redis.NewClient(redisOpt)
	defer rdb.Close()

	// Services
	scoreSvc := service.NewScoreService(db, rdb)
	notifSvc := service.NewNotificationService(db)
	entSvc := service.NewEntitlementService(db, rdb)
	llmClient := llm.NewClient(cfg.OpenRouterAPIKey, cfg.OpenRouterModel)
	insightSvc := service.NewInsightService(db, rdb, llmClient)
	emailSender := email.NewLogSender(logger)

	// Task handlers
	scoreH := worker.NewScoreSnapshotHandler(db, scoreSvc, notifSvc, logger)
	streakH := worker.NewStreakUpdateHandler(db, notifSvc, logger)
	counterH := worker.NewCounterReconcilerHandler(db, logger)
	notifH := worker.NewSendNotificationHandler(emailSender, logger)
	insightGenH := worker.NewInsightGenerateHandler(db, insightSvc, entSvc, notifSvc, logger)

	// Asynq server
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisOpt.Addr, Password: redisOpt.Password, DB: redisOpt.DB},
		asynq.Config{
			Concurrency: 5,
			Queues:      map[string]int{"default": 3, "critical": 6, "low": 1},
			Logger:      &asynqLogger{logger: logger},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(worker.TypeScoreSnapshot, scoreH.ProcessTask)
	mux.HandleFunc(worker.TypeStreakUpdate, streakH.ProcessTask)
	mux.HandleFunc(worker.TypeCounterReconcile, counterH.ProcessTask)
	mux.HandleFunc(worker.TypeSendNotification, notifH.ProcessTask)
	mux.HandleFunc(worker.TypeInsightGenerate, insightGenH.ProcessTask)

	// Scheduler
	scheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{Addr: redisOpt.Addr, Password: redisOpt.Password, DB: redisOpt.DB},
		nil,
	)

	scheduler.Register("0 2 * * 1", worker.NewTask(worker.TypeScoreSnapshot, nil))    // Monday 2am
	scheduler.Register("0 1 * * *", worker.NewTask(worker.TypeStreakUpdate, nil))       // Daily 1am
	scheduler.Register("0 * * * *", worker.NewTask(worker.TypeCounterReconcile, nil))  // Hourly
	scheduler.Register("0 3 * * 1", worker.NewTask(worker.TypeInsightGenerate, nil))  // Monday 3am

	// Start
	go func() {
		if err := srv.Start(mux); err != nil {
			logger.Fatal().Err(err).Msg("asynq server error")
		}
	}()

	go func() {
		if err := scheduler.Start(); err != nil {
			logger.Fatal().Err(err).Msg("asynq scheduler error")
		}
	}()

	// Metrics HTTP server on :9090
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsSrv := &http.Server{Addr: ":9090", Handler: metricsMux}
	go func() {
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("metrics server error")
		}
	}()

	// Poll asynq queue depths every 30s
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: redisOpt.Addr, Password: redisOpt.Password, DB: redisOpt.DB})
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				for _, q := range []string{"default", "critical", "low"} {
					info, err := inspector.GetQueueInfo(q)
					if err == nil {
						observability.AsynqQueueDepth.WithLabelValues(q).Set(float64(info.Pending))
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	logger.Info().Msg("worker running with scheduled tasks")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("shutting down worker")
	cancel()
	srv.Shutdown()
	scheduler.Shutdown()
	inspector.Close()
	metricsSrv.Close()
	logger.Info().Msg("worker stopped")
}

// asynqLogger adapts zerolog to asynq's logger interface
type asynqLogger struct {
	logger zerolog.Logger
}

func (l *asynqLogger) Debug(args ...any) { l.logger.Debug().Msgf("%v", args) }
func (l *asynqLogger) Info(args ...any)  { l.logger.Info().Msgf("%v", args) }
func (l *asynqLogger) Warn(args ...any)  { l.logger.Warn().Msgf("%v", args) }
func (l *asynqLogger) Error(args ...any) { l.logger.Error().Msgf("%v", args) }
func (l *asynqLogger) Fatal(args ...any) { l.logger.Fatal().Msgf("%v", args) }
