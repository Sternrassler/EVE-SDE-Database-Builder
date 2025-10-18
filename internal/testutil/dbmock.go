// Package testutil provides shared testing utilities for EVE SDE Database Builder tests.
package testutil

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// DBInterface defines the minimal database interface needed for testing.
// This allows mocking database operations without needing a real database connection.
type DBInterface interface {
	// Exec executes a query without returning any rows
	Exec(query string, args ...interface{}) (sql.Result, error)

	// Query executes a query that returns rows
	Query(query string, args ...interface{}) (*sql.Rows, error)

	// QueryRow executes a query that is expected to return at most one row
	QueryRow(query string, args ...interface{}) *sql.Row

	// Prepare creates a prepared statement for later queries or executions
	Prepare(query string) (*sql.Stmt, error)

	// Begin starts a transaction
	Begin() (*sql.Tx, error)

	// Close closes the database connection
	Close() error

	// Ping verifies a connection to the database is still alive
	Ping() error
}

// MockDB is a mock implementation of DBInterface for testing.
// It allows configuring expected behavior and recording actual calls.
type MockDB struct {
	// ExecFunc is called when Exec is invoked
	ExecFunc func(query string, args ...interface{}) (sql.Result, error)

	// QueryFunc is called when Query is invoked
	QueryFunc func(query string, args ...interface{}) (*sql.Rows, error)

	// QueryRowFunc is called when QueryRow is invoked
	QueryRowFunc func(query string, args ...interface{}) *sql.Row

	// PrepareFunc is called when Prepare is invoked
	PrepareFunc func(query string) (*sql.Stmt, error)

	// BeginFunc is called when Begin is invoked
	BeginFunc func() (*sql.Tx, error)

	// CloseFunc is called when Close is invoked
	CloseFunc func() error

	// PingFunc is called when Ping is invoked
	PingFunc func() error

	// ExecCalls records all calls to Exec
	ExecCalls []ExecCall

	// QueryCalls records all calls to Query
	QueryCalls []QueryCall

	// QueryRowCalls records all calls to QueryRow
	QueryRowCalls []QueryCall

	// PrepareCalls records all calls to Prepare
	PrepareCalls []string

	// Closed tracks whether Close was called
	Closed bool
}

// ExecCall records a call to Exec
type ExecCall struct {
	Query string
	Args  []interface{}
}

// QueryCall records a call to Query or QueryRow
type QueryCall struct {
	Query string
	Args  []interface{}
}

// NewMockDB creates a new MockDB with default no-op implementations.
func NewMockDB() *MockDB {
	return &MockDB{
		ExecFunc: func(query string, args ...interface{}) (sql.Result, error) {
			return MockResult{}, nil
		},
		QueryFunc: func(query string, args ...interface{}) (*sql.Rows, error) {
			return nil, errors.New("not implemented")
		},
		QueryRowFunc: func(query string, args ...interface{}) *sql.Row {
			return nil
		},
		PrepareFunc: func(query string) (*sql.Stmt, error) {
			return nil, errors.New("not implemented")
		},
		BeginFunc: func() (*sql.Tx, error) {
			return nil, errors.New("not implemented")
		},
		CloseFunc: func() error {
			return nil
		},
		PingFunc: func() error {
			return nil
		},
		ExecCalls:     make([]ExecCall, 0),
		QueryCalls:    make([]QueryCall, 0),
		QueryRowCalls: make([]QueryCall, 0),
		PrepareCalls:  make([]string, 0),
		Closed:        false,
	}
}

// Exec implements DBInterface
func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	m.ExecCalls = append(m.ExecCalls, ExecCall{Query: query, Args: args})
	return m.ExecFunc(query, args...)
}

// Query implements DBInterface
func (m *MockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	m.QueryCalls = append(m.QueryCalls, QueryCall{Query: query, Args: args})
	return m.QueryFunc(query, args...)
}

// QueryRow implements DBInterface
func (m *MockDB) QueryRow(query string, args ...interface{}) *sql.Row {
	m.QueryRowCalls = append(m.QueryRowCalls, QueryCall{Query: query, Args: args})
	return m.QueryRowFunc(query, args...)
}

// Prepare implements DBInterface
func (m *MockDB) Prepare(query string) (*sql.Stmt, error) {
	m.PrepareCalls = append(m.PrepareCalls, query)
	return m.PrepareFunc(query)
}

// Begin implements DBInterface
func (m *MockDB) Begin() (*sql.Tx, error) {
	return m.BeginFunc()
}

// Close implements DBInterface
func (m *MockDB) Close() error {
	m.Closed = true
	return m.CloseFunc()
}

// Ping implements DBInterface
func (m *MockDB) Ping() error {
	return m.PingFunc()
}

// Reset clears all recorded calls and resets the Closed flag
func (m *MockDB) Reset() {
	m.ExecCalls = make([]ExecCall, 0)
	m.QueryCalls = make([]QueryCall, 0)
	m.QueryRowCalls = make([]QueryCall, 0)
	m.PrepareCalls = make([]string, 0)
	m.Closed = false
}

// MockResult is a mock implementation of sql.Result for testing.
type MockResult struct {
	LastID       int64
	AffectedRows int64
	Err          error
}

// LastInsertId implements sql.Result
func (m MockResult) LastInsertId() (int64, error) {
	if m.Err != nil {
		return 0, m.Err
	}
	return m.LastID, nil
}

// RowsAffected implements sql.Result
func (m MockResult) RowsAffected() (int64, error) {
	if m.Err != nil {
		return 0, m.Err
	}
	return m.AffectedRows, nil
}

// NewMockResult creates a MockResult with the given values.
func NewMockResult(lastID, rowsAffected int64) MockResult {
	return MockResult{
		LastID:       lastID,
		AffectedRows: rowsAffected,
		Err:          nil,
	}
}

// NewMockResultWithError creates a MockResult that returns an error.
func NewMockResultWithError(err error) MockResult {
	return MockResult{
		LastID:       0,
		AffectedRows: 0,
		Err:          err,
	}
}

// SQLXAdapter wraps *sqlx.DB to implement DBInterface.
// This allows using real database connections through the same interface.
type SQLXAdapter struct {
	DB *sqlx.DB
}

// NewSQLXAdapter creates a new adapter for *sqlx.DB
func NewSQLXAdapter(db *sqlx.DB) *SQLXAdapter {
	return &SQLXAdapter{DB: db}
}

// Exec implements DBInterface
func (a *SQLXAdapter) Exec(query string, args ...interface{}) (sql.Result, error) {
	return a.DB.Exec(query, args...)
}

// Query implements DBInterface
func (a *SQLXAdapter) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return a.DB.Query(query, args...)
}

// QueryRow implements DBInterface
func (a *SQLXAdapter) QueryRow(query string, args ...interface{}) *sql.Row {
	return a.DB.QueryRow(query, args...)
}

// Prepare implements DBInterface
func (a *SQLXAdapter) Prepare(query string) (*sql.Stmt, error) {
	return a.DB.Prepare(query)
}

// Begin implements DBInterface
func (a *SQLXAdapter) Begin() (*sql.Tx, error) {
	return a.DB.Begin()
}

// Close implements DBInterface
func (a *SQLXAdapter) Close() error {
	return a.DB.Close()
}

// Ping implements DBInterface
func (a *SQLXAdapter) Ping() error {
	return a.DB.Ping()
}

// MockDriver is a minimal mock implementation of database/sql/driver for testing.
// This is used internally for creating test connections.
type MockDriver struct{}

// Open implements driver.Driver
func (d MockDriver) Open(name string) (driver.Conn, error) {
	return &MockConn{}, nil
}

// MockConn is a mock database connection
type MockConn struct{}

// Prepare implements driver.Conn
func (c *MockConn) Prepare(query string) (driver.Stmt, error) {
	return &MockStmt{}, nil
}

// Close implements driver.Conn
func (c *MockConn) Close() error {
	return nil
}

// Begin implements driver.Conn
func (c *MockConn) Begin() (driver.Tx, error) {
	return &MockTx{}, nil
}

// MockStmt is a mock prepared statement
type MockStmt struct{}

// Close implements driver.Stmt
func (s *MockStmt) Close() error {
	return nil
}

// NumInput implements driver.Stmt
func (s *MockStmt) NumInput() int {
	return -1
}

// Exec implements driver.Stmt
func (s *MockStmt) Exec(args []driver.Value) (driver.Result, error) {
	return driver.ResultNoRows, nil
}

// Query implements driver.Stmt
func (s *MockStmt) Query(args []driver.Value) (driver.Rows, error) {
	return &MockRows{}, nil
}

// MockRows is a mock result set
type MockRows struct {
	closed bool
}

// Columns implements driver.Rows
func (r *MockRows) Columns() []string {
	return []string{}
}

// Close implements driver.Rows
func (r *MockRows) Close() error {
	r.closed = true
	return nil
}

// Next implements driver.Rows
func (r *MockRows) Next(dest []driver.Value) error {
	return fmt.Errorf("no rows")
}

// MockTx is a mock transaction
type MockTx struct{}

// Commit implements driver.Tx
func (t *MockTx) Commit() error {
	return nil
}

// Rollback implements driver.Tx
func (t *MockTx) Rollback() error {
	return nil
}
