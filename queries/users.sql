-- name: GetUserByEmail :one
SELECT id, email, password_hash, created_at, updated_at
FROM users
WHERE email = ?;

-- name: GetUserByID :one
SELECT id, email, password_hash, created_at, updated_at
FROM users
WHERE id = ?;

-- name: CreateUser :one
-- Creates a new account at the competitor baseline. Capabilities ('judge' /
-- 'admin') are additive grants stored in user_roles - never a column here - so
-- a fresh account holds no explicit capability until one is granted.
INSERT INTO users (email, password_hash)
VALUES (?, ?)
RETURNING id, email, password_hash, created_at, updated_at;

-- name: ListJudgeEligibleUsers :many
-- Accounts assignable as a trial judge, derived from the additive capability
-- model (user_roles). A user is eligible if they hold the 'judge' capability OR
-- the 'admin' capability (admin is a superset and can judge). DISTINCT collapses
-- users holding both grants. Ordered by email for the assignment dropdown (D6).
SELECT DISTINCT u.id, u.email
FROM users u
JOIN user_roles ur ON ur.user_id = u.id
WHERE ur.capability IN ('judge', 'admin')
ORDER BY u.email;

-- name: ListUsersWithCaps :many
-- Every user with their competitor identity (when any) and a comma-separated,
-- alphabetically sorted list of their explicit account capabilities
-- ('admin'/'judge'), for the users and roles admin (D8). The 'competitor'
-- baseline is implicit and never stored, so caps is empty for competitor-only
-- accounts. Derive the display label from caps.
SELECT
    u.id, u.email, u.created_at,
    c.handle, c.display_name,
    CAST(COALESCE((SELECT group_concat(ur.capability, ',')
       FROM (SELECT capability FROM user_roles
              WHERE user_id = u.id ORDER BY capability) ur), '') AS TEXT) AS caps
FROM users u
LEFT JOIN competitors c ON c.user_id = u.id
ORDER BY u.created_at DESC, u.id DESC;

-- name: UserCapabilities :many
-- The explicit account capabilities a user holds ('judge' / 'admin'). The
-- 'competitor' baseline is implicit and never stored, so an empty result means
-- the user is a competitor only. Ordered for deterministic output.
SELECT capability FROM user_roles
WHERE user_id = ?
ORDER BY capability;

-- name: GrantCapability :exec
-- Grants an account capability to a user. Idempotent: re-granting an existing
-- capability is a no-op (ON CONFLICT DO NOTHING).
INSERT INTO user_roles (user_id, capability)
VALUES (?, ?)
ON CONFLICT (user_id, capability) DO NOTHING;

-- name: RevokeCapability :exec
-- Revokes an account capability from a user. Idempotent: revoking an absent
-- capability deletes zero rows and is not an error.
DELETE FROM user_roles
WHERE user_id = ? AND capability = ?;
