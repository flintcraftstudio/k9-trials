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

-- name: CreateDog :one
-- Inserts a dog under an owner. Only call_name is required; the rest
-- default empty/null so a dog can exist in the registry before competing.
-- date_of_birth is passed as a nullable param. Returns the new row.
INSERT INTO dogs (owner_id, call_name, registered_name, breed, date_of_birth, registration_number)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateDog :exec
-- Saves an edited dog. Scoped by owner_id as well as id so a competitor
-- cannot edit a dog they do not own even with a guessed id.
UPDATE dogs
SET call_name = ?, registered_name = ?, breed = ?, date_of_birth = ?,
    registration_number = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND owner_id = ?;

-- name: DeleteDog :exec
-- Removes a dog, scoped by owner so a competitor can only delete their
-- own. Historical entries keep their denormalized dog_name snapshot, so
-- past results stay readable after the dog row is gone.
DELETE FROM dogs WHERE id = ? AND owner_id = ?;

-- name: GetDogForOwner :one
-- Fetches a dog by id scoped to its owner. Returns sql.ErrNoRows when the
-- id misses or belongs to another competitor, so the edit form 404s
-- rather than leaking another owner dog.
SELECT * FROM dogs WHERE id = ? AND owner_id = ?;

-- name: CountEntriesByDog :one
-- Total entries (any status) recorded against a dog. Drives the
-- no-entries-yet hint on the dogs list.
SELECT COUNT(*) FROM entries WHERE dog_id = ?;

-- name: LastCompetedByDog :one
-- Trial date of a dog most recent finalized entry. Returns sql.ErrNoRows
-- when the dog has never finalized a run.
SELECT t.trial_date
FROM entries e
JOIN trials t ON t.id = e.trial_id
WHERE e.dog_id = ? AND e.status = 'finalized'
ORDER BY t.trial_date DESC
LIMIT 1;
