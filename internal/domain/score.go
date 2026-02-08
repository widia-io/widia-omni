package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AreaScore struct {
	ID          uuid.UUID       `json:"id"`
	WorkspaceID uuid.UUID       `json:"workspace_id"`
	AreaID      uuid.UUID       `json:"area_id"`
	Score       int16           `json:"score"`
	WeekStart   time.Time       `json:"week_start"`
	Breakdown   json.RawMessage `json:"breakdown,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

type LifeScore struct {
	ID          uuid.UUID       `json:"id"`
	WorkspaceID uuid.UUID       `json:"workspace_id"`
	Score       int16           `json:"score"`
	WeekStart   time.Time       `json:"week_start"`
	AreaScores  json.RawMessage `json:"area_scores"`
	CreatedAt   time.Time       `json:"created_at"`
}

type ScoreBreakdown struct {
	HabitsScore float64 `json:"habits_score"`
	GoalsScore  float64 `json:"goals_score"`
	TasksScore  float64 `json:"tasks_score"`
}

type ScoreHistory struct {
	LifeScores []LifeScore `json:"life_scores"`
	AreaScores []AreaScore `json:"area_scores"`
}
