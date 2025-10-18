package database

import (
	"context"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperties_BatchInsert_RowCountPreservation tests that all rows are inserted
func TestProperties_BatchInsert_RowCountPreservation(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("batch insert preserves row count", prop.ForAll(
		func(rowCount int, batchSize int) bool {
			db, err := NewDB(":memory:")
			if err != nil {
				return false
			}
			defer func() { _ = Close(db) }()

			// Create test table
			_, err = db.Exec(`CREATE TABLE test_data (id INTEGER, value INTEGER)`)
			if err != nil {
				return false
			}

			// Generate test data
			columns := []string{"id", "value"}
			rows := make([][]interface{}, rowCount)
			for i := 0; i < rowCount; i++ {
				rows[i] = []interface{}{i + 1, i * 10}
			}

			// Execute batch insert
			ctx := context.Background()
			err = BatchInsert(ctx, db, "test_data", columns, rows, batchSize)
			if err != nil {
				return false
			}

			// Verify row count
			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM test_data").Scan(&count)
			if err != nil {
				return false
			}

			return count == rowCount
		},
		gen.IntRange(1, 1000),    // rowCount
		gen.IntRange(1, 500),     // batchSize
	))

	properties.TestingRun(t)
}

// TestProperties_BatchInsert_BatchSplittingCorrectness tests that batching works correctly regardless of batch size
func TestProperties_BatchInsert_BatchSplittingCorrectness(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("batch splitting produces same result", prop.ForAll(
		func(rowCount int, batchSize1 int, batchSize2 int) bool {
			// Create first database
			db1, err := NewDB(":memory:")
			if err != nil {
				return false
			}
			defer func() { _ = Close(db1) }()

			_, err = db1.Exec(`CREATE TABLE test_data (id INTEGER, value INTEGER)`)
			if err != nil {
				return false
			}

			// Create second database
			db2, err := NewDB(":memory:")
			if err != nil {
				return false
			}
			defer func() { _ = Close(db2) }()

			_, err = db2.Exec(`CREATE TABLE test_data (id INTEGER, value INTEGER)`)
			if err != nil {
				return false
			}

			// Generate test data
			columns := []string{"id", "value"}
			rows := make([][]interface{}, rowCount)
			for i := 0; i < rowCount; i++ {
				rows[i] = []interface{}{i + 1, i * 10}
			}

			// Insert with different batch sizes
			ctx := context.Background()
			err1 := BatchInsert(ctx, db1, "test_data", columns, rows, batchSize1)
			err2 := BatchInsert(ctx, db2, "test_data", columns, rows, batchSize2)

			if err1 != nil || err2 != nil {
				return false
			}

			// Verify both have same count
			var count1, count2 int
			err = db1.QueryRow("SELECT COUNT(*) FROM test_data").Scan(&count1)
			if err != nil {
				return false
			}
			err = db2.QueryRow("SELECT COUNT(*) FROM test_data").Scan(&count2)
			if err != nil {
				return false
			}

			return count1 == count2 && count1 == rowCount
		},
		gen.IntRange(1, 500),     // rowCount
		gen.IntRange(1, 100),     // batchSize1
		gen.IntRange(1, 100),     // batchSize2
	))

	properties.TestingRun(t)
}

// TestProperties_BatchInsert_DataIntegrity tests that data is correctly inserted
func TestProperties_BatchInsert_DataIntegrity(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("batch insert preserves data integrity", prop.ForAll(
		func(values []int, batchSize int) bool {
			if len(values) == 0 {
				return true // Skip empty case
			}

			db, err := NewDB(":memory:")
			if err != nil {
				return false
			}
			defer func() { _ = Close(db) }()

			// Create test table
			_, err = db.Exec(`CREATE TABLE test_data (id INTEGER, value INTEGER)`)
			if err != nil {
				return false
			}

			// Generate test data
			columns := []string{"id", "value"}
			rows := make([][]interface{}, len(values))
			for i, val := range values {
				rows[i] = []interface{}{i + 1, val}
			}

			// Execute batch insert
			ctx := context.Background()
			err = BatchInsert(ctx, db, "test_data", columns, rows, batchSize)
			if err != nil {
				return false
			}

			// Verify each value
			for i, expectedVal := range values {
				var actualVal int
				err = db.QueryRow("SELECT value FROM test_data WHERE id = ?", i+1).Scan(&actualVal)
				if err != nil {
					return false
				}
				if actualVal != expectedVal {
					return false
				}
			}

			return true
		},
		gen.SliceOf(gen.IntRange(-1000, 1000)).SuchThat(func(v interface{}) bool {
			slice := v.([]int)
			return len(slice) >= 1 && len(slice) <= 200
		}),
		gen.IntRange(1, 50),
	))

	properties.TestingRun(t)
}

// TestProperties_BatchInsert_EmptyRowsNoop tests that empty rows is a no-op
func TestProperties_BatchInsert_EmptyRowsNoop(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("empty rows is no-op", prop.ForAll(
		func(batchSize int) bool {
			db, err := NewDB(":memory:")
			if err != nil {
				return false
			}
			defer func() { _ = Close(db) }()

			// Create test table
			_, err = db.Exec(`CREATE TABLE test_data (id INTEGER, value INTEGER)`)
			if err != nil {
				return false
			}

			// Execute batch insert with empty rows
			columns := []string{"id", "value"}
			rows := [][]interface{}{}
			ctx := context.Background()
			err = BatchInsert(ctx, db, "test_data", columns, rows, batchSize)

			if err != nil {
				return false
			}

			// Verify no rows inserted
			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM test_data").Scan(&count)
			if err != nil {
				return false
			}

			return count == 0
		},
		gen.IntRange(1, 1000),
	))

	properties.TestingRun(t)
}

// TestProperties_BatchInsert_InvalidBatchSizeFails tests that invalid batch sizes fail
func TestProperties_BatchInsert_InvalidBatchSizeFails(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("invalid batch size fails", prop.ForAll(
		func(batchSize int) bool {
			db, err := NewDB(":memory:")
			if err != nil {
				return false
			}
			defer func() { _ = Close(db) }()

			// Create test table
			_, err = db.Exec(`CREATE TABLE test_data (id INTEGER)`)
			if err != nil {
				return false
			}

			// Try batch insert with invalid batch size
			columns := []string{"id"}
			rows := [][]interface{}{{1}}
			ctx := context.Background()
			err = BatchInsert(ctx, db, "test_data", columns, rows, batchSize)

			// Should fail for batch size <= 0
			return err != nil
		},
		gen.IntRange(-100, 0),
	))

	properties.TestingRun(t)
}

// TestProperties_BatchInsert_MismatchedColumnCountFails tests that mismatched columns fail
func TestProperties_BatchInsert_MismatchedColumnCountFails(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("mismatched column count fails", prop.ForAll(
		func(columnCount int, valueCount int) bool {
			if columnCount == valueCount {
				return true // Skip when they match
			}

			db, err := NewDB(":memory:")
			if err != nil {
				return false
			}
			defer func() { _ = Close(db) }()

			// Create columns
			columns := make([]string, columnCount)
			for i := 0; i < columnCount; i++ {
				columns[i] = "col" + string(rune('0'+i))
			}

			// Create row with different value count
			values := make([]interface{}, valueCount)
			for i := 0; i < valueCount; i++ {
				values[i] = i
			}
			rows := [][]interface{}{values}

			ctx := context.Background()
			err = BatchInsert(ctx, db, "test_data", columns, rows, 1000)

			// Should fail due to mismatch
			return err != nil
		},
		gen.IntRange(1, 10),  // columnCount
		gen.IntRange(1, 10),  // valueCount
	))

	properties.TestingRun(t)
}

// TestProperties_BatchInsert_TransactionalRollback tests that errors cause rollback
func TestProperties_BatchInsert_TransactionalRollback(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("error causes transaction rollback", prop.ForAll(
		func(validRowCount int) bool {
			db, err := NewDB(":memory:")
			if err != nil {
				return false
			}
			defer func() { _ = Close(db) }()

			// Create test table with NOT NULL constraint
			_, err = db.Exec(`CREATE TABLE test_data (id INTEGER NOT NULL, value INTEGER NOT NULL)`)
			if err != nil {
				return false
			}

			// Generate test data with one invalid row (NULL value)
			columns := []string{"id", "value"}
			rows := make([][]interface{}, validRowCount+1)
			for i := 0; i < validRowCount; i++ {
				rows[i] = []interface{}{i + 1, i * 10}
			}
			// Add invalid row with NULL
			rows[validRowCount] = []interface{}{validRowCount + 1, nil}

			// Execute batch insert (should fail)
			ctx := context.Background()
			err = BatchInsert(ctx, db, "test_data", columns, rows, 1000)

			if err == nil {
				return false // Should have failed
			}

			// Verify NO data was inserted (transaction rolled back)
			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM test_data").Scan(&count)
			if err != nil {
				return false
			}

			return count == 0
		},
		gen.IntRange(1, 100),
	))

	properties.TestingRun(t)
}
