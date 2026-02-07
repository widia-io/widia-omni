SET search_path TO widia_omni;

CREATE TYPE notification_channel AS ENUM ('in_app', 'email', 'push');
CREATE TYPE notification_type AS ENUM (
    'weekly_review', 'streak_at_risk', 'goal_deadline',
    'trial_ending', 'plan_changed', 'score_update',
    'habit_reminder', 'system'
);

CREATE TABLE notifications (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    type            notification_type NOT NULL,
    channel         notification_channel NOT NULL DEFAULT 'in_app',
    title           TEXT NOT NULL,
    body            TEXT,
    data            JSONB,
    is_read         BOOLEAN NOT NULL DEFAULT false,
    read_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
