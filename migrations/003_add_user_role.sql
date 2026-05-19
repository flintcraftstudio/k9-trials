-- +goose Up
ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'admin'
    CHECK (role IN ('admin', 'judge'));

CREATE INDEX idx_users_role ON users(role);

-- +goose Down
DROP INDEX idx_users_role;
ALTER TABLE users DROP COLUMN role;
