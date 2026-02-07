SET search_path TO widia_omni;

CREATE TYPE entitlement_source AS ENUM ('free', 'stripe', 'trial', 'promo', 'admin');

CREATE TABLE workspace_entitlements (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    tier            plan_tier NOT NULL DEFAULT 'free',
    limits          JSONB NOT NULL,
    source          entitlement_source NOT NULL DEFAULT 'free',
    effective_from  TIMESTAMPTZ NOT NULL DEFAULT now(),
    effective_to    TIMESTAMPTZ,
    is_current      BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_ent_current_ws
    ON workspace_entitlements(workspace_id) WHERE is_current = true;
