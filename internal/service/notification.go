package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/widia-io/widia-omni/internal/domain"
)

type NotificationService struct {
	db *pgxpool.Pool
}

func NewNotificationService(db *pgxpool.Pool) *NotificationService {
	return &NotificationService{db: db}
}

func (s *NotificationService) List(ctx context.Context, wsID, userID uuid.UUID, unreadOnly bool, limit, offset int) ([]domain.Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := `
		SELECT id, workspace_id, user_id, type, channel, title, body, data, is_read, read_at, created_at
		FROM notifications
		WHERE workspace_id = $1 AND user_id = $2`
	if unreadOnly {
		query += ` AND is_read = false`
	}
	query += ` ORDER BY created_at DESC LIMIT $3 OFFSET $4`

	rows, err := s.db.Query(ctx, query, wsID, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifs []domain.Notification
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.WorkspaceID, &n.UserID, &n.Type, &n.Channel,
			&n.Title, &n.Body, &n.Data, &n.IsRead, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifs = append(notifs, n)
	}
	return notifs, nil
}

type CreateNotificationRequest struct {
	WorkspaceID uuid.UUID                  `json:"workspace_id"`
	UserID      uuid.UUID                  `json:"user_id"`
	Type        domain.NotificationType    `json:"type"`
	Channel     domain.NotificationChannel `json:"channel"`
	Title       string                     `json:"title"`
	Body        *string                    `json:"body"`
	Data        json.RawMessage            `json:"data"`
}

func (s *NotificationService) Create(ctx context.Context, req CreateNotificationRequest) (*domain.Notification, error) {
	var n domain.Notification
	err := s.db.QueryRow(ctx, `
		INSERT INTO notifications (workspace_id, user_id, type, channel, title, body, data)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, workspace_id, user_id, type, channel, title, body, data, is_read, read_at, created_at
	`, req.WorkspaceID, req.UserID, req.Type, req.Channel, req.Title, req.Body, req.Data).Scan(
		&n.ID, &n.WorkspaceID, &n.UserID, &n.Type, &n.Channel,
		&n.Title, &n.Body, &n.Data, &n.IsRead, &n.ReadAt, &n.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (s *NotificationService) MarkRead(ctx context.Context, wsID, userID, id uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE notifications SET is_read = true, read_at = now()
		WHERE id = $1 AND workspace_id = $2 AND user_id = $3
	`, id, wsID, userID)
	return err
}

func (s *NotificationService) MarkAllRead(ctx context.Context, wsID, userID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE notifications SET is_read = true, read_at = now()
		WHERE workspace_id = $1 AND user_id = $2 AND is_read = false
	`, wsID, userID)
	return err
}

func (s *NotificationService) UnreadCount(ctx context.Context, wsID, userID uuid.UUID) (int, error) {
	var count int
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM notifications
		WHERE workspace_id = $1 AND user_id = $2 AND is_read = false
	`, wsID, userID).Scan(&count)
	return count, err
}
