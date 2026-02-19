package auth

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"sponsor-tracker/internal/database"
)

const SessionDuration = 15 * time.Minute

// UserStore is the subset of database operations needed by the auth service.
type UserStore interface {
	FindUserByID(ctx context.Context, id int) (database.User, bool, error)
	FindUserByUsername(ctx context.Context, username string) (database.User, bool, error)
}

// SessionStore is the subset of database operations needed by the auth service.
type SessionStore interface {
	CreateSession(ctx context.Context, userID int, expiry time.Duration) (string, error)
	FindSession(ctx context.Context, token string) (database.Session, bool, error)
	DeleteSession(ctx context.Context, token string) error
	ExtendSession(ctx context.Context, token string, expiry time.Duration) error
}

// Service handles authentication logic.
type Service struct {
	users    UserStore
	sessions SessionStore
}

// NewService constructs an auth Service.
func NewService(users UserStore, sessions SessionStore) *Service {
	return &Service{users: users, sessions: sessions}
}

// Login verifies credentials and returns a session token on success.
func (s *Service) Login(ctx context.Context, username, password string) (string, error) {
	user, found, err := s.users.FindUserByUsername(ctx, username)
	if err != nil {
		return "", fmt.Errorf("login: %w", err)
	}
	if !found {
		return "", fmt.Errorf("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", fmt.Errorf("invalid credentials")
	}
	token, err := s.sessions.CreateSession(ctx, user.ID, SessionDuration)
	if err != nil {
		return "", fmt.Errorf("login: %w", err)
	}
	return token, nil
}

// Logout deletes the session identified by token.
func (s *Service) Logout(ctx context.Context, token string) error {
	if err := s.sessions.DeleteSession(ctx, token); err != nil {
		return fmt.Errorf("logout: %w", err)
	}
	return nil
}

// Authenticate validates the token, extends the session, and returns the associated user.
func (s *Service) Authenticate(ctx context.Context, token string) (database.User, error) {
	session, found, err := s.sessions.FindSession(ctx, token)
	if err != nil {
		return database.User{}, fmt.Errorf("authenticate: %w", err)
	}
	if !found {
		return database.User{}, fmt.Errorf("invalid or expired session")
	}
	if err := s.sessions.ExtendSession(ctx, token, SessionDuration); err != nil {
		return database.User{}, fmt.Errorf("authenticate: %w", err)
	}
	user, found, err := s.users.FindUserByID(ctx, session.UserID)
	if err != nil {
		return database.User{}, fmt.Errorf("authenticate: %w", err)
	}
	if !found {
		return database.User{}, fmt.Errorf("authenticate: user not found")
	}
	return user, nil
}
