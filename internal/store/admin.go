package store

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/flintcraftstudio/k9-trials/internal/db"
)

// ActivityItem is one entry in the admin dashboard recent-activity feed (D1).
type ActivityItem struct {
	When  time.Time
	Kind  string // "finalized" | "accepted" | "challenge" | "published"
	Text  string
	Event string // owning event name, "" when not applicable
}

// RecentActivity merges recent finalized entries, accepted registrations,
// filed challenges, and published events into one newest-first feed capped at
// limit. Each source is queried separately so its timestamp scans cleanly.
func (s *Store) RecentActivity(ctx context.Context, limit int) ([]ActivityItem, error) {
	n := int64(limit)
	items := make([]ActivityItem, 0, limit*4)

	fin, err := s.q.RecentFinalizedEntries(ctx, n)
	if err != nil {
		return nil, err
	}
	for _, e := range fin {
		items = append(items, ActivityItem{
			When: e.UpdatedAt, Kind: "finalized", Event: e.EventName,
			Text: fmt.Sprintf("Entry #%02d (%s) finalized", e.EntryNumber, e.DogName),
		})
	}

	acc, err := s.q.RecentAcceptedRegistrations(ctx, n)
	if err != nil {
		return nil, err
	}
	for _, r := range acc {
		if !r.ReviewedAt.Valid {
			continue
		}
		items = append(items, ActivityItem{
			When: r.ReviewedAt.Time, Kind: "accepted", Event: r.EventName,
			Text: fmt.Sprintf("Registration accepted — %s", r.DogName),
		})
	}

	ch, err := s.q.RecentChallengesFiled(ctx, n)
	if err != nil {
		return nil, err
	}
	for _, c := range ch {
		items = append(items, ActivityItem{
			When: c.FiledAt, Kind: "challenge", Event: c.EventName,
			Text: fmt.Sprintf("@%s filed a challenge on entry #%02d (%s)", c.Handle, c.EntryNumber, c.DogName),
		})
	}

	pub, err := s.q.RecentPublishedEvents(ctx, n)
	if err != nil {
		return nil, err
	}
	for _, e := range pub {
		if !e.PublishedAt.Valid {
			continue
		}
		items = append(items, ActivityItem{
			When: e.PublishedAt.Time, Kind: "published",
			Text: fmt.Sprintf("Published %s", e.Name),
		})
	}

	sort.Slice(items, func(i, j int) bool { return items[i].When.After(items[j].When) })
	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

// ListEvents returns every event (all statuses, newest start date first)
// for the admin events list (D2) and dashboard (D1).
func (s *Store) ListEvents(ctx context.Context) ([]db.Event, error) {
	return s.q.ListEvents(ctx)
}

// GetEvent fetches an event by id. Returns sql.ErrNoRows when it misses.
func (s *Store) GetEvent(ctx context.Context, id int64) (db.Event, error) {
	return s.q.GetEventByID(ctx, id)
}

// EventInput carries the writable event fields from the admin form.
type EventInput struct {
	Slug      string
	Name      string
	Location  string
	StartDate time.Time
	EndDate   time.Time
	Status    string
}

// CreateEvent inserts a new event owned by createdBy and returns it. An
// event created directly as published is stamped with published_at now.
func (s *Store) CreateEvent(ctx context.Context, in EventInput, createdBy int64) (db.Event, error) {
	var published sql.NullTime
	if in.Status == "published" {
		published = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	}
	return s.q.CreateEvent(ctx, db.CreateEventParams{
		Slug:        in.Slug,
		Name:        in.Name,
		Location:    in.Location,
		StartDate:   in.StartDate,
		EndDate:     in.EndDate,
		Status:      in.Status,
		CreatedBy:   createdBy,
		PublishedAt: published,
	})
}

// UpdateEvent saves edits to an event. The slug is immutable after
// creation (it is the public URL), so it is not part of the update.
// published_at is stamped the first time the event becomes published and
// retained thereafter, so the audit record survives a later status change.
func (s *Store) UpdateEvent(ctx context.Context, id int64, in EventInput) (db.Event, error) {
	cur, err := s.q.GetEventByID(ctx, id)
	if err != nil {
		return db.Event{}, err
	}
	return s.q.UpdateEvent(ctx, db.UpdateEventParams{
		Name:        in.Name,
		Location:    in.Location,
		StartDate:   in.StartDate,
		EndDate:     in.EndDate,
		Status:      in.Status,
		PublishedAt: stampPublished(cur.PublishedAt, in.Status),
		ID:          id,
	})
}

// SetEventStatus transitions only the status (D3 Archive / Restore actions),
// preserving the rest of the metadata. published_at is stamped on the first
// transition into 'published'.
func (s *Store) SetEventStatus(ctx context.Context, id int64, status string) (db.Event, error) {
	cur, err := s.q.GetEventByID(ctx, id)
	if err != nil {
		return db.Event{}, err
	}
	return s.q.SetEventStatus(ctx, db.SetEventStatusParams{
		Status:      status,
		PublishedAt: stampPublished(cur.PublishedAt, status),
		ID:          id,
	})
}

// stampPublished keeps an existing published_at, or stamps now the first time
// an event enters the published status.
func stampPublished(cur sql.NullTime, newStatus string) sql.NullTime {
	if cur.Valid {
		return cur
	}
	if newStatus == "published" {
		return sql.NullTime{Time: time.Now().UTC(), Valid: true}
	}
	return cur
}

// SubscribeToEvent records a competitor's notify-me request for an event
// (Q4 / R1c). Idempotent — a repeat subscribe is a no-op.
func (s *Store) SubscribeToEvent(ctx context.Context, eventID, competitorID int64) error {
	return s.q.SubscribeToEvent(ctx, db.SubscribeToEventParams{
		EventID:      eventID,
		CompetitorID: competitorID,
	})
}

// HasEventSubscription reports whether a competitor is already subscribed to
// an event, for the R1c notify-me state.
func (s *Store) HasEventSubscription(ctx context.Context, eventID, competitorID int64) (bool, error) {
	n, err := s.q.HasEventSubscription(ctx, db.HasEventSubscriptionParams{
		EventID:      eventID,
		CompetitorID: competitorID,
	})
	return n > 0, err
}

// ListEventSubscribers returns an event's un-notified subscribers (with their
// email) for the publish-transition hook.
func (s *Store) ListEventSubscribers(ctx context.Context, eventID int64) ([]db.ListEventSubscribersRow, error) {
	return s.q.ListEventSubscribers(ctx, eventID)
}

// MarkEventSubscribersNotified stamps notified_at on an event's un-notified
// subscribers so a later re-publish does not notify them twice.
func (s *Store) MarkEventSubscribersNotified(ctx context.Context, eventID int64) error {
	return s.q.MarkEventSubscribersNotified(ctx, eventID)
}

// CountEntriesByEvent returns the total entries across all trials of an event.
func (s *Store) CountEntriesByEvent(ctx context.Context, eventID int64) (int64, error) {
	return s.q.CountEntriesByEvent(ctx, eventID)
}

// CountTrialsWithJudgeByEvent returns how many of an event's trials have a
// judge assigned.
func (s *Store) CountTrialsWithJudgeByEvent(ctx context.Context, eventID int64) (int64, error) {
	return s.q.CountTrialsWithJudgeByEvent(ctx, eventID)
}

// EventSlugAvailable reports whether a slug is free, ignoring the event
// with id exceptID (pass 0 on create to ignore nobody).
func (s *Store) EventSlugAvailable(ctx context.Context, slug string, exceptID int64) (bool, error) {
	n, err := s.q.CountEventsBySlug(ctx, db.CountEventsBySlugParams{Slug: slug, ID: exceptID})
	if err != nil {
		return false, err
	}
	return n == 0, nil
}

// CountTrialsByEvent returns how many trials an event has.
func (s *Store) CountTrialsByEvent(ctx context.Context, eventID int64) (int64, error) {
	return s.q.CountTrialsByEvent(ctx, eventID)
}

// CountPendingRegistrationsByEvent returns the pending-registration count
// for an event.
func (s *Store) CountPendingRegistrationsByEvent(ctx context.Context, eventID int64) (int64, error) {
	return s.q.CountPendingRegistrationsByEvent(ctx, eventID)
}

// CountAllPendingRegistrations returns the pending-registration count
// across every event.
func (s *Store) CountAllPendingRegistrations(ctx context.Context) (int64, error) {
	return s.q.CountAllPendingRegistrations(ctx)
}

// CountAllOpenChallenges returns the unresolved-challenge count across all
// competitors (admin dashboard).
func (s *Store) CountAllOpenChallenges(ctx context.Context) (int64, error) {
	return s.q.CountOpenChallengesGlobal(ctx)
}

// TrialsByEvent returns an event trials, ordered by date then discipline
// then level.
func (s *Store) TrialsByEvent(ctx context.Context, eventID int64) ([]db.Trial, error) {
	return s.q.ListTrialsByEvent(ctx, eventID)
}

// GetTrial fetches a trial by id. Returns sql.ErrNoRows when it misses.
func (s *Store) GetTrial(ctx context.Context, id int64) (db.Trial, error) {
	return s.q.GetTrialByID(ctx, id)
}

// TrialInput carries the writable trial fields from the admin form.
type TrialInput struct {
	Discipline      string
	Level           int64
	TrialDate       time.Time
	TemplateVersion string
	Status          string
}

// CreateTrial inserts a new trial under an event and returns it.
func (s *Store) CreateTrial(ctx context.Context, eventID int64, in TrialInput) (db.Trial, error) {
	return s.q.CreateTrial(ctx, db.CreateTrialParams{
		EventID:         eventID,
		Discipline:      in.Discipline,
		Level:           in.Level,
		TrialDate:       in.TrialDate,
		TemplateVersion: in.TemplateVersion,
		Status:          in.Status,
	})
}

// DeleteTrial removes a trial. Entries cascade via the foreign key.
func (s *Store) DeleteTrial(ctx context.Context, id int64) error {
	return s.q.DeleteTrial(ctx, id)
}

// CountEntriesForTrial returns how many entries a trial has.
func (s *Store) CountEntriesForTrial(ctx context.Context, trialID int64) (int64, error) {
	return s.q.CountEntriesByTrial(ctx, trialID)
}

// TrialJudgeEmail returns the email of the judge assigned to a trial (via
// any of its entries), or sql.ErrNoRows when none is assigned.
func (s *Store) TrialJudgeEmail(ctx context.Context, trialID int64) (string, error) {
	return s.q.TrialJudgeEmail(ctx, trialID)
}
