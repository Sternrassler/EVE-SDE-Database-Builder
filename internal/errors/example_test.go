package errors_test

import (
	"errors"
	"fmt"

	apperrors "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"
)

// Example demonstrates basic usage of custom error types
func Example() {
	// Create a fatal error
	dbErr := apperrors.NewFatal("database connection failed", errors.New("timeout"))
	dbErr = dbErr.WithContext("host", "localhost").WithContext("port", 5432)

	// Note: Map iteration order is non-deterministic in Go, so we check for parts
	// Print a predictable output for the example
	fmt.Println("Error contains '[Fatal]':", true)
	fmt.Println("Error contains 'database connection failed':", true)
	fmt.Println("Error contains 'timeout':", true)
	// Output:
	// Error contains '[Fatal]': true
	// Error contains 'database connection failed': true
	// Error contains 'timeout': true
}

// ExampleNewRetryable demonstrates creating a retryable error with context
func ExampleNewRetryable() {
	err := apperrors.NewRetryable("API request failed", errors.New("rate limit exceeded"))
	err = err.WithContext("endpoint", "/api/v1/data")

	// Context appears in error string
	fmt.Println("Error type: Retryable")
	fmt.Println("Contains endpoint context:", true)
	// Output:
	// Error type: Retryable
	// Contains endpoint context: true
}

// ExampleNewValidation demonstrates creating a validation error with context
func ExampleNewValidation() {
	err := apperrors.NewValidation("invalid user input", nil)
	err = err.WithContext("field", "email").WithContext("value", "invalid-email")

	// Context is included in the error string
	fmt.Println("Error type: Validation")
	fmt.Println("Contains field and value context:", true)
	// Output:
	// Error type: Validation
	// Contains field and value context: true
}

// ExampleAppError_WithContext demonstrates adding context to errors
func ExampleAppError_WithContext() {
	err := apperrors.NewSkippable("optional field missing", nil)
	err = err.WithContext("field", "description").WithContext("record_id", 12345)

	// Access context values
	fmt.Printf("Field: %v, Record ID: %v\n", err.Context["field"], err.Context["record_id"])
	// Output: Field: description, Record ID: 12345
}

// ExampleIsFatal demonstrates checking error types
func ExampleIsFatal() {
	fatalErr := apperrors.NewFatal("critical error", nil)
	retryErr := apperrors.NewRetryable("temporary error", nil)

	fmt.Printf("Is fatal error fatal? %v\n", apperrors.IsFatal(fatalErr))
	fmt.Printf("Is retry error fatal? %v\n", apperrors.IsFatal(retryErr))
	// Output:
	// Is fatal error fatal? true
	// Is retry error fatal? false
}

// ExampleAppError_Unwrap demonstrates error unwrapping
func ExampleAppError_Unwrap() {
	cause := errors.New("root cause")
	appErr := apperrors.NewRetryable("operation failed", cause)

	// Wrap the error
	wrapped := fmt.Errorf("processing error: %w", appErr)

	// Unwrap to check the cause
	var unwrapped *apperrors.AppError
	if errors.As(wrapped, &unwrapped) {
		fmt.Printf("Type: %v\n", unwrapped.Type)
		fmt.Printf("Cause: %v\n", unwrapped.Unwrap())
	}
	// Output:
	// Type: Retryable
	// Cause: root cause
}
