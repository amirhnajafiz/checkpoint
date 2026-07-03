-- name: CreateWorkspace :one
INSERT INTO workspaces (user_email)
VALUES ($1)
RETURNING *;

-- name: GetWorkspace :one
SELECT * FROM workspaces
WHERE id = $1;

-- name: ListWorkspaces :many
SELECT * FROM workspaces
ORDER BY id;

-- name: ListWorkspacesByUser :many
SELECT * FROM workspaces
WHERE user_email = $1
ORDER BY id;

-- name: UpdateWorkspace :one
UPDATE workspaces
SET user_email = $2
WHERE id = $1
RETURNING *;

-- name: DeleteWorkspace :exec
DELETE FROM workspaces
WHERE id = $1;
