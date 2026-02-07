package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/widia-io/widia-omni/internal/domain"
)

type WorkspaceService struct {
	db *pgxpool.Pool
}

func NewWorkspaceService(db *pgxpool.Pool) *WorkspaceService {
	return &WorkspaceService{db: db}
}

func (s *WorkspaceService) GetWorkspace(ctx context.Context, wsID, userID uuid.UUID) (*domain.Workspace, error) {
	var w domain.Workspace
	err := s.db.QueryRow(ctx, `
		SELECT w.id, w.name, w.slug, w.owner_id, w.created_at, w.updated_at
		FROM workspaces w
		JOIN workspace_members wm ON wm.workspace_id = w.id
		WHERE w.id = $1 AND wm.user_id = $2
	`, wsID, userID).Scan(&w.ID, &w.Name, &w.Slug, &w.OwnerID, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

type UpdateWorkspaceRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (s *WorkspaceService) UpdateWorkspace(ctx context.Context, wsID uuid.UUID, req UpdateWorkspaceRequest) (*domain.Workspace, error) {
	var w domain.Workspace
	err := s.db.QueryRow(ctx, `
		UPDATE workspaces SET name = $2, slug = $3
		WHERE id = $1
		RETURNING id, name, slug, owner_id, created_at, updated_at
	`, wsID, req.Name, req.Slug).Scan(&w.ID, &w.Name, &w.Slug, &w.OwnerID, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

type WorkspaceUsage struct {
	Counters *domain.WorkspaceCounter  `json:"counters"`
	Limits   *domain.EntitlementLimits `json:"limits"`
}

func (s *WorkspaceService) GetUsage(ctx context.Context, wsID uuid.UUID) (*WorkspaceUsage, error) {
	var c domain.WorkspaceCounter
	err := s.db.QueryRow(ctx, `
		SELECT workspace_id, areas_count, goals_count, habits_count,
			   tasks_created_today, tasks_today_date, transactions_month_count,
			   transactions_month, storage_bytes_used, updated_at
		FROM workspace_counters WHERE workspace_id = $1
	`, wsID).Scan(
		&c.WorkspaceID, &c.AreasCount, &c.GoalsCount, &c.HabitsCount,
		&c.TasksCreatedToday, &c.TasksTodayDate, &c.TransactionsMonthCount,
		&c.TransactionsMonth, &c.StorageBytesUsed, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	var limitsRaw json.RawMessage
	err = s.db.QueryRow(ctx, `
		SELECT limits FROM workspace_entitlements
		WHERE workspace_id = $1 AND is_current = true
	`, wsID).Scan(&limitsRaw)
	if err != nil {
		return nil, err
	}

	limits, err := domain.ParseLimits(limitsRaw)
	if err != nil {
		return nil, err
	}

	return &WorkspaceUsage{Counters: &c, Limits: limits}, nil
}
