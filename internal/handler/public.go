package handler

import (
	"log/slog"
	"net/http"

	"github.com/flintcraftstudio/k9-trials/internal/view/competitors"
	"github.com/flintcraftstudio/k9-trials/internal/view/dogs"
	"github.com/flintcraftstudio/k9-trials/internal/view/events"
)

// EventsList renders the public events index.
func EventsList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := events.ListPage().Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// EventDetail renders the public per-event page.
func EventDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		if err := events.DetailPage(slug).Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// TrialDetail renders the public per-trial leaderboard.
func TrialDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		trialID := r.PathValue("id")
		if err := events.TrialDetailPage(slug, trialID).Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// CompetitorSearch renders the public competitor directory + search.
func CompetitorSearch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := competitors.SearchPage().Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// CompetitorProfile renders a public competitor profile by handle.
func CompetitorProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handle := r.PathValue("handle")
		if err := competitors.ProfilePage(handle).Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// DogProfile renders a public dog profile by id.
func DogProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := dogs.ProfilePage(id).Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}
