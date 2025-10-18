// Package testutil provides shared testing utilities for EVE SDE Database Builder tests.
package testutil

import (
	"bytes"
	"context"
	"sync"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/logger"
)

// LoggerStub is a test logger implementation that captures log messages for testing.
// It provides methods to verify that the correct log messages were emitted during tests.
type LoggerStub struct {
	mu       sync.RWMutex
	messages []LogMessage
	silent   bool // If true, messages are recorded but not stored (for performance)
}

// LogMessage represents a single log entry captured by LoggerStub.
type LogMessage struct {
	Level   string
	Message string
	Fields  []logger.Field
}

// NewLoggerStub creates a new logger stub for testing.
// By default, it records all log messages.
func NewLoggerStub() *LoggerStub {
	return &LoggerStub{
		messages: make([]LogMessage, 0),
		silent:   false,
	}
}

// NewSilentLogger creates a logger stub that doesn't record messages.
// This is useful for tests where logging is needed but verification isn't required.
func NewSilentLogger() *LoggerStub {
	return &LoggerStub{
		messages: make([]LogMessage, 0),
		silent:   true,
	}
}

// Debug implements logger interface for debug level.
func (l *LoggerStub) Debug(msg string, fields ...logger.Field) {
	l.record("DEBUG", msg, fields)
}

// Info implements logger interface for info level.
func (l *LoggerStub) Info(msg string, fields ...logger.Field) {
	l.record("INFO", msg, fields)
}

// Warn implements logger interface for warn level.
func (l *LoggerStub) Warn(msg string, fields ...logger.Field) {
	l.record("WARN", msg, fields)
}

// Error implements logger interface for error level.
func (l *LoggerStub) Error(msg string, fields ...logger.Field) {
	l.record("ERROR", msg, fields)
}

// Fatal implements logger interface for fatal level.
// Note: Unlike the real logger, this does NOT exit the program.
func (l *LoggerStub) Fatal(msg string, fields ...logger.Field) {
	l.record("FATAL", msg, fields)
}

// WithContext creates a new logger with context-specific fields.
// For the stub, this returns a new instance that shares the message buffer.
func (l *LoggerStub) WithContext(_ context.Context) *LoggerStub {
	// For simplicity, return the same logger
	// In a more sophisticated implementation, we could extract context values
	return l
}

// record adds a log message to the internal buffer.
func (l *LoggerStub) record(level, msg string, fields []logger.Field) {
	if l.silent {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.messages = append(l.messages, LogMessage{
		Level:   level,
		Message: msg,
		Fields:  fields,
	})
}

// Messages returns all recorded log messages.
func (l *LoggerStub) Messages() []LogMessage {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Return a copy to prevent modifications
	result := make([]LogMessage, len(l.messages))
	copy(result, l.messages)
	return result
}

// MessageCount returns the number of recorded messages.
func (l *LoggerStub) MessageCount() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.messages)
}

// MessagesAtLevel returns all messages at the specified level.
func (l *LoggerStub) MessagesAtLevel(level string) []LogMessage {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]LogMessage, 0)
	for _, msg := range l.messages {
		if msg.Level == level {
			result = append(result, msg)
		}
	}
	return result
}

// HasMessage checks if a message with the given text was logged at any level.
func (l *LoggerStub) HasMessage(text string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, msg := range l.messages {
		if msg.Message == text {
			return true
		}
	}
	return false
}

// HasMessageAtLevel checks if a message with the given text was logged at the specified level.
func (l *LoggerStub) HasMessageAtLevel(level, text string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, msg := range l.messages {
		if msg.Level == level && msg.Message == text {
			return true
		}
	}
	return false
}

// ContainsMessage checks if any message contains the given substring.
func (l *LoggerStub) ContainsMessage(substring string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, msg := range l.messages {
		if contains(msg.Message, substring) {
			return true
		}
	}
	return false
}

// Reset clears all recorded messages.
func (l *LoggerStub) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.messages = make([]LogMessage, 0)
}

// String returns a human-readable representation of all logged messages.
func (l *LoggerStub) String() string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var buf bytes.Buffer
	for i, msg := range l.messages {
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString("[")
		buf.WriteString(msg.Level)
		buf.WriteString("] ")
		buf.WriteString(msg.Message)

		if len(msg.Fields) > 0 {
			buf.WriteString(" {")
			for j, field := range msg.Fields {
				if j > 0 {
					buf.WriteString(", ")
				}
				buf.WriteString(field.Key)
				buf.WriteString("=")
				// Simple string representation
				buf.WriteString(formatValue(field.Value))
			}
			buf.WriteString("}")
		}
	}
	return buf.String()
}

// LastMessage returns the most recent log message, or nil if no messages.
func (l *LoggerStub) LastMessage() *LogMessage {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if len(l.messages) == 0 {
		return nil
	}
	// Return a copy
	msg := l.messages[len(l.messages)-1]
	return &msg
}

// DebugCount returns the number of debug messages.
func (l *LoggerStub) DebugCount() int {
	return len(l.MessagesAtLevel("DEBUG"))
}

// InfoCount returns the number of info messages.
func (l *LoggerStub) InfoCount() int {
	return len(l.MessagesAtLevel("INFO"))
}

// WarnCount returns the number of warn messages.
func (l *LoggerStub) WarnCount() int {
	return len(l.MessagesAtLevel("WARN"))
}

// ErrorCount returns the number of error messages.
func (l *LoggerStub) ErrorCount() int {
	return len(l.MessagesAtLevel("ERROR"))
}

// FatalCount returns the number of fatal messages.
func (l *LoggerStub) FatalCount() int {
	return len(l.MessagesAtLevel("FATAL"))
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func formatValue(v interface{}) string {
	if v == nil {
		return "nil"
	}

	switch val := v.(type) {
	case string:
		return val
	case int, int8, int16, int32, int64:
		return intToString(val)
	case uint, uint8, uint16, uint32, uint64:
		return intToString(val)
	case float32, float64:
		return floatToString(val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		// For complex types, just return type name
		return "<value>"
	}
}

func intToString(v interface{}) string {
	// Simple integer conversion without fmt package
	switch val := v.(type) {
	case int:
		return itoa(int64(val))
	case int32:
		return itoa(int64(val))
	case int64:
		return itoa(val)
	case uint:
		return itoa(int64(val))
	case uint32:
		return itoa(int64(val))
	case uint64:
		return itoa(int64(val))
	default:
		return "<int>"
	}
}

func floatToString(_ interface{}) string {
	// Simplified float representation
	return "<float>"
}

func itoa(i int64) string {
	if i == 0 {
		return "0"
	}

	neg := i < 0
	if neg {
		i = -i
	}

	var buf [20]byte
	pos := len(buf)

	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}

	if neg {
		pos--
		buf[pos] = '-'
	}

	return string(buf[pos:])
}
