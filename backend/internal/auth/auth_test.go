package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"sponsor-tracker/internal/database"
)

// fakeUserStore is an in-memory UserStore for testing.
type fakeUserStore struct {
	byUsername map[string]database.User
	byID       map[int]database.User
	err        error
}

func (f *fakeUserStore) FindUserByUsername(_ context.Context, username string) (database.User, bool, error) {
	if f.err != nil {
		return database.User{}, false, f.err
	}
	u, ok := f.byUsername[username]
	return u, ok, nil
}

func (f *fakeUserStore) FindUserByID(_ context.Context, id int) (database.User, bool, error) {
	if f.err != nil {
		return database.User{}, false, f.err
	}
	u, ok := f.byID[id]
	return u, ok, nil
}

// fakeSessionStore is an in-memory SessionStore for testing.
type fakeSessionStore struct {
	sessions  map[string]database.Session
	nextToken string
	createErr error
	extendErr error
	deleteErr error
}

func (f *fakeSessionStore) CreateSession(_ context.Context, userID int, expiry time.Duration) (string, error) {
	if f.createErr != nil {
		return "", f.createErr
	}
	s := database.Session{Token: f.nextToken, UserID: userID, ExpiresAt: time.Now().Add(expiry)}
	f.sessions[f.nextToken] = s
	return f.nextToken, nil
}

func (f *fakeSessionStore) FindSession(_ context.Context, token string) (database.Session, bool, error) {
	s, ok := f.sessions[token]
	return s, ok, nil
}

func (f *fakeSessionStore) DeleteSession(_ context.Context, token string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	delete(f.sessions, token)
	return nil
}

func (f *fakeSessionStore) ExtendSession(_ context.Context, token string, expiry time.Duration) error {
	if f.extendErr != nil {
		return f.extendErr
	}
	if s, ok := f.sessions[token]; ok {
		s.ExpiresAt = time.Now().Add(expiry)
		f.sessions[token] = s
	}
	return nil
}

func TestService_Login(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.MinCost)
	alice := database.User{ID: 1, Username: "alice", PasswordHash: string(hash), Role: 10}

	tests := []struct {
		name      string
		users     *fakeUserStore
		sessions  *fakeSessionStore
		username  string
		password  string
		wantErr   bool
	}{
		{
			name:     "correct credentials",
			users:    &fakeUserStore{byUsername: map[string]database.User{"alice": alice}},
			sessions: &fakeSessionStore{sessions: map[string]database.Session{}, nextToken: "tok"},
			username: "alice", password: "correct",
		},
		{
			name:     "user not found",
			users:    &fakeUserStore{byUsername: map[string]database.User{}},
			sessions: &fakeSessionStore{sessions: map[string]database.Session{}},
			username: "nobody", password: "x",
			wantErr: true,
		},
		{
			name:     "wrong password",
			users:    &fakeUserStore{byUsername: map[string]database.User{"alice": alice}},
			sessions: &fakeSessionStore{sessions: map[string]database.Session{}},
			username: "alice", password: "wrong",
			wantErr: true,
		},
		{
			name:     "user store error",
			users:    &fakeUserStore{err: errors.New("db down")},
			sessions: &fakeSessionStore{sessions: map[string]database.Session{}},
			username: "alice", password: "correct",
			wantErr: true,
		},
		{
			name:     "session store error",
			users:    &fakeUserStore{byUsername: map[string]database.User{"alice": alice}},
			sessions: &fakeSessionStore{sessions: map[string]database.Session{}, createErr: errors.New("session db down")},
			username: "alice", password: "correct",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.users, tt.sessions)
			token, err := svc.Login(context.Background(), tt.username, tt.password)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Login() err = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil && token == "" {
				t.Error("Login() returned empty token on success")
			}
		})
	}
}

func TestService_Logout(t *testing.T) {
	const tok = "session-token"

	tests := []struct {
		name     string
		sessions *fakeSessionStore
		wantErr  bool
	}{
		{
			name:     "success",
			sessions: &fakeSessionStore{sessions: map[string]database.Session{tok: {Token: tok}}},
		},
		{
			name:     "delete error",
			sessions: &fakeSessionStore{sessions: map[string]database.Session{}, deleteErr: errors.New("db down")},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(&fakeUserStore{byUsername: map[string]database.User{}, byID: map[int]database.User{}}, tt.sessions)
			err := svc.Logout(context.Background(), tok)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Logout() err = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Authenticate(t *testing.T) {
	const tok = "session-token"
	alice := database.User{ID: 1, Username: "alice", Role: 10}
	session := database.Session{Token: tok, UserID: 1, ExpiresAt: time.Now().Add(time.Minute)}

	tests := []struct {
		name     string
		users    *fakeUserStore
		sessions *fakeSessionStore
		wantUser database.User
		wantErr  bool
	}{
		{
			name:     "valid session",
			users:    &fakeUserStore{byID: map[int]database.User{1: alice}},
			sessions: &fakeSessionStore{sessions: map[string]database.Session{tok: session}},
			wantUser: alice,
		},
		{
			name:     "session not found",
			users:    &fakeUserStore{byID: map[int]database.User{}},
			sessions: &fakeSessionStore{sessions: map[string]database.Session{}},
			wantErr:  true,
		},
		{
			name:     "extend error",
			users:    &fakeUserStore{byID: map[int]database.User{1: alice}},
			sessions: &fakeSessionStore{sessions: map[string]database.Session{tok: session}, extendErr: errors.New("db down")},
			wantErr:  true,
		},
		{
			name:     "user not found after session",
			users:    &fakeUserStore{byID: map[int]database.User{}},
			sessions: &fakeSessionStore{sessions: map[string]database.Session{tok: session}},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.users, tt.sessions)
			user, err := svc.Authenticate(context.Background(), tok)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Authenticate() err = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil && user != tt.wantUser {
				t.Errorf("Authenticate() user = %v, want %v", user, tt.wantUser)
			}
		})
	}
}
