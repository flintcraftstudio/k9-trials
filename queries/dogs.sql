-- name: GetDogByID :one
-- Fetches a dog by primary key. Returns sql.ErrNoRows when the id misses -
-- the handler renders 404.
SELECT * FROM dogs WHERE id = ?;

-- name: ListDogsByOwner :many
-- A competitor dogs, alphabetical by call name. Backs the dogs section
-- of the competitor profile (P6).
SELECT * FROM dogs WHERE owner_id = ? ORDER BY call_name;

-- name: ListFinalizedEntriesByDog :many
-- Chronological (newest-first) finalized entries for one dog, joined to
-- trial + event for the public trial-history list (P7). handler_name is
-- carried per row because the handler of record is not always the owner.
SELECT
    e.id, e.entry_number, e.dog_name, e.handler_name, e.status,
    t.id AS trial_id, t.discipline, t.level, t.trial_date, t.template_version,
    ev.name AS event_name, ev.slug AS event_slug
FROM entries e
JOIN trials t ON t.id = e.trial_id
JOIN events ev ON ev.id = t.event_id
WHERE e.dog_id = ? AND e.status = 'finalized'
ORDER BY t.trial_date DESC, e.id DESC;
