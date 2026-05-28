package handler

import (
	"log/slog"
	"net/http"

	"github.com/flintcraftstudio/k9-trials/internal/view/account"
)

// AccountChallenges lists challenges filed by the logged-in competitor.
func AccountChallenges() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := account.ChallengesListPage().Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// AccountChallengeNew renders the form to file a new challenge against an
// entry. Entry id comes from the ?entry= query parameter.
func AccountChallengeNew() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entryID := r.URL.Query().Get("entry")
		if err := account.ChallengesNewPage(entryID).Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// AccountChallengeSubmit will insert a new row in challenges and redirect
// to /account/challenges.
func AccountChallengeSubmit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/account/challenges", http.StatusSeeOther)
	}
}
