-- name: CreateEvent :one
INSERT INTO events (slug, name, location, start_date, end_date, status, created_by)
VALUES (?, ?, ?, ?, ?, ?, ?)
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
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteEvent :exec
DELETE FROM events WHERE id = ?;

-- name: CountEventsBySlug :one
-- Number of events using a slug, excluding one event id. Backs the live
-- slug-availability check on the event form (pass 0 on create to exclude
-- nobody).
SELECT COUNT(*) FROM events WHERE slug = ? AND id != ?;
