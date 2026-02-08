package domain

import (
	"time"

	"github.com/google/uuid"
)

type LifeArea struct {
	ID          uuid.UUID  `json:"id"`
	WorkspaceID uuid.UUID  `json:"workspace_id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Icon        string     `json:"icon"`
	Color       string     `json:"color"`
	Weight      float64    `json:"weight"`
	SortOrder   int        `json:"sort_order"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}
