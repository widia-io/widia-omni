package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/widia-io/widia-omni/internal/domain"
	"github.com/widia-io/widia-omni/internal/observability"
)

type ProjectService struct {
	db         *pgxpool.Pool
	counterSvc *CounterService
}

func NewProjectService(db *pgxpool.Pool, counterSvc *CounterService) *ProjectService {
	return &ProjectService{db: db, counterSvc: counterSvc}
}

var projectSelectCols = `p.id, p.workspace_id, p.area_id, p.goal_id, p.title, p.description,
	p.status, p.color, p.icon, p.start_date, p.target_date, p.completed_at,
	p.is_archived, p.archived_at, p.sort_order, p.created_at, p.updated_at`

var projectReturnCols = `id, workspace_id, area_id, goal_id, title, description,
	status, color, icon, start_date, target_date, completed_at,
	is_archived, archived_at, sort_order, created_at, updated_at`

func scanProject(row interface{ Scan(dest ...any) error }) (domain.Project, error) {
	var p domain.Project
	err := row.Scan(&p.ID, &p.WorkspaceID, &p.AreaID, &p.GoalID, &p.Title, &p.Description,
		&p.Status, &p.Color, &p.Icon, &p.StartDate, &p.TargetDate, &p.CompletedAt,
		&p.IsArchived, &p.ArchivedAt, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func scanProjectWithCounts(row interface{ Scan(dest ...any) error }) (domain.Project, error) {
	var p domain.Project
	err := row.Scan(&p.ID, &p.WorkspaceID, &p.AreaID, &p.GoalID, &p.Title, &p.Description,
		&p.Status, &p.Color, &p.Icon, &p.StartDate, &p.TargetDate, &p.CompletedAt,
		&p.IsArchived, &p.ArchivedAt, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt,
		&p.TasksTotal, &p.TasksCompleted)
	return p, err
}

type ProjectFilters struct {
	AreaID         *uuid.UUID `json:"area_id"`
	GoalID         *uuid.UUID `json:"goal_id"`
	Status         *string    `json:"status"`
	IncludeArchived bool      `json:"include_archived"`
}

func (s *ProjectService) List(ctx context.Context, wsID uuid.UUID, f ProjectFilters) ([]domain.Project, error) {
	query := `SELECT ` + projectSelectCols + `,
		(SELECT COUNT(*) FROM tasks t WHERE t.project_id = p.id AND t.deleted_at IS NULL) AS tasks_total,
		(SELECT COUNT(*) FROM tasks t WHERE t.project_id = p.id AND t.deleted_at IS NULL AND t.is_completed = true) AS tasks_completed
		FROM projects p WHERE p.workspace_id = $1 AND p.deleted_at IS NULL`
	args := []any{wsID}
	idx := 2

	if !f.IncludeArchived {
		query += ` AND p.is_archived = false`
	}
	if f.AreaID != nil {
		query += fmt.Sprintf(` AND p.area_id = $%d`, idx)
		args = append(args, *f.AreaID)
		idx++
	}
	if f.GoalID != nil {
		query += fmt.Sprintf(` AND p.goal_id = $%d`, idx)
		args = append(args, *f.GoalID)
		idx++
	}
	if f.Status != nil {
		query += fmt.Sprintf(` AND p.status = $%d`, idx)
		args = append(args, *f.Status)
		idx++
	}
	query += ` ORDER BY p.sort_order, p.created_at DESC`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []domain.Project
	for rows.Next() {
		p, err := scanProjectWithCounts(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (s *ProjectService) GetByID(ctx context.Context, wsID, id uuid.UUID) (*domain.Project, error) {
	query := `SELECT ` + projectSelectCols + `,
		(SELECT COUNT(*) FROM tasks t WHERE t.project_id = p.id AND t.deleted_at IS NULL) AS tasks_total,
		(SELECT COUNT(*) FROM tasks t WHERE t.project_id = p.id AND t.deleted_at IS NULL AND t.is_completed = true) AS tasks_completed
		FROM projects p WHERE p.id = $1 AND p.workspace_id = $2 AND p.deleted_at IS NULL`
	p, err := scanProjectWithCounts(s.db.QueryRow(ctx, query, id, wsID))
	if err != nil {
		return nil, err
	}
	return &p, nil
}

type CreateProjectRequest struct {
	AreaID      *uuid.UUID `json:"area_id"`
	GoalID      *uuid.UUID `json:"goal_id"`
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	Color       string     `json:"color"`
	Icon        string     `json:"icon"`
	StartDate   *string    `json:"start_date"`
	TargetDate  *string    `json:"target_date"`
}

func (s *ProjectService) Create(ctx context.Context, wsID uuid.UUID, limits *domain.EntitlementLimits, req CreateProjectRequest) (*domain.Project, error) {
	counters, err := s.counterSvc.GetCounters(ctx, wsID)
	if err != nil {
		return nil, err
	}
	if !limits.CanCreateProject(counters.ProjectsCount) {
		observability.EntitlementLimitReachedTotal.WithLabelValues("projects").Inc()
		return nil, errors.New("project limit reached")
	}

	if req.Color == "" {
		req.Color = "blue"
	}
	if req.Icon == "" {
		req.Icon = "folder"
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var p domain.Project
	err = tx.QueryRow(ctx, `
		INSERT INTO projects (workspace_id, area_id, goal_id, title, description, color, icon, start_date, target_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING `+projectReturnCols,
		wsID, req.AreaID, req.GoalID, req.Title, req.Description, req.Color, req.Icon, req.StartDate, req.TargetDate,
	).Scan(&p.ID, &p.WorkspaceID, &p.AreaID, &p.GoalID, &p.Title, &p.Description,
		&p.Status, &p.Color, &p.Icon, &p.StartDate, &p.TargetDate, &p.CompletedAt,
		&p.IsArchived, &p.ArchivedAt, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	defaultSections := []struct {
		name     string
		position int
	}{
		{"To Do", 0},
		{"In Progress", 1},
		{"Done", 2},
	}
	for _, sec := range defaultSections {
		_, err = tx.Exec(ctx, `INSERT INTO project_sections (project_id, name, position) VALUES ($1, $2, $3)`,
			p.ID, sec.name, sec.position)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &p, nil
}

type UpdateProjectRequest struct {
	AreaID      *uuid.UUID `json:"area_id"`
	GoalID      *uuid.UUID `json:"goal_id"`
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	Status      string     `json:"status"`
	Color       string     `json:"color"`
	Icon        string     `json:"icon"`
	StartDate   *string    `json:"start_date"`
	TargetDate  *string    `json:"target_date"`
}

func (s *ProjectService) Update(ctx context.Context, wsID, id uuid.UUID, req UpdateProjectRequest) (*domain.Project, error) {
	var completedAt *time.Time
	if domain.ProjectStatus(req.Status) == domain.ProjectCompleted {
		now := time.Now()
		completedAt = &now
	}

	var p domain.Project
	err := s.db.QueryRow(ctx, `
		UPDATE projects
		SET area_id = $3, goal_id = $4, title = $5, description = $6, status = $7,
			color = $8, icon = $9, start_date = $10, target_date = $11, completed_at = $12
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		RETURNING `+projectReturnCols,
		id, wsID, req.AreaID, req.GoalID, req.Title, req.Description, req.Status,
		req.Color, req.Icon, req.StartDate, req.TargetDate, completedAt,
	).Scan(&p.ID, &p.WorkspaceID, &p.AreaID, &p.GoalID, &p.Title, &p.Description,
		&p.Status, &p.Color, &p.Icon, &p.StartDate, &p.TargetDate, &p.CompletedAt,
		&p.IsArchived, &p.ArchivedAt, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *ProjectService) Delete(ctx context.Context, wsID, id uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE projects SET deleted_at = now() WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
	`, id, wsID)
	return err
}

func (s *ProjectService) Reorder(ctx context.Context, wsID, id uuid.UUID, sortOrder int) error {
	_, err := s.db.Exec(ctx, `
		UPDATE projects SET sort_order = $3 WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
	`, id, wsID, sortOrder)
	return err
}

func (s *ProjectService) Archive(ctx context.Context, wsID, id uuid.UUID) (*domain.Project, error) {
	p, err := scanProject(s.db.QueryRow(ctx, `
		UPDATE projects SET is_archived = true, archived_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		RETURNING `+projectReturnCols, id, wsID))
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *ProjectService) Unarchive(ctx context.Context, wsID, id uuid.UUID) (*domain.Project, error) {
	p, err := scanProject(s.db.QueryRow(ctx, `
		UPDATE projects SET is_archived = false, archived_at = NULL
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		RETURNING `+projectReturnCols, id, wsID))
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// --- Project Sections ---

func (s *ProjectService) ListSections(ctx context.Context, wsID, projectID uuid.UUID) ([]domain.ProjectSection, error) {
	rows, err := s.db.Query(ctx, `
		SELECT ps.id, ps.project_id, ps.name, ps.position, ps.color, ps.created_at, ps.updated_at
		FROM project_sections ps
		JOIN projects p ON p.id = ps.project_id
		WHERE ps.project_id = $1 AND p.workspace_id = $2 AND p.deleted_at IS NULL
		ORDER BY ps.position
	`, projectID, wsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sections []domain.ProjectSection
	for rows.Next() {
		var sec domain.ProjectSection
		if err := rows.Scan(&sec.ID, &sec.ProjectID, &sec.Name, &sec.Position, &sec.Color,
			&sec.CreatedAt, &sec.UpdatedAt); err != nil {
			return nil, err
		}
		sections = append(sections, sec)
	}
	return sections, nil
}

type CreateProjectSectionRequest struct {
	Name  string  `json:"name"`
	Color *string `json:"color"`
}

func (s *ProjectService) CreateSection(ctx context.Context, wsID, projectID uuid.UUID, req CreateProjectSectionRequest) (*domain.ProjectSection, error) {
	var sec domain.ProjectSection
	err := s.db.QueryRow(ctx, `
		INSERT INTO project_sections (project_id, name, color, position)
		SELECT $1, $2, $3, COALESCE(MAX(position), -1) + 1
		FROM project_sections WHERE project_id = $1
		RETURNING id, project_id, name, position, color, created_at, updated_at
	`, projectID, req.Name, req.Color).Scan(
		&sec.ID, &sec.ProjectID, &sec.Name, &sec.Position, &sec.Color,
		&sec.CreatedAt, &sec.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &sec, nil
}

type UpdateProjectSectionRequest struct {
	Name  string  `json:"name"`
	Color *string `json:"color"`
}

func (s *ProjectService) UpdateSection(ctx context.Context, wsID, projectID, sectionID uuid.UUID, req UpdateProjectSectionRequest) (*domain.ProjectSection, error) {
	var sec domain.ProjectSection
	err := s.db.QueryRow(ctx, `
		UPDATE project_sections ps SET name = $4, color = $5
		FROM projects p
		WHERE ps.id = $1 AND ps.project_id = $2 AND p.id = ps.project_id AND p.workspace_id = $3 AND p.deleted_at IS NULL
		RETURNING ps.id, ps.project_id, ps.name, ps.position, ps.color, ps.created_at, ps.updated_at
	`, sectionID, projectID, wsID, req.Name, req.Color).Scan(
		&sec.ID, &sec.ProjectID, &sec.Name, &sec.Position, &sec.Color,
		&sec.CreatedAt, &sec.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &sec, nil
}

func (s *ProjectService) DeleteSection(ctx context.Context, wsID, projectID, sectionID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		DELETE FROM project_sections ps
		USING projects p
		WHERE ps.id = $1 AND ps.project_id = $2 AND p.id = ps.project_id AND p.workspace_id = $3 AND p.deleted_at IS NULL
	`, sectionID, projectID, wsID)
	return err
}

func (s *ProjectService) ReorderSection(ctx context.Context, wsID, projectID, sectionID uuid.UUID, position int) error {
	_, err := s.db.Exec(ctx, `
		UPDATE project_sections ps SET position = $4
		FROM projects p
		WHERE ps.id = $1 AND ps.project_id = $2 AND p.id = ps.project_id AND p.workspace_id = $3 AND p.deleted_at IS NULL
	`, sectionID, projectID, wsID, position)
	return err
}
