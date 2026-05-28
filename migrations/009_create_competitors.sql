-- +goose Up
-- competitors is the public-facing identity for handlers and dog owners.
-- user_id is nullable so admins can pre-create competitor rows for handlers
-- who don't (yet) have a login account, and so that deleting a user account
-- preserves competitive history under the same competitor row.
CREATE TABLE competitors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER UNIQUE REFERENCES users(id) ON DELETE SET NULL,
    handle TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    bio TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_competitors_user_id ON competitors(user_id);
CREATE INDEX idx_competitors_handle ON competitors(handle);

-- +goose Down
DROP TABLE competitors;
