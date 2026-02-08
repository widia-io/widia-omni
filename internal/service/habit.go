package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/widia-io/widia-omni/internal/domain"
)

type HabitService struct {
	db         *pgxpool.Pool
	counterSvc *CounterService
}

func NewHabitService(db *pgxpool.Pool, counterSvc *CounterService) *HabitService {
	return &HabitService{db: db, counterSvc: counterSvc}
}

func (s *HabitService) List(ctx context.Context, wsID uuid.UUID) ([]domain.Habit, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, area_id, name, color, frequency, target_per_week,
			   is_active, sort_order, created_at, updated_at
		FROM habits
		WHERE workspace_id = $1 AND deleted_at IS NULL
		ORDER BY sort_order
	`, wsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var habits []domain.Habit
	for rows.Next() {
		var h domain.Habit
		if err := rows.Scan(&h.ID, &h.WorkspaceID, &h.AreaID, &h.Name, &h.Color, &h.Frequency,
			&h.TargetPerWeek, &h.IsActive, &h.SortOrder, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, err
		}
		habits = append(habits, h)
	}
	return habits, nil
}

type CreateHabitRequest struct {
	AreaID        *uuid.UUID           `json:"area_id"`
	Name          string               `json:"name"`
	Color         string               `json:"color"`
	Frequency     domain.HabitFrequency `json:"frequency"`
	TargetPerWeek int                  `json:"target_per_week"`
	SortOrder     int                  `json:"sort_order"`
}

func (s *HabitService) Create(ctx context.Context, wsID uuid.UUID, limits *domain.EntitlementLimits, req CreateHabitRequest) (*domain.Habit, error) {
	counters, err := s.counterSvc.GetCounters(ctx, wsID)
	if err != nil {
		return nil, err
	}
	if !limits.CanCreateHabit(counters.HabitsCount) {
		return nil, errors.New("habit limit reached")
	}

	var h domain.Habit
	err = s.db.QueryRow(ctx, `
		INSERT INTO habits (workspace_id, area_id, name, color, frequency, target_per_week, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, workspace_id, area_id, name, color, frequency, target_per_week,
				  is_active, sort_order, created_at, updated_at
	`, wsID, req.AreaID, req.Name, req.Color, req.Frequency, req.TargetPerWeek, req.SortOrder).Scan(
		&h.ID, &h.WorkspaceID, &h.AreaID, &h.Name, &h.Color, &h.Frequency,
		&h.TargetPerWeek, &h.IsActive, &h.SortOrder, &h.CreatedAt, &h.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

type UpdateHabitRequest struct {
	Name          string               `json:"name"`
	Color         string               `json:"color"`
	Frequency     domain.HabitFrequency `json:"frequency"`
	TargetPerWeek int                  `json:"target_per_week"`
	IsActive      bool                 `json:"is_active"`
}

func (s *HabitService) Update(ctx context.Context, wsID, id uuid.UUID, req UpdateHabitRequest) (*domain.Habit, error) {
	var h domain.Habit
	err := s.db.QueryRow(ctx, `
		UPDATE habits
		SET name = $3, color = $4, frequency = $5, target_per_week = $6, is_active = $7, updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		RETURNING id, workspace_id, area_id, name, color, frequency, target_per_week,
				  is_active, sort_order, created_at, updated_at
	`, id, wsID, req.Name, req.Color, req.Frequency, req.TargetPerWeek, req.IsActive).Scan(
		&h.ID, &h.WorkspaceID, &h.AreaID, &h.Name, &h.Color, &h.Frequency,
		&h.TargetPerWeek, &h.IsActive, &h.SortOrder, &h.CreatedAt, &h.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (s *HabitService) Delete(ctx context.Context, wsID, id uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE habits SET deleted_at = now() WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
	`, id, wsID)
	return err
}

type CheckInRequest struct {
	Date      string  `json:"date"`
	Intensity int16   `json:"intensity"`
	Notes     *string `json:"notes"`
}

func (s *HabitService) CheckIn(ctx context.Context, wsID, habitID uuid.UUID, req CheckInRequest) (*domain.HabitEntry, error) {
	var e domain.HabitEntry
	err := s.db.QueryRow(ctx, `
		INSERT INTO habit_entries (habit_id, workspace_id, date, intensity, notes)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (habit_id, date) DO UPDATE SET intensity = $4, notes = $5
		RETURNING id, habit_id, workspace_id, date, intensity, notes, created_at
	`, habitID, wsID, req.Date, req.Intensity, req.Notes).Scan(
		&e.ID, &e.HabitID, &e.WorkspaceID, &e.Date, &e.Intensity, &e.Notes, &e.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *HabitService) DeleteCheckIn(ctx context.Context, wsID, habitID uuid.UUID, date string) error {
	_, err := s.db.Exec(ctx, `
		DELETE FROM habit_entries WHERE habit_id = $1 AND workspace_id = $2 AND date = $3
	`, habitID, wsID, date)
	return err
}

func (s *HabitService) ListEntries(ctx context.Context, wsID uuid.UUID, from, to string) ([]domain.HabitEntry, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, habit_id, workspace_id, date, intensity, notes, created_at
		FROM habit_entries
		WHERE workspace_id = $1 AND date >= $2 AND date <= $3
		ORDER BY date DESC
	`, wsID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domain.HabitEntry
	for rows.Next() {
		var e domain.HabitEntry
		if err := rows.Scan(&e.ID, &e.HabitID, &e.WorkspaceID, &e.Date, &e.Intensity, &e.Notes, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func (s *HabitService) GetStreaks(ctx context.Context, wsID uuid.UUID) ([]domain.HabitStreak, error) {
	rows, err := s.db.Query(ctx, `
		WITH daily AS (
			SELECT h.id AS habit_id, h.name,
				   he.date,
				   he.date - (ROW_NUMBER() OVER (PARTITION BY h.id ORDER BY he.date))::int AS grp
			FROM habits h
			JOIN habit_entries he ON he.habit_id = h.id
			WHERE h.workspace_id = $1 AND h.deleted_at IS NULL
		),
		streaks AS (
			SELECT habit_id, name, grp,
				   COUNT(*) AS streak_len,
				   MAX(date) AS last_date
			FROM daily
			GROUP BY habit_id, name, grp
		)
		SELECT habit_id, name,
			   COALESCE((SELECT streak_len FROM streaks s2 WHERE s2.habit_id = streaks.habit_id AND s2.last_date = CURRENT_DATE ORDER BY streak_len DESC LIMIT 1), 0),
			   MAX(streak_len),
			   MAX(last_date)::text
		FROM streaks
		GROUP BY habit_id, name
	`, wsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.HabitStreak
	for rows.Next() {
		var hs domain.HabitStreak
		if err := rows.Scan(&hs.HabitID, &hs.Name, &hs.CurrentStreak, &hs.LongestStreak, &hs.LastCheckIn); err != nil {
			return nil, err
		}
		result = append(result, hs)
	}
	return result, nil
}
