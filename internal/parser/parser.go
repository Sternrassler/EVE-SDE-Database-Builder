// Package parser provides JSONL parsing interfaces and implementations
// for EVE SDE data files.
package parser

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Parser defines the interface for parsing EVE SDE data files.
// Implementations must handle line-by-line JSONL parsing with error handling.
type Parser interface {
	// ParseFile reads and parses a JSONL file, returning a slice of parsed records.
	// The context can be used for cancellation and timeout control.
	ParseFile(ctx context.Context, path string) ([]interface{}, error)

	// TableName returns the database table name for the parsed data.
	TableName() string

	// Columns returns the list of column names for database insertion.
	Columns() []string
}

// JSONLParser is a generic parser for JSONL files that handles line-by-line parsing.
// It uses Go generics to provide type-safe parsing while maintaining a common interface.
type JSONLParser[T any] struct {
	tableName string
	columns   []string
}

// NewJSONLParser creates a new generic JSONL parser with the specified table name and columns.
func NewJSONLParser[T any](tableName string, columns []string) *JSONLParser[T] {
	return &JSONLParser[T]{
		tableName: tableName,
		columns:   columns,
	}
}

// ParseFile implements the Parser interface for JSONLParser.
// It reads the file line-by-line, unmarshals each JSON object, and returns all records.
// If the context is canceled, parsing stops and returns the context error.
func (p *JSONLParser[T]) ParseFile(ctx context.Context, path string) ([]interface{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	return p.parseReader(ctx, file)
}

// parseReader handles the actual line-by-line JSONL parsing from an io.Reader.
// This method is extracted to facilitate testing with different input sources.
func (p *JSONLParser[T]) parseReader(ctx context.Context, r io.Reader) ([]interface{}, error) {
	scanner := bufio.NewScanner(r)
	// Set buffer size for potentially large JSON lines (10MB max line size)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	var results []interface{}
	lineNum := 0

	for scanner.Scan() {
		lineNum++

		// Check for context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue // Skip empty lines
		}

		var item T
		if err := json.Unmarshal(line, &item); err != nil {
			return nil, fmt.Errorf("line %d: failed to parse JSON: %w", lineNum, err)
		}

		results = append(results, item)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error after line %d: %w", lineNum, err)
	}

	return results, nil
}

// TableName returns the database table name for this parser.
func (p *JSONLParser[T]) TableName() string {
	return p.tableName
}

// Columns returns the list of column names for database insertion.
func (p *JSONLParser[T]) Columns() []string {
	return p.columns
}
