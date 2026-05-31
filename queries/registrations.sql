-- name: CreateRegistration :one
-- Files a request to enter a dog in a trial. status defaults to 'pending'
-- at the schema level; an admin later accepts it (which creates the entry)
-- or rejects it. submitted_by is the acting user, kept separate from the
-- competitor so a club secretary can register on behalf of a handler.
INSERT INTO registrations (trial_id, competitor_id, dog_id, submitted_by, notes)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: RegisteredTrialIDsForDog :many
-- Trial ids a dog is already registered in under an active status, so the
-- registration form can disable trials the dog cannot re-enter and the
-- submit handler can skip duplicates.
SELECT trial_id FROM registrations
WHERE dog_id = ? AND status NOT IN ('withdrawn', 'rejected');

-- name: ListPendingRegistrationsByCompetitor :many
-- Registrations a competitor has filed that are not yet accepted (pending
-- or waitlisted), joined to trial, event, and dog for the account entries
-- list (A5). Once accepted, a registration has an entry and surfaces
-- through the entries path instead.
SELECT
    r.id, r.status, r.submitted_at, r.dog_id,
    d.call_name AS dog_name,
    t.id AS trial_id, t.discipline, t.level, t.trial_date,
    ev.name AS event_name, ev.slug AS event_slug
FROM registrations r
JOIN dogs d ON d.id = r.dog_id
JOIN trials t ON t.id = r.trial_id
JOIN events ev ON ev.id = t.event_id
WHERE r.competitor_id = ? AND r.status IN ('pending', 'waitlisted')
ORDER BY t.trial_date DESC, r.id DESC;

-- name: CountPendingRegistrationsByEvent :one
-- Number of pending registrations across an event trials. Drives the
-- review badge in the admin sidebar and the dashboard at-a-glance (D1/D3).
SELECT COUNT(*)
FROM registrations r
JOIN trials t ON t.id = r.trial_id
WHERE t.event_id = ? AND r.status = 'pending';

-- name: CountAllPendingRegistrations :one
-- Total pending registrations across every event. Drives the admin
-- dashboard needs-review card (D1).
SELECT COUNT(*) FROM registrations WHERE status = 'pending';

-- name: ListRegistrationsByEvent :many
-- Every registration in an event, joined to dog, competitor, submitter, and
-- trial, plus the entry number once accepted. Ordered for the trial-grouped
-- review view (D5).
SELECT
    r.id, r.status, r.notes, r.submitted_at, r.entry_id, r.trial_id,
    d.call_name AS dog_name, d.breed AS dog_breed, d.registration_number AS dog_regno,
    c.handle AS competitor_handle, c.display_name AS competitor_name,
    su.email AS submitted_by_email,
    t.discipline, t.level, t.trial_date,
    e.entry_number AS entry_number
FROM registrations r
JOIN dogs d ON d.id = r.dog_id
JOIN competitors c ON c.id = r.competitor_id
JOIN users su ON su.id = r.submitted_by
JOIN trials t ON t.id = r.trial_id
LEFT JOIN entries e ON e.id = r.entry_id
WHERE t.event_id = ?
ORDER BY t.trial_date, t.discipline, t.level, r.submitted_at;

-- name: GetRegistrationDetail :one
-- One registration with the fields the accept flow denormalizes into the
-- new entry, plus the owning event id for the URL guard.
SELECT
    r.id, r.trial_id, r.competitor_id, r.dog_id, r.status,
    d.call_name AS dog_name, d.breed AS dog_breed,
    c.display_name AS competitor_name,
    t.event_id
FROM registrations r
JOIN dogs d ON d.id = r.dog_id
JOIN competitors c ON c.id = r.competitor_id
JOIN trials t ON t.id = r.trial_id
WHERE r.id = ?;

-- name: AcceptRegistration :exec
-- Marks a registration accepted and links the created entry.
UPDATE registrations
SET status = 'accepted', entry_id = ?, reviewed_by = ?, reviewed_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: SetRegistrationStatus :exec
-- Sets a registration status (waitlisted or rejected) and stamps the
-- reviewer.
UPDATE registrations
SET status = ?, reviewed_by = ?, reviewed_at = CURRENT_TIMESTAMP
WHERE id = ?;
