package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"sort"
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

// listRow pairs a built view row with the data the list needs to count and
// order it: its filter group and its trial date.
type listRow struct {
	row   account.EntryRow
	group string
	date  time.Time
}

// toEntriesListVD assembles the A5 list from finalized/scoring/registered
// entries plus not-yet-accepted registrations. It counts each status group
// for the filter chips and renders only the rows matching the active
// filter, newest trial date first. Finalized rows are scored for their
// points and Q/NQ pill; pending registrations join the "upcoming" group.
func toEntriesListVD(r *http.Request, st *store.Store, entries []db.ListEntriesByHandlerRow, regs []db.ListPendingRegistrationsByCompetitorRow, active string) account.EntriesListViewData {
	all := make([]listRow, 0, len(entries)+len(regs))
	for _, e := range entries {
		group := entryGroup(e.Status)
		withdrawn := e.RegStatus.Valid && e.RegStatus.String == "withdrawn"
		requested := !withdrawn && e.RegStatus.Valid && e.RegStatus.String == "accepted" && e.WithdrawRequestedAt.Valid
		if withdrawn {
			group = "withdrawn"
		}
		all = append(all, listRow{row: entryRowVD(r, st, e, group, requested), group: group, date: e.TrialDate})
	}
	for _, rg := range regs {
		all = append(all, listRow{row: registrationRowVD(rg), group: "upcoming", date: rg.TrialDate})
	}

	// Newest trial date first; stable so same-date rows keep insertion order.
	sort.SliceStable(all, func(i, j int) bool { return all[i].date.After(all[j].date) })

	var upcoming, scoring, finalized, withdrawn int
	rows := make([]account.EntryRow, 0, len(all))
	for _, lr := range all {
		switch lr.group {
		case "upcoming":
			upcoming++
		case "scoring":
			scoring++
		case "finalized":
			finalized++
		case "withdrawn":
			withdrawn++
		}
		if active == "" || lr.group == active {
			rows = append(rows, lr.row)
		}
	}

	return account.EntriesListViewData{
		Total:   len(all),
		Filters: entryFilters(active, len(all), upcoming, scoring, finalized, withdrawn),
		Rows:    rows,
	}
}

// registrationRowVD builds an A5 row for a not-yet-accepted registration.
// It links to the public event since there is no entry detail until an
// admin accepts it.
func registrationRowVD(rg db.ListPendingRegistrationsByCompetitorRow) account.EntryRow {
	label, kind := "Pending", "wait"
	if rg.Status == "waitlisted" {
		label = "Waitlisted"
	}
	return account.EntryRow{
		Href:        "/events/" + rg.EventSlug,
		Title:       rg.EventName + " · " + disciplineLevelLabel(rg.Discipline, rg.Level),
		Sub:         rg.DogName + " · " + shortDate(rg.TrialDate) + " · pending review",
		EventKey:    disciplineKey(rg.Discipline),
		StatusLabel: label,
		StatusKind:  kind,
	}
}

// entryFilters builds the status chip row with per-group counts. The
// Withdrawn chip only appears once at least one entry is withdrawn, so the
// common case stays uncluttered.
func entryFilters(active string, total, upcoming, scoring, finalized, withdrawn int) []account.EntryFilter {
	defs := []struct {
		key, label string
		count      int
	}{
		{"", "All", total},
		{"upcoming", "Upcoming", upcoming},
		{"scoring", "In progress", scoring},
		{"finalized", "Finalized", finalized},
	}
	if withdrawn > 0 {
		defs = append(defs, struct {
			key, label string
			count      int
		}{"withdrawn", "Withdrawn", withdrawn})
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
// requested marks an upcoming entry with a pending withdrawal request.
func entryRowVD(r *http.Request, st *store.Store, e db.ListEntriesByHandlerRow, group string, requested bool) account.EntryRow {
	row := account.EntryRow{
		Href:     fmt.Sprintf("/account/entries/%d", e.ID),
		Title:    e.EventName + " · " + disciplineLevelLabel(e.Discipline, e.Level),
		EventKey: disciplineKey(e.Discipline),
	}
	sub := e.DogName + " · " + entryNumberLabel(e.EntryNumber) + " · " + shortDate(e.TrialDate)
	switch group {
	case "scoring":
		row.StatusLabel, row.StatusKind = "Scoring", "scoring"
	case "finalized":
		fs := evalFinalizedScore(r, st, e.Discipline, e.Level, e.TemplateVersion, e.ID)
		if fs.Passed {
			row.StatusLabel, row.StatusKind = "Finalized · Q", "qual"
		} else {
			row.StatusLabel, row.StatusKind = "Finalized · NQ", "closed"
		}
		row.Points, row.Max, row.HasScore = fs.Points, fs.Max, fs.OK
	case "withdrawn":
		row.StatusLabel, row.StatusKind = "Withdrawn", "closed"
	default:
		if requested {
			row.StatusLabel, row.StatusKind = "Withdrawal requested", "wait"
		} else {
			row.StatusLabel, row.StatusKind = "Upcoming", "wait"
		}
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

	// Withdrawal state from the linked registration (none for entries created
	// outside the registration flow). A withdrawn registration overrides the
	// entry-status branch below; a pending request or a withdrawable entry
	// surfaces an action in the not-yet-run state.
	if reg, ok, err := st.RegistrationForEntry(r.Context(), entry.ID); err != nil {
		slog.Error("entry registration lookup", "entry", entry.ID, "err", err)
	} else if ok {
		switch {
		case reg.Status == "withdrawn":
			d.Withdrawn = true
		case reg.Status == "accepted" && reg.WithdrawRequestedAt.Valid:
			d.WithdrawRequested = true
		case reg.Status == "accepted" && entry.Status == "registered":
			d.CanWithdraw = true
		}
	}

	if d.Withdrawn {
		return d
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
