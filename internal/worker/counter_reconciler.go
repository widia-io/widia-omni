package worker

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/widia-io/widia-omni/internal/observability"
)

type CounterReconcilerHandler struct {
	db     *pgxpool.Pool
	logger zerolog.Logger
}

func NewCounterReconcilerHandler(db *pgxpool.Pool, logger zerolog.Logger) *CounterReconcilerHandler {
	return &CounterReconcilerHandler{db: db, logger: logger}
}

func (h *CounterReconcilerHandler) ProcessTask(ctx context.Context, _ *asynq.Task) error {
	tag, err := h.db.Exec(ctx, `
		UPDATE workspace_counters wc SET
			areas_count = (SELECT COUNT(*) FROM life_areas WHERE workspace_id = wc.workspace_id AND deleted_at IS NULL),
			goals_count = (SELECT COUNT(*) FROM goals WHERE workspace_id = wc.workspace_id AND deleted_at IS NULL),
			habits_count = (SELECT COUNT(*) FROM habits WHERE workspace_id = wc.workspace_id AND deleted_at IS NULL),
			updated_at = now()
	`)
	if err != nil {
		observability.AsynqJobFailuresTotal.WithLabelValues(TypeCounterReconcile).Inc()
		return err
	}

	h.logger.Info().Int64("rows", tag.RowsAffected()).Msg("counter reconciler completed")
	return nil
}
