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

type TaskService struct {
	db         *pgxpool.Pool
	counterSvc *CounterService
}

func NewTaskService(db *pgxpool.Pool, counterSvc *CounterService) *TaskService {
	return &TaskService{db: db, counterSvc: counterSvc}
}

type TaskFilters struct {
	AreaID      *uuid.UUID `json:"area_id"`
	GoalID      *uuid.UUID `json:"goal_id"`
	IsCompleted *bool      `json:"is_completed"`
	DueFrom     *string    `json:"due_from"`
	DueTo       *string    `json:"due_to"`
}

func (s *TaskService) List(ctx context.Context, wsID uuid.UUID, f TaskFilters) ([]domain.Task, error) {
	query := `
		SELECT id, workspace_id, area_id, goal_id, title, description, priority,
			   is_completed, is_focus, due_date, completed_at, created_at, updated_at
		FROM tasks
		WHERE workspace_id = $1 AND deleted_at IS NULL
	`
	args := []any{wsID}
	idx := 2

	if f.AreaID != nil {
		query += fmt.Sprintf(` AND area_id = $%d`, idx)
		args = append(args, *f.AreaID)
		idx++
	}
	if f.GoalID != nil {
		query += fmt.Sprintf(` AND goal_id = $%d`, idx)
		args = append(args, *f.GoalID)
		idx++
	}
	if f.IsCompleted != nil {
		query += fmt.Sprintf(` AND is_completed = $%d`, idx)
		args = append(args, *f.IsCompleted)
		idx++
	}
	if f.DueFrom != nil {
		query += fmt.Sprintf(` AND due_date >= $%d`, idx)
		args = append(args, *f.DueFrom)
		idx++
	}
	if f.DueTo != nil {
		query += fmt.Sprintf(` AND due_date <= $%d`, idx)
		args = append(args, *f.DueTo)
		idx++
	}
	query += ` ORDER BY CASE WHEN is_focus THEN 0 ELSE 1 END, created_at DESC`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(&t.ID, &t.WorkspaceID, &t.AreaID, &t.GoalID, &t.Title, &t.Description,
			&t.Priority, &t.IsCompleted, &t.IsFocus, &t.DueDate, &t.CompletedAt, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

type CreateTaskRequest struct {
	AreaID      *uuid.UUID         `json:"area_id"`
	GoalID      *uuid.UUID         `json:"goal_id"`
	Title       string             `json:"title"`
	Description *string            `json:"description"`
	Priority    domain.TaskPriority `json:"priority"`
	DueDate     *string            `json:"due_date"`
	IsFocus     bool               `json:"is_focus"`
}

func (s *TaskService) Create(ctx context.Context, wsID uuid.UUID, limits *domain.EntitlementLimits, req CreateTaskRequest) (*domain.Task, error) {
	counters, err := s.counterSvc.GetCounters(ctx, wsID)
	if err != nil {
		return nil, err
	}
	if !limits.CanCreateTask(counters.TasksCreatedToday) {
		observability.EntitlementLimitReachedTotal.WithLabelValues("tasks").Inc()
		return nil, errors.New("daily task limit reached")
	}

	var t domain.Task
	err = s.db.QueryRow(ctx, `
		INSERT INTO tasks (workspace_id, area_id, goal_id, title, description, priority, due_date, is_focus)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, workspace_id, area_id, goal_id, title, description, priority,
				  is_completed, is_focus, due_date, completed_at, created_at, updated_at
	`, wsID, req.AreaID, req.GoalID, req.Title, req.Description, req.Priority, req.DueDate, req.IsFocus).Scan(
		&t.ID, &t.WorkspaceID, &t.AreaID, &t.GoalID, &t.Title, &t.Description,
		&t.Priority, &t.IsCompleted, &t.IsFocus, &t.DueDate, &t.CompletedAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	_ = s.counterSvc.IncrementTasksToday(ctx, wsID)
	return &t, nil
}

type UpdateTaskRequest struct {
	AreaID      *uuid.UUID          `json:"area_id"`
	GoalID      *uuid.UUID          `json:"goal_id"`
	Title       string              `json:"title"`
	Description *string             `json:"description"`
	Priority    domain.TaskPriority `json:"priority"`
	DueDate     *string             `json:"due_date"`
}

func (s *TaskService) Update(ctx context.Context, wsID, id uuid.UUID, req UpdateTaskRequest) (*domain.Task, error) {
	var t domain.Task
	err := s.db.QueryRow(ctx, `
		UPDATE tasks
		SET area_id = $3, goal_id = $4, title = $5, description = $6, priority = $7, due_date = $8, updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		RETURNING id, workspace_id, area_id, goal_id, title, description, priority,
				  is_completed, is_focus, due_date, completed_at, created_at, updated_at
	`, id, wsID, req.AreaID, req.GoalID, req.Title, req.Description, req.Priority, req.DueDate).Scan(
		&t.ID, &t.WorkspaceID, &t.AreaID, &t.GoalID, &t.Title, &t.Description,
		&t.Priority, &t.IsCompleted, &t.IsFocus, &t.DueDate, &t.CompletedAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *TaskService) Delete(ctx context.Context, wsID, id uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE tasks SET deleted_at = now() WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
	`, id, wsID)
	return err
}

func (s *TaskService) Complete(ctx context.Context, wsID, id uuid.UUID) (*domain.Task, error) {
	var t domain.Task
	err := s.db.QueryRow(ctx, `
		UPDATE tasks SET is_completed = true, completed_at = now(), updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		RETURNING id, workspace_id, area_id, goal_id, title, description, priority,
				  is_completed, is_focus, due_date, completed_at, created_at, updated_at
	`, id, wsID).Scan(
		&t.ID, &t.WorkspaceID, &t.AreaID, &t.GoalID, &t.Title, &t.Description,
		&t.Priority, &t.IsCompleted, &t.IsFocus, &t.DueDate, &t.CompletedAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *TaskService) ToggleFocus(ctx context.Context, wsID, id uuid.UUID) (*domain.Task, error) {
	var t domain.Task
	err := s.db.QueryRow(ctx, `
		UPDATE tasks SET is_focus = NOT is_focus, updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		RETURNING id, workspace_id, area_id, goal_id, title, description, priority,
				  is_completed, is_focus, due_date, completed_at, created_at, updated_at
	`, id, wsID).Scan(
		&t.ID, &t.WorkspaceID, &t.AreaID, &t.GoalID, &t.Title, &t.Description,
		&t.Priority, &t.IsCompleted, &t.IsFocus, &t.DueDate, &t.CompletedAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
