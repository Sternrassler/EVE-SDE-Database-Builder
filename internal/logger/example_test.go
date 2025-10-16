package logger_test

import (
	"context"
	"time"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/logger"
)

// ExampleNewLogger demonstrates creating a new logger with different configurations
func ExampleNewLogger() {
	// Create a logger with info level and text format for human-readable output
	log := logger.NewLogger("info", "text")
	log.Info("Application initialized")
	
	// Output demonstriert strukturierte Logging-Ausgabe
	// Actual output varies, so we don't include // Output: comment
}

// ExampleNewLogger_json demonstrates JSON-formatted logging
func ExampleNewLogger_json() {
	// Create a logger with JSON format for production
	log := logger.NewLogger("info", "json")
	log.Info("User logged in", logger.Field{Key: "user_id", Value: 12345})
	
	// JSON format is useful for log aggregation systems
}

// ExampleLogger_Info demonstrates basic info logging with structured fields
func ExampleLogger_Info() {
	log := logger.NewLogger("info", "text")
	
	// Log a message with structured fields
	log.Info("Database connection established",
		logger.Field{Key: "host", Value: "localhost"},
		logger.Field{Key: "port", Value: 5432},
		logger.Field{Key: "database", Value: "eve_sde"})
}

// ExampleLogger_Error demonstrates error logging
func ExampleLogger_Error() {
	log := logger.NewLogger("error", "text")
	
	// Log an error with context
	log.Error("Failed to process record",
		logger.Field{Key: "record_id", Value: 999},
		logger.Field{Key: "error", Value: "invalid data"})
}

// ExampleLogger_WithContext demonstrates context-based logging
func ExampleLogger_WithContext() {
	log := logger.NewLogger("info", "text")
	
	// Create a context with request ID
	ctx := context.WithValue(context.Background(), logger.RequestIDKey, "req-abc-123")
	
	// Create a logger that includes context values
	contextLogger := log.WithContext(ctx)
	
	// The RequestID will be automatically included in all log messages
	contextLogger.Info("Processing started")
	contextLogger.Info("Processing completed")
}

// ExampleLogger_LogHTTPRequest demonstrates HTTP request logging
func ExampleLogger_LogHTTPRequest() {
	log := logger.NewLogger("info", "text")
	
	// Log an HTTP request with timing information
	log.LogHTTPRequest("GET", "/api/v1/users", 200, 45*time.Millisecond)
}

// ExampleLogger_LogDBQuery demonstrates database query logging
func ExampleLogger_LogDBQuery() {
	log := logger.NewLogger("info", "text")
	
	// Log a database query with parameters and timing
	log.LogDBQuery(
		"SELECT * FROM users WHERE id = ?",
		[]interface{}{123},
		12*time.Millisecond,
	)
}

// ExampleLogger_LogAppStart demonstrates application startup logging
func ExampleLogger_LogAppStart() {
	log := logger.NewLogger("info", "text")
	
	// Log application startup with version and commit information
	log.LogAppStart("1.0.0", "abc123def")
}

// ExampleSetGlobalLogger demonstrates using the global logger
func ExampleSetGlobalLogger() {
	// Set up a global logger for the application
	logger.SetGlobalLogger(logger.NewLogger("info", "json"))
	
	// Use the global logger via convenience functions
	logger.LogAppStart("1.0.0", "abc123")
	
	// The global logger is accessible from anywhere in the application
}

// ExampleGetGlobalLogger demonstrates retrieving the global logger
func ExampleGetGlobalLogger() {
	// Set a global logger
	logger.SetGlobalLogger(logger.NewLogger("debug", "text"))
	
	// Retrieve and use the global logger
	log := logger.GetGlobalLogger()
	log.Debug("Debug information", logger.Field{Key: "detail", Value: "some data"})
}
