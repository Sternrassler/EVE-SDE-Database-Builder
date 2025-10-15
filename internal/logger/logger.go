// Package logger provides a structured logging wrapper around zerolog
// with configurable log levels and output formats.
package logger

import (
	"context"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/rs/zerolog"
)

var (
	globalLogger *Logger
	mu           sync.RWMutex
)

// ContextKey is the type used for context keys
type ContextKey string

const (
	// RequestIDKey is the context key for request IDs
	RequestIDKey ContextKey = "RequestID"
	// UserIDKey is the context key for user IDs
	UserIDKey ContextKey = "UserID"
)

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

// Logger wraps zerolog.Logger to provide a consistent logging interface
type Logger struct {
	logger zerolog.Logger
}

// NewLogger creates a new Logger instance with the specified log level and format.
// level can be: "debug", "info", "warn", "error", "fatal"
// format can be: "json" or "text"
func NewLogger(level string, format string) *Logger {
	var output io.Writer = os.Stdout

	// Configure output format
	if strings.ToLower(format) == "text" {
		output = zerolog.ConsoleWriter{Out: os.Stdout}
	}

	// Parse log level
	logLevel := parseLogLevel(level)
	zerolog.SetGlobalLevel(logLevel)

	zl := zerolog.New(output).With().Timestamp().Logger()

	return &Logger{
		logger: zl,
	}
}

// parseLogLevel converts a string log level to zerolog.Level
func parseLogLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger *Logger) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger = logger
}

// GetGlobalLogger returns the global logger instance.
// If no global logger has been set, it creates a new one with default settings.
func GetGlobalLogger() *Logger {
	mu.RLock()
	logger := globalLogger
	mu.RUnlock()

	if logger == nil {
		mu.Lock()
		defer mu.Unlock()
		// Double-check after acquiring write lock
		if globalLogger == nil {
			globalLogger = NewLogger("info", "json")
		}
		return globalLogger
	}

	return logger
}

// Debug logs a debug message with optional fields
func (l *Logger) Debug(msg string, fields ...Field) {
	event := l.logger.Debug()
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// Info logs an info message with optional fields
func (l *Logger) Info(msg string, fields ...Field) {
	event := l.logger.Info()
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// Warn logs a warning message with optional fields
func (l *Logger) Warn(msg string, fields ...Field) {
	event := l.logger.Warn()
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// Error logs an error message with optional fields
func (l *Logger) Error(msg string, fields ...Field) {
	event := l.logger.Error()
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// Fatal logs a fatal message with optional fields and exits the program
func (l *Logger) Fatal(msg string, fields ...Field) {
	event := l.logger.Fatal()
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// WithContext creates a new logger with context-specific fields
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// Extract common context values if they exist
	newLogger := l.logger.With()

	// Check for RequestID in context
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		newLogger = newLogger.Interface(string(RequestIDKey), requestID)
	}

	// Check for UserID in context
	if userID := ctx.Value(UserIDKey); userID != nil {
		newLogger = newLogger.Interface(string(UserIDKey), userID)
	}

	return &Logger{
		logger: newLogger.Logger(),
	}
}
