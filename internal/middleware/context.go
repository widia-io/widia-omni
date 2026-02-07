package middleware

import (
	"context"

	"github.com/google/uuid"
	"github.com/widia-io/widia-omni/internal/domain"
)

type ctxKey string

const (
	userIDKey       ctxKey = "user_id"
	workspaceIDKey  ctxKey = "workspace_id"
	roleKey         ctxKey = "workspace_role"
	entitlementsKey ctxKey = "entitlements"
)

func SetUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
}

func SetWorkspaceID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, workspaceIDKey, id)
}

func GetWorkspaceID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(workspaceIDKey).(uuid.UUID)
	return id, ok
}

func SetRole(ctx context.Context, role domain.WorkspaceRole) context.Context {
	return context.WithValue(ctx, roleKey, role)
}

func GetRole(ctx context.Context) (domain.WorkspaceRole, bool) {
	role, ok := ctx.Value(roleKey).(domain.WorkspaceRole)
	return role, ok
}

func SetEntitlements(ctx context.Context, ent *domain.EntitlementLimits) context.Context {
	return context.WithValue(ctx, entitlementsKey, ent)
}

func GetEntitlements(ctx context.Context) (*domain.EntitlementLimits, bool) {
	ent, ok := ctx.Value(entitlementsKey).(*domain.EntitlementLimits)
	return ent, ok
}
