// Package parser provides JSONL parsing interfaces and implementations
// for EVE SDE data files.
package parser

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// StreamFile reads a JSONL file and streams parsed records through channels.
// It provides memory-efficient processing for large files using Go channels
// with backpressure handling and context cancellation support.
//
// The function returns two channels:
//   - dataChan: A channel that emits parsed records of type T
//   - errChan: A channel that emits a single error if parsing fails
//
// Both channels are closed when parsing completes or an error occurs.
// The dataChan is buffered (default 100 items) to handle backpressure.
//
// Context cancellation will stop parsing immediately and close both channels.
//
// Example usage:
//
//	ctx := context.Background()
//	dataChan, errChan := parser.StreamFile[TypeRow](ctx, "types.jsonl")
//
//	for item := range dataChan {
//	    // Process item
//	    if err := processItem(item); err != nil {
//	        // Handle error
//	    }
//	}
//
//	if err := <-errChan; err != nil {
//	    log.Fatal(err)
//	}
func StreamFile[T any](ctx context.Context, path string) (<-chan T, <-chan error) {
	dataChan := make(chan T, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(dataChan)
		defer close(errChan)

		// Open the file
		file, err := os.Open(path)
		if err != nil {
			errChan <- fmt.Errorf("failed to open file %s: %w", path, err)
			return
		}
		defer func() { _ = file.Close() }()

		// Create scanner for line-by-line reading
		scanner := bufio.NewScanner(file)
		// Set buffer size for potentially large JSON lines (10MB max line size)
		scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)

		lineNum := 0
		for scanner.Scan() {
			lineNum++

			// Check for context cancellation before processing each line
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			default:
			}

			line := scanner.Bytes()
			if len(line) == 0 {
				continue // Skip empty lines
			}

			// Parse JSON line
			var item T
			if err := json.Unmarshal(line, &item); err != nil {
				errChan <- fmt.Errorf("line %d: failed to parse JSON: %w", lineNum, err)
				return
			}

			// Send parsed item to channel (blocks if consumer is slow - backpressure)
			select {
			case dataChan <- item:
				// Successfully sent item
			case <-ctx.Done():
				// Context cancelled while waiting to send
				errChan <- ctx.Err()
				return
			}
		}

		// Check for scanner errors
		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("scanner error after line %d: %w", lineNum, err)
			return
		}

		// No error - send nil to indicate success
		errChan <- nil
	}()

	return dataChan, errChan
}
