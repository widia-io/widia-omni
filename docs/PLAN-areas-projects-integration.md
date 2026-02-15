# Plan: Areas, Projects & Integration Rework

> Status: IN PROGRESS | Created: 2026-02-14 | Updated: 2026-02-15 | Phase 3 done: 2026-02-15

## Current State Analysis

### What We Have
| Entity   | CRUD | Hierarchy | Area Link | Goal Link | Entitlements |
|----------|------|-----------|-----------|-----------|-------------|
| Areas    | Full | No        | -         | -         | Free:4 Pro:8 |
| Goals    | Full | parent_id | Optional  | -         | Free:5 Pro:25 |
| Tasks    | Full | parent_id | Optional  | Optional  | Daily limit  |
| Habits   | Full | No        | Optional  | -         | Free:10      |
| Sections | Full | No        | Required  | -         | -            |
| Labels   | Full | No        | -         | -         | -            |

### Problems Identified

**1. Areas are "flat containers"**
- After onboarding, areas are just name+icon+color
- No area detail view aggregating goals/tasks/habits/projects
- No way to see area health/progress at a glance
- Areas feel static — created once and forgotten

**2. Onboarding friction with areas**
- Users start from zero — no templates or suggestions
- Must invent area names/icons/colors from scratch
- No guidance on what a "good set of areas" looks like
- High cognitive load at the very first interaction

**3. No Projects concept**
- Tasks can be grouped by sections within areas, but sections are flat
- Goals track outcomes, but there's nothing for "initiatives" (a group of related tasks working toward a goal)
- Gap between high-level goals and day-to-day tasks
- Example: Goal = "Launch SaaS product", but no place to organize "Build landing page", "Setup Stripe", "Write docs" as a cohesive unit

**4. Weak goal → task connection**
- Tasks can link to a goal_id, but it's just a reference
- No project-level progress tracking
- No way to see "what % of work toward this goal is done"

---

## Proposed Solution

### Phase 1: Area Templates & Onboarding Improvement ✅ DONE
**Impact: Immediate UX win, reduces onboarding drop-off**
**Completed:** 2026-02-15 | **Branch:** `feat/area-templates-projects` | **Commit:** `628db2d`

#### 1.1 Area Template Catalog ✅
8 area templates with i18n (pt-BR + en-US), stored in code (`internal/service/templates.go`).
Uses Lucide icon names + design token colors (matching frontend conventions).

| Area (pt-BR) | Slug | Icon | Color |
|---|---|---|---|
| Saúde & Fitness | saude-fitness | dumbbell | green |
| Carreira & Trabalho | carreira-trabalho | briefcase | orange |
| Finanças | financas | dollar-sign | sand |
| Relacionamentos | relacionamentos | users | rose |
| Desenvolvimento Pessoal | desenvolvimento-pessoal | brain | sage |
| Casa & Ambiente | casa-ambiente | home | blue |
| Lazer & Diversão | lazer-diversao | gamepad-2 | sky |
| Espiritualidade | espiritualidade | sparkles | violet |

**API:**
- `GET /api/v1/onboarding/area-templates?locale=pt-BR` → returns template catalog
- `GET /api/v1/onboarding/goal-suggestions?locale=pt-BR&area_slug=saude-fitness` → returns 2-3 suggestions per area
- Locale fallback: unknown → en-US, pt-* → pt-BR

#### 1.2 Onboarding Flow Refinement ✅
- Step 0: area cards fetched from API (removed hardcoded `SUGGESTED_AREAS`)
- Step 1: goal suggestions shown per selected area as tappable chips + custom input with period badge
- Step 2: habits (unchanged)
- Step 3: complete

**Files changed:**
| File | Action |
|---|---|
| `internal/service/templates.go` | Created — templates data + i18n resolution |
| `internal/handler/onboarding.go` | +2 handlers (GetAreaTemplates, GetGoalSuggestions) |
| `internal/router/router.go` | +2 GET routes |
| `web/src/lib/icons.ts` | +5 Lucide icons (Dumbbell, Brain, Home, Gamepad2, Sparkles) |
| `web/src/types/api.ts` | +AreaTemplate, GoalSuggestion interfaces |
| `web/src/hooks/use-onboarding.ts` | +useAreaTemplates, useGoalSuggestions hooks |
| `web/src/pages/onboarding.tsx` | Rewritten steps 0+1, color token map, icon rendering |

---

### Phase 2: Projects (New Entity) ✅ DONE
**Impact: Core feature — bridges goals and tasks**
**Completed:** 2026-02-15 | **Branch:** `feat/area-templates-projects` | **Commit:** `1b4d689`

#### 2.1 Data Model

```sql
CREATE TYPE project_status AS ENUM (
    'planning', 'active', 'paused', 'completed', 'cancelled'
);

CREATE TABLE projects (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    area_id         UUID REFERENCES life_areas(id) ON DELETE SET NULL,
    goal_id         UUID REFERENCES goals(id) ON DELETE SET NULL,
    title           TEXT NOT NULL,
    description     TEXT,
    status          project_status NOT NULL DEFAULT 'planning',
    color           TEXT NOT NULL DEFAULT '#6366f1',
    icon            TEXT NOT NULL DEFAULT '📁',
    start_date      DATE,
    target_date     DATE,
    completed_at    TIMESTAMPTZ,
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

-- Tasks get a new optional FK
ALTER TABLE tasks ADD COLUMN project_id UUID REFERENCES projects(id) ON DELETE SET NULL;
```

**Relationships:**
```
Area (optional)
└── Project
    ├── linked to Goal (optional) — "this project serves this goal"
    ├── Tasks (project_id FK)
    └── Sections can be reused within project context
```

#### 2.2 Domain Model

```go
type Project struct {
    ID          uuid.UUID      `json:"id"`
    WorkspaceID uuid.UUID      `json:"workspace_id"`
    AreaID      *uuid.UUID     `json:"area_id,omitempty"`
    GoalID      *uuid.UUID     `json:"goal_id,omitempty"`
    Title       string         `json:"title"`
    Description *string        `json:"description,omitempty"`
    Status      ProjectStatus  `json:"status"`
    Color       string         `json:"color"`
    Icon        string         `json:"icon"`
    StartDate   *time.Time     `json:"start_date,omitempty"`
    TargetDate  *time.Time     `json:"target_date,omitempty"`
    CompletedAt *time.Time     `json:"completed_at,omitempty"`
    SortOrder   int            `json:"sort_order"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    DeletedAt   *time.Time     `json:"deleted_at,omitempty"`
    // Computed (not stored)
    TasksTotal     int `json:"tasks_total,omitempty"`
    TasksCompleted int `json:"tasks_completed,omitempty"`
}
```

#### 2.3 API Endpoints

```
GET    /api/v1/projects              (filters: area_id, goal_id, status)
POST   /api/v1/projects
GET    /api/v1/projects/{id}         (includes task counts)
PUT    /api/v1/projects/{id}
DELETE /api/v1/projects/{id}
PATCH  /api/v1/projects/{id}/reorder
```

**Task changes:**
- `POST /api/v1/tasks` accepts `project_id`
- `GET /api/v1/tasks` accepts `?project_id=` filter
- `PUT /api/v1/tasks/{id}` can set/change project_id

#### 2.4 Entitlements

| Tier    | Max Projects |
|---------|-------------|
| Free    | 3           |
| Pro     | 15          |
| Premium | Unlimited   |

**Backend work:**
- Migration: new table + tasks column + counter trigger + entitlement defaults
- Domain model, service, handler (follow existing patterns)
- Update entitlement model with `MaxProjects`
- Update counter model with `projects_count`
- Update dashboard to include `active_projects` count
- Add to public API read routes

---

### Phase 3: Area Detail & Aggregation ✅ DONE
**Impact: Makes areas feel alive and useful**
**Completed:** 2026-02-15 | **Branch:** `main`

#### 3.1 Area Summary Endpoint ✅

`GET /api/v1/areas/{id}/summary` — aggregated stats per area with Redis cache (2min TTL).

Response includes: `goals_active`, `goals_completed`, `projects_active`, `projects_completed`, `tasks_pending`, `tasks_completed_this_week`, `habits_active`, `current_streak_avg`, `area_score`.

#### 3.2 Area GetByID ✅

`GET /api/v1/areas/{id}` — single area fetch.

#### 3.3 Enhanced Area List ✅

`GET /api/v1/areas?include=stats` — inline `goals_count`, `projects_count`, `tasks_pending`, `area_score` per area.

Plain `GET /api/v1/areas` unchanged.

**Files changed:**
| File | Action |
|---|---|
| `internal/service/area.go` | +`AreaStats`, `AreaSummary`, `AreaWithStats` types, +`rdb`, +`GetByID`, `GetSummary`, `ListWithStats` |
| `internal/handler/area.go` | +`GetByID`, `GetSummary` handlers, `List` supports `?include=stats` |
| `internal/router/router.go` | +2 protected routes, +2 public routes, pass `rdb` to area service |

---

### Phase 4: Goal ↔ Project ↔ Task Integration
**Impact: Creates a coherent planning hierarchy**

#### 4.1 Goal Progress from Projects

When a goal has linked projects, calculate goal progress from project task completion:

```
Goal: "Launch SaaS by Q2"
├── Project: "Build MVP"          (8/12 tasks done = 67%)
├── Project: "Marketing Site"     (3/5 tasks done = 60%)
└── Direct tasks (no project)     (2/3 done = 67%)

Aggregate progress: 13/20 = 65%
```

**Auto-update `goals.current_value`** when:
- A task linked to a project linked to a goal is completed
- A direct task linked to a goal is completed
- Set `target_value = total tasks`, `current_value = completed tasks`, `unit = "tasks"`

This is **opt-in** — only applies when goal has `target_value IS NULL` (no manual tracking).

#### 4.2 Cascade Area Assignment

When creating a task/project within an area context, auto-fill `area_id`:
- Creating a task inside a project that belongs to an area → auto-set task.area_id
- Creating a project under a goal that belongs to an area → auto-set project.area_id

This keeps the hierarchy consistent without requiring users to manually tag everything.

---

## Implementation Order & Effort

| Phase | Description                     | Migrations | Services | Handlers | Effort | Status |
|-------|---------------------------------|-----------|----------|----------|--------|--------|
| 1     | Area templates + onboarding     | 0         | 1 new    | 1 modify | Small  | ✅ Done |
| 2     | Projects entity                 | 1         | 1 new + 2 modify | 1 new | Medium | ✅ Done |
| 3     | Area summary/aggregation        | 0         | 1 modify | 1 modify | Small  | ✅ Done |
| 4     | Goal↔Project↔Task integration   | 0         | 2 modify | 0        | Medium | Pending |

**Sequence:** ~~Phase 1~~ → ~~Phase 2~~ → ~~Phase 3~~ → Phase 4

Phase 4 is next — ties goals, projects, and tasks into a coherent hierarchy with auto-progress.

---

## Entity Relationship After Implementation

```
Workspace
├── Area (templates available at onboarding)
│   ├── Goal (optional area link)
│   │   ├── Sub-goal (parent_id)
│   │   └── Project (optional goal link)
│   │       └── Task (project_id)
│   ├── Project (optional area link, optional goal link)
│   │   ├── Task (project_id)
│   │   │   └── Subtask (parent_id)
│   │   └── Section (task organizer)
│   ├── Habit (optional area link)
│   ├── Section (task organizer)
│   └── Task (direct, no project)
└── Unlinked entities (no area)
    ├── Goal, Project, Task, Habit
```

---

## Resolved Decisions

| # | Question | Decision |
|---|----------|----------|
| 1 | Project sections: own or reuse? | **Own sections** — projects get their own `project_sections` table |
| 2 | Kanban board view? | **Yes** — project sections = kanban columns |
| 3 | Archive concept? | **Separate** — `is_archived` boolean, distinct from `status=completed` |
| 4 | Area templates localized? | **Yes** — pt-BR and en-US |
| 5 | Goal hierarchy depth? | **Max 3 levels** — enforced at service layer |
| 6 | Auto-update goal progress? | **Yes** — auto-update when project tasks complete |

### Decision Details

#### D1: Project Sections (Kanban Columns)

```sql
CREATE TABLE project_sections (
    id           UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    project_id   UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    position     INT NOT NULL DEFAULT 0,
    color        TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Tasks link to project sections for kanban placement
ALTER TABLE tasks ADD COLUMN project_section_id UUID
    REFERENCES project_sections(id) ON DELETE SET NULL;
```

Default columns on project creation: `To Do`, `In Progress`, `Done`.

**API:**
```
GET    /api/v1/projects/{id}/sections
POST   /api/v1/projects/{id}/sections
PUT    /api/v1/projects/{id}/sections/{sectionId}
DELETE /api/v1/projects/{id}/sections/{sectionId}
PATCH  /api/v1/projects/{id}/sections/{sectionId}/reorder
```

#### D3: Archive Concept

```sql
ALTER TABLE projects ADD COLUMN is_archived BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE projects ADD COLUMN archived_at TIMESTAMPTZ;
```

- `status=completed` → project finished successfully, visible in area summary
- `is_archived=true` → hidden from default views, accessible via filter
- Archived projects don't count toward entitlement limits

**API:**
```
PATCH /api/v1/projects/{id}/archive
PATCH /api/v1/projects/{id}/unarchive
GET   /api/v1/projects?include_archived=true
```

#### D4: Area Templates i18n ✅ IMPLEMENTED

Templates use **Lucide icon names** + **design token colors** (not emojis/hex) to match frontend conventions.
Locale from `?locale=` query param, fallback: unknown → en-US, pt-* → pt-BR.

See `internal/service/templates.go` for full data. Also includes 2-3 goal suggestions per area per locale.

#### D5: Goal Depth Limit (3 Levels)

```
Level 1: "Get Healthy in 2026"         (area: Health)
  Level 2: "Lose 15kg by June"         (parent: Level 1)
    Level 3: "Cut sugar this month"    (parent: Level 2)
      Level 4: ❌ BLOCKED              (service returns 422)
```

Enforced in `GoalService.Create()`:
```go
func (s *GoalService) validateDepth(ctx context.Context, parentID uuid.UUID) error {
    depth := 1
    current := parentID
    for depth < 3 {
        parent, err := s.getParent(ctx, current)
        if err != nil || parent == nil { return nil }
        current = *parent
        depth++
    }
    return ErrMaxGoalDepthReached
}
```

#### D6: Auto Goal Progress

When a task is completed/reopened and it belongs to a project linked to a goal:

```
task.Complete() →
  if task.project_id != nil:
    project = getProject(task.project_id)
    if project.goal_id != nil:
      goal = getGoal(project.goal_id)
      if goal.target_value == nil:  // auto-tracking mode
        total = countProjectTasks(project.goal_id)
        completed = countCompletedProjectTasks(project.goal_id)
        updateGoalProgress(goal.id, completed, total)
        if completed == total:
          goal.status = "completed"
```

Also applies to direct tasks linked to a goal (no project intermediary).

---

## Updated Phase 2: Projects (Complete Spec)

### Migration (single file)

```sql
-- 000026_projects.up.sql

CREATE TYPE project_status AS ENUM (
    'planning', 'active', 'paused', 'completed', 'cancelled'
);

CREATE TABLE projects (
    id              UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    area_id         UUID REFERENCES life_areas(id) ON DELETE SET NULL,
    goal_id         UUID REFERENCES goals(id) ON DELETE SET NULL,
    title           TEXT NOT NULL,
    description     TEXT,
    status          project_status NOT NULL DEFAULT 'planning',
    color           TEXT NOT NULL DEFAULT '#6366f1',
    icon            TEXT NOT NULL DEFAULT '📁',
    start_date      DATE,
    target_date     DATE,
    completed_at    TIMESTAMPTZ,
    is_archived     BOOLEAN NOT NULL DEFAULT false,
    archived_at     TIMESTAMPTZ,
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE TABLE project_sections (
    id           UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    project_id   UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    position     INT NOT NULL DEFAULT 0,
    color        TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Tasks get project columns
ALTER TABLE tasks ADD COLUMN project_id UUID
    REFERENCES projects(id) ON DELETE SET NULL;
ALTER TABLE tasks ADD COLUMN project_section_id UUID
    REFERENCES project_sections(id) ON DELETE SET NULL;

-- Indexes
CREATE INDEX idx_projects_workspace ON projects(workspace_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_projects_area ON projects(area_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_projects_goal ON projects(goal_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_project_sections_project ON project_sections(project_id);
CREATE INDEX idx_tasks_project ON tasks(project_id) WHERE deleted_at IS NULL;

-- Counter
ALTER TABLE workspace_counters ADD COLUMN projects_count INT NOT NULL DEFAULT 0;

-- Counter triggers (same pattern as existing)
CREATE TRIGGER trg_projects_inc
    AFTER INSERT ON projects
    FOR EACH ROW
    WHEN (NEW.deleted_at IS NULL)
    EXECUTE FUNCTION inc_counter('projects_count');

CREATE TRIGGER trg_projects_dec
    AFTER UPDATE OF deleted_at ON projects
    FOR EACH ROW
    WHEN (OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL)
    EXECUTE FUNCTION dec_counter('projects_count');

-- Entitlement column
ALTER TABLE workspace_entitlements ADD COLUMN max_projects INT NOT NULL DEFAULT 3;

-- Update existing free entitlements
UPDATE workspace_entitlements SET max_projects = 3 WHERE source = 'free';
```

### Full API Surface

```
# Projects CRUD
GET    /api/v1/projects                          ?area_id=&goal_id=&status=&include_archived=
POST   /api/v1/projects
GET    /api/v1/projects/{id}                     (includes task counts, sections)
PUT    /api/v1/projects/{id}
DELETE /api/v1/projects/{id}
PATCH  /api/v1/projects/{id}/reorder
PATCH  /api/v1/projects/{id}/archive
PATCH  /api/v1/projects/{id}/unarchive

# Project Sections (Kanban Columns)
GET    /api/v1/projects/{id}/sections
POST   /api/v1/projects/{id}/sections
PUT    /api/v1/projects/{id}/sections/{sectionId}
DELETE /api/v1/projects/{id}/sections/{sectionId}
PATCH  /api/v1/projects/{id}/sections/{sectionId}/reorder

# Tasks within project (uses existing task endpoints with filter)
GET    /api/v1/tasks?project_id={id}             (list view)
GET    /api/v1/tasks?project_id={id}&group_by=section  (kanban view)

# Public API
GET    /public/v1/projects
GET    /public/v1/projects/{id}
```

---

## Dev Environment: Worktree-Isolated Database ✅ DONE

Worktree has its own Supabase stack with isolated DB, so migrations don't touch main.

### Layout

```
Developer/infra/supabase/          ← main (ports 54321-54323, Redis 6379)
Developer/infra/supabase-wt/       ← worktree (ports 54331-54333, Redis 6389)
Developer/widia/widia-omni/        ← main repo (PORT=8080)
Developer/widia/widia-omni-wt/     ← worktree (PORT=8081, branch feat/area-templates-projects)
```

### Key details

- **Symlinks** (directory-level, not file-level — Docker on macOS can't bind-mount file symlinks):
  - `volumes/db/init/` → `../supabase/volumes/db/init/`
  - `volumes/api/` → `../supabase/volumes/api/`
  - `volumes/promtail/` → `../supabase/volumes/promtail/`
  - `volumes/grafana/` → `../supabase/volumes/grafana/`
- **Copied & edited**: `.env`, `docker-compose.yml`, `volumes/prometheus/prometheus.yml`
- **Worktree .env** needs `?sslmode=disable` on `DATABASE_URL` (local Postgres has no SSL)
- **Start**: `cd infra/supabase-wt && docker compose up -d`
- **Migrate**: `cd widia-omni-wt && make migrate-up`
- **Run API**: `cd widia-omni-wt && make dev-api` (loads .env via Makefile)

### Ports

| Service    | Main  | Worktree |
|------------|-------|----------|
| Kong       | 54321 | 54331    |
| Kong TLS   | 54324 | 54334    |
| DB         | 54322 | 54332    |
| Studio     | 54323 | 54333    |
| Redis      | 6379  | 6389     |
| Prometheus | 9091  | 9092     |
| Loki       | 3100  | 3101     |
| Grafana    | 3001  | 3002     |
| API        | 8080  | 8081     |
