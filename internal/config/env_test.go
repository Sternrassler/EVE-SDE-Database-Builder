package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestEnvVarOverrideDatabasePath tests that ESDEDB_DATABASE_PATH overrides TOML config
func TestEnvVarOverrideDatabasePath(t *testing.T) {
	os.Setenv("ESDEDB_DATABASE_PATH", "/custom/database.db")
	defer os.Unsetenv("ESDEDB_DATABASE_PATH")

	configPath := filepath.Join("testdata", "valid-config.toml")
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Database.Path != "/custom/database.db" {
		t.Errorf("expected database path /custom/database.db, got %s", cfg.Database.Path)
	}
}

// TestEnvVarOverrideWorkerCount tests that ESDEDB_IMPORT_WORKER_COUNT overrides TOML config
func TestEnvVarOverrideWorkerCount(t *testing.T) {
	os.Setenv("ESDEDB_IMPORT_WORKER_COUNT", "16")
	defer os.Unsetenv("ESDEDB_IMPORT_WORKER_COUNT")

	configPath := filepath.Join("testdata", "valid-config.toml")
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Import.Workers != 16 {
		t.Errorf("expected worker count 16, got %d", cfg.Import.Workers)
	}
}

// TestEnvVarOverrideLogLevel tests that ESDEDB_LOG_LEVEL overrides TOML config
func TestEnvVarOverrideLogLevel(t *testing.T) {
	os.Setenv("ESDEDB_LOG_LEVEL", "warn")
	defer os.Unsetenv("ESDEDB_LOG_LEVEL")

	configPath := filepath.Join("testdata", "valid-config.toml")
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Logging.Level != "warn" {
		t.Errorf("expected log level warn, got %s", cfg.Logging.Level)
	}
}

// TestEnvVarOverrideLogFormat tests that ESDEDB_LOGGING_FORMAT overrides TOML config
func TestEnvVarOverrideLogFormat(t *testing.T) {
	os.Setenv("ESDEDB_LOGGING_FORMAT", "text")
	defer os.Unsetenv("ESDEDB_LOGGING_FORMAT")

	configPath := filepath.Join("testdata", "valid-config.toml")
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Logging.Format != "text" {
		t.Errorf("expected log format text, got %s", cfg.Logging.Format)
	}
}

// TestEnvVarOverrideSdePath tests that ESDEDB_SDE_PATH overrides TOML config
func TestEnvVarOverrideSdePath(t *testing.T) {
	os.Setenv("ESDEDB_SDE_PATH", "/custom/sde-path")
	defer os.Unsetenv("ESDEDB_SDE_PATH")

	configPath := filepath.Join("testdata", "valid-config.toml")
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Import.SDEPath != "/custom/sde-path" {
		t.Errorf("expected SDE path /custom/sde-path, got %s", cfg.Import.SDEPath)
	}
}

// TestEnvVarOverrideLanguage tests that ESDEDB_LANGUAGE overrides TOML config
func TestEnvVarOverrideLanguage(t *testing.T) {
	os.Setenv("ESDEDB_LANGUAGE", "ja")
	defer os.Unsetenv("ESDEDB_LANGUAGE")

	configPath := filepath.Join("testdata", "valid-config.toml")
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Import.Language != "ja" {
		t.Errorf("expected language ja, got %s", cfg.Import.Language)
	}
}

// TestEnvVarTypeMismatch tests that invalid worker count in env var is silently ignored
func TestEnvVarTypeMismatch(t *testing.T) {
	tests := []struct {
		name          string
		workerValue   string
		expectedValue int // Expected value from TOML (8) since env var is invalid
	}{
		{
			name:          "Non-numeric string",
			workerValue:   "abc",
			expectedValue: 8, // From valid-config.toml
		},
		{
			name:          "Floating point",
			workerValue:   "4.5",
			expectedValue: 8, // From valid-config.toml
		},
		{
			name:          "Empty string",
			workerValue:   "",
			expectedValue: 8, // From valid-config.toml
		},
		{
			name:          "Special characters",
			workerValue:   "!@#$",
			expectedValue: 8, // From valid-config.toml
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.workerValue != "" {
				os.Setenv("ESDEDB_IMPORT_WORKER_COUNT", tt.workerValue)
				defer os.Unsetenv("ESDEDB_IMPORT_WORKER_COUNT")
			}

			configPath := filepath.Join("testdata", "valid-config.toml")
			cfg, err := Load(configPath)
			if err != nil {
				t.Fatalf("failed to load config: %v", err)
			}

			// Invalid env var should be silently ignored, TOML value should be used
			if cfg.Import.Workers != tt.expectedValue {
				t.Errorf("expected worker count %d (from TOML), got %d", tt.expectedValue, cfg.Import.Workers)
			}
		})
	}
}

// TestEnvVarMissing tests that TOML defaults are used when env vars are not set
func TestEnvVarMissing(t *testing.T) {
	// Ensure no env vars are set
	os.Unsetenv("ESDEDB_DATABASE_PATH")
	os.Unsetenv("ESDEDB_SDE_PATH")
	os.Unsetenv("ESDEDB_LANGUAGE")
	os.Unsetenv("ESDEDB_LOG_LEVEL")
	os.Unsetenv("ESDEDB_LOGGING_FORMAT")
	os.Unsetenv("ESDEDB_IMPORT_WORKER_COUNT")

	configPath := filepath.Join("testdata", "valid-config.toml")
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// All values should come from TOML
	if cfg.Database.Path != "./test_eve.db" {
		t.Errorf("expected database path from TOML ./test_eve.db, got %s", cfg.Database.Path)
	}
	if cfg.Import.SDEPath != "./test-sde" {
		t.Errorf("expected SDE path from TOML ./test-sde, got %s", cfg.Import.SDEPath)
	}
	if cfg.Import.Language != "de" {
		t.Errorf("expected language from TOML de, got %s", cfg.Import.Language)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("expected log level from TOML debug, got %s", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("expected log format from TOML json, got %s", cfg.Logging.Format)
	}
	if cfg.Import.Workers != 8 {
		t.Errorf("expected workers from TOML 8, got %d", cfg.Import.Workers)
	}
}

// TestEnvVarEmptyString tests that empty env vars don't override config
func TestEnvVarEmptyString(t *testing.T) {
	// Set all env vars to empty strings
	os.Setenv("ESDEDB_DATABASE_PATH", "")
	os.Setenv("ESDEDB_SDE_PATH", "")
	os.Setenv("ESDEDB_LANGUAGE", "")
	os.Setenv("ESDEDB_LOG_LEVEL", "")
	os.Setenv("ESDEDB_LOGGING_FORMAT", "")
	os.Setenv("ESDEDB_IMPORT_WORKER_COUNT", "")
	defer func() {
		os.Unsetenv("ESDEDB_DATABASE_PATH")
		os.Unsetenv("ESDEDB_SDE_PATH")
		os.Unsetenv("ESDEDB_LANGUAGE")
		os.Unsetenv("ESDEDB_LOG_LEVEL")
		os.Unsetenv("ESDEDB_LOGGING_FORMAT")
		os.Unsetenv("ESDEDB_IMPORT_WORKER_COUNT")
	}()

	configPath := filepath.Join("testdata", "valid-config.toml")
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Empty env vars should not override TOML values
	if cfg.Database.Path != "./test_eve.db" {
		t.Errorf("empty env var should not override database path, got %s", cfg.Database.Path)
	}
	if cfg.Import.SDEPath != "./test-sde" {
		t.Errorf("empty env var should not override SDE path, got %s", cfg.Import.SDEPath)
	}
	if cfg.Import.Language != "de" {
		t.Errorf("empty env var should not override language, got %s", cfg.Import.Language)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("empty env var should not override log level, got %s", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("empty env var should not override log format, got %s", cfg.Logging.Format)
	}
	if cfg.Import.Workers != 8 {
		t.Errorf("empty env var should not override workers, got %d", cfg.Import.Workers)
	}
}

// TestEnvVarAllOverrides tests that all env vars work simultaneously
func TestEnvVarAllOverrides(t *testing.T) {
	// Set all env vars
	os.Setenv("ESDEDB_DATABASE_PATH", "/env/db.db")
	os.Setenv("ESDEDB_SDE_PATH", "/env/sde")
	os.Setenv("ESDEDB_LANGUAGE", "ru")
	os.Setenv("ESDEDB_LOG_LEVEL", "error")
	os.Setenv("ESDEDB_LOGGING_FORMAT", "text")
	os.Setenv("ESDEDB_IMPORT_WORKER_COUNT", "24")
	defer func() {
		os.Unsetenv("ESDEDB_DATABASE_PATH")
		os.Unsetenv("ESDEDB_SDE_PATH")
		os.Unsetenv("ESDEDB_LANGUAGE")
		os.Unsetenv("ESDEDB_LOG_LEVEL")
		os.Unsetenv("ESDEDB_LOGGING_FORMAT")
		os.Unsetenv("ESDEDB_IMPORT_WORKER_COUNT")
	}()

	configPath := filepath.Join("testdata", "valid-config.toml")
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// All values should come from env vars
	if cfg.Database.Path != "/env/db.db" {
		t.Errorf("expected database path /env/db.db, got %s", cfg.Database.Path)
	}
	if cfg.Import.SDEPath != "/env/sde" {
		t.Errorf("expected SDE path /env/sde, got %s", cfg.Import.SDEPath)
	}
	if cfg.Import.Language != "ru" {
		t.Errorf("expected language ru, got %s", cfg.Import.Language)
	}
	if cfg.Logging.Level != "error" {
		t.Errorf("expected log level error, got %s", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "text" {
		t.Errorf("expected log format text, got %s", cfg.Logging.Format)
	}
	if cfg.Import.Workers != 24 {
		t.Errorf("expected workers 24, got %d", cfg.Import.Workers)
	}
}

// TestEnvVarWorkerCountValidation tests that env var worker count goes through validation
func TestEnvVarWorkerCountValidation(t *testing.T) {
	tests := []struct {
		name        string
		workerValue string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid worker count",
			workerValue: "4",
			expectError: false,
		},
		{
			name:        "Worker count too high",
			workerValue: "50",
			expectError: true,
			errorMsg:    "import.workers must be 0-32 (0=auto, got 50)",
		},
		{
			name:        "Negative worker count",
			workerValue: "-1",
			expectError: true,
			errorMsg:    "import.workers must be 0-32 (0=auto, got -1)",
		},
		{
			name:        "Zero auto-detects",
			workerValue: "0",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("ESDEDB_IMPORT_WORKER_COUNT", tt.workerValue)
			defer os.Unsetenv("ESDEDB_IMPORT_WORKER_COUNT")

			configPath := filepath.Join("testdata", "valid-config.toml")
			cfg, err := Load(configPath)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected validation error, got nil")
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				// For zero, check it was auto-detected
				if tt.workerValue == "0" && cfg.Import.Workers != runtime.NumCPU() {
					t.Errorf("expected workers to be auto-detected to %d, got %d", runtime.NumCPU(), cfg.Import.Workers)
				}
			}
		})
	}
}

// TestEnvVarOverridesDefaultsWhenNoTOML tests env vars override default config when no TOML file
func TestEnvVarOverridesDefaultsWhenNoTOML(t *testing.T) {
	os.Setenv("ESDEDB_DATABASE_PATH", "/env/default.db")
	os.Setenv("ESDEDB_IMPORT_WORKER_COUNT", "12")
	defer func() {
		os.Unsetenv("ESDEDB_DATABASE_PATH")
		os.Unsetenv("ESDEDB_IMPORT_WORKER_COUNT")
	}()

	// Load with non-existent config file (should use defaults + env overrides)
	configPath := filepath.Join("testdata", "nonexistent.toml")
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Env vars should override defaults
	if cfg.Database.Path != "/env/default.db" {
		t.Errorf("expected database path /env/default.db, got %s", cfg.Database.Path)
	}
	if cfg.Import.Workers != 12 {
		t.Errorf("expected workers 12, got %d", cfg.Import.Workers)
	}

	// Other values should be defaults
	if cfg.Import.Language != "en" {
		t.Errorf("expected default language en, got %s", cfg.Import.Language)
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("expected default log level info, got %s", cfg.Logging.Level)
	}
}

// TestEnvVarInvalidLogLevel tests that invalid log level from env var fails validation
func TestEnvVarInvalidLogLevel(t *testing.T) {
	os.Setenv("ESDEDB_LOG_LEVEL", "invalid")
	defer os.Unsetenv("ESDEDB_LOG_LEVEL")

	configPath := filepath.Join("testdata", "valid-config.toml")
	_, err := Load(configPath)

	if err == nil {
		t.Fatal("expected validation error for invalid log level from env var")
	}

	expectedErr := "invalid logging.level: invalid (must be: debug, info, warn, error)"
	if err.Error() != expectedErr {
		t.Errorf("expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

// TestEnvVarInvalidLanguage tests that invalid language from env var fails validation
func TestEnvVarInvalidLanguage(t *testing.T) {
	os.Setenv("ESDEDB_LANGUAGE", "invalid")
	defer os.Unsetenv("ESDEDB_LANGUAGE")

	configPath := filepath.Join("testdata", "valid-config.toml")
	_, err := Load(configPath)

	if err == nil {
		t.Fatal("expected validation error for invalid language from env var")
	}

	expectedErr := "invalid language: invalid (must be: en, de, fr, ja, ru, zh, es, ko)"
	if err.Error() != expectedErr {
		t.Errorf("expected error '%s', got '%s'", expectedErr, err.Error())
	}
}
