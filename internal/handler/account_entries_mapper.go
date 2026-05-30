package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/flintcraftstudio/k9-trials/internal/db"
	"github.com/flintcraftstudio/k9-trials/internal/store"
	"github.com/flintcraftstudio/k9-trials/internal/view/account"
)

// challengeWindow is how long after finalization a competitor may dispute
// a score (§ wireframe A6 — "within 7 days").
const challengeWindow = 7 * 24 * time.Hour

// challengeWindowOpen reports whether the dispute window is still open for
// an entry finalized at the given time.
func challengeWindowOpen(finalizedAt time.Time) bool {
	return time.Since(finalizedAt) <= challengeWindow
}

// entryGroup maps a DB entry status to the A5 filter group key.
func entryGroup(status string) string {
	switch status {
	case "scoring":
		return "scoring"
	case "finalized":
		return "finalized"
	default: // registered / anything pre-scoring
		return "upcoming"
	}
}

// toEntriesListVD assembles the A5 list, counting each status group for the
// filter chips and rendering only the rows matching the active filter.
// Finalized rows are evaluated for their score and Q/NQ pill.
func toEntriesListVD(r *http.Request, st *store.Store, entries []db.ListEntriesByHandlerRow, active string) account.EntriesListViewData {
	var upcoming, scoring, finalized int
	for _, e := range entries {
		switch entryGroup(e.Status) {
		case "upcoming":
			upcoming++
		case "scoring":
			scoring++
		case "finalized":
			finalized++
		}
	}

	rows := make([]account.EntryRow, 0, len(entries))
	for _, e := range entries {
		group := entryGroup(e.Status)
		if active != "" && group != active {
			continue
		}
		rows = append(rows, entryRowVD(r, st, e, group))
	}

	return account.EntriesListViewData{
		Total:   len(entries),
		Filters: entryFilters(active, len(entries), upcoming, scoring, finalized),
		Rows:    rows,
	}
}

// entryFilters builds the status chip row with per-group counts.
func entryFilters(active string, total, upcoming, scoring, finalized int) []account.EntryFilter {
	defs := []struct {
		key, label string
		count      int
	}{
		{"", "All", total},
		{"upcoming", "Upcoming", upcoming},
		{"scoring", "In progress", scoring},
		{"finalized", "Finalized", finalized},
	}
	out := make([]account.EntryFilter, 0, len(defs))
	for _, d := range defs {
		href := "/account/entries"
		if d.key != "" {
			href += "?status=" + d.key
		}
		out = append(out, account.EntryFilter{
			Key:    d.key,
			Label:  d.label,
			Count:  d.count,
			Href:   href,
			Active: active == d.key,
		})
	}
	return out
}

// entryRowVD builds one A5 row, evaluating the score for finalized entries.
func entryRowVD(r *http.Request, st *store.Store, e db.ListEntriesByHandlerRow, group string) account.EntryRow {
	row := account.EntryRow{
		EntryID:  e.ID,
		Title:    e.EventName + " · " + disciplineLevelLabel(e.Discipline, e.Level),
		EventKey: disciplineKey(e.Discipline),
	}
	sub := e.DogName + " · " + entryNumberLabel(e.EntryNumber) + " · " + shortDate(e.TrialDate)
	switch group {
	case "scoring":
		row.StatusLabel, row.StatusKind = "Scoring", "scoring"
	case "finalized":
		pts, passed, ok := evalFinalizedScore(r, st, e.Discipline, e.Level, e.TemplateVersion, e.ID)
		if passed {
			row.StatusLabel, row.StatusKind = "Finalized · Q", "qual"
		} else {
			row.StatusLabel, row.StatusKind = "Finalized · NQ", "closed"
		}
		if ok {
			sub += fmt.Sprintf(" · scored %d", pts)
		}
	default:
		row.StatusLabel, row.StatusKind = "Upcoming", "wait"
	}
	row.Sub = sub
	return row
}

// toEntryDetailVD assembles the A6 page. For finalized entries it runs the
// scoring engine and flattens the per-exercise results, then layers on the
// challenge affordance (open window, already-filed, or window-closed). For
// other statuses it leaves the score payload empty for the neutral state.
func toEntryDetailVD(r *http.Request, st *store.Store, entry db.Entry, trial db.Trial, event db.Event, existing *db.Challenge) account.EntryDetailViewData {
	d := account.EntryDetailViewData{
		EntryID:   entry.ID,
		Eyebrow:   disciplineLevelLabel(trial.Discipline, trial.Level) + " · " + entryNumberLabel(entry.EntryNumber),
		EventName: event.Name,
		EventKey:  disciplineKey(trial.Discipline),
		DogMeta:   ownedEntryMeta(entry, trial),
	}

	switch entry.Status {
	case "finalized":
		d.Finalized = true
		tpl, sheet, inputs, result, err := loadTemplateAndEvaluate(r, st, trial, entry.ID)
		if err != nil {
			// Template/eval failure: fall back to the neutral notice rather
			// than 500, mirroring the public entry page.
			d.Finalized = false
			d.Pending = true
			return d
		}
		flat := flattenExercises(sheet, result, inputs)
		lines := make([]account.ExerciseLine, 0, len(flat))
		for _, fx := range flat {
			lines = append(lines, account.ExerciseLine{
				Num:   fx.Num,
				Name:  fx.Name,
				Score: int(fx.Result.Points),
				Max:   int(fx.MaxPts),
			})
		}
		d.Points = int(result.TotalPoints)
		d.MaxPoints = int(result.MaxPoints)
		d.Passed = result.Passed
		d.Threshold = int(qualifyingThreshold(tpl, sheet))
		d.Exercises = lines
		d.JudgedBy = judgeNameForEntry(entry)
		d.FinalizedDate = fullDate(entry.UpdatedAt)

		switch {
		case existing != nil:
			d.AlreadyFiled = true
			d.ChallengeStatus = existing.Status
		case challengeWindowOpen(entry.UpdatedAt):
			d.CanChallenge = true
			d.ChallengeHref = fmt.Sprintf("/account/challenges/new?entry=%d", entry.ID)
		default:
			d.WindowClosed = true
		}
	case "scoring":
		d.Scoring = true
	default:
		d.Pending = true
	}
	return d
}

// ownedEntryMeta renders the entry sub-line for the owner's own view:
// dog, handler of record, and trial date.
func ownedEntryMeta(entry db.Entry, trial db.Trial) string {
	parts := []string{entry.DogName}
	if entry.HandlerName != "" {
		parts = append(parts, "handled by "+entry.HandlerName)
	}
	parts = append(parts, fullDate(trial.TrialDate))
	out := parts[0]
	for _, p := range parts[1:] {
		out += " · " + p
	}
	return out
}
