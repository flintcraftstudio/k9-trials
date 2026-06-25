-- Append-only score inputs. The evaluation engine reads these tables
-- and projects them into ScoresheetInputs (see internal/scoring).

-- name: RecordCriterionScore :one
INSERT INTO criterion_scores (entry_id, exercise_code, criterion_code, points, judged_by)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: ListCriterionScoresByEntry :many
-- Returns all rows ordered by time. The scoring engine in
-- internal/scoring projects latest-write-wins per
-- (exercise_code, criterion_code) when building ScoresheetInputs.
SELECT * FROM criterion_scores
WHERE entry_id = ?
ORDER BY created_at;

-- name: RecordPenaltyOccurrence :one
INSERT INTO penalty_occurrences (entry_id, exercise_code, event_code, judged_by)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: ListPenaltyOccurrencesByEntry :many
SELECT * FROM penalty_occurrences
WHERE entry_id = ?
ORDER BY created_at;

-- name: RecordAutoTriggerFiring :one
INSERT INTO auto_trigger_firings (entry_id, exercise_code, trigger_code, judged_by)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: ListAutoTriggerFiringsByEntry :many
SELECT * FROM auto_trigger_firings
WHERE entry_id = ?
ORDER BY created_at;

-- name: DeleteAutoTriggerFiring :exec
-- Removes all firings of one trigger on one exercise. Auto-NQ triggers are
-- boolean in effect (fired or not), so the judge toggles them off by deleting
-- the firing rows rather than appending a counter-record.
DELETE FROM auto_trigger_firings
WHERE entry_id = ? AND exercise_code = ? AND trigger_code = ?;

-- name: RecordModifierApplication :one
INSERT INTO modifier_applications (entry_id, modifier_code, judged_by)
VALUES (?, ?, ?)
RETURNING *;

-- name: ListModifierApplicationsByEntry :many
SELECT * FROM modifier_applications
WHERE entry_id = ?
ORDER BY created_at;
