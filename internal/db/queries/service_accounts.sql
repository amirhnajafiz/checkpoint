-- name: CreateServiceAccount :one
INSERT INTO service_accounts (name, description, active, user_email, ttl_seconds)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetServiceAccount :one
SELECT * FROM service_accounts
WHERE id = $1;

-- name: ListUserServiceAccounts :many
SELECT *
FROM service_accounts as sa JOIN service_account_meta as sam ON sa.id = sam.account_id
WHERE user_email = $1
ORDER BY id;

-- name: UpdateServiceAccount :one
UPDATE service_accounts
SET name = $2, description = $3, active = $4, ttl_seconds = $5
WHERE id = $1
RETURNING *;

-- name: DeleteServiceAccount :exec
DELETE FROM service_accounts
WHERE id = $1;
