package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/flintcraftstudio/k9-trials/internal/db"
	"github.com/flintcraftstudio/k9-trials/internal/scoring"
)

// LoadEntryWithTrial returns the entry plus its parent trial and event in
// one call. All three resolve to NotFound if any leg misses — the caller
// renders 404 rather than partial chrome.
func (s *Store) LoadEntryWithTrial(ctx context.Context, entryID int64) (db.Entry, db.Trial, db.Event, error) {
	entry, err := s.q.GetEntryByID(ctx, entryID)
	if err != nil {
		return db.Entry{}, db.Trial{}, db.Event{}, err
	}
	trial, err := s.q.GetTrialByID(ctx, entry.TrialID)
	if err != nil {
		return db.Entry{}, db.Trial{}, db.Event{}, err
	}
	event, err := s.q.GetEventByID(ctx, trial.EventID)
	if err != nil {
		return db.Entry{}, db.Trial{}, db.Event{}, err
	}
	return entry, trial, event, nil
}

// LoadInputsForEntry projects the four append-only input tables into a
// scoring.ScoresheetInputs ready for EvaluateScoresheet. Latest-write-wins
// collapse is applied to criterion scores per (exercise_code, criterion_code) —
// the list queries return ORDER BY created_at, so the last row in each
// group is the winner. Penalty occurrences, auto-trigger firings, and
// modifier applications are kept verbatim because the engine counts them.
func (s *Store) LoadInputsForEntry(ctx context.Context, entryID int64) (scoring.ScoresheetInputs, error) {
	csRows, err := s.q.ListCriterionScoresByEntry(ctx, entryID)
	if err != nil {
		return scoring.ScoresheetInputs{}, fmt.Errorf("list criterion scores: %w", err)
	}
	poRows, err := s.q.ListPenaltyOccurrencesByEntry(ctx, entryID)
	if err != nil {
		return scoring.ScoresheetInputs{}, fmt.Errorf("list penalty occurrences: %w", err)
	}
	atRows, err := s.q.ListAutoTriggerFiringsByEntry(ctx, entryID)
	if err != nil {
		return scoring.ScoresheetInputs{}, fmt.Errorf("list auto trigger firings: %w", err)
	}
	maRows, err := s.q.ListModifierApplicationsByEntry(ctx, entryID)
	if err != nil {
		return scoring.ScoresheetInputs{}, fmt.Errorf("list modifier applications: %w", err)
	}

	type csKey struct{ exercise, criterion string }
	latest := make(map[csKey]scoring.CriterionScore, len(csRows))
	for _, row := range csRows {
		k := csKey{row.ExerciseCode, row.CriterionCode}
		latest[k] = scoring.CriterionScore{
			ExerciseCode:  row.ExerciseCode,
			CriterionCode: row.CriterionCode,
			Points:        scoring.Points(row.Points),
		}
	}
	criterionScores := make([]scoring.CriterionScore, 0, len(latest))
	for _, cs := range latest {
		criterionScores = append(criterionScores, cs)
	}

	penalties := make([]scoring.PenaltyOccurrence, 0, len(poRows))
	for _, row := range poRows {
		penalties = append(penalties, scoring.PenaltyOccurrence{
			ExerciseCode: row.ExerciseCode,
			EventCode:    row.EventCode,
		})
	}

	triggers := make([]scoring.AutoTriggerFiring, 0, len(atRows))
	for _, row := range atRows {
		triggers = append(triggers, scoring.AutoTriggerFiring{
			ExerciseCode: row.ExerciseCode,
			TriggerCode:  row.TriggerCode,
		})
	}

	modifiers := make([]scoring.ModifierApplication, 0, len(maRows))
	for _, row := range maRows {
		modifiers = append(modifiers, scoring.ModifierApplication{
			ModifierCode: row.ModifierCode,
		})
	}

	return scoring.ScoresheetInputs{
		CriterionScores:    criterionScores,
		PenaltyOccurrences: penalties,
		AutoTriggers:       triggers,
		Modifiers:          modifiers,
	}, nil
}

// JudgeQueueResult is the queue payload for one judge's active trial.
// Trial is nil when the judge has no assigned entries yet.
type JudgeQueueResult struct {
	Trial   *db.Trial
	Event   *db.Event
	Entries []db.Entry
}

// ListJudgeQueue picks the most recently-updated trial for entries
// assigned to this judge and returns every entry on that trial (the
// judge's queue is the trial-day roster, not just their slice). Other
// trials are ignored for v1 — a judge runs one trial at a time.
func (s *Store) ListJudgeQueue(ctx context.Context, judgeID int64) (JudgeQueueResult, error) {
	mine, err := s.q.ListEntriesByJudge(ctx, sql.NullInt64{Int64: judgeID, Valid: true})
	if err != nil {
		return JudgeQueueResult{}, fmt.Errorf("list entries by judge: %w", err)
	}
	if len(mine) == 0 {
		return JudgeQueueResult{}, nil
	}

	// Pick the trial id with the most recent UpdatedAt among assigned entries.
	var pickTrialID int64
	var pickAt = mine[0].UpdatedAt
	pickTrialID = mine[0].TrialID
	for _, e := range mine[1:] {
		if e.UpdatedAt.After(pickAt) {
			pickAt = e.UpdatedAt
			pickTrialID = e.TrialID
		}
	}

	trial, err := s.q.GetTrialByID(ctx, pickTrialID)
	if err != nil {
		return JudgeQueueResult{}, fmt.Errorf("get trial: %w", err)
	}
	event, err := s.q.GetEventByID(ctx, trial.EventID)
	if err != nil {
		return JudgeQueueResult{}, fmt.Errorf("get event: %w", err)
	}
	entries, err := s.q.ListEntriesByTrial(ctx, trial.ID)
	if err != nil {
		return JudgeQueueResult{}, fmt.Errorf("list entries by trial: %w", err)
	}
	return JudgeQueueResult{Trial: &trial, Event: &event, Entries: entries}, nil
}
