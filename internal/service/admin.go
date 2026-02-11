package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/widia-io/widia-omni/internal/domain"
)

type AdminService struct {
	db     *pgxpool.Pool
	entSvc *EntitlementService
}

func NewAdminService(db *pgxpool.Pool, entSvc *EntitlementService) *AdminService {
	return &AdminService{db: db, entSvc: entSvc}
}

type AdminMetrics struct {
	TotalUsers      int            `json:"total_users"`
	TotalWorkspaces int            `json:"total_workspaces"`
	ActiveSubs      map[string]int `json:"active_subscriptions"`
}

func (s *AdminService) GetMetrics(ctx context.Context) (*AdminMetrics, error) {
	m := &AdminMetrics{ActiveSubs: make(map[string]int)}

	err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM user_profiles`).Scan(&m.TotalUsers)
	if err != nil {
		return nil, err
	}
	err = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM workspaces`).Scan(&m.TotalWorkspaces)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, `
		SELECT tier, COUNT(*) FROM subscriptions
		WHERE status IN ('active', 'trialing')
		GROUP BY tier
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tier string
		var count int
		if err := rows.Scan(&tier, &count); err != nil {
			return nil, err
		}
		m.ActiveSubs[tier] = count
	}
	return m, nil
}

type AdminUser struct {
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Tier        string    `json:"tier"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

func (s *AdminService) ListUsers(ctx context.Context, limit, offset int) ([]AdminUser, int, error) {
	if limit <= 0 {
		limit = 50
	}

	var total int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM user_profiles`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(ctx, `
		SELECT up.id, up.email, up.display_name, wm.workspace_id,
			   COALESCE(sub.tier, 'free'), COALESCE(sub.status, 'none'), up.created_at
		FROM user_profiles up
		JOIN workspace_members wm ON wm.user_id = up.id AND wm.role = 'owner'
		LEFT JOIN subscriptions sub ON sub.workspace_id = wm.workspace_id
		ORDER BY up.created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []AdminUser
	for rows.Next() {
		var u AdminUser
		if err := rows.Scan(&u.UserID, &u.Email, &u.DisplayName, &u.WorkspaceID,
			&u.Tier, &u.Status, &u.CreatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, nil
}

type AdminUserDetail struct {
	AdminUser
	Counters    *domain.WorkspaceCounter `json:"counters"`
	Entitlement *domain.Entitlement      `json:"entitlement,omitempty"`
}

func (s *AdminService) GetUser(ctx context.Context, userID uuid.UUID) (*AdminUserDetail, error) {
	var u AdminUserDetail
	err := s.db.QueryRow(ctx, `
		SELECT up.id, up.email, up.display_name, wm.workspace_id,
			   COALESCE(sub.tier, 'free'), COALESCE(sub.status, 'none'), up.created_at
		FROM user_profiles up
		JOIN workspace_members wm ON wm.user_id = up.id AND wm.role = 'owner'
		LEFT JOIN subscriptions sub ON sub.workspace_id = wm.workspace_id
		WHERE up.id = $1
	`, userID).Scan(&u.UserID, &u.Email, &u.DisplayName, &u.WorkspaceID,
		&u.Tier, &u.Status, &u.CreatedAt)
	if err != nil {
		return nil, err
	}

	var c domain.WorkspaceCounter
	err = s.db.QueryRow(ctx, `
		SELECT workspace_id, areas_count, goals_count, habits_count, members_count,
			   tasks_created_today, tasks_today_date, transactions_month_count,
			   transactions_month, storage_bytes_used, updated_at
		FROM workspace_counters WHERE workspace_id = $1
	`, u.WorkspaceID).Scan(
		&c.WorkspaceID, &c.AreasCount, &c.GoalsCount, &c.HabitsCount, &c.MembersCount,
		&c.TasksCreatedToday, &c.TasksTodayDate, &c.TransactionsMonthCount,
		&c.TransactionsMonth, &c.StorageBytesUsed, &c.UpdatedAt,
	)
	if err == nil {
		u.Counters = &c
	}

	ent, err := s.entSvc.GetCurrent(ctx, u.WorkspaceID)
	if err == nil {
		u.Entitlement = ent
	}

	return &u, nil
}

type AdminWorkspaceUsage struct {
	WorkspaceID uuid.UUID                 `json:"workspace_id"`
	Counters    *domain.WorkspaceCounter  `json:"counters"`
	Limits      *domain.EntitlementLimits `json:"limits,omitempty"`
}

func (s *AdminService) GetWorkspaceUsage(ctx context.Context, wsID uuid.UUID) (*AdminWorkspaceUsage, error) {
	usage := &AdminWorkspaceUsage{WorkspaceID: wsID}

	var c domain.WorkspaceCounter
	err := s.db.QueryRow(ctx, `
		SELECT workspace_id, areas_count, goals_count, habits_count, members_count,
			   tasks_created_today, tasks_today_date, transactions_month_count,
			   transactions_month, storage_bytes_used, updated_at
		FROM workspace_counters WHERE workspace_id = $1
	`, wsID).Scan(
		&c.WorkspaceID, &c.AreasCount, &c.GoalsCount, &c.HabitsCount, &c.MembersCount,
		&c.TasksCreatedToday, &c.TasksTodayDate, &c.TransactionsMonthCount,
		&c.TransactionsMonth, &c.StorageBytesUsed, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	usage.Counters = &c

	ent, err := s.entSvc.GetCurrent(ctx, wsID)
	if err == nil {
		limits, err := domain.ParseLimits(ent.Limits)
		if err == nil {
			usage.Limits = limits
		}
	}

	return usage, nil
}

type OverrideEntitlementRequest struct {
	WorkspaceID uuid.UUID       `json:"workspace_id"`
	Tier        string          `json:"tier"`
	Limits      json.RawMessage `json:"limits"`
}

func (s *AdminService) OverrideEntitlement(ctx context.Context, req OverrideEntitlementRequest) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE workspace_entitlements
		SET is_current = false, effective_to = now()
		WHERE workspace_id = $1 AND is_current = true
	`, req.WorkspaceID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO workspace_entitlements (workspace_id, tier, limits, source, is_current)
		VALUES ($1, $2, $3, 'admin', true)
	`, req.WorkspaceID, req.Tier, req.Limits)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return s.entSvc.InvalidateCache(ctx, req.WorkspaceID)
}
