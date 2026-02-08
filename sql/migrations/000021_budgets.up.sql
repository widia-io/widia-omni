SET search_path TO widia_omni;

CREATE TABLE budgets (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    category_id     UUID REFERENCES finance_categories(id) ON DELETE CASCADE,
    month           TEXT NOT NULL,
    amount          NUMERIC(12,2) NOT NULL CHECK (amount >= 0),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX budgets_ws_cat_month
    ON budgets(workspace_id, COALESCE(category_id, '00000000-0000-0000-0000-000000000000'::uuid), month);

CREATE INDEX idx_budget_ws_month ON budgets(workspace_id, month);

ALTER TABLE budgets ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_budgets ON budgets
    FOR ALL USING (widia_omni.is_workspace_member(workspace_id))
    WITH CHECK (widia_omni.is_workspace_member(workspace_id));
