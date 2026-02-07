-- name: GetCurrentEntitlement :one
SELECT * FROM workspace_entitlements
WHERE workspace_id = $1 AND is_current = true;

-- name: CreateEntitlement :one
INSERT INTO workspace_entitlements (workspace_id, tier, limits, source, is_current)
VALUES ($1, $2, $3, $4, true)
RETURNING *;

-- name: ExpireEntitlement :exec
UPDATE workspace_entitlements
SET is_current = false, effective_to = now()
WHERE workspace_id = $1 AND is_current = true;
