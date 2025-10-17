// Package database provides SQLite database connection management
// with optimized batch insert functionality.
package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

// ProgressCallback is an optional callback function that reports progress during batch inserts.
// It receives the current row number and total rows being processed.
type ProgressCallback func(current, total int)

// BatchInsert performs optimized batch insertion of rows into a specified table.
// It automatically splits large datasets into batches and wraps all operations in a transaction.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - db: SQLite database connection
//   - table: Name of the target table
//   - columns: Column names for the insert operation
//   - rows: Data rows to insert, each row contains values corresponding to columns
//   - batchSize: Number of rows per INSERT statement (recommended: 1000)
//
// Returns:
//   - error: Any error encountered during the batch insert operation
//
// The function ensures transactional safety: all inserts are rolled back if any error occurs.
// Performance: Optimized for large datasets (50k+ rows) using multi-row INSERT statements.
func BatchInsert(ctx context.Context, db *sqlx.DB, table string, columns []string, rows [][]interface{}, batchSize int) error {
	return BatchInsertWithProgress(ctx, db, table, columns, rows, batchSize, nil)
}

// BatchInsertWithProgress performs batch insertion with optional progress reporting.
// See BatchInsert for parameter documentation.
//
// Additional parameter:
//   - progressCallback: Optional callback function to report progress (can be nil)
func BatchInsertWithProgress(ctx context.Context, db *sqlx.DB, table string, columns []string, rows [][]interface{}, batchSize int, progressCallback ProgressCallback) error {
	// Validate inputs
	if table == "" {
		return fmt.Errorf("table name cannot be empty")
	}
	if len(columns) == 0 {
		return fmt.Errorf("columns cannot be empty")
	}
	if len(rows) == 0 {
		return nil // Nothing to insert
	}
	if batchSize <= 0 {
		return fmt.Errorf("batchSize must be greater than 0")
	}

	// Validate all rows have correct column count
	for i, row := range rows {
		if len(row) != len(columns) {
			return fmt.Errorf("row %d has %d values, expected %d columns", i, len(row), len(columns))
		}
	}

	// Begin transaction
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // Rollback if not committed (ignore error as commit may have succeeded)
	}()

	totalRows := len(rows)
	processedRows := 0

	// Process rows in batches
	for i := 0; i < totalRows; i += batchSize {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("batch insert cancelled: %w", ctx.Err())
		default:
		}

		// Calculate batch end index
		end := i + batchSize
		if end > totalRows {
			end = totalRows
		}

		batch := rows[i:end]
		currentBatchSize := len(batch)

		// Build SQL for this batch
		sql := buildBatchInsertSQL(table, columns, currentBatchSize)

		// Flatten batch data for SQL execution
		args := make([]interface{}, 0, currentBatchSize*len(columns))
		for _, row := range batch {
			args = append(args, row...)
		}

		// Execute batch insert
		_, err := tx.ExecContext(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to insert batch at row %d: %w", i, err)
		}

		processedRows += currentBatchSize

		// Report progress if callback is provided
		if progressCallback != nil {
			progressCallback(processedRows, totalRows)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// buildBatchInsertSQL generates an optimized multi-row INSERT statement.
//
// Example output:
//
//	INSERT INTO invTypes (typeID, typeName, groupID) VALUES (?, ?, ?), (?, ?, ?), (?, ?, ?)
//
// Parameters:
//   - table: Name of the target table
//   - columns: Column names for the insert
//   - batchSize: Number of rows in this batch
//
// Returns:
//   - string: The generated SQL INSERT statement
func buildBatchInsertSQL(table string, columns []string, batchSize int) string {
	var sb strings.Builder

	// Build column list
	sb.WriteString("INSERT INTO ")
	sb.WriteString(table)
	sb.WriteString(" (")
	sb.WriteString(strings.Join(columns, ", "))
	sb.WriteString(") VALUES ")

	// Build value placeholders
	numColumns := len(columns)
	valuePlaceholder := "(" + strings.Repeat("?, ", numColumns-1) + "?)"

	// Add placeholders for each row in batch
	for i := 0; i < batchSize; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(valuePlaceholder)
	}

	return sb.String()
}
