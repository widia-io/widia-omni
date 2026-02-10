SET search_path TO widia_omni;

DROP POLICY IF EXISTS rls_api_keys ON api_keys;
DROP TABLE IF EXISTS api_keys;
