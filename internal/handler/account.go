package handler

import (
	"log/slog"
	"net/http"

	"github.com/flintcraftstudio/k9-trials/internal/view/account"
)

// AccountDashboard renders the logged-in competitor's landing page.
func AccountDashboard() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := account.DashboardPage().Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// AccountProfile renders the profile editor.
func AccountProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := account.ProfilePage().Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// AccountDogs lists every dog owned by the logged-in competitor.
func AccountDogs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := account.DogsListPage().Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// AccountDogsNew renders the new-dog form.
func AccountDogsNew() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := account.DogsFormPage("").Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// AccountDogsEdit renders the edit-dog form for {id}.
func AccountDogsEdit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := account.DogsFormPage(id).Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// AccountEntries lists the logged-in competitor's entries across all events.
func AccountEntries() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := account.EntriesListPage().Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// AccountEntryDetail renders the read-only scoresheet view for an entry
// the logged-in competitor owns. Adds a "Challenge this score" CTA once
// the entry is finalized.
func AccountEntryDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := account.EntryDetailPage(id).Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}
