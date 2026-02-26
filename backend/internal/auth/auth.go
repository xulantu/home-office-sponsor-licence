package auth

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"sponsor-tracker/internal/database"
)

const SessionDuration = 15 * time.Minute
const ResetTokenDuration = 1 * time.Hour

// UserStore is the subset of database operations needed by the auth service.
type UserStore interface {
	FindUserByID(ctx context.Context, id int) (database.User, bool, error)
	FindUserByUsername(ctx context.Context, username string) (database.User, bool, error)
	InsertUser(ctx context.Context, u database.User) (int, error)
	UpdateUserPassword(ctx context.Context, userID int, passwordHash string) error
}

// SessionStore is the subset of database operations needed by the auth service.
type SessionStore interface {
	CreateSession(ctx context.Context, userID int, expiry time.Duration) (string, error)
	FindSession(ctx context.Context, token string) (database.Session, bool, error)
	DeleteSession(ctx context.Context, token string) error
	ExtendSession(ctx context.Context, token string, expiry time.Duration) error
}

// InvitationCodeStore provides access to invitation codes.
type InvitationCodeStore interface {
	FindInvitationCode(ctx context.Context, code string) (database.InvitationCode, bool, error)
	CountInvitationCodeUses(ctx context.Context, codeID int) (int, error)
}

// PasswordResetStore provides access to password reset tokens.
type PasswordResetStore interface {
	FindLatestPasswordResetToken(ctx context.Context, userID int) (database.PasswordResetToken, bool, error)
	CreatePasswordResetToken(ctx context.Context, userID int, token string, expiry time.Duration) error
}

// Service handles authentication logic.
type Service struct {
	users       UserStore
	sessions    SessionStore
	invCodes    InvitationCodeStore
	resetTokens PasswordResetStore
}

// NewService constructs an auth Service.
func NewService(users UserStore, sessions SessionStore, invCodes InvitationCodeStore, resetTokens PasswordResetStore) *Service {
	return &Service{users: users, sessions: sessions, invCodes: invCodes, resetTokens: resetTokens}
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

// Register validates the invitation code, creates a new viewer account, and returns a session token.
func (s *Service) Register(ctx context.Context, username, password, invitationCode string) (string, error) {
	code, found, err := s.invCodes.FindInvitationCode(ctx, invitationCode)
	if err != nil {
		return "", fmt.Errorf("register: %w", err)
	}
	if !found {
		return "", fmt.Errorf("invalid or expired invitation code")
	}
	uses, err := s.invCodes.CountInvitationCodeUses(ctx, code.ID)
	if err != nil {
		return "", fmt.Errorf("register: %w", err)
	}
	if uses >= code.Count {
		return "", fmt.Errorf("invitation code has reached its usage limit")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", fmt.Errorf("register: %w", err)
	}
	userID, err := s.users.InsertUser(ctx, database.User{
		Username:         username,
		PasswordHash:     string(hash),
		Role:             50,
		InvitationCodeID: &code.ID,
	})
	if err != nil {
		return "", fmt.Errorf("register: %w", err)
	}
	token, err := s.sessions.CreateSession(ctx, userID, SessionDuration)
	if err != nil {
		return "", fmt.Errorf("register: %w", err)
	}
	return token, nil
}

// ResetPassword validates the reset token and updates the user's password.
// All validation failures return the same error to prevent user enumeration.
func (s *Service) ResetPassword(ctx context.Context, username, newPassword, resetToken string) error {
	const invalidRequest = "invalid request"
	user, found, err := s.users.FindUserByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("reset password: %w", err)
	}
	if !found {
		return fmt.Errorf(invalidRequest)
	}
	latest, found, err := s.resetTokens.FindLatestPasswordResetToken(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("reset password: %w", err)
	}
	if !found {
		return fmt.Errorf(invalidRequest)
	}
	if latest.Token != resetToken {
		return fmt.Errorf(invalidRequest)
	}
	if user.PasswordLastChangedAt != nil && !latest.CreatedAt.After(*user.PasswordLastChangedAt) {
		return fmt.Errorf(invalidRequest)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return fmt.Errorf("reset password: %w", err)
	}
	if err := s.users.UpdateUserPassword(ctx, user.ID, string(hash)); err != nil {
		return fmt.Errorf("reset password: %w", err)
	}
	return nil
}
