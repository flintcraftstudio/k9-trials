-- Recent-activity sources for the admin dashboard (D1). Each returns its own
-- typed timestamp so the driver parses it cleanly; the handler merges and
-- sorts them into one feed (a cross-table UNION would lose column-type
-- affinity and break time scanning).

-- name: RecentFinalizedEntries :many
SELECT e.entry_number, e.dog_name, e.updated_at, ev.name AS event_name
FROM entries e
JOIN trials t ON t.id = e.trial_id
JOIN events ev ON ev.id = t.event_id
WHERE e.status = 'finalized'
ORDER BY e.updated_at DESC
LIMIT ?;

-- name: RecentAcceptedRegistrations :many
SELECT d.call_name AS dog_name, r.reviewed_at, ev.name AS event_name
FROM registrations r
JOIN dogs d ON d.id = r.dog_id
JOIN trials t ON t.id = r.trial_id
JOIN events ev ON ev.id = t.event_id
WHERE r.status = 'accepted' AND r.reviewed_at IS NOT NULL
ORDER BY r.reviewed_at DESC
LIMIT ?;

-- name: RecentChallengesFiled :many
SELECT ch.filed_at, e.entry_number, e.dog_name, c.handle, ev.name AS event_name
FROM challenges ch
JOIN entries e ON e.id = ch.entry_id
JOIN trials t ON t.id = e.trial_id
JOIN events ev ON ev.id = t.event_id
JOIN competitors c ON c.id = ch.filed_by
ORDER BY ch.filed_at DESC
LIMIT ?;

-- name: RecentPublishedEvents :many
SELECT name, published_at
FROM events
WHERE published_at IS NOT NULL
ORDER BY published_at DESC
LIMIT ?;
