-- name: SetServiceAccountKV :one
INSERT INTO service_account_kv (account_id, xkey, xvalue)
VALUES ($1, $2, $3)
ON CONFLICT (account_id, xkey) DO UPDATE SET xvalue = EXCLUDED.xvalue
RETURNING *;

-- name: ListServiceAccountKV :many
SELECT * FROM service_account_kv
WHERE account_id = $1
ORDER BY xkey;

-- name: DeleteServiceAccountKV :exec
DELETE FROM service_account_kv
WHERE id = $1;
