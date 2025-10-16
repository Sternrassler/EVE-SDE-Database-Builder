package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	apperrors "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/logger"
)

func main() {
	// Example 1: Basic logger with JSON format
	jsonLogger := logger.NewLogger("info", "json")
	jsonLogger.Info("Application started", logger.Field{Key: "version", Value: "0.1.0"})

	// Example 2: Logger with text format for development
	textLogger := logger.NewLogger("debug", "text")
	textLogger.Debug("Debug message", logger.Field{Key: "module", Value: "main"})
	textLogger.Info("Info message with multiple fields",
		logger.Field{Key: "user", Value: "admin"},
		logger.Field{Key: "action", Value: "login"},
	)
	textLogger.Warn("Warning message", logger.Field{Key: "reason", Value: "deprecated feature"})
	textLogger.Error("Error message", logger.Field{Key: "error", Value: "connection timeout"})

	// Example 3: Using global logger
	logger.SetGlobalLogger(logger.NewLogger("info", "json"))
	globalLogger := logger.GetGlobalLogger()
	globalLogger.Info("Using global logger")

	// Example 4: Context-aware logging
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.RequestIDKey, "req-12345")
	ctx = context.WithValue(ctx, logger.UserIDKey, "user-67890")

	ctxLogger := globalLogger.WithContext(ctx)
	ctxLogger.Info("Request processed", logger.Field{Key: "duration_ms", Value: 150})

	// Example 5: Helper functions - HTTP Request Logging
	jsonLogger.LogHTTPRequest("GET", "/api/types", 200, 45*time.Millisecond)
	jsonLogger.LogHTTPRequest("POST", "/api/users", 201, 120*time.Millisecond)
	jsonLogger.LogHTTPRequest("GET", "/api/missing", 404, 15*time.Millisecond)

	// Example 6: Helper functions - Database Query Logging
	jsonLogger.LogDBQuery("SELECT * FROM types WHERE id = ?", []interface{}{123}, 2*time.Millisecond)
	jsonLogger.LogDBQuery("INSERT INTO items (name, value) VALUES (?, ?)", []interface{}{"test", 42}, 5*time.Millisecond)

	// Example 7: Helper functions - Error Logging with Context
	err := errors.New("failed to parse file")
	jsonLogger.LogError(err, map[string]interface{}{
		"file": "types.jsonl",
		"line": 42,
		"type": "parse_error",
	})

	// Example 8: Helper functions - Application Lifecycle
	jsonLogger.LogAppStart("0.1.0", "abc123def")
	jsonLogger.LogAppShutdown("graceful shutdown")

	// Example 9: Using global helper functions
	logger.LogHTTPRequest("GET", "/api/health", 200, 5*time.Millisecond)
	logger.LogDBQuery("SELECT COUNT(*) FROM types", []interface{}{}, 10*time.Millisecond)
	logger.LogErrorGlobal(errors.New("unexpected error"), map[string]interface{}{"context": "example"})
	logger.LogAppStart("1.0.0", "xyz789")
	logger.LogAppShutdown("normal exit")

	// Example 10: Error Context Management (NEW)
	fmt.Println("\n=== Error Context Management Demo ===")

	// Create an error with context
	dbErr := apperrors.NewRetryable("DB locked", errors.New("database is locked")).
		WithContext("table", "invTypes").
		WithContext("operation", "batch_insert").
		WithContext("retry_attempt", 3)

	fmt.Printf("\n1. Error with context (text): %s\n", dbErr.Error())

	// Log the error with automatic context extraction
	jsonLogger.LogAppError(dbErr)

	// Validation error with context
	validationErr := apperrors.NewValidation("invalid input", nil).
		WithContext("field", "email").
		WithContext("value", "not-an-email").
		WithContext("line_number", 42)

	fmt.Printf("\n2. Validation error: %s\n", validationErr.Error())

	// Fatal error with context
	fatalErr := apperrors.NewFatal("connection failed", errors.New("timeout")).
		WithContext("host", "localhost").
		WithContext("port", 5432).
		WithContext("timeout_ms", 5000)

	fmt.Printf("\n3. Fatal error: %s\n", fatalErr.Error())
	jsonLogger.LogAppError(fatalErr)

	fmt.Println("\n=== Demo Complete ===")
}
