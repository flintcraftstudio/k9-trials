package handler

import (
	"log/slog"
	"net/http"

	"github.com/flintcraftstudio/k9-trials/internal/view"
)

// SignupPage renders the competitor self-signup form.
func SignupPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := view.SignupPage().Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// SignupSubmit will create a user (role=competitor) and competitor row,
// then start a session and redirect to /account.
func SignupSubmit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/signup", http.StatusSeeOther)
	}
}
