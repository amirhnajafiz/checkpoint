-- name: BindAccountRole :one
INSERT INTO account_roles (role_id, account_id)
VALUES ($1, $2)
RETURNING *;

-- name: UnbindAccountRole :exec
DELETE FROM account_roles
WHERE role_id = $1 AND account_id = $2;

-- name: ListAccountRoles :many
SELECT * FROM account_roles
ORDER BY role_id, account_id;

-- name: ListRolesByAccount :many
SELECT r.* FROM roles r
JOIN account_roles ar ON ar.role_id = r.id
WHERE ar.account_id = $1
ORDER BY r.id;

-- name: ListAccountsByRole :many
SELECT a.* FROM accounts a
JOIN account_roles ar ON ar.account_id = a.id
WHERE ar.role_id = $1
ORDER BY a.id;
