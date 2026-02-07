# Local Development Setup

## Prerequisites

- Go 1.24+
- Docker & Docker Compose
- [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
- [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html) (optional, for regenerating queries)

## Infrastructure

Local Supabase runs via Docker Compose at `/Users/bruno/Developer/infra/supabase/`.

### Services

| Service | Container | Port | Image |
|---------|-----------|------|-------|
| API Gateway (Kong) | supabase-kong | **54321** | kong:2.8.1 |
| PostgreSQL | supabase-db | **54322** | supabase/postgres:15.8.1.060 |
| Studio (Dashboard) | supabase-studio | **54323** | supabase/studio:latest |
| Redis | supabase-redis | **6379** | redis:7-alpine |
| Auth (GoTrue) | supabase-auth | internal | supabase/gotrue:v2.170.0 |
| REST (PostgREST) | supabase-rest | internal | postgrest/postgrest:v12.2.3 |
| Metadata | supabase-meta | internal | supabase/postgres-meta:v0.84.2 |

### Start Supabase

```bash
cd /Users/bruno/Developer/infra/supabase
docker compose up -d
```

Wait ~15s for all services to become healthy. Verify:

```bash
docker compose ps
curl -s http://localhost:54321/auth/v1/health \
  -H "apikey: <ANON_KEY>"
```

### Stop / Reset

```bash
# stop (keeps data)
docker compose down

# full reset (destroys volumes)
docker compose down -v
```

After a full reset you must re-run migrations.

## App Setup

### 1. Environment

Copy `.env.example` to `.env` and fill with local Supabase values:

```
SUPABASE_URL=http://localhost:54321
SUPABASE_SERVICE_KEY=<SERVICE_ROLE_KEY from infra/.env>
SUPABASE_JWT_SECRET=<JWT_SECRET from infra/.env>
DATABASE_URL=postgresql://postgres:<POSTGRES_PASSWORD>@localhost:54322/postgres
REDIS_URL=redis://localhost:6379
PORT=8080
ENV=development
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
```

All secrets are in `/Users/bruno/Developer/infra/supabase/.env`.

### 2. Run Migrations

```bash
make migrate-up
```

This applies all 20 migration pairs to the `widia_omni` schema. Tables are **never** created in `public`.

### 3. Start the API

```bash
make dev
```

Server runs on `:8080`. The pgxpool connection sets `search_path TO widia_omni` via `AfterConnect`.

## Database

### Schema

All application tables live in the `widia_omni` schema. The only table in `public` is `schema_migrations` (golang-migrate tracking).

Tables (21):

```
workspaces, workspace_members, plans, subscriptions,
workspace_entitlements, workspace_counters, user_profiles,
user_preferences, life_areas, goals, habits, habit_entries,
tasks, finance_categories, transactions, journal_entries,
area_scores, life_scores, notifications, audit_log,
stripe_events_processed
```

### Automatic Provisioning

On user signup via Supabase Auth, the `handle_new_user()` trigger creates:

- Workspace (named `<username>'s Space`)
- Workspace member (owner role)
- User profile
- User preferences (defaults)
- Free plan subscription
- Free tier entitlement
- Workspace counters (all zeros)

### DB Superuser

The Supabase postgres image uses `supabase_admin` as superuser, **not** `postgres`. For admin operations:

```bash
docker exec supabase-db psql -U supabase_admin -d postgres
```

For app-level queries (respects search_path):

```bash
docker exec supabase-db psql -U postgres -d postgres
```

### Migrations

```bash
# apply all
make migrate-up

# rollback last
make migrate-down

# force version (dirty state recovery)
migrate -path sql/migrations \
  -database "$DATABASE_URL" force <VERSION>
```

### Init Script

The file `infra/supabase/volumes/db/init/99-widia.sql` is mounted into the Postgres container at `/docker-entrypoint-initdb.d/migrations/99-widia.sql`. On first DB init it:

1. Overrides internal role passwords to match `POSTGRES_PASSWORD`
2. Creates the `widia_omni` schema
3. Grants permissions to `postgres`, `authenticated`, `anon`, `service_role`

## Verification Checklist

After setup, verify everything works:

```bash
# 1. Health check
curl http://localhost:54321/auth/v1/health \
  -H "apikey: <ANON_KEY>"
# => {"version":"v2.170.0","name":"GoTrue",...}

# 2. Signup
curl -X POST http://localhost:54321/auth/v1/signup \
  -H "apikey: <ANON_KEY>" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@widia.io","password":"TestPass1234"}'
# => {"access_token":"...","user":{"id":"..."}}

# 3. Check trigger provisioning
docker exec supabase-db psql -U postgres -d postgres \
  -c "SELECT name FROM widia_omni.workspaces;"
# => test's Space

# 4. Studio dashboard
open http://localhost:54323

# 5. Redis
docker exec supabase-redis redis-cli ping
# => PONG

# 6. API server
make dev &
curl http://localhost:8080/health
# => {"status":"ok"}
```

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make dev` | Run API server locally |
| `make build` | Build api + worker binaries to `bin/` |
| `make test` | Run tests with race detector |
| `make migrate-up` | Apply all pending migrations |
| `make migrate-down` | Rollback last migration |
| `make sqlc-gen` | Regenerate sqlc query code |
| `make lint` | Run golangci-lint |
