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
