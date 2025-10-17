# Error Recovery Strategies

This document describes the error recovery strategies available in the JSONL parser for handling malformed or invalid data.

## Overview

The parser provides two error handling modes for dealing with invalid JSONL lines:

1. **Skip Mode**: Continues parsing after errors, logging and skipping invalid lines
2. **Fail Fast Mode**: Stops immediately upon encountering the first error

## Error Modes

### ErrorModeSkip

In Skip mode, the parser:
- Continues parsing when encountering invalid JSON
- Logs each error with structured logging (zerolog)
- Tracks skipped line numbers
- Returns all successfully parsed records along with error details
- Supports configurable error thresholds

**Use Case**: When you want maximum data recovery from partially corrupted files.

### ErrorModeFailFast

In FailFast mode, the parser:
- Stops immediately upon the first error
- Returns all records parsed before the error
- Provides detailed error information about the failure point

**Use Case**: When data integrity is critical and any error indicates a serious problem.

## Usage

### Basic Usage

```go
import "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"

// Skip mode - continue on errors
records, errors := parser.ParseWithErrorHandling[MyType](
    "data.jsonl",
    parser.ErrorModeSkip,
    0, // 0 = unlimited errors
)

// FailFast mode - stop on first error
records, errors := parser.ParseWithErrorHandling[MyType](
    "data.jsonl",
    parser.ErrorModeFailFast,
    0, // ignored in FailFast mode
)
```

### Error Threshold Configuration

Control how many errors to tolerate before aborting:

```go
// Allow up to 10 errors before stopping
records, errors := parser.ParseWithErrorHandling[MyType](
    "data.jsonl",
    parser.ErrorModeSkip,
    10, // max 10 errors
)
```

**Note**: Setting `maxErrors` to 0 means unlimited errors (only in Skip mode).

### Detailed Error Reporting with ParseResult

For more detailed information about the parsing results:

```go
import "context"

ctx := context.Background()
result := parser.ParseWithErrorHandlingContext[MyType](
    ctx,
    "data.jsonl",
    parser.ErrorModeSkip,
    0,
)

// Access detailed information
fmt.Printf("Total lines: %d\n", result.TotalLines)
fmt.Printf("Successful records: %d\n", len(result.Records))
fmt.Printf("Skipped lines: %v\n", result.SkippedLines)
fmt.Printf("Error count: %d\n", len(result.Errors))

// Check error status
if result.HasErrors() {
    fmt.Println(result.ErrorSummary())
}

if result.HasFatalErrors() {
    // Handle fatal errors (file not found, threshold exceeded, etc.)
}
```

## Error Types

The parser integrates with the application's error classification system:

- **Skippable Errors**: Individual line parsing failures (invalid JSON)
- **Fatal Errors**: 
  - File not found
  - Scanner errors (file corruption)
  - Error threshold exceeded
  - Context cancellation

```go
import apperrors "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"

for _, err := range errors {
    if apperrors.IsSkippable(err) {
        // Handle skippable error
    } else if apperrors.IsFatal(err) {
        // Handle fatal error
    }
}
```

## Context Support

The parser supports context cancellation and timeouts:

```go
import (
    "context"
    "time"
)

// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result := parser.ParseWithErrorHandlingContext[MyType](
    ctx,
    "large-file.jsonl",
    parser.ErrorModeSkip,
    0,
)

if ctx.Err() != nil {
    // Parsing was cancelled or timed out
}
```

## Logging

All errors are logged using structured logging (zerolog):

```json
{
  "level": "warn",
  "line": 42,
  "error": "unexpected end of JSON input",
  "mode": "Skip",
  "message": "JSON parse error"
}
```

When parsing completes with errors:

```json
{
  "level": "info",
  "total_lines": 1000,
  "skipped_lines": 5,
  "successful_records": 995,
  "mode": "Skip",
  "message": "parsing completed with errors"
}
```

## Best Practices

### When to Use Skip Mode
- Processing large datasets where some data loss is acceptable
- Importing from external sources with potential quality issues
- Development/testing environments
- When you need maximum data recovery

### When to Use FailFast Mode
- Production imports where data integrity is critical
- Validating file format before full processing
- When any error indicates a serious problem
- Schema validation scenarios

### Error Threshold Guidelines
- **Development**: Set to 0 (unlimited) to see all issues
- **Production**: Set based on expected data quality
  - High quality data: 10-50 errors
  - Medium quality: 100-500 errors
  - Consider file size (e.g., 0.1% of total lines)

### Context Usage
- Always use context with timeouts for large files
- Set reasonable timeouts based on file size
- Handle context cancellation gracefully

## Performance Considerations

- **Skip Mode**: Slight overhead for error tracking and logging
- **FailFast Mode**: Faster when encountering early errors
- **Error Threshold**: Minimal overhead for checking count
- **Logging**: Structured logging is optimized (zero-allocation)

## Examples

See `internal/parser/example_error_handling_test.go` for runnable examples.

## Integration with ADR-005

This implementation follows [ADR-005: Error Handling Strategy](../../docs/adr/ADR-005-error-handling-strategy.md):

- ✅ Custom Error Types (AppError with ErrorType)
- ✅ Structured Logging (zerolog)
- ✅ Skippable vs Fatal error classification
- ✅ Context propagation
- ✅ Error wrapping with context

## API Reference

### Types

```go
type ErrorMode int
const (
    ErrorModeSkip ErrorMode = iota
    ErrorModeFailFast
)

type ParseResult[T any] struct {
    Records      []T
    Errors       []error
    SkippedLines []int
    TotalLines   int
}
```

### Functions

```go
// Simple interface
func ParseWithErrorHandling[T any](path string, mode ErrorMode, maxErrors int) ([]T, []error)

// Full interface with context
func ParseWithErrorHandlingContext[T any](ctx context.Context, path string, mode ErrorMode, maxErrors int) ParseResult[T]
```

### Methods

```go
func (r *ParseResult[T]) ErrorSummary() string
func (r *ParseResult[T]) HasErrors() bool
func (r *ParseResult[T]) HasFatalErrors() bool
```
