SET search_path TO widia_omni;

CREATE TYPE plan_tier AS ENUM ('free', 'pro', 'premium');
CREATE TYPE subscription_status AS ENUM (
    'trialing', 'active', 'past_due',
    'canceled', 'paused', 'unpaid'
);

CREATE TABLE plans (
    id                    UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    tier                  plan_tier NOT NULL UNIQUE,
    name                  TEXT NOT NULL,
    price_monthly         NUMERIC(8,2) NOT NULL DEFAULT 0,
    price_yearly          NUMERIC(8,2) NOT NULL DEFAULT 0,
    stripe_price_monthly  TEXT,
    stripe_price_yearly   TEXT,
    limits                JSONB NOT NULL DEFAULT '{}',
    is_active             BOOLEAN NOT NULL DEFAULT true,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE subscriptions (
    id                      UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id            UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    plan_id                 UUID NOT NULL REFERENCES plans(id),
    tier                    plan_tier NOT NULL,
    status                  subscription_status NOT NULL DEFAULT 'active',
    stripe_customer_id      TEXT,
    stripe_subscription_id  TEXT,
    stripe_price_id         TEXT,
    currency                TEXT NOT NULL DEFAULT 'BRL',
    current_period_start    TIMESTAMPTZ,
    current_period_end      TIMESTAMPTZ,
    trial_end               TIMESTAMPTZ,
    cancel_at               TIMESTAMPTZ,
    canceled_at             TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_subs_active_ws
    ON subscriptions(workspace_id) WHERE status IN ('trialing', 'active', 'past_due');
