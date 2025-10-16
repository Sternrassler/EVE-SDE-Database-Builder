// Package database provides SQLite database connection management
// with generic query helper functions.
package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"
	"github.com/jmoiron/sqlx"
)

// QueryRow executes a query that returns a single row and scans it into the provided type.
// It uses sqlx.Get internally for struct scanning with proper field mapping.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - db: Database connection or transaction that implements QueryerContext
//   - query: SQL query string (should return exactly one row)
//   - args: Query arguments for parameter binding
//
// Returns:
//   - T: The scanned result of type T
//   - error: Any error encountered, including ErrNotFound if no rows are returned
//
// Example:
//
//	type User struct {
//	    ID   int    `db:"id"`
//	    Name string `db:"name"`
//	}
//	user, err := QueryRow[User](ctx, db, "SELECT id, name FROM users WHERE id = ?", 1)
func QueryRow[T any](ctx context.Context, db sqlx.QueryerContext, query string, args ...interface{}) (T, error) {
	var result T

	// Use sqlx.GetContext for struct scanning
	// We need a *sqlx.DB or *sqlx.Tx to use GetContext
	switch v := db.(type) {
	case *sqlx.DB:
		err := v.GetContext(ctx, &result, query, args...)
		if err != nil {
			if err == sql.ErrNoRows {
				return result, errors.NewValidation("no rows found", err)
			}
			return result, fmt.Errorf("failed to query row: %w", err)
		}
	case *sqlx.Tx:
		err := v.GetContext(ctx, &result, query, args...)
		if err != nil {
			if err == sql.ErrNoRows {
				return result, errors.NewValidation("no rows found", err)
			}
			return result, fmt.Errorf("failed to query row: %w", err)
		}
	default:
		// Fallback: manually query and scan
		rows, err := db.QueryContext(ctx, query, args...)
		if err != nil {
			return result, fmt.Errorf("failed to execute query: %w", err)
		}
		defer rows.Close()

		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return result, fmt.Errorf("query error: %w", err)
			}
			return result, errors.NewValidation("no rows found", sql.ErrNoRows)
		}

		if err := rows.Scan(&result); err != nil {
			return result, fmt.Errorf("failed to scan row: %w", err)
		}

		if err := rows.Err(); err != nil {
			return result, fmt.Errorf("query error: %w", err)
		}
	}

	return result, nil
}

// QueryAll executes a query that returns multiple rows and scans them into a slice.
// It uses sqlx.Select internally for efficient struct scanning.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - db: Database connection or transaction that implements QueryerContext
//   - query: SQL query string
//   - args: Query arguments for parameter binding
//
// Returns:
//   - []T: A slice of results of type T (empty slice if no rows found)
//   - error: Any error encountered during query execution
//
// Example:
//
//	type User struct {
//	    ID   int    `db:"id"`
//	    Name string `db:"name"`
//	}
//	users, err := QueryAll[User](ctx, db, "SELECT id, name FROM users WHERE active = ?", true)
func QueryAll[T any](ctx context.Context, db sqlx.QueryerContext, query string, args ...interface{}) ([]T, error) {
	var results []T

	// Use sqlx.SelectContext for struct scanning
	switch v := db.(type) {
	case *sqlx.DB:
		err := v.SelectContext(ctx, &results, query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to query all rows: %w", err)
		}
	case *sqlx.Tx:
		err := v.SelectContext(ctx, &results, query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to query all rows: %w", err)
		}
	default:
		// Fallback: manually query and scan
		rows, err := db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to execute query: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var result T
			if err := rows.Scan(&result); err != nil {
				return nil, fmt.Errorf("failed to scan row: %w", err)
			}
			results = append(results, result)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("query error: %w", err)
		}
	}

	// Return empty slice instead of nil if no results
	if results == nil {
		results = []T{}
	}

	return results, nil
}

// Exists checks if a query returns at least one row.
// It's optimized for existence checks and doesn't retrieve actual data.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - db: Database connection or transaction that implements QueryerContext
//   - query: SQL query string (typically a SELECT with WHERE conditions)
//   - args: Query arguments for parameter binding
//
// Returns:
//   - bool: true if at least one row exists, false otherwise
//   - error: Any error encountered during query execution
//
// Example:
//
//	exists, err := Exists(ctx, db, "SELECT 1 FROM users WHERE email = ?", "user@example.com")
//	if err != nil {
//	    return err
//	}
//	if exists {
//	    // User with this email already exists
//	}
func Exists(ctx context.Context, db sqlx.QueryerContext, query string, args ...interface{}) (bool, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return false, fmt.Errorf("failed to execute existence check: %w", err)
	}
	defer rows.Close()

	// Check if at least one row exists
	exists := rows.Next()

	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("existence check error: %w", err)
	}

	return exists, nil
}
