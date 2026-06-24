package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/flintcraftstudio/k9-trials/internal/db"

	"golang.org/x/crypto/bcrypt"
)

// CreateCompetitorAccount provisions a self-service competitor in one
// transaction: a users row (role=competitor) and the matching competitors
// identity linked by user_id. Either both land or neither does, so a
// failed handle insert never leaves an orphan login. Returns the new
// user id (the value the caller starts a session for).
func (s *Store) CreateCompetitorAccount(ctx context.Context, email, password, displayName, handle string) (int64, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		"INSERT INTO users (email, password_hash, role) VALUES (?, ?, ?)",
		email, string(hash), "competitor",
	)
	if err != nil {
		return 0, fmt.Errorf("create user: %w", err)
	}
	userID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	qtx := s.q.WithTx(tx)
	if _, err := qtx.CreateCompetitor(ctx, db.CreateCompetitorParams{
		UserID:      sql.NullInt64{Int64: userID, Valid: true},
		Handle:      handle,
		DisplayName: displayName,
		Bio:         "",
	}); err != nil {
		return 0, fmt.Errorf("create competitor: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return userID, nil
}

// CurrentCompetitor resolves the competitor identity for a logged-in user.
// Returns sql.ErrNoRows when the account has no competitor row (an admin
// seeded without one), which callers translate into a neutral state rather
// than an error.
func (s *Store) CurrentCompetitor(ctx context.Context, userID int64) (db.Competitor, error) {
	return s.q.GetCompetitorByUserID(ctx, sql.NullInt64{Int64: userID, Valid: true})
}

// HandleAvailable reports whether a handle is free, ignoring the
// competitor with id exceptID (pass 0 to ignore nobody — the signup case).
// Used both for live availability checks and to guard a handle edit.
func (s *Store) HandleAvailable(ctx context.Context, handle string, exceptID int64) (bool, error) {
	n, err := s.q.CountCompetitorsByHandle(ctx, db.CountCompetitorsByHandleParams{
		Handle: handle,
		ID:     exceptID,
	})
	if err != nil {
		return false, err
	}
	return n == 0, nil
}

// UpdateCompetitorProfile saves the editable profile fields for a
// competitor.
func (s *Store) UpdateCompetitorProfile(ctx context.Context, id int64, displayName, handle, bio string) error {
	return s.q.UpdateCompetitorProfile(ctx, db.UpdateCompetitorProfileParams{
		DisplayName: displayName,
		Handle:      handle,
		Bio:         bio,
		ID:          id,
	})
}

// ListHandlerEntries returns every entry a competitor has handled across
// all events and statuses, newest trial first. Backs the account entries
// list (A5) and the dashboard up-next / recent clusters (A1).
func (s *Store) ListHandlerEntries(ctx context.Context, competitorID int64) ([]db.ListEntriesByHandlerRow, error) {
	return s.q.ListEntriesByHandler(ctx, sql.NullInt64{Int64: competitorID, Valid: true})
}

// CountOpenChallenges returns how many unresolved challenges a competitor
// has filed (open or under review). Drives the dashboard banner (A1).
func (s *Store) CountOpenChallenges(ctx context.Context, competitorID int64) (int64, error) {
	return s.q.CountOpenChallengesByFiler(ctx, competitorID)
}

// CountOwnerDogs returns how many dogs a competitor owns. Drives the
// dashboard summary line (A1).
func (s *Store) CountOwnerDogs(ctx context.Context, ownerID int64) (int64, error) {
	return s.q.CountDogsByOwner(ctx, ownerID)
}

// OwnerDogs returns a competitor's dogs as plain rows (no activity
// counts), alphabetical by call name. Used by the registration form where
// only the identity fields are needed.
func (s *Store) OwnerDogs(ctx context.Context, ownerID int64) ([]db.Dog, error) {
	return s.q.ListDogsByOwner(ctx, ownerID)
}

// RegisteredTrialIDsForDog returns the set of trial ids a dog already holds
// an active (non-withdrawn, non-rejected) registration in.
func (s *Store) RegisteredTrialIDsForDog(ctx context.Context, dogID int64) (map[int64]bool, error) {
	ids, err := s.q.RegisteredTrialIDsForDog(ctx, dogID)
	if err != nil {
		return nil, err
	}
	set := make(map[int64]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set, nil
}

// CreateRegistration files one pending registration. The caller enforces
// dog ownership and trial eligibility; the (trial_id, dog_id) UNIQUE
// constraint is the final guard against a duplicate.
func (s *Store) CreateRegistration(ctx context.Context, trialID, competitorID, dogID, submittedBy int64, notes string) (db.Registration, error) {
	return s.q.CreateRegistration(ctx, db.CreateRegistrationParams{
		TrialID:      trialID,
		CompetitorID: competitorID,
		DogID:        dogID,
		SubmittedBy:  submittedBy,
		Notes:        notes,
	})
}

// ListPendingRegistrations returns a competitor's not-yet-accepted
// registrations (pending or waitlisted), joined to trial, event, and dog.
func (s *Store) ListPendingRegistrations(ctx context.Context, competitorID int64) ([]db.ListPendingRegistrationsByCompetitorRow, error) {
	return s.q.ListPendingRegistrationsByCompetitor(ctx, competitorID)
}

// ListFiledChallenges returns every challenge a competitor has filed,
// newest first, each joined to the disputed entry's trial and event.
// Backs the challenges list (A7).
func (s *Store) ListFiledChallenges(ctx context.Context, competitorID int64) ([]db.ListChallengesByFilerRow, error) {
	return s.q.ListChallengesByFiler(ctx, competitorID)
}

// ChallengeForEntry returns the most recent challenge a competitor filed
// against one entry, or sql.ErrNoRows when none exists. Used to stop a
// duplicate filing and to label an already-challenged entry.
func (s *Store) ChallengeForEntry(ctx context.Context, entryID, filedBy int64) (db.Challenge, error) {
	return s.q.GetChallengeForEntryByFiler(ctx, db.GetChallengeForEntryByFilerParams{
		EntryID: entryID,
		FiledBy: filedBy,
	})
}

// FileChallenge records a competitor's dispute of a finalized entry and
// returns the new row. The handler enforces ownership, finalized status,
// and no-duplicate before calling.
func (s *Store) FileChallenge(ctx context.Context, entryID, filedBy int64, reason string) (db.Challenge, error) {
	return s.q.CreateChallenge(ctx, db.CreateChallengeParams{
		EntryID: entryID,
		FiledBy: filedBy,
		Reason:  reason,
	})
}

// DogListItem is one dog on the account dogs list (A3) with the derived
// activity hint. LastCompeted is nil when the dog has no finalized entry;
// EntryCount is the total entries (any status) recorded against it.
type DogListItem struct {
	Dog          db.Dog
	EntryCount   int64
	LastCompeted *time.Time
}

// ListOwnerDogs loads a competitor's dogs with each dog's activity hint.
// The per-dog counts are an N+1 bounded by roster size (a handful of dogs
// per competitor), acceptable for this page.
func (s *Store) ListOwnerDogs(ctx context.Context, ownerID int64) ([]DogListItem, error) {
	dogs, err := s.q.ListDogsByOwner(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("list dogs for owner %d: %w", ownerID, err)
	}
	items := make([]DogListItem, 0, len(dogs))
	for _, d := range dogs {
		dogID := sql.NullInt64{Int64: d.ID, Valid: true}
		count, err := s.q.CountEntriesByDog(ctx, dogID)
		if err != nil {
			return nil, fmt.Errorf("count entries for dog %d: %w", d.ID, err)
		}
		item := DogListItem{Dog: d, EntryCount: count}
		last, err := s.q.LastCompetedByDog(ctx, dogID)
		if err == nil {
			item.LastCompeted = &last
		} else if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("last competed for dog %d: %w", d.ID, err)
		}
		items = append(items, item)
	}
	return items, nil
}

// GetOwnerDog fetches one dog scoped to its owner. Returns sql.ErrNoRows
// when the id misses or belongs to another competitor so the edit form
// renders 404 rather than leaking another owner's dog.
func (s *Store) GetOwnerDog(ctx context.Context, dogID, ownerID int64) (db.Dog, error) {
	return s.q.GetDogForOwner(ctx, db.GetDogForOwnerParams{ID: dogID, OwnerID: ownerID})
}

// DogInput carries the writable dog fields from the form. CallName is the
// only required value; the rest may be empty, and DateOfBirth is nil when
// unset.
type DogInput struct {
	CallName           string
	RegisteredName     string
	Breed              string
	DateOfBirth        *time.Time
	RegistrationNumber string
	Sex                string // "male", "female", or "" when unrecorded
}

// CreateDog inserts a dog under an owner and returns the new row.
func (s *Store) CreateDog(ctx context.Context, ownerID int64, in DogInput) (db.Dog, error) {
	return s.q.CreateDog(ctx, db.CreateDogParams{
		OwnerID:            ownerID,
		CallName:           in.CallName,
		RegisteredName:     in.RegisteredName,
		Breed:              in.Breed,
		DateOfBirth:        nullTime(in.DateOfBirth),
		RegistrationNumber: in.RegistrationNumber,
		Sex:                in.Sex,
	})
}

// UpdateDog saves an edited dog, scoped by owner so a competitor cannot
// edit a dog they do not own even with a guessed id.
func (s *Store) UpdateDog(ctx context.Context, dogID, ownerID int64, in DogInput) error {
	return s.q.UpdateDog(ctx, db.UpdateDogParams{
		CallName:           in.CallName,
		RegisteredName:     in.RegisteredName,
		Breed:              in.Breed,
		DateOfBirth:        nullTime(in.DateOfBirth),
		RegistrationNumber: in.RegistrationNumber,
		Sex:                in.Sex,
		ID:                 dogID,
		OwnerID:            ownerID,
	})
}

// DeleteDog removes a dog, scoped by owner.
func (s *Store) DeleteDog(ctx context.Context, dogID, ownerID int64) error {
	return s.q.DeleteDog(ctx, db.DeleteDogParams{ID: dogID, OwnerID: ownerID})
}

// nullTime wraps an optional time into the sql.NullTime sqlc expects.
func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
