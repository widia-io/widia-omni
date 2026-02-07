package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/widia-io/widia-omni/internal/domain"
	"github.com/widia-io/widia-omni/internal/observability"
)

const entitlementCacheTTL = 5 * time.Minute

type tenantData struct {
	WorkspaceID uuid.UUID             `json:"workspace_id"`
	Role        domain.WorkspaceRole  `json:"role"`
	Limits      *domain.EntitlementLimits `json:"limits"`
}

func Tenant(db *pgxpool.Pool, rdb *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := GetUserID(r.Context())
			if !ok {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			td, err := loadTenantData(r.Context(), db, rdb, userID)
			if err != nil {
				http.Error(w, `{"error":"workspace not found"}`, http.StatusForbidden)
				return
			}

			ctx := SetWorkspaceID(r.Context(), td.WorkspaceID)
			ctx = SetRole(ctx, td.Role)
			ctx = SetEntitlements(ctx, td.Limits)
			ctx = observability.WithWorkspaceID(ctx, td.WorkspaceID.String())

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func loadTenantData(ctx context.Context, db *pgxpool.Pool, rdb *redis.Client, userID uuid.UUID) (*tenantData, error) {
	cacheKey := fmt.Sprintf("tenant:%s", userID.String())

	if rdb != nil {
		cached, err := rdb.Get(ctx, cacheKey).Bytes()
		if err == nil {
			var td tenantData
			if json.Unmarshal(cached, &td) == nil {
				return &td, nil
			}
		}
	}

	var wsID uuid.UUID
	var role domain.WorkspaceRole
	err := db.QueryRow(ctx, `
		SELECT wm.workspace_id, wm.role
		FROM workspace_members wm
		JOIN user_profiles up ON up.default_workspace_id = wm.workspace_id
		WHERE wm.user_id = $1 AND up.id = $1
		LIMIT 1
	`, userID).Scan(&wsID, &role)
	if err != nil {
		return nil, fmt.Errorf("workspace lookup: %w", err)
	}

	var limitsRaw json.RawMessage
	err = db.QueryRow(ctx, `
		SELECT limits FROM workspace_entitlements
		WHERE workspace_id = $1 AND is_current = true
	`, wsID).Scan(&limitsRaw)
	if err != nil {
		return nil, fmt.Errorf("entitlement lookup: %w", err)
	}

	limits, err := domain.ParseLimits(limitsRaw)
	if err != nil {
		return nil, fmt.Errorf("parse limits: %w", err)
	}

	td := &tenantData{WorkspaceID: wsID, Role: role, Limits: limits}

	if rdb != nil {
		data, _ := json.Marshal(td)
		rdb.Set(ctx, cacheKey, data, entitlementCacheTTL)
	}

	return td, nil
}
