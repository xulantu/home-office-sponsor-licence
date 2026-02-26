package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// User represents an application user.
type User struct {
	ID                   int
	Username             string
	PasswordHash         string
	Role                 int
	CreatedAt            time.Time
	InvitationCodeID     *int
	PasswordLastChangedAt *time.Time
}

// InsertUser inserts a new user and returns their ID.
func InsertUser(ctx context.Context, q Querier, u User) (int, error) {
	var id int
	err := q.QueryRow(ctx,
		`INSERT INTO users (username, password_hash, role, invitation_code_id)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id`,
		u.Username, u.PasswordHash, u.Role, u.InvitationCodeID,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert user: %w", err)
	}
	return id, nil
}

// UpdateUserPassword sets a new password hash and records the change time.
func UpdateUserPassword(ctx context.Context, q Querier, userID int, passwordHash string) error {
	_, err := q.Exec(ctx,
		`UPDATE users SET password_hash = $1, password_last_changed_at = NOW() WHERE id = $2`,
		passwordHash, userID,
	)
	if err != nil {
		return fmt.Errorf("update user password: %w", err)
	}
	return nil
}

// FindUserByID looks up a user by their ID.
// Returns the user and true if found, or empty and false if not found.
func FindUserByID(ctx context.Context, q Querier, id int) (User, bool, error) {
	var u User
	err := q.QueryRow(ctx,
		`SELECT id, username, password_hash, role, created_at, invitation_code_id, password_last_changed_at
		 FROM users
		 WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.InvitationCodeID, &u.PasswordLastChangedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, false, nil
	}
	if err != nil {
		return User{}, false, fmt.Errorf("find user by id: %w", err)
	}
	return u, true, nil
}

// FindUserByUsername looks up an active user by username.
// Returns the user and true if found, or empty and false if not found.
func FindUserByUsername(ctx context.Context, q Querier, username string) (User, bool, error) {
	var u User
	err := q.QueryRow(ctx,
		`SELECT id, username, password_hash, role, created_at, invitation_code_id, password_last_changed_at
		 FROM users
		 WHERE username = $1`,
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.InvitationCodeID, &u.PasswordLastChangedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, false, nil
	}
	if err != nil {
		return User{}, false, fmt.Errorf("find user by username: %w", err)
	}
	return u, true, nil
}
