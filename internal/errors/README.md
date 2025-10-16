# Custom Error Types Package

This package provides structured error handling with classification for the EVE SDE Database Builder.

## Features

- **Error Classification**: Four error types for precise error handling
  - `ErrorTypeFatal`: Critical errors that cannot be recovered
  - `ErrorTypeRetryable`: Errors that may succeed on retry
  - `ErrorTypeValidation`: Input validation errors
  - `ErrorTypeSkippable`: Errors that can be safely skipped

- **Rich Error Context**: Add contextual information to errors using `WithContext()`
- **Error Chaining**: Full support for Go 1.13+ error wrapping with `Unwrap()`, `Is()`, and `As()`
- **Type Checking Helpers**: Convenient functions to check error types

## Usage

### Creating Errors

```go
import "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"

// Fatal error - critical failure
err := errors.NewFatal("database connection failed", causeErr)

// Retryable error - temporary failure
err := errors.NewRetryable("API request failed", causeErr)

// Validation error - invalid input
err := errors.NewValidation("invalid email format", nil)

// Skippable error - optional operation failed
err := errors.NewSkippable("optional field missing", nil)
```

### Adding Context

The `WithContext()` method provides a fluent, chainable API for adding debugging information:

```go
err := errors.NewRetryable("DB locked", errors.New("database is locked")).
    WithContext("table", "invTypes").
    WithContext("operation", "batch_insert").
    WithContext("retry_attempt", 3)

// Error output includes context:
// [Retryable] DB locked (table=invTypes, operation=batch_insert, retry_attempt=3): database is locked
fmt.Println(err.Error())
```

**Context in Error Messages:**
- Context appears in parentheses after the message
- Multiple entries are comma-separated
- Map iteration order is non-deterministic (Go behavior)
- Empty context results in no parentheses

**Supported Value Types:**
```go
err := errors.NewValidation("invalid data", nil).
    WithContext("field", "email").           // string
    WithContext("line", 42).                 // int
    WithContext("required", true).           // bool
    WithContext("value", 3.14).              // float
    WithContext("tags", []string{"a", "b"}). // slice
    WithContext("metadata", map[string]int{"x": 1}) // map

// Access context directly
fmt.Printf("Field: %v\n", err.Context["field"])
```

### Type Checking

```go
if errors.IsFatal(err) {
    log.Fatal("Fatal error occurred", err)
}

if errors.IsRetryable(err) {
    // Retry the operation
}

if errors.IsValidation(err) {
    // Return validation error to user
}

if errors.IsSkippable(err) {
    // Log and continue
}
```

### Error Wrapping and Unwrapping

```go
// Create an error
appErr := errors.NewRetryable("service unavailable", rootCause)

// Wrap it
wrapped := fmt.Errorf("operation failed: %w", appErr)

// Unwrap and check type
var unwrapped *errors.AppError
if errors.As(wrapped, &unwrapped) {
    fmt.Printf("Error type: %v\n", unwrapped.Type)
}
```

### Integration with Logger

The logger package provides automatic context extraction from `AppError` instances:

```go
import (
    "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"
    "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/logger"
)

log := logger.NewLogger("info", "json")

err := errors.NewRetryable("DB locked", errors.New("database is locked")).
    WithContext("table", "invTypes").
    WithContext("operation", "batch_insert").
    WithContext("retry_attempt", 3)

// Automatically extracts and logs all context fields
log.LogAppError(err)

// JSON output:
// {
//   "level": "error",
//   "error_type": "Retryable",
//   "message": "DB locked",
//   "table": "invTypes",
//   "operation": "batch_insert",
//   "retry_attempt": 3,
//   "cause": "database is locked"
// }
```

**Available Logger Methods:**
- `LogAppError(err error)` - Automatically extracts context from AppError
- `LogError(err error, context map[string]interface{})` - Manual context passing
```

## API Reference

### Types

#### `ErrorType`
```go
type ErrorType int

const (
    ErrorTypeFatal      ErrorType = iota
    ErrorTypeRetryable
    ErrorTypeValidation
    ErrorTypeSkippable
)
```

#### `AppError`
```go
type AppError struct {
    Type    ErrorType
    Message string
    Cause   error
    Context map[string]interface{}
}
```

### Constructor Functions

- `NewFatal(msg string, cause error) *AppError`
- `NewRetryable(msg string, cause error) *AppError`
- `NewValidation(msg string, cause error) *AppError`
- `NewSkippable(msg string, cause error) *AppError`

### Methods

- `(e *AppError) Error() string` - Returns formatted error message
- `(e *AppError) Unwrap() error` - Returns the underlying cause
- `(e *AppError) WithContext(key string, value interface{}) *AppError` - Adds context
- `(e *AppError) Is(target error) bool` - Supports `errors.Is()` comparison

### Helper Functions

- `IsFatal(err error) bool` - Checks if error is fatal
- `IsRetryable(err error) bool` - Checks if error is retryable
- `IsValidation(err error) bool` - Checks if error is validation
- `IsSkippable(err error) bool` - Checks if error is skippable

## Testing

The package has 100% test coverage. Run tests with:

```bash
go test ./internal/errors/
```

For coverage report:

```bash
go test -coverprofile=coverage.out ./internal/errors/
go tool cover -html=coverage.out
```

## Design Rationale

This error package follows ADR-005 (Error Handling Strategy) and provides:

1. **Type Safety**: Compile-time type checking for error handling
2. **Context Preservation**: Rich context information for debugging
3. **Standard Compatibility**: Full support for Go's error handling patterns
4. **Logging Integration**: Seamless integration with the logger package
5. **Decision Support**: Error types guide handling strategy (retry, skip, fail)

## Examples

See `example_test.go` for comprehensive usage examples.
