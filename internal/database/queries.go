package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// User CRUD operations

// CreateUser creates a new user in the database
func (db *DB) CreateUser(ctx context.Context, email, username, passwordHash string) (int64, error) {
	query := `
		INSERT INTO users (email, username, password_hash, created_at)
		VALUES (?, ?, ?, ?)
	`

	result, err := db.ExecContext(ctx, query, email, username, passwordHash, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get user ID: %w", err)
	}

	return userID, nil
}

// GetUserByEmail retrieves a user by their email address
func (db *DB) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, username, password_hash, created_at
		FROM users
		WHERE email = ?
	`

	var user User
	err := db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// GetUserByUsername retrieves a user by their username
func (db *DB) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT id, email, username, password_hash, created_at
		FROM users
		WHERE username = ?
	`

	var user User
	err := db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// GetUserByID retrieves a user by their ID
func (db *DB) GetUserByID(ctx context.Context, userID int64) (*User, error) {
	query := `
		SELECT id, email, username, password_hash, created_at
		FROM users
		WHERE id = ?
	`

	var user User
	err := db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// Session CRUD operations

// CreateSession creates a new session for a user
func (db *DB) CreateSession(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO sessions (user_id, token, expires_at, created_at)
		VALUES (?, ?, ?, ?)
	`

	_, err := db.ExecContext(ctx, query, userID, token, expiresAt, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetSessionByToken retrieves a session by its token
func (db *DB) GetSessionByToken(ctx context.Context, token string) (*Session, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM sessions
		WHERE token = ? AND expires_at > CURRENT_TIMESTAMP
	`

	var session Session
	err := db.QueryRowContext(ctx, query, token).Scan(
		&session.ID, &session.UserID, &session.Token, &session.ExpiresAt, &session.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found or expired")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// DeleteSession deletes a session by its token
func (db *DB) DeleteSession(ctx context.Context, token string) error {
	query := `DELETE FROM sessions WHERE token = ?`

	_, err := db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// DeleteUserSessions deletes all sessions for a specific user
func (db *DB) DeleteUserSessions(ctx context.Context, userID int64) error {
	query := `DELETE FROM sessions WHERE user_id = ?`

	_, err := db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	return nil
}

// Category CRUD operations

// CreateCategory creates a new category
func (db *DB) CreateCategory(ctx context.Context, name string) (int64, error) {
	query := `
		INSERT INTO categories (name, created_at)
		VALUES (?, ?)
	`

	result, err := db.ExecContext(ctx, query, name, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("failed to create category: %w", err)
	}

	categoryID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get category ID: %w", err)
	}

	return categoryID, nil
}

// GetCategoryByName retrieves a category by its name
func (db *DB) GetCategoryByName(ctx context.Context, name string) (*Category, error) {
	query := `
		SELECT id, name, created_at
		FROM categories
		WHERE name = ?
	`

	var category Category
	err := db.QueryRowContext(ctx, query, name).Scan(
		&category.ID, &category.Name, &category.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &category, nil
}

// GetCategoryByID retrieves a category by its ID
func (db *DB) GetCategoryByID(ctx context.Context, categoryID int64) (*Category, error) {
	query := `
		SELECT id, name, created_at
		FROM categories
		WHERE id = ?
	`

	var category Category
	err := db.QueryRowContext(ctx, query, categoryID).Scan(
		&category.ID, &category.Name, &category.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &category, nil
}

// GetAllCategories retrieves all categories
func (db *DB) GetAllCategories(ctx context.Context) ([]Category, error) {
	query := `
		SELECT id, name, created_at
		FROM categories
		ORDER BY name ASC
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		err := rows.Scan(&category.ID, &category.Name, &category.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return categories, nil
}

// Utility functions for existing features compatibility

// GetOrCreateCategory gets a category by name or creates it if it doesn't exist
func (db *DB) GetOrCreateCategory(ctx context.Context, name string) (int64, error) {
	// Try to get existing category first
	category, err := db.GetCategoryByName(ctx, name)
	if err == nil {
		return category.ID, nil
	}

	// Category doesn't exist, create it
	categoryID, err := db.CreateCategory(ctx, name)
	if err != nil {
		return 0, fmt.Errorf("failed to create category: %w", err)
	}

	return categoryID, nil
}

// EmailExists checks if an email is already taken
func (db *DB) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT COUNT(1) FROM users WHERE email = ?`

	var count int
	err := db.QueryRowContext(ctx, query, email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return count > 0, nil
}

// UsernameExists checks if a username is already taken
func (db *DB) UsernameExists(ctx context.Context, username string) (bool, error) {
	query := `SELECT COUNT(1) FROM users WHERE username = ?`

	var count int
	err := db.QueryRowContext(ctx, query, username).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}

	return count > 0, nil
}
