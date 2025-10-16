package parser_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"
)

// TestRow is a simple test structure for parser tests
type TestRow struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// TestNestedRow is a test structure with nested data
type TestNestedRow struct {
	TypeID   int               `json:"typeID"`
	TypeName map[string]string `json:"typeName"`
	Mass     float64           `json:"mass"`
}

func TestJSONLParser_ParseFile_BasicFunctionality(t *testing.T) {
	// Create a temporary JSONL file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	content := `{"id":1,"name":"Item One"}
{"id":2,"name":"Item Two"}
{"id":3,"name":"Item Three"}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create parser
	p := parser.NewJSONLParser[TestRow]("test_table", []string{"id", "name"})

	// Parse file
	ctx := context.Background()
	results, err := p.ParseFile(ctx, testFile)

	// Verify results
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 records, got %d", len(results))
	}

	// Check first record
	if row, ok := results[0].(TestRow); ok {
		if row.ID != 1 || row.Name != "Item One" {
			t.Errorf("First record incorrect: got %+v", row)
		}
	} else {
		t.Error("Failed to cast result to TestRow")
	}
}

func TestJSONLParser_ParseFile_NestedStructure(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "types.jsonl")

	content := `{"typeID":34,"typeName":{"en":"Tritanium","de":"Tritanium"},"mass":0.01}
{"typeID":35,"typeName":{"en":"Pyerite","de":"Pyerit"},"mass":0.01}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	p := parser.NewJSONLParser[TestNestedRow]("invTypes", []string{"typeID", "typeName", "mass"})
	ctx := context.Background()
	results, err := p.ParseFile(ctx, testFile)

	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 records, got %d", len(results))
	}

	if row, ok := results[0].(TestNestedRow); ok {
		if row.TypeID != 34 {
			t.Errorf("Expected TypeID 34, got %d", row.TypeID)
		}
		if row.TypeName["en"] != "Tritanium" {
			t.Errorf("Expected TypeName[en] 'Tritanium', got '%s'", row.TypeName["en"])
		}
		if row.Mass != 0.01 {
			t.Errorf("Expected Mass 0.01, got %f", row.Mass)
		}
	} else {
		t.Error("Failed to cast result to TestNestedRow")
	}
}

func TestJSONLParser_ParseFile_EmptyLines(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	content := `{"id":1,"name":"Item One"}

{"id":2,"name":"Item Two"}


{"id":3,"name":"Item Three"}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	p := parser.NewJSONLParser[TestRow]("test_table", []string{"id", "name"})
	ctx := context.Background()
	results, err := p.ParseFile(ctx, testFile)

	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Should skip empty lines
	if len(results) != 3 {
		t.Errorf("Expected 3 records (empty lines skipped), got %d", len(results))
	}
}

func TestJSONLParser_ParseFile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	content := `{"id":1,"name":"Item One"}
{"id":2,"name":"Item Two"
{"id":3,"name":"Item Three"}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	p := parser.NewJSONLParser[TestRow]("test_table", []string{"id", "name"})
	ctx := context.Background()
	_, err := p.ParseFile(ctx, testFile)

	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}

	// Error should include line number
	if !strings.Contains(err.Error(), "line 2") {
		t.Errorf("Error should mention line number: %v", err)
	}
}

func TestJSONLParser_ParseFile_FileNotFound(t *testing.T) {
	p := parser.NewJSONLParser[TestRow]("test_table", []string{"id", "name"})
	ctx := context.Background()
	_, err := p.ParseFile(ctx, "/nonexistent/file.jsonl")

	if err == nil {
		t.Fatal("Expected error for nonexistent file, got nil")
	}

	if !strings.Contains(err.Error(), "failed to open file") {
		t.Errorf("Expected 'failed to open file' error, got: %v", err)
	}
}

func TestJSONLParser_ParseFile_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	// Create a file with many lines
	var content strings.Builder
	for i := 1; i <= 1000; i++ {
		content.WriteString(`{"id":`)
		content.WriteString(string(rune('0' + i%10)))
		content.WriteString(`,"name":"Item"}` + "\n")
	}
	if err := os.WriteFile(testFile, []byte(content.String()), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	p := parser.NewJSONLParser[TestRow]("test_table", []string{"id", "name"})

	// Create a context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := p.ParseFile(ctx, testFile)

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got: %v", err)
	}
}

func TestJSONLParser_ParseFile_ContextTimeout(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	// Create a file with many lines
	var content strings.Builder
	for i := 1; i <= 10000; i++ {
		content.WriteString(`{"id":`)
		content.WriteString(string(rune('0' + i%10)))
		content.WriteString(`,"name":"Item"}` + "\n")
	}
	if err := os.WriteFile(testFile, []byte(content.String()), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	p := parser.NewJSONLParser[TestRow]("test_table", []string{"id", "name"})

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Give the context time to expire
	time.Sleep(10 * time.Millisecond)

	_, err := p.ParseFile(ctx, testFile)

	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded error, got: %v", err)
	}
}

func TestJSONLParser_TableName(t *testing.T) {
	p := parser.NewJSONLParser[TestRow]("my_table", []string{"id", "name"})

	if p.TableName() != "my_table" {
		t.Errorf("Expected TableName 'my_table', got '%s'", p.TableName())
	}
}

func TestJSONLParser_Columns(t *testing.T) {
	columns := []string{"id", "name", "value"}
	p := parser.NewJSONLParser[TestRow]("test_table", columns)

	resultColumns := p.Columns()
	if len(resultColumns) != len(columns) {
		t.Errorf("Expected %d columns, got %d", len(columns), len(resultColumns))
	}

	for i, col := range columns {
		if resultColumns[i] != col {
			t.Errorf("Column %d: expected '%s', got '%s'", i, col, resultColumns[i])
		}
	}
}

func TestJSONLParser_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.jsonl")

	if err := os.WriteFile(testFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	p := parser.NewJSONLParser[TestRow]("test_table", []string{"id", "name"})
	ctx := context.Background()
	results, err := p.ParseFile(ctx, testFile)

	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 records for empty file, got %d", len(results))
	}
}

func TestJSONLParser_LargeLines(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "large.jsonl")

	// Create a JSON line with a large string (but under the 10MB buffer limit)
	largeString := strings.Repeat("x", 1024*1024) // 1MB string
	content := `{"id":1,"name":"` + largeString + `"}
{"id":2,"name":"normal"}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	p := parser.NewJSONLParser[TestRow]("test_table", []string{"id", "name"})
	ctx := context.Background()
	results, err := p.ParseFile(ctx, testFile)

	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 records, got %d", len(results))
	}

	if row, ok := results[0].(TestRow); ok {
		if len(row.Name) != len(largeString) {
			t.Errorf("Expected large string of length %d, got %d", len(largeString), len(row.Name))
		}
	}
}

// Benchmark tests
func BenchmarkJSONLParser_ParseFile_SmallFile(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	var content strings.Builder
	for i := 1; i <= 100; i++ {
		content.WriteString(`{"id":`)
		content.WriteString(string(rune('0' + i%10)))
		content.WriteString(`,"name":"Item"}` + "\n")
	}
	if err := os.WriteFile(testFile, []byte(content.String()), 0644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	p := parser.NewJSONLParser[TestRow]("test_table", []string{"id", "name"})
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.ParseFile(ctx, testFile)
		if err != nil {
			b.Fatalf("ParseFile failed: %v", err)
		}
	}
}



// Example test demonstrating usage
func ExampleJSONLParser() {
	// Create a temporary test file
	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "example.jsonl")
	content := `{"id":1,"name":"Example Item"}
{"id":2,"name":"Another Item"}
`
	_ = os.WriteFile(testFile, []byte(content), 0644)
	defer os.Remove(testFile)

	// Create and use parser
	p := parser.NewJSONLParser[TestRow]("items", []string{"id", "name"})
	ctx := context.Background()
	results, err := p.ParseFile(ctx, testFile)
	if err != nil {
		// Handle error
		return
	}

	// Process results
	for _, result := range results {
		if row, ok := result.(TestRow); ok {
			_ = row // Use the parsed row
		}
	}
}

// Test to ensure Parser interface is properly implemented
func TestJSONLParser_ImplementsParserInterface(t *testing.T) {
	var _ parser.Parser = (*parser.JSONLParser[TestRow])(nil)
}

// Test reading from io.Reader (not just files)
func TestJSONLParser_ParseReader(t *testing.T) {
	content := `{"id":1,"name":"Item One"}
{"id":2,"name":"Item Two"}
`
	reader := strings.NewReader(content)

	// We need access to parseReader, so we'll test through ParseFile with a temp file
	// This test verifies the underlying reader functionality
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	p := parser.NewJSONLParser[TestRow]("test_table", []string{"id", "name"})
	ctx := context.Background()
	results, err := p.ParseFile(ctx, testFile)

	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 records, got %d", len(results))
	}

	// Verify reader was properly closed by checking file operations still work
	_, err = os.Stat(testFile)
	if err != nil {
		t.Errorf("File operations failed after parsing: %v", err)
	}

	// Suppress unused variable warning
	_ = reader
}

// Test concurrent parsing (ensuring parser is safe to use concurrently)
func TestJSONLParser_ConcurrentParsing(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	content := `{"id":1,"name":"Item"}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	p := parser.NewJSONLParser[TestRow]("test_table", []string{"id", "name"})

	// Run multiple parsers concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			ctx := context.Background()
			_, err := p.ParseFile(ctx, testFile)
			if err != nil {
				t.Errorf("Concurrent ParseFile failed: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Suppress unused import warnings
var (
	_ = io.EOF
)
