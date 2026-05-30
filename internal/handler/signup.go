package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/flintcraftstudio/k9-trials/internal/session"
	"github.com/flintcraftstudio/k9-trials/internal/store"
	"github.com/flintcraftstudio/k9-trials/internal/view"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// handlePattern is the allowed shape of a public handle: lowercase
// letters, digits, and hyphens. It becomes a URL slug so it stays in the
// unreserved character set.
var handlePattern = regexp.MustCompile(`^[a-z0-9-]+$`)

// minPasswordLen is the floor advertised on the signup form.
const minPasswordLen = 8

// SignupPage handles GET /signup and renders the self-signup form. Already
// authenticated callers are bounced to their account rather than shown a
// second signup.
func SignupPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if session.FromContext(r.Context()) != nil {
			http.Redirect(w, r, "/account", http.StatusSeeOther)
			return
		}
		if err := view.SignupPage("", view.SignupValues{}).Render(r.Context(), w); err != nil {
			slog.Error("render error", "err", err)
		}
	}
}

// SignupSubmit handles POST /signup: validates the four fields, provisions
// a competitor account (user + competitor row) in one transaction, starts
// a session, and redirects to /account. Validation failures re-render the
// form fragment with the offending message and the user's prior input.
func SignupSubmit(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		email := strings.TrimSpace(r.FormValue("email"))
		password := r.FormValue("password")
		displayName := strings.TrimSpace(r.FormValue("display_name"))
		handle := strings.ToLower(strings.TrimSpace(r.FormValue("handle")))
		vals := view.SignupValues{Email: email, DisplayName: displayName, Handle: handle}

		fail := func(msg string) {
			if err := view.SignupForm(msg, vals).Render(r.Context(), w); err != nil {
				slog.Error("render error", "err", err)
			}
		}

		switch {
		case email == "" || !strings.Contains(email, "@"):
			fail("Enter a valid email address.")
			return
		case len(password) < minPasswordLen:
			fail("Password must be at least 8 characters.")
			return
		case displayName == "":
			fail("Display name is required.")
			return
		case handle == "" || !handlePattern.MatchString(handle):
			fail("Handle can use only lowercase letters, digits, and hyphens.")
			return
		}

		available, err := st.HandleAvailable(r.Context(), handle, 0)
		if err != nil {
			slog.Error("signup handle check", "err", err)
			fail("Something went wrong. Please try again.")
			return
		}
		if !available {
			fail("That handle is already taken. Pick another.")
			return
		}

		userID, err := st.CreateCompetitorAccount(r.Context(), email, password, displayName, handle)
		if err != nil {
			// A racing duplicate email or handle slips past the pre-check
			// and trips a UNIQUE constraint; surface it as a friendly
			// message rather than a 500.
			if isUniqueViolation(err) {
				fail("That email or handle is already in use.")
				return
			}
			slog.Error("create competitor account", "err", err)
			fail("Something went wrong. Please try again.")
			return
		}

		if err := session.Create(r.Context(), w, st, userID); err != nil {
			slog.Error("session create error", "err", err)
			fail("Account created, but sign-in failed. Try logging in.")
			return
		}

		hxRedirect(w, r, "/account")
	}
}

// SignupHandleCheck handles GET /signup/handle — the live availability
// probe. It validates the handle shape, checks the store, and returns the
// status fragment. A malformed handle reports as unavailable so the user
// sees the corrective hint.
func SignupHandleCheck(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handle := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("handle")))
		if handle == "" {
			renderPublic(w, r, view.HandleStatus("", false, false))
			return
		}
		if !handlePattern.MatchString(handle) {
			renderPublic(w, r, view.HandleStatus(handle, false, true))
			return
		}
		available, err := st.HandleAvailable(r.Context(), handle, 0)
		if err != nil {
			slog.Error("handle check", "err", err)
			renderPublic(w, r, view.HandleStatus(handle, false, false))
			return
		}
		renderPublic(w, r, view.HandleStatus(handle, available, true))
	}
}

// isUniqueViolation reports whether err is a SQLite UNIQUE constraint
// failure — the signal that a handle or email collided despite the
// pre-insert check (a race between two signups).
func isUniqueViolation(err error) bool {
	var se *sqlite.Error
	if errors.As(err, &se) {
		return se.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE
	}
	return false
}
