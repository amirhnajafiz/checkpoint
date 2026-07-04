-- name: CreateServiceAccountMeta :one
INSERT INTO service_account_meta (account_id)
VALUES ($1)
RETURNING *;

-- name: GetServiceAccountMeta :one
SELECT * FROM service_account_meta
WHERE account_id = $1;

-- name: UpdateServiceAccountMeta :one
UPDATE service_account_meta
SET last_used = $2, usage = usage + $3
WHERE account_id = $1
RETURNING *;

-- name: AddServiceAccountUsage :exec
UPDATE service_account_meta
SET usage = usage + $2, last_used = NOW()
WHERE account_id = $1;

-- name: DeleteServiceAccountMeta :exec
DELETE FROM service_account_meta
WHERE account_id = $1;
