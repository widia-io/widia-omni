package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v81/checkout/session"
	stripesub "github.com/stripe/stripe-go/v81/subscription"
	"github.com/stripe/stripe-go/v81/webhook"
	"github.com/widia-io/widia-omni/internal/domain"
)

type BillingService struct {
	db             *pgxpool.Pool
	entitlementSvc *EntitlementService
	referralSvc    *ReferralService
	webhookSecret  string
	successURL     string
	cancelURL      string
}

var ErrBillingPlanUnavailable = errors.New("plan not found or no stripe price configured")

func NewBillingService(db *pgxpool.Pool, entSvc *EntitlementService, referralSvc *ReferralService, stripeKey, webhookSecret, successURL, cancelURL string) *BillingService {
	stripe.Key = stripeKey
	return &BillingService{
		db:             db,
		entitlementSvc: entSvc,
		referralSvc:    referralSvc,
		webhookSecret:  webhookSecret,
		successURL:     successURL,
		cancelURL:      cancelURL,
	}
}

func (s *BillingService) ListPlans(ctx context.Context) ([]domain.Plan, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, tier, name, price_monthly, price_yearly, stripe_price_monthly,
			   stripe_price_yearly, limits, is_active, created_at
		FROM plans WHERE is_active = true ORDER BY price_monthly
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []domain.Plan
	for rows.Next() {
		var p domain.Plan
		if err := rows.Scan(&p.ID, &p.Tier, &p.Name, &p.PriceMonthly, &p.PriceYearly,
			&p.StripePriceMonthly, &p.StripePriceYearly, &p.Limits, &p.IsActive, &p.CreatedAt); err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}
	return plans, nil
}

func (s *BillingService) GetSubscription(ctx context.Context, wsID uuid.UUID) (*domain.Subscription, error) {
	var sub domain.Subscription
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, plan_id, tier, status, stripe_customer_id, stripe_subscription_id,
			   stripe_price_id, currency, current_period_start, current_period_end, trial_end,
			   cancel_at, canceled_at, created_at, updated_at
		FROM subscriptions
		WHERE workspace_id = $1 AND status IN ('trialing', 'active', 'past_due')
	`, wsID).Scan(
		&sub.ID, &sub.WorkspaceID, &sub.PlanID, &sub.Tier, &sub.Status,
		&sub.StripeCustomerID, &sub.StripeSubscriptionID, &sub.StripePriceID,
		&sub.Currency, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd, &sub.TrialEnd,
		&sub.CancelAt, &sub.CanceledAt, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (s *BillingService) CreateCheckoutSession(ctx context.Context, wsID uuid.UUID, tier string, interval string) (string, error) {
	var priceID *string
	col := "stripe_price_monthly"
	if interval == "yearly" {
		col = "stripe_price_yearly"
	}
	err := s.db.QueryRow(ctx, fmt.Sprintf(`SELECT %s FROM plans WHERE tier = $1 AND is_active = true`, col), tier).Scan(&priceID)
	if err != nil || priceID == nil {
		return "", ErrBillingPlanUnavailable
	}

	var customerID *string
	_ = s.db.QueryRow(ctx, `SELECT stripe_customer_id FROM subscriptions WHERE workspace_id = $1 LIMIT 1`, wsID).Scan(&customerID)

	params := &stripe.CheckoutSessionParams{
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(s.successURL),
		CancelURL:  stripe.String(s.cancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{Price: priceID, Quantity: stripe.Int64(1)},
		},
		Metadata: map[string]string{
			"workspace_id": wsID.String(),
			"tier":         tier,
			"interval":     interval,
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"workspace_id": wsID.String(),
				"tier":         tier,
				"interval":     interval,
			},
		},
	}
	if customerID != nil {
		params.Customer = customerID
	}

	sess, err := checkoutsession.New(params)
	if err != nil {
		return "", err
	}
	return sess.URL, nil
}

func (s *BillingService) CreatePortalSession(ctx context.Context, wsID uuid.UUID) (string, error) {
	var customerID *string
	err := s.db.QueryRow(ctx, `SELECT stripe_customer_id FROM subscriptions WHERE workspace_id = $1 AND stripe_customer_id IS NOT NULL LIMIT 1`, wsID).Scan(&customerID)
	if err != nil || customerID == nil {
		return "", errors.New("no stripe customer found")
	}

	params := &stripe.BillingPortalSessionParams{
		Customer: customerID,
	}
	sess, err := session.New(params)
	if err != nil {
		return "", err
	}
	return sess.URL, nil
}

func (s *BillingService) HandleWebhook(ctx context.Context, body io.Reader, signature string) error {
	payload, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	event, err := webhook.ConstructEventWithOptions(payload, signature, s.webhookSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		return fmt.Errorf("invalid signature: %w", err)
	}

	var processed bool
	err = s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM stripe_events_processed WHERE event_id = $1)`, event.ID).Scan(&processed)
	if err != nil {
		return err
	}
	if processed {
		return nil
	}

	switch event.Type {
	case "checkout.session.completed":
		err = s.handleCheckoutCompleted(ctx, event)
	case "customer.subscription.created":
		err = s.handleSubscriptionUpdated(ctx, event)
	case "customer.subscription.updated":
		err = s.handleSubscriptionUpdated(ctx, event)
	case "customer.subscription.deleted":
		err = s.handleSubscriptionDeleted(ctx, event)
	}
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, `INSERT INTO stripe_events_processed (event_id, event_type) VALUES ($1, $2)`, event.ID, event.Type)
	return err
}

func (s *BillingService) handleCheckoutCompleted(ctx context.Context, event stripe.Event) error {
	var checkoutObj stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &checkoutObj); err != nil {
		return fmt.Errorf("decode checkout session event: %w", err)
	}
	if checkoutObj.ID == "" {
		return errors.New("missing checkout session id")
	}

	params := &stripe.CheckoutSessionParams{}
	params.AddExpand("subscription")
	params.AddExpand("customer")

	sess, err := checkoutsession.Get(checkoutObj.ID, params)
	if err != nil {
		return err
	}

	wsIDStr, ok := sess.Metadata["workspace_id"]
	if !ok {
		return errors.New("missing workspace_id in metadata")
	}
	wsID, err := uuid.Parse(wsIDStr)
	if err != nil {
		return err
	}

	if sess.Subscription == nil {
		return errors.New("missing subscription in checkout session")
	}
	sub, err := stripesub.Get(sess.Subscription.ID, nil)
	if err != nil {
		return err
	}

	priceID, hasPrice := firstSubscriptionPriceID(sub)
	if !hasPrice {
		return errors.New("missing price in stripe subscription")
	}
	planID, tier, err := s.resolvePlanByPriceID(ctx, priceID)
	if err != nil {
		fallbackTier, tierErr := parsePlanTier(sess.Metadata["tier"])
		if tierErr != nil {
			return err
		}
		planID, err = s.resolvePlanIDByTier(ctx, fallbackTier)
		if err != nil {
			return err
		}
		tier = fallbackTier
	}
	status := normalizeStripeStatus(sub.Status)
	customerID := stringOrNilPtr(extractCustomerID(sess.Customer, sub.Customer))

	_, err = s.db.Exec(ctx, `
		UPDATE subscriptions
		SET plan_id = $2,
			tier = $3,
			stripe_customer_id = $4,
			stripe_subscription_id = $5,
			stripe_price_id = $6,
			status = $7,
			current_period_start = $8,
			current_period_end = $9,
			trial_end = $10,
			cancel_at = $11,
			canceled_at = $12,
			updated_at = now()
		WHERE workspace_id = $1
	`, wsID, planID, tier, customerID, sub.ID, priceID, status,
		unixToTimePtr(sub.CurrentPeriodStart), unixToTimePtr(sub.CurrentPeriodEnd),
		unixToTimePtr(sub.TrialEnd), unixToTimePtr(sub.CancelAt), unixToTimePtr(sub.CanceledAt))
	if err != nil {
		return err
	}

	return s.syncEntitlementsAndReferral(ctx, wsID, tier, status)
}

func (s *BillingService) handleSubscriptionUpdated(ctx context.Context, event stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return fmt.Errorf("decode subscription event: %w", err)
	}
	if sub.ID == "" {
		return errors.New("missing subscription id")
	}

	wsID, err := s.resolveWorkspaceIDForSubscription(ctx, sub.ID, sub.Metadata)
	if err != nil {
		return err
	}

	var (
		planID uuid.UUID
		tier   domain.PlanTier
	)
	priceID, hasPrice := firstSubscriptionPriceID(&sub)
	if hasPrice {
		planID, tier, err = s.resolvePlanByPriceID(ctx, priceID)
		if err != nil {
			fallbackTier, tierErr := parsePlanTier(sub.Metadata["tier"])
			if tierErr != nil {
				return err
			}
			planID, err = s.resolvePlanIDByTier(ctx, fallbackTier)
			if err != nil {
				return err
			}
			tier = fallbackTier
		}
	} else {
		err = s.db.QueryRow(ctx, `SELECT plan_id, tier FROM subscriptions WHERE workspace_id = $1 LIMIT 1`, wsID).Scan(&planID, &tier)
		if err != nil {
			return err
		}
	}

	status := normalizeStripeStatus(sub.Status)
	customerID := stringOrNilPtr(extractCustomerID(sub.Customer))
	var stripePriceID *string
	if hasPrice {
		stripePriceID = &priceID
	}

	_, err = s.db.Exec(ctx, `
		UPDATE subscriptions
		SET plan_id = $2,
			tier = $3,
			stripe_customer_id = $4,
			stripe_subscription_id = $5,
			stripe_price_id = $6,
			status = $7,
			current_period_start = $8,
			current_period_end = $9,
			trial_end = $10,
			cancel_at = $11,
			canceled_at = $12,
			updated_at = now()
		WHERE workspace_id = $1
	`, wsID, planID, tier, customerID, sub.ID, stripePriceID, status,
		unixToTimePtr(sub.CurrentPeriodStart), unixToTimePtr(sub.CurrentPeriodEnd),
		unixToTimePtr(sub.TrialEnd), unixToTimePtr(sub.CancelAt), unixToTimePtr(sub.CanceledAt))
	if err != nil {
		return err
	}

	return s.syncEntitlementsAndReferral(ctx, wsID, tier, status)
}

func (s *BillingService) handleSubscriptionDeleted(ctx context.Context, event stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return fmt.Errorf("decode subscription deleted event: %w", err)
	}
	if sub.ID == "" {
		return errors.New("missing subscription id")
	}

	wsID, err := s.resolveWorkspaceIDForSubscription(ctx, sub.ID, sub.Metadata)
	if err != nil {
		return err
	}
	freePlanID, err := s.resolvePlanIDByTier(ctx, domain.TierFree)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, `
		UPDATE subscriptions
		SET plan_id = $2,
			tier = 'free',
			status = 'canceled',
			stripe_price_id = NULL,
			canceled_at = COALESCE($3, now()),
			current_period_end = $4,
			updated_at = now()
		WHERE workspace_id = $1
	`, wsID, freePlanID, unixToTimePtr(sub.CanceledAt), unixToTimePtr(sub.CurrentPeriodEnd))
	if err != nil {
		return err
	}

	return s.entitlementSvc.DeriveFromSubscription(ctx, wsID, domain.TierFree)
}

func (s *BillingService) resolveWorkspaceIDForSubscription(ctx context.Context, subID string, metadata map[string]string) (uuid.UUID, error) {
	var wsID uuid.UUID
	err := s.db.QueryRow(ctx, `SELECT workspace_id FROM subscriptions WHERE stripe_subscription_id = $1`, subID).Scan(&wsID)
	if err == nil {
		return wsID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, err
	}
	wsIDRaw, ok := metadata["workspace_id"]
	if !ok {
		return uuid.Nil, fmt.Errorf("subscription %s not linked to workspace", subID)
	}
	wsID, err = uuid.Parse(wsIDRaw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid workspace_id metadata: %w", err)
	}
	return wsID, nil
}

func (s *BillingService) resolvePlanByPriceID(ctx context.Context, priceID string) (uuid.UUID, domain.PlanTier, error) {
	var (
		planID uuid.UUID
		tier   domain.PlanTier
	)
	err := s.db.QueryRow(ctx, `
		SELECT id, tier FROM plans
		WHERE stripe_price_monthly = $1 OR stripe_price_yearly = $1
		LIMIT 1
	`, priceID).Scan(&planID, &tier)
	return planID, tier, err
}

func (s *BillingService) resolvePlanIDByTier(ctx context.Context, tier domain.PlanTier) (uuid.UUID, error) {
	var planID uuid.UUID
	err := s.db.QueryRow(ctx, `SELECT id FROM plans WHERE tier = $1 LIMIT 1`, tier).Scan(&planID)
	return planID, err
}

func (s *BillingService) syncEntitlementsAndReferral(ctx context.Context, wsID uuid.UUID, tier domain.PlanTier, status domain.SubscriptionStatus) error {
	if status != domain.StatusActive && status != domain.StatusTrialing {
		return nil
	}
	if err := s.entitlementSvc.DeriveFromSubscription(ctx, wsID, tier); err != nil {
		return err
	}
	if status == domain.StatusActive && s.referralSvc != nil && (tier == domain.TierPro || tier == domain.TierPremium) {
		if err := s.referralSvc.ProcessConversion(ctx, wsID); err != nil {
			return err
		}
	}
	return nil
}

func parsePlanTier(raw string) (domain.PlanTier, error) {
	switch domain.PlanTier(strings.ToLower(strings.TrimSpace(raw))) {
	case domain.TierFree, domain.TierPro, domain.TierPremium:
		return domain.PlanTier(strings.ToLower(strings.TrimSpace(raw))), nil
	default:
		return "", fmt.Errorf("invalid plan tier %q", raw)
	}
}

func firstSubscriptionPriceID(sub *stripe.Subscription) (string, bool) {
	if sub == nil || sub.Items == nil || len(sub.Items.Data) == 0 || sub.Items.Data[0].Price == nil {
		return "", false
	}
	if sub.Items.Data[0].Price.ID == "" {
		return "", false
	}
	return sub.Items.Data[0].Price.ID, true
}

func normalizeStripeStatus(status stripe.SubscriptionStatus) domain.SubscriptionStatus {
	switch status {
	case stripe.SubscriptionStatusTrialing:
		return domain.StatusTrialing
	case stripe.SubscriptionStatusActive:
		return domain.StatusActive
	case stripe.SubscriptionStatusPastDue:
		return domain.StatusPastDue
	case stripe.SubscriptionStatusCanceled:
		return domain.StatusCanceled
	case stripe.SubscriptionStatusPaused:
		return domain.StatusPaused
	case stripe.SubscriptionStatusUnpaid:
		return domain.StatusUnpaid
	case stripe.SubscriptionStatusIncomplete, stripe.SubscriptionStatusIncompleteExpired:
		return domain.StatusPastDue
	default:
		return domain.StatusPastDue
	}
}

func extractCustomerID(customers ...*stripe.Customer) string {
	for _, customer := range customers {
		if customer != nil && customer.ID != "" {
			return customer.ID
		}
	}
	return ""
}

func stringOrNilPtr(v string) *string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return &v
}

func unixToTimePtr(ts int64) *time.Time {
	if ts <= 0 {
		return nil
	}
	t := time.Unix(ts, 0).UTC()
	return &t
}
