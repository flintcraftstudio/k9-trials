-- +goose Up
-- A challenge is a competitor's dispute of a finalized scoresheet. Resolution
-- (when status becomes 'resolved') happens through normal scoring inputs —
-- the append-only score_inputs trail naturally records the change. This
-- table stores only the dispute itself and its outcome notes.
CREATE TABLE challenges (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entry_id INTEGER NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    filed_by INTEGER NOT NULL REFERENCES competitors(id),
    reason TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'open'
        CHECK (status IN ('open', 'under_review', 'resolved', 'dismissed')),
    resolution_notes TEXT NOT NULL DEFAULT '',
    resolved_by INTEGER REFERENCES users(id),
    resolved_at DATETIME,
    filed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_challenges_entry_id ON challenges(entry_id);
CREATE INDEX idx_challenges_filed_by ON challenges(filed_by);
CREATE INDEX idx_challenges_status ON challenges(status);

-- +goose Down
DROP TABLE challenges;
