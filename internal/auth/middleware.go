package auth

import (
	"context"
	"net/http"
)

// Middleware provides authentication middleware
type Middleware struct {
	sessionService *SessionService
}

// NewMiddleware creates a new authentication middleware
func NewMiddleware(sessionService *SessionService) *Middleware {
	return &Middleware{sessionService: sessionService}
}

// RequireAuth middleware that requires user to be authenticated
func (m *Middleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, authenticated := m.sessionService.GetCurrentUserID(r)
		if !authenticated {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Add user ID to request context
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetUserFromContext extracts user ID from request context
func GetUserFromContext(r *http.Request) (int64, bool) {
	userID, ok := r.Context().Value("userID").(int64)
	return userID, ok
}

// OptionalAuth middleware that adds user info to context if authenticated
func (m *Middleware) OptionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, authenticated := m.sessionService.GetCurrentUserID(r)
		if authenticated {
			ctx := context.WithValue(r.Context(), "userID", userID)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	}
}
