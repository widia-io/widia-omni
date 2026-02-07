SET search_path TO widia_omni;

DROP TRIGGER IF EXISTS on_auth_user_created ON auth.users;
DROP FUNCTION IF EXISTS widia_omni.handle_new_user;

DO $$
DECLARE tbl TEXT;
BEGIN
    FOR tbl IN SELECT unnest(ARRAY[
        'workspaces', 'user_profiles', 'user_preferences',
        'subscriptions', 'life_areas', 'goals', 'habits',
        'tasks', 'journal_entries', 'workspace_counters'
    ])
    LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS trg_%s_updated_at ON widia_omni.%I', tbl, tbl);
    END LOOP;
END;
$$;

DROP FUNCTION IF EXISTS widia_omni.update_updated_at;

DROP TRIGGER IF EXISTS trg_habits_dec ON habits;
DROP TRIGGER IF EXISTS trg_habits_inc ON habits;
DROP TRIGGER IF EXISTS trg_goals_dec ON goals;
DROP TRIGGER IF EXISTS trg_goals_inc ON goals;
DROP TRIGGER IF EXISTS trg_areas_dec ON life_areas;
DROP TRIGGER IF EXISTS trg_areas_inc ON life_areas;
DROP FUNCTION IF EXISTS widia_omni.decrement_counter;
DROP FUNCTION IF EXISTS widia_omni.increment_counter;
