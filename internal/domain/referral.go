package domain

import (
	"time"

	"github.com/google/uuid"
)

type ReferralAttributionStatus string

const (
	ReferralAttributionPending   ReferralAttributionStatus = "pending"
	ReferralAttributionConverted ReferralAttributionStatus = "converted"
	ReferralAttributionExpired   ReferralAttributionStatus = "expired"
)

type ReferralCreditStatus string

const (
	ReferralCreditAvailable ReferralCreditStatus = "available"
	ReferralCreditConsumed  ReferralCreditStatus = "consumed"
	ReferralCreditExpired   ReferralCreditStatus = "expired"
)

type ReferralCode struct {
	WorkspaceID   uuid.UUID  `json:"workspace_id"`
	Code          string     `json:"code"`
	CreatedBy     uuid.UUID  `json:"created_by"`
	CreatedAt     time.Time  `json:"created_at"`
	RegeneratedAt *time.Time `json:"regenerated_at,omitempty"`
}

type ReferralAttribution struct {
	ID                  uuid.UUID                 `json:"id"`
	ReferralCode        string                    `json:"referral_code"`
	ReferrerWorkspaceID uuid.UUID                 `json:"referrer_workspace_id"`
	ReferredWorkspaceID uuid.UUID                 `json:"referred_workspace_id"`
	ReferredUserID      *uuid.UUID                `json:"referred_user_id,omitempty"`
	ExpiresAt           time.Time                 `json:"expires_at"`
	Status              ReferralAttributionStatus `json:"status"`
	ConvertedAt         *time.Time                `json:"converted_at,omitempty"`
	CreatedAt           time.Time                 `json:"created_at"`
}

type ReferralCredit struct {
	ID            uuid.UUID            `json:"id"`
	AttributionID uuid.UUID            `json:"attribution_id"`
	WorkspaceID   uuid.UUID            `json:"workspace_id"`
	CreditType    string               `json:"credit_type"`
	Days          int                  `json:"days"`
	Status        ReferralCreditStatus `json:"status"`
	ExpiresAt     *time.Time           `json:"expires_at,omitempty"`
	ConsumedAt    *time.Time           `json:"consumed_at,omitempty"`
	CreatedAt     time.Time            `json:"created_at"`
}

type ReferralStats struct {
	Pending   int `json:"pending"`
	Converted int `json:"converted"`
	Expired   int `json:"expired"`
}

type ReferralMe struct {
	Code         string        `json:"code"`
	ShareURL     string        `json:"share_url"`
	Stats        ReferralStats `json:"stats"`
	CreditDays   int           `json:"credit_days"`
	HasAvailable bool          `json:"has_available"`
}
