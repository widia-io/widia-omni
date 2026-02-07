SET search_path TO widia_omni;

CREATE OR REPLACE FUNCTION widia_omni.is_workspace_member(ws_id UUID)
RETURNS BOOLEAN AS $$
    SELECT EXISTS (
        SELECT 1 FROM widia_omni.workspace_members
        WHERE workspace_id = ws_id AND user_id = auth.uid()
    );
$$ LANGUAGE sql SECURITY DEFINER STABLE SET search_path = widia_omni;

-- Workspaces
ALTER TABLE workspaces ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_workspaces ON workspaces
    FOR ALL USING (widia_omni.is_workspace_member(id)) WITH CHECK (owner_id = auth.uid());

-- Workspace members
ALTER TABLE workspace_members ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_ws_members ON workspace_members
    FOR ALL USING (widia_omni.is_workspace_member(workspace_id));

-- User profiles (own only)
ALTER TABLE user_profiles ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_profiles ON user_profiles
    FOR ALL USING (auth.uid() = id) WITH CHECK (auth.uid() = id);

-- User preferences (own only)
ALTER TABLE user_preferences ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_prefs ON user_preferences
    FOR ALL USING (auth.uid() = user_id) WITH CHECK (auth.uid() = user_id);

-- All workspace-scoped business tables
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
        EXECUTE format('ALTER TABLE widia_omni.%I ENABLE ROW LEVEL SECURITY', tbl);
        EXECUTE format(
            'CREATE POLICY rls_%s ON widia_omni.%I FOR ALL USING (widia_omni.is_workspace_member(workspace_id)) WITH CHECK (widia_omni.is_workspace_member(workspace_id))',
            tbl, tbl
        );
    END LOOP;
END;
$$;

-- Notifications (own only + workspace member)
ALTER TABLE notifications ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_notif ON notifications
    FOR ALL USING (auth.uid() = user_id AND widia_omni.is_workspace_member(workspace_id));

-- Audit log (select by workspace, insert via function only)
ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_audit ON audit_log
    FOR SELECT USING (widia_omni.is_workspace_member(workspace_id));

-- Plans (public read)
ALTER TABLE plans ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_plans ON plans FOR SELECT USING (true);
