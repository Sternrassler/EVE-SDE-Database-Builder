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
package database
