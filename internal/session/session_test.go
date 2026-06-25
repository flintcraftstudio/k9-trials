package session

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// mockStore implements Store for testing.
type mockStore struct {
	sessions map[string]mockSession
	users    map[int64]mockUser
}

type mockSession struct {
	userID    int64
	expiresAt time.Time
}

type mockUser struct {
	id    int64
	email string
	caps  []string
}

func newMockStore() *mockStore {
	return &mockStore{
		sessions: make(map[string]mockSession),
		users:    make(map[int64]mockUser),
	}
}

func (m *mockStore) CreateSession(_ context.Context, token string, userID int64, expiresAt time.Time) error {
	m.sessions[token] = mockSession{userID: userID, expiresAt: expiresAt}
	return nil
}

func (m *mockStore) GetSession(_ context.Context, token string) (int64, time.Time, error) {
	s, ok := m.sessions[token]
	if !ok || time.Now().After(s.expiresAt) {
		return 0, time.Time{}, sql.ErrNoRows
	}
	return s.userID, s.expiresAt, nil
}

func (m *mockStore) DeleteSession(_ context.Context, token string) error {
	delete(m.sessions, token)
	return nil
}

func (m *mockStore) GetUserByID(_ context.Context, id int64) (int64, string, error) {
	u, ok := m.users[id]
	if !ok {
		return 0, "", sql.ErrNoRows
	}
	return u.id, u.email, nil
}

func (m *mockStore) UserCapabilities(_ context.Context, id int64) ([]string, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return u.caps, nil
}

// addUser seeds a user with the given capability grants ('judge'/'admin').
// The competitor baseline is implicit, so pass no caps for a plain competitor.
func (m *mockStore) addUser(id int64, email string, caps ...string) {
	m.users[id] = mockUser{id: id, email: email, caps: caps}
}

// okHandler returns 200 with the user email if authenticated, or "anonymous".
func okHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if u := FromContext(r.Context()); u != nil {
			w.Write([]byte(u.Email))
		} else {
			w.Write([]byte("anonymous"))
		}
	}
}

func TestRequireAuth_RedirectsWhenUnauthenticated(t *testing.T) {
	handler := RequireAuth(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc != "/login" {
		t.Fatalf("expected redirect to /login, got %q", loc)
	}
}

func TestRequireAuth_AllowsAuthenticated(t *testing.T) {
	handler := RequireAuth(okHandler())

	ctx := withUser(context.Background(), &User{ID: 1, Email: "admin@example.com"})
	req := httptest.NewRequest(http.MethodGet, "/admin", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "admin@example.com" {
		t.Fatalf("expected body %q, got %q", "admin@example.com", rec.Body.String())
	}
}

func TestMiddleware_NoCookie_PassesThrough(t *testing.T) {
	store := newMockStore()
	handler := Middleware(store)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "anonymous" {
		t.Fatalf("expected body %q, got %q", "anonymous", rec.Body.String())
	}
}

func TestMiddleware_ValidSession_AttachesUser(t *testing.T) {
	store := newMockStore()
	store.addUser(1, "user@example.com")
	store.sessions["valid-token"] = mockSession{
		userID:    1,
		expiresAt: time.Now().Add(time.Hour),
	}

	handler := Middleware(store)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: "valid-token"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "user@example.com" {
		t.Fatalf("expected body %q, got %q", "user@example.com", rec.Body.String())
	}
}

func TestMiddleware_ExpiredSession_NoUser(t *testing.T) {
	store := newMockStore()
	store.addUser(1, "user@example.com")
	store.sessions["expired-token"] = mockSession{
		userID:    1,
		expiresAt: time.Now().Add(-time.Hour),
	}

	handler := Middleware(store)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: "expired-token"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Body.String() != "anonymous" {
		t.Fatalf("expected body %q, got %q", "anonymous", rec.Body.String())
	}
}

func TestMiddleware_InvalidToken_NoUser(t *testing.T) {
	store := newMockStore()

	handler := Middleware(store)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: "nonexistent-token"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Body.String() != "anonymous" {
		t.Fatalf("expected body %q, got %q", "anonymous", rec.Body.String())
	}
}

func TestMiddleware_UserDeleted_NoUser(t *testing.T) {
	store := newMockStore()
	// Session exists but user does not
	store.sessions["orphan-token"] = mockSession{
		userID:    99,
		expiresAt: time.Now().Add(time.Hour),
	}

	handler := Middleware(store)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: "orphan-token"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Body.String() != "anonymous" {
		t.Fatalf("expected body %q, got %q", "anonymous", rec.Body.String())
	}
}

func TestFullChain_ProtectedRoute_Unauthenticated(t *testing.T) {
	store := newMockStore()
	handler := Middleware(store)(RequireAuth(okHandler()))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/login" {
		t.Fatalf("expected redirect to /login, got %q", loc)
	}
}

func TestFullChain_ProtectedRoute_Authenticated(t *testing.T) {
	store := newMockStore()
	store.addUser(1, "admin@example.com")
	store.sessions["good-token"] = mockSession{
		userID:    1,
		expiresAt: time.Now().Add(time.Hour),
	}

	handler := Middleware(store)(RequireAuth(okHandler()))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: "good-token"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "admin@example.com" {
		t.Fatalf("expected body %q, got %q", "admin@example.com", rec.Body.String())
	}
}

func TestCreate_SetsSecureCookie(t *testing.T) {
	store := newMockStore()
	store.addUser(1, "user@example.com")

	rec := httptest.NewRecorder()
	err := Create(context.Background(), rec, store, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}

	c := cookies[0]
	if c.Name != cookieName {
		t.Fatalf("expected cookie name %q, got %q", cookieName, c.Name)
	}
	if c.Value == "" {
		t.Fatal("expected non-empty cookie value")
	}
	if !c.HttpOnly {
		t.Fatal("expected HttpOnly cookie")
	}
	if !c.Secure {
		t.Fatal("expected Secure cookie")
	}
	if c.SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected SameSite=Lax, got %v", c.SameSite)
	}

	// Verify session was persisted in store
	if len(store.sessions) != 1 {
		t.Fatalf("expected 1 session in store, got %d", len(store.sessions))
	}
}

func TestDestroy_ClearsCookieAndSession(t *testing.T) {
	store := newMockStore()
	store.sessions["to-delete"] = mockSession{
		userID:    1,
		expiresAt: time.Now().Add(time.Hour),
	}

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: "to-delete"})
	rec := httptest.NewRecorder()

	err := Destroy(context.Background(), rec, req, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Session should be removed from store
	if len(store.sessions) != 0 {
		t.Fatalf("expected 0 sessions in store, got %d", len(store.sessions))
	}

	// Cookie should be cleared (MaxAge = -1)
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].MaxAge != -1 {
		t.Fatalf("expected MaxAge=-1, got %d", cookies[0].MaxAge)
	}
}

func TestFromContext_NilWhenNoUser(t *testing.T) {
	if u := FromContext(context.Background()); u != nil {
		t.Fatalf("expected nil, got %+v", u)
	}
}

func TestRequireAdmin_AllowsAdmin(t *testing.T) {
	handler := RequireAdmin(okHandler())
	ctx := withUser(context.Background(), &User{ID: 1, Email: "a@x", Caps: []string{"admin"}})
	req := httptest.NewRequest(http.MethodGet, "/admin", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireAdmin_ForbidsJudge(t *testing.T) {
	handler := RequireAdmin(okHandler())
	ctx := withUser(context.Background(), &User{ID: 2, Email: "j@x", Caps: []string{"judge"}})
	req := httptest.NewRequest(http.MethodGet, "/admin", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestRequireAdmin_ForbidsCompetitor(t *testing.T) {
	// A plain competitor (no caps) is blocked from admin surfaces.
	handler := RequireAdmin(okHandler())
	ctx := withUser(context.Background(), &User{ID: 3, Email: "c@x"})
	req := httptest.NewRequest(http.MethodGet, "/admin", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestRequireAdmin_RedirectsAnonymous(t *testing.T) {
	handler := RequireAdmin(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", rec.Code)
	}
}

func TestRequireJudge_AllowsJudge(t *testing.T) {
	handler := RequireJudge(okHandler())
	ctx := withUser(context.Background(), &User{ID: 1, Email: "j@x", Caps: []string{"judge"}})
	req := httptest.NewRequest(http.MethodGet, "/judge", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// TestRequireJudge_AdminIsSuperset asserts the admin-as-superset rule: an admin
// account (no explicit "judge" cap) still passes a RequireJudge gate.
func TestRequireJudge_AdminIsSuperset(t *testing.T) {
	handler := RequireJudge(okHandler())
	ctx := withUser(context.Background(), &User{ID: 1, Email: "a@x", Caps: []string{"admin"}})
	req := httptest.NewRequest(http.MethodGet, "/judge", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("admin superset: expected 200, got %d", rec.Code)
	}
}

// TestRequireJudge_ForbidsCompetitor confirms a plain competitor (no caps) is
// blocked from judge surfaces.
func TestRequireJudge_ForbidsCompetitor(t *testing.T) {
	handler := RequireJudge(okHandler())
	ctx := withUser(context.Background(), &User{ID: 2, Email: "c@x"})
	req := httptest.NewRequest(http.MethodGet, "/judge", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

// TestJudgeHoldsBaseline confirms a judge account (which also holds the implicit
// competitor baseline) passes BOTH a judge gate and the RequireAuth competitor
// gate — the core multi-role guarantee: a judge can enter their own dog.
func TestJudgeHoldsBaseline(t *testing.T) {
	u := &User{ID: 1, Email: "judge@x", Caps: []string{"judge"}}

	if !u.IsJudge() {
		t.Fatal("expected IsJudge() true for judge cap")
	}
	if !u.IsCompetitor() {
		t.Fatal("expected IsCompetitor() true (universal baseline)")
	}

	// Judge gate passes.
	jrec := httptest.NewRecorder()
	jctx := withUser(context.Background(), u)
	RequireJudge(okHandler()).ServeHTTP(jrec, httptest.NewRequest(http.MethodGet, "/judge", nil).WithContext(jctx))
	if jrec.Code != http.StatusOK {
		t.Fatalf("judge gate: expected 200, got %d", jrec.Code)
	}

	// Competitor (RequireAuth) gate passes too.
	crec := httptest.NewRecorder()
	cctx := withUser(context.Background(), u)
	RequireAuth(okHandler()).ServeHTTP(crec, httptest.NewRequest(http.MethodGet, "/account/dogs", nil).WithContext(cctx))
	if crec.Code != http.StatusOK {
		t.Fatalf("competitor gate: expected 200, got %d", crec.Code)
	}
}

func TestHas_NilUserAndMembership(t *testing.T) {
	var nilU *User
	if nilU.Has("admin") {
		t.Fatal("nil user must not hold any capability")
	}
	u := &User{Caps: []string{"judge"}}
	if !u.Has("judge") {
		t.Fatal("expected Has(judge) true")
	}
	if u.Has("admin") {
		t.Fatal("expected Has(admin) false — Has does not apply superset rule")
	}
}
