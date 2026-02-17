SET search_path TO widia_omni;

CREATE TYPE project_status AS ENUM ('planning','active','paused','completed','cancelled');

CREATE TABLE projects (
    id            UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id  UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    area_id       UUID REFERENCES life_areas(id) ON DELETE SET NULL,
    goal_id       UUID REFERENCES goals(id) ON DELETE SET NULL,
    title         TEXT NOT NULL,
    description   TEXT,
    status        project_status NOT NULL DEFAULT 'planning',
    color         TEXT NOT NULL DEFAULT 'blue',
    icon          TEXT NOT NULL DEFAULT 'folder',
    start_date    DATE,
    target_date   DATE,
    completed_at  TIMESTAMPTZ,
    is_archived   BOOLEAN NOT NULL DEFAULT false,
    archived_at   TIMESTAMPTZ,
    sort_order    INT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

CREATE TABLE project_sections (
    id          UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    position    INT NOT NULL DEFAULT 0,
    color       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE tasks ADD COLUMN project_id UUID REFERENCES projects(id) ON DELETE SET NULL;
ALTER TABLE tasks ADD COLUMN project_section_id UUID REFERENCES project_sections(id) ON DELETE SET NULL;

CREATE INDEX idx_projects_workspace ON projects(workspace_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_projects_area ON projects(area_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_projects_goal ON projects(goal_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_project_sections_project ON project_sections(project_id);
CREATE INDEX idx_tasks_project ON tasks(project_id) WHERE deleted_at IS NULL;

ALTER TABLE workspace_counters ADD COLUMN projects_count INT NOT NULL DEFAULT 0;

-- Extend counter triggers to handle projects
CREATE OR REPLACE FUNCTION widia_omni.increment_counter()
RETURNS TRIGGER AS $$
BEGIN
    CASE TG_TABLE_NAME
        WHEN 'life_areas' THEN UPDATE workspace_counters SET areas_count = areas_count + 1, updated_at = now() WHERE workspace_id = NEW.workspace_id;
        WHEN 'goals' THEN UPDATE workspace_counters SET goals_count = goals_count + 1, updated_at = now() WHERE workspace_id = NEW.workspace_id;
        WHEN 'habits' THEN UPDATE workspace_counters SET habits_count = habits_count + 1, updated_at = now() WHERE workspace_id = NEW.workspace_id;
        WHEN 'projects' THEN UPDATE workspace_counters SET projects_count = projects_count + 1, updated_at = now() WHERE workspace_id = NEW.workspace_id;
    END CASE;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER SET search_path = widia_omni;

CREATE OR REPLACE FUNCTION widia_omni.decrement_counter()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        CASE TG_TABLE_NAME
            WHEN 'life_areas' THEN UPDATE workspace_counters SET areas_count = GREATEST(areas_count - 1, 0), updated_at = now() WHERE workspace_id = NEW.workspace_id;
            WHEN 'goals' THEN UPDATE workspace_counters SET goals_count = GREATEST(goals_count - 1, 0), updated_at = now() WHERE workspace_id = NEW.workspace_id;
            WHEN 'habits' THEN UPDATE workspace_counters SET habits_count = GREATEST(habits_count - 1, 0), updated_at = now() WHERE workspace_id = NEW.workspace_id;
            WHEN 'projects' THEN UPDATE workspace_counters SET projects_count = GREATEST(projects_count - 1, 0), updated_at = now() WHERE workspace_id = NEW.workspace_id;
        END CASE;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER SET search_path = widia_omni;

CREATE TRIGGER trg_projects_inc AFTER INSERT ON projects FOR EACH ROW WHEN (NEW.deleted_at IS NULL) EXECUTE FUNCTION widia_omni.increment_counter();
CREATE TRIGGER trg_projects_dec AFTER UPDATE ON projects FOR EACH ROW EXECUTE FUNCTION widia_omni.decrement_counter();

-- RLS
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_projects ON projects FOR ALL
    USING (widia_omni.is_workspace_member(workspace_id))
    WITH CHECK (widia_omni.is_workspace_member(workspace_id));

ALTER TABLE project_sections ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_project_sections ON project_sections FOR ALL
    USING (EXISTS (SELECT 1 FROM projects p WHERE p.id = project_id AND widia_omni.is_workspace_member(p.workspace_id)))
    WITH CHECK (EXISTS (SELECT 1 FROM projects p WHERE p.id = project_id AND widia_omni.is_workspace_member(p.workspace_id)));

-- updated_at triggers
CREATE TRIGGER trg_projects_updated_at BEFORE UPDATE ON projects FOR EACH ROW EXECUTE FUNCTION widia_omni.update_updated_at();
CREATE TRIGGER trg_project_sections_updated_at BEFORE UPDATE ON project_sections FOR EACH ROW EXECUTE FUNCTION widia_omni.update_updated_at();

-- Entitlement limits: add max_projects to plans
UPDATE plans SET limits = limits || '{"max_projects": 3}'::jsonb WHERE tier = 'free';
UPDATE plans SET limits = limits || '{"max_projects": 15}'::jsonb WHERE tier = 'pro';
UPDATE plans SET limits = limits || '{"max_projects": -1}'::jsonb WHERE tier = 'premium';

-- Backfill existing workspace entitlements
UPDATE workspace_entitlements SET limits = limits || '{"max_projects": 3}'::jsonb
WHERE tier = 'free' AND NOT (limits ? 'max_projects');
UPDATE workspace_entitlements SET limits = limits || '{"max_projects": 15}'::jsonb
WHERE tier = 'pro' AND NOT (limits ? 'max_projects');
UPDATE workspace_entitlements SET limits = limits || '{"max_projects": -1}'::jsonb
WHERE tier = 'premium' AND NOT (limits ? 'max_projects');
