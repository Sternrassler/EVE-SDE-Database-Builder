// Package database provides SQLite database connection management
// for the EVE SDE Database Builder.
//
// This package implements the database layer according to ADR-001 (SQLite-Only Approach)
// and ADR-002 (Database Layer Design). It provides optimized SQLite connections with
// performance-tuned PRAGMAs for efficient data import and querying.
//
// Features:
//   - WAL (Write-Ahead Logging) mode for better concurrency
//   - Optimized cache settings for performance
//   - Foreign key constraint enforcement
//   - Connection pooling configuration
//   - Health checks via Ping
//   - Graceful connection closure
//   - High-performance batch insert for large datasets (50k+ rows)
//   - Transaction wrapper with automatic rollback and panic recovery
//
// # Basic Usage
//
//	db, err := database.NewDB("path/to/database.db")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer database.Close(db)
//
//	// Use db for queries...
//
// For in-memory databases (useful for testing):
//
//	db, err := database.NewDB(":memory:")
//
// # Batch Insert
//
// The package provides optimized batch insert functionality for importing large datasets.
// This is particularly useful for EVE SDE data imports (500k+ invTypes, etc.).
//
//	// Prepare data
//	columns := []string{"typeID", "typeName", "groupID"}
//	rows := [][]interface{}{
//		{34, "Tritanium", 18},
//		{35, "Pyerite", 18},
//		// ... more rows
//	}
//
//	// Perform batch insert
//	ctx := context.Background()
//	err = database.BatchInsert(ctx, db, "invTypes", columns, rows, 1000)
//
// For progress reporting during large imports:
//
//	progressCallback := func(current, total int) {
//		fmt.Printf("Imported %d/%d rows\n", current, total)
//	}
//	err = database.BatchInsertWithProgress(ctx, db, table, columns, rows, 1000, progressCallback)
//
// Performance characteristics:
//   - 10k rows: ~15ms
//   - 100k rows: ~130ms
//   - Automatic transaction management with rollback on error
//   - Configurable batch size (recommended: 1000 rows per statement)
//
// # Transaction Wrapper
//
// The package provides a safe transaction wrapper that handles commit, rollback,
// and panic recovery automatically. This ensures transactional safety for all database operations.
//
//	err = database.WithTransaction(ctx, db, func(tx *sqlx.Tx) error {
//		_, err := tx.Exec("INSERT INTO users (id, name) VALUES (?, ?)", 1, "Alice")
//		if err != nil {
//			return err // Transaction will be rolled back
//		}
//		_, err = tx.Exec("INSERT INTO roles (user_id, role) VALUES (?, ?)", 1, "admin")
//		return err // Transaction will be committed if no error
//	})
//
// Transaction options can be specified using functional options:
//
//	// Read-only transaction with serializable isolation
//	err = database.WithTransaction(ctx, db, func(tx *sqlx.Tx) error {
//		var count int
//		return tx.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
//	}, database.WithReadOnly(), database.WithIsolationLevel(sql.LevelSerializable))
//
// The transaction wrapper automatically:
//   - Commits on successful completion
//   - Rolls back on error return
//   - Rolls back and re-raises panic
//   - Respects context cancellation
//
// # Testing Utilities
//
// The package provides testing utilities for creating in-memory databases with
// all migrations automatically applied. This simplifies test setup and ensures
// consistency across tests.
//
//	func TestMyFeature(t *testing.T) {
//		// Create in-memory database with all migrations applied
//		db := database.NewTestDB(t)
//		// Database is automatically cleaned up when test completes
//
//		// Insert test data
//		_, err := db.Exec("INSERT INTO invTypes (typeID, typeName, groupID) VALUES (1, 'Test', 1)")
//		if err != nil {
//			t.Fatal(err)
//		}
//
//		// Run your tests...
//	}
//
// For manual migration control:
//
//	db, _ := database.NewDB(":memory:")
//	defer database.Close(db)
//
//	// Apply all migrations from migrations/sqlite directory
//	if err := database.ApplyMigrations(db); err != nil {
//		log.Fatal(err)
//	}
//
// Testing utilities features:
//   - Automatic migration application (all .sql files in migrations/sqlite)
//   - Automatic cleanup via t.Cleanup()
//   - Independent database per test (no shared state)
//   - All PRAGMAs pre-configured (foreign keys, WAL mode, cache settings)
package database
