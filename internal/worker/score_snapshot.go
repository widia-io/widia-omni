package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/widia-io/widia-omni/internal/domain"
	"github.com/widia-io/widia-omni/internal/service"
)

type ScoreSnapshotHandler struct {
	db       *pgxpool.Pool
	scoreSvc *service.ScoreService
	notifSvc *service.NotificationService
	logger   zerolog.Logger
}

func NewScoreSnapshotHandler(db *pgxpool.Pool, scoreSvc *service.ScoreService, notifSvc *service.NotificationService, logger zerolog.Logger) *ScoreSnapshotHandler {
	return &ScoreSnapshotHandler{db: db, scoreSvc: scoreSvc, notifSvc: notifSvc, logger: logger}
}

func (h *ScoreSnapshotHandler) ProcessTask(ctx context.Context, _ *asynq.Task) error {
	now := time.Now().UTC()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()-time.Monday+7)%7)
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, time.UTC)

	rows, err := h.db.Query(ctx, `
		SELECT w.id, wm.user_id FROM workspaces w
		JOIN workspace_members wm ON wm.workspace_id = w.id AND wm.role = 'owner'
		JOIN subscriptions s ON s.workspace_id = w.id AND s.status IN ('active', 'trialing')
	`)
	if err != nil {
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
			h.logger.Error().Err(err).Msg("score snapshot: scan error")
			continue
		}
		targets = append(targets, t)
	}

	for _, t := range targets {
		if err := h.scoreSvc.Calculate(ctx, t.wsID, weekStart); err != nil {
			h.logger.Error().Err(err).Str("workspace_id", t.wsID.String()).Msg("score snapshot: calculation failed")
			continue
		}

		body := "Your weekly life score has been updated."
		h.notifSvc.Create(ctx, service.CreateNotificationRequest{
			WorkspaceID: t.wsID,
			UserID:      t.userID,
			Type:        domain.NotifScoreUpdate,
			Channel:     domain.ChannelInApp,
			Title:       "Weekly Score Updated",
			Body:        &body,
			Data:        json.RawMessage(`{"week_start":"` + weekStart.Format("2006-01-02") + `"}`),
		})
	}

	h.logger.Info().Int("workspaces", len(targets)).Msg("score snapshot completed")
	return nil
}
