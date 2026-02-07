SET search_path TO widia_omni;

CREATE TABLE workspace_counters (
    workspace_id                UUID PRIMARY KEY REFERENCES workspaces(id) ON DELETE CASCADE,
    areas_count                 INT NOT NULL DEFAULT 0,
    goals_count                 INT NOT NULL DEFAULT 0,
    habits_count                INT NOT NULL DEFAULT 0,
    tasks_created_today         INT NOT NULL DEFAULT 0,
    tasks_today_date            DATE NOT NULL DEFAULT CURRENT_DATE,
    transactions_month_count    INT NOT NULL DEFAULT 0,
    transactions_month          TEXT NOT NULL DEFAULT to_char(CURRENT_DATE, 'YYYY-MM'),
    storage_bytes_used          BIGINT NOT NULL DEFAULT 0,
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now()
);
