# Testing the API

## Start the Server

```bash
set -a && source .env && set +a && go run ./cmd/api
```

Server starts on `:8080`. Logs are JSON (zerolog).

## Endpoints

### Health & Readiness (no auth)

```bash
curl http://localhost:8080/health
# {"status":"ok"}

curl http://localhost:8080/ready
# {"status":"ready"}  (pings DB + Redis)
```

### Auth (no auth required)

**Register** — creates user in Supabase Auth, trigger auto-provisions workspace + profile + subscription + entitlement + counters:

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"you@example.com","password":"YourPass1234"}'
```

Response:
```json
{
  "access_token": "eyJ...",
  "refresh_token": "...",
  "user": {
    "id": "uuid",
    "email": "you@example.com",
    "role": "authenticated"
  }
}
```

**Login:**

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"you@example.com","password":"YourPass1234"}'
```

**Refresh token:**

```bash
curl -X POST http://localhost:8080/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<REFRESH_TOKEN>"}'
```

**Logout:**

```bash
curl -X POST http://localhost:8080/auth/logout \
  -H "Authorization: Bearer $TOKEN"
```

### Protected Endpoints (Bearer token required)

Save the `access_token` from register/login:

```bash
TOKEN="eyJ..."
```

**Get profile:**

```bash
curl http://localhost:8080/api/v1/me \
  -H "Authorization: Bearer $TOKEN"
```

```json
{
  "id": "uuid",
  "display_name": "you",
  "email": "you@example.com",
  "timezone": "America/Sao_Paulo",
  "locale": "pt-BR",
  "default_workspace_id": "uuid",
  "onboarding_completed": false,
  "created_at": "...",
  "updated_at": "..."
}
```

**Update profile:**

```bash
curl -X PUT http://localhost:8080/api/v1/me \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"display_name":"Demo User","timezone":"America/Sao_Paulo"}'
```

**Get preferences:**

```bash
curl http://localhost:8080/api/v1/me/preferences \
  -H "Authorization: Bearer $TOKEN"
```

**Update preferences:**

```bash
curl -X PUT http://localhost:8080/api/v1/me/preferences \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"theme":"dark","currency":"BRL","week_starts_on":1}'
```

**Delete account:**

```bash
curl -X DELETE http://localhost:8080/api/v1/me \
  -H "Authorization: Bearer $TOKEN"
```

**Get workspace:**

```bash
curl http://localhost:8080/api/v1/workspace \
  -H "Authorization: Bearer $TOKEN"
```

```json
{
  "id": "uuid",
  "name": "you's Space",
  "slug": "ws-abc123",
  "owner_id": "uuid",
  "created_at": "...",
  "updated_at": "..."
}
```

**Update workspace:**

```bash
curl -X PUT http://localhost:8080/api/v1/workspace \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"My Workspace","slug":"my-ws"}'
```

**Get workspace usage:**

```bash
curl http://localhost:8080/api/v1/workspace/usage \
  -H "Authorization: Bearer $TOKEN"
```

### Metrics (no auth)

```bash
curl http://localhost:8080/metrics
```

Returns Prometheus format with:
- `http_requests_total` — counter by method + status
- `http_request_duration_seconds` — histogram by method + path

## Error Responses

All errors follow the same shape:

```json
{"error": "description of the error"}
```

| Status | Meaning |
|--------|---------|
| 400 | Bad request / invalid JSON |
| 401 | Missing or invalid JWT |
| 403 | Not a workspace member / insufficient permissions |
| 404 | Resource not found |
| 429 | Rate limit exceeded (check `X-RateLimit-*` headers) |
| 500 | Internal server error |

## Middleware Pipeline

Requests to `/api/v1/*` pass through:

```
RequestID → Logger → CORS → Metrics → Auth → Tenant → RateLimit → Handler
```

- **Auth** validates JWT signature using `SUPABASE_JWT_SECRET`, extracts `user_id`
- **Tenant** resolves `workspace_id` + role + entitlements from DB (cached in Redis 5min)
- **RateLimit** enforces token bucket per workspace (Redis-backed)

## Quick Full Flow

```bash
# 1. Register
RESP=$(curl -s -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@widia.io","password":"TestPass1234"}')
TOKEN=$(echo $RESP | python3 -c "import sys,json; print(json.load(sys.stdin)['access_token'])")

# 2. Profile
curl -s http://localhost:8080/api/v1/me -H "Authorization: Bearer $TOKEN"

# 3. Update
curl -s -X PUT http://localhost:8080/api/v1/me \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"display_name":"Tester"}'

# 4. Workspace
curl -s http://localhost:8080/api/v1/workspace -H "Authorization: Bearer $TOKEN"

# 5. Verify in DB
docker exec supabase-db psql -U postgres -d postgres \
  -c "SELECT display_name, email FROM widia_omni.user_profiles;"
```

## Studio

Browse tables visually at http://localhost:54323.
