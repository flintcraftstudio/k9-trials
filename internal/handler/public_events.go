package handler

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"sort"
	"strconv"

	"github.com/a-h/templ"

	"github.com/flintcraftstudio/k9-trials/internal/db"
	"github.com/flintcraftstudio/k9-trials/internal/scoring"
	"github.com/flintcraftstudio/k9-trials/internal/session"
	"github.com/flintcraftstudio/k9-trials/internal/store"
	"github.com/flintcraftstudio/k9-trials/internal/view/entries"
	"github.com/flintcraftstudio/k9-trials/internal/view/events"
)

// renderPublic renders a templ component for the public surface, logging
// any render error without leaking it to the client.
func renderPublic(w http.ResponseWriter, r *http.Request, c templ.Component) {
	if err := c.Render(r.Context(), w); err != nil {
		slog.Error("render error", "path", r.URL.Path, "err", err)
	}
}

// disciplineOrder is the canonical display order for discipline chips and
// the filter row. Keeps OB/PR/TR/DT stable regardless of trial insertion
// order.
var disciplineOrder = []string{"OB", "PR", "TR", "DT"}

// EventsList serves GET /events — the public events index (P1). Supports a
// ?discipline= filter (OB/PR/TR/DT); any other value falls back to "all".
// htmx requests (the filter chips) receive only the event-grid fragment;
// full navigations receive the page shell.
func EventsList(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("discipline")
		if !validDiscipline(filter) {
			filter = "" // "all"
		}

		all, err := st.ListPublicEvents(r.Context())
		if err != nil {
			slog.Error("events list load", "err", err)
			http.Error(w, "events unavailable", http.StatusInternalServerError)
			return
		}

		cards := make([]events.EventCard, 0, len(all))
		for _, ewt := range all {
			if filter != "" && !eventHasDiscipline(ewt, filter) {
				continue
			}
			cards = append(cards, toEventCard(ewt))
		}

		data := events.ListViewData{
			Count:   len(cards),
			Filters: disciplineFilters(filter),
			Events:  cards,
		}

		if r.Header.Get("HX-Request") == "true" {
			renderPublic(w, r, events.EventGridFragment(data))
			return
		}
		renderPublic(w, r, events.ListPage(data))
	}
}

// validDiscipline reports whether code is one of the four known
// discipline codes.
func validDiscipline(code string) bool {
	switch code {
	case "OB", "PR", "TR", "DT":
		return true
	}
	return false
}

// eventHasDiscipline reports whether any trial in the event uses the given
// discipline code.
func eventHasDiscipline(ewt store.EventWithTrials, code string) bool {
	for _, t := range ewt.Trials {
		if t.Discipline == code {
			return true
		}
	}
	return false
}

// eventDisciplines returns the unique discipline codes present in an
// event's trials, in canonical OB/PR/TR/DT order.
func eventDisciplines(ewt store.EventWithTrials) []string {
	seen := make(map[string]bool)
	for _, t := range ewt.Trials {
		seen[t.Discipline] = true
	}
	out := make([]string, 0, len(seen))
	for _, code := range disciplineOrder {
		if seen[code] {
			out = append(out, code)
		}
	}
	return out
}

// toEventCard maps an event + its trials into the index card view struct.
func toEventCard(ewt store.EventWithTrials) events.EventCard {
	codes := eventDisciplines(ewt)
	tags := make([]events.DisciplineTag, 0, len(codes))
	for _, c := range codes {
		tags = append(tags, events.DisciplineTag{Label: disciplineLabel(c), Key: disciplineKey(c)})
	}
	eventKey := "obedience"
	if len(codes) > 0 {
		eventKey = disciplineKey(codes[0])
	}
	return events.EventCard{
		Slug:        ewt.Event.Slug,
		Name:        ewt.Event.Name,
		Location:    ewt.Event.Location,
		DateRange:   dateRange(ewt.Event.StartDate, ewt.Event.EndDate),
		TrialCount:  len(ewt.Trials),
		Disciplines: tags,
		RegOpen:     registrationOpen(ewt.Event.Status),
		EventKey:    eventKey,
	}
}

// disciplineFilters builds the filter chip row, marking the active chip.
// active is "" for the All chip or a discipline code.
func disciplineFilters(active string) []events.DisciplineFilter {
	filters := []events.DisciplineFilter{
		{Code: "", Label: "All", Href: "/events", Active: active == ""},
	}
	for _, c := range disciplineOrder {
		filters = append(filters, events.DisciplineFilter{
			Code:   c,
			Label:  disciplineLabel(c),
			Href:   "/events?discipline=" + c,
			Active: active == c,
		})
	}
	return filters
}

// EventDetail serves GET /events/{slug} — public event detail (P2). Draft
// events are treated as not-found for the public; only published/closed
// events render. 404 on an unknown or draft slug.
func EventDetail(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		ewt, err := st.LoadPublicEvent(r.Context(), slug)
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			slog.Error("event detail load", "slug", slug, "err", err)
			http.Error(w, "event unavailable", http.StatusInternalServerError)
			return
		}
		if ewt.Event.Status == "draft" || ewt.Event.Status == "archived" {
			http.NotFound(w, r)
			return
		}
		renderPublic(w, r, events.DetailPage(toDetailViewData(ewt, session.FromContext(r.Context()) != nil)))
	}
}

// toDetailViewData maps an event + trials into the detail view struct.
// Trials render grouped by date in template order (ListTrialsByEvent
// already sorts by trial_date, discipline, level).
func toDetailViewData(ewt store.EventWithTrials, loggedIn bool) events.DetailViewData {
	codes := eventDisciplines(ewt)
	tags := make([]events.DisciplineTag, 0, len(codes))
	for _, c := range codes {
		tags = append(tags, events.DisciplineTag{Label: disciplineLabel(c), Key: disciplineKey(c)})
	}
	eventKey := "obedience"
	if len(codes) > 0 {
		eventKey = disciplineKey(codes[0])
	}
	rows := make([]events.TrialRow, 0, len(ewt.Trials))
	for _, t := range ewt.Trials {
		rows = append(rows, events.TrialRow{
			ID:            t.ID,
			DisciplineLvl: disciplineLevelLabel(t.Discipline, t.Level),
			EventKey:      disciplineKey(t.Discipline),
			Date:          t.TrialDate.UTC().Format("Mon 2 Jan"),
		})
	}
	return events.DetailViewData{
		Slug:        ewt.Event.Slug,
		Name:        ewt.Event.Name,
		Location:    ewt.Event.Location,
		DateRange:   dateRange(ewt.Event.StartDate, ewt.Event.EndDate),
		Disciplines: tags,
		RegOpen:     registrationOpen(ewt.Event.Status),
		LoggedIn:    loggedIn,
		TrialCount:  len(ewt.Trials),
		Trials:      rows,
		EventKey:    eventKey,
	}
}

// TrialDetail serves GET /events/{slug}/trials/{id} — the public trial
// leaderboard (P3). Finalized entries are evaluated through the scoring
// engine to derive points + pass/fail; in-progress entries show a
// "scoring" pill with no partial points. 404 when the trial id misses or
// doesn't belong to the slug's event.
func TrialDetail(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		trialID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || trialID <= 0 {
			http.NotFound(w, r)
			return
		}
		trial, event, entries, err := st.LoadTrialWithEntries(r.Context(), trialID)
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			slog.Error("trial detail load", "trial", trialID, "err", err)
			http.Error(w, "trial unavailable", http.StatusInternalServerError)
			return
		}
		// Guard the slug/trial relationship so the URL is self-consistent.
		if event.Slug != slug || event.Status == "draft" || event.Status == "archived" {
			http.NotFound(w, r)
			return
		}
		renderPublic(w, r, events.TrialDetailPage(toTrialDetailViewData(r, st, trial, event, entries)))
	}
}

// toTrialDetailViewData evaluates each finalized entry and assembles the
// leaderboard. Qualifying finalized rows are ranked by points descending;
// NQ and in-progress rows carry no rank. Entries whose template can't be
// resolved are shown without a score rather than dropped.
func toTrialDetailViewData(r *http.Request, st *store.Store, trial db.Trial, event db.Event, entries []db.Entry) events.TrialDetailViewData {
	tpl, sheet, tplOK := lookupTemplateForTrial(trial)

	var finalized, scoring_, upcoming int
	rows := make([]events.LeaderRow, 0, len(entries))
	for _, e := range entries {
		row := events.LeaderRow{
			EntryID: e.ID,
			DogName: e.DogName,
			Handler: handlerShort(e.HandlerName),
			K9ID:    "",
		}
		switch e.Status {
		case "finalized":
			finalized++
			row.Finalized = true
			if tplOK {
				if inputs, err := st.LoadInputsForEntry(r.Context(), e.ID); err == nil {
					if res, err := scoring.EvaluateScoresheet(inputs, sheet, tpl); err == nil {
						row.Points = int(res.TotalPoints)
						row.Qualified = qualified(res)
						row.NQ = !res.Passed
					}
				}
			}
		case "scoring":
			scoring_++
			row.Scoring = true
		default: // registered / anything pre-scoring
			upcoming++
		}
		rows = append(rows, row)
	}

	// Rank: qualifying finalized rows by points desc, then assign 1..n.
	// NQ, scoring, and upcoming rows keep Rank 0 (rendered without a placing).
	ranked := make([]int, 0, len(rows))
	for i, row := range rows {
		if row.Finalized && row.Qualified {
			ranked = append(ranked, i)
		}
	}
	sort.SliceStable(ranked, func(a, b int) bool {
		return rows[ranked[a]].Points > rows[ranked[b]].Points
	})
	for placing, idx := range ranked {
		rows[idx].Rank = placing + 1
	}
	// Present qualifying rows first (in placing order), then the rest in
	// roster order so spectators see the standings up top.
	ordered := make([]events.LeaderRow, 0, len(rows))
	inRanked := make(map[int]bool, len(ranked))
	for _, idx := range ranked {
		ordered = append(ordered, rows[idx])
		inRanked[idx] = true
	}
	for i, row := range rows {
		if !inRanked[i] {
			ordered = append(ordered, row)
		}
	}

	return events.TrialDetailViewData{
		EventSlug:      event.Slug,
		EventName:      event.Name,
		DisciplineLvl:  disciplineLevelLabel(trial.Discipline, trial.Level),
		EventKey:       disciplineKey(trial.Discipline),
		Date:           fullDate(trial.TrialDate),
		TotalEntries:   len(entries),
		FinalizedCount: finalized,
		ScoringCount:   scoring_,
		UpcomingCount:  upcoming,
		Rows:           ordered,
	}
}

// EntryDetail serves GET /entries/{id} — the public read-only entry page
// (P4). Finalized entries render the full scoresheet breakdown; scoring
// and registered entries render a neutral state notice with no partial
// points. 404 when the id misses or the parent event isn't public.
func EntryDetail(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entryID, ok := parseEntryID(r)
		if !ok {
			http.NotFound(w, r)
			return
		}
		entry, trial, event, err := st.LoadEntryWithTrial(r.Context(), entryID)
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			slog.Error("entry detail load", "entry", entryID, "err", err)
			http.Error(w, "entry unavailable", http.StatusInternalServerError)
			return
		}
		if event.Status == "draft" || event.Status == "archived" {
			http.NotFound(w, r)
			return
		}
		renderPublic(w, r, entries.DetailPage(toEntryDetailViewData(r, st, entry, trial, event)))
	}
}

// toEntryDetailViewData assembles the public entry view. For finalized
// entries it runs the scoring engine and flattens the per-exercise
// results; for other statuses it leaves the score payload empty and lets
// the template render the appropriate notice.
func toEntryDetailViewData(r *http.Request, st *store.Store, entry db.Entry, trial db.Trial, event db.Event) entries.DetailViewData {
	d := entries.DetailViewData{
		Eyebrow:   disciplineLevelLabel(trial.Discipline, trial.Level) + " · " + entryNumberLabel(entry.EntryNumber),
		EventName: event.Name,
		EventSlug: event.Slug,
		TrialID:   trial.ID,
		EventKey:  disciplineKey(trial.Discipline),
		DogName:   entry.DogName,
		DogMeta:   entryDogMeta(entry),
	}

	switch entry.Status {
	case "finalized":
		d.Finalized = true
		tpl, sheet, inputs, result, err := loadTemplateAndEvaluate(r, st, trial, entry.ID)
		if err != nil {
			// Template/eval failure: fall back to the score-less notice
			// rather than 500, so a missing template registration doesn't
			// take down a public page.
			slog.Error("entry detail evaluate", "entry", entry.ID, "err", err)
			d.Finalized = false
			d.Pending = true
			return d
		}
		flat := flattenExercises(sheet, result, inputs)
		lines := make([]entries.ExerciseLine, 0, len(flat))
		for _, fx := range flat {
			lines = append(lines, entries.ExerciseLine{
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
	case "scoring":
		d.Scoring = true
	default:
		d.Pending = true
	}
	return d
}

// entryNumberLabel renders "Entry 14" / "Entry —" for unassigned numbers.
func entryNumberLabel(n int64) string {
	if n <= 0 {
		return "Entry —"
	}
	return "Entry " + strconv.FormatInt(n, 10)
}

// entryDogMeta renders the "Breed · handled by Handler" sub-line. The K9
// registration number lives on the dog row (not yet joined here), so it's
// omitted until the dog FK is wired through.
func entryDogMeta(entry db.Entry) string {
	parts := []string{}
	if entry.DogBreed != "" {
		parts = append(parts, entry.DogBreed)
	}
	if entry.HandlerName != "" {
		parts = append(parts, "handled by "+entry.HandlerName)
	}
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for _, p := range parts[1:] {
		out += " · " + p
	}
	return out
}

// judgeNameForEntry resolves a display name for the entry's judge. The
// entry stores only judge_id and no users join is wired into the public
// path yet, so this returns "" for now; judgedLine drops the "Judged by"
// clause when empty. Swap for a real lookup once users.display_name lands.
func judgeNameForEntry(entry db.Entry) string {
	return ""
}
