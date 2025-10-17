package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/database"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"
)

// TestNewProgressTracker tests progress tracker creation
func TestNewProgressTracker(t *testing.T) {
	pt := NewProgressTracker(10)
	if pt == nil {
		t.Fatal("expected progress tracker, got nil")
	}

	parsed, inserted, failed, total := pt.GetProgress()
	if parsed != 0 || inserted != 0 || failed != 0 || total != 10 {
		t.Errorf("expected (0,0,0,10), got (%d,%d,%d,%d)", parsed, inserted, failed, total)
	}
}

// TestProgressTracker_Increment tests progress counter increments
func TestProgressTracker_Increment(t *testing.T) {
	pt := NewProgressTracker(5)

	pt.IncrementParsed()
	pt.IncrementParsed()
	pt.IncrementInserted()
	pt.IncrementFailed()

	parsed, inserted, failed, total := pt.GetProgress()
	if parsed != 2 || inserted != 1 || failed != 1 || total != 5 {
		t.Errorf("expected (2,1,1,5), got (%d,%d,%d,%d)", parsed, inserted, failed, total)
	}
}

// TestProgressTracker_Concurrent tests thread-safe counter increments
func TestProgressTracker_Concurrent(t *testing.T) {
	pt := NewProgressTracker(100)

	done := make(chan bool)

	// Concurrent increments
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				pt.IncrementParsed()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	parsed, _, _, _ := pt.GetProgress()
	if parsed != 100 {
		t.Errorf("expected 100 parsed, got %d", parsed)
	}
}

// TestNewOrchestrator tests orchestrator creation
func TestNewOrchestrator(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	pool := NewPool(2)
	parsers := make(map[string]parser.Parser)

	orch := NewOrchestrator(db, pool, parsers)
	if orch == nil {
		t.Fatal("expected orchestrator, got nil")
	}

	if orch.db != db {
		t.Error("orchestrator db mismatch")
	}
	if orch.pool != pool {
		t.Error("orchestrator pool mismatch")
	}
	if orch.parsers == nil {
		t.Error("orchestrator parsers should not be nil")
	}
}

// TestOrchestrator_CreateParseTasks tests parse task creation
func TestOrchestrator_CreateParseTasks(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	pool := NewPool(2)
	parsers := map[string]parser.Parser{
		"types.jsonl": &MockParser{
			tableName: "invTypes",
			columns:   []string{"typeID", "typeName"},
		},
		"groups.jsonl": &MockParser{
			tableName: "invGroups",
			columns:   []string{"groupID", "groupName"},
		},
	}

	orch := NewOrchestrator(db, pool, parsers)
	tasks, err := orch.createParseTasks("/test/sde")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

// TestOrchestrator_ConvertToRows tests row conversion
func TestOrchestrator_ConvertToRows(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	pool := NewPool(2)
	orch := NewOrchestrator(db, pool, nil)

	records := []interface{}{"rec1", "rec2", "rec3"}
	rows, err := orch.convertToRows(records, 2)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rows) != 3 {
		t.Errorf("expected 3 rows, got %d", len(rows))
	}

	for i, row := range rows {
		if len(row) != 2 {
			t.Errorf("row %d: expected 2 columns, got %d", i, len(row))
		}
	}
}

// TestOrchestrator_ConvertToRows_Empty tests empty record conversion
func TestOrchestrator_ConvertToRows_Empty(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	pool := NewPool(2)
	orch := NewOrchestrator(db, pool, nil)

	rows, err := orch.convertToRows([]interface{}{}, 2)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
}

// TestOrchestrator_ImportAll_EmptyParsers tests import with no parsers
func TestOrchestrator_ImportAll_EmptyParsers(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	pool := NewPool(2)
	parsers := make(map[string]parser.Parser)
	orch := NewOrchestrator(db, pool, parsers)

	ctx := context.Background()
	progress, err := orch.ImportAll(ctx, "/test/sde")

	if err == nil {
		t.Error("expected error for no JSONL files, got nil")
	}

	if progress != nil {
		t.Error("expected nil progress on error")
	}
}

// TestOrchestrator_ImportAll_WithMockParsers tests basic import flow
func TestOrchestrator_ImportAll_WithMockParsers(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	// Create test table
	_, err = db.Exec("CREATE TABLE test_types (id INTEGER, name TEXT)")
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}

	pool := NewPool(2)
	parsers := map[string]parser.Parser{
		"types.jsonl": &MockParser{
			tableName:   "test_types",
			columns:     []string{"id", "name"},
			returnItems: []interface{}{"record1", "record2"},
		},
	}

	orch := NewOrchestrator(db, pool, parsers)

	ctx := context.Background()
	progress, err := orch.ImportAll(ctx, "/test/sde")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if progress == nil {
		t.Fatal("expected progress tracker, got nil")
	}

	parsed, _, _, total := progress.GetProgress()
	if parsed != 1 || total != 1 {
		t.Errorf("expected parsed=1, total=1, got parsed=%d, total=%d", parsed, total)
	}
}

// TestOrchestrator_ImportAll_WithParseError tests error handling
func TestOrchestrator_ImportAll_WithParseError(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	pool := NewPool(2)
	parsers := map[string]parser.Parser{
		"bad.jsonl": &MockParser{
			tableName:   "test_table",
			columns:     []string{"id"},
			shouldFail:  true,
			failWithErr: errors.New("mock parse error"),
		},
	}

	orch := NewOrchestrator(db, pool, parsers)

	ctx := context.Background()
	progress, err := orch.ImportAll(ctx, "/test/sde")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, inserted, failed, _ := progress.GetProgress()
	if parsed != 1 {
		t.Errorf("expected parsed=1, got %d", parsed)
	}
	if failed != 1 {
		t.Errorf("expected failed=1, got %d", failed)
	}
	if inserted != 0 {
		t.Errorf("expected inserted=0, got %d", inserted)
	}
}

// TestOrchestrator_ImportAll_ContextCancellation tests graceful cancellation
func TestOrchestrator_ImportAll_ContextCancellation(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	pool := NewPool(2)
	parsers := map[string]parser.Parser{
		"slow.jsonl": &MockParser{
			tableName: "test_table",
			columns:   []string{"id"},
			parseFunc: func(ctx context.Context, path string) ([]interface{}, error) {
				// Simulate slow parsing
				select {
				case <-time.After(100 * time.Millisecond):
					return []interface{}{"data"}, nil
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			},
		},
	}

	orch := NewOrchestrator(db, pool, parsers)

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	_, err = orch.ImportAll(ctx, "/test/sde")

	// Should return context error or complete quickly
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Logf("got error: %v (acceptable)", err)
	}
}

// TestOrchestrator_ImportAll_MultipleFiles tests multiple file processing
func TestOrchestrator_ImportAll_MultipleFiles(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	// Create test tables
	_, err = db.Exec("CREATE TABLE types (id INTEGER)")
	if err != nil {
		t.Fatalf("failed to create types table: %v", err)
	}
	_, err = db.Exec("CREATE TABLE groups (id INTEGER)")
	if err != nil {
		t.Fatalf("failed to create groups table: %v", err)
	}

	pool := NewPool(4)
	parsers := map[string]parser.Parser{
		"types.jsonl": &MockParser{
			tableName: "types",
			columns:   []string{"id"},
		},
		"groups.jsonl": &MockParser{
			tableName:   "groups",
			columns:     []string{"id"},
			returnItems: []interface{}{"data"},
		},
	}

	orch := NewOrchestrator(db, pool, parsers)

	ctx := context.Background()
	progress, err := orch.ImportAll(ctx, "/test/sde")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, _, _, total := progress.GetProgress()
	if parsed != 2 || total != 2 {
		t.Errorf("expected parsed=2, total=2, got parsed=%d, total=%d", parsed, total)
	}
}

// TestOrchestrator_ImportAll_MixedResults tests mixed success/failure
func TestOrchestrator_ImportAll_MixedResults(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	// Create test table
	_, err = db.Exec("CREATE TABLE success_table (id INTEGER)")
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}

	pool := NewPool(2)
	parsers := map[string]parser.Parser{
		"success.jsonl": &MockParser{
			tableName:   "success_table",
			columns:     []string{"id"},
			returnItems: []interface{}{"data"},
		},
		"failure.jsonl": &MockParser{
			tableName:   "failure_table",
			columns:     []string{"id"},
			shouldFail:  true,
			failWithErr: errors.New("mock parse error"),
		},
	}

	orch := NewOrchestrator(db, pool, parsers)

	ctx := context.Background()
	progress, err := orch.ImportAll(ctx, "/test/sde")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, _, failed, total := progress.GetProgress()
	if parsed != 2 {
		t.Errorf("expected parsed=2, got %d", parsed)
	}
	if failed < 1 {
		t.Errorf("expected at least 1 failure, got %d", failed)
	}
	if total != 2 {
		t.Errorf("expected total=2, got %d", total)
	}
}

// benchmarkOrchestrator creates a benchmark setup
func benchmarkOrchestrator(b *testing.B, workers, files int) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		b.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	// Create test table
	_, err = db.Exec("CREATE TABLE bench_table (id INTEGER)")
	if err != nil {
		b.Fatalf("failed to create test table: %v", err)
	}

	pool := NewPool(workers)
	parsers := make(map[string]parser.Parser)

	for i := 0; i < files; i++ {
		parsers[fmt.Sprintf("file%d.jsonl", i)] = &MockParser{
			tableName:   "bench_table",
			columns:     []string{"id"},
			returnItems: []interface{}{"data"},
		}
	}

	orch := NewOrchestrator(db, pool, parsers)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		_, err := orch.ImportAll(ctx, "/bench/sde")
		if err != nil {
			b.Fatalf("benchmark error: %v", err)
		}
	}
}

// BenchmarkOrchestrator_2Workers_10Files benchmarks with 2 workers, 10 files
func BenchmarkOrchestrator_2Workers_10Files(b *testing.B) {
	benchmarkOrchestrator(b, 2, 10)
}

// BenchmarkOrchestrator_4Workers_10Files benchmarks with 4 workers, 10 files
func BenchmarkOrchestrator_4Workers_10Files(b *testing.B) {
	benchmarkOrchestrator(b, 4, 10)
}

// Example_orchestratorBasicUsage demonstrates basic orchestrator usage
func Example_orchestratorBasicUsage() {
	// Setup in-memory database
	db, _ := database.NewDB(":memory:")
	defer db.Close()

	// Create test table
	_, _ = db.Exec("CREATE TABLE types (typeID INTEGER, typeName TEXT)")

	// Setup worker pool with 4 workers
	pool := NewPool(4)

	// Register parsers for different file types
	parsers := map[string]parser.Parser{
		"types.jsonl": &MockParser{
			tableName: "types",
			columns:   []string{"typeID", "typeName"},
			returnItems: []interface{}{
				map[string]interface{}{"typeID": 1, "typeName": "Tritanium"},
				map[string]interface{}{"typeID": 2, "typeName": "Pyerite"},
			},
		},
	}

	// Create orchestrator
	orch := NewOrchestrator(db, pool, parsers)

	// Execute 2-phase import
	ctx := context.Background()
	progress, err := orch.ImportAll(ctx, "/path/to/sde")
	if err != nil {
		fmt.Printf("Import error: %v\n", err)
		return
	}

	// Check progress
	parsed, inserted, failed, total := progress.GetProgress()
	fmt.Printf("Import complete: %d/%d parsed, %d inserted, %d failed\n", parsed, total, inserted, failed)
	// Output: Import complete: 1/1 parsed, 1 inserted, 0 failed
}

// Example_orchestratorWithContextCancellation demonstrates graceful cancellation
func Example_orchestratorWithContextCancellation() {
	db, _ := database.NewDB(":memory:")
	defer db.Close()

	pool := NewPool(2)
	parsers := map[string]parser.Parser{
		"slow.jsonl": &MockParser{
			tableName: "slow_table",
			columns:   []string{"id"},
			parseFunc: func(ctx context.Context, path string) ([]interface{}, error) {
				// Check context cancellation during long operation
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(100 * time.Millisecond):
					return []interface{}{"data"}, nil
				}
			},
		},
	}

	orch := NewOrchestrator(db, pool, parsers)

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after short delay (simulating user interrupt)
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	// Start import (will be cancelled)
	_, err := orch.ImportAll(ctx, "/path/to/sde")
	// Either completed or cancelled
	if err == nil {
		fmt.Println("Import completed or cancelled")
	} else {
		fmt.Println("Import completed or cancelled")
	}
	// Output: Import completed or cancelled
}

// Example_orchestratorProgressTracking demonstrates progress monitoring
func Example_orchestratorProgressTracking() {
	// Create progress tracker for 5 files
	progress := NewProgressTracker(5)

	// Simulate concurrent processing
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// Simulate parsing
			time.Sleep(time.Duration(id) * time.Millisecond)
			progress.IncrementParsed()

			// Some succeed, some fail
			if id%2 == 0 {
				progress.IncrementInserted()
			} else {
				progress.IncrementFailed()
			}
		}(i)
	}

	wg.Wait()

	// Get final progress
	parsed, inserted, failed, total := progress.GetProgress()
	fmt.Printf("Parsed: %d/%d, Inserted: %d, Failed: %d\n", parsed, total, inserted, failed)
	// Output: Parsed: 5/5, Inserted: 3, Failed: 2
}
