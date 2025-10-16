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
//
// Usage:
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
package database
