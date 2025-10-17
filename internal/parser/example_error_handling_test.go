package parser_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"
)

// Example_parseWithErrorHandling_skipMode demonstrates using Skip mode to parse a file with errors
func Example_parseWithErrorHandling_skipMode() {
	// Create a temporary test file with some invalid lines
	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "example_skip.jsonl")
	content := `{"id":1,"name":"Valid Item"}
invalid json line
{"id":2,"name":"Another Valid Item"}
`
	_ = os.WriteFile(testFile, []byte(content), 0644)
	defer func() { _ = os.Remove(testFile) }()

	// Parse with Skip mode - invalid lines are logged and skipped
	records, errors := parser.ParseWithErrorHandling[TestRow](
		testFile,
		parser.ErrorModeSkip,
		0, // 0 means unlimited errors
	)

	fmt.Printf("Parsed %d records\n", len(records))
	fmt.Printf("Encountered %d errors\n", len(errors))

	// Output:
	// Parsed 2 records
	// Encountered 1 errors
}

// Example_parseWithErrorHandling_failFast demonstrates using FailFast mode
func Example_parseWithErrorHandling_failFast() {
	// Create a temporary test file with an invalid line
	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "example_failfast.jsonl")
	content := `{"id":1,"name":"Valid Item"}
invalid json line
{"id":2,"name":"This won't be parsed"}
`
	_ = os.WriteFile(testFile, []byte(content), 0644)
	defer func() { _ = os.Remove(testFile) }()

	// Parse with FailFast mode - stops at first error
	records, errors := parser.ParseWithErrorHandling[TestRow](
		testFile,
		parser.ErrorModeFailFast,
		0,
	)

	fmt.Printf("Parsed %d records before error\n", len(records))
	fmt.Printf("Stopped with %d error(s)\n", len(errors))

	// Output:
	// Parsed 1 records before error
	// Stopped with 1 error(s)
}

// Example_parseWithErrorHandling_errorThreshold demonstrates error threshold configuration
func Example_parseWithErrorHandling_errorThreshold() {
	// Create a temporary test file with multiple invalid lines
	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "example_threshold.jsonl")
	content := `{"id":1,"name":"Valid"}
invalid line 1
invalid line 2
{"id":2,"name":"Valid"}
invalid line 3
`
	_ = os.WriteFile(testFile, []byte(content), 0644)
	defer func() { _ = os.Remove(testFile) }()

	// Parse with Skip mode and max 2 errors
	records, errors := parser.ParseWithErrorHandling[TestRow](
		testFile,
		parser.ErrorModeSkip,
		2, // Stop after 2 errors
	)

	fmt.Printf("Parsed %d records\n", len(records))
	fmt.Printf("Hit error threshold: %d errors\n", len(errors))

	// Output:
	// Parsed 1 records
	// Hit error threshold: 3 errors
}

// Example_parseResult demonstrates using ParseResult for detailed error reporting
func Example_parseResult() {
	// Create a temporary test file
	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "example_result.jsonl")
	content := `{"id":1,"name":"Item One"}
invalid line
{"id":2,"name":"Item Two"}
`
	_ = os.WriteFile(testFile, []byte(content), 0644)
	defer func() { _ = os.Remove(testFile) }()

	// Use ParseWithErrorHandlingContext for detailed results
	ctx := context.Background()
	result := parser.ParseWithErrorHandlingContext[TestRow](
		ctx,
		testFile,
		parser.ErrorModeSkip,
		0,
	)

	// Check result status
	if result.HasErrors() {
		fmt.Printf("Summary: %s\n", result.ErrorSummary())
		fmt.Printf("Skipped lines: %v\n", result.SkippedLines)
	}

	// Use the successfully parsed records
	fmt.Printf("Successfully parsed %d records\n", len(result.Records))

	// Output:
	// Summary: Encountered 1 error(s) while parsing 3 lines. Successfully parsed 2 records. Skipped 1 lines.
	// Skipped lines: [2]
	// Successfully parsed 2 records
}
