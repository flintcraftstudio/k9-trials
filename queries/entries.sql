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
-- template_version is included so finalized rows can be re-evaluated. The
-- linked registration's status and withdraw_requested_at carry the
-- withdrawal state (null for entries created outside the registration flow).
SELECT
    e.id, e.entry_number, e.dog_name, e.dog_breed, e.handler_name,
    e.status, e.dog_id,
    t.id AS trial_id, t.discipline, t.level, t.trial_date, t.template_version,
    ev.name AS event_name, ev.slug AS event_slug,
    rg.status AS reg_status, rg.withdraw_requested_at AS withdraw_requested_at
FROM entries e
JOIN trials t ON t.id = e.trial_id
JOIN events ev ON ev.id = t.event_id
LEFT JOIN registrations rg ON rg.entry_id = e.id
WHERE e.handler_id = ?
ORDER BY t.trial_date DESC, e.id DESC;

-- name: MaxEntryNumberByTrial :one
-- Highest entry number used in a trial, or 0 when it has no entries. The
-- accept flow assigns the next number.
SELECT CAST(COALESCE(MAX(entry_number), 0) AS INTEGER) AS max_number FROM entries WHERE trial_id = ?;

-- name: CreateEntryForRegistration :one
-- Creates the entry an accepted registration produces, carrying the
-- competitor and dog foreign keys plus the print-program name snapshot.
INSERT INTO entries (
    trial_id, judge_id, entry_number, handler_name, dog_name, dog_breed,
    status, dog_id, handler_id
)
VALUES (?, ?, ?, ?, ?, ?, 'registered', ?, ?)
RETURNING *;

-- name: TrialJudgeID :one
-- The judge id assigned to a trial, resolved through any of its entries.
-- Returns sql.ErrNoRows when no judge is assigned yet.
SELECT judge_id FROM entries
WHERE trial_id = ? AND judge_id IS NOT NULL
LIMIT 1;

-- name: AssignTrialJudge :exec
-- Sets the judge on every entry in a trial. Judge assignment is per-entry
-- in this schema, so assigning a trial judge bulk-updates its entries.
UPDATE entries SET judge_id = ?, updated_at = CURRENT_TIMESTAMP WHERE trial_id = ?;

-- name: JudgeHandlesEntryInTrial :one
-- Conflict-of-interest probe for the assign-time advisory: counts entries in
-- the given trial whose handler (entries.handler_id -> competitors.id) is the
-- competitor identity owned by the candidate judge's user account
-- (competitors.user_id). The first ? is the trial id, the second the judge's
-- user id. A non-zero count means the judge handles at least one dog entered in
-- that trial. Advisory only — the caller warns but does not block the
-- assignment.
SELECT COUNT(*) AS conflicts
FROM entries e
JOIN competitors c ON c.id = e.handler_id
WHERE c.user_id = ?
  AND e.trial_id = ?
  AND e.handler_id IS NOT NULL;
