-- name: CreateTrial :one
INSERT INTO trials (event_id, discipline, level, trial_date, template_version, status)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetTrialByID :one
SELECT * FROM trials WHERE id = ?;

-- name: ListTrialsByEvent :many
SELECT * FROM trials
WHERE event_id = ?
ORDER BY trial_date, discipline, level;

-- name: ListTrialsByJudge :many
SELECT DISTINCT t.*
FROM trials t
JOIN entries e ON e.trial_id = t.id
WHERE e.judge_id = ?
ORDER BY t.trial_date DESC, t.id DESC;

-- name: UpdateTrialStatus :one
UPDATE trials
SET status = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteTrial :exec
DELETE FROM trials WHERE id = ?;
