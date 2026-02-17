SET search_path TO widia_omni;

-- Remove entitlement limits
UPDATE plans SET limits = limits - 'max_projects';
UPDATE workspace_entitlements SET limits = limits - 'max_projects';

-- Drop triggers
DROP TRIGGER IF EXISTS trg_project_sections_updated_at ON project_sections;
DROP TRIGGER IF EXISTS trg_projects_updated_at ON project_sections;
DROP TRIGGER IF EXISTS trg_projects_dec ON projects;
DROP TRIGGER IF EXISTS trg_projects_inc ON projects;

-- Drop RLS
DROP POLICY IF EXISTS rls_project_sections ON project_sections;
ALTER TABLE project_sections DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS rls_projects ON projects;
ALTER TABLE projects DISABLE ROW LEVEL SECURITY;

-- Restore counter functions without projects
CREATE OR REPLACE FUNCTION widia_omni.increment_counter()
RETURNS TRIGGER AS $$
BEGIN
    CASE TG_TABLE_NAME
        WHEN 'life_areas' THEN UPDATE workspace_counters SET areas_count = areas_count + 1, updated_at = now() WHERE workspace_id = NEW.workspace_id;
        WHEN 'goals' THEN UPDATE workspace_counters SET goals_count = goals_count + 1, updated_at = now() WHERE workspace_id = NEW.workspace_id;
        WHEN 'habits' THEN UPDATE workspace_counters SET habits_count = habits_count + 1, updated_at = now() WHERE workspace_id = NEW.workspace_id;
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
        END CASE;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER SET search_path = widia_omni;

-- Drop indexes
DROP INDEX IF EXISTS idx_tasks_project;
DROP INDEX IF EXISTS idx_project_sections_project;
DROP INDEX IF EXISTS idx_projects_goal;
DROP INDEX IF EXISTS idx_projects_area;
DROP INDEX IF EXISTS idx_projects_workspace;

-- Remove task columns
ALTER TABLE tasks DROP COLUMN IF EXISTS project_section_id;
ALTER TABLE tasks DROP COLUMN IF EXISTS project_id;

-- Drop tables
DROP TABLE IF EXISTS project_sections;
DROP TABLE IF EXISTS projects;

-- Drop column from counters
ALTER TABLE workspace_counters DROP COLUMN IF EXISTS projects_count;

-- Drop enum
DROP TYPE IF EXISTS project_status;
