-- name: GetCompetitorByHandle :one
-- Fetches a competitor by their public URL handle. Returns sql.ErrNoRows
-- when the handle does not resolve - the handler renders 404.
SELECT * FROM competitors WHERE handle = ?;

-- name: GetCompetitorByID :one
-- Fetches a competitor by primary key. Used to resolve a dog owner for
-- the owner link on the dog profile (P7).
SELECT * FROM competitors WHERE id = ?;

-- name: ListRecentCompetitors :many
-- Newest competitors first, capped by ?. Backs the recently-active
-- cluster on the directory (P5) when there is no search term. Ordering is
-- by row age, not last-competed - a cheap proxy until activity volume
-- justifies a date join.
SELECT * FROM competitors
ORDER BY created_at DESC, id DESC
LIMIT ?;

-- name: SearchCompetitors :many
-- Substring search across competitor handle/display_name and the
-- call_name / registered_name / registration_number of their dogs. The
-- inner query collects matching competitor ids (a LEFT JOIN to dogs can
-- multiply rows, so the outer query de-dupes via IN). @term is the raw
-- needle; the handler trims it. Capped by @lim.
SELECT * FROM competitors
WHERE id IN (
    SELECT c.id FROM competitors c
    LEFT JOIN dogs d ON d.owner_id = c.id
    WHERE c.handle LIKE '%' || @term || '%'
       OR c.display_name LIKE '%' || @term || '%'
       OR d.call_name LIKE '%' || @term || '%'
       OR d.registered_name LIKE '%' || @term || '%'
       OR d.registration_number LIKE '%' || @term || '%'
)
ORDER BY display_name
LIMIT @lim;

-- name: CountDogsByOwner :one
-- Number of dogs a competitor owns. Drives the dog-count chip on the
-- directory cards and profile header.
SELECT COUNT(*) FROM dogs WHERE owner_id = ?;

-- name: CountFinalizedByHandler :one
-- Number of finalized entries a competitor has handled. Drives the
-- finalized-entry-count chip.
SELECT COUNT(*) FROM entries WHERE handler_id = ? AND status = 'finalized';

-- name: LastCompetedByHandler :one
-- Trial date of the competitor most recent finalized entry. Returns
-- sql.ErrNoRows when they have never finalized a run - callers treat that
-- as has-not-competed-yet rather than an error.
SELECT t.trial_date
FROM entries e
JOIN trials t ON t.id = e.trial_id
WHERE e.handler_id = ? AND e.status = 'finalized'
ORDER BY t.trial_date DESC
LIMIT 1;

-- name: ListFinalizedEntriesByHandler :many
-- Chronological (newest-first) finalized entries handled by a competitor,
-- joined to their trial + event for the public event-history list (P6).
-- template_version is included so the caller can re-evaluate the score.
SELECT
    e.id, e.entry_number, e.dog_name, e.dog_breed, e.handler_name, e.status,
    t.id AS trial_id, t.discipline, t.level, t.trial_date, t.template_version,
    ev.name AS event_name, ev.slug AS event_slug
FROM entries e
JOIN trials t ON t.id = e.trial_id
JOIN events ev ON ev.id = t.event_id
WHERE e.handler_id = ? AND e.status = 'finalized'
ORDER BY t.trial_date DESC, e.id DESC;
