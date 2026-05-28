-- +goose Up
-- A registration is a request to enter a dog in a trial. Once accepted,
-- entry_id points at the entries row that holds the assigned entry number
-- and scoring state. submitted_by is separate from competitor_id so a
-- club secretary (a different user) can register on behalf of a handler.
CREATE TABLE registrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    trial_id INTEGER NOT NULL REFERENCES trials(id) ON DELETE CASCADE,
    competitor_id INTEGER NOT NULL REFERENCES competitors(id),
    dog_id INTEGER NOT NULL REFERENCES dogs(id),
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'accepted', 'waitlisted', 'withdrawn', 'rejected')),
    notes TEXT NOT NULL DEFAULT '',
    submitted_by INTEGER NOT NULL REFERENCES users(id),
    submitted_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    reviewed_by INTEGER REFERENCES users(id),
    reviewed_at DATETIME,
    entry_id INTEGER REFERENCES entries(id),
    UNIQUE (trial_id, dog_id)
);

CREATE INDEX idx_registrations_trial_id ON registrations(trial_id);
CREATE INDEX idx_registrations_competitor_id ON registrations(competitor_id);
CREATE INDEX idx_registrations_status ON registrations(status);

-- +goose Down
DROP TABLE registrations;
