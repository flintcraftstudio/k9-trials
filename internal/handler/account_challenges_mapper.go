package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/flintcraftstudio/k9-trials/internal/db"
	"github.com/flintcraftstudio/k9-trials/internal/scoring"
	"github.com/flintcraftstudio/k9-trials/internal/store"
	"github.com/flintcraftstudio/k9-trials/internal/view/account"
)

// toChallengesListVD maps the filed-challenge rows into the A7 view. It
// counts each status group for the filter chips, surfaces the most recent
// update in the header, and renders only the rows matching the active
// filter. The rows arrive newest-first from the query, so order is kept.
func toChallengesListVD(rows []db.ListChallengesByFilerRow, active string) account.ChallengesListViewData {
	var open, review, resolved, dismissed int
	var lastUpdate time.Time
	out := make([]account.ChallengeRow, 0, len(rows))
	for _, c := range rows {
		switch c.Status {
		case "open":
			open++
		case "under_review":
			review++
		case "resolved":
			resolved++
		case "dismissed":
			dismissed++
		}
		if c.UpdatedAt.After(lastUpdate) {
			lastUpdate = c.UpdatedAt
		}
		if active != "" && c.Status != active {
			continue
		}
		out = append(out, account.ChallengeRow{
			EntryID: c.EntryID,
			Title:   c.EventName + " · " + disciplineLevelLabel(c.Discipline, c.Level),
			Sub:     c.DogName + " · " + entryNumberLabel(c.EntryNumber) + " · " + shortDate(c.TrialDate),
			Filed:   challengeRowDetail(c),
			Status:  c.Status,
		})
	}

	vd := account.ChallengesListViewData{
		Total:   len(rows),
		Filters: challengeFilters(active, len(rows), open, review, resolved, dismissed),
		Rows:    out,
	}
	if !lastUpdate.IsZero() {
		vd.LastUpdate = relativeTime(lastUpdate)
	}
	return vd
}

// challengeRowDetail renders the second meta line on a challenge row:
// when it was filed plus a status-dependent progress clause, mirroring the
// A7 mockup ("Filed 5 days ago · admin started review yesterday").
func challengeRowDetail(c db.ListChallengesByFilerRow) string {
	filed := "Filed " + relativeTime(c.FiledAt)
	switch c.Status {
	case "open":
		return filed + " · waiting on admin"
	case "under_review":
		return filed + " · admin started review " + relativeTime(c.UpdatedAt)
	case "resolved":
		return filed + " · resolved " + relativeTime(c.UpdatedAt)
	case "dismissed":
		return filed + " · dismissed " + relativeTime(c.UpdatedAt)
	}
	return filed
}

// challengeFilters builds the status chip row with per-status counts. The
// "Review" chip is the shorter label for under_review used on the list.
func challengeFilters(active string, total, open, review, resolved, dismissed int) []account.ChallengeFilter {
	defs := []struct {
		key, label string
		count      int
	}{
		{"", "All", total},
		{"open", "Open", open},
		{"under_review", "Review", review},
		{"resolved", "Resolved", resolved},
		{"dismissed", "Dismissed", dismissed},
	}
	out := make([]account.ChallengeFilter, 0, len(defs))
	for _, d := range defs {
		href := "/account/challenges"
		if d.key != "" {
			href += "?status=" + d.key
		}
		out = append(out, account.ChallengeFilter{
			Key:    d.key,
			Label:  d.label,
			Count:  d.count,
			Href:   href,
			Active: active == d.key,
		})
	}
	return out
}

// toChallengeNewVD builds the A8 form view: the disputed-entry summary with
// a scoresheet excerpt (result pill + NQ reason / score summary) so the
// competitor sees exactly what they are disputing, plus the already-filed
// state when a prior challenge exists.
func toChallengeNewVD(r *http.Request, st *store.Store, entry db.Entry, trial db.Trial, event db.Event, existing *db.Challenge) account.ChallengeNewViewData {
	judge := challengeJudgeName(r, st, trial)
	vd := account.ChallengeNewViewData{
		EntryID:        entry.ID,
		DisputeTitle:   event.Name + " · " + disciplineLevelLabel(trial.Discipline, trial.Level) + " · " + entryNumberLabel(entry.EntryNumber),
		DisputeSub:     challengeDisputeSub(entry, trial, judge),
		EventKey:       disciplineKey(trial.Discipline),
		ScoresheetHref: fmt.Sprintf("/account/entries/%d", entry.ID),
	}
	if tpl, sheet, _, result, err := loadTemplateAndEvaluate(r, st, trial, entry.ID); err == nil {
		if result.Passed {
			vd.ResultLabel, vd.ResultKind = "Q", "qual"
		} else {
			vd.ResultLabel, vd.ResultKind = "NQ", "closed"
		}
		vd.ExcerptLabel, vd.Excerpt = challengeExcerpt(tpl, sheet, result)
	}
	if existing != nil {
		vd.AlreadyFiled = true
		vd.ChallengeStatus = existing.Status
	}
	return vd
}

// challengeDisputeSub renders the "dog · date · judged by · finalized"
// sub-line for the disputed-entry card. The judge clause is dropped when the
// judge name is unavailable, mirroring the public scoresheet's judgedLine.
func challengeDisputeSub(entry db.Entry, trial db.Trial, judge string) string {
	sub := entry.DogName + " · " + fullDate(trial.TrialDate)
	if judge != "" {
		sub += " · judged by " + judge
	}
	return sub + " · finalized"
}

// challengeJudgeName resolves the display name of the judge assigned to the
// trial, or "" when none is assigned or the lookup fails — the sub-line then
// drops the "judged by" clause.
func challengeJudgeName(r *http.Request, st *store.Store, trial db.Trial) string {
	email, err := st.TrialJudgeEmail(r.Context(), trial.ID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			slog.Error("challenge judge lookup", "trial", trial.ID, "err", err)
		}
		return ""
	}
	return judgeName(email)
}

// challengeExcerpt builds the scoresheet excerpt line for the disputing
// card: the most specific reason the entry landed where it did. An AutoNQ
// trigger description is quoted verbatim; otherwise an Insufficient tally or
// a below-threshold summary explains an NQ, and a qualifying entry gets a
// neutral score summary.
func challengeExcerpt(tpl scoring.ScoresheetTemplate, sheet scoring.ConcreteScoresheet, result scoring.ScoresheetResult) (label, text string) {
	if reasons := firedTriggerReasons(tpl, result); len(reasons) > 0 {
		return "NQ reason —", "“" + reasons[0] + "”"
	}
	pts, max := int(result.TotalPoints), int(result.MaxPoints)
	if !result.Passed {
		if result.InsufficientCount > 0 {
			return "NQ reason —", fmt.Sprintf("%s scored Insufficient — below the qualifying standard.", exerciseCountWord(result.InsufficientCount))
		}
		threshold := int(qualifyingThreshold(tpl, sheet))
		return "NQ reason —", fmt.Sprintf("Final score %d of %d — below the %d-point qualifying line.", pts, max, threshold)
	}
	return "Result —", fmt.Sprintf("Qualifying score · %d of %d points.", pts, max)
}

// firedTriggerReasons returns the descriptions of every AutoNQ trigger that
// fired against the scoresheet, in concrete-scoresheet order, by resolving
// the fired trigger codes against the template's AutoTrigger definitions.
func firedTriggerReasons(tpl scoring.ScoresheetTemplate, result scoring.ScoresheetResult) []string {
	desc := make(map[string]string)
	for _, ph := range tpl.Phases {
		for _, ex := range ph.Exercises {
			for _, at := range ex.AutoTriggers {
				desc[at.Code] = at.Description
			}
		}
	}
	var out []string
	for _, ex := range result.PerExercise {
		for _, code := range ex.AutoNQFired {
			if d := desc[code]; d != "" {
				out = append(out, d)
			}
		}
	}
	return out
}

// exerciseCountWord renders "1 exercise" / "3 exercises".
func exerciseCountWord(n int) string {
	if n == 1 {
		return "1 exercise"
	}
	return fmt.Sprintf("%d exercises", n)
}
