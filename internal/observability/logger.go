package observability

import (
	"context"
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/widia-io/widia-omni/internal/config"
)

type ctxKey string

const (
	loggerKey      ctxKey = "logger"
	requestIDKey   ctxKey = "request_id"
	userIDKey      ctxKey = "user_id"
	workspaceIDKey ctxKey = "workspace_id"
)

func NewLogger(cfg *config.Config) zerolog.Logger {
	level := zerolog.InfoLevel
	if cfg.Env == "development" {
		level = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(level)

	_ = os.MkdirAll("logs", 0o755)
	logFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		logFile = nil
	}

	var writers []io.Writer
	if cfg.Env == "development" {
		writers = append(writers, zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"})
	} else {
		writers = append(writers, os.Stdout)
	}
	if logFile != nil {
		writers = append(writers, logFile)
	}

	multi := zerolog.MultiLevelWriter(writers...)

	return zerolog.New(multi).
		With().
		Timestamp().
		Str("service", "widia-omni").
		Str("env", cfg.Env).
		Logger()
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
