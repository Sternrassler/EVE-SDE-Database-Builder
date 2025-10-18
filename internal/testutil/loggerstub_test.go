package testutil_test

import (
	"testing"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/logger"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/testutil"
)

func TestLoggerStub_Debug(t *testing.T) {
	stub := testutil.NewLoggerStub()

	stub.Debug("debug message")

	if stub.MessageCount() != 1 {
		t.Fatalf("expected 1 message, got %d", stub.MessageCount())
	}

	msg := stub.Messages()[0]
	if msg.Level != "DEBUG" {
		t.Errorf("expected level DEBUG, got %s", msg.Level)
	}
	if msg.Message != "debug message" {
		t.Errorf("expected message 'debug message', got %q", msg.Message)
	}
}

func TestLoggerStub_Info(t *testing.T) {
	stub := testutil.NewLoggerStub()

	stub.Info("info message")

	if stub.MessageCount() != 1 {
		t.Fatalf("expected 1 message, got %d", stub.MessageCount())
	}

	msg := stub.Messages()[0]
	if msg.Level != "INFO" {
		t.Errorf("expected level INFO, got %s", msg.Level)
	}
	if msg.Message != "info message" {
		t.Errorf("expected message 'info message', got %q", msg.Message)
	}
}

func TestLoggerStub_Warn(t *testing.T) {
	stub := testutil.NewLoggerStub()

	stub.Warn("warning message")

	if stub.MessageCount() != 1 {
		t.Fatalf("expected 1 message, got %d", stub.MessageCount())
	}

	msg := stub.Messages()[0]
	if msg.Level != "WARN" {
		t.Errorf("expected level WARN, got %s", msg.Level)
	}
	if msg.Message != "warning message" {
		t.Errorf("expected message 'warning message', got %q", msg.Message)
	}
}

func TestLoggerStub_Error(t *testing.T) {
	stub := testutil.NewLoggerStub()

	stub.Error("error message")

	if stub.MessageCount() != 1 {
		t.Fatalf("expected 1 message, got %d", stub.MessageCount())
	}

	msg := stub.Messages()[0]
	if msg.Level != "ERROR" {
		t.Errorf("expected level ERROR, got %s", msg.Level)
	}
	if msg.Message != "error message" {
		t.Errorf("expected message 'error message', got %q", msg.Message)
	}
}

func TestLoggerStub_Fatal(t *testing.T) {
	stub := testutil.NewLoggerStub()

	// Fatal should record but NOT exit (unlike real logger)
	stub.Fatal("fatal message")

	if stub.MessageCount() != 1 {
		t.Fatalf("expected 1 message, got %d", stub.MessageCount())
	}

	msg := stub.Messages()[0]
	if msg.Level != "FATAL" {
		t.Errorf("expected level FATAL, got %s", msg.Level)
	}
	if msg.Message != "fatal message" {
		t.Errorf("expected message 'fatal message', got %q", msg.Message)
	}
}

func TestLoggerStub_WithFields(t *testing.T) {
	stub := testutil.NewLoggerStub()

	fields := []logger.Field{
		{Key: "user", Value: "alice"},
		{Key: "count", Value: 42},
	}

	stub.Info("operation completed", fields...)

	if stub.MessageCount() != 1 {
		t.Fatalf("expected 1 message, got %d", stub.MessageCount())
	}

	msg := stub.Messages()[0]
	if len(msg.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(msg.Fields))
	}

	if msg.Fields[0].Key != "user" || msg.Fields[0].Value != "alice" {
		t.Errorf("unexpected first field: %+v", msg.Fields[0])
	}

	if msg.Fields[1].Key != "count" || msg.Fields[1].Value != 42 {
		t.Errorf("unexpected second field: %+v", msg.Fields[1])
	}
}

func TestLoggerStub_MultipleLevels(t *testing.T) {
	stub := testutil.NewLoggerStub()

	stub.Debug("debug 1")
	stub.Info("info 1")
	stub.Warn("warn 1")
	stub.Error("error 1")
	stub.Debug("debug 2")

	if stub.MessageCount() != 5 {
		t.Fatalf("expected 5 messages, got %d", stub.MessageCount())
	}

	if stub.DebugCount() != 2 {
		t.Errorf("expected 2 debug messages, got %d", stub.DebugCount())
	}

	if stub.InfoCount() != 1 {
		t.Errorf("expected 1 info message, got %d", stub.InfoCount())
	}

	if stub.WarnCount() != 1 {
		t.Errorf("expected 1 warn message, got %d", stub.WarnCount())
	}

	if stub.ErrorCount() != 1 {
		t.Errorf("expected 1 error message, got %d", stub.ErrorCount())
	}

	if stub.FatalCount() != 0 {
		t.Errorf("expected 0 fatal messages, got %d", stub.FatalCount())
	}
}

func TestLoggerStub_MessagesAtLevel(t *testing.T) {
	stub := testutil.NewLoggerStub()

	stub.Info("info 1")
	stub.Debug("debug 1")
	stub.Info("info 2")
	stub.Error("error 1")
	stub.Info("info 3")

	infoMessages := stub.MessagesAtLevel("INFO")
	if len(infoMessages) != 3 {
		t.Fatalf("expected 3 info messages, got %d", len(infoMessages))
	}

	expectedInfoMessages := []string{"info 1", "info 2", "info 3"}
	for i, expected := range expectedInfoMessages {
		if infoMessages[i].Message != expected {
			t.Errorf("info message %d: expected %q, got %q", i, expected, infoMessages[i].Message)
		}
	}
}

func TestLoggerStub_HasMessage(t *testing.T) {
	stub := testutil.NewLoggerStub()

	stub.Info("test message")
	stub.Debug("another message")

	if !stub.HasMessage("test message") {
		t.Error("expected to find 'test message'")
	}

	if !stub.HasMessage("another message") {
		t.Error("expected to find 'another message'")
	}

	if stub.HasMessage("nonexistent message") {
		t.Error("did not expect to find 'nonexistent message'")
	}
}

func TestLoggerStub_HasMessageAtLevel(t *testing.T) {
	stub := testutil.NewLoggerStub()

	stub.Info("test message")
	stub.Debug("test message")

	if !stub.HasMessageAtLevel("INFO", "test message") {
		t.Error("expected to find 'test message' at INFO level")
	}

	if !stub.HasMessageAtLevel("DEBUG", "test message") {
		t.Error("expected to find 'test message' at DEBUG level")
	}

	if stub.HasMessageAtLevel("ERROR", "test message") {
		t.Error("did not expect to find 'test message' at ERROR level")
	}
}

func TestLoggerStub_ContainsMessage(t *testing.T) {
	stub := testutil.NewLoggerStub()

	stub.Info("processing user request")
	stub.Debug("completed successfully")

	if !stub.ContainsMessage("user request") {
		t.Error("expected to find message containing 'user request'")
	}

	if !stub.ContainsMessage("completed") {
		t.Error("expected to find message containing 'completed'")
	}

	if stub.ContainsMessage("failed") {
		t.Error("did not expect to find message containing 'failed'")
	}
}

func TestLoggerStub_LastMessage(t *testing.T) {
	stub := testutil.NewLoggerStub()

	// No messages initially
	if stub.LastMessage() != nil {
		t.Error("expected nil last message when no messages")
	}

	stub.Info("first message")
	stub.Debug("second message")
	stub.Error("third message")

	last := stub.LastMessage()
	if last == nil {
		t.Fatal("expected non-nil last message")
	}

	if last.Level != "ERROR" {
		t.Errorf("expected last message level ERROR, got %s", last.Level)
	}

	if last.Message != "third message" {
		t.Errorf("expected last message 'third message', got %q", last.Message)
	}
}

func TestLoggerStub_Reset(t *testing.T) {
	stub := testutil.NewLoggerStub()

	// Add some messages
	stub.Info("message 1")
	stub.Debug("message 2")
	stub.Error("message 3")

	if stub.MessageCount() != 3 {
		t.Fatalf("expected 3 messages before reset, got %d", stub.MessageCount())
	}

	// Reset
	stub.Reset()

	if stub.MessageCount() != 0 {
		t.Errorf("expected 0 messages after reset, got %d", stub.MessageCount())
	}

	if stub.LastMessage() != nil {
		t.Error("expected nil last message after reset")
	}
}

func TestLoggerStub_String(t *testing.T) {
	stub := testutil.NewLoggerStub()

	stub.Info("first")
	stub.Debug("second")
	stub.Error("third")

	str := stub.String()
	if str == "" {
		t.Error("expected non-empty string representation")
	}

	// Check that all messages are present
	if !contains(str, "INFO") {
		t.Error("expected string to contain 'INFO'")
	}
	if !contains(str, "DEBUG") {
		t.Error("expected string to contain 'DEBUG'")
	}
	if !contains(str, "ERROR") {
		t.Error("expected string to contain 'ERROR'")
	}
	if !contains(str, "first") {
		t.Error("expected string to contain 'first'")
	}
	if !contains(str, "second") {
		t.Error("expected string to contain 'second'")
	}
	if !contains(str, "third") {
		t.Error("expected string to contain 'third'")
	}
}

func TestLoggerStub_StringWithFields(t *testing.T) {
	stub := testutil.NewLoggerStub()

	fields := []logger.Field{
		{Key: "user", Value: "alice"},
		{Key: "action", Value: "login"},
	}
	stub.Info("user logged in", fields...)

	str := stub.String()
	if !contains(str, "user") {
		t.Error("expected string to contain field key 'user'")
	}
	if !contains(str, "action") {
		t.Error("expected string to contain field key 'action'")
	}
}

func TestSilentLogger(t *testing.T) {
	stub := testutil.NewSilentLogger()

	// Log many messages
	for i := 0; i < 100; i++ {
		stub.Info("message")
		stub.Debug("debug")
		stub.Error("error")
	}

	// Silent logger should not record messages
	if stub.MessageCount() != 0 {
		t.Errorf("expected silent logger to have 0 messages, got %d", stub.MessageCount())
	}

	if stub.LastMessage() != nil {
		t.Error("expected silent logger to have nil last message")
	}
}

func TestLoggerStub_Concurrency(t *testing.T) {
	stub := testutil.NewLoggerStub()

	// Test concurrent logging
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				stub.Info("concurrent message")
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 100 messages
	if stub.MessageCount() != 100 {
		t.Errorf("expected 100 messages from concurrent logging, got %d", stub.MessageCount())
	}
}

// Helper function for string contains
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
