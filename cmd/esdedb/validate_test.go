package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestValidateCmd_ValidConfig(t *testing.T) {
	// Create a temporary valid config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "valid-config.toml")

	validConfig := `version = "1.0.0"

[database]
path = "./test.db"

[import]
sde_path = "./sde"
language = "en"
workers = 4

[logging]
level = "info"
format = "text"

[update]
enabled = false
`
	err := os.WriteFile(configFile, []byte(validConfig), 0644)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	// Set up the command
	configPath = configFile
	cmd := newValidateCmd()

	// Execute the command
	err = cmd.Execute()
	if err != nil {
		t.Errorf("expected no error for valid config, got: %v", err)
	}
}

func TestValidateCmd_InvalidConfig_EmptyDatabasePath(t *testing.T) {
	// Create a temporary invalid config file (empty database path)
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "invalid-config.toml")

	invalidConfig := `version = "1.0.0"

[database]
path = ""

[import]
sde_path = "./sde"
language = "en"
workers = 4

[logging]
level = "info"
format = "text"

[update]
enabled = false
`
	err := os.WriteFile(configFile, []byte(invalidConfig), 0644)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	// Set up the command
	configPath = configFile
	cmd := newValidateCmd()

	// Execute the command - should return error
	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid config with empty database path, got nil")
	}

	expectedErr := "database.path is required"
	if err != nil && err.Error() != expectedErr {
		t.Errorf("expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestValidateCmd_InvalidConfig_InvalidLanguage(t *testing.T) {
	// Create a temporary invalid config file (invalid language)
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "invalid-lang.toml")

	invalidConfig := `version = "1.0.0"

[database]
path = "./test.db"

[import]
sde_path = "./sde"
language = "invalid"
workers = 4

[logging]
level = "info"
format = "text"

[update]
enabled = false
`
	err := os.WriteFile(configFile, []byte(invalidConfig), 0644)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	// Set up the command
	configPath = configFile
	cmd := newValidateCmd()

	// Execute the command - should return error
	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid language, got nil")
	}
}

func TestValidateCmd_InvalidConfig_WorkersOutOfRange(t *testing.T) {
	// Create a temporary invalid config file (workers out of range)
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "invalid-workers.toml")

	invalidConfig := `version = "1.0.0"

[database]
path = "./test.db"

[import]
sde_path = "./sde"
language = "en"
workers = 100

[logging]
level = "info"
format = "text"

[update]
enabled = false
`
	err := os.WriteFile(configFile, []byte(invalidConfig), 0644)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	// Set up the command
	configPath = configFile
	cmd := newValidateCmd()

	// Execute the command - should return error
	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for workers out of range, got nil")
	}
}

func TestValidateCmd_NonExistentFile_UsesDefaults(t *testing.T) {
	// Point to a non-existent file
	configPath = "/tmp/does-not-exist-validate-test.toml"

	// Make sure the file doesn't exist
	os.Remove(configPath)

	// Set up the command
	cmd := newValidateCmd()

	// Execute the command - should succeed with defaults
	err := cmd.Execute()
	if err != nil {
		t.Errorf("expected no error when file doesn't exist (uses defaults), got: %v", err)
	}
}

func TestValidateCmd_Help(t *testing.T) {
	// Test that help text is available
	cmd := newValidateCmd()

	if cmd.Use != "validate" {
		t.Errorf("expected Use to be 'validate', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}
}

func TestValidateCmd_Output(t *testing.T) {
	// Create a temporary valid config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test-output.toml")

	validConfig := `version = "1.0.0"

[database]
path = "./output-test.db"

[import]
sde_path = "./sde-test"
language = "de"
workers = 8

[logging]
level = "debug"
format = "json"

[update]
enabled = false
`
	err := os.WriteFile(configFile, []byte(validConfig), 0644)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	// Set config path
	configPath = configFile
	cmd := newValidateCmd()

	// Execute the command
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Note: Output verification is done in manual testing since fmt.Printf writes to os.Stdout
	// The important part is that the command executes without error
}

// TestValidateCmd_Integration tests the validate command as it would be called from CLI
func TestValidateCmd_Integration(t *testing.T) {
	// Create a temporary valid config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "integration.toml")

	validConfig := `version = "1.0.0"

[database]
path = "./integration.db"
journal_mode = "WAL"
cache_size_mb = 64

[import]
sde_path = "./sde-integration"
language = "en"
workers = 4

[logging]
level = "info"
format = "text"

[update]
enabled = false
`
	err := os.WriteFile(configFile, []byte(validConfig), 0644)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	// Simulate CLI execution
	rootCmd := &cobra.Command{Use: "esdedb"}
	configPath = configFile
	validateCmd := newValidateCmd()
	rootCmd.AddCommand(validateCmd)

	// Execute
	rootCmd.SetArgs([]string{"validate"})
	err = rootCmd.Execute()
	if err != nil {
		t.Errorf("integration test failed: %v", err)
	}
}
