-- name: CreateAccount :one
INSERT INTO accounts (name, description, workspace_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetAccount :one
SELECT * FROM accounts
WHERE id = $1;

-- name: ListAccountsByWorkspace :many
SELECT * FROM accounts
WHERE workspace_id = $1
ORDER BY id;

-- name: UpdateAccount :one
UPDATE accounts
SET name = $2, description = $3
WHERE id = $1
RETURNING *;

-- name: DeleteAccount :exec
DELETE FROM accounts
WHERE id = $1;
