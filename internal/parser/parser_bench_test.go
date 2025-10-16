package parser_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"
)

// BenchmarkParseJSONL_1k benchmarks parsing 1k lines
// Target: < 1s (establishes baseline for small files)
func BenchmarkParseJSONL_1k(b *testing.B) {
	benchmarkParseJSONL(b, 1000)
}

// BenchmarkParseJSONL_10k benchmarks parsing 10k lines
// Target: < 1s (tests medium-sized files)
func BenchmarkParseJSONL_10k(b *testing.B) {
	benchmarkParseJSONL(b, 10000)
}

// BenchmarkParseJSONL_100k benchmarks parsing 100k lines
// Target: < 1s (primary performance goal from epic)
func BenchmarkParseJSONL_100k(b *testing.B) {
	benchmarkParseJSONL(b, 100000)
}

// BenchmarkParseJSONL_500k benchmarks parsing 500k lines
// Target: < 5s (stress test for very large files)
func BenchmarkParseJSONL_500k(b *testing.B) {
	benchmarkParseJSONL(b, 500000)
}

// benchmarkParseJSONL is a helper function that benchmarks JSONL parsing with specified line count
// It creates a temporary test file, runs the parser, and measures performance with memory allocation tracking
func benchmarkParseJSONL(b *testing.B, lineCount int) {
	// Create temporary directory and test file
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, fmt.Sprintf("test_%d.jsonl", lineCount))

	// Generate JSONL content with realistic data structure
	// Using simple but representative JSON objects to match actual EVE SDE data patterns
	var content strings.Builder
	for i := 1; i <= lineCount; i++ {
		content.WriteString(fmt.Sprintf(`{"id":%d,"name":"Item_%d"}`, i, i))
		content.WriteString("\n")
	}

	// Write test file once (outside benchmark loop)
	if err := os.WriteFile(testFile, []byte(content.String()), 0644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	// Create parser instance
	p := parser.NewJSONLParser[TestRow]("test_table", []string{"id", "name"})
	ctx := context.Background()

	// Report memory allocations
	b.ReportAllocs()

	// Reset timer to exclude setup time
	b.ResetTimer()

	// Run benchmark iterations
	for i := 0; i < b.N; i++ {
		results, err := p.ParseFile(ctx, testFile)
		if err != nil {
			b.Fatalf("ParseFile failed: %v", err)
		}

		// Verify correct number of results (prevents compiler optimization)
		if len(results) != lineCount {
			b.Fatalf("Expected %d results, got %d", lineCount, len(results))
		}
	}

	b.StopTimer()
}

// BenchmarkParseJSONL_1k_NestedData benchmarks 1k lines with nested structures
// Tests performance impact of complex JSON parsing
func BenchmarkParseJSONL_1k_NestedData(b *testing.B) {
	benchmarkParseJSONLNested(b, 1000)
}

// BenchmarkParseJSONL_10k_NestedData benchmarks 10k lines with nested structures
func BenchmarkParseJSONL_10k_NestedData(b *testing.B) {
	benchmarkParseJSONLNested(b, 10000)
}

// BenchmarkParseJSONL_100k_NestedData benchmarks 100k lines with nested structures
func BenchmarkParseJSONL_100k_NestedData(b *testing.B) {
	benchmarkParseJSONLNested(b, 100000)
}

// benchmarkParseJSONLNested benchmarks parsing with nested data structures
// Simulates more realistic EVE SDE data with maps and complex structures
func benchmarkParseJSONLNested(b *testing.B, lineCount int) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, fmt.Sprintf("nested_%d.jsonl", lineCount))

	// Generate JSONL with nested structure (similar to invTypes with translations)
	var content strings.Builder
	for i := 1; i <= lineCount; i++ {
		content.WriteString(fmt.Sprintf(
			`{"typeID":%d,"typeName":{"en":"Item_%d","de":"Artikel_%d"},"mass":%f}`,
			i, i, i, float64(i)*0.01,
		))
		content.WriteString("\n")
	}

	if err := os.WriteFile(testFile, []byte(content.String()), 0644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	p := parser.NewJSONLParser[TestNestedRow]("invTypes", []string{"typeID", "typeName", "mass"})
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		results, err := p.ParseFile(ctx, testFile)
		if err != nil {
			b.Fatalf("ParseFile failed: %v", err)
		}

		if len(results) != lineCount {
			b.Fatalf("Expected %d results, got %d", lineCount, len(results))
		}
	}

	b.StopTimer()
}
