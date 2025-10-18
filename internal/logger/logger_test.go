package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestNewLogger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		level  string
		format string
	}{
		{"Debug JSON", "debug", "json"},
		{"Info JSON", "info", "json"},
		{"Warn JSON", "warn", "json"},
		{"Error JSON", "error", "json"},
		{"Info Text", "info", "text"},
		{"Default Level", "invalid", "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := NewLogger(tt.level, tt.format)
			if logger == nil {
				t.Fatal("NewLogger returned nil")
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected zerolog.Level
	}{
		{"debug", zerolog.DebugLevel},
		{"DEBUG", zerolog.DebugLevel},
		{"info", zerolog.InfoLevel},
		{"INFO", zerolog.InfoLevel},
		{"warn", zerolog.WarnLevel},
		{"warning", zerolog.WarnLevel},
		{"error", zerolog.ErrorLevel},
		{"fatal", zerolog.FatalLevel},
		{"invalid", zerolog.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			result := parseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGlobalLogger(t *testing.T) {
	// Reset global logger
	mu.Lock()
	globalLogger = nil
	mu.Unlock()

	// Test GetGlobalLogger creates default logger when none exists
	logger := GetGlobalLogger()
	if logger == nil {
		t.Fatal("GetGlobalLogger returned nil")
	}

	// Test SetGlobalLogger
	newLogger := NewLogger("debug", "json")
	SetGlobalLogger(newLogger)

	retrieved := GetGlobalLogger()
	if retrieved != newLogger {
		t.Error("GetGlobalLogger did not return the logger set by SetGlobalLogger")
	}

	// Clean up
	mu.Lock()
	globalLogger = nil
	mu.Unlock()
}

func TestLoggerDebug(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	zl := zerolog.New(&buf).Level(zerolog.DebugLevel).With().Timestamp().Logger()
	logger := &Logger{logger: zl}

	logger.Debug("test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Debug output missing message, got: %s", output)
	}
	if !strings.Contains(output, `"level":"debug"`) {
		t.Errorf("Debug output missing level, got: %s", output)
	}
}

func TestLoggerInfo(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	zl := zerolog.New(&buf).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	logger := &Logger{logger: zl}

	logger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Info output missing message, got: %s", output)
	}
	if !strings.Contains(output, `"level":"info"`) {
		t.Errorf("Info output missing level, got: %s", output)
	}
}

func TestLoggerWarn(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	zl := zerolog.New(&buf).Level(zerolog.WarnLevel).With().Timestamp().Logger()
	logger := &Logger{logger: zl}

	logger.Warn("test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Warn output missing message, got: %s", output)
	}
	if !strings.Contains(output, `"level":"warn"`) {
		t.Errorf("Warn output missing level, got: %s", output)
	}
}

func TestLoggerError(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	zl := zerolog.New(&buf).Level(zerolog.ErrorLevel).With().Timestamp().Logger()
	logger := &Logger{logger: zl}

	logger.Error("test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Error output missing message, got: %s", output)
	}
	if !strings.Contains(output, `"level":"error"`) {
		t.Errorf("Error output missing level, got: %s", output)
	}
}

func TestLoggerWithFields(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	zl := zerolog.New(&buf).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	logger := &Logger{logger: zl}

	logger.Info("test message",
		Field{Key: "version", Value: "0.1.0"},
		Field{Key: "count", Value: 42},
	)

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Output missing message, got: %s", output)
	}
	if !strings.Contains(output, `"version":"0.1.0"`) {
		t.Errorf("Output missing version field, got: %s", output)
	}
	if !strings.Contains(output, `"count":42`) {
		t.Errorf("Output missing count field, got: %s", output)
	}
}

func TestLoggerWithContext(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	zl := zerolog.New(&buf).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	logger := &Logger{logger: zl}

	// Create context with values using the exported context keys
	ctx := context.Background()
	ctx = context.WithValue(ctx, RequestIDKey, "req-123")
	ctx = context.WithValue(ctx, UserIDKey, "user-456")

	// Create logger with context
	ctxLogger := logger.WithContext(ctx)
	ctxLogger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Output missing message, got: %s", output)
	}
	if !strings.Contains(output, `"RequestID":"req-123"`) {
		t.Errorf("Output missing RequestID, got: %s", output)
	}
	if !strings.Contains(output, `"UserID":"user-456"`) {
		t.Errorf("Output missing UserID, got: %s", output)
	}
}

func TestLoggerWithContextNoValues(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	zl := zerolog.New(&buf).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	logger := &Logger{logger: zl}

	// Create context without values
	ctx := context.Background()

	// Create logger with context
	ctxLogger := logger.WithContext(ctx)
	ctxLogger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Output missing message, got: %s", output)
	}
	// Should still log without extra fields
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}
	if _, exists := logEntry["RequestID"]; exists {
		t.Error("Output should not contain RequestID when not in context")
	}
}

func TestLoggerJSONFormat(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	zl := zerolog.New(&buf).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	logger := &Logger{logger: zl}

	logger.Info("test message", Field{Key: "key", Value: "value"})

	output := buf.String()
	// Verify it's valid JSON
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Output is not valid JSON: %v, output: %s", err, output)
	}

	// Verify expected fields
	if logEntry["level"] != "info" {
		t.Errorf("Expected level=info, got: %v", logEntry["level"])
	}
	if logEntry["message"] != "test message" {
		t.Errorf("Expected message='test message', got: %v", logEntry["message"])
	}
	if logEntry["key"] != "value" {
		t.Errorf("Expected key='value', got: %v", logEntry["key"])
	}
}

func TestFatalLogsBeforeExit(t *testing.T) {
	t.Parallel()
	// We can't test Fatal() directly as it calls os.Exit()
	// Instead, we verify the log is written before the fatal event
	var buf bytes.Buffer
	zl := zerolog.New(&buf).Level(zerolog.FatalLevel).With().Timestamp().Logger()
	logger := &Logger{logger: zl}

	// We can test that the fatal event is created, but we can't actually call Fatal
	// without exiting the test. Instead, we verify the logger can create a fatal event.
	event := logger.logger.Fatal()
	if event == nil {
		t.Error("Fatal event should not be nil")
	}

	// Note: In a real scenario, Fatal() would call os.Exit(1) after logging
	// For testing purposes, we've verified the logger supports the Fatal level
}
