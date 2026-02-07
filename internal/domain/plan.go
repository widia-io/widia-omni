package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type PlanTier string

const (
	TierFree    PlanTier = "free"
	TierPro     PlanTier = "pro"
	TierPremium PlanTier = "premium"
)

type Plan struct {
	ID                uuid.UUID       `json:"id"`
	Tier              PlanTier        `json:"tier"`
	Name              string          `json:"name"`
	PriceMonthly      float64         `json:"price_monthly"`
	PriceYearly       float64         `json:"price_yearly"`
	StripePriceMonthly *string        `json:"stripe_price_monthly,omitempty"`
	StripePriceYearly  *string        `json:"stripe_price_yearly,omitempty"`
	Limits            json.RawMessage `json:"limits"`
	IsActive          bool            `json:"is_active"`
	CreatedAt         time.Time       `json:"created_at"`
}

type SubscriptionStatus string

const (
	StatusTrialing SubscriptionStatus = "trialing"
	StatusActive   SubscriptionStatus = "active"
	StatusPastDue  SubscriptionStatus = "past_due"
	StatusCanceled SubscriptionStatus = "canceled"
	StatusPaused   SubscriptionStatus = "paused"
	StatusUnpaid   SubscriptionStatus = "unpaid"
)

type Subscription struct {
	ID                     uuid.UUID          `json:"id"`
	WorkspaceID            uuid.UUID          `json:"workspace_id"`
	PlanID                 uuid.UUID          `json:"plan_id"`
	Tier                   PlanTier           `json:"tier"`
	Status                 SubscriptionStatus `json:"status"`
	StripeCustomerID       *string            `json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID   *string            `json:"stripe_subscription_id,omitempty"`
	StripePriceID          *string            `json:"stripe_price_id,omitempty"`
	Currency               string             `json:"currency"`
	CurrentPeriodStart     *time.Time         `json:"current_period_start,omitempty"`
	CurrentPeriodEnd       *time.Time         `json:"current_period_end,omitempty"`
	TrialEnd               *time.Time         `json:"trial_end,omitempty"`
	CancelAt               *time.Time         `json:"cancel_at,omitempty"`
	CanceledAt             *time.Time         `json:"canceled_at,omitempty"`
	CreatedAt              time.Time          `json:"created_at"`
	UpdatedAt              time.Time          `json:"updated_at"`
}
