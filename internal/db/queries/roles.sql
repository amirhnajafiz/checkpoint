-- name: CreateRole :one
INSERT INTO roles (name, description, workspace_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetRole :one
SELECT * FROM roles
WHERE id = $1;

-- name: ListRolesByWorkspace :many
SELECT * FROM roles
WHERE workspace_id = $1
ORDER BY id;

-- name: UpdateRole :one
UPDATE roles
SET name = $2, description = $3
WHERE id = $1
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE id = $1;
