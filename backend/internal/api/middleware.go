package api

import (
	"context"
	"net/http"

	"sponsor-tracker/internal/database"
)

type contextKey string

const userContextKey contextKey = "user"

// UserFromContext retrieves the authenticated user from the request context.
func UserFromContext(ctx context.Context) (database.User, bool) {
	u, ok := ctx.Value(userContextKey).(database.User)
	return u, ok
}

// requireRole returns a middleware that allows only users with role <= minRole.
// Lower role numbers have more permissions (10 = admin, 50 = viewer).
func (s *Server) requireRole(minRole int, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Error(w, "unauthorised", http.StatusUnauthorized)
			return
		}
		user, err := s.auth.Authenticate(r.Context(), cookie.Value)
		if err != nil {
			http.Error(w, "unauthorised", http.StatusUnauthorized)
			return
		}
		if user.Role > minRole {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next(w, r.WithContext(ctx))
	}
}
