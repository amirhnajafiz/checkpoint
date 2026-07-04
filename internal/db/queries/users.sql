-- name: UpsertUser :one
INSERT INTO users (email)
VALUES ($1)
ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
RETURNING *;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;
