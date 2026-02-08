package domain

import (
	"time"

	"github.com/google/uuid"
)

type JournalEntry struct {
	ID          uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Date        time.Time `json:"date"`
	Mood        *int16    `json:"mood,omitempty"`
	Energy      *int16    `json:"energy,omitempty"`
	Wins        []string  `json:"wins,omitempty"`
	Challenges  []string  `json:"challenges,omitempty"`
	Gratitude   []string  `json:"gratitude,omitempty"`
	Notes       *string   `json:"notes,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
