package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/widia-io/widia-omni/internal/domain"
	"github.com/widia-io/widia-omni/internal/observability"
)

type AreaService struct {
	db         *pgxpool.Pool
	counterSvc *CounterService
}

func NewAreaService(db *pgxpool.Pool, counterSvc *CounterService) *AreaService {
	return &AreaService{db: db, counterSvc: counterSvc}
}

func (s *AreaService) List(ctx context.Context, wsID uuid.UUID) ([]domain.LifeArea, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, name, slug, icon, color, weight, sort_order, is_active,
			   created_at, updated_at
		FROM life_areas
		WHERE workspace_id = $1 AND deleted_at IS NULL
		ORDER BY sort_order
	`, wsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []domain.LifeArea
	for rows.Next() {
		var a domain.LifeArea
		if err := rows.Scan(&a.ID, &a.WorkspaceID, &a.Name, &a.Slug, &a.Icon, &a.Color,
			&a.Weight, &a.SortOrder, &a.IsActive, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		areas = append(areas, a)
	}
	return areas, nil
}

type CreateAreaRequest struct {
	Name      string  `json:"name"`
	Slug      string  `json:"slug"`
	Icon      string  `json:"icon"`
	Color     string  `json:"color"`
	Weight    float64 `json:"weight"`
	SortOrder int     `json:"sort_order"`
}

func (s *AreaService) Create(ctx context.Context, wsID uuid.UUID, limits *domain.EntitlementLimits, req CreateAreaRequest) (*domain.LifeArea, error) {
	counters, err := s.counterSvc.GetCounters(ctx, wsID)
	if err != nil {
		return nil, err
	}
	if !limits.CanCreateArea(counters.AreasCount) {
		observability.EntitlementLimitReachedTotal.WithLabelValues("areas").Inc()
		return nil, errors.New("area limit reached")
	}

	var a domain.LifeArea
	err = s.db.QueryRow(ctx, `
		INSERT INTO life_areas (workspace_id, name, slug, icon, color, weight, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, workspace_id, name, slug, icon, color, weight, sort_order, is_active,
				  created_at, updated_at
	`, wsID, req.Name, req.Slug, req.Icon, req.Color, req.Weight, req.SortOrder).Scan(
		&a.ID, &a.WorkspaceID, &a.Name, &a.Slug, &a.Icon, &a.Color,
		&a.Weight, &a.SortOrder, &a.IsActive, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

type UpdateAreaRequest struct {
	Name     string  `json:"name"`
	Icon     string  `json:"icon"`
	Color    string  `json:"color"`
	Weight   float64 `json:"weight"`
	IsActive bool    `json:"is_active"`
}

func (s *AreaService) Update(ctx context.Context, wsID, id uuid.UUID, req UpdateAreaRequest) (*domain.LifeArea, error) {
	var a domain.LifeArea
	err := s.db.QueryRow(ctx, `
		UPDATE life_areas
		SET name = $3, icon = $4, color = $5, weight = $6, is_active = $7, updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		RETURNING id, workspace_id, name, slug, icon, color, weight, sort_order, is_active,
				  created_at, updated_at
	`, id, wsID, req.Name, req.Icon, req.Color, req.Weight, req.IsActive).Scan(
		&a.ID, &a.WorkspaceID, &a.Name, &a.Slug, &a.Icon, &a.Color,
		&a.Weight, &a.SortOrder, &a.IsActive, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *AreaService) Delete(ctx context.Context, wsID, id uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE life_areas SET deleted_at = now() WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
	`, id, wsID)
	return err
}

func (s *AreaService) Reorder(ctx context.Context, wsID, id uuid.UUID, sortOrder int) error {
	_, err := s.db.Exec(ctx, `
		UPDATE life_areas SET sort_order = $3, updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
	`, id, wsID, sortOrder)
	return err
}
