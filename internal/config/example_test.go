package config_test

import (
	"fmt"
	"os"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/config"
)

// ExampleDefaultConfig demonstrates getting default configuration
func ExampleDefaultConfig() {
	cfg := config.DefaultConfig()

	fmt.Printf("Database Path: %s\n", cfg.Database.Path)
	fmt.Printf("Journal Mode: %s\n", cfg.Database.JournalMode)
	fmt.Printf("Language: %s\n", cfg.Import.Language)
	fmt.Printf("Log Level: %s\n", cfg.Logging.Level)
	// Output:
	// Database Path: ./eve_sde.db
	// Journal Mode: WAL
	// Language: en
	// Log Level: info
}

// ExampleLoad demonstrates loading configuration from a TOML file
func ExampleLoad() {
	// Create a temporary config file for demonstration
	tmpfile, _ := os.CreateTemp("", "config-*.toml")
	defer os.Remove(tmpfile.Name())

	content := `version = "1.0.0"

[database]
path = "./test.db"
journal_mode = "WAL"
cache_size_mb = 128

[import]
sde_path = "./test-sde"
language = "de"
workers = 4

[logging]
level = "debug"
format = "json"
`
	tmpfile.WriteString(content)
	tmpfile.Close()

	cfg, err := config.Load(tmpfile.Name())
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Database Path: %s\n", cfg.Database.Path)
	fmt.Printf("Language: %s\n", cfg.Import.Language)
	fmt.Printf("Workers: %d\n", cfg.Import.Workers)
	fmt.Printf("Log Level: %s\n", cfg.Logging.Level)
	// Output:
	// Database Path: ./test.db
	// Language: de
	// Workers: 4
	// Log Level: debug
}

// ExampleLoad_nonexistent demonstrates that missing config file uses defaults
func ExampleLoad_nonexistent() {
	// Load from a non-existent file (should use defaults)
	cfg, err := config.Load("/nonexistent/config.toml")

	if err == nil {
		fmt.Printf("Database Path: %s\n", cfg.Database.Path)
		fmt.Printf("Language: %s\n", cfg.Import.Language)
	}
	// Output:
	// Database Path: ./eve_sde.db
	// Language: en
}

// ExampleConfig_Validate demonstrates configuration validation
func ExampleConfig_Validate() {
	cfg := config.Config{
		Database: config.DatabaseConfig{
			Path:        "./eve.db",
			JournalMode: "WAL",
			CacheSizeMB: 64,
		},
		Import: config.ImportConfig{
			SDEPath:  "./sde",
			Language: "en",
			Workers:  4,
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	err := cfg.Validate()
	if err == nil {
		fmt.Println("Configuration is valid")
	}
	// Output: Configuration is valid
}

// ExampleConfig_Validate_invalid demonstrates validation failure
func ExampleConfig_Validate_invalid() {
	cfg := config.Config{
		Database: config.DatabaseConfig{
			Path: "", // Invalid: empty path
		},
		Import: config.ImportConfig{
			SDEPath: "./sde",
		},
	}

	err := cfg.Validate()
	if err != nil {
		fmt.Println("Validation failed: database.path is required")
	}
	// Output: Validation failed: database.path is required
}

// ExampleConfig_Validate_autoWorkers demonstrates auto worker configuration
func ExampleConfig_Validate_autoWorkers() {
	cfg := config.DefaultConfig()
	cfg.Import.Workers = 0 // Auto mode

	err := cfg.Validate()
	if err == nil {
		// After validation, Workers is set to runtime.NumCPU()
		fmt.Printf("Workers auto-configured: %v\n", cfg.Import.Workers > 0)
	}
	// Output: Workers auto-configured: true
}

// Example_envOverride demonstrates environment variable override
func Example_envOverride() {
	// Set environment variable
	os.Setenv("ESDEDB_DATABASE_PATH", "/custom/path.db")
	os.Setenv("ESDEDB_LANGUAGE", "de")
	defer os.Unsetenv("ESDEDB_DATABASE_PATH")
	defer os.Unsetenv("ESDEDB_LANGUAGE")

	// Load config (env vars will override defaults)
	cfg, err := config.Load("/nonexistent/config.toml")
	if err == nil {
		fmt.Printf("Database Path: %s\n", cfg.Database.Path)
		fmt.Printf("Language: %s\n", cfg.Import.Language)
	}
	// Output:
	// Database Path: /custom/path.db
	// Language: de
}

// ExampleDatabaseConfig demonstrates database configuration
func ExampleDatabaseConfig() {
	cfg := config.DefaultConfig()

	fmt.Printf("Path: %s\n", cfg.Database.Path)
	fmt.Printf("Journal Mode: %s\n", cfg.Database.JournalMode)
	fmt.Printf("Cache Size: %dMB\n", cfg.Database.CacheSizeMB)
	// Output:
	// Path: ./eve_sde.db
	// Journal Mode: WAL
	// Cache Size: 64MB
}

// ExampleImportConfig demonstrates import configuration
func ExampleImportConfig() {
	cfg := config.DefaultConfig()

	fmt.Printf("SDE Path: %s\n", cfg.Import.SDEPath)
	fmt.Printf("Language: %s\n", cfg.Import.Language)
	fmt.Printf("Workers (0=auto): %d\n", cfg.Import.Workers)
	// Output:
	// SDE Path: ./sde-JSONL
	// Language: en
	// Workers (0=auto): 0
}

// ExampleLoggingConfig demonstrates logging configuration
func ExampleLoggingConfig() {
	cfg := config.DefaultConfig()

	fmt.Printf("Level: %s\n", cfg.Logging.Level)
	fmt.Printf("Format: %s\n", cfg.Logging.Format)
	// Output:
	// Level: info
	// Format: text
}
