package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/flintcraftstudio/k9-trials/internal/db"
)

// CompetitorCard is one competitor in the directory (P5) with the derived
// counts shown on the card. LastCompeted is nil when the competitor has no
// finalized entries yet.
type CompetitorCard struct {
	Competitor     db.Competitor
	DogCount       int64
	FinalizedCount int64
	LastCompeted   *time.Time
}

// ListCompetitorCards backs the directory (P5). With an empty search term
// it returns the most recent competitors; otherwise it substring-searches
// handle/display_name and dog names/registration numbers. Per-card counts
// are fetched individually — an N+1 bounded by limit, acceptable for the
// small result sets the directory renders.
func (s *Store) ListCompetitorCards(ctx context.Context, search string, limit int64) ([]CompetitorCard, error) {
	search = strings.TrimSpace(search)
	var rows []db.Competitor
	var err error
	if search == "" {
		rows, err = s.q.ListRecentCompetitors(ctx, limit)
	} else {
		rows, err = s.q.SearchCompetitors(ctx, db.SearchCompetitorsParams{
			Term: sql.NullString{String: search, Valid: true},
			Lim:  limit,
		})
	}
	if err != nil {
		return nil, fmt.Errorf("list competitors: %w", err)
	}

	cards := make([]CompetitorCard, 0, len(rows))
	for _, c := range rows {
		card, err := s.competitorCard(ctx, c)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	return cards, nil
}

// competitorCard layers the derived counts onto a competitor row.
func (s *Store) competitorCard(ctx context.Context, c db.Competitor) (CompetitorCard, error) {
	handlerID := sql.NullInt64{Int64: c.ID, Valid: true}
	dogCount, err := s.q.CountDogsByOwner(ctx, c.ID)
	if err != nil {
		return CompetitorCard{}, fmt.Errorf("count dogs for competitor %d: %w", c.ID, err)
	}
	finalized, err := s.q.CountFinalizedByHandler(ctx, handlerID)
	if err != nil {
		return CompetitorCard{}, fmt.Errorf("count finalized for competitor %d: %w", c.ID, err)
	}
	card := CompetitorCard{Competitor: c, DogCount: dogCount, FinalizedCount: finalized}
	last, err := s.q.LastCompetedByHandler(ctx, handlerID)
	if err == nil {
		card.LastCompeted = &last
	} else if !errors.Is(err, sql.ErrNoRows) {
		return CompetitorCard{}, fmt.Errorf("last competed for competitor %d: %w", c.ID, err)
	}
	return card, nil
}

// CompetitorProfile bundles everything the public profile page (P6) needs:
// the competitor, their dogs, and their finalized event history (raw rows;
// the handler runs the scoring engine for per-row points).
type CompetitorProfile struct {
	Competitor db.Competitor
	Dogs       []db.Dog
	History    []db.ListFinalizedEntriesByHandlerRow
}

// LoadCompetitorProfile resolves a competitor by handle plus their dogs and
// finalized history. Returns sql.ErrNoRows (from GetCompetitorByHandle)
// when the handle doesn't resolve so the handler renders 404.
func (s *Store) LoadCompetitorProfile(ctx context.Context, handle string) (CompetitorProfile, error) {
	c, err := s.q.GetCompetitorByHandle(ctx, handle)
	if err != nil {
		return CompetitorProfile{}, err
	}
	dogs, err := s.q.ListDogsByOwner(ctx, c.ID)
	if err != nil {
		return CompetitorProfile{}, fmt.Errorf("list dogs for competitor %d: %w", c.ID, err)
	}
	history, err := s.q.ListFinalizedEntriesByHandler(ctx, sql.NullInt64{Int64: c.ID, Valid: true})
	if err != nil {
		return CompetitorProfile{}, fmt.Errorf("list history for competitor %d: %w", c.ID, err)
	}
	return CompetitorProfile{Competitor: c, Dogs: dogs, History: history}, nil
}

// DogProfile bundles everything the public dog page (P7) needs: the dog,
// its owning competitor, and its finalized trial history (raw rows).
type DogProfile struct {
	Dog     db.Dog
	Owner   db.Competitor
	History []db.ListFinalizedEntriesByDogRow
}

// LoadDogProfile resolves a dog by id plus its owner and finalized history.
// Returns sql.ErrNoRows (from GetDogByID) when the id misses so the handler
// renders 404.
func (s *Store) LoadDogProfile(ctx context.Context, dogID int64) (DogProfile, error) {
	dog, err := s.q.GetDogByID(ctx, dogID)
	if err != nil {
		return DogProfile{}, err
	}
	owner, err := s.q.GetCompetitorByID(ctx, dog.OwnerID)
	if err != nil {
		return DogProfile{}, fmt.Errorf("get owner %d for dog %d: %w", dog.OwnerID, dogID, err)
	}
	history, err := s.q.ListFinalizedEntriesByDog(ctx, sql.NullInt64{Int64: dogID, Valid: true})
	if err != nil {
		return DogProfile{}, fmt.Errorf("list history for dog %d: %w", dogID, err)
	}
	return DogProfile{Dog: dog, Owner: owner, History: history}, nil
}
