package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// PasswordResetToken represents a password reset token row.
type PasswordResetToken struct {
	ID        int
	UserID    int
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// FindLatestPasswordResetToken returns the most recently inserted, unexpired
// token row for the given user. The caller must compare the returned token
// string against the user-supplied value and check token.CreatedAt against
// user.PasswordLastChangedAt to prevent reuse after a successful reset.
// Returns the token and true if found, or empty and false if not found.
func FindLatestPasswordResetToken(ctx context.Context, q Querier, userID int) (PasswordResetToken, bool, error) {
	var t PasswordResetToken
	err := q.QueryRow(ctx,
		`SELECT id, user_id, token, created_at, expires_at
		 FROM password_reset_tokens
		 WHERE user_id = $1 AND expires_at > NOW()
		 ORDER BY id DESC
		 LIMIT 1`,
		userID,
	).Scan(&t.ID, &t.UserID, &t.Token, &t.CreatedAt, &t.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return PasswordResetToken{}, false, nil
	}
	if err != nil {
		return PasswordResetToken{}, false, fmt.Errorf("find latest password reset token: %w", err)
	}
	return t, true, nil
}

// CreatePasswordResetToken inserts a new password reset token for the given user.
func CreatePasswordResetToken(ctx context.Context, q Querier, userID int, token string, expiry time.Duration) error {
	_, err := q.Exec(ctx,
		`INSERT INTO password_reset_tokens (user_id, token, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, token, time.Now().Add(expiry),
	)
	if err != nil {
		return fmt.Errorf("create password reset token: %w", err)
	}
	return nil
}
