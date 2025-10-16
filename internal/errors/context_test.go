package errors

import (
	"errors"
	"strings"
	"testing"
)

// TestWithContext_Single tests adding a single context entry
func TestWithContext_Single(t *testing.T) {
	err := NewRetryable("database error", nil)
	err = err.WithContext("table", "invTypes")

	if err.Context["table"] != "invTypes" {
		t.Errorf("expected context table='invTypes', got %v", err.Context["table"])
	}

	if len(err.Context) != 1 {
		t.Errorf("expected 1 context entry, got %d", len(err.Context))
	}
}

// TestWithContext_Multiple tests adding multiple context entries
func TestWithContext_Multiple(t *testing.T) {
	err := NewRetryable("database error", nil)
	err = err.WithContext("table", "invTypes").
		WithContext("operation", "batch_insert").
		WithContext("retry_attempt", 3)

	if err.Context["table"] != "invTypes" {
		t.Errorf("expected context table='invTypes', got %v", err.Context["table"])
	}

	if err.Context["operation"] != "batch_insert" {
		t.Errorf("expected context operation='batch_insert', got %v", err.Context["operation"])
	}

	if err.Context["retry_attempt"] != 3 {
		t.Errorf("expected context retry_attempt=3, got %v", err.Context["retry_attempt"])
	}

	if len(err.Context) != 3 {
		t.Errorf("expected 3 context entries, got %d", len(err.Context))
	}
}

// TestWithContext_InErrorString tests that context appears in error string
func TestWithContext_InErrorString(t *testing.T) {
	cause := errors.New("connection timeout")
	err := NewRetryable("DB locked", cause).
		WithContext("table", "invTypes").
		WithContext("operation", "batch_insert").
		WithContext("retry_attempt", 3)

	errStr := err.Error()

	// Check that all context keys and values appear in error string
	expectedParts := []string{
		"[Retryable]",
		"DB locked",
		"table=invTypes",
		"operation=batch_insert",
		"retry_attempt=3",
		"connection timeout",
	}

	for _, part := range expectedParts {
		if !strings.Contains(errStr, part) {
			t.Errorf("error string '%s' should contain '%s'", errStr, part)
		}
	}
}

// TestWithContext_NoContext tests error string without context
func TestWithContext_NoContext(t *testing.T) {
	err := NewFatal("critical error", nil)
	errStr := err.Error()

	// Should not contain parentheses when no context
	if strings.Contains(errStr, "(") || strings.Contains(errStr, ")") {
		t.Errorf("error string without context should not contain parentheses, got: %s", errStr)
	}

	expected := "[Fatal] critical error"
	if errStr != expected {
		t.Errorf("expected '%s', got '%s'", expected, errStr)
	}
}

// TestWithContext_Chaining tests that chaining returns the same error instance
func TestWithContext_Chaining(t *testing.T) {
	err1 := NewValidation("invalid input", nil)
	err2 := err1.WithContext("field", "username")
	err3 := err2.WithContext("value", "john@doe")

	// All should point to the same error instance
	if err1 != err2 || err2 != err3 {
		t.Error("WithContext should return the same error instance for chaining")
	}

	if len(err1.Context) != 2 {
		t.Errorf("expected 2 context entries in original error, got %d", len(err1.Context))
	}
}

// TestWithContext_PreservedOnUnwrap tests that context is preserved when unwrapping
func TestWithContext_PreservedOnUnwrap(t *testing.T) {
	cause := errors.New("root cause")
	err := NewRetryable("operation failed", cause).
		WithContext("attempt", 1).
		WithContext("resource", "database")

	// Unwrap should give us the cause
	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Errorf("expected unwrapped error to be %v, got %v", cause, unwrapped)
	}

	// But context should still be accessible on the original error
	if err.Context["attempt"] != 1 {
		t.Errorf("context should be preserved, expected attempt=1, got %v", err.Context["attempt"])
	}

	if err.Context["resource"] != "database" {
		t.Errorf("context should be preserved, expected resource='database', got %v", err.Context["resource"])
	}
}

// TestWithContext_DifferentTypes tests context with different value types
func TestWithContext_DifferentTypes(t *testing.T) {
	err := NewSkippable("skipped record", nil).
		WithContext("string", "text").
		WithContext("int", 42).
		WithContext("bool", true).
		WithContext("float", 3.14).
		WithContext("slice", []string{"a", "b", "c"}).
		WithContext("map", map[string]int{"x": 1, "y": 2})

	// Verify all types are stored correctly
	if err.Context["string"] != "text" {
		t.Errorf("expected string='text', got %v", err.Context["string"])
	}
	if err.Context["int"] != 42 {
		t.Errorf("expected int=42, got %v", err.Context["int"])
	}
	if err.Context["bool"] != true {
		t.Errorf("expected bool=true, got %v", err.Context["bool"])
	}
	if err.Context["float"] != 3.14 {
		t.Errorf("expected float=3.14, got %v", err.Context["float"])
	}

	// Error string should contain all values (format may vary for complex types)
	errStr := err.Error()
	if !strings.Contains(errStr, "string=text") {
		t.Errorf("error string should contain string context, got: %s", errStr)
	}
	if !strings.Contains(errStr, "int=42") {
		t.Errorf("error string should contain int context, got: %s", errStr)
	}
	if !strings.Contains(errStr, "bool=true") {
		t.Errorf("error string should contain bool context, got: %s", errStr)
	}
}

// TestWithContext_OrderIndependent tests that context order doesn't affect equality
func TestWithContext_OrderIndependent(t *testing.T) {
	err1 := NewFatal("error", nil).
		WithContext("key1", "value1").
		WithContext("key2", "value2")

	err2 := NewFatal("error", nil).
		WithContext("key2", "value2").
		WithContext("key1", "value1")

	// Both should have the same context entries
	if len(err1.Context) != 2 || len(err2.Context) != 2 {
		t.Error("both errors should have 2 context entries")
	}

	if err1.Context["key1"] != "value1" || err2.Context["key1"] != "value1" {
		t.Error("key1 should be 'value1' in both errors")
	}

	if err1.Context["key2"] != "value2" || err2.Context["key2"] != "value2" {
		t.Error("key2 should be 'value2' in both errors")
	}

	// Error strings may differ in order (map iteration is non-deterministic)
	// but should contain the same parts
	errStr1 := err1.Error()
	errStr2 := err2.Error()

	if !strings.Contains(errStr1, "key1=value1") || !strings.Contains(errStr2, "key1=value1") {
		t.Error("both error strings should contain key1=value1")
	}

	if !strings.Contains(errStr1, "key2=value2") || !strings.Contains(errStr2, "key2=value2") {
		t.Error("both error strings should contain key2=value2")
	}
}

// TestWithContext_EmptyKey tests adding context with empty key
func TestWithContext_EmptyKey(t *testing.T) {
	err := NewValidation("validation error", nil)
	err = err.WithContext("", "empty key")

	// Empty key should still be added
	if err.Context[""] != "empty key" {
		t.Errorf("expected context with empty key, got %v", err.Context[""])
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "=empty key") {
		t.Errorf("error string should contain empty key context, got: %s", errStr)
	}
}

// TestWithContext_NilValue tests adding context with nil value
func TestWithContext_NilValue(t *testing.T) {
	err := NewRetryable("retry error", nil)
	err = err.WithContext("nullable", nil)

	// Nil value should be stored
	if err.Context["nullable"] != nil {
		t.Errorf("expected nil value, got %v", err.Context["nullable"])
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "nullable=") {
		t.Errorf("error string should contain nullable key, got: %s", errStr)
	}
}

// TestWithContext_Overwrite tests that context values can be overwritten
func TestWithContext_Overwrite(t *testing.T) {
	err := NewFatal("error", nil)
	err = err.WithContext("key", "value1")
	err = err.WithContext("key", "value2")

	if err.Context["key"] != "value2" {
		t.Errorf("expected context key='value2' after overwrite, got %v", err.Context["key"])
	}

	if len(err.Context) != 1 {
		t.Errorf("expected 1 context entry after overwrite, got %d", len(err.Context))
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "key=value2") {
		t.Errorf("error string should contain overwritten value, got: %s", errStr)
	}

	if strings.Contains(errStr, "value1") {
		t.Errorf("error string should not contain old value, got: %s", errStr)
	}
}

// TestWithContext_ComplexExample tests the example from the issue
func TestWithContext_ComplexExample(t *testing.T) {
	cause := errors.New("database is locked")
	err := NewRetryable("DB locked", cause).
		WithContext("table", "invTypes").
		WithContext("operation", "batch_insert").
		WithContext("retry_attempt", 3)

	errStr := err.Error()

	// Verify the format matches the expected output from the issue
	// Expected: "DB locked (table=invTypes, operation=batch_insert, retry_attempt=3): original cause"
	// Note: Map iteration order is non-deterministic, so we check for parts

	if !strings.HasPrefix(errStr, "[Retryable] DB locked (") {
		t.Errorf("error should start with '[Retryable] DB locked (', got: %s", errStr)
	}

	if !strings.Contains(errStr, "table=invTypes") {
		t.Errorf("error should contain 'table=invTypes', got: %s", errStr)
	}

	if !strings.Contains(errStr, "operation=batch_insert") {
		t.Errorf("error should contain 'operation=batch_insert', got: %s", errStr)
	}

	if !strings.Contains(errStr, "retry_attempt=3") {
		t.Errorf("error should contain 'retry_attempt=3', got: %s", errStr)
	}

	if !strings.HasSuffix(errStr, "): database is locked") {
		t.Errorf("error should end with '): database is locked', got: %s", errStr)
	}
}

// TestWithContext_WithWrappedErrors tests context with wrapped errors using errors.Is/As
func TestWithContext_WithWrappedErrors(t *testing.T) {
	cause := errors.New("network timeout")
	appErr := NewRetryable("service unavailable", cause).
		WithContext("service", "auth").
		WithContext("endpoint", "/api/login")

	// Wrap with standard error
	wrapped := errors.Join(appErr, errors.New("additional context"))

	// Extract AppError from wrapped error
	var extractedErr *AppError
	if !errors.As(wrapped, &extractedErr) {
		t.Fatal("should be able to extract AppError from wrapped error")
	}

	// Context should be preserved
	if extractedErr.Context["service"] != "auth" {
		t.Errorf("expected service='auth', got %v", extractedErr.Context["service"])
	}

	if extractedErr.Context["endpoint"] != "/api/login" {
		t.Errorf("expected endpoint='/api/login', got %v", extractedErr.Context["endpoint"])
	}
}

// TestWithContext_OnAllErrorTypes tests context on all error types
func TestWithContext_OnAllErrorTypes(t *testing.T) {
	tests := []struct {
		name        string
		constructor func(string, error) *AppError
		errorType   ErrorType
	}{
		{"Fatal", NewFatal, ErrorTypeFatal},
		{"Retryable", NewRetryable, ErrorTypeRetryable},
		{"Validation", NewValidation, ErrorTypeValidation},
		{"Skippable", NewSkippable, ErrorTypeSkippable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor("test error", nil).
				WithContext("type_test", tt.name).
				WithContext("numeric", 123)

			if err.Type != tt.errorType {
				t.Errorf("expected type %v, got %v", tt.errorType, err.Type)
			}

			if err.Context["type_test"] != tt.name {
				t.Errorf("expected context type_test='%s', got %v", tt.name, err.Context["type_test"])
			}

			errStr := err.Error()
			if !strings.Contains(errStr, "type_test="+tt.name) {
				t.Errorf("error string should contain context, got: %s", errStr)
			}
		})
	}
}
