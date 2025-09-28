package auth

import (
	"database/sql"
	"fmt"
	"net/mail"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// AuthService handles user authentication
type AuthService struct {
	db *sql.DB
}

// NewAuthService creates a new authentication service
func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{db: db}
}

// RegisterUser creates a new user account
func (a *AuthService) RegisterUser(email, username, password string) error {
	// Validate email format
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email format")
	}

	// Validate input
	if strings.TrimSpace(username) == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if strings.Contains(username, " ") {
		return fmt.Errorf("username cannot contain spaces")
	}
	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}

	// Check if email already exists
	var exists bool
	err := a.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", email).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check email availability: %w", err)
	}
	if exists {
		return fmt.Errorf("email already registered")
	}

	// Check if username already exists
	err = a.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", username).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check username availability: %w", err)
	}
	if exists {
		return fmt.Errorf("username already taken")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Insert user
	_, err = a.db.Exec(
		"INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)",
		username, email, string(hashedPassword),
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// AuthenticateUser validates user credentials and returns user ID
func (a *AuthService) AuthenticateUser(email, password string) (int64, error) {
	var userID int64
	var hashedPassword string

	err := a.db.QueryRow(
		"SELECT id, password_hash FROM users WHERE email = ?",
		email,
	).Scan(&userID, &hashedPassword)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("invalid email or password")
		}
		return 0, fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return 0, fmt.Errorf("invalid email or password")
	}

	return userID, nil
}

// GetUserByID retrieves user information by ID
func (a *AuthService) GetUserByID(userID int64) (*User, error) {
	var user User
	err := a.db.QueryRow(
		"SELECT id, username, email, created_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// User represents a user in the system
type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}
