package logger

import (
	"errors"
	"time"

	apperrors "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"
)

// LogHTTPRequest logs an HTTP request with method, URL, status code, and duration
func (l *Logger) LogHTTPRequest(method, url string, statusCode int, duration time.Duration) {
	l.Info("HTTP request",
		Field{Key: "method", Value: method},
		Field{Key: "url", Value: url},
		Field{Key: "status_code", Value: statusCode},
		Field{Key: "duration_ms", Value: duration.Milliseconds()},
	)
}

// LogDBQuery logs a database query with the query string, arguments, and duration
func (l *Logger) LogDBQuery(query string, args []interface{}, duration time.Duration) {
	l.Info("Database query",
		Field{Key: "query", Value: query},
		Field{Key: "args", Value: args},
		Field{Key: "duration_ms", Value: duration.Milliseconds()},
	)
}

// LogError logs an error with additional context information
func (l *Logger) LogError(err error, context map[string]interface{}) {
	fields := []Field{
		{Key: "error", Value: err.Error()},
	}

	// Add context fields
	for key, value := range context {
		fields = append(fields, Field{Key: key, Value: value})
	}

	l.Error("Error occurred", fields...)
}

// LogAppError logs an AppError with automatic context extraction
func (l *Logger) LogAppError(err error) {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		fields := []Field{
			{Key: "error_type", Value: appErr.Type.String()},
			{Key: "message", Value: appErr.Message},
		}

		// Add all context fields from the error
		for key, value := range appErr.Context {
			fields = append(fields, Field{Key: key, Value: value})
		}

		// Add the cause if present
		if appErr.Cause != nil {
			fields = append(fields, Field{Key: "cause", Value: appErr.Cause.Error()})
		}

		l.Error("Application error", fields...)
	} else {
		// Fallback for non-AppError types
		l.Error("Error occurred", Field{Key: "error", Value: err.Error()})
	}
}

// LogAppStart logs application startup with version and commit information
func (l *Logger) LogAppStart(version, commit string) {
	l.Info("Application started",
		Field{Key: "version", Value: version},
		Field{Key: "commit", Value: commit},
	)
}

// LogAppShutdown logs application shutdown with reason
func (l *Logger) LogAppShutdown(reason string) {
	l.Info("Application shutting down",
		Field{Key: "reason", Value: reason},
	)
}

// LogHTTPRequestGlobal logs an HTTP request using the global logger
func LogHTTPRequest(method, url string, statusCode int, duration time.Duration) {
	GetGlobalLogger().LogHTTPRequest(method, url, statusCode, duration)
}

// LogDBQueryGlobal logs a database query using the global logger
func LogDBQuery(query string, args []interface{}, duration time.Duration) {
	GetGlobalLogger().LogDBQuery(query, args, duration)
}

// LogErrorGlobal logs an error with context using the global logger
func LogErrorGlobal(err error, context map[string]interface{}) {
	GetGlobalLogger().LogError(err, context)
}

// LogAppError logs an AppError with automatic context extraction using the global logger
func LogAppError(err error) {
	GetGlobalLogger().LogAppError(err)
}

// LogAppStartGlobal logs application startup using the global logger
func LogAppStart(version, commit string) {
	GetGlobalLogger().LogAppStart(version, commit)
}

// LogAppShutdownGlobal logs application shutdown using the global logger
func LogAppShutdown(reason string) {
	GetGlobalLogger().LogAppShutdown(reason)
}
