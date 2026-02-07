SET search_path TO widia_omni;

INSERT INTO plans (tier, name, price_monthly, price_yearly, limits) VALUES
('free', 'Free', 0, 0, '{
    "max_areas": 4, "max_goals": 5, "max_habits": 5,
    "max_tasks_per_day": 5, "max_transactions_per_month": 0,
    "journal_enabled": true, "finance_enabled": false,
    "export_enabled": false, "score_history_weeks": 4,
    "api_rate_limit_per_minute": 30, "storage_mb": 50,
    "ai_insights": false, "api_access": false
}'::jsonb),
('pro', 'Pro', 19.90, 199.00, '{
    "max_areas": 8, "max_goals": 25, "max_habits": 15,
    "max_tasks_per_day": 20, "max_transactions_per_month": 500,
    "journal_enabled": true, "finance_enabled": true,
    "export_enabled": true, "score_history_weeks": 52,
    "api_rate_limit_per_minute": 120, "storage_mb": 500,
    "ai_insights": false, "api_access": false
}'::jsonb),
('premium', 'Premium', 39.90, 399.00, '{
    "max_areas": -1, "max_goals": -1, "max_habits": -1,
    "max_tasks_per_day": -1, "max_transactions_per_month": -1,
    "journal_enabled": true, "finance_enabled": true,
    "export_enabled": true, "score_history_weeks": -1,
    "api_rate_limit_per_minute": 300, "storage_mb": 5000,
    "ai_insights": true, "api_access": true
}'::jsonb);
