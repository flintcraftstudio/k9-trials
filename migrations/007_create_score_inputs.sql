-- +goose Up
-- Append-only input tables. Each row records a single judge action.
-- The evaluation engine reads the latest write per logical key
-- (criterion_scores: per (entry_id, exercise_code, criterion_code);
-- modifier_applications: per (entry_id, modifier_code)).
-- penalty_occurrences and auto_trigger_firings are counted, not
-- latest-write-wins.

CREATE TABLE criterion_scores (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entry_id INTEGER NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    exercise_code TEXT NOT NULL,
    criterion_code TEXT NOT NULL,
    points INTEGER NOT NULL,
    judged_by INTEGER NOT NULL REFERENCES users(id),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_criterion_scores_lookup
    ON criterion_scores(entry_id, exercise_code, criterion_code, created_at);

CREATE TABLE penalty_occurrences (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entry_id INTEGER NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    exercise_code TEXT NOT NULL,
    event_code TEXT NOT NULL,
    judged_by INTEGER NOT NULL REFERENCES users(id),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_penalty_occurrences_entry
    ON penalty_occurrences(entry_id, exercise_code);

CREATE TABLE auto_trigger_firings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entry_id INTEGER NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    exercise_code TEXT NOT NULL,
    trigger_code TEXT NOT NULL,
    judged_by INTEGER NOT NULL REFERENCES users(id),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_auto_trigger_firings_entry
    ON auto_trigger_firings(entry_id, exercise_code);

CREATE TABLE modifier_applications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entry_id INTEGER NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    modifier_code TEXT NOT NULL,
    judged_by INTEGER NOT NULL REFERENCES users(id),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_modifier_applications_entry
    ON modifier_applications(entry_id, modifier_code, created_at);

-- +goose Down
DROP TABLE modifier_applications;
DROP TABLE auto_trigger_firings;
DROP TABLE penalty_occurrences;
DROP TABLE criterion_scores;
