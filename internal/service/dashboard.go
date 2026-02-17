package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type DashboardService struct {
	db  *pgxpool.Pool
	rdb *redis.Client
}

func NewDashboardService(db *pgxpool.Pool, rdb *redis.Client) *DashboardService {
	return &DashboardService{db: db, rdb: rdb}
}

type DashboardData struct {
	AreasCount          int    `json:"areas_count"`
	ActiveGoals         int    `json:"active_goals"`
	ActiveProjects      int    `json:"active_projects"`
	TodayTasks          int    `json:"today_tasks"`
	CompletedToday      int    `json:"completed_today"`
	HabitsToday         int    `json:"habits_today"`
	CurrentStreaks      int    `json:"current_streaks"`
	LifeScore           *int16 `json:"life_score"`
	JournalToday        bool   `json:"journal_today"`
	UnreadNotifications int    `json:"unread_notifications"`
}

func (s *DashboardService) GetDashboard(ctx context.Context, wsID, userID uuid.UUID) (*DashboardData, error) {
	cacheKey := fmt.Sprintf("ws:%s:dash", wsID.String())

	if s.rdb != nil {
		cached, err := s.rdb.Get(ctx, cacheKey).Bytes()
		if err == nil {
			var d DashboardData
			if json.Unmarshal(cached, &d) == nil {
				return &d, nil
			}
		}
	}

	var d DashboardData
	err := s.db.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM life_areas WHERE workspace_id = $1 AND deleted_at IS NULL AND is_active = true),
			(SELECT COUNT(*) FROM goals WHERE workspace_id = $1 AND deleted_at IS NULL AND status NOT IN ('completed', 'cancelled')),
			(SELECT COUNT(*) FROM projects WHERE workspace_id = $1 AND deleted_at IS NULL AND status IN ('planning','active') AND is_archived = false),
			(SELECT COUNT(*) FROM tasks WHERE workspace_id = $1 AND deleted_at IS NULL AND is_completed = false AND (due_date IS NULL OR due_date <= CURRENT_DATE)),
			(SELECT COUNT(*) FROM tasks WHERE workspace_id = $1 AND deleted_at IS NULL AND is_completed = true AND completed_at::date = CURRENT_DATE),
			(SELECT COUNT(DISTINCT he.habit_id) FROM habit_entries he JOIN habits h ON h.id = he.habit_id WHERE h.workspace_id = $1 AND h.deleted_at IS NULL AND he.date = CURRENT_DATE),
			(SELECT COUNT(*) FROM habits WHERE workspace_id = $1 AND deleted_at IS NULL AND is_active = true AND id IN (
				SELECT habit_id FROM habit_entries WHERE date = CURRENT_DATE - INTERVAL '1 day'
			)),
			(SELECT score FROM life_scores WHERE workspace_id = $1 ORDER BY week_start DESC LIMIT 1),
			(SELECT EXISTS(SELECT 1 FROM journal_entries WHERE workspace_id = $1 AND date = CURRENT_DATE)),
			(SELECT COUNT(*) FROM notifications WHERE workspace_id = $1 AND user_id = $2 AND is_read = false)
	`, wsID, userID).Scan(&d.AreasCount, &d.ActiveGoals, &d.ActiveProjects, &d.TodayTasks, &d.CompletedToday,
		&d.HabitsToday, &d.CurrentStreaks, &d.LifeScore, &d.JournalToday, &d.UnreadNotifications)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		data, _ := json.Marshal(d)
		s.rdb.Set(ctx, cacheKey, data, 2*time.Minute)
	}

	return &d, nil
}
