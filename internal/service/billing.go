package service

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/webhook"
	"github.com/widia-io/widia-omni/internal/domain"
)

type BillingService struct {
	db             *pgxpool.Pool
	entitlementSvc *EntitlementService
	webhookSecret  string
	successURL     string
	cancelURL      string
}

func NewBillingService(db *pgxpool.Pool, entSvc *EntitlementService, stripeKey, webhookSecret, successURL, cancelURL string) *BillingService {
	stripe.Key = stripeKey
	return &BillingService{
		db:             db,
		entitlementSvc: entSvc,
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
		return "", errors.New("plan not found or no stripe price configured")
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

	event, err := webhook.ConstructEvent(payload, signature, s.webhookSecret)
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
	sess, err := checkoutsession.Get(event.Data.Object["id"].(string), nil)
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
		return nil
	}

	_, err = s.db.Exec(ctx, `
		UPDATE subscriptions
		SET stripe_customer_id = $2, stripe_subscription_id = $3, status = 'active', updated_at = now()
		WHERE workspace_id = $1
	`, wsID, sess.Customer.ID, sess.Subscription.ID)
	return err
}

func (s *BillingService) handleSubscriptionUpdated(ctx context.Context, event stripe.Event) error {
	subID, _ := event.Data.Object["id"].(string)
	status, _ := event.Data.Object["status"].(string)

	var wsID uuid.UUID
	err := s.db.QueryRow(ctx, `SELECT workspace_id FROM subscriptions WHERE stripe_subscription_id = $1`, subID).Scan(&wsID)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, `
		UPDATE subscriptions SET status = $2, updated_at = now()
		WHERE stripe_subscription_id = $1
	`, subID, status)
	if err != nil {
		return err
	}

	if status == "active" {
		var tier domain.PlanTier
		_ = s.db.QueryRow(ctx, `SELECT tier FROM subscriptions WHERE stripe_subscription_id = $1`, subID).Scan(&tier)
		return s.entitlementSvc.DeriveFromSubscription(ctx, wsID, tier)
	}
	return nil
}

func (s *BillingService) handleSubscriptionDeleted(ctx context.Context, event stripe.Event) error {
	subID, _ := event.Data.Object["id"].(string)

	var wsID uuid.UUID
	err := s.db.QueryRow(ctx, `SELECT workspace_id FROM subscriptions WHERE stripe_subscription_id = $1`, subID).Scan(&wsID)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, `
		UPDATE subscriptions SET status = 'canceled', canceled_at = now(), updated_at = now()
		WHERE stripe_subscription_id = $1
	`, subID)
	if err != nil {
		return err
	}

	return s.entitlementSvc.DeriveFromSubscription(ctx, wsID, domain.TierFree)
}
