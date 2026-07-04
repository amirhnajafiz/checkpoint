-- name: CreateServiceAccountMeta :one
INSERT INTO service_account_meta (account_id)
VALUES ($1)
RETURNING *;

-- name: GetServiceAccountMeta :one
SELECT * FROM service_account_meta
WHERE account_id = $1;

-- name: UpdateServiceAccountMeta :one
UPDATE service_account_meta
SET last_used = $1, usage = usage + $2
WHERE account_id = $1
RETURNING *;

-- name: DeleteServiceAccountMeta :exec
DELETE FROM service_account_meta
WHERE account_id = $1;
