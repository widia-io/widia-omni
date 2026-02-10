SET search_path TO widia_omni;

-- Labels: standalone colored tags per workspace
CREATE TABLE labels (
    id           UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    color        TEXT NOT NULL DEFAULT 'gray',
    position     INT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ,
    UNIQUE(workspace_id, name)
);

-- Task-Label junction (many-to-many)
CREATE TABLE task_labels (
    task_id  UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    label_id UUID NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, label_id)
);

-- Sections: group tasks within an area
CREATE TABLE sections (
    id           UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    area_id      UUID NOT NULL REFERENCES life_areas(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    position     INT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ
);

-- Extend tasks table
ALTER TABLE tasks ADD COLUMN parent_id UUID REFERENCES tasks(id) ON DELETE CASCADE;
ALTER TABLE tasks ADD COLUMN section_id UUID REFERENCES sections(id) ON DELETE SET NULL;
ALTER TABLE tasks ADD COLUMN position INT NOT NULL DEFAULT 0;
ALTER TABLE tasks ADD COLUMN duration_minutes INT;
ALTER TABLE tasks ALTER COLUMN due_date TYPE TIMESTAMPTZ USING due_date::timestamptz;

-- Indexes
CREATE INDEX idx_labels_ws ON labels(workspace_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_task_labels_task ON task_labels(task_id);
CREATE INDEX idx_task_labels_label ON task_labels(label_id);
CREATE INDEX idx_sections_ws ON sections(workspace_id, area_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_parent ON tasks(parent_id) WHERE parent_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_tasks_section ON tasks(section_id) WHERE section_id IS NOT NULL AND deleted_at IS NULL;

-- RLS
ALTER TABLE labels ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_labels ON labels FOR ALL USING (is_workspace_member(workspace_id)) WITH CHECK (is_workspace_member(workspace_id));
ALTER TABLE sections ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_sections ON sections FOR ALL USING (is_workspace_member(workspace_id)) WITH CHECK (is_workspace_member(workspace_id));
