SET search_path TO widia_omni;

CREATE INDEX idx_goals_ws_period ON goals(workspace_id, period, start_date, end_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_goals_parent ON goals(parent_id) WHERE parent_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_goals_active ON goals(workspace_id, status) WHERE status NOT IN ('completed', 'cancelled') AND deleted_at IS NULL;

CREATE INDEX idx_habit_entries_ws ON habit_entries(workspace_id, date DESC);
CREATE INDEX idx_habit_entries_habit ON habit_entries(habit_id, date DESC);
CREATE INDEX idx_habits_active ON habits(workspace_id) WHERE is_active = true AND deleted_at IS NULL;

CREATE INDEX idx_tasks_ws ON tasks(workspace_id, due_date, is_completed) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_focus ON tasks(workspace_id, is_focus, due_date) WHERE is_focus = true AND is_completed = false AND deleted_at IS NULL;
CREATE INDEX idx_tasks_pending ON tasks(workspace_id, due_date) WHERE is_completed = false AND deleted_at IS NULL;

CREATE INDEX idx_txn_ws_date ON transactions(workspace_id, date DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_txn_ws_type ON transactions(workspace_id, type, date DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_txn_ws_cat ON transactions(workspace_id, category_id, date DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_txn_tags ON transactions USING GIN(tags) WHERE deleted_at IS NULL;

CREATE INDEX idx_journal_ws ON journal_entries(workspace_id, date DESC);
CREATE INDEX idx_area_scores_ws ON area_scores(workspace_id, week_start DESC);
CREATE INDEX idx_life_scores_ws ON life_scores(workspace_id, week_start DESC);
CREATE INDEX idx_notif_unread ON notifications(workspace_id, user_id, created_at DESC) WHERE is_read = false;
CREATE INDEX idx_audit_ws ON audit_log(workspace_id, created_at DESC);
CREATE INDEX idx_audit_entity ON audit_log(entity_type, entity_id);
CREATE INDEX idx_stripe_events_date ON stripe_events_processed(processed_at);
