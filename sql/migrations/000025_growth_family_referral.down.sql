SET search_path TO widia_omni;

-- Revert signup auto-setup function to pre-M9 version.
CREATE OR REPLACE FUNCTION handle_new_user()
RETURNS TRIGGER AS $$
DECLARE
    v_ws_id UUID;
    v_plan_id UUID;
    v_limits JSONB;
    v_name TEXT;
BEGIN
    v_name := COALESCE(NEW.raw_user_meta_data->>'display_name', split_part(NEW.email, '@', 1));

    INSERT INTO workspaces (name, slug, owner_id)
    VALUES (v_name || '''s Space', 'ws-' || substr(gen_random_uuid()::text, 1, 12), NEW.id)
    RETURNING id INTO v_ws_id;

    INSERT INTO workspace_members (workspace_id, user_id, role) VALUES (v_ws_id, NEW.id, 'owner');
    INSERT INTO user_profiles (id, display_name, email, default_workspace_id) VALUES (NEW.id, v_name, NEW.email, v_ws_id);
    INSERT INTO user_preferences (user_id) VALUES (NEW.id);

    SELECT id, limits INTO v_plan_id, v_limits FROM plans WHERE tier = 'free' LIMIT 1;
    INSERT INTO subscriptions (workspace_id, plan_id, tier, status) VALUES (v_ws_id, v_plan_id, 'free', 'active');
    INSERT INTO workspace_entitlements (workspace_id, tier, limits, source, is_current) VALUES (v_ws_id, 'free', v_limits, 'free', true);
    INSERT INTO workspace_counters (workspace_id) VALUES (v_ws_id);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER SET search_path = widia_omni;

-- Remove member counter triggers/functions.
DROP TRIGGER IF EXISTS trg_ws_members_inc ON workspace_members;
DROP TRIGGER IF EXISTS trg_ws_members_dec ON workspace_members;
DROP FUNCTION IF EXISTS increment_members_counter();
DROP FUNCTION IF EXISTS decrement_members_counter();

-- Drop referral/family tables and helper function.
DROP TABLE IF EXISTS referral_credits;
DROP TABLE IF EXISTS referral_attributions;
DROP TABLE IF EXISTS workspace_referral_codes;
DROP TABLE IF EXISTS workspace_invites;
DROP FUNCTION IF EXISTS generate_referral_code();

DROP TYPE IF EXISTS referral_credit_status;
DROP TYPE IF EXISTS referral_attribution_status;

ALTER TABLE workspace_counters
    DROP COLUMN IF EXISTS members_count;

-- Remove limits/flags keys added in M9.
UPDATE plans
SET limits = limits
    - 'max_members'
    - 'family_enabled'
    - 'referral_enabled'
    - 'mobile_pwa_enabled';

UPDATE workspace_entitlements
SET limits = limits
    - 'max_members'
    - 'family_enabled'
    - 'referral_enabled'
    - 'mobile_pwa_enabled';
