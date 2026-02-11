SET search_path TO widia_omni;

-- Drop RLS
DROP POLICY IF EXISTS rls_sections ON sections;
DROP POLICY IF EXISTS rls_labels ON labels;

-- Drop indexes
DROP INDEX IF EXISTS idx_tasks_section;
DROP INDEX IF EXISTS idx_tasks_parent;
DROP INDEX IF EXISTS idx_sections_ws;
DROP INDEX IF EXISTS idx_task_labels_label;
DROP INDEX IF EXISTS idx_task_labels_task;
DROP INDEX IF EXISTS idx_labels_ws;

-- Revert tasks columns
ALTER TABLE tasks ALTER COLUMN due_date TYPE DATE USING due_date::date;
ALTER TABLE tasks DROP COLUMN IF EXISTS duration_minutes;
ALTER TABLE tasks DROP COLUMN IF EXISTS position;
ALTER TABLE tasks DROP COLUMN IF EXISTS section_id;
ALTER TABLE tasks DROP COLUMN IF EXISTS parent_id;

-- Drop tables
DROP TABLE IF EXISTS task_labels;
DROP TABLE IF EXISTS sections;
DROP TABLE IF EXISTS labels;
