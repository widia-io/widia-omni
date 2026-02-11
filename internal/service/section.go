package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/widia-io/widia-omni/internal/domain"
)

type SectionService struct {
	db *pgxpool.Pool
}

func NewSectionService(db *pgxpool.Pool) *SectionService {
	return &SectionService{db: db}
}

func (s *SectionService) List(ctx context.Context, wsID uuid.UUID, areaID *uuid.UUID) ([]domain.Section, error) {
	query := `
		SELECT id, workspace_id, area_id, name, position, created_at, updated_at
		FROM sections
		WHERE workspace_id = $1 AND deleted_at IS NULL
	`
	args := []any{wsID}
	if areaID != nil {
		query += ` AND area_id = $2`
		args = append(args, *areaID)
	}
	query += ` ORDER BY position, name`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sections []domain.Section
	for rows.Next() {
		var sec domain.Section
		if err := rows.Scan(&sec.ID, &sec.WorkspaceID, &sec.AreaID, &sec.Name, &sec.Position,
			&sec.CreatedAt, &sec.UpdatedAt); err != nil {
			return nil, err
		}
		sections = append(sections, sec)
	}
	return sections, nil
}

type CreateSectionRequest struct {
	AreaID   uuid.UUID `json:"area_id"`
	Name     string    `json:"name"`
	Position int       `json:"position"`
}

func (s *SectionService) Create(ctx context.Context, wsID uuid.UUID, req CreateSectionRequest) (*domain.Section, error) {
	var sec domain.Section
	err := s.db.QueryRow(ctx, `
		INSERT INTO sections (workspace_id, area_id, name, position)
		VALUES ($1, $2, $3, $4)
		RETURNING id, workspace_id, area_id, name, position, created_at, updated_at
	`, wsID, req.AreaID, req.Name, req.Position).Scan(
		&sec.ID, &sec.WorkspaceID, &sec.AreaID, &sec.Name, &sec.Position, &sec.CreatedAt, &sec.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &sec, nil
}

type UpdateSectionRequest struct {
	Name     string `json:"name"`
	Position int    `json:"position"`
}

func (s *SectionService) Update(ctx context.Context, wsID, id uuid.UUID, req UpdateSectionRequest) (*domain.Section, error) {
	var sec domain.Section
	err := s.db.QueryRow(ctx, `
		UPDATE sections
		SET name = $3, position = $4, updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		RETURNING id, workspace_id, area_id, name, position, created_at, updated_at
	`, id, wsID, req.Name, req.Position).Scan(
		&sec.ID, &sec.WorkspaceID, &sec.AreaID, &sec.Name, &sec.Position, &sec.CreatedAt, &sec.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &sec, nil
}

func (s *SectionService) Delete(ctx context.Context, wsID, id uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE sections SET deleted_at = now() WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
	`, id, wsID)
	return err
}

func (s *SectionService) Reorder(ctx context.Context, wsID, id uuid.UUID, position int) error {
	_, err := s.db.Exec(ctx, `
		UPDATE sections SET position = $3, updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
	`, id, wsID, position)
	return err
}
