-- +goose Up
-- A competitor's request to be emailed when a not-yet-open event opens
-- registration (Q4 / R1c). The publish transition (draft -> published) is the
-- trigger; notified_at records when the notification was sent so a later
-- re-publish does not re-notify. Email delivery is still unwired — recipients
-- are logged like the D6 notify-judges stub.
CREATE TABLE event_subscriptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    competitor_id INTEGER NOT NULL REFERENCES competitors(id),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    notified_at DATETIME,
    UNIQUE (event_id, competitor_id)
);

CREATE INDEX idx_event_subscriptions_event_id ON event_subscriptions(event_id);

-- +goose Down
DROP TABLE event_subscriptions;
