-- name: SubscribeToEvent :exec
-- Records a competitor's notify-me request for an event. Idempotent: a repeat
-- subscribe is a no-op via the (event_id, competitor_id) UNIQUE.
INSERT INTO event_subscriptions (event_id, competitor_id)
VALUES (?, ?)
ON CONFLICT (event_id, competitor_id) DO NOTHING;

-- name: HasEventSubscription :one
-- Whether a competitor is already subscribed to an event, for the R1c
-- notify-me state.
SELECT COUNT(*) FROM event_subscriptions
WHERE event_id = ? AND competitor_id = ?;

-- name: ListEventSubscribers :many
-- Subscribers awaiting notification for an event, with the email to send to.
-- Drives the publish-transition hook (recipients are logged, not yet mailed).
SELECT es.id, es.competitor_id, c.display_name, u.email
FROM event_subscriptions es
JOIN competitors c ON c.id = es.competitor_id
JOIN users u ON u.id = c.user_id
WHERE es.event_id = ? AND es.notified_at IS NULL
ORDER BY es.id;

-- name: MarkEventSubscribersNotified :exec
-- Stamps notified_at on an event's un-notified subscribers so a later
-- re-publish does not notify them again.
UPDATE event_subscriptions
SET notified_at = CURRENT_TIMESTAMP
WHERE event_id = ? AND notified_at IS NULL;
