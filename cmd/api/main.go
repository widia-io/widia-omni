package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/widia-io/widia-omni/internal/config"
	"github.com/widia-io/widia-omni/internal/observability"
	"github.com/widia-io/widia-omni/internal/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger := observability.NewLogger(cfg)
	logger.Info().Str("env", cfg.Env).Int("port", cfg.Port).Msg("starting api server")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to parse redis url")
	}
	rdb := redis.NewClient(opt)
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Warn().Err(err).Msg("redis not available, running without cache")
		rdb = nil
	} else {
		logger.Info().Msg("redis connected")
	}

	r := router.New(cfg, logger, db, rdb)

	// Background: refresh subscription gauge every 5min
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		observability.RefreshSubscriptionGauge(ctx, db, logger) // initial
		for {
			select {
			case <-ticker.C:
				observability.RefreshSubscriptionGauge(ctx, db, logger)
			case <-ctx.Done():
				return
			}
		}
	}()

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("server error")
		}
	}()

	logger.Info().Msgf("server listening on :%d", cfg.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("shutting down server")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal().Err(err).Msg("server forced shutdown")
	}

	logger.Info().Msg("server stopped")
}
