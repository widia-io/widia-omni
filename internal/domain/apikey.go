package domain

import (
	"time"

	"github.com/google/uuid"
)

type APIKey struct {
	ID          uuid.UUID  `json:"id"`
	WorkspaceID uuid.UUID  `json:"workspace_id"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	Name        string     `json:"name"`
	KeyPrefix   string     `json:"key_prefix"`
	KeyHash     string     `json:"-"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type APIKeyWithSecret struct {
	APIKey
	RawKey string `json:"key"`
}
