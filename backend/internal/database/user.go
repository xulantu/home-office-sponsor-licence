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
	ID           int
	Username     string
	PasswordHash string
	Role         int
	CreatedAt    time.Time
}

// InsertUser inserts a new user and returns their ID.
func InsertUser(ctx context.Context, q Querier, u User) (int, error) {
	var id int
	err := q.QueryRow(ctx,
		`INSERT INTO users (username, password_hash, role)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		u.Username, u.PasswordHash, u.Role,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert user: %w", err)
	}
	return id, nil
}

// FindUserByUsername looks up an active user by username.
// Returns the user and true if found, or empty and false if not found.
func FindUserByUsername(ctx context.Context, q Querier, username string) (User, bool, error) {
	var u User
	err := q.QueryRow(ctx,
		`SELECT id, username, password_hash, role, created_at
		 FROM users
		 WHERE username = $1`,
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, false, nil
	}
	if err != nil {
		return User{}, false, fmt.Errorf("find user by username: %w", err)
	}
	return u, true, nil
}
