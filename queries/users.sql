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

-- name: ListAssignableJudges :many
-- Users who can be assigned to judge a trial: judges and admins (an admin
-- can judge). Ordered by email for the assignment dropdown (D6).
SELECT id, email FROM users
WHERE role IN ('judge', 'admin')
ORDER BY email;

-- name: ListUsersWithCompetitor :many
-- Every user with their competitor identity (when they have one), for the
-- users and roles admin (D8). handle and display_name are null for users
-- without a competitor row (judges, admins).
SELECT
    u.id, u.email, u.role, u.created_at,
    c.handle, c.display_name
FROM users u
LEFT JOIN competitors c ON c.user_id = u.id
ORDER BY u.created_at DESC, u.id DESC;

-- name: UpdateUserRole :exec
-- Changes a user role (admin / judge / competitor).
UPDATE users SET role = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;
