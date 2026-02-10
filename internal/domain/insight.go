package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type InsightType string

const (
	InsightWeeklySummary InsightType = "weekly_summary"
	InsightOnDemand      InsightType = "on_demand"
)

type Insight struct {
	ID               uuid.UUID       `json:"id"`
	WorkspaceID      uuid.UUID       `json:"workspace_id"`
	Type             InsightType     `json:"type"`
	WeekStart        time.Time       `json:"week_start"`
	Content          json.RawMessage `json:"content"`
	Model            string          `json:"model"`
	PromptTokens     int             `json:"prompt_tokens"`
	CompletionTokens int             `json:"completion_tokens"`
	CreatedAt        time.Time       `json:"created_at"`
}

type InsightContent struct {
	Summary         string              `json:"summary"`
	Highlights      []string            `json:"highlights"`
	Concerns        []string            `json:"concerns"`
	Patterns        InsightPatterns     `json:"patterns"`
	Correlations    InsightCorrelations `json:"correlations"`
	Recommendations []Recommendation    `json:"recommendations"`
	AreaBreakdown   []AreaBreakdown     `json:"area_breakdown"`
}

type InsightPatterns struct {
	BestHabitDays      []string `json:"best_habit_days"`
	MoodTrend          string   `json:"mood_trend"`
	EnergyTrend        string   `json:"energy_trend"`
	TaskCompletionPct  float64  `json:"task_completion_pct"`
	HabitConsistencyPct float64 `json:"habit_consistency_pct"`
}

type InsightCorrelations struct {
	MoodProductivity string `json:"mood_productivity"`
	HabitScore       string `json:"habit_score"`
}

type Recommendation struct {
	Area     string `json:"area"`
	Action   string `json:"action"`
	Priority string `json:"priority"`
}

type AreaBreakdown struct {
	AreaName   string  `json:"area_name"`
	ScoreDelta float64 `json:"score_delta"`
	Summary    string  `json:"summary"`
	Suggestion string  `json:"suggestion"`
}
