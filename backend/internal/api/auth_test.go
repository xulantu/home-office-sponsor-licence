package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sponsor-tracker/internal/database"
)

type fakeAuth struct {
	loginToken string
	loginErr   error
	logoutErr  error
	authUser   database.User
	authErr    error
}

func (f *fakeAuth) Login(_ context.Context, _, _ string) (string, error) {
	return f.loginToken, f.loginErr
}

func (f *fakeAuth) Logout(_ context.Context, _ string) error {
	return f.logoutErr
}

func (f *fakeAuth) Authenticate(_ context.Context, _ string) (database.User, error) {
	return f.authUser, f.authErr
}

type fakeData struct{}

func (f *fakeData) GetAll(_ context.Context, _, _ int, _ string) (*database.DataResponse, error) {
	return &database.DataResponse{}, nil
}

func newTestServer(a *fakeAuth) *Server {
	return NewServer(nil, &fakeData{}, a)
}

func TestHandleLogin(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		auth       *fakeAuth
		wantCode   int
		wantCookie bool
	}{
		{
			name:       "success",
			body:       `{"username":"alice","password":"correct"}`,
			auth:       &fakeAuth{loginToken: "tok"},
			wantCode:   http.StatusNoContent,
			wantCookie: true,
		},
		{
			name:     "bad credentials",
			body:     `{"username":"alice","password":"wrong"}`,
			auth:     &fakeAuth{loginErr: errors.New("invalid credentials")},
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "invalid JSON",
			body:     `not-json`,
			auth:     &fakeAuth{},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer(tt.auth)
			r := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			s.handleLogin(w, r)

			if w.Code != tt.wantCode {
				t.Errorf("status = %d, want %d", w.Code, tt.wantCode)
			}
			hasCookie := w.Header().Get("Set-Cookie") != ""
			if hasCookie != tt.wantCookie {
				t.Errorf("hasCookie = %v, want %v", hasCookie, tt.wantCookie)
			}
		})
	}
}

func TestHandleLogout(t *testing.T) {
	tests := []struct {
		name     string
		cookie   string
		auth     *fakeAuth
		wantCode int
	}{
		{
			name:     "no cookie",
			auth:     &fakeAuth{},
			wantCode: http.StatusNoContent,
		},
		{
			name:     "with valid cookie",
			cookie:   "session_token=tok",
			auth:     &fakeAuth{},
			wantCode: http.StatusNoContent,
		},
		{
			name:     "logout error",
			cookie:   "session_token=tok",
			auth:     &fakeAuth{logoutErr: errors.New("db down")},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer(tt.auth)
			r := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
			if tt.cookie != "" {
				r.Header.Set("Cookie", tt.cookie)
			}
			w := httptest.NewRecorder()
			s.handleLogout(w, r)

			if w.Code != tt.wantCode {
				t.Errorf("status = %d, want %d", w.Code, tt.wantCode)
			}
		})
	}
}

func TestHandleMe(t *testing.T) {
	alice := database.User{ID: 1, Username: "alice", Role: 10}

	tests := []struct {
		name     string
		cookie   string
		auth     *fakeAuth
		wantCode int
		wantUser *database.User
	}{
		{
			name:     "no cookie",
			auth:     &fakeAuth{},
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "invalid token",
			cookie:   "session_token=bad",
			auth:     &fakeAuth{authErr: errors.New("not found")},
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "valid token",
			cookie:   "session_token=tok",
			auth:     &fakeAuth{authUser: alice},
			wantCode: http.StatusOK,
			wantUser: &alice,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer(tt.auth)
			r := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
			if tt.cookie != "" {
				r.Header.Set("Cookie", tt.cookie)
			}
			w := httptest.NewRecorder()
			s.handleMe(w, r)

			if w.Code != tt.wantCode {
				t.Errorf("status = %d, want %d", w.Code, tt.wantCode)
			}
			if tt.wantUser != nil {
				var got struct {
					ID       int    `json:"id"`
					Username string `json:"username"`
					Role     int    `json:"role"`
				}
				if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if got.ID != tt.wantUser.ID || got.Username != tt.wantUser.Username || got.Role != tt.wantUser.Role {
					t.Errorf("got %+v, want %+v", got, *tt.wantUser)
				}
			}
		})
	}
}

func TestRequireRole(t *testing.T) {
	admin := database.User{ID: 1, Username: "alice", Role: 10}
	viewer := database.User{ID: 2, Username: "bob", Role: 50}

	tests := []struct {
		name     string
		cookie   string
		auth     *fakeAuth
		minRole  int
		wantCode int
	}{
		{
			name:     "no cookie",
			auth:     &fakeAuth{},
			minRole:  10,
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "invalid session",
			cookie:   "session_token=bad",
			auth:     &fakeAuth{authErr: errors.New("not found")},
			minRole:  10,
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "viewer forbidden on admin route",
			cookie:   "session_token=tok",
			auth:     &fakeAuth{authUser: viewer},
			minRole:  10,
			wantCode: http.StatusForbidden,
		},
		{
			name:     "admin allowed on admin route",
			cookie:   "session_token=tok",
			auth:     &fakeAuth{authUser: admin},
			minRole:  10,
			wantCode: http.StatusOK,
		},
		{
			name:     "admin allowed on viewer route",
			cookie:   "session_token=tok",
			auth:     &fakeAuth{authUser: admin},
			minRole:  50,
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer(tt.auth)
			handler := s.requireRole(tt.minRole, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.cookie != "" {
				r.Header.Set("Cookie", tt.cookie)
			}
			w := httptest.NewRecorder()
			handler(w, r)

			if w.Code != tt.wantCode {
				t.Errorf("status = %d, want %d", w.Code, tt.wantCode)
			}
		})
	}
}
