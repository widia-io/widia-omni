package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type EntitlementSource string

const (
	SourceFree   EntitlementSource = "free"
	SourceStripe EntitlementSource = "stripe"
	SourceTrial  EntitlementSource = "trial"
	SourcePromo  EntitlementSource = "promo"
	SourceAdmin  EntitlementSource = "admin"
)

type Entitlement struct {
	ID            uuid.UUID         `json:"id"`
	WorkspaceID   uuid.UUID         `json:"workspace_id"`
	Tier          string            `json:"tier"`
	Limits        json.RawMessage   `json:"limits"`
	Source        EntitlementSource `json:"source"`
	EffectiveFrom time.Time         `json:"effective_from"`
	EffectiveTo   *time.Time        `json:"effective_to,omitempty"`
	IsCurrent     bool              `json:"is_current"`
	CreatedAt     time.Time         `json:"created_at"`
}

type EntitlementLimits struct {
	MaxAreas                int  `json:"max_areas"`
	MaxGoals                int  `json:"max_goals"`
	MaxHabits               int  `json:"max_habits"`
	MaxMembers              int  `json:"max_members"`
	MaxTasksPerDay          int  `json:"max_tasks_per_day"`
	MaxTransactionsPerMonth int  `json:"max_transactions_per_month"`
	JournalEnabled          bool `json:"journal_enabled"`
	FinanceEnabled          bool `json:"finance_enabled"`
	ExportEnabled           bool `json:"export_enabled"`
	FamilyEnabled           bool `json:"family_enabled"`
	ReferralEnabled         bool `json:"referral_enabled"`
	MobilePWAEnabled        bool `json:"mobile_pwa_enabled"`
	ScoreHistoryWeeks       int  `json:"score_history_weeks"`
	APIRateLimitPerMinute   int  `json:"api_rate_limit_per_minute"`
	StorageMB               int  `json:"storage_mb"`
	AIInsights              bool `json:"ai_insights"`
	APIAccess               bool `json:"api_access"`
}

func ParseLimits(raw json.RawMessage) (*EntitlementLimits, error) {
	var l EntitlementLimits
	if err := json.Unmarshal(raw, &l); err != nil {
		return nil, err
	}
	return &l, nil
}

func (l *EntitlementLimits) IsUnlimited(val int) bool {
	return val == -1
}

func (l *EntitlementLimits) CanCreateArea(currentCount int) bool {
	return l.IsUnlimited(l.MaxAreas) || currentCount < l.MaxAreas
}

func (l *EntitlementLimits) CanCreateGoal(currentCount int) bool {
	return l.IsUnlimited(l.MaxGoals) || currentCount < l.MaxGoals
}

func (l *EntitlementLimits) CanCreateHabit(currentCount int) bool {
	return l.IsUnlimited(l.MaxHabits) || currentCount < l.MaxHabits
}

func (l *EntitlementLimits) CanCreateTask(todayCount int) bool {
	return l.IsUnlimited(l.MaxTasksPerDay) || todayCount < l.MaxTasksPerDay
}

func (l *EntitlementLimits) CanCreateTransaction(monthCount int) bool {
	return l.IsUnlimited(l.MaxTransactionsPerMonth) || monthCount < l.MaxTransactionsPerMonth
}

func (l *EntitlementLimits) CanAddMember(currentCount int) bool {
	return l.IsUnlimited(l.MaxMembers) || currentCount < l.MaxMembers
}

func (l *EntitlementLimits) CanUseStorage(currentBytes, additionalBytes int64) bool {
	if l.IsUnlimited(l.StorageMB) {
		return true
	}
	limitBytes := int64(l.StorageMB) * 1024 * 1024
	return currentBytes+additionalBytes <= limitBytes
}
