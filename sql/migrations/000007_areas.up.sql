SET search_path TO widia_omni;

CREATE TABLE life_areas (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    slug            TEXT NOT NULL,
    icon            TEXT NOT NULL DEFAULT '🎯',
    color           TEXT NOT NULL DEFAULT '#d97757',
    weight          NUMERIC(3,2) NOT NULL DEFAULT 1.0,
    sort_order      INT NOT NULL DEFAULT 0,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,
    UNIQUE(workspace_id, slug)
);
