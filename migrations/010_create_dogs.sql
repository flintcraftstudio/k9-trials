-- +goose Up
-- owner_id points at the primary owner. Co-ownership can be modeled later
-- with a dog_owners join table without changing this column.
CREATE TABLE dogs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_id INTEGER NOT NULL REFERENCES competitors(id),
    call_name TEXT NOT NULL,
    registered_name TEXT NOT NULL DEFAULT '',
    breed TEXT NOT NULL DEFAULT '',
    date_of_birth DATE,
    registration_number TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_dogs_owner_id ON dogs(owner_id);

-- +goose Down
DROP TABLE dogs;
