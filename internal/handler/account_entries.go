package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/flintcraftstudio/k9-trials/internal/db"
	"github.com/flintcraftstudio/k9-trials/internal/store"
	"github.com/flintcraftstudio/k9-trials/internal/view/account"
)

// AccountEntries serves GET /account/entries — the competitor's entries
// across all events (A5), with status filter chips. htmx filter requests
// receive only the results fragment.
func AccountEntries(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := currentCompetitor(w, r, st)
		if !ok {
			return
		}
		entries, err := st.ListHandlerEntries(r.Context(), c.ID)
		if err != nil {
			slog.Error("account entries", "competitor", c.ID, "err", err)
			http.Error(w, "account unavailable", http.StatusInternalServerError)
			return
		}
		regs, err := st.ListPendingRegistrations(r.Context(), c.ID)
		if err != nil {
			slog.Error("account pending registrations", "competitor", c.ID, "err", err)
			http.Error(w, "account unavailable", http.StatusInternalServerError)
			return
		}
		filter := r.URL.Query().Get("status")
		if !validEntryFilter(filter) {
			filter = ""
		}
		data := toEntriesListVD(r, st, entries, regs, filter)
		if r.Header.Get("HX-Request") == "true" {
			renderPublic(w, r, account.EntriesResults(data))
			return
		}
		renderPublic(w, r, account.EntriesListPage(data))
	}
}

// AccountEntryDetail serves GET /account/entries/{id} — the read-only
// scoresheet for an entry the competitor owns (A6). 404 when the id misses
// or the entry was handled by someone else.
func AccountEntryDetail(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := currentCompetitor(w, r, st)
		if !ok {
			return
		}
		entry, trial, event, ok := loadOwnedEntry(w, r, st, c)
		if !ok {
			return
		}
		var existing *db.Challenge
		if ch, err := st.ChallengeForEntry(r.Context(), entry.ID, c.ID); err == nil {
			existing = &ch
		} else if !errors.Is(err, sql.ErrNoRows) {
			slog.Error("entry challenge lookup", "entry", entry.ID, "err", err)
		}
		renderPublic(w, r, account.EntryDetailPage(toEntryDetailVD(r, st, entry, trial, event, existing)))
	}
}

// AccountEntryWithdraw serves POST /account/entries/{id}/withdraw — a
// competitor's request to withdraw their own accepted entry (Q1). It routes
// to an admin for confirmation rather than voiding the entry immediately, and
// is only offered before the run is scored.
func AccountEntryWithdraw(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, ok := currentCompetitor(w, r, st)
		if !ok {
			return
		}
		entry, _, _, ok := loadOwnedEntry(w, r, st, c)
		if !ok {
			return
		}
		if entry.Status != "registered" {
			http.Error(w, "this entry can no longer be withdrawn", http.StatusConflict)
			return
		}
		if err := st.RequestEntryWithdrawal(r.Context(), entry.ID, c.ID); err != nil {
			slog.Error("request withdrawal", "entry", entry.ID, "err", err)
			http.Error(w, "could not request withdrawal", http.StatusInternalServerError)
			return
		}
		hxRedirect(w, r, fmt.Sprintf("/account/entries/%d", entry.ID))
	}
}

// loadOwnedEntry parses the {id} path segment, loads the entry with its
// trial and event, and confirms the logged-in competitor is the handler of
// record. It writes a 404 (and returns ok=false) on a missing id, a missing
// row, or an entry owned by someone else, so a guessed id never leaks
// another competitor's scoresheet.
func loadOwnedEntry(w http.ResponseWriter, r *http.Request, st *store.Store, c db.Competitor) (db.Entry, db.Trial, db.Event, bool) {
	entryID, ok := parseEntryID(r)
	if !ok {
		http.NotFound(w, r)
		return db.Entry{}, db.Trial{}, db.Event{}, false
	}
	return ownedEntryByID(w, r, st, c, entryID)
}

// ownedEntryByID is the ownership-checked entry load shared by the
// path-based entry detail (A6) and the query-based challenge form (A8).
func ownedEntryByID(w http.ResponseWriter, r *http.Request, st *store.Store, c db.Competitor, entryID int64) (db.Entry, db.Trial, db.Event, bool) {
	entry, trial, event, err := st.LoadEntryWithTrial(r.Context(), entryID)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return db.Entry{}, db.Trial{}, db.Event{}, false
	}
	if err != nil {
		slog.Error("owned entry load", "entry", entryID, "err", err)
		http.Error(w, "account unavailable", http.StatusInternalServerError)
		return db.Entry{}, db.Trial{}, db.Event{}, false
	}
	if !entry.HandlerID.Valid || entry.HandlerID.Int64 != c.ID {
		http.NotFound(w, r)
		return db.Entry{}, db.Trial{}, db.Event{}, false
	}
	return entry, trial, event, true
}

// validEntryFilter reports whether key is one of the recognized status
// filters (empty means "all").
func validEntryFilter(key string) bool {
	switch key {
	case "", "upcoming", "scoring", "finalized", "withdrawn":
		return true
	}
	return false
}
