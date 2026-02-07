package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type UserProfile struct {
	ID                   uuid.UUID  `json:"id"`
	DisplayName          string     `json:"display_name"`
	Email                string     `json:"email"`
	AvatarURL            *string    `json:"avatar_url,omitempty"`
	Timezone             string     `json:"timezone"`
	Locale               string     `json:"locale"`
	DefaultWorkspaceID   *uuid.UUID `json:"default_workspace_id,omitempty"`
	OnboardingCompleted  bool       `json:"onboarding_completed"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type UserPreferences struct {
	UserID            uuid.UUID       `json:"user_id"`
	WeekStartsOn      int16          `json:"week_starts_on"`
	DailyFocusLimit    int16          `json:"daily_focus_limit"`
	NotificationEmail  bool           `json:"notification_email"`
	NotificationPush   bool           `json:"notification_push"`
	WeeklyReviewDay    int16          `json:"weekly_review_day"`
	WeeklyReviewTime   string         `json:"weekly_review_time"`
	Theme              string         `json:"theme"`
	Currency           string         `json:"currency"`
	ScoreWeights       json.RawMessage `json:"score_weights,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}
