-- name: GetWorkspaceByID :one
SELECT w.*, wm.role
FROM workspaces w
JOIN workspace_members wm ON wm.workspace_id = w.id
WHERE w.id = $1 AND wm.user_id = $2;

-- name: GetWorkspacesByUserID :many
SELECT w.*, wm.role
FROM workspaces w
JOIN workspace_members wm ON wm.workspace_id = w.id
WHERE wm.user_id = $1
ORDER BY w.created_at;

-- name: UpdateWorkspace :one
UPDATE workspaces
SET name = $2, slug = $3
WHERE id = $1
RETURNING *;
