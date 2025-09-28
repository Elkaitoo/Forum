package auth

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// SessionService handles user sessions
type SessionService struct {
	db *sql.DB
}

// NewSessionService creates a new session service
func NewSessionService(db *sql.DB) *SessionService {
	return &SessionService{db: db}
}

// CreateSession creates a new session for a user
func (s *SessionService) CreateSession(userID int64) (string, error) {
	// Generate session token
	sessionToken := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour) // 24 hour session

	// Delete any existing sessions for this user (single session per user)
	_, err := s.db.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	if err != nil {
		return "", fmt.Errorf("failed to clean existing sessions: %w", err)
	}

	// Insert new session
	_, err = s.db.Exec(
		"INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)",
		sessionToken, userID, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return sessionToken, nil
}

// ValidateSession checks if a session token is valid and returns user ID
func (s *SessionService) ValidateSession(token string) (int64, error) {
	var userID int64
	var expiresAt time.Time

	err := s.db.QueryRow(
		"SELECT user_id, expires_at FROM sessions WHERE token = ?",
		token,
	).Scan(&userID, &expiresAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("invalid session")
		}
		return 0, fmt.Errorf("failed to validate session: %w", err)
	}

	// Check if session has expired
	if time.Now().After(expiresAt) {
		// Clean up expired session
		s.db.Exec("DELETE FROM sessions WHERE token = ?", token)
		return 0, fmt.Errorf("session expired")
	}

	return userID, nil
}

// DeleteSession removes a session (logout)
func (s *SessionService) DeleteSession(token string) error {
	_, err := s.db.Exec("DELETE FROM sessions WHERE token = ?", token)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// GetCurrentUserID extracts user ID from session cookie
func (s *SessionService) GetCurrentUserID(r *http.Request) (int64, bool) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return 0, false
	}

	userID, err := s.ValidateSession(cookie.Value)
	if err != nil {
		return 0, false
	}

	return userID, true
}

// SetSessionCookie sets the session cookie in the response
func (s *SessionService) SetSessionCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24 hours
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

// ClearSessionCookie removes the session cookie
func (s *SessionService) ClearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}
