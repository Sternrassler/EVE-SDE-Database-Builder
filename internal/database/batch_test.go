package database

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

// TestBatchInsert_BasicFunctionality tests basic batch insert with small dataset
func TestBatchInsert_BasicFunctionality(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table
	_, err = db.Exec(`
		CREATE TABLE test_data (
			id INTEGER,
			name TEXT,
			value INTEGER
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Prepare test data
	columns := []string{"id", "name", "value"}
	rows := [][]interface{}{
		{1, "first", 100},
		{2, "second", 200},
		{3, "third", 300},
	}

	// Execute batch insert
	ctx := context.Background()
	err = BatchInsert(ctx, db, "test_data", columns, rows, 1000)
	if err != nil {
		t.Fatalf("BatchInsert failed: %v", err)
	}

	// Verify data was inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_data").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count rows: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 rows, got %d", count)
	}

	// Verify data integrity
	var name string
	var value int
	err = db.QueryRow("SELECT name, value FROM test_data WHERE id = 2").Scan(&name, &value)
	if err != nil {
		t.Fatalf("Failed to query row: %v", err)
	}

	if name != "second" || value != 200 {
		t.Errorf("Expected (second, 200), got (%s, %d)", name, value)
	}
}

// TestBatchInsert_EmptyRows tests handling of empty row slice
func TestBatchInsert_EmptyRows(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table
	_, err = db.Exec(`CREATE TABLE test_data (id INTEGER, name TEXT)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Test with empty rows
	columns := []string{"id", "name"}
	rows := [][]interface{}{}

	ctx := context.Background()
	err = BatchInsert(ctx, db, "test_data", columns, rows, 1000)
	if err != nil {
		t.Errorf("BatchInsert with empty rows should not fail: %v", err)
	}

	// Verify no data was inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_data").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count rows: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 rows, got %d", count)
	}
}

// TestBatchInsert_BatchSplitting tests that large datasets are split into multiple batches
func TestBatchInsert_BatchSplitting(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table
	_, err = db.Exec(`CREATE TABLE test_data (id INTEGER, value INTEGER)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Prepare 1500 rows (should be split into 2 batches with batchSize=1000)
	columns := []string{"id", "value"}
	rows := make([][]interface{}, 1500)
	for i := 0; i < 1500; i++ {
		rows[i] = []interface{}{i + 1, i * 10}
	}

	// Track progress to verify batching
	var progressCalls atomic.Int32
	progressCallback := func(current, total int) {
		progressCalls.Add(1)
		t.Logf("Progress: %d/%d", current, total)
	}

	// Execute batch insert with batch size of 1000
	ctx := context.Background()
	err = BatchInsertWithProgress(ctx, db, "test_data", columns, rows, 1000, progressCallback)
	if err != nil {
		t.Fatalf("BatchInsert failed: %v", err)
	}

	// Verify all data was inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_data").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count rows: %v", err)
	}

	if count != 1500 {
		t.Errorf("Expected 1500 rows, got %d", count)
	}

	// Verify progress callback was called (should be 2 times: after first 1000, then after remaining 500)
	calls := progressCalls.Load()
	if calls != 2 {
		t.Errorf("Expected 2 progress callbacks, got %d", calls)
	}

	// Verify data integrity for a few random rows
	testCases := []struct {
		id    int
		value int
	}{
		{1, 0},
		{500, 4990},
		{1000, 9990},
		{1500, 14990},
	}

	for _, tc := range testCases {
		var value int
		err = db.QueryRow("SELECT value FROM test_data WHERE id = ?", tc.id).Scan(&value)
		if err != nil {
			t.Errorf("Failed to query row %d: %v", tc.id, err)
			continue
		}
		if value != tc.value {
			t.Errorf("Row %d: expected value %d, got %d", tc.id, tc.value, value)
		}
	}
}

// TestBatchInsert_TransactionRollback tests that errors trigger rollback
func TestBatchInsert_TransactionRollback(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table with NOT NULL constraint
	_, err = db.Exec(`
		CREATE TABLE test_data (
			id INTEGER NOT NULL,
			name TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Prepare test data with an invalid row (NULL value)
	columns := []string{"id", "name"}
	rows := [][]interface{}{
		{1, "valid"},
		{2, nil}, // This will violate NOT NULL constraint
		{3, "also_valid"},
	}

	// Execute batch insert (should fail)
	ctx := context.Background()
	err = BatchInsert(ctx, db, "test_data", columns, rows, 1000)
	if err == nil {
		t.Fatal("BatchInsert should have failed with constraint violation")
	}

	// Verify NO data was inserted (transaction rolled back)
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_data").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count rows: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 rows after rollback, got %d", count)
	}
}

// TestBatchInsert_LargeDataset tests performance with 10k rows
func TestBatchInsert_LargeDataset(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table
	_, err = db.Exec(`
		CREATE TABLE test_data (
			id INTEGER,
			name TEXT,
			value INTEGER,
			description TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Prepare 10k rows
	columns := []string{"id", "name", "value", "description"}
	rows := make([][]interface{}, 10000)
	for i := 0; i < 10000; i++ {
		rows[i] = []interface{}{
			i + 1,
			fmt.Sprintf("item_%d", i),
			i * 100,
			fmt.Sprintf("Description for item %d", i),
		}
	}

	// Execute batch insert with timing
	ctx := context.Background()
	start := time.Now()
	err = BatchInsert(ctx, db, "test_data", columns, rows, 1000)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("BatchInsert failed: %v", err)
	}

	// Verify all data was inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_data").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count rows: %v", err)
	}

	if count != 10000 {
		t.Errorf("Expected 10000 rows, got %d", count)
	}

	// Performance check: should complete in less than 1 second
	t.Logf("Inserted 10k rows in %v", duration)
	if duration > time.Second {
		t.Errorf("Performance issue: 10k rows took %v (expected < 1s)", duration)
	}
}

// TestBatchInsert_ValidationErrors tests input validation
func TestBatchInsert_ValidationErrors(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	ctx := context.Background()

	tests := []struct {
		name      string
		table     string
		columns   []string
		rows      [][]interface{}
		batchSize int
		wantErr   bool
	}{
		{
			name:      "empty table name",
			table:     "",
			columns:   []string{"id"},
			rows:      [][]interface{}{{1}},
			batchSize: 1000,
			wantErr:   true,
		},
		{
			name:      "empty columns",
			table:     "test",
			columns:   []string{},
			rows:      [][]interface{}{{1}},
			batchSize: 1000,
			wantErr:   true,
		},
		{
			name:      "invalid batch size",
			table:     "test",
			columns:   []string{"id"},
			rows:      [][]interface{}{{1}},
			batchSize: 0,
			wantErr:   true,
		},
		{
			name:      "negative batch size",
			table:     "test",
			columns:   []string{"id"},
			rows:      [][]interface{}{{1}},
			batchSize: -1,
			wantErr:   true,
		},
		{
			name:      "mismatched column count",
			table:     "test",
			columns:   []string{"id", "name"},
			rows:      [][]interface{}{{1}}, // Only 1 value, expected 2
			batchSize: 1000,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := BatchInsert(ctx, db, tt.table, tt.columns, tt.rows, tt.batchSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("BatchInsert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestBatchInsert_ContextCancellation tests that context cancellation is handled
func TestBatchInsert_ContextCancellation(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table
	_, err = db.Exec(`CREATE TABLE test_data (id INTEGER, value INTEGER)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Prepare large dataset
	columns := []string{"id", "value"}
	rows := make([][]interface{}, 5000)
	for i := 0; i < 5000; i++ {
		rows[i] = []interface{}{i + 1, i * 10}
	}

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	// Execute batch insert (should fail due to cancellation)
	err = BatchInsert(ctx, db, "test_data", columns, rows, 1000)
	if err == nil {
		t.Error("BatchInsert should have failed with cancelled context")
	}

	// Verify no data was committed
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_data").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count rows: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 rows after cancellation, got %d", count)
	}
}

// TestBuildBatchInsertSQL tests SQL generation
func TestBuildBatchInsertSQL(t *testing.T) {
	tests := []struct {
		name      string
		table     string
		columns   []string
		batchSize int
		expected  string
	}{
		{
			name:      "single row",
			table:     "users",
			columns:   []string{"id", "name"},
			batchSize: 1,
			expected:  "INSERT INTO users (id, name) VALUES (?, ?)",
		},
		{
			name:      "multiple rows",
			table:     "users",
			columns:   []string{"id", "name"},
			batchSize: 3,
			expected:  "INSERT INTO users (id, name) VALUES (?, ?), (?, ?), (?, ?)",
		},
		{
			name:      "many columns",
			table:     "invTypes",
			columns:   []string{"typeID", "groupID", "typeName", "description", "mass"},
			batchSize: 2,
			expected:  "INSERT INTO invTypes (typeID, groupID, typeName, description, mass) VALUES (?, ?, ?, ?, ?), (?, ?, ?, ?, ?)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildBatchInsertSQL(tt.table, tt.columns, tt.batchSize)
			if result != tt.expected {
				t.Errorf("buildBatchInsertSQL() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// BenchmarkBatchInsert_100k benchmarks insertion of 100k rows
func BenchmarkBatchInsert_100k(b *testing.B) {
	db, err := NewDB(":memory:")
	if err != nil {
		b.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		_ = Close(db)
	}()

	// Create test table
	_, err = db.Exec(`
		CREATE TABLE test_data (
			id INTEGER,
			name TEXT,
			value INTEGER,
			description TEXT
		)
	`)
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	// Prepare 100k rows
	columns := []string{"id", "name", "value", "description"}
	rows := make([][]interface{}, 100000)
	for i := 0; i < 100000; i++ {
		rows[i] = []interface{}{
			i + 1,
			fmt.Sprintf("item_%d", i),
			i * 100,
			fmt.Sprintf("Description for item %d", i),
		}
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clear table for each iteration
		if i > 0 {
			_, _ = db.Exec("DELETE FROM test_data")
		}

		err = BatchInsert(ctx, db, "test_data", columns, rows, 1000)
		if err != nil {
			b.Fatalf("BatchInsert failed: %v", err)
		}
	}
	b.StopTimer()

	// Verify data was inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_data").Scan(&count)
	if err != nil {
		b.Fatalf("Failed to count rows: %v", err)
	}

	if count != 100000 {
		b.Errorf("Expected 100000 rows, got %d", count)
	}
}

// BenchmarkBatchInsert_DifferentBatchSizes benchmarks different batch sizes
func BenchmarkBatchInsert_DifferentBatchSizes(b *testing.B) {
	batchSizes := []int{100, 500, 1000, 2000, 5000}

	for _, batchSize := range batchSizes {
		b.Run(fmt.Sprintf("BatchSize_%d", batchSize), func(b *testing.B) {
			db, err := NewDB(":memory:")
			if err != nil {
				b.Fatalf("Failed to create database: %v", err)
			}
			defer func() {
				_ = Close(db)
			}()

			// Create test table
			_, err = db.Exec(`CREATE TABLE test_data (id INTEGER, value INTEGER)`)
			if err != nil {
				b.Fatalf("Failed to create table: %v", err)
			}

			// Prepare 10k rows
			columns := []string{"id", "value"}
			rows := make([][]interface{}, 10000)
			for i := 0; i < 10000; i++ {
				rows[i] = []interface{}{i + 1, i * 10}
			}

			ctx := context.Background()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if i > 0 {
					_, _ = db.Exec("DELETE FROM test_data")
				}

				err = BatchInsert(ctx, db, "test_data", columns, rows, batchSize)
				if err != nil {
					b.Fatalf("BatchInsert failed: %v", err)
				}
			}
		})
	}
}
