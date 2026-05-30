-- name: CreateEntry :one
INSERT INTO entries (trial_id, judge_id, entry_number, handler_name, dog_name, dog_breed, status)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetEntryByID :one
SELECT * FROM entries WHERE id = ?;

-- name: ListEntriesByTrial :many
SELECT * FROM entries
WHERE trial_id = ?
ORDER BY entry_number;

-- name: ListFinalizedEntriesByTrial :many
SELECT * FROM entries
WHERE trial_id = ? AND status = 'finalized'
ORDER BY entry_number;

-- name: ListEntriesByJudge :many
SELECT * FROM entries
WHERE judge_id = ?
ORDER BY trial_id, entry_number;

-- name: AssignEntryJudge :one
UPDATE entries
SET judge_id = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: UpdateEntryStatus :one
UPDATE entries
SET status = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: UpdateEntry :one
UPDATE entries
SET handler_name = ?, dog_name = ?, dog_breed = ?, entry_number = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteEntry :exec
DELETE FROM entries WHERE id = ?;

-- name: ListEntriesByHandler :many
-- Every entry handled by a competitor across all events and statuses,
-- joined to trial + event. Newest trial date first. Backs the account
-- entries list (A5) and the dashboard up-next / recent clusters (A1).
-- template_version is included so finalized rows can be re-evaluated.
SELECT
    e.id, e.entry_number, e.dog_name, e.dog_breed, e.handler_name,
    e.status, e.dog_id,
    t.id AS trial_id, t.discipline, t.level, t.trial_date, t.template_version,
    ev.name AS event_name, ev.slug AS event_slug
FROM entries e
JOIN trials t ON t.id = e.trial_id
JOIN events ev ON ev.id = t.event_id
WHERE e.handler_id = ?
ORDER BY t.trial_date DESC, e.id DESC;
