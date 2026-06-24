-- name: CreateEvent :one
INSERT INTO events (slug, name, location, start_date, end_date, status, created_by, published_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetEventByID :one
SELECT * FROM events WHERE id = ?;

-- name: GetEventBySlug :one
SELECT * FROM events WHERE slug = ?;

-- name: ListEvents :many
SELECT * FROM events
ORDER BY start_date DESC, id DESC;

-- name: ListPublishedEvents :many
SELECT * FROM events
WHERE status IN ('published', 'closed')
ORDER BY start_date DESC, id DESC;

-- name: UpdateEvent :one
UPDATE events
SET name = ?, location = ?, start_date = ?, end_date = ?, status = ?,
    published_at = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: SetEventStatus :one
-- Lifecycle transition that touches only the status (and published_at when
-- entering 'published'). Backs the D3 Archive / Restore actions without
-- disturbing the rest of the metadata form.
UPDATE events
SET status = ?, published_at = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteEvent :exec
DELETE FROM events WHERE id = ?;

-- name: CountEntriesByEvent :one
-- Total entries (any status) across all trials of an event. Drives the D3
-- at-a-glance "N entries across all trials" line.
SELECT COUNT(*) FROM entries e
JOIN trials t ON t.id = e.trial_id
WHERE t.event_id = ?;

-- name: CountTrialsWithJudgeByEvent :one
-- How many of an event's trials have a judge assigned (resolved through any
-- entry carrying a judge_id). Drives the D3 at-a-glance judge-coverage line.
SELECT COUNT(DISTINCT t.id) FROM trials t
JOIN entries e ON e.trial_id = t.id
WHERE t.event_id = ? AND e.judge_id IS NOT NULL;

-- name: CountEventsBySlug :one
-- Number of events using a slug, excluding one event id. Backs the live
-- slug-availability check on the event form (pass 0 on create to exclude
-- nobody).
SELECT COUNT(*) FROM events WHERE slug = ? AND id != ?;
