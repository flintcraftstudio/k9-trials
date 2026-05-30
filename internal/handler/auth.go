package handler

import (
	"log/slog"
	"net/http"

	"github.com/flintcraftstudio/k9-trials/internal/session"
	"github.com/flintcraftstudio/k9-trials/internal/store"
	"github.com/flintcraftstudio/k9-trials/internal/view"

	"golang.org/x/crypto/bcrypt"
)

// LoginPage handles GET /login and renders the login form.
func LoginPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if session.FromContext(r.Context()) != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		if err := view.LoginPage("", "").Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// LoginSubmit handles POST /login, validates credentials, creates a session, and redirects.
func LoginSubmit(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		if email == "" || password == "" {
			if err := view.LoginForm("Email and password are required.", email).Render(r.Context(), w); err != nil {
				slog.Error("render error", "err", err)
			}
			return
		}

		userID, _, passwordHash, _, err := s.GetUserByEmail(r.Context(), email)
		if err != nil {
			if err := view.LoginForm("Invalid email or password.", email).Render(r.Context(), w); err != nil {
				slog.Error("render error", "err", err)
			}
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
			if err := view.LoginForm("Invalid email or password.", email).Render(r.Context(), w); err != nil {
				slog.Error("render error", "err", err)
			}
			return
		}

		if err := session.Create(r.Context(), w, s, userID); err != nil {
			slog.Error("session create error", "err", err)
			if err := view.LoginForm("Something went wrong. Please try again.", email).Render(r.Context(), w); err != nil {
				slog.Error("render error", "err", err)
			}
			return
		}

		hxRedirect(w, r, "/")
	}
}

// Logout handles POST /logout, destroys the session, and redirects.
func Logout(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := session.Destroy(r.Context(), w, r, s); err != nil {
			slog.Error("session destroy error", "err", err)
		}
		hxRedirect(w, r, "/")
	}
}
