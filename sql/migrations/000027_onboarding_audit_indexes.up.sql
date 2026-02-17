SET search_path TO widia_omni;

CREATE INDEX idx_audit_log_action_created_at ON audit_log(action, created_at DESC);
CREATE INDEX idx_audit_log_user_action_created_at ON audit_log(user_id, action, created_at DESC);
CREATE INDEX idx_audit_log_workspace_action_created_at ON audit_log(workspace_id, action, created_at DESC);
