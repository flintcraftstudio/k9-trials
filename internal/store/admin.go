package store

import (
	"context"
	"time"

	"github.com/flintcraftstudio/k9-trials/internal/db"
)

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

// CreateEvent inserts a new event owned by createdBy and returns it.
func (s *Store) CreateEvent(ctx context.Context, in EventInput, createdBy int64) (db.Event, error) {
	return s.q.CreateEvent(ctx, db.CreateEventParams{
		Slug:      in.Slug,
		Name:      in.Name,
		Location:  in.Location,
		StartDate: in.StartDate,
		EndDate:   in.EndDate,
		Status:    in.Status,
		CreatedBy: createdBy,
	})
}

// UpdateEvent saves edits to an event. The slug is immutable after
// creation (it is the public URL), so it is not part of the update.
func (s *Store) UpdateEvent(ctx context.Context, id int64, in EventInput) (db.Event, error) {
	return s.q.UpdateEvent(ctx, db.UpdateEventParams{
		Name:      in.Name,
		Location:  in.Location,
		StartDate: in.StartDate,
		EndDate:   in.EndDate,
		Status:    in.Status,
		ID:        id,
	})
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
