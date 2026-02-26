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
	insertErr  error
}

func (f *fakeUserStore) InsertUser(_ context.Context, u database.User) (int, error) {
	if f.insertErr != nil {
		return 0, f.insertErr
	}
	id := len(f.byID) + 1
	u.ID = id
	f.byID[id] = u
	f.byUsername[u.Username] = u
	return id, nil
}

func (f *fakeUserStore) UpdateUserPassword(_ context.Context, userID int, hash string) error {
	if u, ok := f.byID[userID]; ok {
		u.PasswordHash = hash
		f.byID[userID] = u
	}
	return nil
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

type fakeInvitationCodeStore struct {
	code      database.InvitationCode
	codeFound bool
	codeErr   error
	uses      int
	usesErr   error
}

func (f *fakeInvitationCodeStore) FindInvitationCode(_ context.Context, _ string) (database.InvitationCode, bool, error) {
	return f.code, f.codeFound, f.codeErr
}

func (f *fakeInvitationCodeStore) CountInvitationCodeUses(_ context.Context, _ int) (int, error) {
	return f.uses, f.usesErr
}

type fakePasswordResetStore struct {
	token      database.PasswordResetToken
	tokenFound bool
	tokenErr   error
	createErr  error
}

func (f *fakePasswordResetStore) FindLatestPasswordResetToken(_ context.Context, _ int) (database.PasswordResetToken, bool, error) {
	return f.token, f.tokenFound, f.tokenErr
}

func (f *fakePasswordResetStore) CreatePasswordResetToken(_ context.Context, _ int, _ string, _ time.Duration) error {
	return f.createErr
}

func TestService_Register(t *testing.T) {
	code := database.InvitationCode{ID: 1, Count: 5}
	newUsers := func() *fakeUserStore {
		return &fakeUserStore{byUsername: map[string]database.User{}, byID: map[int]database.User{}}
	}
	tests := []struct {
		name    string
		inv     *fakeInvitationCodeStore
		sesErr  error
		wantErr bool
	}{
		{"success", &fakeInvitationCodeStore{code: code, codeFound: true, uses: 0}, nil, false},
		{"code not found", &fakeInvitationCodeStore{codeFound: false}, nil, true},
		{"code at limit", &fakeInvitationCodeStore{code: code, codeFound: true, uses: 5}, nil, true},
		{"session error", &fakeInvitationCodeStore{code: code, codeFound: true}, errors.New("db down"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ses := &fakeSessionStore{sessions: map[string]database.Session{}, nextToken: "tok", createErr: tt.sesErr}
			svc := NewService(newUsers(), ses, tt.inv, &fakePasswordResetStore{})
			tok, err := svc.Register(context.Background(), "alice", "password123", "invite")
			if (err != nil) != tt.wantErr {
				t.Fatalf("Register() err = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil && tok == "" {
				t.Error("Register() returned empty token on success")
			}
		})
	}
}

func TestService_ResetPassword(t *testing.T) {
	now := time.Now()
	past := now.Add(-time.Hour)
	hash, _ := bcrypt.GenerateFromPassword([]byte("old"), bcrypt.MinCost)
	alice := database.User{ID: 1, Username: "alice", PasswordHash: string(hash)}
	aliceChanged := database.User{ID: 1, Username: "alice", PasswordHash: string(hash), PasswordLastChangedAt: &now}
	validTok := database.PasswordResetToken{UserID: 1, Token: "tok123", CreatedAt: now, ExpiresAt: now.Add(time.Hour)}
	staleTok := database.PasswordResetToken{UserID: 1, Token: "tok123", CreatedAt: past, ExpiresAt: now.Add(time.Hour)}

	tests := []struct {
		name       string
		users      *fakeUserStore
		resetStore *fakePasswordResetStore
		token      string
		wantErr    bool
	}{
		{"success",
			&fakeUserStore{byUsername: map[string]database.User{"alice": alice}, byID: map[int]database.User{1: alice}},
			&fakePasswordResetStore{token: validTok, tokenFound: true}, "tok123", false},
		{"user not found",
			&fakeUserStore{byUsername: map[string]database.User{}, byID: map[int]database.User{}},
			&fakePasswordResetStore{}, "tok123", true},
		{"token not found",
			&fakeUserStore{byUsername: map[string]database.User{"alice": alice}, byID: map[int]database.User{1: alice}},
			&fakePasswordResetStore{tokenFound: false}, "tok123", true},
		{"token mismatch",
			&fakeUserStore{byUsername: map[string]database.User{"alice": alice}, byID: map[int]database.User{1: alice}},
			&fakePasswordResetStore{token: validTok, tokenFound: true}, "wrong", true},
		{"token predates password change",
			&fakeUserStore{byUsername: map[string]database.User{"alice": aliceChanged}, byID: map[int]database.User{1: aliceChanged}},
			&fakePasswordResetStore{token: staleTok, tokenFound: true}, "tok123", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ses := &fakeSessionStore{sessions: map[string]database.Session{}}
			svc := NewService(tt.users, ses, &fakeInvitationCodeStore{}, tt.resetStore)
			err := svc.ResetPassword(context.Background(), "alice", "newpass", tt.token)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ResetPassword() err = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
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
			svc := NewService(tt.users, tt.sessions, &fakeInvitationCodeStore{}, &fakePasswordResetStore{})
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
			svc := NewService(&fakeUserStore{byUsername: map[string]database.User{}, byID: map[int]database.User{}}, tt.sessions, &fakeInvitationCodeStore{}, &fakePasswordResetStore{})
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
			svc := NewService(tt.users, tt.sessions, &fakeInvitationCodeStore{}, &fakePasswordResetStore{})
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
