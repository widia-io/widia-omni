SET search_path TO widia_omni;

CREATE TYPE task_priority AS ENUM ('low', 'medium', 'high', 'critical');

CREATE TABLE tasks (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    area_id         UUID REFERENCES life_areas(id) ON DELETE SET NULL,
    goal_id         UUID REFERENCES goals(id) ON DELETE SET NULL,
    title           TEXT NOT NULL,
    description     TEXT,
    priority        task_priority NOT NULL DEFAULT 'medium',
    is_completed    BOOLEAN NOT NULL DEFAULT false,
    is_focus        BOOLEAN NOT NULL DEFAULT false,
    due_date        DATE,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);
