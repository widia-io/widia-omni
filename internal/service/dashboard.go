package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DashboardService struct {
	db *pgxpool.Pool
}

func NewDashboardService(db *pgxpool.Pool) *DashboardService {
	return &DashboardService{db: db}
}

type DashboardData struct {
	AreasCount           int    `json:"areas_count"`
	ActiveGoals          int    `json:"active_goals"`
	TodayTasks           int    `json:"today_tasks"`
	CompletedToday       int    `json:"completed_today"`
	HabitsToday          int    `json:"habits_today"`
	CurrentStreaks       int    `json:"current_streaks"`
	LifeScore            *int16 `json:"life_score"`
	JournalToday         bool   `json:"journal_today"`
	UnreadNotifications  int    `json:"unread_notifications"`
}

func (s *DashboardService) GetDashboard(ctx context.Context, wsID, userID uuid.UUID) (*DashboardData, error) {
	var d DashboardData
	err := s.db.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM life_areas WHERE workspace_id = $1 AND deleted_at IS NULL AND is_active = true),
			(SELECT COUNT(*) FROM goals WHERE workspace_id = $1 AND deleted_at IS NULL AND status NOT IN ('completed', 'cancelled')),
			(SELECT COUNT(*) FROM tasks WHERE workspace_id = $1 AND deleted_at IS NULL AND is_completed = false AND (due_date IS NULL OR due_date <= CURRENT_DATE)),
			(SELECT COUNT(*) FROM tasks WHERE workspace_id = $1 AND deleted_at IS NULL AND is_completed = true AND completed_at::date = CURRENT_DATE),
			(SELECT COUNT(DISTINCT he.habit_id) FROM habit_entries he JOIN habits h ON h.id = he.habit_id WHERE h.workspace_id = $1 AND h.deleted_at IS NULL AND he.date = CURRENT_DATE),
			(SELECT COUNT(*) FROM habits WHERE workspace_id = $1 AND deleted_at IS NULL AND is_active = true AND id IN (
				SELECT habit_id FROM habit_entries WHERE date = CURRENT_DATE - INTERVAL '1 day'
			)),
			(SELECT score FROM life_scores WHERE workspace_id = $1 ORDER BY week_start DESC LIMIT 1),
			(SELECT EXISTS(SELECT 1 FROM journal_entries WHERE workspace_id = $1 AND date = CURRENT_DATE)),
			(SELECT COUNT(*) FROM notifications WHERE workspace_id = $1 AND user_id = $2 AND is_read = false)
	`, wsID, userID).Scan(&d.AreasCount, &d.ActiveGoals, &d.TodayTasks, &d.CompletedToday,
		&d.HabitsToday, &d.CurrentStreaks, &d.LifeScore, &d.JournalToday, &d.UnreadNotifications)
	if err != nil {
		return nil, err
	}
	return &d, nil
}
