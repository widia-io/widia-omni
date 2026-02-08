package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/widia-io/widia-omni/internal/domain"
)

type ExportService struct {
	db *pgxpool.Pool
}

func NewExportService(db *pgxpool.Pool) *ExportService {
	return &ExportService{db: db}
}

type ExportData struct {
	ExportedAt    time.Time              `json:"exported_at"`
	Areas         []domain.LifeArea      `json:"areas"`
	Goals         []domain.Goal          `json:"goals"`
	Habits        []domain.Habit         `json:"habits"`
	Tasks         []domain.Task          `json:"tasks"`
	Journal       []domain.JournalEntry  `json:"journal"`
	LifeScores    []domain.LifeScore     `json:"life_scores"`
	AreaScores    []domain.AreaScore     `json:"area_scores"`
	Notifications []domain.Notification  `json:"notifications"`
}

func (s *ExportService) Export(ctx context.Context, wsID, userID uuid.UUID) (*ExportData, error) {
	data := &ExportData{ExportedAt: time.Now().UTC()}

	// Areas
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, name, slug, icon, color, weight, sort_order, is_active, created_at, updated_at
		FROM life_areas WHERE workspace_id = $1 AND deleted_at IS NULL ORDER BY sort_order
	`, wsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var a domain.LifeArea
		if err := rows.Scan(&a.ID, &a.WorkspaceID, &a.Name, &a.Slug, &a.Icon, &a.Color,
			&a.Weight, &a.SortOrder, &a.IsActive, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		data.Areas = append(data.Areas, a)
	}
	rows.Close()

	// Goals
	rows, err = s.db.Query(ctx, `
		SELECT id, workspace_id, area_id, parent_id, title, description, period, status,
			   target_value, current_value, unit, start_date, end_date, completed_at, created_at, updated_at
		FROM goals WHERE workspace_id = $1 AND deleted_at IS NULL ORDER BY created_at
	`, wsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var g domain.Goal
		if err := rows.Scan(&g.ID, &g.WorkspaceID, &g.AreaID, &g.ParentID, &g.Title, &g.Description,
			&g.Period, &g.Status, &g.TargetValue, &g.CurrentValue, &g.Unit,
			&g.StartDate, &g.EndDate, &g.CompletedAt, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		data.Goals = append(data.Goals, g)
	}
	rows.Close()

	// Habits
	rows, err = s.db.Query(ctx, `
		SELECT id, workspace_id, area_id, name, color, frequency, target_per_week, is_active, sort_order,
			   created_at, updated_at
		FROM habits WHERE workspace_id = $1 AND deleted_at IS NULL ORDER BY sort_order
	`, wsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var h domain.Habit
		if err := rows.Scan(&h.ID, &h.WorkspaceID, &h.AreaID, &h.Name, &h.Color,
			&h.Frequency, &h.TargetPerWeek, &h.IsActive, &h.SortOrder,
			&h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, err
		}
		data.Habits = append(data.Habits, h)
	}
	rows.Close()

	// Tasks
	rows, err = s.db.Query(ctx, `
		SELECT id, workspace_id, area_id, goal_id, title, description, priority, is_completed, is_focus,
			   due_date, completed_at, created_at, updated_at
		FROM tasks WHERE workspace_id = $1 AND deleted_at IS NULL ORDER BY created_at
	`, wsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(&t.ID, &t.WorkspaceID, &t.AreaID, &t.GoalID, &t.Title, &t.Description,
			&t.Priority, &t.IsCompleted, &t.IsFocus, &t.DueDate, &t.CompletedAt,
			&t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		data.Tasks = append(data.Tasks, t)
	}
	rows.Close()

	// Journal
	rows, err = s.db.Query(ctx, `
		SELECT id, workspace_id, date, mood, energy, wins, challenges, gratitude, notes, tags,
			   created_at, updated_at
		FROM journal_entries WHERE workspace_id = $1 ORDER BY date DESC
	`, wsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var e domain.JournalEntry
		if err := rows.Scan(&e.ID, &e.WorkspaceID, &e.Date, &e.Mood, &e.Energy,
			&e.Wins, &e.Challenges, &e.Gratitude, &e.Notes, &e.Tags,
			&e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		data.Journal = append(data.Journal, e)
	}
	rows.Close()

	// Life Scores
	rows, err = s.db.Query(ctx, `
		SELECT id, workspace_id, score, week_start, area_scores, created_at
		FROM life_scores WHERE workspace_id = $1 ORDER BY week_start DESC
	`, wsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var ls domain.LifeScore
		if err := rows.Scan(&ls.ID, &ls.WorkspaceID, &ls.Score, &ls.WeekStart, &ls.AreaScores, &ls.CreatedAt); err != nil {
			return nil, err
		}
		data.LifeScores = append(data.LifeScores, ls)
	}
	rows.Close()

	// Area Scores
	rows, err = s.db.Query(ctx, `
		SELECT id, workspace_id, area_id, score, week_start, breakdown, created_at
		FROM area_scores WHERE workspace_id = $1 ORDER BY week_start DESC
	`, wsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var as domain.AreaScore
		if err := rows.Scan(&as.ID, &as.WorkspaceID, &as.AreaID, &as.Score, &as.WeekStart, &as.Breakdown, &as.CreatedAt); err != nil {
			return nil, err
		}
		data.AreaScores = append(data.AreaScores, as)
	}
	rows.Close()

	// Notifications
	rows, err = s.db.Query(ctx, `
		SELECT id, workspace_id, user_id, type, channel, title, body, data, is_read, read_at, created_at
		FROM notifications WHERE workspace_id = $1 AND user_id = $2 ORDER BY created_at DESC
	`, wsID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.WorkspaceID, &n.UserID, &n.Type, &n.Channel,
			&n.Title, &n.Body, &n.Data, &n.IsRead, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, err
		}
		data.Notifications = append(data.Notifications, n)
	}

	return data, nil
}
