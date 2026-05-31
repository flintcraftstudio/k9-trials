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

-- name: CountEntriesByTrial :one
-- Number of entries recorded in a trial. Drives the entry-count hint on
-- the admin trials list (D4).
SELECT COUNT(*) FROM entries WHERE trial_id = ?;

-- name: CountTrialsByEvent :one
-- Number of trials in an event. Drives the trial-count column on the admin
-- events list (D2) and the event editor at-a-glance panel (D3).
SELECT COUNT(*) FROM trials WHERE event_id = ?;

-- name: TrialJudgeEmail :one
-- Email of the judge assigned to a trial, resolved through any entry that
-- carries a judge_id. Returns sql.ErrNoRows when no judge is assigned yet,
-- which the trials list (D4) renders as no-judge.
SELECT u.email
FROM entries e
JOIN users u ON u.id = e.judge_id
WHERE e.trial_id = ? AND e.judge_id IS NOT NULL
LIMIT 1;
