package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestDefaultConfig tests that default configuration values are correct
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Version
	if cfg.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", cfg.Version)
	}

	// Database defaults
	if cfg.Database.Path != "./eve_sde.db" {
		t.Errorf("expected database path ./eve_sde.db, got %s", cfg.Database.Path)
	}
	if cfg.Database.JournalMode != "WAL" {
		t.Errorf("expected journal mode WAL, got %s", cfg.Database.JournalMode)
	}
	if cfg.Database.CacheSizeMB != 64 {
		t.Errorf("expected cache size 64, got %d", cfg.Database.CacheSizeMB)
	}

	// Import defaults
	if cfg.Import.SDEPath != "./sde-JSONL" {
		t.Errorf("expected SDE path ./sde-JSONL, got %s", cfg.Import.SDEPath)
	}
	if cfg.Import.Language != "en" {
		t.Errorf("expected language en, got %s", cfg.Import.Language)
	}
	if cfg.Import.Workers != 0 {
		t.Errorf("expected workers 0 (auto), got %d", cfg.Import.Workers)
	}

	// Logging defaults
	if cfg.Logging.Level != "info" {
		t.Errorf("expected log level info, got %s", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "text" {
		t.Errorf("expected log format text, got %s", cfg.Logging.Format)
	}

	// Update defaults
	if cfg.Update.Enabled != false {
		t.Errorf("expected update enabled false, got %v", cfg.Update.Enabled)
	}
}

// TestLoadConfigFromTOML tests loading configuration from a TOML file
func TestLoadConfigFromTOML(t *testing.T) {
	configPath := filepath.Join("testdata", "valid-config.toml")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify values from the TOML file
	if cfg.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", cfg.Version)
	}
	if cfg.Database.Path != "./test_eve.db" {
		t.Errorf("expected database path ./test_eve.db, got %s", cfg.Database.Path)
	}
	if cfg.Database.JournalMode != "WAL" {
		t.Errorf("expected journal mode WAL, got %s", cfg.Database.JournalMode)
	}
	if cfg.Database.CacheSizeMB != 128 {
		t.Errorf("expected cache size 128, got %d", cfg.Database.CacheSizeMB)
	}
	if cfg.Import.SDEPath != "./test-sde" {
		t.Errorf("expected SDE path ./test-sde, got %s", cfg.Import.SDEPath)
	}
	if cfg.Import.Language != "de" {
		t.Errorf("expected language de, got %s", cfg.Import.Language)
	}
	if cfg.Import.Workers != 8 {
		t.Errorf("expected workers 8, got %d", cfg.Import.Workers)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("expected log level debug, got %s", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("expected log format json, got %s", cfg.Logging.Format)
	}
	if cfg.Update.Enabled != true {
		t.Errorf("expected update enabled true, got %v", cfg.Update.Enabled)
	}
	if cfg.Update.CheckURL != "https://test.example.com/version.json" {
		t.Errorf("expected check URL https://test.example.com/version.json, got %s", cfg.Update.CheckURL)
	}
}

// TestLoadConfigNonExistent tests loading when config file doesn't exist
func TestLoadConfigNonExistent(t *testing.T) {
	configPath := filepath.Join("testdata", "nonexistent.toml")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("loading nonexistent config should use defaults, got error: %v", err)
	}

	// Should have default values
	if cfg.Database.Path != "./eve_sde.db" {
		t.Errorf("expected default database path, got %s", cfg.Database.Path)
	}
	if cfg.Import.Language != "en" {
		t.Errorf("expected default language en, got %s", cfg.Import.Language)
	}
}

// TestEnvVarOverride tests that environment variables override TOML config
func TestEnvVarOverride(t *testing.T) {
	// Set environment variables
	_ = os.Setenv("ESDEDB_DATABASE_PATH", "/custom/db.db")
	_ = os.Setenv("ESDEDB_SDE_PATH", "/custom/sde")
	_ = os.Setenv("ESDEDB_LANGUAGE", "fr")
	_ = os.Setenv("ESDEDB_LOG_LEVEL", "error")
	defer func() {
		_ = os.Unsetenv("ESDEDB_DATABASE_PATH")
		_ = os.Unsetenv("ESDEDB_SDE_PATH")
		_ = os.Unsetenv("ESDEDB_LANGUAGE")
		_ = os.Unsetenv("ESDEDB_LOG_LEVEL")
	}()

	configPath := filepath.Join("testdata", "valid-config.toml")
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Environment variables should override TOML values
	if cfg.Database.Path != "/custom/db.db" {
		t.Errorf("expected env var database path /custom/db.db, got %s", cfg.Database.Path)
	}
	if cfg.Import.SDEPath != "/custom/sde" {
		t.Errorf("expected env var SDE path /custom/sde, got %s", cfg.Import.SDEPath)
	}
	if cfg.Import.Language != "fr" {
		t.Errorf("expected env var language fr, got %s", cfg.Import.Language)
	}
	if cfg.Logging.Level != "error" {
		t.Errorf("expected env var log level error, got %s", cfg.Logging.Level)
	}
}

// TestEnvVarOverrideDefaults tests env vars override defaults when no config file
func TestEnvVarOverrideDefaults(t *testing.T) {
	_ = os.Setenv("ESDEDB_DATABASE_PATH", "/env/db.db")
	defer func() { _ = os.Unsetenv("ESDEDB_DATABASE_PATH") }()

	configPath := filepath.Join("testdata", "nonexistent.toml")
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Env var should override default
	if cfg.Database.Path != "/env/db.db" {
		t.Errorf("expected env var database path /env/db.db, got %s", cfg.Database.Path)
	}
	// Other values should be defaults
	if cfg.Import.Language != "en" {
		t.Errorf("expected default language en, got %s", cfg.Import.Language)
	}
}

// TestValidationMissingDatabasePath tests validation fails when database path is empty
func TestValidationMissingDatabasePath(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Database.Path = ""

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error for missing database path")
	}
	if err.Error() != "database.path is required" {
		t.Errorf("expected error 'database.path is required', got %v", err)
	}
}

// TestValidationMissingPathsFromFile tests validation fails when loading config with empty paths
func TestValidationMissingPathsFromFile(t *testing.T) {
	configPath := filepath.Join("testdata", "invalid-config.toml")

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected validation error for missing database path")
	}
	// Should fail on database.path first
	if err.Error() != "database.path is required" {
		t.Errorf("expected error 'database.path is required', got %v", err)
	}
}

// TestValidationMissingSdePath tests validation fails when SDE path is missing
func TestValidationMissingSdePath(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Import.SDEPath = ""

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error for missing SDE path")
	}
	if err.Error() != "import.sde_path is required" {
		t.Errorf("expected error 'import.sde_path is required', got %v", err)
	}
}

// TestValidationInvalidWorkers tests validation fails for invalid worker counts
func TestValidationInvalidWorkers(t *testing.T) {
	tests := []struct {
		name    string
		workers int
		wantErr bool
	}{
		{"Negative workers", -1, true},
		{"Too many workers", 33, true},
		{"Valid explicit", 8, false},
		{"Valid max", 32, false},
		{"Valid min", 1, false},
		{"Auto detect (will be converted)", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Import.Workers = tt.workers

			err := cfg.Validate()
			if tt.wantErr && err == nil {
				t.Errorf("expected validation error for workers=%d", tt.workers)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected validation error for workers=%d: %v", tt.workers, err)
			}
		})
	}
}

// TestValidationInvalidWorkersFromFile tests loading config with invalid workers
func TestValidationInvalidWorkersFromFile(t *testing.T) {
	configPath := filepath.Join("testdata", "invalid-workers.toml")

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected validation error for invalid workers")
	}
	// Error message should mention workers range
	if err.Error() != "import.workers must be 0-32 (0=auto, got 50)" {
		t.Errorf("expected workers range error, got %v", err)
	}
}

// TestValidationInvalidLanguage tests validation fails for invalid language
func TestValidationInvalidLanguage(t *testing.T) {
	configPath := filepath.Join("testdata", "invalid-language.toml")

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected validation error for invalid language")
	}
	// Error should mention invalid language
	expectedErr := "invalid language: invalid (must be: en, de, fr, ja, ru, zh, es, ko)"
	if err.Error() != expectedErr {
		t.Errorf("expected error '%s', got %v", expectedErr, err)
	}
}

// TestValidationInvalidLogLevel tests validation fails for invalid log level
func TestValidationInvalidLogLevel(t *testing.T) {
	configPath := filepath.Join("testdata", "invalid-log-level.toml")

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected validation error for invalid log level")
	}
	// Error should mention invalid log level
	expectedErr := "invalid logging.level: invalid (must be: debug, info, warn, error)"
	if err.Error() != expectedErr {
		t.Errorf("expected error '%s', got %v", expectedErr, err)
	}
}

// TestValidationAllLanguages tests that all valid languages pass validation
func TestValidationAllLanguages(t *testing.T) {
	validLanguages := []string{"en", "de", "fr", "ja", "ru", "zh", "es", "ko"}

	for _, lang := range validLanguages {
		t.Run(lang, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Import.Language = lang

			err := cfg.Validate()
			if err != nil {
				t.Errorf("language %s should be valid, got error: %v", lang, err)
			}
		})
	}
}

// TestValidationAllLogLevels tests that all valid log levels pass validation
func TestValidationAllLogLevels(t *testing.T) {
	validLevels := []string{"debug", "info", "warn", "error"}

	for _, level := range validLevels {
		t.Run(level, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Logging.Level = level

			err := cfg.Validate()
			if err != nil {
				t.Errorf("log level %s should be valid, got error: %v", level, err)
			}
		})
	}
}

// TestWorkerCountAutoDetect tests that workers=0 is converted to runtime.NumCPU()
func TestWorkerCountAutoDetect(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Import.Workers = 0

	err := cfg.Validate()
	if err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	expectedWorkers := runtime.NumCPU()
	if cfg.Import.Workers != expectedWorkers {
		t.Errorf("expected workers to be auto-detected to %d (runtime.NumCPU()), got %d",
			expectedWorkers, cfg.Import.Workers)
	}
}

// TestWorkerCountAutoDetectFromFile tests auto-detection works when loading from file
func TestWorkerCountAutoDetectFromFile(t *testing.T) {
	// Create a temp config with workers = 0
	configPath := filepath.Join("testdata", "auto-workers.toml")
	content := `version = "1.0.0"

[database]
path = "./test.db"

[import]
sde_path = "./sde"
language = "en"
workers = 0  # Auto-detect

[logging]
level = "info"
format = "text"
`
	err := os.WriteFile(configPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create temp config: %v", err)
	}
	defer func() { _ = os.Remove(configPath) }()

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	expectedWorkers := runtime.NumCPU()
	if cfg.Import.Workers != expectedWorkers {
		t.Errorf("expected workers to be auto-detected to %d (runtime.NumCPU()), got %d",
			expectedWorkers, cfg.Import.Workers)
	}
}

// TestLoadEmptyConfig tests loading an empty config file
func TestLoadEmptyConfig(t *testing.T) {
	configPath := filepath.Join("testdata", "empty-config.toml")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("loading empty config should use defaults, got error: %v", err)
	}

	// Should have default values
	if cfg.Database.Path != "./eve_sde.db" {
		t.Errorf("expected default database path, got %s", cfg.Database.Path)
	}
	if cfg.Import.Language != "en" {
		t.Errorf("expected default language en, got %s", cfg.Import.Language)
	}
}

// TestApplyEnvVarsWithEmptyValues tests that empty env vars don't override config
func TestApplyEnvVarsWithEmptyValues(t *testing.T) {
	// Set empty environment variable
	_ = os.Setenv("ESDEDB_DATABASE_PATH", "")
	defer func() { _ = os.Unsetenv("ESDEDB_DATABASE_PATH") }()

	cfg := DefaultConfig()
	originalPath := cfg.Database.Path
	applyEnvVars(&cfg)

	// Empty env var should not override
	if cfg.Database.Path != originalPath {
		t.Errorf("empty env var should not override config, got %s", cfg.Database.Path)
	}
}

// TestInvalidTOMLSyntax tests handling of invalid TOML syntax
func TestInvalidTOMLSyntax(t *testing.T) {
	// Create a temp file with invalid TOML
	configPath := filepath.Join("testdata", "invalid-syntax.toml")
	invalidContent := `
[database
path = "missing bracket"
`
	err := os.WriteFile(configPath, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("failed to create temp config: %v", err)
	}
	defer func() { _ = os.Remove(configPath) }()

	_, err = Load(configPath)
	if err == nil {
		t.Fatal("expected error for invalid TOML syntax")
	}
	// Error should mention parsing failure
	if err.Error() == "" {
		t.Error("error message should not be empty")
	}
}

// TestValidateCompleteConfig tests validation with a fully specified valid config
func TestValidateCompleteConfig(t *testing.T) {
	cfg := Config{
		Version: "1.0.0",
		Database: DatabaseConfig{
			Path:        "/valid/path.db",
			JournalMode: "WAL",
			CacheSizeMB: 128,
		},
		Import: ImportConfig{
			SDEPath:  "/valid/sde",
			Language: "en",
			Workers:  4,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		Update: UpdateConfig{
			Enabled:  true,
			CheckURL: "https://example.com/version",
		},
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("valid config should pass validation, got error: %v", err)
	}
}
