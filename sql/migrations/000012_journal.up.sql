SET search_path TO widia_omni;

CREATE TABLE journal_entries (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    date            DATE NOT NULL,
    mood            SMALLINT CHECK (mood BETWEEN 1 AND 5),
    energy          SMALLINT CHECK (energy BETWEEN 1 AND 5),
    wins            TEXT[],
    challenges      TEXT[],
    gratitude       TEXT[],
    notes           TEXT,
    tags            TEXT[],
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(workspace_id, date)
);
