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

```go
err := errors.NewValidation("invalid input", nil)
err = err.WithContext("field", "email")
err = err.WithContext("value", "invalid-email")
err = err.WithContext("rule", "rfc5322")

// Access context
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

```go
import (
    "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"
    "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/logger"
)

log := logger.GetGlobalLogger()

err := errors.NewFatal("critical failure", cause)
err = err.WithContext("component", "database")

// Log with error details
log.Error("Operation failed",
    logger.Field{Key: "error", Value: err.Error()},
    logger.Field{Key: "type", Value: err.Type.String()},
    logger.Field{Key: "context", Value: err.Context},
)
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
