-- +goose Up
CREATE TABLE trials (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    discipline TEXT NOT NULL
        CHECK (discipline IN ('OB', 'PR', 'TR', 'DT')),
    level INTEGER NOT NULL
        CHECK (level IN (1, 2, 3)),
    trial_date DATE NOT NULL,
    template_version TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'in_progress', 'complete')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (event_id, discipline, level, trial_date)
);

CREATE INDEX idx_trials_event_id ON trials(event_id);
CREATE INDEX idx_trials_status ON trials(status);
CREATE INDEX idx_trials_date ON trials(trial_date);

-- +goose Down
DROP TABLE trials;
