package domain

import (
	"time"

	"github.com/google/uuid"
)

type ProjectStatus string

const (
	ProjectPlanning  ProjectStatus = "planning"
	ProjectActive    ProjectStatus = "active"
	ProjectPaused    ProjectStatus = "paused"
	ProjectCompleted ProjectStatus = "completed"
	ProjectCancelled ProjectStatus = "cancelled"
)

type Project struct {
	ID             uuid.UUID     `json:"id"`
	WorkspaceID    uuid.UUID     `json:"workspace_id"`
	AreaID         *uuid.UUID    `json:"area_id,omitempty"`
	GoalID         *uuid.UUID    `json:"goal_id,omitempty"`
	Title          string        `json:"title"`
	Description    *string       `json:"description,omitempty"`
	Status         ProjectStatus `json:"status"`
	Color          string        `json:"color"`
	Icon           string        `json:"icon"`
	StartDate      *time.Time    `json:"start_date,omitempty"`
	TargetDate     *time.Time    `json:"target_date,omitempty"`
	CompletedAt    *time.Time    `json:"completed_at,omitempty"`
	IsArchived     bool          `json:"is_archived"`
	ArchivedAt     *time.Time    `json:"archived_at,omitempty"`
	SortOrder      int           `json:"sort_order"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	DeletedAt      *time.Time    `json:"deleted_at,omitempty"`
	TasksTotal     int           `json:"tasks_total"`
	TasksCompleted int           `json:"tasks_completed"`
}

type ProjectSection struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"project_id"`
	Name      string    `json:"name"`
	Position  int       `json:"position"`
	Color     *string   `json:"color,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
