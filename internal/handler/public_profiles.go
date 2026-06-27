package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flintcraftstudio/k9-trials/internal/scoring"
	"github.com/flintcraftstudio/k9-trials/internal/scoring/templates"
	"github.com/flintcraftstudio/k9-trials/internal/store"
	"github.com/flintcraftstudio/k9-trials/internal/view/competitors"
	"github.com/flintcraftstudio/k9-trials/internal/view/dogs"
)

// directoryLimit caps how many competitors the directory renders for both
// the recently-active list and search results.
const directoryLimit = 24

// CompetitorSearch serves GET /competitors — the public directory (P5).
// With no ?q= it lists recently-added competitors; with a term it
// substring-searches handles, names, and dog identifiers. htmx requests
// (the live search input) receive only the results fragment.
func CompetitorSearch(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		cards, err := st.ListCompetitorCards(r.Context(), q, directoryLimit)
		if err != nil {
			slog.Error("competitor directory load", "err", err)
			http.Error(w, "directory unavailable", http.StatusInternalServerError)
			return
		}
		data := competitors.SearchViewData{
			Query:    q,
			Searched: q != "",
			Cards:    toDirectoryCards(cards),
		}
		if r.Header.Get("HX-Request") == "true" {
			renderPublic(w, r, competitors.ResultsFragment(data))
			return
		}
		renderPublic(w, r, competitors.SearchPage(data))
	}
}

// toDirectoryCards maps store cards into the view structs, formatting the
// last-competed timestamp as a relative phrase.
func toDirectoryCards(cards []store.CompetitorCard) []competitors.DirectoryCard {
	out := make([]competitors.DirectoryCard, 0, len(cards))
	for _, c := range cards {
		last := ""
		if c.LastCompeted != nil {
			last = relativeTime(*c.LastCompeted)
		}
		out = append(out, competitors.DirectoryCard{
			Handle:         c.Competitor.Handle,
			DisplayName:    c.Competitor.DisplayName,
			DogCount:       int(c.DogCount),
			FinalizedCount: int(c.FinalizedCount),
			LastCompeted:   last,
		})
	}
	return out
}

// CompetitorProfile serves GET /competitors/{handle} — public profile
// (P6). 404 when the handle doesn't resolve.
func CompetitorProfile(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handle := r.PathValue("handle")
		prof, err := st.LoadCompetitorProfile(r.Context(), handle)
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			slog.Error("competitor profile load", "handle", handle, "err", err)
			http.Error(w, "profile unavailable", http.StatusInternalServerError)
			return
		}
		renderPublic(w, r, competitors.ProfilePage(toCompetitorProfileVD(r, st, prof)))
	}
}

// toCompetitorProfileVD builds the P6 view: dog roster + finalized history
// (each row evaluated for points). The history title leads with the dog so
// a handler running multiple dogs reads clearly.
func toCompetitorProfileVD(r *http.Request, st *store.Store, prof store.CompetitorProfile) competitors.ProfileViewData {
	dogTags := make([]competitors.DogTag, 0, len(prof.Dogs))
	for _, d := range prof.Dogs {
		dogTags = append(dogTags, competitors.DogTag{
			ID:       d.ID,
			CallName: d.CallName,
			Meta:     dogMetaLine(d.Breed, d.DateOfBirth),
			RegNo:    d.RegistrationNumber,
		})
	}

	history := make([]competitors.HistoryRow, 0, len(prof.History))
	for _, e := range prof.History {
		fs := evalFinalizedScore(r, st, e.Discipline, e.Level, e.TemplateVersion, e.ID)
		history = append(history, competitors.HistoryRow{
			EntryID:   e.ID,
			Title:     e.DogName + " · " + disciplineLevelLabel(e.Discipline, e.Level),
			Sub:       e.EventName + " · " + shortDate(e.TrialDate),
			Points:    fs.Points,
			Max:       fs.Max,
			Qualified: fs.Passed,
			HasScore:  fs.OK,
		})
	}

	return competitors.ProfileViewData{
		Handle:      prof.Competitor.Handle,
		DisplayName: prof.Competitor.DisplayName,
		Initials:    judgeInitials(prof.Competitor.DisplayName),
		Bio:         prof.Competitor.Bio,
		DogCount:    len(prof.Dogs),
		Dogs:        dogTags,
		History:     history,
	}
}

// DogProfile serves GET /dogs/{id} — public dog profile (P7). 404 when the
// id is malformed or misses.
func DogProfile(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id <= 0 {
			http.NotFound(w, r)
			return
		}
		prof, err := st.LoadDogProfile(r.Context(), id)
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			slog.Error("dog profile load", "dog", id, "err", err)
			http.Error(w, "profile unavailable", http.StatusInternalServerError)
			return
		}
		renderPublic(w, r, dogs.ProfilePage(toDogProfileVD(r, st, prof)))
	}
}

// toDogProfileVD builds the P7 view. History rows carry the handler of
// record (not always the owner) in the sub-line.
func toDogProfileVD(r *http.Request, st *store.Store, prof store.DogProfile) dogs.ProfileViewData {
	history := make([]dogs.HistoryRow, 0, len(prof.History))
	for _, e := range prof.History {
		fs := evalFinalizedScore(r, st, e.Discipline, e.Level, e.TemplateVersion, e.ID)
		sub := e.EventName
		if e.HandlerName != "" {
			sub += " · " + handlerShort(e.HandlerName)
		}
		sub += " · " + shortDate(e.TrialDate)
		history = append(history, dogs.HistoryRow{
			EntryID:   e.ID,
			Title:     disciplineLevelLabel(e.Discipline, e.Level),
			Sub:       sub,
			Points:    fs.Points,
			Max:       fs.Max,
			Qualified: fs.Passed,
			HasScore:  fs.OK,
		})
	}
	return dogs.ProfileViewData{
		CallName:       prof.Dog.CallName,
		RegisteredName: prof.Dog.RegisteredName,
		Breed:          prof.Dog.Breed,
		Age:            ageLabel(prof.Dog.DateOfBirth),
		RegNo:          prof.Dog.RegistrationNumber,
		OwnerHandle:    prof.Owner.Handle,
		OwnerName:      prof.Owner.DisplayName,
		History:        history,
	}
}

// finalizedScore is a finalized entry's evaluated result, flattened for
// list-row rendering. OK is false when no template is registered or
// evaluation fails; callers then render the row without a score.
type finalizedScore struct {
	Points  int
	Max     int
	Percent int // 100 * points / max, rounded
	Passed  bool
	OK      bool
}

// evalFinalizedScore looks up the trial's template, builds a concrete
// sheet, loads the entry's logged inputs, and runs the scoring engine.
// Returns OK=false (a zero score) when no template is registered or
// evaluation fails — callers render the row without a score rather than
// failing the whole page.
func evalFinalizedScore(r *http.Request, st *store.Store, discipline string, level int64, version string, entryID int64) finalizedScore {
	tpl, ok := templates.Lookup(scoring.Discipline(discipline), scoring.Level(level), version)
	if !ok {
		return finalizedScore{}
	}
	sheet, err := tpl.BuildConcrete(nil)
	if err != nil {
		return finalizedScore{}
	}
	inputs, err := st.LoadInputsForEntry(r.Context(), entryID)
	if err != nil {
		return finalizedScore{}
	}
	res, err := scoring.EvaluateScoresheet(inputs, sheet, tpl)
	if err != nil {
		return finalizedScore{}
	}
	return finalizedScore{
		Points:  int(res.TotalPoints),
		Max:     int(res.MaxPoints),
		Percent: int(math.Round(res.Percent)),
		Passed:  res.Passed,
		OK:      true,
	}
}

// dogMetaLine composes the "Breed · age" sub-line for a dog card, dropping
// whichever part is unknown.
func dogMetaLine(breed string, dob sql.NullTime) string {
	parts := []string{}
	if breed != "" {
		parts = append(parts, breed)
	}
	if age := ageLabel(dob); age != "" {
		parts = append(parts, age)
	}
	return strings.Join(parts, " · ")
}

// ageLabel renders a dog's age in whole years ("4y") from its date of
// birth, or "" when the DOB isn't recorded.
func ageLabel(dob sql.NullTime) string {
	if !dob.Valid {
		return ""
	}
	years := yearsSince(dob.Time, time.Now())
	if years < 0 {
		return ""
	}
	return fmt.Sprintf("%dy", years)
}

// yearsSince returns whole years between then and now, accounting for
// whether the anniversary has occurred this year.
func yearsSince(then, now time.Time) int {
	years := now.Year() - then.Year()
	anniversaryThisYear := time.Date(now.Year(), then.Month(), then.Day(), 0, 0, 0, 0, time.UTC)
	nowDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if nowDay.Before(anniversaryThisYear) {
		years--
	}
	return years
}

// relativeTime renders a coarse "X ago" phrase for a past timestamp. Used
// for the directory's last-competed hint, where precision below a day
// isn't useful.
func relativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < 24*time.Hour:
		return "today"
	case d < 48*time.Hour:
		return "yesterday"
	case d < 7*24*time.Hour:
		return agoUnits(int(d.Hours()/24), "day")
	case d < 30*24*time.Hour:
		return agoUnits(int(d.Hours()/(24*7)), "week")
	case d < 365*24*time.Hour:
		return agoUnits(int(d.Hours()/(24*30)), "month")
	default:
		return agoUnits(int(d.Hours()/(24*365)), "year")
	}
}

// agoUnits renders "1 week ago" / "3 weeks ago" with correct pluralization.
func agoUnits(n int, unit string) string {
	if n == 1 {
		return "1 " + unit + " ago"
	}
	return fmt.Sprintf("%d %ss ago", n, unit)
}
