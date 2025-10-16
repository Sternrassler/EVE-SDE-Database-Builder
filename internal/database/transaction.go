// Package database provides SQLite database connection management
// with transaction wrapper functionality.
package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// TxOption is a functional option for configuring transaction behavior.
type TxOption func(*sql.TxOptions)

// WithIsolationLevel configures the isolation level for the transaction.
// SQLite supports: ReadUncommitted, ReadCommitted, RepeatableRead, Serializable
func WithIsolationLevel(level sql.IsolationLevel) TxOption {
	return func(opts *sql.TxOptions) {
		opts.Isolation = level
	}
}

// WithReadOnly configures the transaction as read-only.
func WithReadOnly() TxOption {
	return func(opts *sql.TxOptions) {
		opts.ReadOnly = true
	}
}

// WithTransaction executes a function within a database transaction.
// It automatically handles commit on success, rollback on error or panic,
// and supports context cancellation.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - db: Database connection
//   - fn: Function to execute within the transaction
//   - opts: Optional transaction configuration options
//
// Returns:
//   - error: Any error from the transaction function or transaction management
//
// The transaction is automatically rolled back if:
//   - The function returns an error
//   - A panic occurs (panic is re-raised after rollback)
//   - The context is cancelled
func WithTransaction(ctx context.Context, db *sqlx.DB, fn func(*sqlx.Tx) error, opts ...TxOption) error {
	// Build transaction options
	txOpts := &sql.TxOptions{}
	for _, opt := range opts {
		opt(txOpts)
	}

	// Begin transaction
	tx, err := db.BeginTxx(ctx, txOpts)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure rollback on panic or error
	defer func() {
		if p := recover(); p != nil {
			// Rollback on panic
			_ = tx.Rollback()
			panic(p) // Re-raise panic after rollback
		} else if err != nil {
			// Rollback on error
			_ = tx.Rollback()
		} else {
			// Commit on success
			err = tx.Commit()
			if err != nil {
				err = fmt.Errorf("failed to commit transaction: %w", err)
			}
		}
	}()

	// Execute transaction function
	err = fn(tx)
	return err
}
