package observability

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
)

var (
	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path", "status"})

	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	ActiveSubscriptions = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "active_subscriptions",
		Help: "Number of active subscriptions by tier",
	}, []string{"tier"})

	AsynqQueueDepth = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "asynq_queue_depth",
		Help: "Number of pending tasks in asynq queues",
	}, []string{"queue"})

	AsynqJobFailuresTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "asynq_job_failures_total",
		Help: "Total number of asynq job failures",
	}, []string{"task_type"})

	EntitlementLimitReachedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "entitlement_limit_reached_total",
		Help: "Total times an entitlement limit was hit",
	}, []string{"limit_type"})
)

func RefreshSubscriptionGauge(ctx context.Context, db *pgxpool.Pool, logger zerolog.Logger) {
	rows, err := db.Query(ctx, `
		SELECT tier, COUNT(*) FROM subscriptions
		WHERE status IN ('active', 'trialing')
		GROUP BY tier
	`)
	if err != nil {
		logger.Error().Err(err).Msg("failed to refresh subscription gauge")
		return
	}
	defer rows.Close()

	ActiveSubscriptions.Reset()
	for rows.Next() {
		var tier string
		var count int
		if err := rows.Scan(&tier, &count); err != nil {
			continue
		}
		ActiveSubscriptions.WithLabelValues(tier).Set(float64(count))
	}
}

type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newMetricsResponseWriter(w http.ResponseWriter) *metricsResponseWriter {
	return &metricsResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (m *metricsResponseWriter) WriteHeader(code int) {
	m.statusCode = code
	m.ResponseWriter.WriteHeader(code)
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		mrw := newMetricsResponseWriter(w)

		next.ServeHTTP(mrw, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(mrw.statusCode)

		HTTPRequestDuration.WithLabelValues(r.Method, r.URL.Path, status).Observe(duration)
		HTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
	})
}
