package database

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// Session represents an authenticated user session.
type Session struct {
	Token     string
	UserID    int
	ExpiresAt time.Time
	CreatedAt time.Time
}

// CreateSession generates a random token, inserts a session, and returns the token.
func CreateSession(ctx context.Context, q Querier, userID int, expiry time.Duration) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("create session: generate token: %w", err)
	}
	token := hex.EncodeToString(b)
	expiresAt := time.Now().Add(expiry)

	_, err := q.Exec(ctx,
		`INSERT INTO sessions (token, user_id, expires_at) VALUES ($1, $2, $3)`,
		token, userID, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	return token, nil
}

// FindSession looks up a session by token, returning it only if not expired.
// Returns the session and true if found and valid, or empty and false if not.
func FindSession(ctx context.Context, q Querier, token string) (Session, bool, error) {
	var s Session
	err := q.QueryRow(ctx,
		`SELECT token, user_id, expires_at, created_at
		 FROM sessions
		 WHERE token = $1 AND expires_at > NOW()`,
		token,
	).Scan(&s.Token, &s.UserID, &s.ExpiresAt, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, false, nil
	}
	if err != nil {
		return Session{}, false, fmt.Errorf("find session: %w", err)
	}
	return s, true, nil
}

// DeleteSession removes a session by token (logout).
func DeleteSession(ctx context.Context, q Querier, token string) error {
	_, err := q.Exec(ctx,
		`DELETE FROM sessions WHERE token = $1`,
		token,
	)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// ExtendSession resets the expiry of a session to now + expiry duration.
func ExtendSession(ctx context.Context, q Querier, token string, expiry time.Duration) error {
	_, err := q.Exec(ctx,
		`UPDATE sessions SET expires_at = $1 WHERE token = $2`,
		time.Now().Add(expiry), token,
	)
	if err != nil {
		return fmt.Errorf("extend session: %w", err)
	}
	return nil
}

// DeleteExpiredSessions removes all expired sessions.
func DeleteExpiredSessions(ctx context.Context, q Querier) error {
	_, err := q.Exec(ctx,
		`DELETE FROM sessions WHERE expires_at <= NOW()`,
	)
	if err != nil {
		return fmt.Errorf("delete expired sessions: %w", err)
	}
	return nil
}
