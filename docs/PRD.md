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
│   │   ├── 020_seed.sql
│   │   └── 021_budgets.sql
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
-- BUDGETS (entitlement gated: Pro+)
-- ============================================================
CREATE TABLE budgets (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    category_id     UUID REFERENCES finance_categories(id) ON DELETE CASCADE,
    month           TEXT NOT NULL,       -- YYYY-MM
    amount          NUMERIC(12,2) NOT NULL CHECK (amount >= 0),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Expression index for upsert ON CONFLICT (handles NULL category_id)
CREATE UNIQUE INDEX budgets_ws_cat_month
    ON budgets(workspace_id, COALESCE(category_id, '00000000-0000-0000-0000-000000000000'::uuid), month);
CREATE INDEX idx_budget_ws_month ON budgets(workspace_id, month);

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
GET    /api/v1/finances/summary          # ?month=YYYY-MM (default: current)
GET    /api/v1/finances/transactions     # ?type,category_id,area_id,date_from,date_to,tag,limit,offset
POST   /api/v1/finances/transactions     # [limit: max_transactions_per_month]
PUT    /api/v1/finances/transactions/:id
DELETE /api/v1/finances/transactions/:id # soft delete
GET    /api/v1/finances/categories
POST   /api/v1/finances/categories
PUT    /api/v1/finances/categories/:id
DELETE /api/v1/finances/categories/:id   # soft delete
GET    /api/v1/finances/budgets          # ?month=YYYY-MM (default: current)
POST   /api/v1/finances/budgets          # upsert (ON CONFLICT update)
DELETE /api/v1/finances/budgets/:id      # hard delete

# ─── Journal ───
GET    /api/v1/journal                   # paginated, date DESC
GET    /api/v1/journal/:date
PUT    /api/v1/journal/:date             # upsert (one per ws+date) [gate: journal_enabled]
DELETE /api/v1/journal/:date

# ─── Scores ───
GET    /api/v1/scores/current
GET    /api/v1/scores/history            # ?weeks=N [gate: score_history_weeks]

# ─── Notifications ───
GET    /api/v1/notifications             # ?unread=true&limit=N&offset=N
PATCH  /api/v1/notifications/:id/read
PATCH  /api/v1/notifications/read-all
GET    /api/v1/notifications/unread-count

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

### Milestone 2 — Core + Billing Skeleton ✅
- [x] Life Areas CRUD (soft delete, counter triggers)
- [x] Goals CRUD (hierarchy, progress)
- [x] Habits + Entries + Streaks
- [x] Tasks + Daily Focus
- [x] Dashboard aggregation
- [x] Onboarding wizard API
- [x] Stripe checkout + portal + webhook
- [x] stripe_events_processed idempotency
- [x] Webhook → subscription → entitlement sync
- [x] All create endpoints enforce limits via counters

### Milestone 3 — Engagement + Compliance ✅
- [x] Score engine (area + life scores, weighted formula)
- [x] asynq: score snapshots (weekly), streak updates (daily), counter reconciler (hourly)
- [x] Journal + mood/energy tracking (upsert per date)
- [x] Notifications (in-app CRUD, unread count, mark read/all)
- [x] Email stub (Sender interface + LogSender, swap for Resend later)
- [x] Audit log service (insert_audit_log SECURITY DEFINER)
- [x] Data export (GDPR, entitlement-gated)
- [x] Dashboard enhanced (life_score, journal_today, unread_notifications)

### Milestone 4 — Finance Module ✅
- [x] Finance categories CRUD (create, list, update, soft delete)
- [x] Transactions CRUD (create, list with filters, update, soft delete)
- [x] Monthly summary + analytics (totals by type, category breakdown, net balance)
- [x] Budget tracking (upsert via ON CONFLICT, list, hard delete, comparison in summary)
- [x] Transaction counter enforcement (monthly limit via counterSvc)
- [x] Finance entitlement gate on all endpoints (finance_enabled check)

### Milestone 5 — Scale & Polish ✅
- [x] Redis caching (dashboard, scores)
- [x] Storage usage tracking
- [x] Admin endpoints + entitlement overrides
- [x] Prometheus metrics refinement
- [x] OpenAPI spec complete

### Milestone 6 — Frontend ✅
- [x] React SPA scaffold (Vite + React 19 + TS + Tailwind v4 + shadcn/ui)
- [x] Auth pages (login, register, forgot/reset password)
- [x] App layout (sidebar + header + notifications)
- [x] Dashboard (life score ring, areas, goals, habits heatmap, today focus, mood, weekly stats)
- [x] Core CRUD pages (areas, goals, habits, tasks, journal, finances)
- [x] Landing page + pricing
- [x] Onboarding wizard UI
- [x] Billing management UI
- [x] Notifications panel
- [x] Settings (profile, preferences, workspace, account)

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

### M2 — Core + Billing Skeleton (2026-02-07)

19 new files + 1 modified, `go build` + `go vet` clean. All endpoints tested via curl against local Supabase.

| Phase | Files | Detail |
|-------|-------|--------|
| Phase 1 — Domain Models | 4 new | LifeArea, Goal (status/period enums), Habit + HabitEntry + HabitStreak (frequency enum), Task (priority enum) |
| Phase 2 — Services | 7 new | AreaService (CRUD + reorder + limit enforcement), GoalService (CRUD + filtered list + auto-complete on progress), HabitService (CRUD + check-in upsert + entries + streak SQL), TaskService (CRUD + complete + toggle focus + daily limit + counter increment), BillingService (Stripe checkout/portal/webhook + idempotency + entitlement sync), OnboardingService (status + batch setup + complete), DashboardService (aggregated counts) |
| Phase 3 — Handlers | 8 new | area, goal, habit, task, billing, stripe_webhook (isolated, no auth), onboarding, dashboard — all follow existing pattern |
| Phase 4 — Router | 1 modified | Wired all services/handlers, registered routes under `/api/v1` (protected) + `/webhooks/stripe` (public), finance routes placeholder for M4 |
| Phase 5 — Dependencies | go.mod | Added `github.com/stripe/stripe-go/v81` |
| Bugfix — DATE scanning | 3 domain files | pgx binary format can't scan DATE into `string` → changed `StartDate`, `EndDate`, `DueDate`, `Date`, `TasksTodayDate` to `time.Time` |

**Key behaviors verified:**
- Limit enforcement: free plan (4 areas) blocks 5th create, soft delete decrements counter allowing new create
- Dashboard aggregation: areas count, active goals, today tasks, completed today, habits today, streaks
- Habit streaks: SQL window function calculation
- Goal auto-complete: progress update auto-sets status=completed when target reached
- Stripe webhook: signature verification + idempotency via `stripe_events_processed` + subscription lifecycle → entitlement sync
- Onboarding: batch setup (areas/goals/habits) + completion flag

### M3 — Engagement + Compliance (2026-02-08)

19 new files + 6 modified, `go build` + `go vet` clean. All endpoints tested via curl against local Supabase + Redis.

| Phase | Files | Detail |
|-------|-------|--------|
| 3A — Journal + Notifications + Scores | 11 new, 1 modified | JournalEntry, AreaScore, LifeScore, ScoreBreakdown, Notification (type/channel enums), AuditEntry domain models. JournalService (list/get/upsert/delete, one entry per ws+date). ScoreService (history, current, calculate with habit 50%/goal 30%/task 20% formula). NotificationService (list/create/markRead/markAllRead/unreadCount). AuditService (insert_audit_log wrapper). Handlers for journal, score, notification. Router wiring. |
| 3B — Worker Infrastructure | 6 new, 2 modified | asynq-based worker binary (full rewrite from stub). ScoreSnapshotHandler (weekly Monday 2am, all active workspaces). StreakUpdateHandler (daily 1am, detects at-risk streaks, creates notifications). CounterReconcilerHandler (hourly, fixes areas/goals/habits count drift). SendNotificationHandler (email dispatch via interface). Email Sender interface + LogSender stub. Added `hibiken/asynq` dependency. |
| 3C — GDPR Export + Dashboard | 2 new, 3 modified | ExportService (all workspace data as JSON, entitlement-gated). ExportHandler (POST /me/export). DashboardService enhanced with life_score (nullable), journal_today (bool), unread_notifications (count). DashboardHandler updated to pass userID. Router updated with export route. |
| Bugfix — interval cast | 1 file | pgx binary protocol can't concat int param with string for interval → changed to `make_interval(weeks => $2)` |

**Key behaviors verified:**
- Journal CRUD: upsert per workspace+date, list paginated desc, delete
- Notifications: list (with unread filter), unread count, mark read, mark all read
- Scores: current (graceful nil when empty), history (parameterized weeks)
- Dashboard: life_score=null before first compute, journal_today=true after entry, unread_notifications count
- Export: blocked on free plan (export_enabled=false), returns full data on pro+
- Entitlement gating: journal_enabled, export_enabled, score_history_weeks all enforced

### M4 — Finance Module (2026-02-08)

5 new files + 1 modified, `go build` + `go vet` clean. All 24 endpoint tests passed via curl against local Supabase + Redis.

| Area | Files | Detail |
|------|-------|--------|
| Migration | 2 new (up/down) | `000021_budgets` — budgets table with expression unique index on (workspace_id, COALESCE(category_id, nil_uuid), month), RLS policy |
| Domain | 1 new | FinanceCategory, Transaction (TransactionType enum), Budget models |
| Service | 1 new | FinanceService — categories CRUD, transactions CRUD with dynamic filters (type/category/area/date/tag/limit/offset), budgets CRUD with upsert ON CONFLICT, monthly summary (totals by type, category breakdown, budget comparison with % spent) |
| Handler | 1 new | FinanceHandler — 12 endpoints, all gated with `finance_enabled` entitlement check. `financeGate()` helper for DRY gate logic. Counter-checked transaction creates. |
| Router | 1 modified | Wired FinanceService + FinanceHandler, registered `/finances` route group with 12 routes |

**Key behaviors verified:**
- Finance gate: free plan returns 403 "finance not available on your plan"
- Categories: create (income/expense/investment), list (3), update name, soft delete (3→2)
- Transactions: create with counter enforcement, list all (4), filter by type (2 expenses), filter by date range (2), filter by tag (2 with "food"), pagination (limit=2 offset=1), update amount, soft delete (4→3)
- Budgets: upsert creates new budget, upsert same category+month updates amount (600→600), list by month (2), hard delete
- Summary: income=5000, expenses=375, investments=1000, net_balance=3625, 3 category breakdowns, 2 budget comparisons (Groceries 62.5% spent, Stocks 0%)
- Counter: transactions_month_count=4 after 4 creates, transactions_month=2026-02

### M5 — Scale & Polish (2026-02-08)

21 files changed (3 new + 18 modified), `go build` + `go vet` clean.

| Phase | Files | Detail |
|-------|-------|--------|
| Phase 1 — Redis Caching | 4 modified | DashboardService: added `rdb *redis.Client`, cache key `ws:{wsID}:dash` TTL 2min. ScoreService: added `rdb *redis.Client`, cache `ws:{wsID}:score:current` + `ws:{wsID}:score:hist:{weeks}` TTL 10min, `InvalidateCache` with SCAN. ScoreSnapshotWorker: calls InvalidateCache after Calculate. Router: updated constructors to pass rdb. |
| Phase 2 — Storage Tracking | 2 modified | CounterService: `IncrementStorage(ctx, wsID, bytes)` and `DecrementStorage(ctx, wsID, bytes)` with GREATEST(0, ...) floor. EntitlementLimits: `CanUseStorage(currentBytes, additionalBytes)` converting StorageMB to bytes. |
| Phase 3 — Admin Endpoints | 3 new, 1 modified | AdminAuth middleware (X-Service-Key header). AdminService: GetMetrics (users/workspaces/subs by tier), ListUsers (paginated join), GetUser (detail + counters + entitlement), GetWorkspaceUsage (counters + limits), OverrideEntitlement (tx: deactivate current → insert new source='admin' → invalidate cache). AdminHandler (5 endpoints). Router: `/admin` group with 5 routes. |
| Phase 4 — Prometheus Metrics | 11 modified | 4 new metrics: `active_subscriptions` GaugeVec (tier), `asynq_queue_depth` GaugeVec (queue), `asynq_job_failures_total` CounterVec (task_type), `entitlement_limit_reached_total` CounterVec (limit_type). `RefreshSubscriptionGauge` func (5min goroutine in api). Worker: metrics HTTP on :9090, Inspector polling queue depths every 30s. 4 worker files: increment failure counter on errors. 5 service files (area/goal/habit/task/finance): increment limit counter on rejection. |
| Phase 5 — OpenAPI Spec | 1 rewritten | Full `api/openapi.yaml` (~2800 lines): 60+ endpoints, 30+ schemas, 18 tags, bearerAuth + serviceKeyAuth, standard error responses (400/401/403/404/429), all request/response bodies. |

### M6 — Frontend React SPA (2026-02-08)

87 files added under `web/`, `npm run build` clean (1.69s), all pages verified via Playwright E2E.

| Phase | Files | Detail |
|-------|-------|--------|
| Phase 1 — Scaffold | ~25 | Vite + React 19 + TS + Tailwind v4 + shadcn/ui. Design tokens from `docs/sample-page.html` (dark theme, Poppins/Lora/JetBrains Mono). api-client.ts (fetch wrapper, token injection, 401 refresh). Zustand stores (auth, ui). TanStack Query v5. React Router v7 with lazy routes. |
| Phase 2 — Auth Pages | 6 | Login, register, forgot-password, reset-password. Auth hooks (useLogin, useRegister). AuthLayout centered card. Post-login redirect: onboarding if incomplete, else dashboard. |
| Phase 3 — App Layout | 4 | 72px fixed sidebar (7 nav items with lucide icons + tooltips), collapsible. Header with greeting (Bom dia/Boa tarde/Boa noite), date (JetBrains Mono), notification bell, avatar initials. |
| Phase 4 — Dashboard | 12 | Life score animated SVG ring. Areas grid (3x2 cards with progress). Goals card with progress bars. Habits heatmap (7-day grid). Today focus list with complete toggle. Mood selector (5 emojis). Weekly stats (3 stat cards). Responsive breakpoints at 1200px/900px. |
| Phase 5 — Core CRUD | 18 | Areas (card grid + create/edit dialog). Goals (list + filters + progress). Habits (list + check-in + streak badge). Tasks (list + checkbox + priority/area filters). Journal (date nav + mood/energy + wins/challenges/gratitude). Finances (4 tabs: transactions, categories, budgets, summary). |
| Phase 6 — Landing Page | 7 | Public marketing page: nav, hero with gradient text, 6 feature cards, 3 pricing tiers (Gratis R$0, Pro R$19, Premium R$39), footer. |
| Phase 7 — Onboarding Wizard | 2 | 4-step wizard: select areas → create goals → create habits → complete. Progress bar. Resumes from correct step via status API. |
| Phase 8 — Billing | 3 | Subscription status card, plan comparison grid (handles `-1` as unlimited), Stripe checkout redirect, Stripe portal link. Success/cancel callback pages. |
| Phase 9 — Notifications | 2 | Bell icon with unread count badge. Dropdown panel with emoji type icons, mark read, mark all read, click-outside dismiss. 60s polling. |
| Phase 10 — Settings | 2 | 4 tabs: Profile (name, email disabled, timezone), Preferences (currency, focus limit, week start), Workspace (name + usage progress bars), Account (data export, danger zone with delete confirmation). |

**Bugs found & fixed during E2E testing:**
- Onboarding crash: `status?.steps.habits` accessed before query resolved → deferred `derivedStep` after `isLoading` guard
- Billing "-1 areas" display: backend uses `-1` for unlimited → fixed condition to `< 0 || >= 999`
