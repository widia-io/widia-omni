SET search_path TO widia_omni;

CREATE TYPE goal_status AS ENUM (
    'not_started', 'on_track', 'at_risk',
    'behind', 'completed', 'cancelled'
);
CREATE TYPE goal_period AS ENUM ('yearly', 'quarterly', 'monthly', 'weekly');

CREATE TABLE goals (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    area_id         UUID REFERENCES life_areas(id) ON DELETE SET NULL,
    parent_id       UUID REFERENCES goals(id) ON DELETE CASCADE,
    title           TEXT NOT NULL,
    description     TEXT,
    period          goal_period NOT NULL DEFAULT 'quarterly',
    status          goal_status NOT NULL DEFAULT 'not_started',
    target_value    NUMERIC,
    current_value   NUMERIC DEFAULT 0,
    unit            TEXT,
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);
