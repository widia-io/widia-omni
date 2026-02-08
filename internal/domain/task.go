package domain

import (
	"time"

	"github.com/google/uuid"
)

type TaskPriority string

const (
	PriorityLow      TaskPriority = "low"
	PriorityMedium   TaskPriority = "medium"
	PriorityHigh     TaskPriority = "high"
	PriorityCritical TaskPriority = "critical"
)

type Task struct {
	ID          uuid.UUID    `json:"id"`
	WorkspaceID uuid.UUID    `json:"workspace_id"`
	AreaID      *uuid.UUID   `json:"area_id,omitempty"`
	GoalID      *uuid.UUID   `json:"goal_id,omitempty"`
	Title       string       `json:"title"`
	Description *string      `json:"description,omitempty"`
	Priority    TaskPriority `json:"priority"`
	IsCompleted bool         `json:"is_completed"`
	IsFocus     bool         `json:"is_focus"`
	DueDate     *time.Time   `json:"due_date,omitempty"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	DeletedAt   *time.Time   `json:"deleted_at,omitempty"`
}
