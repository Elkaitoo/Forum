package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// DB holds the database connection and configuration
type DB struct {
	*sql.DB
	dsn string
}

// Config holds database configuration
type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// DefaultConfig returns a default database configuration
func DefaultConfig() *Config {
	return &Config{
		DSN:             "forum.db",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	}
}

// NewDB creates a new database connection with the given configuration
func NewDB(config *Config) (*DB, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Add SQLite-specific parameters for better performance and safety
	dsn := fmt.Sprintf("%s?_foreign_keys=on&_journal_mode=WAL&_timeout=10000&_synchronous=NORMAL", config.DSN)

	sqlDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{
		DB:  sqlDB,
		dsn: dsn,
	}

	log.Printf("Database connected successfully: %s", config.DSN)
	return db, nil
}

// InitializeDatabase runs the migration script to set up all tables
func (db *DB) InitializeDatabase() error {
	// Read the migrations file
	migrationPath := filepath.Join("internal", "database", "migrations.sql")
	migrationSQL, err := ioutil.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations file: %w", err)
	}

	// Execute the migration
	if _, err := db.Exec(string(migrationSQL)); err != nil {
		return fmt.Errorf("failed to execute migrations: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.DB != nil {
		return db.DB.Close()
	}
	return nil
}

// GetContextWithTimeout creates a context with timeout
func GetContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// HealthCheck verifies the database connection is working
func (db *DB) HealthCheck() error {
	ctx, cancel := GetContextWithTimeout(5 * time.Second)
	defer cancel()

	var result int
	err := db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected health check result: %d", result)
	}

	return nil
}

// CleanExpiredSessions removes expired sessions from the database
func (db *DB) CleanExpiredSessions() error {
	ctx, cancel := GetContextWithTimeout(10 * time.Second)
	defer cancel()

	result, err := db.ExecContext(ctx, "DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP")
	if err != nil {
		return fmt.Errorf("failed to clean expired sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Warning: could not get rows affected count: %v", err)
	} else if rowsAffected > 0 {
		log.Printf("Cleaned %d expired sessions", rowsAffected)
	}

	return nil
}

// GetStats returns basic database statistics
func (db *DB) GetStats() sql.DBStats {
	return db.DB.Stats()
}
