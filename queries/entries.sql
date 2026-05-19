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
