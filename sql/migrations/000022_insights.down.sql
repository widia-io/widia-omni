SET search_path TO widia_omni;

DROP POLICY IF EXISTS rls_workspace_insights ON workspace_insights;
DROP TABLE IF EXISTS workspace_insights;
