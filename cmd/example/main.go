package main

import (
	"context"

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
}
