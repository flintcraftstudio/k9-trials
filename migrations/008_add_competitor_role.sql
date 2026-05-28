-- +goose Up
-- +goose StatementBegin
-- SQLite cannot ALTER a CHECK constraint in place, so we rebuild the table
-- to add 'competitor' to the allowed roles.
CREATE TABLE users_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'competitor'
        CHECK (role IN ('admin', 'judge', 'competitor')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO users_new (id, email, password_hash, role, created_at, updated_at)
    SELECT id, email, password_hash, role, created_at, updated_at FROM users;
DROP INDEX IF EXISTS idx_users_role;
DROP TABLE users;
ALTER TABLE users_new RENAME TO users;
CREATE INDEX idx_users_role ON users(role);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE users_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'admin'
        CHECK (role IN ('admin', 'judge')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO users_new (id, email, password_hash, role, created_at, updated_at)
    SELECT id, email, password_hash,
           CASE WHEN role = 'competitor' THEN 'admin' ELSE role END,
           created_at, updated_at FROM users;
DROP INDEX IF EXISTS idx_users_role;
DROP TABLE users;
ALTER TABLE users_new RENAME TO users;
CREATE INDEX idx_users_role ON users(role);
-- +goose StatementEnd
