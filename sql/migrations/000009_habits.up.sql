SET search_path TO widia_omni;

CREATE TYPE habit_frequency AS ENUM ('daily', 'weekly', 'custom');

CREATE TABLE habits (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    area_id         UUID REFERENCES life_areas(id) ON DELETE SET NULL,
    name            TEXT NOT NULL,
    color           TEXT NOT NULL DEFAULT '#788c5d',
    frequency       habit_frequency NOT NULL DEFAULT 'daily',
    target_per_week INT NOT NULL DEFAULT 3,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE TABLE habit_entries (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    habit_id        UUID NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    date            DATE NOT NULL,
    intensity       SMALLINT NOT NULL DEFAULT 3 CHECK (intensity BETWEEN 1 AND 3),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(habit_id, date)
);
