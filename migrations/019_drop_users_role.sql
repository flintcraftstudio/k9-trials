-- +goose Up
-- Step 5 of the multi-role-accounts migration: drop the legacy users.role
-- column. Authorization is now driven entirely by the additive capability model
-- (user_roles); nothing reads users.role anymore.
--
-- A plain `ALTER TABLE users DROP COLUMN role` fails while idx_users_role
-- references the column ("error in index idx_users_role after drop column").
-- Dropping that index first lets the direct DROP COLUMN succeed. This is
-- preferred over a full table rebuild: it preserves the users table identity,
-- its PK and email UNIQUE constraint, and — critically — the user_roles foreign
-- key (REFERENCES users(id) ON DELETE CASCADE), which a drop/recreate of the
-- users table would put at risk under PRAGMA foreign_keys=ON. The inline
-- CHECK/DEFAULT on the column go away with it.
DROP INDEX IF EXISTS idx_users_role;
ALTER TABLE users DROP COLUMN role;

-- +goose Down
-- Restore the column so the migration is reversible. Historical role values are
-- NOT recoverable — capabilities (user_roles) are the source of truth now, so
-- every restored row gets the 'competitor' default. The original column was
-- NOT NULL DEFAULT 'competitor' CHECK (role IN ('admin','judge','competitor'));
-- SQLite's ADD COLUMN cannot re-attach a CHECK constraint, so only the NOT NULL
-- DEFAULT is restored. The idx_users_role index is recreated to match the
-- pre-drop schema.
ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'competitor';
CREATE INDEX idx_users_role ON users(role);
