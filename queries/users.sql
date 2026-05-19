-- name: GetUserByEmail :one
SELECT id, email, password_hash, role, created_at, updated_at
FROM users
WHERE email = ?;

-- name: GetUserByID :one
SELECT id, email, password_hash, role, created_at, updated_at
FROM users
WHERE id = ?;

-- name: CreateUser :one
INSERT INTO users (email, password_hash, role)
VALUES (?, ?, ?)
RETURNING id, email, password_hash, role, created_at, updated_at;

-- name: ListJudges :many
SELECT id, email, created_at, updated_at
FROM users
WHERE role = 'judge'
ORDER BY email;
