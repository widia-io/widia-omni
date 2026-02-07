SET search_path TO widia_omni;

CREATE TABLE audit_log (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id    UUID REFERENCES workspaces(id) ON DELETE SET NULL,
    user_id         UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    action          TEXT NOT NULL,
    entity_type     TEXT,
    entity_id       UUID,
    metadata        JSONB,
    ip_address      INET,
    user_agent      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION widia_omni.insert_audit_log(
    p_workspace_id UUID, p_user_id UUID, p_action TEXT,
    p_entity_type TEXT DEFAULT NULL, p_entity_id UUID DEFAULT NULL,
    p_metadata JSONB DEFAULT NULL, p_ip INET DEFAULT NULL, p_ua TEXT DEFAULT NULL
) RETURNS UUID AS $$
DECLARE v_id UUID;
BEGIN
    INSERT INTO widia_omni.audit_log (workspace_id, user_id, action, entity_type, entity_id, metadata, ip_address, user_agent)
    VALUES (p_workspace_id, p_user_id, p_action, p_entity_type, p_entity_id, p_metadata, p_ip, p_ua)
    RETURNING id INTO v_id;
    RETURN v_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER SET search_path = widia_omni;
