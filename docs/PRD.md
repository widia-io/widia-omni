# Mission Control — Arquitetura SaaS (v2)

## 🎯 Visão do Produto

Mission Control é uma plataforma SaaS de gestão de vida pessoal.
Cada workspace tem seu espaço isolado com metas, hábitos, finanças, journaling e um Life Score que consolida tudo.

**Modelo de negócio:** Freemium → Pro → Premium

---

## 📐 Arquitetura de Alto Nível

```
                    ┌──────────────────────────────────────┐
                    │           CDN (Cloudflare)            │
                    │     Landing Page + App Frontend       │
                    └──────────────┬───────────────────────┘
                                   │
                    ┌──────────────▼───────────────────────┐
                    │          TRAEFIK (Reverse Proxy)      │
                    │    mc.app → frontend                  │
                    │    api.mc.app → go backend             │
                    └──────┬───────────────┬───────────────┘
                           │               │
          ┌────────────────▼──┐   ┌────────▼──────────────┐
          │   FRONTEND (SPA)  │   │     GO API SERVER      │
          │   React + Vite    │   │                        │
          │                   │   │  ┌──────────────────┐  │
          │  - Landing page   │   │  │  Auth Middleware  │  │
          │  - Onboarding     │   │  │  (JWT → tenant)   │  │
          │  - Dashboard      │   │  ├──────────────────┤  │
          │  - Settings       │   │  │  Rate Limiter     │  │
          │  - Billing        │   │  │  (token bucket    │  │
          │                   │   │  │   per ws+route)   │  │
          └───────────────────┘   │  ├──────────────────┤  │
                                  │  │  Entitlement Gate │  │
                                  │  │  (limits via      │  │
                                  │  │   counters O(1))  │  │
                                  │  ├──────────────────┤  │
                                  │  │  Handlers         │  │
                                  │  │  Services         │  │
                                  │  │  Repository       │  │
                                  │  └──────────────────┘  │
                                  └──┬──────┬──────┬───────┘
                                     │      │      │
                    ┌────────────────▼┐  ┌──▼──┐  ┌▼────────────┐
                    │   SUPABASE      │  │REDIS│  │  STRIPE      │
                    │                 │  │     │  │  (billing)   │
                    │  PostgreSQL     │  │cache│  └──────────────┘
                    │  Auth           │  │rate │
                    │  Realtime       │  │limit│  ┌──────────────┐
                    │  Storage        │  │     │  │  RESEND       │
                    │                 │  └─────┘  │  (emails)    │
                    └─────────────────┘           └──────────────┘
                                  │
                    ┌─────────────▼────────────────────────┐
                    │         GO WORKER (asynq)             │
                    │                                      │
                    │  Scheduled (cron):                    │
                    │   - score_snapshot  (weekly)          │
                    │   - streak_update   (daily)           │
                    │   - weekly_review   (weekly)          │
                    │   - trial_expiry    (daily)           │
                    │   - counter_reconciler (hourly)       │
                    │                                      │
                    │  Event-driven:                        │
                    │   - send_notification                 │
                    │   - process_stripe_event              │
                    │   - export_user_data                  │
                    │                                      │
                    │  DLQ + retries (3x, exp backoff)     │
                    └──────────────────────────────────────┘
```

---

## 🏗️ Multi-Tenancy: Workspace-First

**Decisão: workspace_id em todas as tabelas de negócio desde o dia 1.**

Hoje 1 user = 1 workspace. Amanhã suporta Family/Couple/Team sem migrar o banco.

```
Camada 1 — Database (RLS)
  └─ Todas tabelas filtradas por workspace_id
  └─ is_workspace_member() → SECURITY DEFINER function
  └─ Service role NUNCA exposto ao client

Camada 2 — Application (Tenant Context)
  └─ JWT → user_id → workspace_id + role + entitlements
  └─ Todo repo recebe workspace_id obrigatório

Camada 3 — Entitlement Gate
  └─ Entitlements = fonte de verdade (não o plano direto)
  └─ Counters O(1) para limites (sem COUNT(*))
  └─ Rate limit: token bucket por workspace_id + route_group
```

---

## 🗂️ Estrutura do Projeto Go

```
mission-control/
├── cmd/
│   ├── api/main.go                     # API server
│   ├── worker/main.go                  # asynq worker
│   └── migrate/main.go                # Migrations
│
├── internal/
│   ├── config/config.go               # Env vars + defaults
│   │
│   ├── middleware/
│   │   ├── auth.go                     # JWT validation → user ctx
│   │   ├── tenant.go                   # user → workspace → entitlements ctx
│   │   ├── entitlement.go             # Feature gate + limit check (O(1))
│   │   ├── ratelimit.go               # Token bucket per ws + route group
│   │   ├── idempotency.go            # Idempotency-Key header
│   │   ├── requestid.go
│   │   ├── logger.go                  # Request logging
│   │   └── cors.go
│   │
│   ├── handler/
│   │   ├── auth.go
│   │   ├── onboarding.go
│   │   ├── user.go
│   │   ├── workspace.go
│   │   ├── billing.go
│   │   ├── stripe_webhook.go          # Isolado: sem auth, com signature
│   │   ├── area.go
│   │   ├── goal.go
│   │   ├── habit.go
│   │   ├── task.go
│   │   ├── finance.go
│   │   ├── journal.go
│   │   ├── dashboard.go
│   │   ├── notification.go
│   │   ├── export.go
│   │   ├── admin.go
│   │   └── health.go
│   │
│   ├── service/
│   │   ├── auth.go
│   │   ├── onboarding.go
│   │   ├── user.go
│   │   ├── workspace.go
│   │   ├── entitlement.go             # Derive entitlements from subscription
│   │   ├── billing.go                 # Stripe integration
│   │   ├── counter.go                 # Counter management + reconciler
│   │   ├── area.go
│   │   ├── goal.go
│   │   ├── habit.go
│   │   ├── task.go
│   │   ├── score.go
│   │   ├── finance.go
│   │   ├── journal.go
│   │   ├── dashboard.go
│   │   ├── notification.go
│   │   ├── export.go
│   │   └── audit.go                   # Audit via SECURITY DEFINER
│   │
│   ├── repository/sqlc/               # Auto-generated
│   │
│   ├── domain/
│   │   ├── workspace.go               # Workspace + Members + Role
│   │   ├── user.go
│   │   ├── entitlement.go            # Entitlements + source enum
│   │   ├── plan.go
│   │   ├── counter.go                # WorkspaceCounters
│   │   ├── area.go
│   │   ├── goal.go
│   │   ├── habit.go
│   │   ├── task.go
│   │   ├── finance.go
│   │   ├── journal.go
│   │   ├── score.go
│   │   └── notification.go
│   │
│   ├── billing/
│   │   ├── stripe.go                  # Stripe client
│   │   ├── webhook.go                # Signature + idempotency + out-of-order
│   │   └── plans.go
│   │
│   ├── worker/
│   │   ├── tasks.go                   # asynq task type constants
│   │   ├── score_snapshot.go
│   │   ├── streak_update.go
│   │   ├── weekly_review.go
│   │   ├── trial_expiry.go
│   │   ├── counter_reconciler.go     # Fix drift hourly
│   │   ├── send_notification.go
│   │   ├── process_stripe_event.go
│   │   └── export_data.go
│   │
│   ├── email/
│   │   ├── resend.go
│   │   └── templates.go
│   │
│   ├── observability/
│   │   ├── logger.go                  # zerolog structured JSON
│   │   ├── metrics.go                # Prometheus
│   │   └── tracing.go               # OpenTelemetry (optional)
│   │
│   └── router/router.go
│
├── api/openapi.yaml                   # OpenAPI 3.1 spec
│
├── sql/
│   ├── migrations/
│   │   ├── 001_extensions.sql
│   │   ├── 002_workspaces.sql
│   │   ├── 003_plans_subscriptions.sql
│   │   ├── 004_entitlements.sql
│   │   ├── 005_counters.sql
│   │   ├── 006_profiles_preferences.sql
│   │   ├── 007_areas.sql
│   │   ├── 008_goals.sql
│   │   ├── 009_habits.sql
│   │   ├── 010_tasks.sql
│   │   ├── 011_finances.sql
│   │   ├── 012_journal.sql
│   │   ├── 013_scores.sql
│   │   ├── 014_notifications.sql
│   │   ├── 015_audit.sql
│   │   ├── 016_stripe_events.sql
│   │   ├── 017_rls.sql
│   │   ├── 018_indexes.sql
│   │   ├── 019_triggers.sql
│   │   └── 020_seed.sql
│   └── queries/
│       ├── workspaces.sql
│       ├── entitlements.sql
│       ├── counters.sql
│       ├── users.sql
│       ├── areas.sql
│       ├── goals.sql
│       ├── habits.sql
│       ├── tasks.sql
│       ├── finances.sql
│       ├── journal.sql
│       ├── dashboard.sql
│       ├── audit.sql
│       └── admin.sql
│
├── sqlc.yaml
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── .env.example
```

---

## 🗃️ Database Schema

```sql
-- ============================================================
-- EXTENSIONS
-- ============================================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gist";

-- ============================================================
-- WORKSPACES (tenant principal)
-- ============================================================
CREATE TABLE workspaces (
    id          UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    name        TEXT NOT NULL,
    slug        TEXT NOT NULL UNIQUE,
    owner_id    UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TYPE workspace_role AS ENUM ('owner', 'admin', 'member', 'viewer');

CREATE TABLE workspace_members (
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    role         workspace_role NOT NULL DEFAULT 'member',
    joined_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (workspace_id, user_id)
);

CREATE INDEX idx_ws_members_user ON workspace_members(user_id);

-- ============================================================
-- USER PROFILES
-- ============================================================
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

-- ============================================================
-- PLANS & SUBSCRIPTIONS (espelho do Stripe)
-- ============================================================
CREATE TYPE plan_tier AS ENUM ('free', 'pro', 'premium');
CREATE TYPE subscription_status AS ENUM (
    'trialing', 'active', 'past_due',
    'canceled', 'paused', 'unpaid'
);

CREATE TABLE plans (
    id                    UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    tier                  plan_tier NOT NULL UNIQUE,
    name                  TEXT NOT NULL,
    price_monthly         NUMERIC(8,2) NOT NULL DEFAULT 0,
    price_yearly          NUMERIC(8,2) NOT NULL DEFAULT 0,
    stripe_price_monthly  TEXT,
    stripe_price_yearly   TEXT,
    limits                JSONB NOT NULL DEFAULT '{}',
    is_active             BOOLEAN NOT NULL DEFAULT true,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE subscriptions (
    id                      UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id            UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    plan_id                 UUID NOT NULL REFERENCES plans(id),
    tier                    plan_tier NOT NULL,            -- denormalized
    status                  subscription_status NOT NULL DEFAULT 'active',
    stripe_customer_id      TEXT,
    stripe_subscription_id  TEXT,
    stripe_price_id         TEXT,
    currency                TEXT NOT NULL DEFAULT 'BRL',
    current_period_start    TIMESTAMPTZ,
    current_period_end      TIMESTAMPTZ,
    trial_end               TIMESTAMPTZ,
    cancel_at               TIMESTAMPTZ,
    canceled_at             TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_subs_active_ws
    ON subscriptions(workspace_id) WHERE status IN ('trialing', 'active', 'past_due');

-- ============================================================
-- ENTITLEMENTS (fonte de verdade do que o workspace pode usar)
-- ============================================================
CREATE TYPE entitlement_source AS ENUM ('free', 'stripe', 'trial', 'promo', 'admin');

CREATE TABLE workspace_entitlements (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    tier            plan_tier NOT NULL DEFAULT 'free',
    limits          JSONB NOT NULL,
    source          entitlement_source NOT NULL DEFAULT 'free',
    effective_from  TIMESTAMPTZ NOT NULL DEFAULT now(),
    effective_to    TIMESTAMPTZ,
    is_current      BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_ent_current_ws
    ON workspace_entitlements(workspace_id) WHERE is_current = true;

-- ============================================================
-- WORKSPACE COUNTERS (O(1) limit checks)
-- ============================================================
CREATE TABLE workspace_counters (
    workspace_id                UUID PRIMARY KEY REFERENCES workspaces(id) ON DELETE CASCADE,
    areas_count                 INT NOT NULL DEFAULT 0,
    goals_count                 INT NOT NULL DEFAULT 0,
    habits_count                INT NOT NULL DEFAULT 0,
    tasks_created_today         INT NOT NULL DEFAULT 0,
    tasks_today_date            DATE NOT NULL DEFAULT CURRENT_DATE,
    transactions_month_count    INT NOT NULL DEFAULT 0,
    transactions_month          TEXT NOT NULL DEFAULT to_char(CURRENT_DATE, 'YYYY-MM'),
    storage_bytes_used          BIGINT NOT NULL DEFAULT 0,
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- USER PREFERENCES
-- ============================================================
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

-- ============================================================
-- LIFE AREAS
-- ============================================================
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

-- ============================================================
-- GOALS
-- ============================================================
CREATE TYPE goal_status AS ENUM (
    'not_started', 'on_track', 'at_risk',
    'behind', 'completed', 'cancelled'
);
CREATE TYPE goal_period AS ENUM ('yearly', 'quarterly', 'monthly', 'weekly');

CREATE TABLE goals (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    area_id         UUID REFERENCES life_areas(id) ON DELETE SET NULL,
    parent_id       UUID REFERENCES goals(id) ON DELETE CASCADE,
    title           TEXT NOT NULL,
    description     TEXT,
    period          goal_period NOT NULL DEFAULT 'quarterly',
    status          goal_status NOT NULL DEFAULT 'not_started',
    target_value    NUMERIC,
    current_value   NUMERIC DEFAULT 0,
    unit            TEXT,
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

-- ============================================================
-- HABITS
-- ============================================================
CREATE TYPE habit_frequency AS ENUM ('daily', 'weekly', 'custom');

CREATE TABLE habits (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    area_id         UUID REFERENCES life_areas(id) ON DELETE SET NULL,
    name            TEXT NOT NULL,
    color           TEXT NOT NULL DEFAULT '#788c5d',
    frequency       habit_frequency NOT NULL DEFAULT 'daily',
    target_per_week INT NOT NULL DEFAULT 3,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE TABLE habit_entries (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    habit_id        UUID NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    date            DATE NOT NULL,
    intensity       SMALLINT NOT NULL DEFAULT 3 CHECK (intensity BETWEEN 1 AND 3),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(habit_id, date)
);

-- ============================================================
-- TASKS
-- ============================================================
CREATE TYPE task_priority AS ENUM ('low', 'medium', 'high', 'critical');

CREATE TABLE tasks (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    area_id         UUID REFERENCES life_areas(id) ON DELETE SET NULL,
    goal_id         UUID REFERENCES goals(id) ON DELETE SET NULL,
    title           TEXT NOT NULL,
    description     TEXT,
    priority        task_priority NOT NULL DEFAULT 'medium',
    is_completed    BOOLEAN NOT NULL DEFAULT false,
    is_focus        BOOLEAN NOT NULL DEFAULT false,
    due_date        DATE,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

-- ============================================================
-- FINANCES (entitlement gated: Pro+)
-- ============================================================
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

-- ============================================================
-- JOURNAL
-- ============================================================
CREATE TABLE journal_entries (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    date            DATE NOT NULL,
    mood            SMALLINT CHECK (mood BETWEEN 1 AND 5),
    energy          SMALLINT CHECK (energy BETWEEN 1 AND 5),
    wins            TEXT[],
    challenges      TEXT[],
    gratitude       TEXT[],
    notes           TEXT,
    tags            TEXT[],
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(workspace_id, date)
);

-- ============================================================
-- SCORES
-- ============================================================
CREATE TABLE area_scores (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    area_id         UUID NOT NULL REFERENCES life_areas(id) ON DELETE CASCADE,
    score           SMALLINT NOT NULL CHECK (score BETWEEN 0 AND 100),
    week_start      DATE NOT NULL,
    breakdown       JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(workspace_id, area_id, week_start)
);

CREATE TABLE life_scores (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    score           SMALLINT NOT NULL CHECK (score BETWEEN 0 AND 100),
    week_start      DATE NOT NULL,
    area_scores     JSONB NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(workspace_id, week_start)
);

-- ============================================================
-- NOTIFICATIONS
-- ============================================================
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

-- ============================================================
-- AUDIT LOG (insert via SECURITY DEFINER only)
-- ============================================================
CREATE TABLE audit_log (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id    UUID REFERENCES workspaces(id) ON DELETE SET NULL,
    user_id         UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    action          TEXT NOT NULL,
    entity_type     TEXT,
    entity_id       UUID,
    metadata        JSONB,
    ip_address      INET,
    user_agent      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION insert_audit_log(
    p_workspace_id UUID, p_user_id UUID, p_action TEXT,
    p_entity_type TEXT DEFAULT NULL, p_entity_id UUID DEFAULT NULL,
    p_metadata JSONB DEFAULT NULL, p_ip INET DEFAULT NULL, p_ua TEXT DEFAULT NULL
) RETURNS UUID AS $$
DECLARE v_id UUID;
BEGIN
    INSERT INTO audit_log (workspace_id, user_id, action, entity_type, entity_id, metadata, ip_address, user_agent)
    VALUES (p_workspace_id, p_user_id, p_action, p_entity_type, p_entity_id, p_metadata, p_ip, p_ua)
    RETURNING id INTO v_id;
    RETURN v_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- ============================================================
-- STRIPE EVENTS (webhook idempotency)
-- ============================================================
CREATE TABLE stripe_events_processed (
    event_id        TEXT PRIMARY KEY,
    event_type      TEXT NOT NULL,
    processed_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- INDEXES
-- ============================================================
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

-- ============================================================
-- ROW LEVEL SECURITY
-- ============================================================
CREATE OR REPLACE FUNCTION is_workspace_member(ws_id UUID)
RETURNS BOOLEAN AS $$
    SELECT EXISTS (
        SELECT 1 FROM workspace_members
        WHERE workspace_id = ws_id AND user_id = auth.uid()
    );
$$ LANGUAGE sql SECURITY DEFINER STABLE;

-- Workspaces
ALTER TABLE workspaces ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_workspaces ON workspaces
    FOR ALL USING (is_workspace_member(id)) WITH CHECK (owner_id = auth.uid());

-- Workspace members
ALTER TABLE workspace_members ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_ws_members ON workspace_members
    FOR ALL USING (is_workspace_member(workspace_id));

-- User profiles (own only)
ALTER TABLE user_profiles ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_profiles ON user_profiles
    FOR ALL USING (auth.uid() = id) WITH CHECK (auth.uid() = id);

-- User preferences (own only)
ALTER TABLE user_preferences ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_prefs ON user_preferences
    FOR ALL USING (auth.uid() = user_id) WITH CHECK (auth.uid() = user_id);

-- All workspace-scoped business tables
DO $$
DECLARE tbl TEXT;
BEGIN
    FOR tbl IN SELECT unnest(ARRAY[
        'subscriptions', 'workspace_entitlements', 'workspace_counters',
        'life_areas', 'goals', 'habits', 'habit_entries',
        'tasks', 'finance_categories', 'transactions',
        'journal_entries', 'area_scores', 'life_scores'
    ])
    LOOP
        EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', tbl);
        EXECUTE format(
            'CREATE POLICY rls_%s ON %I FOR ALL USING (is_workspace_member(workspace_id)) WITH CHECK (is_workspace_member(workspace_id))',
            tbl, tbl
        );
    END LOOP;
END;
$$;

-- Notifications (own only + workspace member)
ALTER TABLE notifications ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_notif ON notifications
    FOR ALL USING (auth.uid() = user_id AND is_workspace_member(workspace_id));

-- Audit log (select by workspace, insert via function only)
ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_audit ON audit_log
    FOR SELECT USING (is_workspace_member(workspace_id));

-- Plans (public read)
ALTER TABLE plans ENABLE ROW LEVEL SECURITY;
CREATE POLICY rls_plans ON plans FOR SELECT USING (true);

-- ============================================================
-- COUNTER TRIGGERS
-- ============================================================
CREATE OR REPLACE FUNCTION increment_counter()
RETURNS TRIGGER AS $$
BEGIN
    CASE TG_TABLE_NAME
        WHEN 'life_areas' THEN
            UPDATE workspace_counters SET areas_count = areas_count + 1, updated_at = now() WHERE workspace_id = NEW.workspace_id;
        WHEN 'goals' THEN
            UPDATE workspace_counters SET goals_count = goals_count + 1, updated_at = now() WHERE workspace_id = NEW.workspace_id;
        WHEN 'habits' THEN
            UPDATE workspace_counters SET habits_count = habits_count + 1, updated_at = now() WHERE workspace_id = NEW.workspace_id;
    END CASE;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE OR REPLACE FUNCTION decrement_counter()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        CASE TG_TABLE_NAME
            WHEN 'life_areas' THEN
                UPDATE workspace_counters SET areas_count = GREATEST(areas_count - 1, 0), updated_at = now() WHERE workspace_id = NEW.workspace_id;
            WHEN 'goals' THEN
                UPDATE workspace_counters SET goals_count = GREATEST(goals_count - 1, 0), updated_at = now() WHERE workspace_id = NEW.workspace_id;
            WHEN 'habits' THEN
                UPDATE workspace_counters SET habits_count = GREATEST(habits_count - 1, 0), updated_at = now() WHERE workspace_id = NEW.workspace_id;
        END CASE;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE TRIGGER trg_areas_inc AFTER INSERT ON life_areas FOR EACH ROW WHEN (NEW.deleted_at IS NULL) EXECUTE FUNCTION increment_counter();
CREATE TRIGGER trg_areas_dec AFTER UPDATE ON life_areas FOR EACH ROW EXECUTE FUNCTION decrement_counter();
CREATE TRIGGER trg_goals_inc AFTER INSERT ON goals FOR EACH ROW WHEN (NEW.deleted_at IS NULL) EXECUTE FUNCTION increment_counter();
CREATE TRIGGER trg_goals_dec AFTER UPDATE ON goals FOR EACH ROW EXECUTE FUNCTION decrement_counter();
CREATE TRIGGER trg_habits_inc AFTER INSERT ON habits FOR EACH ROW WHEN (NEW.deleted_at IS NULL) EXECUTE FUNCTION increment_counter();
CREATE TRIGGER trg_habits_dec AFTER UPDATE ON habits FOR EACH ROW EXECUTE FUNCTION decrement_counter();

-- ============================================================
-- UPDATED_AT TRIGGERS
-- ============================================================
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = now(); RETURN NEW; END;
$$ LANGUAGE plpgsql;

DO $$
DECLARE tbl TEXT;
BEGIN
    FOR tbl IN SELECT unnest(ARRAY[
        'workspaces', 'user_profiles', 'user_preferences',
        'subscriptions', 'life_areas', 'goals', 'habits',
        'tasks', 'journal_entries', 'workspace_counters'
    ])
    LOOP
        EXECUTE format(
            'CREATE TRIGGER trg_%s_updated_at BEFORE UPDATE ON %I FOR EACH ROW EXECUTE FUNCTION update_updated_at()',
            tbl, tbl
        );
    END LOOP;
END;
$$;

-- ============================================================
-- AUTO-SETUP ON SIGNUP
-- ============================================================
CREATE OR REPLACE FUNCTION handle_new_user()
RETURNS TRIGGER AS $$
DECLARE
    v_ws_id UUID;
    v_plan_id UUID;
    v_limits JSONB;
    v_name TEXT;
BEGIN
    v_name := COALESCE(NEW.raw_user_meta_data->>'display_name', split_part(NEW.email, '@', 1));

    INSERT INTO workspaces (name, slug, owner_id)
    VALUES (v_name || '''s Space', 'ws-' || substr(gen_random_uuid()::text, 1, 12), NEW.id)
    RETURNING id INTO v_ws_id;

    INSERT INTO workspace_members (workspace_id, user_id, role) VALUES (v_ws_id, NEW.id, 'owner');
    INSERT INTO user_profiles (id, display_name, email, default_workspace_id) VALUES (NEW.id, v_name, NEW.email, v_ws_id);
    INSERT INTO user_preferences (user_id) VALUES (NEW.id);

    SELECT id, limits INTO v_plan_id, v_limits FROM plans WHERE tier = 'free' LIMIT 1;
    INSERT INTO subscriptions (workspace_id, plan_id, tier, status) VALUES (v_ws_id, v_plan_id, 'free', 'active');
    INSERT INTO workspace_entitlements (workspace_id, tier, limits, source, is_current) VALUES (v_ws_id, 'free', v_limits, 'free', true);
    INSERT INTO workspace_counters (workspace_id) VALUES (v_ws_id);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE TRIGGER on_auth_user_created
    AFTER INSERT ON auth.users
    FOR EACH ROW EXECUTE FUNCTION handle_new_user();

-- ============================================================
-- SEED: PLANS
-- ============================================================
INSERT INTO plans (tier, name, price_monthly, price_yearly, limits) VALUES
('free', 'Free', 0, 0, '{
    "max_areas": 4, "max_goals": 5, "max_habits": 5,
    "max_tasks_per_day": 5, "max_transactions_per_month": 0,
    "journal_enabled": true, "finance_enabled": false,
    "export_enabled": false, "score_history_weeks": 4,
    "api_rate_limit_per_minute": 30, "storage_mb": 50,
    "ai_insights": false, "api_access": false
}'::jsonb),
('pro', 'Pro', 19.90, 199.00, '{
    "max_areas": 8, "max_goals": 25, "max_habits": 15,
    "max_tasks_per_day": 20, "max_transactions_per_month": 500,
    "journal_enabled": true, "finance_enabled": true,
    "export_enabled": true, "score_history_weeks": 52,
    "api_rate_limit_per_minute": 120, "storage_mb": 500,
    "ai_insights": false, "api_access": false
}'::jsonb),
('premium', 'Premium', 39.90, 399.00, '{
    "max_areas": -1, "max_goals": -1, "max_habits": -1,
    "max_tasks_per_day": -1, "max_transactions_per_month": -1,
    "journal_enabled": true, "finance_enabled": true,
    "export_enabled": true, "score_history_weeks": -1,
    "api_rate_limit_per_minute": 300, "storage_mb": 5000,
    "ai_insights": true, "api_access": true
}'::jsonb);
```

---

## 🔗 API Routes

```
# ─── Health / Observability ───
GET    /health
GET    /ready
GET    /metrics                         # Prometheus

# ─── Auth ───
POST   /auth/register
POST   /auth/login
POST   /auth/refresh
POST   /auth/logout
POST   /auth/forgot-password
POST   /auth/reset-password
POST   /auth/verify-email

# ─── Stripe Webhook (isolated) ───
POST   /webhooks/stripe                 # Signature + idempotency

# ─── Onboarding ───
GET    /api/v1/onboarding/status
POST   /api/v1/onboarding/areas
POST   /api/v1/onboarding/goals
POST   /api/v1/onboarding/habits
POST   /api/v1/onboarding/complete

# ─── User ───
GET    /api/v1/me
PUT    /api/v1/me
GET    /api/v1/me/preferences
PUT    /api/v1/me/preferences
DELETE /api/v1/me                       # GDPR
POST   /api/v1/me/export               # GDPR [idempotent]

# ─── Workspace ───
GET    /api/v1/workspace
PUT    /api/v1/workspace
GET    /api/v1/workspace/members        # Future: family
GET    /api/v1/workspace/usage          # Counters vs limits

# ─── Billing ───
GET    /api/v1/billing/plans
GET    /api/v1/billing/subscription
GET    /api/v1/billing/entitlements
POST   /api/v1/billing/checkout         # [idempotent]
POST   /api/v1/billing/portal

# ─── Dashboard ───
GET    /api/v1/dashboard
GET    /api/v1/dashboard/weekly-review

# ─── Areas ───
GET    /api/v1/areas
POST   /api/v1/areas                    # [limit: max_areas]
PUT    /api/v1/areas/:id
PATCH  /api/v1/areas/:id/reorder
DELETE /api/v1/areas/:id                # soft delete

# ─── Goals ───
GET    /api/v1/goals
GET    /api/v1/goals/:id
POST   /api/v1/goals                    # [limit: max_goals]
PUT    /api/v1/goals/:id
DELETE /api/v1/goals/:id
PATCH  /api/v1/goals/:id/progress

# ─── Habits ───
GET    /api/v1/habits
POST   /api/v1/habits                   # [limit: max_habits]
PUT    /api/v1/habits/:id
DELETE /api/v1/habits/:id
GET    /api/v1/habits/entries
POST   /api/v1/habits/:id/check-in
DELETE /api/v1/habits/:id/check-in/:date
GET    /api/v1/habits/streaks

# ─── Tasks ───
GET    /api/v1/tasks
POST   /api/v1/tasks                    # [limit: max_tasks_per_day]
PUT    /api/v1/tasks/:id
DELETE /api/v1/tasks/:id
PATCH  /api/v1/tasks/:id/complete
PATCH  /api/v1/tasks/:id/focus

# ─── Finances ─── [gate: finance_enabled]
GET    /api/v1/finances/summary
GET    /api/v1/finances/transactions
POST   /api/v1/finances/transactions    # [limit: max_transactions_per_month]
PUT    /api/v1/finances/transactions/:id
DELETE /api/v1/finances/transactions/:id
GET    /api/v1/finances/categories
POST   /api/v1/finances/categories

# ─── Journal ───
GET    /api/v1/journal
GET    /api/v1/journal/:date
POST   /api/v1/journal
PUT    /api/v1/journal/:date

# ─── Scores ───
GET    /api/v1/scores/history           # [gate: score_history_weeks]

# ─── Notifications ───
GET    /api/v1/notifications
PATCH  /api/v1/notifications/:id/read
PATCH  /api/v1/notifications/read-all

# ─── Admin (service role) ───
GET    /admin/metrics
GET    /admin/users
GET    /admin/users/:id
GET    /admin/workspaces/:id/usage
POST   /admin/entitlements/override
```

---

## 💳 Planos & Entitlements

```
┌──────────────────┬──────────┬──────────┬───────────┐
│     Feature      │   FREE   │   PRO    │  PREMIUM  │
├──────────────────┼──────────┼──────────┼───────────┤
│ Áreas            │    4     │    8     │ Ilimitado │
│ Metas            │    5     │   25     │ Ilimitado │
│ Hábitos          │    5     │   15     │ Ilimitado │
│ Tarefas/dia      │    5     │   20     │ Ilimitado │
│ Journal          │    ✓     │    ✓     │     ✓     │
│ Finanças         │    ✗     │    ✓     │     ✓     │
│ Score History    │  4 sem   │  1 ano   │ Ilimitado │
│ Export           │    ✗     │    ✓     │     ✓     │
│ AI Insights      │    ✗     │    ✗     │     ✓     │
│ API Access       │    ✗     │    ✗     │     ✓     │
│ Storage          │  50 MB   │  500 MB  │   5 GB    │
│ Members          │    1     │    1     │    5      │
│ Preço/mês        │  Grátis  │ R$19,90  │  R$39,90  │
│ Preço/ano        │  Grátis  │ R$199    │  R$399    │
└──────────────────┴──────────┴──────────┴───────────┘
```

### Entitlement Flow

```
Stripe Event (webhook)
   │
   ▼
stripe_events_processed (check event_id → idempotent)
   │ new?
   ▼
subscriptions (update status, period, stripe IDs)
   │
   ▼
workspace_entitlements (old.is_current=false, insert new with is_current=true)
   │
   ▼
Redis cache invalidation (ws:{id}:entitlements)
   │
   ▼
Next request loads new entitlements → new limits applied
```

---

## 📊 Observability

```
Logs:     zerolog → JSON → stdout (collected by Docker/Traefik)
          request_id + workspace_id + user_id em todo request

Metrics:  Prometheus /metrics
          http_request_duration_seconds (p50, p95, p99)
          http_requests_total (by status, method)
          active_subscriptions (by tier)
          asynq_queue_depth
          asynq_job_failures_total
          entitlement_limit_reached_total

Tracing:  OpenTelemetry (optional, plug when needed)
```

---

## 🐳 Docker

```yaml
services:
  api:
    build: { context: ., target: production }
    ports: ["8080:8080"]
    environment:
      - SUPABASE_URL=${SUPABASE_URL}
      - SUPABASE_SERVICE_KEY=${SUPABASE_SERVICE_KEY}
      - SUPABASE_JWT_SECRET=${SUPABASE_JWT_SECRET}
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=redis://redis:6379
      - STRIPE_SECRET_KEY=${STRIPE_SECRET_KEY}
      - STRIPE_WEBHOOK_SECRET=${STRIPE_WEBHOOK_SECRET}
      - RESEND_API_KEY=${RESEND_API_KEY}
    depends_on: { redis: { condition: service_healthy } }
    deploy: { replicas: 2, resources: { limits: { memory: 256M } } }
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.mc-api.rule=Host(`api.missioncontrol.app`)"
      - "traefik.http.routers.mc-api.tls.certresolver=letsencrypt"
      - "traefik.http.services.mc-api.loadbalancer.server.port=8080"

  worker:
    build: { context: ., target: production }
    command: ["/mission-control", "worker"]
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=redis://redis:6379
      - RESEND_API_KEY=${RESEND_API_KEY}
      - STRIPE_SECRET_KEY=${STRIPE_SECRET_KEY}
    depends_on: { redis: { condition: service_healthy } }

  redis:
    image: redis:7-alpine
    command: redis-server --maxmemory 128mb --maxmemory-policy allkeys-lru
    volumes: [redis_data:/data]
    healthcheck: { test: ["CMD", "redis-cli", "ping"], interval: 10s }

volumes:
  redis_data:
```

---

## 📦 Dependencies

```
github.com/go-chi/chi/v5              # Router
github.com/jackc/pgx/v5               # Postgres
github.com/redis/go-redis/v9          # Redis
github.com/hibiken/asynq              # Job queue
github.com/golang-jwt/jwt/v5          # JWT
github.com/stripe/stripe-go/v81       # Stripe
github.com/rs/zerolog                  # Logging
github.com/caarlos0/env/v11           # Config
github.com/prometheus/client_golang    # Metrics
go.opentelemetry.io/otel              # Tracing (optional)
```

---

## 🚀 Milestones

### Milestone 1 — Foundation ✅
- [x] Supabase + full migrations
- [x] Go scaffold: config, Chi, zerolog
- [x] Auth middleware (JWT → user → workspace → entitlements)
- [x] Entitlement gate middleware (feature + limit check O(1))
- [x] Rate limit (token bucket per workspace + route group)
- [x] Request ID + CORS + observability logger
- [x] User profile + preferences CRUD
- [x] Health / ready / metrics endpoints
- [x] Docker + Traefik
- [x] OpenAPI spec skeleton

### Milestone 2 — Core + Billing Skeleton
- [ ] Life Areas CRUD (soft delete, counter triggers)
- [ ] Goals CRUD (hierarchy, progress)
- [ ] Habits + Entries + Streaks
- [ ] Tasks + Daily Focus
- [ ] Dashboard aggregation
- [ ] Onboarding wizard API
- [ ] Stripe checkout + portal + webhook
- [ ] stripe_events_processed idempotency
- [ ] Webhook → subscription → entitlement sync
- [ ] All create endpoints enforce limits via counters

### Milestone 3 — Engagement + Compliance
- [ ] Score engine (area + life scores)
- [ ] asynq: score snapshots, streak updates, counter reconciler
- [ ] Journal + mood tracking
- [ ] Notifications (in-app + email Resend)
- [ ] Weekly review + trial expiry emails
- [ ] Audit log (SECURITY DEFINER)
- [ ] Data export + account deletion (GDPR)

### Milestone 4 — Finance Module
- [ ] Finance categories + transactions CRUD
- [ ] Monthly summary + analytics
- [ ] Budget tracking
- [ ] Transaction counter triggers

### Milestone 5 — Scale & Polish
- [ ] Redis caching (dashboard, entitlements, scores)
- [ ] Counter reconciler worker
- [ ] Storage usage tracking
- [ ] Admin endpoints + entitlement overrides
- [ ] Prometheus metrics refinement
- [ ] OpenAPI spec complete

### Milestone 6 — Frontend
- [ ] React SPA (from HTML mock)
- [ ] Landing page + pricing
- [ ] Onboarding wizard UI
- [ ] Billing management UI

### Milestone 7 — Growth
- [ ] AI Insights (Premium)
- [ ] Public API (Premium)
- [ ] Family plan (workspace members)
- [ ] Mobile app
- [ ] Referral system

---

## Changelog

### M1 — Foundation (2026-02-07)

~80 files, all compiling (`go build` + `go vet` clean).

| Phase | Arquivos | Detalhe |
|-------|----------|---------|
| Phase 0 — Project Init | go.mod, .gitignore, .env.example, Makefile, dirs | Go 1.24, chi/v5, pgx/v5, go-redis/v9, zerolog, caarlos0/env, prometheus |
| Phase 1 — Config + Observability | config.go, logger.go, metrics.go | Env parsing via struct tags, zerolog JSON, Prometheus histogram + counter |
| Phase 2 — Migrations | 40 SQL files (20 up/down) | Extensions, workspaces, plans, entitlements, counters, profiles, areas, goals, habits, tasks, finances, journal, scores, notifications, audit, stripe events, RLS, indexes, triggers, seed |
| Phase 3 — sqlc | sqlc.yaml + 4 query files | workspaces.sql, entitlements.sql, counters.sql, users.sql |
| Phase 4 — Domain | 5 models | Workspace, User, Entitlement (com helpers CanCreate*), Plan, Counter |
| Phase 5 — Middleware | 7 files + context.go | RequestID, Logger, CORS, Auth (JWT/HMAC), Tenant (Redis cache 5min), Entitlement gate, RateLimit (token bucket) |
| Phase 6 — Services | 5 services | Auth (Supabase REST direto), User, Workspace, Entitlement (cache + derive), Counter |
| Phase 7 — Handlers | 5 handlers | Health (/health, /ready), Auth (register/login/refresh/logout/forgot/reset/verify), User (CRUD + prefs), Workspace (get/update/usage), Response helpers |
| Phase 8 — Router + Server | router.go, cmd/api, cmd/worker | Chi router com middleware stack, graceful shutdown, worker stub |
| Phase 9 — Docker | Dockerfile, docker-compose.yml, .dockerignore | Multi-stage build, Traefik v3, Redis 7, replicas 2, LetsEncrypt |
| Phase 10 — OpenAPI | api/openapi.yaml | OpenAPI 3.1, JWT bearer, todos endpoints M1, schemas completos |

**Pendente para rodar:**
1. Setup Supabase em `/Users/bruno/Developer/infra`
2. `.env` com valores reais
3. `make migrate-up`
4. `make dev` → `curl localhost:8080/health`
