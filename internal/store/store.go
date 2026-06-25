package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/flintcraftstudio/k9-trials/internal/db"

	"golang.org/x/crypto/bcrypt"
)

// Store wraps the database connection and provides query methods. The
// hand-written methods on Store cover the auth surface (users, sessions);
// for everything else (events, trials, entries, scores), use Q() to
// reach the sqlc-generated typed queries.
type Store struct {
	db *sql.DB
	q  *db.Queries
}

// New creates a new Store.
func New(database *sql.DB) *Store {
	return &Store{db: database, q: db.New(database)}
}

// Q returns the sqlc-generated query interface.
func (s *Store) Q() *db.Queries {
	return s.q
}

// DB returns the underlying *sql.DB. Use sparingly — prefer Q() or the
// hand-written methods. Needed for transaction handling and migrations.
func (s *Store) DB() *sql.DB {
	return s.db
}

// CreateSession inserts a new session row.
func (s *Store) CreateSession(ctx context.Context, token string, userID int64, expiresAt time.Time) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)",
		token, userID, expiresAt,
	)
	return err
}

// GetSession retrieves a valid (non-expired) session by token.
func (s *Store) GetSession(ctx context.Context, token string) (int64, time.Time, error) {
	var userID int64
	var expiresAt time.Time
	err := s.db.QueryRowContext(ctx,
		"SELECT user_id, expires_at FROM sessions WHERE id = ? AND expires_at > CURRENT_TIMESTAMP",
		token,
	).Scan(&userID, &expiresAt)
	return userID, expiresAt, err
}

// DeleteSession removes a session by token.
func (s *Store) DeleteSession(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM sessions WHERE id = ?", token)
	return err
}

// GetUserByID retrieves a user's id and email by their ID.
func (s *Store) GetUserByID(ctx context.Context, id int64) (int64, string, error) {
	var userID int64
	var email string
	err := s.db.QueryRowContext(ctx,
		"SELECT id, email FROM users WHERE id = ?",
		id,
	).Scan(&userID, &email)
	return userID, email, err
}

// GetUserByEmail retrieves a user by email, returning id, email, and password hash.
func (s *Store) GetUserByEmail(ctx context.Context, email string) (int64, string, string, error) {
	var id int64
	var userEmail, passwordHash string
	err := s.db.QueryRowContext(ctx,
		"SELECT id, email, password_hash FROM users WHERE email = ?",
		email,
	).Scan(&id, &userEmail, &passwordHash)
	return id, userEmail, passwordHash, err
}

// CreateUser inserts a new account at the competitor baseline with a
// bcrypt-hashed password. Capabilities ('judge'/'admin') are additive grants in
// user_roles — call GrantCapability after creation to elevate an account.
func (s *Store) CreateUser(ctx context.Context, email, password string) (int64, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	result, err := s.db.ExecContext(ctx,
		"INSERT INTO users (email, password_hash) VALUES (?, ?)",
		email, string(hash),
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// DeleteExpiredSessions removes all expired sessions.
func (s *Store) DeleteExpiredSessions(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM sessions WHERE expires_at <= CURRENT_TIMESTAMP")
	return err
}

// UserCapabilities returns the explicit account capabilities a user holds
// ('judge' / 'admin'), sorted for determinism. The 'competitor' baseline is
// implicit and never stored, so a nil/empty slice means competitor-only.
func (s *Store) UserCapabilities(ctx context.Context, userID int64) ([]string, error) {
	return s.q.UserCapabilities(ctx, userID)
}

// GrantCapability grants an account capability to a user. Idempotent: granting
// a capability the user already holds is a no-op and returns no error.
func (s *Store) GrantCapability(ctx context.Context, userID int64, cap string) error {
	return s.q.GrantCapability(ctx, db.GrantCapabilityParams{UserID: userID, Capability: cap})
}

// RevokeCapability revokes an account capability from a user. Idempotent:
// revoking a capability the user does not hold deletes zero rows and returns
// no error.
func (s *Store) RevokeCapability(ctx context.Context, userID int64, cap string) error {
	return s.q.RevokeCapability(ctx, db.RevokeCapabilityParams{UserID: userID, Capability: cap})
}
