SET search_path TO widia_omni;

DROP POLICY IF EXISTS rls_plans ON plans;
ALTER TABLE plans DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS rls_audit ON audit_log;
ALTER TABLE audit_log DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS rls_notif ON notifications;
ALTER TABLE notifications DISABLE ROW LEVEL SECURITY;

DO $$
DECLARE tbl TEXT;
BEGIN
    FOR tbl IN SELECT unnest(ARRAY[
        'subscriptions', 'workspace_entitlements', 'workspace_counters',
        'life_areas', 'goals', 'habits', 'habit_entries',
        'tasks', 'finance_categories', 'transactions',
        'journal_entries', 'area_scores', 'life_scores'
    ])
    LOOP
        EXECUTE format('DROP POLICY IF EXISTS rls_%s ON widia_omni.%I', tbl, tbl);
        EXECUTE format('ALTER TABLE widia_omni.%I DISABLE ROW LEVEL SECURITY', tbl);
    END LOOP;
END;
$$;

DROP POLICY IF EXISTS rls_prefs ON user_preferences;
ALTER TABLE user_preferences DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS rls_profiles ON user_profiles;
ALTER TABLE user_profiles DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS rls_ws_members ON workspace_members;
ALTER TABLE workspace_members DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS rls_workspaces ON workspaces;
ALTER TABLE workspaces DISABLE ROW LEVEL SECURITY;

DROP FUNCTION IF EXISTS widia_omni.is_workspace_member;
