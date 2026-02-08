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

type StreakUpdateHandler struct {
	db       *pgxpool.Pool
	notifSvc *service.NotificationService
	logger   zerolog.Logger
}

func NewStreakUpdateHandler(db *pgxpool.Pool, notifSvc *service.NotificationService, logger zerolog.Logger) *StreakUpdateHandler {
	return &StreakUpdateHandler{db: db, notifSvc: notifSvc, logger: logger}
}

func (h *StreakUpdateHandler) ProcessTask(ctx context.Context, _ *asynq.Task) error {
	rows, err := h.db.Query(ctx, `
		SELECT h.id, h.name, h.workspace_id, wm.user_id
		FROM habits h
		JOIN workspace_members wm ON wm.workspace_id = h.workspace_id AND wm.role = 'owner'
		WHERE h.deleted_at IS NULL AND h.is_active = true
			AND h.id IN (
				SELECT habit_id FROM habit_entries WHERE date = CURRENT_DATE - INTERVAL '1 day'
			)
			AND h.id NOT IN (
				SELECT habit_id FROM habit_entries WHERE date = CURRENT_DATE
			)
	`)
	if err != nil {
		observability.AsynqJobFailuresTotal.WithLabelValues(TypeStreakUpdate).Inc()
		return err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var habitID, wsID, userID uuid.UUID
		var habitName string
		if err := rows.Scan(&habitID, &habitName, &wsID, &userID); err != nil {
			h.logger.Error().Err(err).Msg("streak update: scan error")
			continue
		}

		body := "Your streak for \"" + habitName + "\" is at risk! Check in today to keep it going."
		data, _ := json.Marshal(map[string]string{"habit_id": habitID.String(), "habit_name": habitName})
		h.notifSvc.Create(ctx, service.CreateNotificationRequest{
			WorkspaceID: wsID,
			UserID:      userID,
			Type:        domain.NotifStreakAtRisk,
			Channel:     domain.ChannelInApp,
			Title:       "Streak at Risk: " + habitName,
			Body:        &body,
			Data:        data,
		})
		count++
	}

	h.logger.Info().Int("at_risk", count).Msg("streak update completed")
	return nil
}
