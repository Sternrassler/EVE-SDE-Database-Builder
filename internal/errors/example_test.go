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

	fmt.Println(dbErr.Error())
	// Output: [Fatal] database connection failed: timeout
}

// ExampleNewRetryable demonstrates creating a retryable error
func ExampleNewRetryable() {
	err := apperrors.NewRetryable("API request failed", errors.New("rate limit exceeded"))
	err = err.WithContext("endpoint", "/api/v1/data")

	fmt.Println(err.Error())
	// Output: [Retryable] API request failed: rate limit exceeded
}

// ExampleNewValidation demonstrates creating a validation error
func ExampleNewValidation() {
	err := apperrors.NewValidation("invalid user input", nil)
	err = err.WithContext("field", "email").WithContext("value", "invalid-email")

	fmt.Println(err.Error())
	// Output: [Validation] invalid user input
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
