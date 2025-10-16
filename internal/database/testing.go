// Package database provides SQLite database connection management
// with optimized PRAGMAs for performance.
package database

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/jmoiron/sqlx"
)

// NewTestDB creates an in-memory SQLite database for testing purposes.
// The database is automatically cleaned up when the test completes.
//
// This function:
//   - Creates an in-memory SQLite database (":memory:")
//   - Applies all migrations from migrations/sqlite directory
//   - Registers automatic cleanup via t.Cleanup()
//
// Parameters:
//   - t: Testing context for cleanup registration and error reporting
//
// Returns:
//   - *sqlx.DB: Initialized in-memory database connection with all migrations applied
//
// Example:
//
//	func TestMyFeature(t *testing.T) {
//	    db := NewTestDB(t)
//	    // Use db for testing...
//	}
func NewTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	// Create in-memory database
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewTestDB: failed to create in-memory database: %v", err)
	}

	// Apply all migrations
	if err := ApplyMigrations(db); err != nil {
		db.Close()
		t.Fatalf("NewTestDB: failed to apply migrations: %v", err)
	}

	// Register cleanup
	t.Cleanup(func() {
		if err := Close(db); err != nil {
			t.Errorf("NewTestDB cleanup: failed to close database: %v", err)
		}
	})

	return db
}

// ApplyMigrations applies all SQL migration files from the migrations/sqlite directory
// to the given database connection in sorted order (001, 002, 003, etc.).
//
// Migration files are expected to be:
//   - Located in ../../migrations/sqlite/ relative to this package
//   - Named with numeric prefixes (e.g., 001_inv_types.sql, 002_inv_groups.sql)
//   - Idempotent (using CREATE TABLE IF NOT EXISTS, CREATE INDEX IF NOT EXISTS)
//
// Parameters:
//   - db: Database connection to apply migrations to
//
// Returns:
//   - error: Any error encountered while reading or executing migrations
//
// Example:
//
//	db, _ := NewDB(":memory:")
//	if err := ApplyMigrations(db); err != nil {
//	    log.Fatalf("Failed to apply migrations: %v", err)
//	}
func ApplyMigrations(db *sqlx.DB) error {
	// Determine migrations directory path
	migrationsDir := filepath.Join("..", "..", "migrations", "sqlite")

	// Read all migration files
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Filter and sort SQL files
	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Apply each migration in order
	for _, filename := range migrationFiles {
		migrationPath := filepath.Join(migrationsDir, filename)
		migrationSQL, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		// Execute migration
		if _, err := db.Exec(string(migrationSQL)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}
	}

	return nil
}
