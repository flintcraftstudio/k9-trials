package store

import (
	"context"
	"fmt"

	"github.com/flintcraftstudio/k9-trials/internal/db"
)

// EventWithTrials bundles an event with its trials. Used by the public
// events index (P1) and event detail (P2) so the view can derive the
// discipline set, trial count, and per-trial rows without a second round
// trip from the handler.
type EventWithTrials struct {
	Event  db.Event
	Trials []db.Trial
}

// ListPublicEvents returns every published or closed event paired with
// its trials, newest first (the underlying ListPublishedEvents orders by
// start_date desc). Draft events are excluded — they are admin-only until
// published. The per-event trial fetch is a small N+1; acceptable while
// the event count is low, revisit with a JOIN query if it grows.
func (s *Store) ListPublicEvents(ctx context.Context) ([]EventWithTrials, error) {
	events, err := s.q.ListPublishedEvents(ctx)
	if err != nil {
		return nil, fmt.Errorf("list published events: %w", err)
	}
	out := make([]EventWithTrials, 0, len(events))
	for _, e := range events {
		trials, err := s.q.ListTrialsByEvent(ctx, e.ID)
		if err != nil {
			return nil, fmt.Errorf("list trials for event %d: %w", e.ID, err)
		}
		out = append(out, EventWithTrials{Event: e, Trials: trials})
	}
	return out, nil
}

// LoadPublicEvent returns one published or closed event by slug plus its
// trials. Returns sql.ErrNoRows (propagated from GetEventBySlug) when the
// slug doesn't resolve; the handler renders 404. A draft event resolves
// here but the handler is responsible for hiding it — see EventDetail.
func (s *Store) LoadPublicEvent(ctx context.Context, slug string) (EventWithTrials, error) {
	event, err := s.q.GetEventBySlug(ctx, slug)
	if err != nil {
		return EventWithTrials{}, err
	}
	trials, err := s.q.ListTrialsByEvent(ctx, event.ID)
	if err != nil {
		return EventWithTrials{}, fmt.Errorf("list trials for event %d: %w", event.ID, err)
	}
	return EventWithTrials{Event: event, Trials: trials}, nil
}

// LoadTrialWithEntries returns a trial, its parent event, and every entry
// on it (ordered by entry number). Returns sql.ErrNoRows when the trial
// id misses. Backs the public leaderboard (P3).
func (s *Store) LoadTrialWithEntries(ctx context.Context, trialID int64) (db.Trial, db.Event, []db.Entry, error) {
	trial, err := s.q.GetTrialByID(ctx, trialID)
	if err != nil {
		return db.Trial{}, db.Event{}, nil, err
	}
	event, err := s.q.GetEventByID(ctx, trial.EventID)
	if err != nil {
		return db.Trial{}, db.Event{}, nil, fmt.Errorf("get event %d: %w", trial.EventID, err)
	}
	entries, err := s.q.ListEntriesByTrial(ctx, trial.ID)
	if err != nil {
		return db.Trial{}, db.Event{}, nil, fmt.Errorf("list entries for trial %d: %w", trial.ID, err)
	}
	return trial, event, entries, nil
}
