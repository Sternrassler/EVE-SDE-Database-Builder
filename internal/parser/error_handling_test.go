package parser_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	apperrors "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"
)

// TestErrorMode_String tests the string representation of ErrorMode
func TestErrorMode_String(t *testing.T) {
	tests := []struct {
		mode     parser.ErrorMode
		expected string
	}{
		{parser.ErrorModeSkip, "Skip"},
		{parser.ErrorModeFailFast, "FailFast"},
		{parser.ErrorMode(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.expected {
				t.Errorf("ErrorMode.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestParseWithErrorHandling_SkipMode_ValidFile tests Skip mode with a valid file
func TestParseWithErrorHandling_SkipMode_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	content := `{"id":1,"name":"Item One"}
{"id":2,"name":"Item Two"}
{"id":3,"name":"Item Three"}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	records, errors := parser.ParseWithErrorHandling[TestRow](testFile, parser.ErrorModeSkip, 0)

	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %d: %v", len(errors), errors)
	}

	if len(records) != 3 {
		t.Errorf("Expected 3 records, got %d", len(records))
	}

	// Verify first record
	if records[0].ID != 1 || records[0].Name != "Item One" {
		t.Errorf("First record incorrect: got %+v", records[0])
	}
}

// TestParseWithErrorHandling_SkipMode_WithInvalidLines tests Skip mode with invalid JSON lines
func TestParseWithErrorHandling_SkipMode_WithInvalidLines(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	content := `{"id":1,"name":"Item One"}
{"id":2,"name":"Item Two"
{"id":3,"name":"Item Three"}
invalid json here
{"id":4,"name":"Item Four"}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	records, errors := parser.ParseWithErrorHandling[TestRow](testFile, parser.ErrorModeSkip, 0)

	// Should have 2 errors (lines 2 and 4)
	if len(errors) != 2 {
		t.Errorf("Expected 2 errors, got %d: %v", len(errors), errors)
	}

	// Should have successfully parsed 3 records (lines 1, 3, 5)
	if len(records) != 3 {
		t.Errorf("Expected 3 records, got %d", len(records))
	}

	// Verify records are correct
	expectedIDs := []int{1, 3, 4}
	for i, record := range records {
		if record.ID != expectedIDs[i] {
			t.Errorf("Record %d: expected ID %d, got %d", i, expectedIDs[i], record.ID)
		}
	}

	// Verify errors are marked as skippable
	for i, err := range errors {
		if !apperrors.IsSkippable(err) {
			t.Errorf("Error %d should be skippable: %v", i, err)
		}
	}
}

// TestParseWithErrorHandling_FailFastMode_WithInvalidLines tests FailFast mode
func TestParseWithErrorHandling_FailFastMode_WithInvalidLines(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	content := `{"id":1,"name":"Item One"}
{"id":2,"name":"Item Two"
{"id":3,"name":"Item Three"}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	records, errors := parser.ParseWithErrorHandling[TestRow](testFile, parser.ErrorModeFailFast, 0)

	// Should have exactly 1 error (first invalid line)
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d: %v", len(errors), errors)
	}

	// Should have only parsed 1 record (line 1) before failing
	if len(records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(records))
	}

	// Verify the successful record
	if records[0].ID != 1 {
		t.Errorf("Expected ID 1, got %d", records[0].ID)
	}
}

// TestParseWithErrorHandling_SkipMode_ErrorThreshold tests error threshold in Skip mode
func TestParseWithErrorHandling_SkipMode_ErrorThreshold(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	// Create file with 5 invalid lines
	content := `{"id":1,"name":"Item One"}
invalid line 1
invalid line 2
{"id":2,"name":"Item Two"}
invalid line 3
invalid line 4
{"id":3,"name":"Item Three"}
invalid line 5
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Set max errors to 3
	records, errors := parser.ParseWithErrorHandling[TestRow](testFile, parser.ErrorModeSkip, 3)

	// Should have 4 errors (3 parse errors + 1 threshold exceeded error)
	if len(errors) != 4 {
		t.Errorf("Expected 4 errors (3 parse + 1 threshold), got %d: %v", len(errors), errors)
	}

	// Should have parsed fewer than 3 records due to threshold
	if len(records) >= 3 {
		t.Errorf("Expected fewer than 3 records due to threshold, got %d", len(records))
	}

	// Last error should be fatal (threshold exceeded)
	lastErr := errors[len(errors)-1]
	if !apperrors.IsFatal(lastErr) {
		t.Errorf("Last error should be fatal (threshold exceeded): %v", lastErr)
	}

	// Check threshold error message
	if !strings.Contains(lastErr.Error(), "threshold exceeded") {
		t.Errorf("Last error should mention threshold: %v", lastErr)
	}
}

// TestParseWithErrorHandling_SkipMode_UnlimitedErrors tests Skip mode with no error limit
func TestParseWithErrorHandling_SkipMode_UnlimitedErrors(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	// Create file with many invalid lines
	var content strings.Builder
	for i := 1; i <= 10; i++ {
		content.WriteString(fmt.Sprintf(`{"id":%d,"name":"Item %d"}`, i, i))
		content.WriteString("\n")
		content.WriteString(fmt.Sprintf("invalid line %d", i))
		content.WriteString("\n")
	}

	if err := os.WriteFile(testFile, []byte(content.String()), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// maxErrors = 0 means unlimited
	records, errors := parser.ParseWithErrorHandling[TestRow](testFile, parser.ErrorModeSkip, 0)

	// Should have 10 errors (one per invalid line)
	if len(errors) != 10 {
		t.Errorf("Expected 10 errors, got %d", len(errors))
	}

	// Should have successfully parsed 10 records
	if len(records) != 10 {
		t.Errorf("Expected 10 records, got %d", len(records))
	}

	// All errors should be skippable
	for i, err := range errors {
		if !apperrors.IsSkippable(err) {
			t.Errorf("Error %d should be skippable: %v", i, err)
		}
	}
}

// TestParseWithErrorHandling_FileNotFound tests error handling for missing files
func TestParseWithErrorHandling_FileNotFound(t *testing.T) {
	records, errors := parser.ParseWithErrorHandling[TestRow]("/nonexistent/file.jsonl", parser.ErrorModeSkip, 0)

	if len(errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errors))
	}

	if len(records) != 0 {
		t.Errorf("Expected 0 records, got %d", len(records))
	}

	// Error should be fatal
	if !apperrors.IsFatal(errors[0]) {
		t.Errorf("File not found error should be fatal: %v", errors[0])
	}

	if !strings.Contains(errors[0].Error(), "failed to open file") {
		t.Errorf("Error should mention file open failure: %v", errors[0])
	}
}

// TestParseWithErrorHandling_EmptyFile tests parsing an empty file
func TestParseWithErrorHandling_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.jsonl")

	if err := os.WriteFile(testFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	records, errors := parser.ParseWithErrorHandling[TestRow](testFile, parser.ErrorModeSkip, 0)

	if len(errors) > 0 {
		t.Errorf("Expected no errors for empty file, got %d: %v", len(errors), errors)
	}

	if len(records) != 0 {
		t.Errorf("Expected 0 records, got %d", len(records))
	}
}

// TestParseWithErrorHandling_EmptyLines tests handling of empty lines
func TestParseWithErrorHandling_EmptyLines(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	content := `{"id":1,"name":"Item One"}

{"id":2,"name":"Item Two"}


{"id":3,"name":"Item Three"}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	records, errors := parser.ParseWithErrorHandling[TestRow](testFile, parser.ErrorModeSkip, 0)

	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %d: %v", len(errors), errors)
	}

	// Empty lines should be skipped silently
	if len(records) != 3 {
		t.Errorf("Expected 3 records, got %d", len(records))
	}
}

// TestParseWithErrorHandlingContext tests context cancellation
func TestParseWithErrorHandlingContext(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	// Create a large file
	var content strings.Builder
	for i := 1; i <= 10000; i++ {
		content.WriteString(`{"id":`)
		content.WriteString(string(rune('0' + i%10)))
		content.WriteString(`,"name":"Item"}` + "\n")
	}
	if err := os.WriteFile(testFile, []byte(content.String()), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result := parser.ParseWithErrorHandlingContext[TestRow](ctx, testFile, parser.ErrorModeSkip, 0)

	if !result.HasErrors() {
		t.Error("Expected errors due to context cancellation")
	}

	if !result.HasFatalErrors() {
		t.Error("Cancellation error should be fatal")
	}
}

// TestParseWithErrorHandlingContext_Timeout tests context timeout
func TestParseWithErrorHandlingContext_Timeout(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	// Create a large file
	var content strings.Builder
	for i := 1; i <= 10000; i++ {
		content.WriteString(`{"id":`)
		content.WriteString(string(rune('0' + i%10)))
		content.WriteString(`,"name":"Item"}` + "\n")
	}
	if err := os.WriteFile(testFile, []byte(content.String()), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Give context time to expire
	time.Sleep(10 * time.Millisecond)

	result := parser.ParseWithErrorHandlingContext[TestRow](ctx, testFile, parser.ErrorModeSkip, 0)

	if !result.HasErrors() {
		t.Error("Expected errors due to timeout")
	}
}

// TestParseResult_ErrorSummary tests the ErrorSummary method
func TestParseResult_ErrorSummary(t *testing.T) {
	tests := []struct {
		name     string
		result   parser.ParseResult[TestRow]
		contains string
	}{
		{
			name: "No errors",
			result: parser.ParseResult[TestRow]{
				Records:    []TestRow{{ID: 1, Name: "Test"}},
				Errors:     []error{},
				TotalLines: 1,
			},
			contains: "No errors",
		},
		{
			name: "With errors",
			result: parser.ParseResult[TestRow]{
				Records:      []TestRow{{ID: 1, Name: "Test"}},
				Errors:       []error{apperrors.NewSkippable("test error", nil)},
				SkippedLines: []int{2},
				TotalLines:   2,
			},
			contains: "Encountered 1 error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := tt.result.ErrorSummary()
			if !strings.Contains(summary, tt.contains) {
				t.Errorf("ErrorSummary() = %v, should contain %v", summary, tt.contains)
			}
		})
	}
}

// TestParseResult_HasErrors tests the HasErrors method
func TestParseResult_HasErrors(t *testing.T) {
	tests := []struct {
		name     string
		result   parser.ParseResult[TestRow]
		expected bool
	}{
		{
			name: "No errors",
			result: parser.ParseResult[TestRow]{
				Errors: []error{},
			},
			expected: false,
		},
		{
			name: "With errors",
			result: parser.ParseResult[TestRow]{
				Errors: []error{apperrors.NewSkippable("test", nil)},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.HasErrors(); got != tt.expected {
				t.Errorf("HasErrors() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestParseResult_HasFatalErrors tests the HasFatalErrors method
func TestParseResult_HasFatalErrors(t *testing.T) {
	tests := []struct {
		name     string
		result   parser.ParseResult[TestRow]
		expected bool
	}{
		{
			name: "No errors",
			result: parser.ParseResult[TestRow]{
				Errors: []error{},
			},
			expected: false,
		},
		{
			name: "Only skippable errors",
			result: parser.ParseResult[TestRow]{
				Errors: []error{apperrors.NewSkippable("test", nil)},
			},
			expected: false,
		},
		{
			name: "With fatal error",
			result: parser.ParseResult[TestRow]{
				Errors: []error{
					apperrors.NewSkippable("test", nil),
					apperrors.NewFatal("fatal", nil),
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.HasFatalErrors(); got != tt.expected {
				t.Errorf("HasFatalErrors() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestParseWithErrorHandling_NestedStructure tests parsing with nested structures
func TestParseWithErrorHandling_NestedStructure(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "nested.jsonl")

	content := `{"typeID":34,"typeName":{"en":"Tritanium"},"mass":0.01}
invalid line
{"typeID":35,"typeName":{"en":"Pyerite"},"mass":0.01}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	records, errors := parser.ParseWithErrorHandling[TestNestedRow](testFile, parser.ErrorModeSkip, 0)

	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	}

	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}

	// Verify nested structure was parsed correctly
	if records[0].TypeID != 34 {
		t.Errorf("Expected TypeID 34, got %d", records[0].TypeID)
	}
	if records[0].TypeName["en"] != "Tritanium" {
		t.Errorf("Expected Tritanium, got %s", records[0].TypeName["en"])
	}
}

// TestParseWithErrorHandling_LargeFile tests parsing a large file with errors
func TestParseWithErrorHandling_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "large.jsonl")

	// Create a file with 1000 lines, every 10th line is invalid
	var content strings.Builder
	for i := 1; i <= 1000; i++ {
		if i%10 == 0 {
			content.WriteString("invalid line\n")
		} else {
			content.WriteString(fmt.Sprintf(`{"id":%d,"name":"Item %d"}`, i, i))
			content.WriteString("\n")
		}
	}

	if err := os.WriteFile(testFile, []byte(content.String()), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	records, errors := parser.ParseWithErrorHandling[TestRow](testFile, parser.ErrorModeSkip, 0)

	// Should have 100 errors (every 10th line)
	if len(errors) != 100 {
		t.Errorf("Expected 100 errors, got %d", len(errors))
	}

	// Should have successfully parsed 900 records
	if len(records) != 900 {
		t.Errorf("Expected 900 records, got %d", len(records))
	}
}

// Benchmark tests for performance evaluation
func BenchmarkParseWithErrorHandling_SkipMode(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "bench.jsonl")

	// Create file with some invalid lines
	var content strings.Builder
	for i := 1; i <= 1000; i++ {
		if i%100 == 0 {
			content.WriteString("invalid line\n")
		} else {
			content.WriteString(`{"id":1,"name":"Item"}` + "\n")
		}
	}
	if err := os.WriteFile(testFile, []byte(content.String()), 0644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParseWithErrorHandling[TestRow](testFile, parser.ErrorModeSkip, 0)
	}
}

func BenchmarkParseWithErrorHandling_FailFastMode(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "bench.jsonl")

	// Create file with some invalid lines
	var content strings.Builder
	for i := 1; i <= 1000; i++ {
		if i%100 == 0 {
			content.WriteString("invalid line\n")
		} else {
			content.WriteString(`{"id":1,"name":"Item"}` + "\n")
		}
	}
	if err := os.WriteFile(testFile, []byte(content.String()), 0644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParseWithErrorHandling[TestRow](testFile, parser.ErrorModeFailFast, 0)
	}
}
