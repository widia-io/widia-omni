package domain

import (
	"time"

	"github.com/google/uuid"
)

type GoalStatus string

const (
	GoalNotStarted GoalStatus = "not_started"
	GoalOnTrack    GoalStatus = "on_track"
	GoalAtRisk     GoalStatus = "at_risk"
	GoalBehind     GoalStatus = "behind"
	GoalCompleted  GoalStatus = "completed"
	GoalCancelled  GoalStatus = "cancelled"
)

type GoalPeriod string

const (
	PeriodYearly    GoalPeriod = "yearly"
	PeriodQuarterly GoalPeriod = "quarterly"
	PeriodMonthly   GoalPeriod = "monthly"
	PeriodWeekly    GoalPeriod = "weekly"
)

type Goal struct {
	ID           uuid.UUID  `json:"id"`
	WorkspaceID  uuid.UUID  `json:"workspace_id"`
	AreaID       *uuid.UUID `json:"area_id,omitempty"`
	ParentID     *uuid.UUID `json:"parent_id,omitempty"`
	Title        string     `json:"title"`
	Description  *string    `json:"description,omitempty"`
	Period       GoalPeriod `json:"period"`
	Status       GoalStatus `json:"status"`
	TargetValue  *float64   `json:"target_value,omitempty"`
	CurrentValue float64    `json:"current_value"`
	Unit         *string    `json:"unit,omitempty"`
	StartDate    time.Time  `json:"start_date"`
	EndDate      time.Time  `json:"end_date"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}
