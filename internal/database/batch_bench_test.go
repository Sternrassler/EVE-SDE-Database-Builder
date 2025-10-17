package database

import (
	"context"
	"fmt"
	"testing"
)

// BenchmarkBatchInsert_10k benchmarks insertion of 10k rows with default batch size
func BenchmarkBatchInsert_10k(b *testing.B) {
	benchmarkBatchInsertRows(b, 10000, 1000)
}

// BenchmarkBatchInsert_50k benchmarks insertion of 50k rows with default batch size
func BenchmarkBatchInsert_50k(b *testing.B) {
	benchmarkBatchInsertRows(b, 50000, 1000)
}

// BenchmarkBatchInsert_500k benchmarks insertion of 500k rows with default batch size
func BenchmarkBatchInsert_500k(b *testing.B) {
	benchmarkBatchInsertRows(b, 500000, 1000)
}

// BenchmarkBatchInsert_BatchSize_100 benchmarks batch size of 100 with 50k rows
func BenchmarkBatchInsert_BatchSize_100(b *testing.B) {
	benchmarkBatchInsertRows(b, 50000, 100)
}

// BenchmarkBatchInsert_BatchSize_500 benchmarks batch size of 500 with 50k rows
func BenchmarkBatchInsert_BatchSize_500(b *testing.B) {
	benchmarkBatchInsertRows(b, 50000, 500)
}

// BenchmarkBatchInsert_BatchSize_1000 benchmarks batch size of 1000 with 50k rows
func BenchmarkBatchInsert_BatchSize_1000(b *testing.B) {
	benchmarkBatchInsertRows(b, 50000, 1000)
}

// BenchmarkBatchInsert_BatchSize_5000 benchmarks batch size of 5000 with 50k rows
func BenchmarkBatchInsert_BatchSize_5000(b *testing.B) {
	benchmarkBatchInsertRows(b, 50000, 5000)
}

// benchmarkBatchInsertRows is a helper function that benchmarks batch insert with specified parameters
func benchmarkBatchInsertRows(b *testing.B, rowCount int, batchSize int) {
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

	// Prepare test data
	columns := []string{"id", "name", "value", "description"}
	rows := make([][]interface{}, rowCount)
	for i := 0; i < rowCount; i++ {
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
			_, err := db.Exec("DELETE FROM test_data")
			if err != nil {
				b.Fatalf("Failed to clear table: %v", err)
			}
		}

		err = BatchInsert(ctx, db, "test_data", columns, rows, batchSize)
		if err != nil {
			b.Fatalf("BatchInsert failed: %v", err)
		}
	}
	b.StopTimer()

	// Verify data was inserted correctly
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_data").Scan(&count)
	if err != nil {
		b.Fatalf("Failed to count rows: %v", err)
	}

	if count != rowCount {
		b.Errorf("Expected %d rows, got %d", rowCount, count)
	}
}
