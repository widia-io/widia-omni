package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotifWeeklyReview NotificationType = "weekly_review"
	NotifStreakAtRisk NotificationType = "streak_at_risk"
	NotifGoalDeadline NotificationType = "goal_deadline"
	NotifTrialEnding  NotificationType = "trial_ending"
	NotifPlanChanged  NotificationType = "plan_changed"
	NotifScoreUpdate  NotificationType = "score_update"
	NotifHabitRemind  NotificationType = "habit_reminder"
	NotifSystem       NotificationType = "system"
)

type NotificationChannel string

const (
	ChannelInApp NotificationChannel = "in_app"
	ChannelEmail NotificationChannel = "email"
	ChannelPush  NotificationChannel = "push"
)

type Notification struct {
	ID          uuid.UUID           `json:"id"`
	WorkspaceID uuid.UUID           `json:"workspace_id"`
	UserID      uuid.UUID           `json:"user_id"`
	Type        NotificationType    `json:"type"`
	Channel     NotificationChannel `json:"channel"`
	Title       string              `json:"title"`
	Body        *string             `json:"body,omitempty"`
	Data        json.RawMessage     `json:"data,omitempty"`
	IsRead      bool                `json:"is_read"`
	ReadAt      *time.Time          `json:"read_at,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
}
