SET search_path TO widia_omni;

-- ============================================================
-- REFERRAL TYPES
-- ============================================================
CREATE TYPE referral_attribution_status AS ENUM ('pending', 'converted', 'expired');
CREATE TYPE referral_credit_status AS ENUM ('available', 'consumed', 'expired');

-- ============================================================
-- COUNTERS: MEMBERS
-- ============================================================
ALTER TABLE workspace_counters
    ADD COLUMN IF NOT EXISTS members_count INT NOT NULL DEFAULT 1;

UPDATE workspace_counters wc
SET members_count = sub.cnt
FROM (
    SELECT workspace_id, COUNT(*)::INT AS cnt
    FROM workspace_members
    GROUP BY workspace_id
) sub
WHERE wc.workspace_id = sub.workspace_id;

CREATE OR REPLACE FUNCTION increment_members_counter()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE workspace_counters
    SET members_count = members_count + 1, updated_at = now()
    WHERE workspace_id = NEW.workspace_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER SET search_path = widia_omni;

CREATE OR REPLACE FUNCTION decrement_members_counter()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE workspace_counters
    SET members_count = GREATEST(members_count - 1, 0), updated_at = now()
    WHERE workspace_id = OLD.workspace_id;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER SET search_path = widia_omni;

DROP TRIGGER IF EXISTS trg_ws_members_inc ON workspace_members;
CREATE TRIGGER trg_ws_members_inc
AFTER INSERT ON workspace_members
FOR EACH ROW
EXECUTE FUNCTION increment_members_counter();

DROP TRIGGER IF EXISTS trg_ws_members_dec ON workspace_members;
CREATE TRIGGER trg_ws_members_dec
AFTER DELETE ON workspace_members
FOR EACH ROW
EXECUTE FUNCTION decrement_members_counter();

-- ============================================================
-- FAMILY INVITES
-- ============================================================
CREATE TABLE workspace_invites (
    id            UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id  UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    email         TEXT NOT NULL,
    role          workspace_role NOT NULL DEFAULT 'member',
    token_hash    TEXT NOT NULL UNIQUE,
    invited_by    UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    expires_at    TIMESTAMPTZ NOT NULL,
    accepted_at   TIMESTAMPTZ,
    accepted_by   UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    revoked_at    TIMESTAMPTZ,
    revoked_by    UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT workspace_invites_status_check CHECK (
        NOT (accepted_at IS NOT NULL AND revoked_at IS NOT NULL)
    )
);

CREATE INDEX idx_workspace_invites_ws ON workspace_invites(workspace_id, created_at DESC);
CREATE INDEX idx_workspace_invites_email ON workspace_invites(LOWER(email));
CREATE INDEX idx_workspace_invites_pending ON workspace_invites(workspace_id, created_at DESC)
WHERE accepted_at IS NULL AND revoked_at IS NULL;

ALTER TABLE workspace_invites ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_workspace_invites ON workspace_invites
    FOR ALL USING (is_workspace_member(workspace_id))
    WITH CHECK (is_workspace_member(workspace_id));

-- ============================================================
-- REFERRAL STRUCTURES
-- ============================================================
CREATE TABLE workspace_referral_codes (
    workspace_id   UUID PRIMARY KEY REFERENCES workspaces(id) ON DELETE CASCADE,
    code           TEXT NOT NULL UNIQUE,
    created_by     UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    regenerated_at TIMESTAMPTZ
);

ALTER TABLE workspace_referral_codes ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_workspace_referral_codes ON workspace_referral_codes
    FOR ALL USING (is_workspace_member(workspace_id))
    WITH CHECK (is_workspace_member(workspace_id));

CREATE TABLE referral_attributions (
    id                    UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    referral_code         TEXT NOT NULL,
    referrer_workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    referred_workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    referred_user_id      UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    expires_at            TIMESTAMPTZ NOT NULL,
    status                referral_attribution_status NOT NULL DEFAULT 'pending',
    converted_at          TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (referred_workspace_id),
    UNIQUE (referrer_workspace_id, referred_workspace_id),
    CONSTRAINT referral_attributions_status_check CHECK (
        (status = 'converted' AND converted_at IS NOT NULL) OR
        (status <> 'converted')
    )
);

CREATE INDEX idx_referral_attr_referrer ON referral_attributions(referrer_workspace_id, status, created_at DESC);
CREATE INDEX idx_referral_attr_referred ON referral_attributions(referred_workspace_id);
CREATE INDEX idx_referral_attr_pending_exp ON referral_attributions(expires_at)
WHERE status = 'pending';

ALTER TABLE referral_attributions ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_referral_attributions ON referral_attributions
    FOR ALL USING (
        is_workspace_member(referrer_workspace_id)
        OR is_workspace_member(referred_workspace_id)
    )
    WITH CHECK (
        is_workspace_member(referrer_workspace_id)
        OR is_workspace_member(referred_workspace_id)
    );

CREATE TABLE referral_credits (
    id             UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    attribution_id UUID NOT NULL REFERENCES referral_attributions(id) ON DELETE CASCADE,
    workspace_id   UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    credit_type    TEXT NOT NULL CHECK (credit_type IN ('referrer_bonus', 'referred_bonus')),
    days           INT NOT NULL DEFAULT 30 CHECK (days > 0),
    status         referral_credit_status NOT NULL DEFAULT 'available',
    expires_at     TIMESTAMPTZ,
    consumed_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (attribution_id, workspace_id, credit_type)
);

CREATE INDEX idx_referral_credits_ws ON referral_credits(workspace_id, status, created_at DESC);

ALTER TABLE referral_credits ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_referral_credits ON referral_credits
    FOR ALL USING (is_workspace_member(workspace_id))
    WITH CHECK (is_workspace_member(workspace_id));

-- ============================================================
-- REFERRAL CODE GENERATION + BACKFILL
-- ============================================================
CREATE OR REPLACE FUNCTION generate_referral_code()
RETURNS TEXT AS $$
DECLARE
    v_code TEXT;
BEGIN
    LOOP
        v_code := UPPER(SUBSTRING(REPLACE(gen_random_uuid()::text, '-', '') FROM 1 FOR 8));
        EXIT WHEN NOT EXISTS (
            SELECT 1 FROM workspace_referral_codes WHERE code = v_code
        );
    END LOOP;
    RETURN v_code;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER SET search_path = widia_omni;

INSERT INTO workspace_referral_codes (workspace_id, code, created_by)
SELECT w.id, generate_referral_code(), w.owner_id
FROM workspaces w
ON CONFLICT (workspace_id) DO NOTHING;

-- ============================================================
-- LIMITS + FLAGS
-- ============================================================
UPDATE plans
SET limits = jsonb_set(
    jsonb_set(
        jsonb_set(
            jsonb_set(
                limits,
                '{max_members}',
                to_jsonb(CASE WHEN tier = 'premium' THEN 5 ELSE 1 END),
                true
            ),
            '{family_enabled}',
            to_jsonb(false),
            true
        ),
        '{referral_enabled}',
        to_jsonb(false),
        true
    ),
    '{mobile_pwa_enabled}',
    to_jsonb(false),
    true
);

UPDATE workspace_entitlements
SET limits = jsonb_set(
    jsonb_set(
        jsonb_set(
            jsonb_set(
                limits,
                '{max_members}',
                to_jsonb(CASE WHEN tier = 'premium' THEN 5 ELSE 1 END),
                true
            ),
            '{family_enabled}',
            to_jsonb(false),
            true
        ),
        '{referral_enabled}',
        to_jsonb(false),
        true
    ),
    '{mobile_pwa_enabled}',
    to_jsonb(false),
    true
);

-- ============================================================
-- SIGNUP AUTO-SETUP: REFERRAL SUPPORT
-- ============================================================
CREATE OR REPLACE FUNCTION handle_new_user()
RETURNS TRIGGER AS $$
DECLARE
    v_ws_id UUID;
    v_plan_id UUID;
    v_limits JSONB;
    v_name TEXT;
    v_referral_code TEXT;
    v_referrer_ws_id UUID;
BEGIN
    v_name := COALESCE(NEW.raw_user_meta_data->>'display_name', split_part(NEW.email, '@', 1));
    v_referral_code := NULLIF(TRIM(COALESCE(NEW.raw_user_meta_data->>'referral_code', '')), '');

    INSERT INTO workspaces (name, slug, owner_id)
    VALUES (v_name || '''s Space', 'ws-' || substr(gen_random_uuid()::text, 1, 12), NEW.id)
    RETURNING id INTO v_ws_id;

    INSERT INTO workspace_members (workspace_id, user_id, role)
    VALUES (v_ws_id, NEW.id, 'owner');

    INSERT INTO user_profiles (id, display_name, email, default_workspace_id)
    VALUES (NEW.id, v_name, NEW.email, v_ws_id);

    INSERT INTO user_preferences (user_id)
    VALUES (NEW.id);

    SELECT id, limits
    INTO v_plan_id, v_limits
    FROM plans
    WHERE tier = 'free'
    LIMIT 1;

    INSERT INTO subscriptions (workspace_id, plan_id, tier, status)
    VALUES (v_ws_id, v_plan_id, 'free', 'active');

    INSERT INTO workspace_entitlements (workspace_id, tier, limits, source, is_current)
    VALUES (v_ws_id, 'free', v_limits, 'free', true);

    INSERT INTO workspace_counters (workspace_id)
    VALUES (v_ws_id);

    INSERT INTO workspace_referral_codes (workspace_id, code, created_by)
    VALUES (v_ws_id, generate_referral_code(), NEW.id)
    ON CONFLICT (workspace_id) DO NOTHING;

    IF v_referral_code IS NOT NULL THEN
        SELECT workspace_id INTO v_referrer_ws_id
        FROM workspace_referral_codes
        WHERE UPPER(code) = UPPER(v_referral_code)
        LIMIT 1;

        IF v_referrer_ws_id IS NOT NULL AND v_referrer_ws_id <> v_ws_id THEN
            INSERT INTO referral_attributions (
                referral_code,
                referrer_workspace_id,
                referred_workspace_id,
                referred_user_id,
                expires_at,
                status
            )
            VALUES (
                UPPER(v_referral_code),
                v_referrer_ws_id,
                v_ws_id,
                NEW.id,
                now() + INTERVAL '30 days',
                'pending'
            )
            ON CONFLICT (referred_workspace_id) DO NOTHING;
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER SET search_path = widia_omni;
