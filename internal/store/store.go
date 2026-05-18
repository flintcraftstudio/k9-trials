package store

import (
	"context"
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Store wraps the database connection and provides query methods.
type Store struct {
	db *sql.DB
}

// New creates a new Store.
func New(db *sql.DB) *Store {
	return &Store{db: db}
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

// CreateUser inserts a new user with a bcrypt-hashed password.
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
