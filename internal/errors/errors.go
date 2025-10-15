// Package errors provides custom error types with classification
// for precise error handling in the EVE SDE Database Builder.
package errors

import (
	"errors"
	"fmt"
)

// ErrorType represents the classification of an error
type ErrorType int

const (
	// ErrorTypeFatal represents critical errors that cannot be recovered
	ErrorTypeFatal ErrorType = iota
	// ErrorTypeRetryable represents errors that may succeed on retry
	ErrorTypeRetryable
	// ErrorTypeValidation represents input validation errors
	ErrorTypeValidation
	// ErrorTypeSkippable represents errors that can be safely skipped
	ErrorTypeSkippable
)

// String returns a string representation of the ErrorType
func (e ErrorType) String() string {
	switch e {
	case ErrorTypeFatal:
		return "Fatal"
	case ErrorTypeRetryable:
		return "Retryable"
	case ErrorTypeValidation:
		return "Validation"
	case ErrorTypeSkippable:
		return "Skippable"
	default:
		return "Unknown"
	}
}

// AppError represents a structured application error with type classification
type AppError struct {
	Type    ErrorType
	Message string
	Cause   error
	Context map[string]interface{}
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type.String(), e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Type.String(), e.Message)
}

// Unwrap returns the underlying cause error for error chain support
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithContext adds context information to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// Is implements error comparison for errors.Is support
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Type == t.Type && e.Message == t.Message
}

// NewFatal creates a new fatal error
func NewFatal(msg string, cause error) *AppError {
	return &AppError{
		Type:    ErrorTypeFatal,
		Message: msg,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// NewRetryable creates a new retryable error
func NewRetryable(msg string, cause error) *AppError {
	return &AppError{
		Type:    ErrorTypeRetryable,
		Message: msg,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// NewValidation creates a new validation error
func NewValidation(msg string, cause error) *AppError {
	return &AppError{
		Type:    ErrorTypeValidation,
		Message: msg,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// NewSkippable creates a new skippable error
func NewSkippable(msg string, cause error) *AppError {
	return &AppError{
		Type:    ErrorTypeSkippable,
		Message: msg,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// IsFatal checks if the error is a fatal error
func IsFatal(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeFatal
	}
	return false
}

// IsRetryable checks if the error is a retryable error
func IsRetryable(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeRetryable
	}
	return false
}

// IsValidation checks if the error is a validation error
func IsValidation(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeValidation
	}
	return false
}

// IsSkippable checks if the error is a skippable error
func IsSkippable(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeSkippable
	}
	return false
}
