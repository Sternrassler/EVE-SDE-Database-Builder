package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// TestErrorTypes tests creation of all error types
func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name        string
		constructor func(string, error) *AppError
		errorType   ErrorType
		typeString  string
	}{
		{
			name:        "Fatal Error",
			constructor: NewFatal,
			errorType:   ErrorTypeFatal,
			typeString:  "Fatal",
		},
		{
			name:        "Retryable Error",
			constructor: NewRetryable,
			errorType:   ErrorTypeRetryable,
			typeString:  "Retryable",
		},
		{
			name:        "Validation Error",
			constructor: NewValidation,
			errorType:   ErrorTypeValidation,
			typeString:  "Validation",
		},
		{
			name:        "Skippable Error",
			constructor: NewSkippable,
			errorType:   ErrorTypeSkippable,
			typeString:  "Skippable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cause := errors.New("underlying error")
			err := tt.constructor("test error", cause)

			if err == nil {
				t.Fatal("constructor returned nil")
			}

			if err.Type != tt.errorType {
				t.Errorf("expected type %v, got %v", tt.errorType, err.Type)
			}

			if err.Message != "test error" {
				t.Errorf("expected message 'test error', got %s", err.Message)
			}

			if err.Cause != cause {
				t.Errorf("expected cause to be %v, got %v", cause, err.Cause)
			}

			if err.Context == nil {
				t.Error("expected context to be initialized")
			}

			if err.Type.String() != tt.typeString {
				t.Errorf("expected type string %s, got %s", tt.typeString, err.Type.String())
			}
		})
	}
}

// TestErrorWithNilCause tests error creation with nil cause
func TestErrorWithNilCause(t *testing.T) {
	err := NewFatal("test error", nil)

	if err == nil {
		t.Fatal("constructor returned nil")
	}

	if err.Cause != nil {
		t.Errorf("expected nil cause, got %v", err.Cause)
	}

	// Error message should not include cause
	if strings.Contains(err.Error(), ": ") {
		t.Errorf("error message should not contain cause separator when cause is nil, got: %s", err.Error())
	}
}

// TestErrorMessageFormat tests the error message formatting
func TestErrorMessageFormat(t *testing.T) {
	tests := []struct {
		name            string
		errorFunc       func() *AppError
		expectedContain []string
	}{
		{
			name: "Error with cause",
			errorFunc: func() *AppError {
				return NewFatal("database connection failed", errors.New("timeout"))
			},
			expectedContain: []string{"[Fatal]", "database connection failed", "timeout"},
		},
		{
			name: "Error without cause",
			errorFunc: func() *AppError {
				return NewValidation("invalid input", nil)
			},
			expectedContain: []string{"[Validation]", "invalid input"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errorFunc()
			errMsg := err.Error()

			for _, expected := range tt.expectedContain {
				if !strings.Contains(errMsg, expected) {
					t.Errorf("error message '%s' should contain '%s'", errMsg, expected)
				}
			}
		})
	}
}

// TestErrorUnwrap tests error unwrapping
func TestErrorUnwrap(t *testing.T) {
	cause := errors.New("root cause")
	err := NewRetryable("temporary failure", cause)

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Errorf("expected unwrapped error to be %v, got %v", cause, unwrapped)
	}

	// Test with nil cause
	errNoCause := NewFatal("critical error", nil)
	unwrappedNil := errNoCause.Unwrap()
	if unwrappedNil != nil {
		t.Errorf("expected nil unwrapped error, got %v", unwrappedNil)
	}
}

// TestErrorIs tests error comparison using errors.Is
func TestErrorIs(t *testing.T) {
	err1 := NewFatal("database error", nil)
	err2 := NewFatal("database error", nil)
	err3 := NewFatal("network error", nil)
	err4 := NewRetryable("database error", nil)

	// Same type and message should match
	if !errors.Is(err1, err2) {
		t.Error("errors with same type and message should be equal")
	}

	// Different message should not match
	if errors.Is(err1, err3) {
		t.Error("errors with different messages should not be equal")
	}

	// Different type should not match
	if errors.Is(err1, err4) {
		t.Error("errors with different types should not be equal")
	}

	// Non-AppError should not match
	stdErr := errors.New("standard error")
	if errors.Is(err1, stdErr) {
		t.Error("AppError should not match standard error")
	}
}

// TestErrorAs tests error type assertion using errors.As
func TestErrorAs(t *testing.T) {
	cause := errors.New("underlying")
	err := NewRetryable("test error", cause)

	// Wrap it in another error
	wrapped := fmt.Errorf("wrapped: %w", err)

	var appErr *AppError
	if !errors.As(wrapped, &appErr) {
		t.Fatal("errors.As should find AppError in wrapped error")
	}

	if appErr.Type != ErrorTypeRetryable {
		t.Errorf("expected type %v, got %v", ErrorTypeRetryable, appErr.Type)
	}

	if appErr.Message != "test error" {
		t.Errorf("expected message 'test error', got %s", appErr.Message)
	}
}

// TestWithContext tests adding context to errors
func TestWithContext(t *testing.T) {
	err := NewValidation("invalid data", nil)

	// Add single context
	err = err.WithContext("field", "username")
	if err.Context["field"] != "username" {
		t.Errorf("expected context field='username', got %v", err.Context["field"])
	}

	// Add multiple contexts
	err = err.WithContext("value", "john@doe").WithContext("rule", "email_format")

	if err.Context["value"] != "john@doe" {
		t.Errorf("expected context value='john@doe', got %v", err.Context["value"])
	}

	if err.Context["rule"] != "email_format" {
		t.Errorf("expected context rule='email_format', got %v", err.Context["rule"])
	}

	// Verify all contexts are present
	if len(err.Context) != 3 {
		t.Errorf("expected 3 context entries, got %d", len(err.Context))
	}
}

// TestWithContextOverwrite tests that context values can be overwritten
func TestWithContextOverwrite(t *testing.T) {
	err := NewFatal("error", nil)
	err = err.WithContext("key", "value1")
	err = err.WithContext("key", "value2")

	if err.Context["key"] != "value2" {
		t.Errorf("expected context key='value2', got %v", err.Context["key"])
	}
}

// TestErrorTypeHelpers tests the helper functions for error type checking
func TestErrorTypeHelpers(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		isFatal bool
		isRetry bool
		isValid bool
		isSkip  bool
	}{
		{
			name:    "Fatal Error",
			err:     NewFatal("fatal", nil),
			isFatal: true,
			isRetry: false,
			isValid: false,
			isSkip:  false,
		},
		{
			name:    "Retryable Error",
			err:     NewRetryable("retry", nil),
			isFatal: false,
			isRetry: true,
			isValid: false,
			isSkip:  false,
		},
		{
			name:    "Validation Error",
			err:     NewValidation("validation", nil),
			isFatal: false,
			isRetry: false,
			isValid: true,
			isSkip:  false,
		},
		{
			name:    "Skippable Error",
			err:     NewSkippable("skip", nil),
			isFatal: false,
			isRetry: false,
			isValid: false,
			isSkip:  true,
		},
		{
			name:    "Standard Error",
			err:     errors.New("standard"),
			isFatal: false,
			isRetry: false,
			isValid: false,
			isSkip:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if IsFatal(tt.err) != tt.isFatal {
				t.Errorf("IsFatal expected %v, got %v", tt.isFatal, IsFatal(tt.err))
			}
			if IsRetryable(tt.err) != tt.isRetry {
				t.Errorf("IsRetryable expected %v, got %v", tt.isRetry, IsRetryable(tt.err))
			}
			if IsValidation(tt.err) != tt.isValid {
				t.Errorf("IsValidation expected %v, got %v", tt.isValid, IsValidation(tt.err))
			}
			if IsSkippable(tt.err) != tt.isSkip {
				t.Errorf("IsSkippable expected %v, got %v", tt.isSkip, IsSkippable(tt.err))
			}
		})
	}
}

// TestErrorTypeHelpersWithWrappedErrors tests helper functions with wrapped errors
func TestErrorTypeHelpersWithWrappedErrors(t *testing.T) {
	appErr := NewFatal("fatal error", nil)
	wrapped := fmt.Errorf("wrapper: %w", appErr)

	if !IsFatal(wrapped) {
		t.Error("IsFatal should work with wrapped errors")
	}

	if IsRetryable(wrapped) {
		t.Error("IsRetryable should return false for wrapped fatal error")
	}
}

// TestErrorTypeString tests the String method of ErrorType
func TestErrorTypeString(t *testing.T) {
	tests := []struct {
		errorType ErrorType
		expected  string
	}{
		{ErrorTypeFatal, "Fatal"},
		{ErrorTypeRetryable, "Retryable"},
		{ErrorTypeValidation, "Validation"},
		{ErrorTypeSkippable, "Skippable"},
		{ErrorType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.errorType.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.errorType.String())
			}
		})
	}
}

// TestContextInitialization tests that context is properly initialized
func TestContextInitialization(t *testing.T) {
	err := NewFatal("test", nil)

	// Context should be initialized but empty
	if err.Context == nil {
		t.Fatal("context should not be nil")
	}

	if len(err.Context) != 0 {
		t.Errorf("expected empty context, got %d entries", len(err.Context))
	}
}

// TestErrorChaining tests complex error wrapping scenarios
func TestErrorChaining(t *testing.T) {
	// Create a chain of errors
	rootErr := errors.New("root cause")
	appErr := NewRetryable("service unavailable", rootErr)
	wrapperErr := fmt.Errorf("failed to process: %w", appErr)

	// Test unwrapping through the chain
	var unwrapped *AppError
	if !errors.As(wrapperErr, &unwrapped) {
		t.Fatal("should be able to extract AppError from chain")
	}

	if unwrapped.Type != ErrorTypeRetryable {
		t.Errorf("expected ErrorTypeRetryable, got %v", unwrapped.Type)
	}

	// Test that we can reach the root cause
	if !errors.Is(wrapperErr, rootErr) {
		t.Error("should be able to find root cause in error chain")
	}
}

// TestNilErrorHandling tests handling of nil errors
func TestNilErrorHandling(t *testing.T) {
	// Helper functions should handle nil gracefully
	if IsFatal(nil) {
		t.Error("IsFatal(nil) should return false")
	}
	if IsRetryable(nil) {
		t.Error("IsRetryable(nil) should return false")
	}
	if IsValidation(nil) {
		t.Error("IsValidation(nil) should return false")
	}
	if IsSkippable(nil) {
		t.Error("IsSkippable(nil) should return false")
	}
}

// TestWithContextOnNilContext tests adding context when Context map is nil
func TestWithContextOnNilContext(t *testing.T) {
	// Create AppError without using constructor (to test nil context initialization)
	err := &AppError{
		Type:    ErrorTypeFatal,
		Message: "test",
		Cause:   nil,
		Context: nil, // Explicitly set to nil
	}

	// Adding context should initialize the map
	err = err.WithContext("key", "value")

	if err.Context == nil {
		t.Fatal("WithContext should initialize nil context map")
	}

	if err.Context["key"] != "value" {
		t.Errorf("expected context key='value', got %v", err.Context["key"])
	}
}
