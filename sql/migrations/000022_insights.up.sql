SET search_path TO widia_omni;

CREATE TABLE workspace_insights (
    id                UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id      UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    type              TEXT NOT NULL CHECK (type IN ('weekly_summary', 'on_demand')),
    week_start        DATE NOT NULL,
    content           JSONB NOT NULL,
    model             TEXT NOT NULL,
    prompt_tokens     INT NOT NULL DEFAULT 0,
    completion_tokens INT NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_insights_ws_created ON workspace_insights(workspace_id, created_at DESC);
CREATE INDEX idx_insights_ws_type_week ON workspace_insights(workspace_id, type, week_start);
ALTER TABLE workspace_insights ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_workspace_insights ON workspace_insights
    FOR ALL USING (is_workspace_member(workspace_id))
    WITH CHECK (is_workspace_member(workspace_id));
