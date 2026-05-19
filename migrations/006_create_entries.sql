-- +goose Up
CREATE TABLE entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    trial_id INTEGER NOT NULL REFERENCES trials(id) ON DELETE CASCADE,
    judge_id INTEGER REFERENCES users(id),
    entry_number INTEGER NOT NULL,
    handler_name TEXT NOT NULL,
    dog_name TEXT NOT NULL,
    dog_breed TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'registered'
        CHECK (status IN ('registered', 'scoring', 'finalized')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (trial_id, entry_number)
);

CREATE INDEX idx_entries_trial_id ON entries(trial_id);
CREATE INDEX idx_entries_judge_id ON entries(judge_id);
CREATE INDEX idx_entries_status ON entries(status);

-- +goose Down
DROP TABLE entries;
