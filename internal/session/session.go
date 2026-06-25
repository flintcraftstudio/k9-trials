package session

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"time"
)

const (
	cookieName     = "session_token"
	sessionMaxAge  = 7 * 24 * time.Hour
)

// Secure controls the Secure flag on the session cookie. It defaults to true
// (production-safe) and is set once at startup. Set it to false for local
// development over plain HTTP, where browsers refuse to store a Secure cookie
// served over http://localhost and the session would never persist.
var Secure = true

type ctxKey struct{}

// User represents the authenticated user attached to a request context.
//
// Caps holds the account's additive capability grants ('judge'/'admin'), loaded
// once at session resolution. The 'competitor' baseline is universal and never
// stored in Caps. All authorization and display decisions derive from Caps —
// there is no single-role field (the legacy users.role column was dropped).
type User struct {
	ID    int64
	Email string
	Caps  []string
}

// Has reports whether the user holds the named capability.
// Contract: pure membership test over u.Caps; returns false for a nil user.
// It does NOT apply the admin-superset rule — gate helpers layer that on top.
func (u *User) Has(cap string) bool {
	if u == nil {
		return false
	}
	for _, c := range u.Caps {
		if c == cap {
			return true
		}
	}
	return false
}

// IsAdmin reports whether this user holds the admin capability.
// Contract: reads Caps (Has("admin")).
func (u *User) IsAdmin() bool { return u.Has("admin") }

// IsJudge reports whether this user holds the judge capability (judge-eligible).
// Contract: reads Caps (Has("judge")). Does not imply admin.
func (u *User) IsJudge() bool { return u.Has("judge") }

// IsCompetitor reports whether this user has the competitor baseline.
// Contract: competitor is the universal baseline for every authenticated
// account, so this is true for any non-nil user regardless of Caps.
func (u *User) IsCompetitor() bool { return u != nil }

// Label returns the single human-readable role label to show for this user,
// derived from capabilities: "Admin" if the user holds the
// admin capability, else "Judge" if judge, else "Competitor" (the universal
// baseline). Use this for display chips so the surface reflects capabilities.
func (u *User) Label() string {
	switch {
	case u.IsAdmin():
		return "Admin"
	case u.IsJudge():
		return "Judge"
	default:
		return "Competitor"
	}
}

// CapsLabel returns the single display label for a set of capability strings,
// using the same precedence as User.Label (admin > judge > competitor). It lets
// callers that hold raw caps (e.g. the admin user list) render a label without
// constructing a User. An empty/nil slice yields "Competitor".
func CapsLabel(caps []string) string {
	hasAdmin, hasJudge := false, false
	for _, c := range caps {
		switch c {
		case "admin":
			hasAdmin = true
		case "judge":
			hasJudge = true
		}
	}
	switch {
	case hasAdmin:
		return "Admin"
	case hasJudge:
		return "Judge"
	default:
		return "Competitor"
	}
}

// FromContext returns the authenticated user from the request context, or nil.
func FromContext(ctx context.Context) *User {
	u, _ := ctx.Value(ctxKey{}).(*User)
	return u
}

// withUser attaches a user to the context.
func withUser(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, ctxKey{}, u)
}

// Store defines the session persistence interface.
type Store interface {
	CreateSession(ctx context.Context, token string, userID int64, expiresAt time.Time) error
	GetSession(ctx context.Context, token string) (userID int64, expiresAt time.Time, err error)
	DeleteSession(ctx context.Context, token string) error
	GetUserByID(ctx context.Context, id int64) (userID int64, email string, err error)
	// UserCapabilities returns the explicit capability grants ('judge'/'admin')
	// for a user. The competitor baseline is implicit and not included.
	UserCapabilities(ctx context.Context, userID int64) ([]string, error)
}

// Middleware loads the session from the cookie and attaches the user to the request context.
func Middleware(store Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cookieName)
			if err != nil || cookie.Value == "" {
				next.ServeHTTP(w, r)
				return
			}

			userID, expiresAt, err := store.GetSession(r.Context(), cookie.Value)
			if err != nil || time.Now().After(expiresAt) {
				next.ServeHTTP(w, r)
				return
			}

			id, email, err := store.GetUserByID(r.Context(), userID)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			// Load capability grants once per request, at session resolution.
			// A load failure leaves Caps empty (competitor baseline only) rather
			// than dropping the session — the user is still authenticated.
			caps, _ := store.UserCapabilities(r.Context(), id)

			ctx := withUser(r.Context(), &User{ID: id, Email: email, Caps: caps})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth wraps a handler and redirects unauthenticated users to /login.
// Contract: passes any logged-in account through — every authenticated user
// holds the competitor baseline, so this is the canonical competitor gate.
// Unauthenticated callers are redirected to /login (303 See Other).
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if FromContext(r.Context()) == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireCapability wraps a handler so only users who hold at least one of the
// named capabilities may proceed. Contract: the admin capability is a superset
// and always passes, regardless of which caps are requested. Anonymous callers
// are redirected to /login (303 See Other); authenticated-but-unauthorized
// callers get 403 Forbidden. Authorization reads u.Caps only.
func RequireCapability(next http.Handler, caps ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := FromContext(r.Context())
		if u == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if u.Has("admin") {
			next.ServeHTTP(w, r)
			return
		}
		for _, c := range caps {
			if u.Has(c) {
				next.ServeHTTP(w, r)
				return
			}
		}
		http.Error(w, "Forbidden", http.StatusForbidden)
	})
}

// RequireAdmin permits only users holding the admin capability.
// Contract: thin alias for RequireCapability(next, "admin"); reads Caps.
func RequireAdmin(next http.Handler) http.Handler {
	return RequireCapability(next, "admin")
}

// RequireJudge permits users holding the judge capability; admins pass as a
// superset. Contract: thin alias for RequireCapability(next, "judge"); reads Caps.
func RequireJudge(next http.Handler) http.Handler {
	return RequireCapability(next, "judge")
}

// Create generates a new session token, persists it, and sets the cookie.
func Create(ctx context.Context, w http.ResponseWriter, store Store, userID int64) error {
	token, err := generateToken()
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(sessionMaxAge)
	if err := store.CreateSession(ctx, token, userID, expiresAt); err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   Secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionMaxAge.Seconds()),
	})
	return nil
}

// Destroy removes the session from the store and clears the cookie.
func Destroy(ctx context.Context, w http.ResponseWriter, r *http.Request, store Store) error {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return nil
	}

	if err := store.DeleteSession(ctx, cookie.Value); err != nil && err != sql.ErrNoRows {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   Secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	return nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
