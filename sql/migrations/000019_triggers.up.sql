SET search_path TO widia_omni;

-- Counter triggers
CREATE OR REPLACE FUNCTION widia_omni.increment_counter()
RETURNS TRIGGER AS $$
BEGIN
    CASE TG_TABLE_NAME
        WHEN 'life_areas' THEN
            UPDATE widia_omni.workspace_counters SET areas_count = areas_count + 1, updated_at = now() WHERE workspace_id = NEW.workspace_id;
        WHEN 'goals' THEN
            UPDATE widia_omni.workspace_counters SET goals_count = goals_count + 1, updated_at = now() WHERE workspace_id = NEW.workspace_id;
        WHEN 'habits' THEN
            UPDATE widia_omni.workspace_counters SET habits_count = habits_count + 1, updated_at = now() WHERE workspace_id = NEW.workspace_id;
    END CASE;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER SET search_path = widia_omni;

CREATE OR REPLACE FUNCTION widia_omni.decrement_counter()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        CASE TG_TABLE_NAME
            WHEN 'life_areas' THEN
                UPDATE widia_omni.workspace_counters SET areas_count = GREATEST(areas_count - 1, 0), updated_at = now() WHERE workspace_id = NEW.workspace_id;
            WHEN 'goals' THEN
                UPDATE widia_omni.workspace_counters SET goals_count = GREATEST(goals_count - 1, 0), updated_at = now() WHERE workspace_id = NEW.workspace_id;
            WHEN 'habits' THEN
                UPDATE widia_omni.workspace_counters SET habits_count = GREATEST(habits_count - 1, 0), updated_at = now() WHERE workspace_id = NEW.workspace_id;
        END CASE;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER SET search_path = widia_omni;

CREATE TRIGGER trg_areas_inc AFTER INSERT ON life_areas FOR EACH ROW WHEN (NEW.deleted_at IS NULL) EXECUTE FUNCTION widia_omni.increment_counter();
CREATE TRIGGER trg_areas_dec AFTER UPDATE ON life_areas FOR EACH ROW EXECUTE FUNCTION widia_omni.decrement_counter();
CREATE TRIGGER trg_goals_inc AFTER INSERT ON goals FOR EACH ROW WHEN (NEW.deleted_at IS NULL) EXECUTE FUNCTION widia_omni.increment_counter();
CREATE TRIGGER trg_goals_dec AFTER UPDATE ON goals FOR EACH ROW EXECUTE FUNCTION widia_omni.decrement_counter();
CREATE TRIGGER trg_habits_inc AFTER INSERT ON habits FOR EACH ROW WHEN (NEW.deleted_at IS NULL) EXECUTE FUNCTION widia_omni.increment_counter();
CREATE TRIGGER trg_habits_dec AFTER UPDATE ON habits FOR EACH ROW EXECUTE FUNCTION widia_omni.decrement_counter();

-- updated_at triggers
CREATE OR REPLACE FUNCTION widia_omni.update_updated_at()
RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = now(); RETURN NEW; END;
$$ LANGUAGE plpgsql SET search_path = widia_omni;

DO $$
DECLARE tbl TEXT;
BEGIN
    FOR tbl IN SELECT unnest(ARRAY[
        'workspaces', 'user_profiles', 'user_preferences',
        'subscriptions', 'life_areas', 'goals', 'habits',
        'tasks', 'journal_entries', 'workspace_counters'
    ])
    LOOP
        EXECUTE format(
            'CREATE TRIGGER trg_%s_updated_at BEFORE UPDATE ON widia_omni.%I FOR EACH ROW EXECUTE FUNCTION widia_omni.update_updated_at()',
            tbl, tbl
        );
    END LOOP;
END;
$$;

-- Auto-setup on signup
CREATE OR REPLACE FUNCTION widia_omni.handle_new_user()
RETURNS TRIGGER AS $$
DECLARE
    v_ws_id UUID;
    v_plan_id UUID;
    v_limits JSONB;
    v_name TEXT;
BEGIN
    v_name := COALESCE(NEW.raw_user_meta_data->>'display_name', split_part(NEW.email, '@', 1));

    INSERT INTO widia_omni.workspaces (name, slug, owner_id)
    VALUES (v_name || '''s Space', 'ws-' || substr(gen_random_uuid()::text, 1, 12), NEW.id)
    RETURNING id INTO v_ws_id;

    INSERT INTO widia_omni.workspace_members (workspace_id, user_id, role) VALUES (v_ws_id, NEW.id, 'owner');
    INSERT INTO widia_omni.user_profiles (id, display_name, email, default_workspace_id) VALUES (NEW.id, v_name, NEW.email, v_ws_id);
    INSERT INTO widia_omni.user_preferences (user_id) VALUES (NEW.id);

    SELECT id, limits INTO v_plan_id, v_limits FROM widia_omni.plans WHERE tier = 'free' LIMIT 1;
    INSERT INTO widia_omni.subscriptions (workspace_id, plan_id, tier, status) VALUES (v_ws_id, v_plan_id, 'free', 'active');
    INSERT INTO widia_omni.workspace_entitlements (workspace_id, tier, limits, source, is_current) VALUES (v_ws_id, 'free', v_limits, 'free', true);
    INSERT INTO widia_omni.workspace_counters (workspace_id) VALUES (v_ws_id);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER SET search_path = widia_omni;

CREATE TRIGGER on_auth_user_created
    AFTER INSERT ON auth.users
    FOR EACH ROW EXECUTE FUNCTION widia_omni.handle_new_user();
