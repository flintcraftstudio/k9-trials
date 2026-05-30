package handler

import (
	"net/http"

	"github.com/flintcraftstudio/k9-trials/internal/db"
	"github.com/flintcraftstudio/k9-trials/internal/store"
	"github.com/flintcraftstudio/k9-trials/internal/view/account"
)

// toChallengesListVD maps the filed-challenge rows into the A7 view.
func toChallengesListVD(rows []db.ListChallengesByFilerRow) account.ChallengesListViewData {
	out := make([]account.ChallengeRow, 0, len(rows))
	for _, c := range rows {
		out = append(out, account.ChallengeRow{
			EntryID: c.EntryID,
			Title:   c.EventName + " · " + disciplineLevelLabel(c.Discipline, c.Level),
			Sub:     c.DogName + " · " + entryNumberLabel(c.EntryNumber) + " · " + shortDate(c.TrialDate),
			Filed:   "Filed " + relativeTime(c.FiledAt),
			Status:  c.Status,
		})
	}
	return account.ChallengesListViewData{Total: len(out), Rows: out}
}

// toChallengeNewVD builds the A8 form view: the disputed-entry summary plus
// the already-filed state when a prior challenge exists. The summary line
// includes the qualifying result so the competitor sees what they are
// disputing.
func toChallengeNewVD(r *http.Request, st *store.Store, entry db.Entry, trial db.Trial, event db.Event, existing *db.Challenge) account.ChallengeNewViewData {
	vd := account.ChallengeNewViewData{
		EntryID:      entry.ID,
		DisputeTitle: event.Name + " · " + disciplineLevelLabel(trial.Discipline, trial.Level) + " · " + entryNumberLabel(entry.EntryNumber),
		DisputeSub:   challengeDisputeSub(r, st, entry, trial),
		EventKey:     disciplineKey(trial.Discipline),
	}
	if existing != nil {
		vd.AlreadyFiled = true
		vd.ChallengeStatus = existing.Status
	}
	return vd
}

// challengeDisputeSub renders the "dog · date · result" sub-line for the
// disputed-entry card, appending the qualifying result when the score
// evaluates.
func challengeDisputeSub(r *http.Request, st *store.Store, entry db.Entry, trial db.Trial) string {
	sub := entry.DogName + " · " + fullDate(trial.TrialDate)
	if _, passed, ok := evalFinalizedScore(r, st, trial.Discipline, trial.Level, trial.TemplateVersion, entry.ID); ok {
		if passed {
			sub += " · result Q"
		} else {
			sub += " · result NQ"
		}
	}
	return sub
}
