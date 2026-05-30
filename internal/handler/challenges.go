package handler

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/flintcraftstudio/k9-trials/internal/db"
	"github.com/flintcraftstudio/k9-trials/internal/store"
	"github.com/flintcraftstudio/k9-trials/internal/view/account"
)

// AccountChallenges serves GET /account/challenges — every dispute the
// competitor has filed (A7), newest first.
func AccountChallenges(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := currentCompetitor(w, r, st)
		if !ok {
			return
		}
		rows, err := st.ListFiledChallenges(r.Context(), c.ID)
		if err != nil {
			slog.Error("challenges list", "competitor", c.ID, "err", err)
			http.Error(w, "account unavailable", http.StatusInternalServerError)
			return
		}
		renderPublic(w, r, account.ChallengesListPage(toChallengesListVD(rows)))
	}
}

// AccountChallengeNew serves GET /account/challenges/new?entry={id} — the
// file-a-challenge form (A8), prefilled with the disputed entry summary.
// The entry must be owned by the competitor and finalized.
func AccountChallengeNew(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := currentCompetitor(w, r, st)
		if !ok {
			return
		}
		entryID, ok := parseQueryEntryID(r)
		if !ok {
			http.NotFound(w, r)
			return
		}
		entry, trial, event, ok := ownedEntryByID(w, r, st, c, entryID)
		if !ok {
			return
		}
		// Only finalized entries can be disputed; a non-finalized entry has
		// no score to challenge.
		if entry.Status != "finalized" {
			http.Redirect(w, r, "/account/entries/"+strconv.FormatInt(entry.ID, 10), http.StatusSeeOther)
			return
		}
		existing := lookupExistingChallenge(r, st, entry.ID, c.ID)
		renderPublic(w, r, account.ChallengesNewPage(toChallengeNewVD(r, st, entry, trial, event, existing)))
	}
}

// AccountChallengeSubmit serves POST /account/challenges — validates and
// records a dispute against a finalized entry the competitor owns, then
// redirects to the challenges list. Re-renders the form fragment on a
// validation error.
func AccountChallengeSubmit(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := currentCompetitor(w, r, st)
		if !ok {
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		entryID, err := strconv.ParseInt(r.FormValue("entry"), 10, 64)
		if err != nil || entryID <= 0 {
			http.NotFound(w, r)
			return
		}
		entry, trial, event, ok := ownedEntryByID(w, r, st, c, entryID)
		if !ok {
			return
		}
		if entry.Status != "finalized" {
			http.Error(w, "entry is not finalized", http.StatusConflict)
			return
		}

		// A dispute already on file: send the competitor to their list
		// rather than stacking a second one.
		if existing := lookupExistingChallenge(r, st, entry.ID, c.ID); existing != nil {
			hxRedirect(w, r, "/account/challenges")
			return
		}

		reason := strings.TrimSpace(r.FormValue("reason"))
		vd := toChallengeNewVD(r, st, entry, trial, event, nil)
		vd.Reason = reason

		if !challengeWindowOpen(entry.UpdatedAt) {
			vd.Err = "The 7-day window to challenge this score has closed."
			renderPublic(w, r, account.ChallengeForm(vd))
			return
		}
		if reason == "" {
			vd.Err = "Tell us why you are disputing this score."
			renderPublic(w, r, account.ChallengeForm(vd))
			return
		}

		if _, err := st.FileChallenge(r.Context(), entry.ID, c.ID, reason); err != nil {
			slog.Error("file challenge", "entry", entry.ID, "err", err)
			vd.Err = "Something went wrong. Please try again."
			renderPublic(w, r, account.ChallengeForm(vd))
			return
		}
		hxRedirect(w, r, "/account/challenges")
	}
}

// parseQueryEntryID pulls and validates the ?entry= query parameter.
func parseQueryEntryID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.URL.Query().Get("entry"), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// lookupExistingChallenge returns the competitor's prior challenge for an
// entry, or nil when none exists. A lookup error other than no-rows is
// logged and treated as "none" so it never blocks the page.
func lookupExistingChallenge(r *http.Request, st *store.Store, entryID, filerID int64) *db.Challenge {
	ch, err := st.ChallengeForEntry(r.Context(), entryID, filerID)
	if err == nil {
		return &ch
	}
	if !errors.Is(err, sql.ErrNoRows) {
		slog.Error("challenge lookup", "entry", entryID, "err", err)
	}
	return nil
}
