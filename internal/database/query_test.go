package database

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	apperrors "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"
)

// TestQueryRow_SingleResult tests QueryRow with a successful single row query
func TestQueryRow_SingleResult(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table
	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert test data
	_, err = db.Exec("INSERT INTO users (id, name, email) VALUES (?, ?, ?)", 1, "Alice", "alice@example.com")
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	// Test struct type
	type User struct {
		ID    int    `db:"id"`
		Name  string `db:"name"`
		Email string `db:"email"`
	}

	ctx := context.Background()
	user, err := QueryRow[User](ctx, db, "SELECT id, name, email FROM users WHERE id = ?", 1)
	if err != nil {
		t.Fatalf("QueryRow failed: %v", err)
	}

	if user.ID != 1 {
		t.Errorf("Expected ID 1, got %d", user.ID)
	}
	if user.Name != "Alice" {
		t.Errorf("Expected name 'Alice', got '%s'", user.Name)
	}
	if user.Email != "alice@example.com" {
		t.Errorf("Expected email 'alice@example.com', got '%s'", user.Email)
	}
}

// TestQueryRow_PrimitiveType tests QueryRow with primitive types
func TestQueryRow_PrimitiveType(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table
	_, err = db.Exec(`CREATE TABLE counters (value INTEGER)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert test data
	_, err = db.Exec("INSERT INTO counters (value) VALUES (?)", 42)
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	ctx := context.Background()
	value, err := QueryRow[int](ctx, db, "SELECT value FROM counters")
	if err != nil {
		t.Fatalf("QueryRow failed: %v", err)
	}

	if value != 42 {
		t.Errorf("Expected value 42, got %d", value)
	}
}

// TestQueryRow_NoRows tests QueryRow with no matching rows (should return custom error)
func TestQueryRow_NoRows(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table
	_, err = db.Exec(`CREATE TABLE users (id INTEGER, name TEXT)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	type User struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	ctx := context.Background()
	_, err = QueryRow[User](ctx, db, "SELECT id, name FROM users WHERE id = ?", 999)
	if err == nil {
		t.Fatal("Expected error for no rows, got nil")
	}

	// Check that it's wrapped as a validation error
	if !apperrors.IsValidation(err) {
		t.Errorf("Expected validation error, got: %v", err)
	}

	// Verify that the original sql.ErrNoRows is in the error chain
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("Expected sql.ErrNoRows in error chain, got: %v", err)
	}
}

// TestQueryRow_WithTransaction tests QueryRow within a transaction
func TestQueryRow_WithTransaction(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table
	_, err = db.Exec(`CREATE TABLE products (id INTEGER, name TEXT, price REAL)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	type Product struct {
		ID    int     `db:"id"`
		Name  string  `db:"name"`
		Price float64 `db:"price"`
	}

	ctx := context.Background()
	err = WithTransaction(ctx, db, func(tx *sqlx.Tx) error {
		// Insert within transaction
		_, err := tx.Exec("INSERT INTO products (id, name, price) VALUES (?, ?, ?)", 1, "Widget", 9.99)
		if err != nil {
			return err
		}

		// Query within transaction using QueryRow
		product, err := QueryRow[Product](ctx, tx, "SELECT id, name, price FROM products WHERE id = ?", 1)
		if err != nil {
			return err
		}

		if product.Name != "Widget" {
			t.Errorf("Expected name 'Widget', got '%s'", product.Name)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}
}

// TestQueryRow_ContextCancellation tests QueryRow with cancelled context
func TestQueryRow_ContextCancellation(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table
	_, err = db.Exec(`CREATE TABLE test (id INTEGER)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	type Test struct {
		ID int `db:"id"`
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = QueryRow[Test](ctx, db, "SELECT id FROM test")
	if err == nil {
		t.Fatal("Expected error for cancelled context, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled error, got: %v", err)
	}
}

// TestQueryAll_MultipleRows tests QueryAll with multiple rows
func TestQueryAll_MultipleRows(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table
	_, err = db.Exec(`
		CREATE TABLE items (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			quantity INTEGER
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert test data
	testData := []struct {
		id       int
		name     string
		quantity int
	}{
		{1, "Item A", 10},
		{2, "Item B", 20},
		{3, "Item C", 30},
	}

	for _, td := range testData {
		_, err = db.Exec("INSERT INTO items (id, name, quantity) VALUES (?, ?, ?)", td.id, td.name, td.quantity)
		if err != nil {
			t.Fatalf("Failed to insert data: %v", err)
		}
	}

	type Item struct {
		ID       int    `db:"id"`
		Name     string `db:"name"`
		Quantity int    `db:"quantity"`
	}

	ctx := context.Background()
	items, err := QueryAll[Item](ctx, db, "SELECT id, name, quantity FROM items ORDER BY id")
	if err != nil {
		t.Fatalf("QueryAll failed: %v", err)
	}

	if len(items) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(items))
	}

	for i, item := range items {
		if item.ID != testData[i].id {
			t.Errorf("Item %d: expected ID %d, got %d", i, testData[i].id, item.ID)
		}
		if item.Name != testData[i].name {
			t.Errorf("Item %d: expected name '%s', got '%s'", i, testData[i].name, item.Name)
		}
		if item.Quantity != testData[i].quantity {
			t.Errorf("Item %d: expected quantity %d, got %d", i, testData[i].quantity, item.Quantity)
		}
	}
}

// TestQueryAll_EmptyResult tests QueryAll with no matching rows (should return empty slice)
func TestQueryAll_EmptyResult(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table
	_, err = db.Exec(`CREATE TABLE items (id INTEGER, name TEXT)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	type Item struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	ctx := context.Background()
	items, err := QueryAll[Item](ctx, db, "SELECT id, name FROM items WHERE id > ?", 100)
	if err != nil {
		t.Fatalf("QueryAll failed: %v", err)
	}

	if items == nil {
		t.Fatal("Expected empty slice, got nil")
	}

	if len(items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(items))
	}
}

// TestQueryAll_WithFilter tests QueryAll with WHERE clause
func TestQueryAll_WithFilter(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table
	_, err = db.Exec(`CREATE TABLE products (id INTEGER, category TEXT, price REAL)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert test data
	_, err = db.Exec("INSERT INTO products VALUES (1, 'electronics', 99.99), (2, 'books', 19.99), (3, 'electronics', 149.99)")
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	type Product struct {
		ID       int     `db:"id"`
		Category string  `db:"category"`
		Price    float64 `db:"price"`
	}

	ctx := context.Background()
	products, err := QueryAll[Product](ctx, db, "SELECT id, category, price FROM products WHERE category = ?", "electronics")
	if err != nil {
		t.Fatalf("QueryAll failed: %v", err)
	}

	if len(products) != 2 {
		t.Fatalf("Expected 2 products, got %d", len(products))
	}

	for _, p := range products {
		if p.Category != "electronics" {
			t.Errorf("Expected category 'electronics', got '%s'", p.Category)
		}
	}
}

// TestQueryAll_WithTransaction tests QueryAll within a transaction
func TestQueryAll_WithTransaction(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table
	_, err = db.Exec(`CREATE TABLE orders (id INTEGER, status TEXT)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	type Order struct {
		ID     int    `db:"id"`
		Status string `db:"status"`
	}

	ctx := context.Background()
	err = WithTransaction(ctx, db, func(tx *sqlx.Tx) error {
		// Insert within transaction
		_, err := tx.Exec("INSERT INTO orders VALUES (1, 'pending'), (2, 'pending'), (3, 'completed')")
		if err != nil {
			return err
		}

		// Query within transaction
		orders, err := QueryAll[Order](ctx, tx, "SELECT id, status FROM orders WHERE status = ?", "pending")
		if err != nil {
			return err
		}

		if len(orders) != 2 {
			t.Errorf("Expected 2 pending orders, got %d", len(orders))
		}

		return nil
	})
	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}
}

// TestExists_True tests Exists when rows exist
func TestExists_True(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table
	_, err = db.Exec(`CREATE TABLE users (id INTEGER, email TEXT)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert test data
	_, err = db.Exec("INSERT INTO users (id, email) VALUES (?, ?)", 1, "user@example.com")
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	ctx := context.Background()
	exists, err := Exists(ctx, db, "SELECT 1 FROM users WHERE email = ?", "user@example.com")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}

	if !exists {
		t.Error("Expected exists to be true")
	}
}

// TestExists_False tests Exists when no rows exist
func TestExists_False(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table
	_, err = db.Exec(`CREATE TABLE users (id INTEGER, email TEXT)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	ctx := context.Background()
	exists, err := Exists(ctx, db, "SELECT 1 FROM users WHERE email = ?", "nonexistent@example.com")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}

	if exists {
		t.Error("Expected exists to be false")
	}
}

// TestExists_WithTransaction tests Exists within a transaction
func TestExists_WithTransaction(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table
	_, err = db.Exec(`CREATE TABLE sessions (id INTEGER, token TEXT)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	ctx := context.Background()
	err = WithTransaction(ctx, db, func(tx *sqlx.Tx) error {
		// Insert within transaction
		_, err := tx.Exec("INSERT INTO sessions (id, token) VALUES (?, ?)", 1, "abc123")
		if err != nil {
			return err
		}

		// Check existence within transaction
		exists, err := Exists(ctx, tx, "SELECT 1 FROM sessions WHERE token = ?", "abc123")
		if err != nil {
			return err
		}

		if !exists {
			t.Error("Expected session to exist within transaction")
		}

		return nil
	})
	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}
}

// TestExists_EmptyTable tests Exists on empty table
func TestExists_EmptyTable(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table
	_, err = db.Exec(`CREATE TABLE empty_table (id INTEGER)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	ctx := context.Background()
	exists, err := Exists(ctx, db, "SELECT 1 FROM empty_table")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}

	if exists {
		t.Error("Expected exists to be false for empty table")
	}
}

// TestExists_ContextCancellation tests Exists with cancelled context
func TestExists_ContextCancellation(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table
	_, err = db.Exec(`CREATE TABLE test (id INTEGER)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = Exists(ctx, db, "SELECT 1 FROM test")
	if err == nil {
		t.Fatal("Expected error for cancelled context, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled error, got: %v", err)
	}
}

// TestQueryAll_ContextTimeout tests QueryAll with timeout
func TestQueryAll_ContextTimeout(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table
	_, err = db.Exec(`CREATE TABLE test (id INTEGER)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	type Test struct {
		ID int `db:"id"`
	}

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Sleep a bit to ensure timeout
	time.Sleep(10 * time.Millisecond)

	_, err = QueryAll[Test](ctx, db, "SELECT id FROM test")
	if err == nil {
		t.Fatal("Expected error for timeout context, got nil")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded error, got: %v", err)
	}
}

// mockQueryer implements sqlx.QueryerContext for testing fallback paths
type mockQueryer struct {
	db *sqlx.DB
}

func (m *mockQueryer) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return m.db.QueryContext(ctx, query, args...)
}

func (m *mockQueryer) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	return m.db.QueryxContext(ctx, query, args...)
}

func (m *mockQueryer) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	return m.db.QueryRowxContext(ctx, query, args...)
}

// TestQueryRow_FallbackPath tests QueryRow with a custom QueryerContext (not *sqlx.DB or *sqlx.Tx)
func TestQueryRow_FallbackPath(t *testing.T) {
	// Create a real database to get actual rows
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table and insert data
	_, err = db.Exec(`CREATE TABLE test (value INTEGER)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	_, err = db.Exec("INSERT INTO test (value) VALUES (?)", 42)
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	// Use mock queryer that wraps the db to trigger fallback path
	ctx := context.Background()
	mock := &mockQueryer{db: db}
	value, err := QueryRow[int](ctx, mock, "SELECT value FROM test")
	if err != nil {
		t.Fatalf("QueryRow with fallback path failed: %v", err)
	}

	if value != 42 {
		t.Errorf("Expected value 42, got %d", value)
	}
}

// TestQueryRow_FallbackPath_NoRows tests QueryRow fallback path with no rows
func TestQueryRow_FallbackPath_NoRows(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create empty table
	_, err = db.Exec(`CREATE TABLE test (value INTEGER)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Use mock queryer with empty table
	ctx := context.Background()
	mock := &mockQueryer{db: db}
	_, err = QueryRow[int](ctx, mock, "SELECT value FROM test")
	if err == nil {
		t.Fatal("Expected error for no rows, got nil")
	}

	if !apperrors.IsValidation(err) {
		t.Errorf("Expected validation error, got: %v", err)
	}
}

// mockQueryerError implements sqlx.QueryerContext for testing error paths
type mockQueryerError struct {
	err error
}

func (m *mockQueryerError) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, m.err
}

func (m *mockQueryerError) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	return nil, m.err
}

func (m *mockQueryerError) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	return nil
}

// TestQueryRow_FallbackPath_QueryError tests QueryRow fallback path with query error
func TestQueryRow_FallbackPath_QueryError(t *testing.T) {
	ctx := context.Background()
	mock := &mockQueryerError{err: sql.ErrConnDone}

	_, err := QueryRow[int](ctx, mock, "SELECT value FROM test")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !errors.Is(err, sql.ErrConnDone) {
		t.Errorf("Expected sql.ErrConnDone in error chain, got: %v", err)
	}
}

// TestQueryAll_FallbackPath tests QueryAll with a custom QueryerContext
func TestQueryAll_FallbackPath(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table and insert data
	_, err = db.Exec(`CREATE TABLE test (value INTEGER)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	_, err = db.Exec("INSERT INTO test (value) VALUES (1), (2), (3)")
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	// Use mock queryer to trigger fallback path
	ctx := context.Background()
	mock := &mockQueryer{db: db}
	values, err := QueryAll[int](ctx, mock, "SELECT value FROM test ORDER BY value")
	if err != nil {
		t.Fatalf("QueryAll with fallback path failed: %v", err)
	}

	if len(values) != 3 {
		t.Fatalf("Expected 3 values, got %d", len(values))
	}

	for i, v := range values {
		if v != i+1 {
			t.Errorf("Expected value %d, got %d", i+1, v)
		}
	}
}

// TestQueryAll_FallbackPath_EmptyResult tests QueryAll fallback path with no rows
func TestQueryAll_FallbackPath_EmptyResult(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create empty table
	_, err = db.Exec(`CREATE TABLE test (value INTEGER)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Use mock queryer with empty table
	ctx := context.Background()
	mock := &mockQueryer{db: db}
	values, err := QueryAll[int](ctx, mock, "SELECT value FROM test")
	if err != nil {
		t.Fatalf("QueryAll failed: %v", err)
	}

	if values == nil {
		t.Fatal("Expected empty slice, got nil")
	}

	if len(values) != 0 {
		t.Errorf("Expected 0 values, got %d", len(values))
	}
}

// TestQueryAll_FallbackPath_QueryError tests QueryAll fallback path with query error
func TestQueryAll_FallbackPath_QueryError(t *testing.T) {
	ctx := context.Background()
	mock := &mockQueryerError{err: sql.ErrConnDone}

	_, err := QueryAll[int](ctx, mock, "SELECT value FROM test")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !errors.Is(err, sql.ErrConnDone) {
		t.Errorf("Expected sql.ErrConnDone in error chain, got: %v", err)
	}
}

// TestQueryRow_FallbackPath_ScanError tests QueryRow fallback path with scan error
func TestQueryRow_FallbackPath_ScanError(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table with string column
	_, err = db.Exec(`CREATE TABLE test (value TEXT)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	_, err = db.Exec("INSERT INTO test (value) VALUES (?)", "not_an_int")
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	// Try to scan string into int - should fail with scan error in fallback path
	ctx := context.Background()
	mock := &mockQueryer{db: db}
	_, err = QueryRow[int](ctx, mock, "SELECT value FROM test")
	if err == nil {
		t.Fatal("Expected scan error, got nil")
	}

	// Should contain "scan" in error message
	if !strings.Contains(err.Error(), "failed to scan row") {
		t.Errorf("Expected scan error, got: %v", err)
	}
}

// TestQueryAll_FallbackPath_ScanError tests QueryAll fallback path with scan error
func TestQueryAll_FallbackPath_ScanError(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table with string column
	_, err = db.Exec(`CREATE TABLE test (value TEXT)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	_, err = db.Exec("INSERT INTO test (value) VALUES (?)", "not_an_int")
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	// Try to scan string into int - should fail with scan error in fallback path
	ctx := context.Background()
	mock := &mockQueryer{db: db}
	_, err = QueryAll[int](ctx, mock, "SELECT value FROM test")
	if err == nil {
		t.Fatal("Expected scan error, got nil")
	}

	// Should contain "scan" in error message
	if !strings.Contains(err.Error(), "failed to scan row") {
		t.Errorf("Expected scan error, got: %v", err)
	}
}
