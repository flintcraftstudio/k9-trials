package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/flintcraftstudio/k9-trials/internal/db"
)

// ListRegistrationsByEvent returns every registration in an event for the
// review screen (D5), joined to dog, competitor, submitter, and trial.
func (s *Store) ListRegistrationsByEvent(ctx context.Context, eventID int64) ([]db.ListRegistrationsByEventRow, error) {
	return s.q.ListRegistrationsByEvent(ctx, eventID)
}

// AcceptRegistration is the bridge from competitor self-service to the
// scoring pipeline: it creates the entry an accepted registration produces
// (assigning the next entry number and inheriting any judge already on the
// trial) and links it back to the registration — all in one transaction.
// Returns the new entry number. Errors if the registration is not pending.
func (s *Store) AcceptRegistration(ctx context.Context, regID, reviewerUserID int64) (int64, error) {
	reg, err := s.q.GetRegistrationDetail(ctx, regID)
	if err != nil {
		return 0, err
	}
	if reg.Status != "pending" {
		return 0, fmt.Errorf("registration %d is %s, not pending", regID, reg.Status)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	qtx := s.q.WithTx(tx)

	maxNum, err := qtx.MaxEntryNumberByTrial(ctx, reg.TrialID)
	if err != nil {
		return 0, fmt.Errorf("max entry number: %w", err)
	}
	entryNumber := maxNum + 1

	// Inherit the judge already assigned to the trial, if any, so a late
	// accept stays consistent with the rest of the trial.
	judge, err := qtx.TrialJudgeID(ctx, reg.TrialID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("trial judge: %w", err)
	}

	entry, err := qtx.CreateEntryForRegistration(ctx, db.CreateEntryForRegistrationParams{
		TrialID:     reg.TrialID,
		JudgeID:     judge,
		EntryNumber: entryNumber,
		HandlerName: reg.CompetitorName,
		DogName:     reg.DogName,
		DogBreed:    reg.DogBreed,
		DogID:       sql.NullInt64{Int64: reg.DogID, Valid: true},
		HandlerID:   sql.NullInt64{Int64: reg.CompetitorID, Valid: true},
	})
	if err != nil {
		return 0, fmt.Errorf("create entry: %w", err)
	}

	if err := qtx.AcceptRegistration(ctx, db.AcceptRegistrationParams{
		EntryID:    sql.NullInt64{Int64: entry.ID, Valid: true},
		ReviewedBy: sql.NullInt64{Int64: reviewerUserID, Valid: true},
		ID:         regID,
	}); err != nil {
		return 0, fmt.Errorf("accept registration: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return entryNumber, nil
}

// SetRegistrationStatus sets a registration to waitlisted or rejected and
// stamps the reviewer. The handler validates the target status and that the
// registration belongs to the event being reviewed.
func (s *Store) SetRegistrationStatus(ctx context.Context, regID, reviewerUserID int64, status string) error {
	return s.q.SetRegistrationStatus(ctx, db.SetRegistrationStatusParams{
		Status:     status,
		ReviewedBy: sql.NullInt64{Int64: reviewerUserID, Valid: true},
		ID:         regID,
	})
}

// ConfirmRegistrationWithdrawal grants a pending withdrawal request (Q1):
// the registration becomes withdrawn and the reviewer is stamped. The linked
// entry row and its entry_number are retained for audit (the number is not
// freed). No-op unless the registration is accepted with a pending request.
func (s *Store) ConfirmRegistrationWithdrawal(ctx context.Context, regID, reviewerUserID int64) error {
	return s.q.ConfirmRegistrationWithdrawal(ctx, db.ConfirmRegistrationWithdrawalParams{
		ReviewedBy: sql.NullInt64{Int64: reviewerUserID, Valid: true},
		ID:         regID,
	})
}

// RegistrationRef is the minimal registration identity the review handlers
// pass around: the row id and its owning event for the redirect target.
type RegistrationRef struct {
	ID      int64
	EventID int64
}

// GetRegistrationDetail returns the registration fields the review actions
// need (including the owning event id for the URL guard).
func (s *Store) GetRegistrationDetail(ctx context.Context, regID int64) (db.GetRegistrationDetailRow, error) {
	return s.q.GetRegistrationDetail(ctx, regID)
}

// JudgeEligibleUsers lists the accounts assignable as a trial judge, derived
// from account capabilities (user_roles): a user is eligible if they hold the
// 'judge' capability, or 'admin' (a superset that can judge). This replaces the
// legacy role='judge' lookup so a competitor-role account that has been granted
// the judge capability is correctly assignable.
func (s *Store) JudgeEligibleUsers(ctx context.Context) ([]db.ListJudgeEligibleUsersRow, error) {
	return s.q.ListJudgeEligibleUsers(ctx)
}

// JudgeHandlesEntryInTrial reports the conflict-of-interest advisory: whether
// the candidate judge (a users.id) handles any dog entered in the given trial.
// It is true when the trial has an entry whose handler_id resolves to a
// competitor row owned by judgeUserID (competitors.user_id = judgeUserID).
// Advisory only — callers warn but do not block on a conflict.
func (s *Store) JudgeHandlesEntryInTrial(ctx context.Context, trialID, judgeUserID int64) (bool, error) {
	n, err := s.q.JudgeHandlesEntryInTrial(ctx, db.JudgeHandlesEntryInTrialParams{
		UserID:  sql.NullInt64{Int64: judgeUserID, Valid: true},
		TrialID: trialID,
	})
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// UserHasCapability reports whether a user holds a specific account capability
// ('judge'/'admin'). Used to verify a target before assigning them as a judge:
// only judge-eligible accounts may be written to entries.judge_id. The
// 'competitor' baseline is implicit and never stored, so it is not testable
// here (every authenticated account holds it by definition).
func (s *Store) UserHasCapability(ctx context.Context, userID int64, cap string) (bool, error) {
	caps, err := s.q.UserCapabilities(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, c := range caps {
		if c == cap {
			return true, nil
		}
	}
	return false, nil
}

// TrialJudgeID returns the judge id assigned to a trial (via its entries),
// or ok=false when none is assigned.
func (s *Store) TrialJudgeID(ctx context.Context, trialID int64) (int64, bool, error) {
	id, err := s.q.TrialJudgeID(ctx, trialID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	if !id.Valid {
		return 0, false, nil
	}
	return id.Int64, true, nil
}

// AssignTrialJudge sets the judge on every entry in a trial.
func (s *Store) AssignTrialJudge(ctx context.Context, trialID, judgeID int64) error {
	return s.q.AssignTrialJudge(ctx, db.AssignTrialJudgeParams{
		JudgeID: sql.NullInt64{Int64: judgeID, Valid: true},
		TrialID: trialID,
	})
}

// ChallengeListRow is one row of the paginated admin challenge queue (D7).
type ChallengeListRow struct {
	ID          int64
	Status      string
	FiledAt     time.Time
	EntryID     int64
	EntryNumber int64
	DogName     string
	Discipline  string
	Level       int64
	EventName   string
	FilerHandle string
}

// challengeSortOrders whitelists the ORDER BY clause for each sort key. The
// queue list query is hand-built (sqlc can't parameterise a dynamic ORDER BY),
// so caller input is mapped through this table and never interpolated.
var challengeSortOrders = map[string]string{
	"newest": "ch.filed_at DESC",
	"oldest": "ch.filed_at ASC",
	"status": "ch.status ASC, ch.filed_at DESC",
}

// ChallengeSortValid reports whether sort is a known queue sort key.
func ChallengeSortValid(sort string) bool {
	_, ok := challengeSortOrders[sort]
	return ok
}

// ListChallengesPage returns one page of the cross-event challenge queue (D7).
// status filters by exact status when non-empty; sort is a key into
// challengeSortOrders (unknown keys fall back to newest); limit/offset window
// the result.
func (s *Store) ListChallengesPage(ctx context.Context, status, sort string, limit, offset int64) ([]ChallengeListRow, error) {
	order, ok := challengeSortOrders[sort]
	if !ok {
		order = challengeSortOrders["newest"]
	}
	query := `
SELECT
    ch.id, ch.status, ch.filed_at, ch.entry_id,
    e.entry_number, e.dog_name,
    t.discipline, t.level,
    ev.name, c.handle
FROM challenges ch
JOIN entries e ON e.id = ch.entry_id
JOIN trials t ON t.id = e.trial_id
JOIN events ev ON ev.id = t.event_id
JOIN competitors c ON c.id = ch.filed_by
WHERE (? = '' OR ch.status = ?)
ORDER BY ` + order + `
LIMIT ? OFFSET ?`
	rows, err := s.db.QueryContext(ctx, query, status, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ChallengeListRow
	for rows.Next() {
		var r ChallengeListRow
		if err := rows.Scan(
			&r.ID, &r.Status, &r.FiledAt, &r.EntryID,
			&r.EntryNumber, &r.DogName,
			&r.Discipline, &r.Level,
			&r.EventName, &r.FilerHandle,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// CountChallenges returns the number of challenges matching the status filter
// (empty = all), for the queue's pagination.
func (s *Store) CountChallenges(ctx context.Context, status string) (int64, error) {
	return s.q.CountChallenges(ctx, status)
}

// ChallengeStatusCounts returns the global challenge tally keyed by status,
// independent of any filter — drives the filter-chip counts and header summary.
func (s *Store) ChallengeStatusCounts(ctx context.Context) (map[string]int, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT status, COUNT(*) FROM challenges GROUP BY status`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]int, 4)
	for rows.Next() {
		var status string
		var n int
		if err := rows.Scan(&status, &n); err != nil {
			return nil, err
		}
		out[status] = n
	}
	return out, rows.Err()
}

// GetChallengeDetail returns one challenge with its disputed entry context.
func (s *Store) GetChallengeDetail(ctx context.Context, id int64) (db.GetChallengeDetailRow, error) {
	return s.q.GetChallengeDetail(ctx, id)
}

// UpdateChallengeStatus advances a challenge. resolvedBy is the acting
// admin user when resolving or dismissing; pass 0 (and a zero time) while
// only starting review.
func (s *Store) UpdateChallengeStatus(ctx context.Context, id int64, status, notes string, resolvedBy int64, resolvedAt time.Time) error {
	by := sql.NullInt64{}
	at := sql.NullTime{}
	if resolvedBy != 0 {
		by = sql.NullInt64{Int64: resolvedBy, Valid: true}
	}
	if !resolvedAt.IsZero() {
		at = sql.NullTime{Time: resolvedAt, Valid: true}
	}
	return s.q.UpdateChallengeStatus(ctx, db.UpdateChallengeStatusParams{
		Status:          status,
		ResolutionNotes: notes,
		ResolvedBy:      by,
		ResolvedAt:      at,
		ID:              id,
	})
}

// ListUsersWithCaps returns every user with their competitor identity (when
// any) and their explicit account capabilities (comma-separated, sorted), for
// the users and roles admin (D8). The 'competitor' baseline is implicit, so an
// empty caps string means competitor-only. The display label is derived from
// caps.
func (s *Store) ListUsersWithCaps(ctx context.Context) ([]db.ListUsersWithCapsRow, error) {
	return s.q.ListUsersWithCaps(ctx)
}
