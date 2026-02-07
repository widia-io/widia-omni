package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/widia-io/widia-omni/internal/observability"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func Logger(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqID := observability.RequestIDFromContext(r.Context())

			l := logger.With().
				Str("request_id", reqID).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Logger()

			ctx := observability.WithLogger(r.Context(), l)
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(sw, r.WithContext(ctx))

			l.Info().
				Int("status", sw.status).
				Dur("duration_ms", time.Since(start)).
				Msg("request completed")
		})
	}
}
