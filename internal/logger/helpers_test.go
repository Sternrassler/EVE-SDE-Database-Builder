package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	apperrors "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"
	"github.com/rs/zerolog"
)

func TestLogHTTPRequest(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf).With().Timestamp().Logger()
	logger := &Logger{logger: zl}

	logger.LogHTTPRequest("GET", "/api/types", 200, 45*time.Millisecond)

	output := buf.String()

	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v, output: %s", err, output)
	}

	// Verify message
	if logEntry["message"] != "HTTP request" {
		t.Errorf("Expected message='HTTP request', got: %v", logEntry["message"])
	}

	// Verify all required fields
	if logEntry["method"] != "GET" {
		t.Errorf("Expected method='GET', got: %v", logEntry["method"])
	}
	if logEntry["url"] != "/api/types" {
		t.Errorf("Expected url='/api/types', got: %v", logEntry["url"])
	}
	if logEntry["status_code"] != float64(200) {
		t.Errorf("Expected status_code=200, got: %v", logEntry["status_code"])
	}
	if logEntry["duration_ms"] != float64(45) {
		t.Errorf("Expected duration_ms=45, got: %v", logEntry["duration_ms"])
	}
}

func TestLogHTTPRequestMultipleStatuses(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		url        string
		statusCode int
		duration   time.Duration
	}{
		{"Success", "GET", "/api/users", 200, 10 * time.Millisecond},
		{"Not Found", "GET", "/api/missing", 404, 5 * time.Millisecond},
		{"Server Error", "POST", "/api/data", 500, 100 * time.Millisecond},
		{"Created", "POST", "/api/items", 201, 50 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			zl := zerolog.New(&buf).With().Timestamp().Logger()
			logger := &Logger{logger: zl}

			logger.LogHTTPRequest(tt.method, tt.url, tt.statusCode, tt.duration)

			output := buf.String()
			var logEntry map[string]interface{}
			if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
				t.Fatalf("Failed to parse JSON output: %v", err)
			}

			if logEntry["method"] != tt.method {
				t.Errorf("Expected method=%s, got: %v", tt.method, logEntry["method"])
			}
			if logEntry["status_code"] != float64(tt.statusCode) {
				t.Errorf("Expected status_code=%d, got: %v", tt.statusCode, logEntry["status_code"])
			}
		})
	}
}

func TestLogDBQuery(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf).With().Timestamp().Logger()
	logger := &Logger{logger: zl}

	logger.LogDBQuery("SELECT * FROM types WHERE id = ?", []interface{}{123}, 2*time.Millisecond)

	output := buf.String()

	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v, output: %s", err, output)
	}

	// Verify message
	if logEntry["message"] != "Database query" {
		t.Errorf("Expected message='Database query', got: %v", logEntry["message"])
	}

	// Verify query field
	if logEntry["query"] != "SELECT * FROM types WHERE id = ?" {
		t.Errorf("Expected query to match, got: %v", logEntry["query"])
	}

	// Verify args field (as array)
	args, ok := logEntry["args"].([]interface{})
	if !ok {
		t.Fatalf("Expected args to be array, got: %T", logEntry["args"])
	}
	if len(args) != 1 || args[0] != float64(123) {
		t.Errorf("Expected args=[123], got: %v", args)
	}

	// Verify duration
	if logEntry["duration_ms"] != float64(2) {
		t.Errorf("Expected duration_ms=2, got: %v", logEntry["duration_ms"])
	}
}

func TestLogDBQueryWithMultipleArgs(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf).With().Timestamp().Logger()
	logger := &Logger{logger: zl}

	args := []interface{}{"test", 42, true}
	logger.LogDBQuery("INSERT INTO items (name, count, active) VALUES (?, ?, ?)", args, 10*time.Millisecond)

	output := buf.String()
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify args
	loggedArgs, ok := logEntry["args"].([]interface{})
	if !ok {
		t.Fatalf("Expected args to be array, got: %T", logEntry["args"])
	}
	if len(loggedArgs) != 3 {
		t.Errorf("Expected 3 args, got: %d", len(loggedArgs))
	}
}

func TestLogError(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf).With().Timestamp().Logger()
	logger := &Logger{logger: zl}

	err := errors.New("connection timeout")
	context := map[string]interface{}{
		"file": "types.jsonl",
		"line": 42,
	}

	logger.LogError(err, context)

	output := buf.String()

	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v, output: %s", err, output)
	}

	// Verify level is error
	if logEntry["level"] != "error" {
		t.Errorf("Expected level='error', got: %v", logEntry["level"])
	}

	// Verify message
	if logEntry["message"] != "Error occurred" {
		t.Errorf("Expected message='Error occurred', got: %v", logEntry["message"])
	}

	// Verify error field
	if logEntry["error"] != "connection timeout" {
		t.Errorf("Expected error='connection timeout', got: %v", logEntry["error"])
	}

	// Verify context fields
	if logEntry["file"] != "types.jsonl" {
		t.Errorf("Expected file='types.jsonl', got: %v", logEntry["file"])
	}
	if logEntry["line"] != float64(42) {
		t.Errorf("Expected line=42, got: %v", logEntry["line"])
	}
}

func TestLogErrorEmptyContext(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf).With().Timestamp().Logger()
	logger := &Logger{logger: zl}

	err := errors.New("test error")
	logger.LogError(err, map[string]interface{}{})

	output := buf.String()
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Should still have error field
	if logEntry["error"] != "test error" {
		t.Errorf("Expected error='test error', got: %v", logEntry["error"])
	}
}

func TestLogAppStart(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf).With().Timestamp().Logger()
	logger := &Logger{logger: zl}

	logger.LogAppStart("0.1.0", "abc123def")

	output := buf.String()

	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v, output: %s", err, output)
	}

	// Verify message
	if logEntry["message"] != "Application started" {
		t.Errorf("Expected message='Application started', got: %v", logEntry["message"])
	}

	// Verify version field
	if logEntry["version"] != "0.1.0" {
		t.Errorf("Expected version='0.1.0', got: %v", logEntry["version"])
	}

	// Verify commit field
	if logEntry["commit"] != "abc123def" {
		t.Errorf("Expected commit='abc123def', got: %v", logEntry["commit"])
	}
}

func TestLogAppShutdown(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf).With().Timestamp().Logger()
	logger := &Logger{logger: zl}

	logger.LogAppShutdown("graceful shutdown")

	output := buf.String()

	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v, output: %s", err, output)
	}

	// Verify message
	if logEntry["message"] != "Application shutting down" {
		t.Errorf("Expected message='Application shutting down', got: %v", logEntry["message"])
	}

	// Verify reason field
	if logEntry["reason"] != "graceful shutdown" {
		t.Errorf("Expected reason='graceful shutdown', got: %v", logEntry["reason"])
	}
}

func TestLogAppShutdownSignalInterrupt(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf).With().Timestamp().Logger()
	logger := &Logger{logger: zl}

	logger.LogAppShutdown("received SIGINT")

	output := buf.String()
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if logEntry["reason"] != "received SIGINT" {
		t.Errorf("Expected reason='received SIGINT', got: %v", logEntry["reason"])
	}
}

// Test global helper functions
func TestGlobalHelperFunctions(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf).With().Timestamp().Logger()
	logger := &Logger{logger: zl}
	SetGlobalLogger(logger)

	// Test LogHTTPRequest global
	LogHTTPRequest("POST", "/api/test", 201, 30*time.Millisecond)
	output := buf.String()
	if !strings.Contains(output, "HTTP request") {
		t.Error("Global LogHTTPRequest did not produce expected output")
	}
	buf.Reset()

	// Test LogDBQuery global
	LogDBQuery("SELECT 1", []interface{}{}, 1*time.Millisecond)
	output = buf.String()
	if !strings.Contains(output, "Database query") {
		t.Error("Global LogDBQuery did not produce expected output")
	}
	buf.Reset()

	// Test LogErrorGlobal
	LogErrorGlobal(errors.New("test"), map[string]interface{}{"key": "value"})
	output = buf.String()
	if !strings.Contains(output, "Error occurred") {
		t.Error("Global LogErrorGlobal did not produce expected output")
	}
	buf.Reset()

	// Test LogAppStart global
	LogAppStart("1.0.0", "commit123")
	output = buf.String()
	if !strings.Contains(output, "Application started") {
		t.Error("Global LogAppStart did not produce expected output")
	}
	buf.Reset()

	// Test LogAppShutdown global
	LogAppShutdown("normal exit")
	output = buf.String()
	if !strings.Contains(output, "Application shutting down") {
		t.Error("Global LogAppShutdown did not produce expected output")
	}

	// Clean up
	mu.Lock()
	globalLogger = nil
	mu.Unlock()
}

// TestLogAppError tests logging AppError with automatic context extraction
func TestLogAppError(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf).With().Timestamp().Logger()
	logger := &Logger{logger: zl}

	// Create an AppError with context
	cause := errors.New("database is locked")
	appErr := apperrors.NewRetryable("DB locked", cause).
		WithContext("table", "invTypes").
		WithContext("operation", "batch_insert").
		WithContext("retry_attempt", 3)

	logger.LogAppError(appErr)

	output := buf.String()

	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v, output: %s", err, output)
	}

	// Verify level is error
	if logEntry["level"] != "error" {
		t.Errorf("Expected level='error', got: %v", logEntry["level"])
	}

	// Verify message
	if logEntry["message"] != "Application error" {
		t.Errorf("Expected message='Application error', got: %v", logEntry["message"])
	}

	// Verify error type
	if logEntry["error_type"] != "Retryable" {
		t.Errorf("Expected error_type='Retryable', got: %v", logEntry["error_type"])
	}

	// Verify error message
	if !strings.Contains(logEntry["message"].(string), "Application error") {
		t.Errorf("Expected message to contain 'Application error', got: %v", logEntry["message"])
	}

	// Verify context fields are automatically extracted
	if logEntry["table"] != "invTypes" {
		t.Errorf("Expected table='invTypes', got: %v", logEntry["table"])
	}
	if logEntry["operation"] != "batch_insert" {
		t.Errorf("Expected operation='batch_insert', got: %v", logEntry["operation"])
	}
	if logEntry["retry_attempt"] != float64(3) {
		t.Errorf("Expected retry_attempt=3, got: %v", logEntry["retry_attempt"])
	}

	// Verify cause is logged
	if logEntry["cause"] != "database is locked" {
		t.Errorf("Expected cause='database is locked', got: %v", logEntry["cause"])
	}
}

// TestLogAppErrorStandardError tests LogAppError with a standard error (not AppError)
func TestLogAppErrorStandardError(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf).With().Timestamp().Logger()
	logger := &Logger{logger: zl}

	// Use a standard error
	stdErr := errors.New("standard error")
	logger.LogAppError(stdErr)

	output := buf.String()

	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Should still log the error, but without AppError fields
	if logEntry["level"] != "error" {
		t.Errorf("Expected level='error', got: %v", logEntry["level"])
	}

	if logEntry["error"] != "standard error" {
		t.Errorf("Expected error='standard error', got: %v", logEntry["error"])
	}
}

// TestLogAppErrorGlobal tests the global LogAppError function
func TestLogAppErrorGlobal(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf).With().Timestamp().Logger()
	logger := &Logger{logger: zl}
	SetGlobalLogger(logger)

	appErr := apperrors.NewValidation("invalid input", nil).
		WithContext("field", "email")

	LogAppError(appErr)

	output := buf.String()
	if !strings.Contains(output, "Application error") {
		t.Error("Global LogAppError did not produce expected output")
	}

	// Clean up
	mu.Lock()
	globalLogger = nil
	mu.Unlock()
}
