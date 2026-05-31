-- name: CreateChallenge :one
-- Files a dispute against a finalized entry. status defaults to 'open' at
-- the schema level. Returns the new row.
INSERT INTO challenges (entry_id, filed_by, reason)
VALUES (?, ?, ?)
RETURNING *;

-- name: ListChallengesByFiler :many
-- Every challenge a competitor has filed, newest first, joined to the
-- disputed entry plus its trial and event for the list rows (A7).
SELECT
    c.id, c.entry_id, c.reason, c.status, c.filed_at, c.updated_at,
    e.entry_number, e.dog_name,
    t.discipline, t.level, t.trial_date,
    ev.name AS event_name, ev.slug AS event_slug
FROM challenges c
JOIN entries e ON e.id = c.entry_id
JOIN trials t ON t.id = e.trial_id
JOIN events ev ON ev.id = t.event_id
WHERE c.filed_by = ?
ORDER BY c.filed_at DESC;

-- name: CountOpenChallengesByFiler :one
-- Number of a competitor unresolved challenges (open or under review).
-- Drives the dashboard open-challenge banner (A1).
SELECT COUNT(*) FROM challenges
WHERE filed_by = ? AND status IN ('open', 'under_review');

-- name: GetChallengeForEntryByFiler :one
-- Most recent challenge a competitor filed against one entry, if any.
-- Used to stop duplicate filings and to label an already-challenged
-- entry. Returns sql.ErrNoRows when none exists.
SELECT * FROM challenges
WHERE entry_id = ? AND filed_by = ?
ORDER BY filed_at DESC
LIMIT 1;

-- name: CountOpenChallengesGlobal :one
-- Total unresolved challenges (open or under review) across all
-- competitors. Drives the admin dashboard needs-review card (D1).
SELECT COUNT(*) FROM challenges WHERE status IN ('open', 'under_review');

-- name: ListAllChallenges :many
-- Every challenge across all events, newest first, joined to the disputed
-- entry, its trial and event, and the filer. Backs the admin review queue
-- (D7).
SELECT
    ch.id, ch.status, ch.filed_at, ch.entry_id,
    e.entry_number, e.dog_name,
    t.discipline, t.level,
    ev.name AS event_name,
    c.handle AS filer_handle
FROM challenges ch
JOIN entries e ON e.id = ch.entry_id
JOIN trials t ON t.id = e.trial_id
JOIN events ev ON ev.id = t.event_id
JOIN competitors c ON c.id = ch.filed_by
ORDER BY ch.filed_at DESC;

-- name: GetChallengeDetail :one
-- One challenge with everything the review detail needs: the dispute, the
-- entry under dispute, its trial and event, and the filer identity.
SELECT
    ch.id, ch.status, ch.reason, ch.resolution_notes, ch.filed_at, ch.entry_id,
    e.entry_number, e.dog_name, e.status AS entry_status,
    t.discipline, t.level, t.trial_date,
    ev.name AS event_name, ev.slug AS event_slug,
    c.handle AS filer_handle, c.display_name AS filer_name
FROM challenges ch
JOIN entries e ON e.id = ch.entry_id
JOIN trials t ON t.id = e.trial_id
JOIN events ev ON ev.id = t.event_id
JOIN competitors c ON c.id = ch.filed_by
WHERE ch.id = ?;

-- name: UpdateChallengeStatus :exec
-- Advances a challenge through its workflow. resolved_by and resolved_at
-- are set when resolving or dismissing, null while under review.
UPDATE challenges
SET status = ?, resolution_notes = ?, resolved_by = ?, resolved_at = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;
