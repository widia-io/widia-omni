SET search_path TO widia_omni;

CREATE TABLE api_keys (
    id           UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    created_by   UUID NOT NULL,
    name         TEXT NOT NULL DEFAULT '',
    key_hash     TEXT NOT NULL,
    key_prefix   TEXT NOT NULL,
    last_used_at TIMESTAMPTZ,
    revoked_at   TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_api_keys_hash ON api_keys(key_hash) WHERE revoked_at IS NULL;
CREATE INDEX idx_api_keys_workspace ON api_keys(workspace_id) WHERE revoked_at IS NULL;
ALTER TABLE api_keys ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_api_keys ON api_keys
    FOR ALL USING (is_workspace_member(workspace_id))
    WITH CHECK (is_workspace_member(workspace_id));
