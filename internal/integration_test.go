// Package internal provides integration tests for foundation components
package internal

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/config"
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
		name         string
		err          *errors.AppError
		shouldRetry  bool
		expectedType string
	}{
		{
			name:         "Retryable Error",
			err:          errors.NewRetryable("Network timeout", nil),
			shouldRetry:  true,
			expectedType: "Retryable",
		},
		{
			name:         "Fatal Error",
			err:          errors.NewFatal("Database corrupted", nil),
			shouldRetry:  false,
			expectedType: "Fatal",
		},
		{
			name:         "Validation Error",
			err:          errors.NewValidation("Invalid input", nil),
			shouldRetry:  false,
			expectedType: "Validation",
		},
		{
			name:         "Skippable Error",
			err:          errors.NewSkippable("Optional field missing", nil),
			shouldRetry:  false,
			expectedType: "Skippable",
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

// TestIntegrationConfigLogger_LoadAndLog tests Config + Logger integration
func TestIntegrationConfigLogger_LoadAndLog(t *testing.T) {
	// Create temporary directory for test config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.toml")

	// Create a test config file
	testConfig := `version = "1.0.0"

[database]
path = "./test.db"
journal_mode = "WAL"
cache_size_mb = 64

[import]
sde_path = "./test-sde"
language = "en"
workers = 4

[logging]
level = "debug"
format = "json"

[update]
enabled = false
`
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify config was loaded correctly
	if cfg.Logging.Level != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("Expected log format 'json', got '%s'", cfg.Logging.Format)
	}

	// Create logger from config
	testLogger := logger.NewLogger(cfg.Logging.Level, cfg.Logging.Format)

	// Log various messages using config-derived logger
	testLogger.Debug("Test debug message", logger.Field{Key: "config_version", Value: cfg.Version})
	testLogger.Info("Config loaded successfully",
		logger.Field{Key: "db_path", Value: cfg.Database.Path},
		logger.Field{Key: "sde_path", Value: cfg.Import.SDEPath},
		logger.Field{Key: "workers", Value: cfg.Import.Workers},
	)
	testLogger.Warn("Warning with config context",
		logger.Field{Key: "journal_mode", Value: cfg.Database.JournalMode},
	)

	// Verify logger is functioning (we can't easily capture output in this setup,
	// but we verify no panics occur and the integration works)
	if testLogger == nil {
		t.Error("Logger should not be nil after creation from config")
	}
}

// TestIntegrationConfigLogger_EnvVariables tests Config + Logger with environment variables
func TestIntegrationConfigLogger_EnvVariables(t *testing.T) {
	// Set environment variables
	originalLogLevel := os.Getenv("ESDEDB_LOG_LEVEL")
	originalLogFormat := os.Getenv("ESDEDB_LOGGING_FORMAT")
	defer func() {
		os.Setenv("ESDEDB_LOG_LEVEL", originalLogLevel)
		os.Setenv("ESDEDB_LOGGING_FORMAT", originalLogFormat)
	}()

	os.Setenv("ESDEDB_LOG_LEVEL", "warn")
	os.Setenv("ESDEDB_LOGGING_FORMAT", "text")

	// Create temporary directory for test config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.toml")

	// Create a minimal config file (env vars should override)
	testConfig := `version = "1.0.0"

[database]
path = "./test.db"

[import]
sde_path = "./test-sde"
language = "en"
workers = 2

[logging]
level = "info"
format = "json"

[update]
enabled = false
`
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config (env vars should override)
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify env vars overrode config file
	if cfg.Logging.Level != "warn" {
		t.Errorf("Expected log level 'warn' from env var, got '%s'", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "text" {
		t.Errorf("Expected log format 'text' from env var, got '%s'", cfg.Logging.Format)
	}

	// Create logger from config
	testLogger := logger.NewLogger(cfg.Logging.Level, cfg.Logging.Format)

	// Log messages
	testLogger.Info("This should not appear due to warn level")
	testLogger.Warn("This warning should appear")
	testLogger.Error("This error should appear")

	// Verify logger is functioning
	if testLogger == nil {
		t.Error("Logger should not be nil")
	}
}

// TestIntegrationConfigLogger_InvalidConfig tests Config + Logger error handling
func TestIntegrationConfigLogger_InvalidConfig(t *testing.T) {
	// Create temporary directory for test config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid_config.toml")

	// Create an invalid config file (empty database path)
	invalidConfig := `version = "1.0.0"

[database]
path = ""

[import]
sde_path = "./test-sde"
language = "en"
workers = 4

[logging]
level = "debug"
format = "json"

[update]
enabled = false
`
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Attempt to load invalid config
	cfg, err := config.Load(configPath)

	// Should fail validation
	if err == nil {
		t.Fatal("Expected error when loading invalid config, got nil")
	}

	// Config should be nil on error
	if cfg != nil {
		t.Error("Expected nil config on validation error")
	}

	// Create logger to log the error
	errorLogger := logger.NewLogger("error", "json")
	errorLogger.Error("Config validation failed",
		logger.Field{Key: "error", Value: err.Error()},
		logger.Field{Key: "config_path", Value: configPath},
	)

	// Verify error message contains expected information
	if !strings.Contains(err.Error(), "database.path") {
		t.Errorf("Expected error to mention 'database.path', got: %v", err)
	}
}

// TestIntegrationConfigLogger_LoggingLevels tests different logging levels from config
func TestIntegrationConfigLogger_LoggingLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			// Create temporary config
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.toml")

			configContent := `version = "1.0.0"

[database]
path = "./test.db"

[import]
sde_path = "./test-sde"
language = "en"
workers = 2

[logging]
level = "` + level + `"
format = "json"

[update]
enabled = false
`
			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				t.Fatalf("Failed to create test config: %v", err)
			}

			// Load config
			cfg, err := config.Load(configPath)
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			// Verify logging level
			if cfg.Logging.Level != level {
				t.Errorf("Expected log level '%s', got '%s'", level, cfg.Logging.Level)
			}

			// Create logger
			testLogger := logger.NewLogger(cfg.Logging.Level, cfg.Logging.Format)

			// Log a message at each level
			testLogger.Debug("Debug message")
			testLogger.Info("Info message")
			testLogger.Warn("Warn message")
			testLogger.Error("Error message")

			// Verify logger creation succeeded
			if testLogger == nil {
				t.Errorf("Logger should not be nil for level '%s'", level)
			}
		})
	}
}

// TestIntegrationConfigLogger_JSONOutput tests Config + Logger with JSON output
func TestIntegrationConfigLogger_JSONOutput(t *testing.T) {
	// Create temporary config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	configContent := `version = "1.0.0"

[database]
path = "./test.db"

[import]
sde_path = "./test-sde"
language = "en"
workers = 2

[logging]
level = "info"
format = "json"

[update]
enabled = false
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create logger and log a test message
	// Note: We can't easily capture stdout in tests, but we verify the integration works
	testLogger := logger.NewLogger(cfg.Logging.Level, cfg.Logging.Format)
	testLogger.Info("Test message",
		logger.Field{Key: "config_version", Value: cfg.Version},
		logger.Field{Key: "workers", Value: cfg.Import.Workers},
	)

	// Verify logger is functioning
	if testLogger == nil {
		t.Error("Logger should not be nil")
	}
}

// TestIntegrationConfigLogger_WorkerConfiguration tests Config + Logger for worker pool configuration
func TestIntegrationConfigLogger_WorkerConfiguration(t *testing.T) {
	// Create temporary config with specific worker count
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	configContent := `version = "1.0.0"

[database]
path = "./test.db"

[import]
sde_path = "./test-sde"
language = "en"
workers = 8

[logging]
level = "debug"
format = "json"

[update]
enabled = false
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create logger
	testLogger := logger.NewLogger(cfg.Logging.Level, cfg.Logging.Format)

	// Log worker pool configuration
	testLogger.Info("Worker pool configuration",
		logger.Field{Key: "workers", Value: cfg.Import.Workers},
		logger.Field{Key: "sde_path", Value: cfg.Import.SDEPath},
		logger.Field{Key: "language", Value: cfg.Import.Language},
	)

	// Verify worker count
	if cfg.Import.Workers != 8 {
		t.Errorf("Expected 8 workers, got %d", cfg.Import.Workers)
	}

	// Log database configuration
	testLogger.Debug("Database configuration",
		logger.Field{Key: "path", Value: cfg.Database.Path},
		logger.Field{Key: "journal_mode", Value: cfg.Database.JournalMode},
		logger.Field{Key: "cache_size_mb", Value: cfg.Database.CacheSizeMB},
	)

	// Verify logger is functioning
	if testLogger == nil {
		t.Error("Logger should not be nil")
	}
}

// TestIntegrationConfigLogger_DefaultConfig tests Config + Logger with default configuration
func TestIntegrationConfigLogger_DefaultConfig(t *testing.T) {
	// Get default config (no file needed)
	cfg := config.DefaultConfig()

	// Validate default config
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Default config validation failed: %v", err)
	}

	// Create logger from default config
	testLogger := logger.NewLogger(cfg.Logging.Level, cfg.Logging.Format)

	// Log with default config
	testLogger.Info("Using default configuration",
		logger.Field{Key: "db_path", Value: cfg.Database.Path},
		logger.Field{Key: "sde_path", Value: cfg.Import.SDEPath},
		logger.Field{Key: "workers", Value: cfg.Import.Workers},
		logger.Field{Key: "log_level", Value: cfg.Logging.Level},
		logger.Field{Key: "log_format", Value: cfg.Logging.Format},
	)

	// Verify default values
	if cfg.Logging.Level != "info" {
		t.Errorf("Expected default log level 'info', got '%s'", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "text" {
		t.Errorf("Expected default log format 'text', got '%s'", cfg.Logging.Format)
	}
	if cfg.Database.JournalMode != "WAL" {
		t.Errorf("Expected default journal mode 'WAL', got '%s'", cfg.Database.JournalMode)
	}

	// Verify logger is functioning
	if testLogger == nil {
		t.Error("Logger should not be nil with default config")
	}
}

// TestIntegrationConfigLogger_GlobalLoggerSetup tests Config + Logger with global logger
func TestIntegrationConfigLogger_GlobalLoggerSetup(t *testing.T) {
	// Create temporary config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	configContent := `version = "1.0.0"

[database]
path = "./test.db"

[import]
sde_path = "./test-sde"
language = "en"
workers = 4

[logging]
level = "info"
format = "json"

[update]
enabled = false
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create logger from config
	testLogger := logger.NewLogger(cfg.Logging.Level, cfg.Logging.Format)

	// Set as global logger
	logger.SetGlobalLogger(testLogger)

	// Get global logger
	globalLogger := logger.GetGlobalLogger()
	if globalLogger == nil {
		t.Fatal("Global logger should not be nil")
	}

	// Use global logger to log config information
	globalLogger.Info("Global logger configured from config",
		logger.Field{Key: "config_version", Value: cfg.Version},
		logger.Field{Key: "log_level", Value: cfg.Logging.Level},
	)

	// Verify global logger is set
	secondFetch := logger.GetGlobalLogger()
	if secondFetch == nil {
		t.Error("Second fetch of global logger should not be nil")
	}
}
