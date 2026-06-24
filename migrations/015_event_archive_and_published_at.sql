-- +goose NO TRANSACTION
-- +goose Up
-- Adds the 'archived' status to events and a published_at timestamp for the
-- D3 audit block. SQLite cannot ALTER an existing CHECK constraint, so the
-- table is rebuilt per the documented procedure. foreign_keys is toggled off
-- around the swap so trials (which reference events.id ON DELETE CASCADE) are
-- not cascade-deleted when the old table is dropped.
PRAGMA foreign_keys=OFF;

CREATE TABLE events_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    location TEXT NOT NULL DEFAULT '',
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'published', 'closed', 'archived')),
    created_by INTEGER NOT NULL REFERENCES users(id),
    published_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (end_date >= start_date)
);

INSERT INTO events_new (id, slug, name, location, start_date, end_date, status, created_by, created_at, updated_at)
SELECT id, slug, name, location, start_date, end_date, status, created_by, created_at, updated_at FROM events;

DROP TABLE events;

ALTER TABLE events_new RENAME TO events;

CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_start_date ON events(start_date);

PRAGMA foreign_keys=ON;

-- +goose Down
-- Reverts to the original schema (no archived status, no published_at). Fails
-- if any event is currently archived, which is the intended guard.
PRAGMA foreign_keys=OFF;

CREATE TABLE events_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    location TEXT NOT NULL DEFAULT '',
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'published', 'closed')),
    created_by INTEGER NOT NULL REFERENCES users(id),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (end_date >= start_date)
);

INSERT INTO events_old (id, slug, name, location, start_date, end_date, status, created_by, created_at, updated_at)
SELECT id, slug, name, location, start_date, end_date, status, created_by, created_at, updated_at FROM events;

DROP TABLE events;

ALTER TABLE events_old RENAME TO events;

CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_start_date ON events(start_date);

PRAGMA foreign_keys=ON;
