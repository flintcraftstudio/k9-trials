package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/a-h/templ"

	"github.com/flintcraftstudio/k9-trials/internal/db"
	"github.com/flintcraftstudio/k9-trials/internal/scoring"
	"github.com/flintcraftstudio/k9-trials/internal/scoring/templates"
	"github.com/flintcraftstudio/k9-trials/internal/session"
	"github.com/flintcraftstudio/k9-trials/internal/store"
	"github.com/flintcraftstudio/k9-trials/internal/view/judge"
)

func renderJudge(w http.ResponseWriter, r *http.Request, c templ.Component) {
	if err := c.Render(r.Context(), w); err != nil {
		slog.Error("render error", "err", err)
	}
}

// loadTemplateAndEvaluate is shared by every per-entry handler: it
// resolves the trial's template, builds a concrete sheet, loads logged
// inputs, and runs the engine. Returns sheet + result so callers can map
// straight into view structs without re-deriving anything.
func loadTemplateAndEvaluate(
	r *http.Request,
	st *store.Store,
	trial db.Trial,
	entryID int64,
) (scoring.ScoresheetTemplate, scoring.ConcreteScoresheet, scoring.ScoresheetInputs, scoring.ScoresheetResult, error) {
	tpl, ok := templates.Lookup(
		scoring.Discipline(trial.Discipline),
		scoring.Level(trial.Level),
		trial.TemplateVersion,
	)
	if !ok {
		return scoring.ScoresheetTemplate{}, scoring.ConcreteScoresheet{}, scoring.ScoresheetInputs{}, scoring.ScoresheetResult{},
			fmt.Errorf("no template registered for %s/L%d/%s", trial.Discipline, trial.Level, trial.TemplateVersion)
	}
	sheet, err := tpl.BuildConcrete(nil)
	if err != nil {
		return tpl, scoring.ConcreteScoresheet{}, scoring.ScoresheetInputs{}, scoring.ScoresheetResult{}, fmt.Errorf("build concrete: %w", err)
	}
	inputs, err := st.LoadInputsForEntry(r.Context(), entryID)
	if err != nil {
		return tpl, sheet, scoring.ScoresheetInputs{}, scoring.ScoresheetResult{}, fmt.Errorf("load inputs: %w", err)
	}
	result, err := scoring.EvaluateScoresheet(inputs, sheet, tpl)
	if err != nil {
		return tpl, sheet, inputs, scoring.ScoresheetResult{}, fmt.Errorf("evaluate: %w", err)
	}
	return tpl, sheet, inputs, result, nil
}

// parseEntryID pulls and validates the {id} path segment.
func parseEntryID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// notFound is the canonical 404 for missing entries. Logs at info so
// scrapers don't fill the error stream.
func notFound(w http.ResponseWriter, r *http.Request, err error) {
	slog.Info("judge 404", "path", r.URL.Path, "err", err)
	http.NotFound(w, r)
}

// entryAssignedTo reports whether u is authorized to act on entry at the row
// level: the user must be the entry's assigned judge (entry.judge_id == u.ID),
// with admins exempt as a superset (an admin may score/finalize any entry).
// RequireJudge already gated the surface; this gates the specific row. A nil
// user or an entry with no assigned judge is never authorized (except admins).
func entryAssignedTo(entry db.Entry, u *session.User) bool {
	if u == nil {
		return false
	}
	if u.IsAdmin() {
		return true
	}
	return entry.JudgeID.Valid && entry.JudgeID.Int64 == u.ID
}

// guardEntryAuthority enforces the per-entry scoring authority rule. It returns
// true when the request may proceed; otherwise it has already written a 403 (or
// a redirect for an anonymous caller) and the handler must return. Centralizes
// the row-level check shared by every per-entry judge handler.
func guardEntryAuthority(w http.ResponseWriter, r *http.Request, entry db.Entry) bool {
	u := session.FromContext(r.Context())
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return false
	}
	if entryAssignedTo(entry, u) {
		return true
	}
	slog.Info("judge entry authority denied",
		"path", r.URL.Path, "entry", entry.ID, "user", u.ID)
	http.Error(w, "You are not the assigned judge for this entry.", http.StatusForbidden)
	return false
}

// JudgeQueue renders the run queue (B1).
func JudgeQueue(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := session.FromContext(r.Context())
		if u == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		queue, err := st.ListJudgeQueue(r.Context(), u.ID)
		if err != nil {
			slog.Error("judge queue load", "err", err)
			http.Error(w, "queue unavailable", http.StatusInternalServerError)
			return
		}

		data := judge.QueueViewData{}
		if queue.Trial == nil {
			// Judge has no assigned entries — render an empty trial so the
			// chrome (event name, judge name) still appears.
			data.Trial = judge.Trial{
				Name:       "No assignments yet",
				Class:      "",
				Discipline: judge.DiscOB,
				JudgeName:  judgeName(u.Email),
				JudgeInits: judgeInitials(u.Email),
			}
			renderJudge(w, r, judge.QueuePage(data))
			return
		}

		tpl, sheet, ok := lookupTemplateForTrial(*queue.Trial)
		// scoring totals only used to populate per-row score labels; if
		// no template is registered we still render the queue with empty
		// score columns rather than 500.
		runs := make([]judge.Run, 0, len(queue.Entries))
		for _, entry := range queue.Entries {
			label := ""
			if ok && entry.Status == "finalized" {
				inputs, err := st.LoadInputsForEntry(r.Context(), entry.ID)
				if err == nil {
					if res, err := scoring.EvaluateScoresheet(inputs, sheet, tpl); err == nil {
						label = scoreLabel(entry, res.TotalPoints)
					}
				}
			}
			runs = append(runs, toRun(entry, label))
		}

		data.Trial = toTrial(*queue.Trial, *queue.Event, u.Email)
		data.Runs = runs
		data.Counts = judge.Tally(runs)
		renderJudge(w, r, judge.QueuePage(data))
	}
}

// lookupTemplateForTrial bundles the template + concrete-sheet build.
// Returns ok=false when no template is registered for the trial.
func lookupTemplateForTrial(trial db.Trial) (scoring.ScoresheetTemplate, scoring.ConcreteScoresheet, bool) {
	tpl, ok := templates.Lookup(
		scoring.Discipline(trial.Discipline),
		scoring.Level(trial.Level),
		trial.TemplateVersion,
	)
	if !ok {
		return scoring.ScoresheetTemplate{}, scoring.ConcreteScoresheet{}, false
	}
	sheet, err := tpl.BuildConcrete(nil)
	if err != nil {
		return tpl, scoring.ConcreteScoresheet{}, false
	}
	return tpl, sheet, true
}

// JudgeGate renders the identity gate (B2) for a given entry.
func JudgeGate(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := session.FromContext(r.Context())
		entryID, ok := parseEntryID(r)
		if !ok {
			http.NotFound(w, r)
			return
		}
		entry, trial, event, err := st.LoadEntryWithTrial(r.Context(), entryID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				notFound(w, r, err)
				return
			}
			slog.Error("judge gate load", "err", err)
			http.Error(w, "gate unavailable", http.StatusInternalServerError)
			return
		}
		if !guardEntryAuthority(w, r, entry) {
			return
		}
		data := judge.GateViewData{
			Trial: toTrial(trial, event, u.Email),
			Run:   toRun(entry, ""),
		}
		renderJudge(w, r, judge.GatePage(data))
	}
}

// JudgeScore renders the active scoresheet (B3-O or B3-D, picked by
// the trial's discipline).
func JudgeScore(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := session.FromContext(r.Context())
		entryID, ok := parseEntryID(r)
		if !ok {
			http.NotFound(w, r)
			return
		}
		entry, trial, event, err := st.LoadEntryWithTrial(r.Context(), entryID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				notFound(w, r, err)
				return
			}
			slog.Error("judge score load", "err", err)
			http.Error(w, "scoresheet unavailable", http.StatusInternalServerError)
			return
		}
		if !guardEntryAuthority(w, r, entry) {
			return
		}

		tpl, sheet, inputs, result, err := loadTemplateAndEvaluate(r, st, trial, entryID)
		if err != nil {
			slog.Error("judge score evaluate", "err", err)
			http.Error(w, "scoresheet unavailable", http.StatusInternalServerError)
			return
		}

		flat := flattenExercises(sheet, result, inputs)
		exercises, activeIdx := toExercises(flat)

		data := judge.ScoreViewData{
			Trial:      toTrial(trial, event, u.Email),
			Run:        toRun(entry, ""),
			Discipline: judge.Discipline(trial.Discipline),
			Exercises:  exercises,
			ActiveIdx:  activeIdx,
			Score:      float64(result.TotalPoints),
			ScoreMax:   float64(sheet.MaxPoints),
			NeedToPass: qualifyingThreshold(tpl, sheet),
		}

		if data.Discipline == judge.DiscDT {
			renderJudge(w, r, judge.DetectionScorePage(data))
			return
		}
		renderJudge(w, r, judge.ObedienceScorePage(data))
	}
}

// JudgeReview renders run review (B4).
func JudgeReview(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := session.FromContext(r.Context())
		entryID, ok := parseEntryID(r)
		if !ok {
			http.NotFound(w, r)
			return
		}
		entry, trial, event, err := st.LoadEntryWithTrial(r.Context(), entryID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				notFound(w, r, err)
				return
			}
			slog.Error("judge review load", "err", err)
			http.Error(w, "review unavailable", http.StatusInternalServerError)
			return
		}
		if !guardEntryAuthority(w, r, entry) {
			return
		}
		tpl, sheet, inputs, result, err := loadTemplateAndEvaluate(r, st, trial, entryID)
		if err != nil {
			slog.Error("judge review evaluate", "err", err)
			http.Error(w, "review unavailable", http.StatusInternalServerError)
			return
		}
		flat := flattenExercises(sheet, result, inputs)
		unscoredCount, unscoredEx := unscoredSummary(flat)

		data := judge.ReviewViewData{
			Trial:            toTrial(trial, event, u.Email),
			Run:              toRun(entry, ""),
			Exercises:        toReviewExercises(flat),
			Provisional:      float64(result.TotalPoints),
			Max:              float64(sheet.MaxPoints),
			Qualifying:       qualifyingThreshold(tpl, sheet),
			UnscoredCount:    unscoredCount,
			UnscoredExercise: unscoredEx,
		}
		renderJudge(w, r, judge.ReviewPage(data))
	}
}

// JudgeSubmit renders submit confirmation (B5).
func JudgeSubmit(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := session.FromContext(r.Context())
		entryID, ok := parseEntryID(r)
		if !ok {
			http.NotFound(w, r)
			return
		}
		entry, trial, event, err := st.LoadEntryWithTrial(r.Context(), entryID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				notFound(w, r, err)
				return
			}
			slog.Error("judge submit load", "err", err)
			http.Error(w, "submit unavailable", http.StatusInternalServerError)
			return
		}
		if !guardEntryAuthority(w, r, entry) {
			return
		}
		_, sheet, inputs, result, err := loadTemplateAndEvaluate(r, st, trial, entryID)
		if err != nil {
			slog.Error("judge submit evaluate", "err", err)
			http.Error(w, "submit unavailable", http.StatusInternalServerError)
			return
		}
		flat := flattenExercises(sheet, result, inputs)
		scored := 0
		for _, fx := range flat {
			if fx.HasInput {
				scored++
			}
		}

		data := judge.SubmitViewData{
			Trial:       toTrial(trial, event, u.Email),
			Run:         toRun(entry, ""),
			Total:       float64(result.TotalPoints),
			Qualifying:  result.Passed,
			ExercisesOK: fmt.Sprintf("%d/%d exercises", scored, len(flat)),
			DedSummary:  "",
		}
		renderJudge(w, r, judge.SubmitPage(data))
	}
}

// JudgeLocked renders the read-only locked run (B6).
func JudgeLocked(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := session.FromContext(r.Context())
		entryID, ok := parseEntryID(r)
		if !ok {
			http.NotFound(w, r)
			return
		}
		entry, trial, event, err := st.LoadEntryWithTrial(r.Context(), entryID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				notFound(w, r, err)
				return
			}
			slog.Error("judge locked load", "err", err)
			http.Error(w, "locked view unavailable", http.StatusInternalServerError)
			return
		}
		if !guardEntryAuthority(w, r, entry) {
			return
		}
		tpl, sheet, inputs, result, err := loadTemplateAndEvaluate(r, st, trial, entryID)
		if err != nil {
			slog.Error("judge locked evaluate", "err", err)
			http.Error(w, "locked view unavailable", http.StatusInternalServerError)
			return
		}
		flat := flattenExercises(sheet, result, inputs)

		data := judge.LockedViewData{
			Trial:      toTrial(trial, event, u.Email),
			Run:        toRun(entry, scoreLabel(entry, result.TotalPoints)),
			Total:      float64(result.TotalPoints),
			Max:        float64(sheet.MaxPoints),
			Qualifying: qualifyingThreshold(tpl, sheet),
			Exercises:  toReviewExercises(flat),
		}
		data.SubmittedBy.Name = judgeName(u.Email)
		data.SubmittedBy.Inits = judgeInitials(u.Email)
		data.SubmittedBy.At = entry.UpdatedAt.Format("15.04")
		renderJudge(w, r, judge.LockedPage(data))
	}
}
