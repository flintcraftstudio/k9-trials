package handler

import (
	"net/http"

	"github.com/flintcraftstudio/k9-trials/internal/store"
	"github.com/flintcraftstudio/k9-trials/internal/view/competitors"
	"github.com/flintcraftstudio/k9-trials/internal/view/dogs"
)

// Public read-only handlers for the events surface (P1–P4) live in
// public_events.go. The competitor directory + profile and dog profile
// (P5–P7) are stubbed here pending their sqlc queries; they take the
// store now so wiring is stable when the bodies land.

// CompetitorSearch renders the public competitor directory + search (P5).
func CompetitorSearch(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderPublic(w, r, competitors.SearchPage())
	}
}

// CompetitorProfile renders a public competitor profile by handle (P6).
func CompetitorProfile(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handle := r.PathValue("handle")
		renderPublic(w, r, competitors.ProfilePage(handle))
	}
}

// DogProfile renders a public dog profile by id (P7).
func DogProfile(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		renderPublic(w, r, dogs.ProfilePage(id))
	}
}
