SET search_path TO widia_omni;

CREATE TYPE transaction_type AS ENUM ('income', 'expense', 'investment', 'transfer');

CREATE TABLE finance_categories (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    type            transaction_type NOT NULL,
    color           TEXT,
    icon            TEXT,
    parent_id       UUID REFERENCES finance_categories(id) ON DELETE SET NULL,
    is_system       BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE TABLE transactions (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    category_id     UUID REFERENCES finance_categories(id) ON DELETE SET NULL,
    area_id         UUID REFERENCES life_areas(id) ON DELETE SET NULL,
    type            transaction_type NOT NULL,
    amount          NUMERIC(12,2) NOT NULL,
    description     TEXT,
    date            DATE NOT NULL,
    is_recurring    BOOLEAN NOT NULL DEFAULT false,
    recurrence_rule TEXT,
    tags            TEXT[],
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);
