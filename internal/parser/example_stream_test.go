package parser_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"
)

// ExampleStreamFile demonstrates basic usage of the streaming parser
func ExampleStreamFile() {
	// Create a temporary test file
	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "example_stream.jsonl")
	content := `{"id":1,"name":"Tritanium"}
{"id":2,"name":"Pyerite"}
{"id":3,"name":"Mexallon"}
`
	_ = os.WriteFile(testFile, []byte(content), 0644)
	defer func() { _ = os.Remove(testFile) }()

	// Stream the file
	ctx := context.Background()
	dataChan, errChan := parser.StreamFile[TestRow](ctx, testFile)

	// Process items as they arrive
	count := 0
	for item := range dataChan {
		count++
		fmt.Printf("Item %d: %s\n", item.ID, item.Name)
	}

	// Check for errors
	if err := <-errChan; err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total items: %d\n", count)
	// Output:
	// Item 1: Tritanium
	// Item 2: Pyerite
	// Item 3: Mexallon
	// Total items: 3
}

// ExampleStreamFile_withCancellation demonstrates context cancellation
func ExampleStreamFile_withCancellation() {
	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "example_cancel.jsonl")

	// Create a large file to ensure cancellation happens before all items are read
	file, _ := os.Create(testFile)
	for i := 1; i <= 1000; i++ {
		_, _ = fmt.Fprintf(file, `{"id":%d,"name":"Item %d"}`, i, i)
		_, _ = fmt.Fprintln(file)
	}
	_ = file.Close()
	defer func() { _ = os.Remove(testFile) }()

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dataChan, errChan := parser.StreamFile[TestRow](ctx, testFile)

	// Process only the first 2 items
	count := 0
	for item := range dataChan {
		if count < 2 {
			fmt.Printf("Processing: %s\n", item.Name)
		}
		count++
		if count == 2 {
			cancel() // Stop processing
		}
	}

	// Check error (should be context.Canceled)
	err := <-errChan
	if err == context.Canceled {
		fmt.Println("Cancelled successfully")
	}

	// Output:
	// Processing: Item 1
	// Processing: Item 2
	// Cancelled successfully
}

// ExampleStreamFile_memoryEfficient demonstrates memory-efficient processing of large files
func ExampleStreamFile_memoryEfficient() {
	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "example_large.jsonl")

	// Create a large file (simulated)
	file, _ := os.Create(testFile)
	for i := 1; i <= 100; i++ {
		_, _ = fmt.Fprintf(file, `{"id":%d,"name":"Item %d"}`, i, i)
		_, _ = fmt.Fprintln(file)
	}
	_ = file.Close()
	defer func() { _ = os.Remove(testFile) }()

	ctx := context.Background()
	dataChan, errChan := parser.StreamFile[TestRow](ctx, testFile)

	// Process items without accumulating them in memory
	count := 0
	for item := range dataChan {
		// Process each item individually
		// (no accumulation in memory)
		_ = item
		count++
	}

	if err := <-errChan; err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Processed %d items with minimal memory\n", count)
	// Output:
	// Processed 100 items with minimal memory
}
