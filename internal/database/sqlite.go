// Package database provides SQLite database connection management
// with optimized PRAGMAs for performance.
package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// NewDB creates and initializes a new SQLite database connection
// with optimized PRAGMAs for performance according to ADR-001 and ADR-002.
//
// Parameters:
//   - path: File path to the SQLite database file. Use ":memory:" for in-memory databases.
//
// Returns:
//   - *sqlx.DB: Initialized database connection
//   - error: Any error encountered during connection setup
func NewDB(path string) (*sqlx.DB, error) {
	// Open database connection
	db, err := sqlx.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool limits
	db.SetMaxOpenConns(1) // SQLite works best with single writer
	db.SetMaxIdleConns(1)

	// Apply performance PRAGMAs
	if err := applyPragmas(db, path); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to apply PRAGMAs: %w", err)
	}

	// Verify connection is working
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// applyPragmas applies SQLite performance optimizations as defined in ADR-001.
//
// PRAGMAs applied:
//   - journal_mode = WAL: Write-Ahead Logging for better concurrency (skipped for :memory: databases)
//   - synchronous = NORMAL: Balance between safety and performance
//   - foreign_keys = ON: Enforce referential integrity
//   - cache_size = -64000: 64MB cache for better performance
//   - temp_store = MEMORY: Store temporary tables in memory
//   - busy_timeout = 5000: Wait up to 5 seconds if database is locked
func applyPragmas(db *sqlx.DB, path string) error {
	pragmas := []string{
		"PRAGMA synchronous = NORMAL",
		"PRAGMA foreign_keys = ON",
		"PRAGMA cache_size = -64000",
		"PRAGMA temp_store = MEMORY",
		"PRAGMA busy_timeout = 5000",
	}

	// Only use WAL mode for file-based databases
	// WAL mode can cause issues with :memory: databases and race detector
	if path != ":memory:" {
		pragmas = append([]string{"PRAGMA journal_mode = WAL"}, pragmas...)
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return fmt.Errorf("failed to execute '%s': %w", pragma, err)
		}
	}

	return nil
}

// Close gracefully closes the database connection.
//
// Parameters:
//   - db: Database connection to close
//
// Returns:
//   - error: Any error encountered during closure
func Close(db *sqlx.DB) error {
	if db == nil {
		return nil
	}
	if err := db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}
