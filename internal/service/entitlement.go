package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/widia-io/widia-omni/internal/domain"
)

type EntitlementService struct {
	db  *pgxpool.Pool
	rdb *redis.Client
}

func NewEntitlementService(db *pgxpool.Pool, rdb *redis.Client) *EntitlementService {
	return &EntitlementService{db: db, rdb: rdb}
}

func (s *EntitlementService) GetCurrent(ctx context.Context, wsID uuid.UUID) (*domain.Entitlement, error) {
	cacheKey := fmt.Sprintf("ws:%s:ent", wsID.String())

	if s.rdb != nil {
		cached, err := s.rdb.Get(ctx, cacheKey).Bytes()
		if err == nil {
			var ent domain.Entitlement
			if json.Unmarshal(cached, &ent) == nil {
				return &ent, nil
			}
		}
	}

	var ent domain.Entitlement
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, tier, limits, source, effective_from, effective_to, is_current, created_at
		FROM workspace_entitlements
		WHERE workspace_id = $1 AND is_current = true
	`, wsID).Scan(
		&ent.ID, &ent.WorkspaceID, &ent.Tier, &ent.Limits, &ent.Source,
		&ent.EffectiveFrom, &ent.EffectiveTo, &ent.IsCurrent, &ent.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		data, _ := json.Marshal(ent)
		s.rdb.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return &ent, nil
}

func (s *EntitlementService) InvalidateCache(ctx context.Context, wsID uuid.UUID) error {
	if s.rdb == nil {
		return nil
	}
	cacheKey := fmt.Sprintf("ws:%s:ent", wsID.String())
	tenantKey := fmt.Sprintf("tenant:*")
	_ = s.rdb.Del(ctx, cacheKey)
	// Also clear tenant cache entries that might reference this workspace
	iter := s.rdb.Scan(ctx, 0, tenantKey, 100).Iterator()
	for iter.Next(ctx) {
		s.rdb.Del(ctx, iter.Val())
	}
	return nil
}

func (s *EntitlementService) DeriveFromSubscription(ctx context.Context, wsID uuid.UUID, tier domain.PlanTier) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE workspace_entitlements
		SET is_current = false, effective_to = now()
		WHERE workspace_id = $1 AND is_current = true
	`, wsID)
	if err != nil {
		return err
	}

	var limitsRaw json.RawMessage
	err = tx.QueryRow(ctx, `SELECT limits FROM plans WHERE tier = $1`, tier).Scan(&limitsRaw)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO workspace_entitlements (workspace_id, tier, limits, source, is_current)
		VALUES ($1, $2, $3, 'stripe', true)
	`, wsID, tier, limitsRaw)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return s.InvalidateCache(ctx, wsID)
}
