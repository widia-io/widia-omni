package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/widia-io/widia-omni/internal/domain"
)

type LabelService struct {
	db *pgxpool.Pool
}

func NewLabelService(db *pgxpool.Pool) *LabelService {
	return &LabelService{db: db}
}

func (s *LabelService) List(ctx context.Context, wsID uuid.UUID) ([]domain.Label, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, name, color, position, created_at, updated_at
		FROM labels
		WHERE workspace_id = $1 AND deleted_at IS NULL
		ORDER BY position, name
	`, wsID)
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

type CreateLabelRequest struct {
	Name     string `json:"name"`
	Color    string `json:"color"`
	Position int    `json:"position"`
}

func (s *LabelService) Create(ctx context.Context, wsID uuid.UUID, req CreateLabelRequest) (*domain.Label, error) {
	var l domain.Label
	err := s.db.QueryRow(ctx, `
		INSERT INTO labels (workspace_id, name, color, position)
		VALUES ($1, $2, $3, $4)
		RETURNING id, workspace_id, name, color, position, created_at, updated_at
	`, wsID, req.Name, req.Color, req.Position).Scan(
		&l.ID, &l.WorkspaceID, &l.Name, &l.Color, &l.Position, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

type UpdateLabelRequest struct {
	Name     string `json:"name"`
	Color    string `json:"color"`
	Position int    `json:"position"`
}

func (s *LabelService) Update(ctx context.Context, wsID, id uuid.UUID, req UpdateLabelRequest) (*domain.Label, error) {
	var l domain.Label
	err := s.db.QueryRow(ctx, `
		UPDATE labels
		SET name = $3, color = $4, position = $5, updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		RETURNING id, workspace_id, name, color, position, created_at, updated_at
	`, id, wsID, req.Name, req.Color, req.Position).Scan(
		&l.ID, &l.WorkspaceID, &l.Name, &l.Color, &l.Position, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (s *LabelService) Delete(ctx context.Context, wsID, id uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE labels SET deleted_at = now() WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
	`, id, wsID)
	return err
}
