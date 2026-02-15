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

var taskSelectCols = `t.id, t.workspace_id, t.area_id, t.goal_id, t.parent_id, t.section_id,
	t.project_id, t.project_section_id,
	t.title, t.description, t.priority, t.position,
	t.is_completed, t.is_focus, t.due_date, t.duration_minutes,
	t.completed_at, t.created_at, t.updated_at`

var taskReturnCols = `id, workspace_id, area_id, goal_id, parent_id, section_id,
	project_id, project_section_id,
	title, description, priority, position,
	is_completed, is_focus, due_date, duration_minutes,
	completed_at, created_at, updated_at`

func scanTask(row interface{ Scan(dest ...any) error }) (domain.Task, error) {
	var t domain.Task
	err := row.Scan(&t.ID, &t.WorkspaceID, &t.AreaID, &t.GoalID, &t.ParentID, &t.SectionID,
		&t.ProjectID, &t.ProjectSectionID,
		&t.Title, &t.Description, &t.Priority, &t.Position,
		&t.IsCompleted, &t.IsFocus, &t.DueDate, &t.DurationMinutes,
		&t.CompletedAt, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

type TaskFilters struct {
	AreaID      *uuid.UUID `json:"area_id"`
	GoalID      *uuid.UUID `json:"goal_id"`
	SectionID   *uuid.UUID `json:"section_id"`
	ProjectID   *uuid.UUID `json:"project_id"`
	ParentID    *uuid.UUID `json:"parent_id"`
	LabelID     *uuid.UUID `json:"label_id"`
	IsCompleted *bool      `json:"is_completed"`
	HasParent   *bool      `json:"has_parent"`
	DueFrom     *string    `json:"due_from"`
	DueTo       *string    `json:"due_to"`
}

func (s *TaskService) List(ctx context.Context, wsID uuid.UUID, f TaskFilters) ([]domain.Task, error) {
	query := `SELECT ` + taskSelectCols + ` FROM tasks t WHERE t.workspace_id = $1 AND t.deleted_at IS NULL`
	args := []any{wsID}
	idx := 2

	if f.AreaID != nil {
		query += fmt.Sprintf(` AND t.area_id = $%d`, idx)
		args = append(args, *f.AreaID)
		idx++
	}
	if f.GoalID != nil {
		query += fmt.Sprintf(` AND t.goal_id = $%d`, idx)
		args = append(args, *f.GoalID)
		idx++
	}
	if f.SectionID != nil {
		query += fmt.Sprintf(` AND t.section_id = $%d`, idx)
		args = append(args, *f.SectionID)
		idx++
	}
	if f.ProjectID != nil {
		query += fmt.Sprintf(` AND t.project_id = $%d`, idx)
		args = append(args, *f.ProjectID)
		idx++
	}
	if f.ParentID != nil {
		query += fmt.Sprintf(` AND t.parent_id = $%d`, idx)
		args = append(args, *f.ParentID)
		idx++
	}
	if f.LabelID != nil {
		query += fmt.Sprintf(` AND EXISTS (SELECT 1 FROM task_labels tl WHERE tl.task_id = t.id AND tl.label_id = $%d)`, idx)
		args = append(args, *f.LabelID)
		idx++
	}
	if f.IsCompleted != nil {
		query += fmt.Sprintf(` AND t.is_completed = $%d`, idx)
		args = append(args, *f.IsCompleted)
		idx++
	}
	if f.HasParent != nil {
		if *f.HasParent {
			query += ` AND t.parent_id IS NOT NULL`
		} else {
			query += ` AND t.parent_id IS NULL`
		}
	}
	if f.DueFrom != nil {
		query += fmt.Sprintf(` AND t.due_date >= $%d`, idx)
		args = append(args, *f.DueFrom)
		idx++
	}
	if f.DueTo != nil {
		query += fmt.Sprintf(` AND t.due_date <= $%d`, idx)
		args = append(args, *f.DueTo)
		idx++
	}
	query += ` ORDER BY CASE WHEN t.is_focus THEN 0 ELSE 1 END, t.position ASC, t.created_at DESC`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []domain.Task
	taskIdx := map[uuid.UUID]int{}
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		taskIdx[t.ID] = len(tasks)
		tasks = append(tasks, t)
	}

	if len(tasks) == 0 {
		return tasks, nil
	}

	ids := make([]uuid.UUID, len(tasks))
	for i, t := range tasks {
		ids[i] = t.ID
	}

	labelRows, err := s.db.Query(ctx, `
		SELECT tl.task_id, l.id, l.workspace_id, l.name, l.color, l.position, l.created_at, l.updated_at
		FROM task_labels tl
		JOIN labels l ON l.id = tl.label_id AND l.deleted_at IS NULL
		WHERE tl.task_id = ANY($1)
	`, ids)
	if err != nil {
		return nil, err
	}
	defer labelRows.Close()

	for labelRows.Next() {
		var taskID uuid.UUID
		var l domain.Label
		if err := labelRows.Scan(&taskID, &l.ID, &l.WorkspaceID, &l.Name, &l.Color, &l.Position,
			&l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		if i, ok := taskIdx[taskID]; ok {
			tasks[i].Labels = append(tasks[i].Labels, l)
		}
	}

	return tasks, nil
}

type CreateTaskRequest struct {
	AreaID           *uuid.UUID          `json:"area_id"`
	GoalID           *uuid.UUID          `json:"goal_id"`
	ParentID         *uuid.UUID          `json:"parent_id"`
	SectionID        *uuid.UUID          `json:"section_id"`
	ProjectID        *uuid.UUID          `json:"project_id"`
	ProjectSectionID *uuid.UUID          `json:"project_section_id"`
	Title            string              `json:"title"`
	Description      *string             `json:"description"`
	Priority         domain.TaskPriority `json:"priority"`
	DueDate          *string             `json:"due_date"`
	IsFocus          bool                `json:"is_focus"`
	DurationMinutes  *int                `json:"duration_minutes"`
	LabelIDs         []uuid.UUID         `json:"label_ids"`
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
		INSERT INTO tasks (workspace_id, area_id, goal_id, parent_id, section_id,
			project_id, project_section_id,
			title, description, priority, due_date, is_focus, duration_minutes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING `+taskReturnCols,
		wsID, req.AreaID, req.GoalID, req.ParentID, req.SectionID,
		req.ProjectID, req.ProjectSectionID,
		req.Title, req.Description, req.Priority, req.DueDate, req.IsFocus, req.DurationMinutes,
	).Scan(&t.ID, &t.WorkspaceID, &t.AreaID, &t.GoalID, &t.ParentID, &t.SectionID,
		&t.ProjectID, &t.ProjectSectionID,
		&t.Title, &t.Description, &t.Priority, &t.Position,
		&t.IsCompleted, &t.IsFocus, &t.DueDate, &t.DurationMinutes,
		&t.CompletedAt, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if len(req.LabelIDs) > 0 {
		if err := s.setLabels(ctx, t.ID, req.LabelIDs); err != nil {
			return nil, err
		}
		t.Labels, _ = s.getTaskLabels(ctx, t.ID)
	}

	_ = s.counterSvc.IncrementTasksToday(ctx, wsID)
	return &t, nil
}

type UpdateTaskRequest struct {
	AreaID           *uuid.UUID          `json:"area_id"`
	GoalID           *uuid.UUID          `json:"goal_id"`
	SectionID        *uuid.UUID          `json:"section_id"`
	ProjectID        *uuid.UUID          `json:"project_id"`
	ProjectSectionID *uuid.UUID          `json:"project_section_id"`
	Title            string              `json:"title"`
	Description      *string             `json:"description"`
	Priority         domain.TaskPriority `json:"priority"`
	DueDate          *string             `json:"due_date"`
	DurationMinutes  *int                `json:"duration_minutes"`
	LabelIDs         *[]uuid.UUID        `json:"label_ids"`
}

func (s *TaskService) Update(ctx context.Context, wsID, id uuid.UUID, req UpdateTaskRequest) (*domain.Task, error) {
	var t domain.Task
	err := s.db.QueryRow(ctx, `
		UPDATE tasks
		SET area_id = $3, goal_id = $4, section_id = $5, project_id = $6, project_section_id = $7,
			title = $8, description = $9, priority = $10, due_date = $11, duration_minutes = $12,
			updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		RETURNING `+taskReturnCols,
		id, wsID, req.AreaID, req.GoalID, req.SectionID, req.ProjectID, req.ProjectSectionID,
		req.Title, req.Description, req.Priority, req.DueDate, req.DurationMinutes,
	).Scan(&t.ID, &t.WorkspaceID, &t.AreaID, &t.GoalID, &t.ParentID, &t.SectionID,
		&t.ProjectID, &t.ProjectSectionID,
		&t.Title, &t.Description, &t.Priority, &t.Position,
		&t.IsCompleted, &t.IsFocus, &t.DueDate, &t.DurationMinutes,
		&t.CompletedAt, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if req.LabelIDs != nil {
		if err := s.setLabels(ctx, id, *req.LabelIDs); err != nil {
			return nil, err
		}
		t.Labels, _ = s.getTaskLabels(ctx, id)
	}

	return &t, nil
}

func (s *TaskService) Delete(ctx context.Context, wsID, id uuid.UUID) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE tasks SET deleted_at = now() WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
	`, id, wsID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		_ = s.counterSvc.DecrementTasksToday(ctx, wsID)
	}
	return nil
}

func (s *TaskService) Complete(ctx context.Context, wsID, id uuid.UUID) (*domain.Task, error) {
	t, err := s.updateAndReturn(ctx, `
		UPDATE tasks SET is_completed = true, completed_at = now(), updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		RETURNING `+taskReturnCols, id, wsID)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TaskService) Reopen(ctx context.Context, wsID, id uuid.UUID) (*domain.Task, error) {
	t, err := s.updateAndReturn(ctx, `
		UPDATE tasks SET is_completed = false, completed_at = NULL, updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		RETURNING `+taskReturnCols, id, wsID)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TaskService) ToggleFocus(ctx context.Context, wsID, id uuid.UUID) (*domain.Task, error) {
	t, err := s.updateAndReturn(ctx, `
		UPDATE tasks SET is_focus = NOT is_focus, updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		RETURNING `+taskReturnCols, id, wsID)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TaskService) Reorder(ctx context.Context, wsID, id uuid.UUID, position int) error {
	_, err := s.db.Exec(ctx, `
		UPDATE tasks SET position = $3, updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
	`, id, wsID, position)
	return err
}

func (s *TaskService) updateAndReturn(ctx context.Context, query string, args ...any) (*domain.Task, error) {
	t, err := scanTask(s.db.QueryRow(ctx, query, args...))
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *TaskService) setLabels(ctx context.Context, taskID uuid.UUID, labelIDs []uuid.UUID) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM task_labels WHERE task_id = $1`, taskID)
	if err != nil {
		return err
	}

	for _, lid := range labelIDs {
		_, err = tx.Exec(ctx, `INSERT INTO task_labels (task_id, label_id) VALUES ($1, $2)`, taskID, lid)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *TaskService) getTaskLabels(ctx context.Context, taskID uuid.UUID) ([]domain.Label, error) {
	rows, err := s.db.Query(ctx, `
		SELECT l.id, l.workspace_id, l.name, l.color, l.position, l.created_at, l.updated_at
		FROM task_labels tl
		JOIN labels l ON l.id = tl.label_id AND l.deleted_at IS NULL
		WHERE tl.task_id = $1
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []domain.Label
	for rows.Next() {
		var l domain.Label
		if err := rows.Scan(&l.ID, &l.WorkspaceID, &l.Name, &l.Color, &l.Position,
			&l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		labels = append(labels, l)
	}
	return labels, nil
}
