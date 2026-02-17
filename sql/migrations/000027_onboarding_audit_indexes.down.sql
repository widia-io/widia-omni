SET search_path TO widia_omni;

DROP INDEX IF EXISTS idx_audit_log_workspace_action_created_at;
DROP INDEX IF EXISTS idx_audit_log_user_action_created_at;
DROP INDEX IF EXISTS idx_audit_log_action_created_at;
