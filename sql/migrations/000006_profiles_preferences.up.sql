SET search_path TO widia_omni;

CREATE TABLE user_profiles (
    id                      UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
    display_name            TEXT NOT NULL,
    email                   TEXT NOT NULL,
    avatar_url              TEXT,
    timezone                TEXT NOT NULL DEFAULT 'America/Sao_Paulo',
    locale                  TEXT NOT NULL DEFAULT 'pt-BR',
    default_workspace_id    UUID REFERENCES workspaces(id) ON DELETE SET NULL,
    onboarding_completed    BOOLEAN NOT NULL DEFAULT false,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE user_preferences (
    user_id             UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
    week_starts_on      SMALLINT NOT NULL DEFAULT 1,
    daily_focus_limit   SMALLINT NOT NULL DEFAULT 5,
    notification_email  BOOLEAN NOT NULL DEFAULT true,
    notification_push   BOOLEAN NOT NULL DEFAULT true,
    weekly_review_day   SMALLINT NOT NULL DEFAULT 0,
    weekly_review_time  TIME NOT NULL DEFAULT '09:00',
    theme               TEXT NOT NULL DEFAULT 'dark',
    currency            TEXT NOT NULL DEFAULT 'BRL',
    score_weights       JSONB DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
