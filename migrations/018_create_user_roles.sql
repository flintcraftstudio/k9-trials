-- +goose Up
-- user_roles holds additive account capabilities. 'competitor' is the universal
-- implicit baseline for every authenticated account and is NEVER stored here —
-- only explicit 'judge' and 'admin' grants are rows. A user may hold any
-- combination (a judge who also competes simply has a 'judge' row). The
-- composite PRIMARY KEY already indexes user_id (leftmost column), so lookups
-- by user_id need no separate index.
CREATE TABLE user_roles (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    capability TEXT NOT NULL CHECK (capability IN ('judge', 'admin')),
    PRIMARY KEY (user_id, capability)
);

-- Backfill capabilities from the existing single-string users.role. Competitor
-- is implicit, so no row is inserted for role='competitor'.
INSERT INTO user_roles (user_id, capability)
    SELECT id, 'judge' FROM users WHERE role = 'judge';
INSERT INTO user_roles (user_id, capability)
    SELECT id, 'admin' FROM users WHERE role = 'admin';

-- NOTE: users.role is intentionally NOT dropped here. session/store still read
-- it; dropping now would break the build. The drop is deferred to step 2 of the
-- multi-role-accounts migration.

-- +goose Down
DROP TABLE user_roles;
