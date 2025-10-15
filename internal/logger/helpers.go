package logger

import (
	"time"
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

// LogAppStartGlobal logs application startup using the global logger
func LogAppStart(version, commit string) {
	GetGlobalLogger().LogAppStart(version, commit)
}

// LogAppShutdownGlobal logs application shutdown using the global logger
func LogAppShutdown(reason string) {
	GetGlobalLogger().LogAppShutdown(reason)
}
