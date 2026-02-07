package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/widia-io/widia-omni/internal/config"
	"github.com/widia-io/widia-omni/internal/observability"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger := observability.NewLogger(cfg)
	logger.Info().Str("env", cfg.Env).Msg("worker started, no tasks registered")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("worker stopped")
}
