package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditEntry struct {
	ID          uuid.UUID       `json:"id"`
	WorkspaceID uuid.UUID       `json:"workspace_id"`
	UserID      uuid.UUID       `json:"user_id"`
	Action      string          `json:"action"`
	EntityType  string          `json:"entity_type"`
	EntityID    *uuid.UUID      `json:"entity_id,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	IPAddress   *string         `json:"ip_address,omitempty"`
	UserAgent   *string         `json:"user_agent,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}
