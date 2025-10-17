package parser_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"
)

// TestStreamFile_BasicFunctionality tests basic streaming with a small file
func TestStreamFile_BasicFunctionality(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	content := `{"id":1,"name":"Item One"}
{"id":2,"name":"Item Two"}
{"id":3,"name":"Item Three"}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ctx := context.Background()
	dataChan, errChan := parser.StreamFile[TestRow](ctx, testFile)

	var results []TestRow
	for item := range dataChan {
		results = append(results, item)
	}

	err := <-errChan
	if err != nil {
		t.Fatalf("StreamFile failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 records, got %d", len(results))
	}

	// Verify first record
	if results[0].ID != 1 || results[0].Name != "Item One" {
		t.Errorf("First record incorrect: got %+v", results[0])
	}
}

// TestStreamFile_EmptyLines tests that empty lines are skipped
func TestStreamFile_EmptyLines(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	content := `{"id":1,"name":"Item One"}

{"id":2,"name":"Item Two"}


{"id":3,"name":"Item Three"}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ctx := context.Background()
	dataChan, errChan := parser.StreamFile[TestRow](ctx, testFile)

	count := 0
	for range dataChan {
		count++
	}

	err := <-errChan
	if err != nil {
		t.Fatalf("StreamFile failed: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 records (empty lines skipped), got %d", count)
	}
}

// TestStreamFile_FileNotFound tests error handling for missing file
func TestStreamFile_FileNotFound(t *testing.T) {
	ctx := context.Background()
	dataChan, errChan := parser.StreamFile[TestRow](ctx, "/nonexistent/file.jsonl")

	// Drain data channel (should be empty)
	for range dataChan {
		t.Error("Should not receive any data for nonexistent file")
	}

	err := <-errChan
	if err == nil {
		t.Fatal("Expected error for nonexistent file, got nil")
	}

	if !strings.Contains(err.Error(), "failed to open file") {
		t.Errorf("Expected 'failed to open file' error, got: %v", err)
	}
}

// TestStreamFile_InvalidJSON tests error handling for malformed JSON
func TestStreamFile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	content := `{"id":1,"name":"Item One"}
{"id":2,"name":"Item Two"
{"id":3,"name":"Item Three"}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ctx := context.Background()
	dataChan, errChan := parser.StreamFile[TestRow](ctx, testFile)

	count := 0
	for range dataChan {
		count++
	}

	err := <-errChan
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}

	// Error should include line number
	if !strings.Contains(err.Error(), "line 2") {
		t.Errorf("Error should mention line number: %v", err)
	}

	// Should have received one valid item before error
	if count != 1 {
		t.Errorf("Expected 1 item before error, got %d", count)
	}
}

// TestStreamFile_ContextCancellation tests that context cancellation stops parsing
func TestStreamFile_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	// Create a file with many lines
	var content strings.Builder
	for i := 1; i <= 1000; i++ {
		content.WriteString(fmt.Sprintf(`{"id":%d,"name":"Item %d"}`, i, i))
		content.WriteString("\n")
	}
	if err := os.WriteFile(testFile, []byte(content.String()), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure cancel is called on all paths
	dataChan, errChan := parser.StreamFile[TestRow](ctx, testFile)

	// Read a few items then cancel
	count := 0
	for range dataChan {
		count++
		if count == 10 {
			cancel()
		}
		// Continue draining channel until it closes
	}

	err := <-errChan
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got: %v", err)
	}

	// Should have received around 10 items (might be slightly more due to buffering)
	if count > 120 {
		t.Errorf("Expected around 10 items before cancellation, got %d", count)
	}
}

// TestStreamFile_ContextTimeout tests that context timeout stops parsing
func TestStreamFile_ContextTimeout(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	// Create a file with many lines
	var content strings.Builder
	for i := 1; i <= 10000; i++ {
		content.WriteString(fmt.Sprintf(`{"id":%d,"name":"Item %d"}`, i, i))
		content.WriteString("\n")
	}
	if err := os.WriteFile(testFile, []byte(content.String()), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	dataChan, errChan := parser.StreamFile[TestRow](ctx, testFile)

	// Drain channel slowly to allow timeout to trigger
	count := 0
	for range dataChan {
		count++
		time.Sleep(100 * time.Microsecond) // Slow consumer
	}

	err := <-errChan
	if err != context.DeadlineExceeded && err != context.Canceled {
		t.Errorf("Expected context timeout error, got: %v", err)
	}
}

// TestStreamFile_Backpressure tests that streaming handles slow consumers properly
func TestStreamFile_Backpressure(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	// Create a file with many lines
	var content strings.Builder
	for i := 1; i <= 500; i++ {
		content.WriteString(fmt.Sprintf(`{"id":%d,"name":"Item %d"}`, i, i))
		content.WriteString("\n")
	}
	if err := os.WriteFile(testFile, []byte(content.String()), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ctx := context.Background()
	dataChan, errChan := parser.StreamFile[TestRow](ctx, testFile)

	// Slow consumer
	count := 0
	for range dataChan {
		count++
		time.Sleep(1 * time.Millisecond) // Simulate slow processing
	}

	err := <-errChan
	if err != nil {
		t.Fatalf("StreamFile failed: %v", err)
	}

	if count != 500 {
		t.Errorf("Expected 500 items, got %d", count)
	}
}

// TestStreamFile_ConcurrentConsumers tests multiple consumers reading from the same stream
func TestStreamFile_ConcurrentConsumers(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	// Create test file
	var content strings.Builder
	for i := 1; i <= 100; i++ {
		content.WriteString(fmt.Sprintf(`{"id":%d,"name":"Item %d"}`, i, i))
		content.WriteString("\n")
	}
	if err := os.WriteFile(testFile, []byte(content.String()), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ctx := context.Background()
	dataChan, errChan := parser.StreamFile[TestRow](ctx, testFile)

	// Multiple goroutines consuming from the same channel
	var wg sync.WaitGroup
	var mu sync.Mutex
	totalCount := 0

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			count := 0
			for range dataChan {
				count++
			}
			mu.Lock()
			totalCount += count
			mu.Unlock()
		}()
	}

	wg.Wait()
	err := <-errChan
	if err != nil {
		t.Fatalf("StreamFile failed: %v", err)
	}

	// Total items consumed should be 100 (distributed among consumers)
	if totalCount != 100 {
		t.Errorf("Expected 100 total items consumed, got %d", totalCount)
	}
}

// TestStreamFile_EmptyFile tests handling of empty files
func TestStreamFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.jsonl")

	if err := os.WriteFile(testFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ctx := context.Background()
	dataChan, errChan := parser.StreamFile[TestRow](ctx, testFile)

	count := 0
	for range dataChan {
		count++
	}

	err := <-errChan
	if err != nil {
		t.Fatalf("StreamFile failed: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 records for empty file, got %d", count)
	}
}

// TestStreamFile_NestedStructure tests streaming with nested JSON structures
func TestStreamFile_NestedStructure(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "types.jsonl")

	content := `{"typeID":34,"typeName":{"en":"Tritanium","de":"Tritanium"},"mass":0.01}
{"typeID":35,"typeName":{"en":"Pyerite","de":"Pyerit"},"mass":0.01}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ctx := context.Background()
	dataChan, errChan := parser.StreamFile[TestNestedRow](ctx, testFile)

	var results []TestNestedRow
	for item := range dataChan {
		results = append(results, item)
	}

	err := <-errChan
	if err != nil {
		t.Fatalf("StreamFile failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 records, got %d", len(results))
	}

	if results[0].TypeID != 34 {
		t.Errorf("Expected TypeID 34, got %d", results[0].TypeID)
	}
	if results[0].TypeName["en"] != "Tritanium" {
		t.Errorf("Expected TypeName[en] 'Tritanium', got '%s'", results[0].TypeName["en"])
	}
}

// TestStreamFile_MemoryEfficiency tests memory usage for large files
func TestStreamFile_MemoryEfficiency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory efficiency test in short mode")
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "large.jsonl")

	// Create a file with 500k lines (similar to invTypes.jsonl)
	file, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Each line is roughly 100 bytes
	for i := 1; i <= 500000; i++ {
		_, _ = fmt.Fprintf(file, `{"id":%d,"name":"Item %d with some extra data to make it realistic"}`, i, i)
		_, _ = fmt.Fprintln(file)
	}
	_ = file.Close()

	// Measure memory before
	runtime.GC()
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	ctx := context.Background()
	dataChan, errChan := parser.StreamFile[TestRow](ctx, testFile)

	// Process items without accumulating them in memory
	count := 0
	var peakMem uint64
	for item := range dataChan {
		count++
		_ = item // Simulate processing

		// Check memory periodically
		if count%10000 == 0 {
			runtime.GC()
			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)
			allocated := mem.Alloc - memBefore.Alloc
			if allocated > peakMem {
				peakMem = allocated
			}
		}
	}

	err = <-errChan
	if err != nil {
		t.Fatalf("StreamFile failed: %v", err)
	}

	if count != 500000 {
		t.Errorf("Expected 500000 items, got %d", count)
	}

	// Check final memory usage
	runtime.GC()
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	memUsedMB := float64(memAfter.Alloc-memBefore.Alloc) / 1024 / 1024
	peakMemMB := float64(peakMem) / 1024 / 1024

	t.Logf("Memory usage: %.2f MB (peak: %.2f MB)", memUsedMB, peakMemMB)

	// Requirement: Memory < 100MB for 500k lines
	// We use peak memory to account for periodic spikes
	if peakMemMB > 100 {
		t.Errorf("Memory usage %.2f MB exceeds requirement of 100 MB", peakMemMB)
	}
}

// BenchmarkStreamFile_LargeFile benchmarks streaming performance
func BenchmarkStreamFile_LargeFile(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	// Create a file with 10k lines
	var content strings.Builder
	for i := 1; i <= 10000; i++ {
		content.WriteString(fmt.Sprintf(`{"id":%d,"name":"Item %d"}`, i, i))
		content.WriteString("\n")
	}
	if err := os.WriteFile(testFile, []byte(content.String()), 0644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dataChan, errChan := parser.StreamFile[TestRow](ctx, testFile)

		for range dataChan {
			// Consume items
		}

		if err := <-errChan; err != nil {
			b.Fatalf("StreamFile failed: %v", err)
		}
	}
}

// BenchmarkStreamFile_vs_ParseFile compares streaming vs batch parsing
func BenchmarkStreamFile_vs_ParseFile(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	// Create test file with 1000 lines
	var content strings.Builder
	for i := 1; i <= 1000; i++ {
		content.WriteString(fmt.Sprintf(`{"id":%d,"name":"Item %d"}`, i, i))
		content.WriteString("\n")
	}
	if err := os.WriteFile(testFile, []byte(content.String()), 0644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	ctx := context.Background()

	b.Run("StreamFile", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			dataChan, errChan := parser.StreamFile[TestRow](ctx, testFile)
			for range dataChan {
			}
			<-errChan
		}
	})

	b.Run("ParseFile", func(b *testing.B) {
		p := parser.NewJSONLParser[TestRow]("test_table", []string{"id", "name"})
		for i := 0; i < b.N; i++ {
			_, err := p.ParseFile(ctx, testFile)
			if err != nil {
				b.Fatalf("ParseFile failed: %v", err)
			}
		}
	})
}
