package domain

import (
	"time"

	"github.com/google/uuid"
)

type HabitFrequency string

const (
	FreqDaily  HabitFrequency = "daily"
	FreqWeekly HabitFrequency = "weekly"
	FreqCustom HabitFrequency = "custom"
)

type Habit struct {
	ID            uuid.UUID      `json:"id"`
	WorkspaceID   uuid.UUID      `json:"workspace_id"`
	AreaID        *uuid.UUID     `json:"area_id,omitempty"`
	Name          string         `json:"name"`
	Color         string         `json:"color"`
	Frequency     HabitFrequency `json:"frequency"`
	TargetPerWeek int            `json:"target_per_week"`
	IsActive      bool           `json:"is_active"`
	SortOrder     int            `json:"sort_order"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     *time.Time     `json:"deleted_at,omitempty"`
}

type HabitEntry struct {
	ID          uuid.UUID `json:"id"`
	HabitID     uuid.UUID `json:"habit_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Date        time.Time `json:"date"`
	Intensity   int16     `json:"intensity"`
	Notes       *string   `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type HabitStreak struct {
	HabitID       uuid.UUID `json:"habit_id"`
	Name          string    `json:"name"`
	CurrentStreak int       `json:"current_streak"`
	LongestStreak int       `json:"longest_streak"`
	LastCheckIn   *string   `json:"last_check_in,omitempty"`
}
