package observability

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/widia-io/widia-omni/internal/config"
)

type ctxKey string

const (
	loggerKey     ctxKey = "logger"
	requestIDKey  ctxKey = "request_id"
	userIDKey     ctxKey = "user_id"
	workspaceIDKey ctxKey = "workspace_id"
)

func NewLogger(cfg *config.Config) zerolog.Logger {
	level := zerolog.InfoLevel
	if cfg.Env == "development" {
		level = zerolog.DebugLevel
	}

	zerolog.SetGlobalLevel(level)

	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", "widia-omni").
		Str("env", cfg.Env).
		Logger()

	return logger
}

func WithLogger(ctx context.Context, logger zerolog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func FromContext(ctx context.Context) zerolog.Logger {
	if logger, ok := ctx.Value(loggerKey).(zerolog.Logger); ok {
		return logger
	}
	return zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

func UserIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(userIDKey).(string); ok {
		return id
	}
	return ""
}

func WithWorkspaceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, workspaceIDKey, id)
}

func WorkspaceIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(workspaceIDKey).(string); ok {
		return id
	}
	return ""
}
