package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OnboardingService struct {
	db *pgxpool.Pool
}

func NewOnboardingService(db *pgxpool.Pool) *OnboardingService {
	return &OnboardingService{db: db}
}

type OnboardingStatus struct {
	OnboardingCompleted bool `json:"onboarding_completed"`
	AreasCount          int  `json:"areas_count"`
	GoalsCount          int  `json:"goals_count"`
	HabitsCount         int  `json:"habits_count"`
}

func (s *OnboardingService) GetStatus(ctx context.Context, userID, wsID uuid.UUID) (*OnboardingStatus, error) {
	var st OnboardingStatus
	err := s.db.QueryRow(ctx, `SELECT onboarding_completed FROM user_profiles WHERE id = $1`, userID).Scan(&st.OnboardingCompleted)
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM life_areas WHERE workspace_id = $1 AND deleted_at IS NULL),
			(SELECT COUNT(*) FROM goals WHERE workspace_id = $1 AND deleted_at IS NULL),
			(SELECT COUNT(*) FROM habits WHERE workspace_id = $1 AND deleted_at IS NULL)
	`, wsID).Scan(&st.AreasCount, &st.GoalsCount, &st.HabitsCount)
	if err != nil {
		return nil, err
	}
	return &st, nil
}

type SetupAreaItem struct {
	Name      string  `json:"name"`
	Slug      string  `json:"slug"`
	Icon      string  `json:"icon"`
	Color     string  `json:"color"`
	Weight    float64 `json:"weight"`
	SortOrder int     `json:"sort_order"`
}

func (s *OnboardingService) SetupAreas(ctx context.Context, wsID uuid.UUID, areas []SetupAreaItem) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, a := range areas {
		_, err := tx.Exec(ctx, `
			INSERT INTO life_areas (workspace_id, name, slug, icon, color, weight, sort_order)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, wsID, a.Name, a.Slug, a.Icon, a.Color, a.Weight, a.SortOrder)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

type SetupGoalItem struct {
	AreaID    *uuid.UUID `json:"area_id"`
	Title     string     `json:"title"`
	Period    string     `json:"period"`
	StartDate string     `json:"start_date"`
	EndDate   string     `json:"end_date"`
}

func (s *OnboardingService) SetupGoals(ctx context.Context, wsID uuid.UUID, goals []SetupGoalItem) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, g := range goals {
		_, err := tx.Exec(ctx, `
			INSERT INTO goals (workspace_id, area_id, title, period, start_date, end_date)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, wsID, g.AreaID, g.Title, g.Period, g.StartDate, g.EndDate)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

type SetupHabitItem struct {
	AreaID        *uuid.UUID `json:"area_id"`
	Name          string     `json:"name"`
	Color         string     `json:"color"`
	Frequency     string     `json:"frequency"`
	TargetPerWeek int        `json:"target_per_week"`
}

func (s *OnboardingService) SetupHabits(ctx context.Context, wsID uuid.UUID, habits []SetupHabitItem) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, h := range habits {
		_, err := tx.Exec(ctx, `
			INSERT INTO habits (workspace_id, area_id, name, color, frequency, target_per_week)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, wsID, h.AreaID, h.Name, h.Color, h.Frequency, h.TargetPerWeek)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *OnboardingService) Complete(ctx context.Context, userID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `UPDATE user_profiles SET onboarding_completed = true WHERE id = $1`, userID)
	return err
}
