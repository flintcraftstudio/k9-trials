package handler

import (
	"log/slog"
	"net/http"

	"github.com/flintcraftstudio/k9-trials/internal/view/account"
)

// RegisterPage renders the event registration form.
func RegisterPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		if err := account.RegisterPage(slug).Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// RegisterSubmit will create a pending registration for each selected
// trial within the event, then redirect to /account/entries.
func RegisterSubmit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		http.Redirect(w, r, "/events/"+slug+"/register", http.StatusSeeOther)
	}
}
