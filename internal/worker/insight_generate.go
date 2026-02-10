package worker

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/widia-io/widia-omni/internal/domain"
	"github.com/widia-io/widia-omni/internal/observability"
	"github.com/widia-io/widia-omni/internal/service"
)

type InsightGenerateHandler struct {
	db         *pgxpool.Pool
	insightSvc *service.InsightService
	entSvc     *service.EntitlementService
	notifSvc   *service.NotificationService
	logger     zerolog.Logger
}

func NewInsightGenerateHandler(db *pgxpool.Pool, insightSvc *service.InsightService, entSvc *service.EntitlementService, notifSvc *service.NotificationService, logger zerolog.Logger) *InsightGenerateHandler {
	return &InsightGenerateHandler{db: db, insightSvc: insightSvc, entSvc: entSvc, notifSvc: notifSvc, logger: logger}
}

func (h *InsightGenerateHandler) ProcessTask(ctx context.Context, _ *asynq.Task) error {
	rows, err := h.db.Query(ctx, `
		SELECT w.id, wm.user_id FROM workspaces w
		JOIN workspace_members wm ON wm.workspace_id = w.id AND wm.role = 'owner'
		JOIN subscriptions s ON s.workspace_id = w.id AND s.status IN ('active', 'trialing')
	`)
	if err != nil {
		observability.AsynqJobFailuresTotal.WithLabelValues(TypeInsightGenerate).Inc()
		return err
	}
	defer rows.Close()

	type wsUser struct {
		wsID   uuid.UUID
		userID uuid.UUID
	}
	var targets []wsUser
	for rows.Next() {
		var t wsUser
		if err := rows.Scan(&t.wsID, &t.userID); err != nil {
			h.logger.Error().Err(err).Msg("insight generate: scan error")
			continue
		}
		targets = append(targets, t)
	}

	var generated int
	for _, t := range targets {
		ent, err := h.entSvc.GetCurrent(ctx, t.wsID)
		if err != nil {
			h.logger.Error().Err(err).Str("workspace_id", t.wsID.String()).Msg("insight generate: get entitlement")
			continue
		}
		limits, err := domain.ParseLimits(ent.Limits)
		if err != nil || !limits.AIInsights {
			continue
		}

		_, err = h.insightSvc.Generate(ctx, t.wsID, domain.InsightWeeklySummary)
		if err != nil {
			h.logger.Error().Err(err).Str("workspace_id", t.wsID.String()).Msg("insight generate: generation failed")
			continue
		}

		body := "Your weekly AI insights are ready."
		h.notifSvc.Create(ctx, service.CreateNotificationRequest{
			WorkspaceID: t.wsID,
			UserID:      t.userID,
			Type:        domain.NotifSystem,
			Channel:     domain.ChannelInApp,
			Title:       "Weekly AI Insights Ready",
			Body:        &body,
			Data:        json.RawMessage(`{}`),
		})
		generated++
	}

	h.logger.Info().Int("workspaces", len(targets)).Int("generated", generated).Msg("insight generation completed")
	return nil
}
