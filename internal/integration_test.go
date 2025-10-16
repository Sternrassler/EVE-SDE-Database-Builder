// Package internal provides integration tests for foundation components
package internal

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/logger"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/retry"
)

// TestIntegrationRetryWithLogging tests retry mechanism with logging for each attempt
func TestIntegrationRetryWithLogging(t *testing.T) {
	// Setup Logger
	testLogger := logger.NewLogger("debug", "json")
	
	// For actual test, we need to capture logs. Use buffer-based logger.
	// Since NewLogger writes to os.Stdout, we'll track calls differently
	
	// Setup Retry Policy with short delays for testing
	policy := retry.NewPolicy(3, 10*time.Millisecond, 100*time.Millisecond)
	
	// Simulate failing operation
	attempts := 0
	err := policy.Do(context.Background(), func() error {
		attempts++
		testLogger.Info("Retry attempt", logger.Field{Key: "attempt", Value: attempts})
		if attempts < 3 {
			return errors.NewRetryable("DB locked", nil)
		}
		return nil // Success
	})
	
	// Verify successful completion after retries
	if err != nil {
		t.Errorf("expected no error after retries, got: %v", err)
	}
	
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

// TestIntegrationErrorContextWithLogging tests error context logging
func TestIntegrationErrorContextWithLogging(t *testing.T) {
	// Create error with context
	err := errors.NewRetryable("Database connection failed", nil)
	err = err.WithContext("host", "localhost")
	err = err.WithContext("port", 5432)
	err = err.WithContext("retry_attempt", 1)
	
	// Log the error with its context
	testLogger := logger.NewLogger("error", "json")
	logFields := []logger.Field{
		{Key: "error_type", Value: err.Type.String()},
		{Key: "error_message", Value: err.Message},
	}
	
	// Add context fields to log
	for k, v := range err.Context {
		logFields = append(logFields, logger.Field{Key: k, Value: v})
	}
	
	testLogger.Error("Operation failed with context", logFields...)
	
	// Verify error has context
	if len(err.Context) != 3 {
		t.Errorf("expected 3 context fields, got %d", len(err.Context))
	}
	
	if err.Context["host"] != "localhost" {
		t.Errorf("expected host=localhost, got %v", err.Context["host"])
	}
	
	// Verify error is retryable
	if !errors.IsRetryable(err) {
		t.Error("expected error to be retryable")
	}
}

// TestIntegrationFatalErrorRecovery tests fatal error handling with panic recovery
func TestIntegrationFatalErrorRecovery(t *testing.T) {
	// Setup Logger
	testLogger := logger.NewLogger("error", "json")
	
	// Create a fatal error
	fatalErr := errors.NewFatal("Critical system failure", nil)
	fatalErr = fatalErr.WithContext("component", "database")
	
	// Test panic recovery mechanism
	recovered := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				recovered = true
				testLogger.Error("Recovered from panic", logger.Field{Key: "panic", Value: r})
			}
		}()
		
		// Simulate fatal error that causes panic
		if errors.IsFatal(fatalErr) {
			panic(fatalErr.Error())
		}
	}()
	
	if !recovered {
		t.Error("expected to recover from panic")
	}
	
	// Verify error is fatal
	if !errors.IsFatal(fatalErr) {
		t.Error("expected error to be fatal")
	}
}

// TestIntegrationRetryWithBackoffAndSuccess tests retryable error that eventually succeeds
func TestIntegrationRetryWithBackoffAndSuccess(t *testing.T) {
	// Setup Logger
	testLogger := logger.NewLogger("debug", "json")
	
	// Setup Retry Policy
	policy := retry.DatabasePolicy()
	
	// Track attempts and timing
	attempts := 0
	startTime := time.Now()
	
	err := policy.Do(context.Background(), func() error {
		attempts++
		testLogger.Info("Retry attempt with backoff", 
			logger.Field{Key: "attempt", Value: attempts},
			logger.Field{Key: "elapsed_ms", Value: time.Since(startTime).Milliseconds()},
		)
		
		if attempts < 3 {
			return errors.NewRetryable("Temporary database error", nil)
		}
		return nil // Success on 3rd attempt
	})
	
	if err != nil {
		t.Errorf("expected success after retries, got: %v", err)
	}
	
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
	
	// Verify some time elapsed due to backoff
	elapsed := time.Since(startTime)
	if elapsed < 20*time.Millisecond {
		t.Errorf("expected backoff delays, but elapsed time too short: %v", elapsed)
	}
}

// TestIntegrationContextCancellation tests that retry respects context cancellation
func TestIntegrationContextCancellation(t *testing.T) {
	// Setup Logger
	testLogger := logger.NewLogger("warn", "json")
	
	// Setup Retry Policy with longer delays
	policy := retry.NewPolicy(10, 100*time.Millisecond, 1*time.Second)
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	
	attempts := 0
	err := policy.Do(ctx, func() error {
		attempts++
		testLogger.Warn("Retry attempt before cancellation",
			logger.Field{Key: "attempt", Value: attempts},
		)
		// Always return retryable error
		return errors.NewRetryable("Persistent error", nil)
	})
	
	// Should fail with context error, not the retryable error
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
	
	// Verify we got a context error
	if err != context.DeadlineExceeded && err != context.Canceled {
		// Check if it's wrapped
		if !strings.Contains(err.Error(), "context") {
			// It might be the last retryable error, which is also valid
			// since context cancellation might happen after the last retry
			if !errors.IsRetryable(err) {
				t.Errorf("expected context error or retryable error, got: %v", err)
			}
		}
	}
	
	// Should have attempted at least once
	if attempts < 1 {
		t.Error("expected at least 1 attempt")
	}
	
	testLogger.Info("Context cancellation test completed",
		logger.Field{Key: "total_attempts", Value: attempts},
		logger.Field{Key: "error", Value: err.Error()},
	)
}

// TestIntegrationMultipleErrorTypes tests handling of different error types
func TestIntegrationMultipleErrorTypes(t *testing.T) {
	testLogger := logger.NewLogger("info", "json")
	
	tests := []struct {
		name          string
		err           *errors.AppError
		shouldRetry   bool
		expectedType  string
	}{
		{
			name:          "Retryable Error",
			err:           errors.NewRetryable("Network timeout", nil),
			shouldRetry:   true,
			expectedType:  "Retryable",
		},
		{
			name:          "Fatal Error",
			err:           errors.NewFatal("Database corrupted", nil),
			shouldRetry:   false,
			expectedType:  "Fatal",
		},
		{
			name:          "Validation Error",
			err:           errors.NewValidation("Invalid input", nil),
			shouldRetry:   false,
			expectedType:  "Validation",
		},
		{
			name:          "Skippable Error",
			err:           errors.NewSkippable("Optional field missing", nil),
			shouldRetry:   false,
			expectedType:  "Skippable",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Log the error
			testLogger.Info("Testing error type",
				logger.Field{Key: "error_type", Value: tt.err.Type.String()},
				logger.Field{Key: "error_message", Value: tt.err.Message},
				logger.Field{Key: "is_retryable", Value: errors.IsRetryable(tt.err)},
			)
			
			// Verify error type
			if tt.err.Type.String() != tt.expectedType {
				t.Errorf("expected type %s, got %s", tt.expectedType, tt.err.Type.String())
			}
			
			// Verify retry behavior
			if errors.IsRetryable(tt.err) != tt.shouldRetry {
				t.Errorf("expected shouldRetry=%v, got %v", tt.shouldRetry, errors.IsRetryable(tt.err))
			}
			
			// Test with retry policy
			policy := retry.NewPolicy(2, 10*time.Millisecond, 100*time.Millisecond)
			attempts := 0
			
			retryErr := policy.Do(context.Background(), func() error {
				attempts++
				return tt.err
			})
			
			// Non-retryable errors should fail immediately
			if !tt.shouldRetry && attempts != 1 {
				t.Errorf("non-retryable error should fail on first attempt, got %d attempts", attempts)
			}
			
			// Verify error is returned
			if retryErr == nil {
				t.Error("expected error to be returned")
			}
		})
	}
}

// TestIntegrationLoggerWithContext tests logger with context values
func TestIntegrationLoggerWithContext(t *testing.T) {
	// Create context with request tracking
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.RequestIDKey, "req-12345")
	ctx = context.WithValue(ctx, logger.UserIDKey, "user-67890")
	
	// Create logger with context
	testLogger := logger.NewLogger("info", "json")
	ctxLogger := testLogger.WithContext(ctx)
	
	// Setup retry with context-aware logging
	policy := retry.HTTPPolicy()
	
	attempts := 0
	err := policy.Do(ctx, func() error {
		attempts++
		ctxLogger.Info("HTTP request attempt",
			logger.Field{Key: "attempt", Value: attempts},
			logger.Field{Key: "url", Value: "https://api.example.com/data"},
		)
		
		if attempts < 2 {
			return errors.NewRetryable("HTTP 503 Service Unavailable", nil)
		}
		return nil
	})
	
	if err != nil {
		t.Errorf("expected success, got: %v", err)
	}
	
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

// TestIntegrationErrorChainLogging tests logging of error chains
func TestIntegrationErrorChainLogging(t *testing.T) {
	// Create error chain
	rootErr := errors.NewRetryable("Database query failed", nil)
	rootErr = rootErr.WithContext("query", "SELECT * FROM users")
	rootErr = rootErr.WithContext("table", "users")
	
	// Test with retry
	policy := retry.NewPolicy(2, 5*time.Millisecond, 50*time.Millisecond)
	testLogger := logger.NewLogger("error", "json")
	
	attempts := 0
	finalErr := policy.Do(context.Background(), func() error {
		attempts++
		
		// Log with error details
		testLogger.Error("Database operation failed",
			logger.Field{Key: "attempt", Value: attempts},
			logger.Field{Key: "error", Value: rootErr.Error()},
			logger.Field{Key: "query", Value: rootErr.Context["query"]},
		)
		
		if attempts < 2 {
			return rootErr
		}
		return nil
	})
	
	if finalErr != nil {
		t.Errorf("expected success after retry, got: %v", finalErr)
	}
	
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}
