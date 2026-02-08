package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/widia-io/widia-omni/internal/domain"
	"github.com/widia-io/widia-omni/internal/observability"
)

type GoalService struct {
	db         *pgxpool.Pool
	counterSvc *CounterService
}

func NewGoalService(db *pgxpool.Pool, counterSvc *CounterService) *GoalService {
	return &GoalService{db: db, counterSvc: counterSvc}
}

type GoalFilters struct {
	AreaID *uuid.UUID         `json:"area_id"`
	Period *domain.GoalPeriod `json:"period"`
	Status *domain.GoalStatus `json:"status"`
}

func (s *GoalService) List(ctx context.Context, wsID uuid.UUID, f GoalFilters) ([]domain.Goal, error) {
	query := `
		SELECT id, workspace_id, area_id, parent_id, title, description, period, status,
			   target_value, current_value, unit, start_date, end_date, completed_at,
			   created_at, updated_at
		FROM goals
		WHERE workspace_id = $1 AND deleted_at IS NULL
	`
	args := []any{wsID}
	idx := 2

	if f.AreaID != nil {
		query += fmt.Sprintf(` AND area_id = $%d`, idx)
		args = append(args, *f.AreaID)
		idx++
	}
	if f.Period != nil {
		query += fmt.Sprintf(` AND period = $%d`, idx)
		args = append(args, *f.Period)
		idx++
	}
	if f.Status != nil {
		query += fmt.Sprintf(` AND status = $%d`, idx)
		args = append(args, *f.Status)
		idx++
	}
	query += ` ORDER BY created_at DESC`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []domain.Goal
	for rows.Next() {
		var g domain.Goal
		if err := rows.Scan(&g.ID, &g.WorkspaceID, &g.AreaID, &g.ParentID, &g.Title, &g.Description,
			&g.Period, &g.Status, &g.TargetValue, &g.CurrentValue, &g.Unit, &g.StartDate, &g.EndDate,
			&g.CompletedAt, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		goals = append(goals, g)
	}
	return goals, nil
}

func (s *GoalService) GetByID(ctx context.Context, wsID, id uuid.UUID) (*domain.Goal, error) {
	var g domain.Goal
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, area_id, parent_id, title, description, period, status,
			   target_value, current_value, unit, start_date, end_date, completed_at,
			   created_at, updated_at
		FROM goals
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
	`, id, wsID).Scan(&g.ID, &g.WorkspaceID, &g.AreaID, &g.ParentID, &g.Title, &g.Description,
		&g.Period, &g.Status, &g.TargetValue, &g.CurrentValue, &g.Unit, &g.StartDate, &g.EndDate,
		&g.CompletedAt, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

type CreateGoalRequest struct {
	AreaID      *uuid.UUID        `json:"area_id"`
	ParentID    *uuid.UUID        `json:"parent_id"`
	Title       string            `json:"title"`
	Description *string           `json:"description"`
	Period      domain.GoalPeriod `json:"period"`
	TargetValue *float64          `json:"target_value"`
	Unit        *string           `json:"unit"`
	StartDate   string            `json:"start_date"`
	EndDate     string            `json:"end_date"`
}

func (s *GoalService) Create(ctx context.Context, wsID uuid.UUID, limits *domain.EntitlementLimits, req CreateGoalRequest) (*domain.Goal, error) {
	counters, err := s.counterSvc.GetCounters(ctx, wsID)
	if err != nil {
		return nil, err
	}
	if !limits.CanCreateGoal(counters.GoalsCount) {
		observability.EntitlementLimitReachedTotal.WithLabelValues("goals").Inc()
		return nil, errors.New("goal limit reached")
	}

	var g domain.Goal
	err = s.db.QueryRow(ctx, `
		INSERT INTO goals (workspace_id, area_id, parent_id, title, description, period,
						   target_value, unit, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, workspace_id, area_id, parent_id, title, description, period, status,
				  target_value, current_value, unit, start_date, end_date, completed_at,
				  created_at, updated_at
	`, wsID, req.AreaID, req.ParentID, req.Title, req.Description, req.Period,
		req.TargetValue, req.Unit, req.StartDate, req.EndDate).Scan(
		&g.ID, &g.WorkspaceID, &g.AreaID, &g.ParentID, &g.Title, &g.Description,
		&g.Period, &g.Status, &g.TargetValue, &g.CurrentValue, &g.Unit, &g.StartDate, &g.EndDate,
		&g.CompletedAt, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

type UpdateGoalRequest struct {
	AreaID      *uuid.UUID         `json:"area_id"`
	Title       string             `json:"title"`
	Description *string            `json:"description"`
	Period      domain.GoalPeriod  `json:"period"`
	Status      domain.GoalStatus  `json:"status"`
	TargetValue *float64           `json:"target_value"`
	Unit        *string            `json:"unit"`
	StartDate   string             `json:"start_date"`
	EndDate     string             `json:"end_date"`
}

func (s *GoalService) Update(ctx context.Context, wsID, id uuid.UUID, req UpdateGoalRequest) (*domain.Goal, error) {
	var g domain.Goal
	err := s.db.QueryRow(ctx, `
		UPDATE goals
		SET area_id = $3, title = $4, description = $5, period = $6, status = $7,
			target_value = $8, unit = $9, start_date = $10, end_date = $11, updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		RETURNING id, workspace_id, area_id, parent_id, title, description, period, status,
				  target_value, current_value, unit, start_date, end_date, completed_at,
				  created_at, updated_at
	`, id, wsID, req.AreaID, req.Title, req.Description, req.Period, req.Status,
		req.TargetValue, req.Unit, req.StartDate, req.EndDate).Scan(
		&g.ID, &g.WorkspaceID, &g.AreaID, &g.ParentID, &g.Title, &g.Description,
		&g.Period, &g.Status, &g.TargetValue, &g.CurrentValue, &g.Unit, &g.StartDate, &g.EndDate,
		&g.CompletedAt, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (s *GoalService) Delete(ctx context.Context, wsID, id uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE goals SET deleted_at = now() WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
	`, id, wsID)
	return err
}

func (s *GoalService) UpdateProgress(ctx context.Context, wsID, id uuid.UUID, currentValue float64) (*domain.Goal, error) {
	var g domain.Goal
	err := s.db.QueryRow(ctx, `
		UPDATE goals
		SET current_value = $3,
			status = CASE WHEN target_value IS NOT NULL AND $3 >= target_value THEN 'completed'::goal_status ELSE status END,
			completed_at = CASE WHEN target_value IS NOT NULL AND $3 >= target_value THEN now() ELSE completed_at END,
			updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		RETURNING id, workspace_id, area_id, parent_id, title, description, period, status,
				  target_value, current_value, unit, start_date, end_date, completed_at,
				  created_at, updated_at
	`, id, wsID, currentValue).Scan(
		&g.ID, &g.WorkspaceID, &g.AreaID, &g.ParentID, &g.Title, &g.Description,
		&g.Period, &g.Status, &g.TargetValue, &g.CurrentValue, &g.Unit, &g.StartDate, &g.EndDate,
		&g.CompletedAt, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

