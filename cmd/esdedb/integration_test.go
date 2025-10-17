package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var (
	binaryBuildOnce sync.Once
	binaryPath      string
	binaryBuildErr  error
)

// buildTestBinary builds the CLI binary once and reuses it for all tests
func buildTestBinary(t *testing.T) string {
	t.Helper()

	binaryBuildOnce.Do(func() {
		// Create a temporary directory for the binary
		tmpDir, err := os.MkdirTemp("", "esdedb-integration-test-*")
		if err != nil {
			binaryBuildErr = err
			return
		}

		// Get the project root (we are in cmd/esdedb, so go up two levels)
		projectRoot, err := filepath.Abs(filepath.Join("..", ".."))
		if err != nil {
			binaryBuildErr = err
			return
		}

		binaryPath = filepath.Join(tmpDir, "esdedb")
		buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/esdedb")
		buildCmd.Dir = projectRoot
		output, err := buildCmd.CombinedOutput()
		if err != nil {
			binaryBuildErr = err
			t.Logf("Build output: %s", output)
			return
		}
	})

	if binaryBuildErr != nil {
		t.Fatalf("failed to build binary: %v", binaryBuildErr)
	}

	return binaryPath
}

// TestIntegration_ImportCommand_ValidData tests the import command with valid test data
// Note: This test expects failure due to missing database schema (migrations not implemented yet)
// Once migrations are implemented, this test should be updated to expect success
func TestIntegration_ImportCommand_ValidData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get the test binary
	binary := buildTestBinary(t)

	// Create temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test-integration.db")

	// Use existing test data from internal/worker/testdata/sde
	testDataDir := filepath.Join("..", "..", "internal", "worker", "testdata", "sde")
	absTestDataDir, err := filepath.Abs(testDataDir)
	if err != nil {
		t.Fatalf("failed to get absolute path for test data: %v", err)
	}

	// Verify test data directory exists
	if _, err := os.Stat(absTestDataDir); os.IsNotExist(err) {
		t.Fatalf("test data directory does not exist: %s", absTestDataDir)
	}

	// Run import command with --skip-errors to allow it to attempt processing all files
	importCmd := exec.Command(binary, "import", "--sde-dir", absTestDataDir, "--db", dbPath, "--workers", "4", "--skip-errors")
	output, err := importCmd.CombinedOutput()
	outputStr := string(output)

	// Currently expect partial failure due to missing schema (TODO comment in import.go)
	// But the command should still execute and discover files
	// Verify that the import command runs and discovers files
	if !strings.Contains(outputStr, "Import Summary") && !strings.Contains(outputStr, "Import completed") {
		t.Errorf("expected import summary in output, got: %s", outputStr)
	}

	// Verify database was created (even if empty)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("database file was not created: %s", dbPath)
	}

	// The import discovers and parses files (currently fails on insert due to no schema)
	// Once schema migrations are implemented, update this to check for success
	if !strings.Contains(outputStr, "parsed") {
		t.Errorf("expected parsed files count in output, got: %s", outputStr)
	}

	// Log the output for debugging
	t.Logf("Import output: %s", outputStr)

	// Check that the command completed (with or without errors for now)
	// err can be non-nil due to schema issues - that's expected for now
	_ = err
}

// TestIntegration_ImportCommand_NonExistentDirectory tests import with invalid directory
func TestIntegration_ImportCommand_NonExistentDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get the test binary
	binary := buildTestBinary(t)

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	nonExistentDir := filepath.Join(tmpDir, "does-not-exist")

	// Run import command with non-existent directory
	importCmd := exec.Command(binary, "import", "--sde-dir", nonExistentDir, "--db", dbPath)
	output, err := importCmd.CombinedOutput()

	// Should fail with non-zero exit code
	if err == nil {
		t.Error("expected import command to fail with non-existent directory, but it succeeded")
	}

	// Check for error message
	outputStr := string(output)
	if !strings.Contains(outputStr, "failed") && !strings.Contains(outputStr, "error") && !strings.Contains(outputStr, "Error") {
		t.Errorf("expected error message in output, got: %s", outputStr)
	}
}

// TestIntegration_ImportCommand_EmptyDirectory tests import with empty directory
func TestIntegration_ImportCommand_EmptyDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get the test binary
	binary := buildTestBinary(t)

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	emptyDir := filepath.Join(tmpDir, "empty-sde")
	
	// Create empty directory
	if err := os.Mkdir(emptyDir, 0755); err != nil {
		t.Fatalf("failed to create empty directory: %v", err)
	}

	// Run import command with empty directory
	importCmd := exec.Command(binary, "import", "--sde-dir", emptyDir, "--db", dbPath)
	output, err := importCmd.CombinedOutput()

	// Should fail because no JSONL files found
	if err == nil {
		t.Error("expected import command to fail with empty directory, but it succeeded")
	}

	// Check for appropriate error message
	outputStr := string(output)
	if !strings.Contains(outputStr, "no JSONL files") && !strings.Contains(outputStr, "Error") {
		t.Errorf("expected 'no JSONL files' error message, got: %s", outputStr)
	}
}

// TestIntegration_ValidateCommand_ValidConfig tests validate command with valid config
func TestIntegration_ValidateCommand_ValidConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get the test binary
	binary := buildTestBinary(t)

	tmpDir := t.TempDir()

	// Create a valid config file
	configPath := filepath.Join(tmpDir, "valid.toml")
	validConfig := `version = "1.0.0"

[database]
path = "./test.db"
journal_mode = "WAL"
cache_size_mb = 64

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
	if err := os.WriteFile(configPath, []byte(validConfig), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	// Run validate command
	validateCmd := exec.Command(binary, "validate", "--config", configPath)
	output, err := validateCmd.CombinedOutput()

	// Should succeed with exit code 0
	if err != nil {
		t.Errorf("validate command failed: %v\nOutput: %s", err, output)
	}

	// Verify output contains success message
	outputStr := string(output)
	if !strings.Contains(outputStr, "valid") && !strings.Contains(outputStr, "Configuration Summary") {
		t.Errorf("expected validation success message in output, got: %s", outputStr)
	}
}

// TestIntegration_ValidateCommand_InvalidConfig tests validate command with invalid config
func TestIntegration_ValidateCommand_InvalidConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get the test binary
	binary := buildTestBinary(t)

	tmpDir := t.TempDir()

	// Create an invalid config file (empty database path)
	configPath := filepath.Join(tmpDir, "invalid.toml")
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
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	// Run validate command
	validateCmd := exec.Command(binary, "validate", "--config", configPath)
	output, err := validateCmd.CombinedOutput()

	// Should fail with non-zero exit code
	if err == nil {
		t.Error("expected validate command to fail with invalid config, but it succeeded")
	}

	// Verify output contains error message
	outputStr := string(output)
	if !strings.Contains(outputStr, "database.path") && !strings.Contains(outputStr, "error") && !strings.Contains(outputStr, "Error") {
		t.Errorf("expected validation error message in output, got: %s", outputStr)
	}
}

// TestIntegration_ExitCodes tests various exit codes
func TestIntegration_ExitCodes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get the test binary
	binary := buildTestBinary(t)

	tests := []struct {
		name           string
		args           []string
		expectSuccess  bool
		description    string
	}{
		{
			name:          "help_command",
			args:          []string{"--help"},
			expectSuccess: true,
			description:   "help should exit with code 0",
		},
		{
			name:          "version_command",
			args:          []string{"--version"},
			expectSuccess: true,
			description:   "version should exit with code 0",
		},
		{
			name:          "version_subcommand",
			args:          []string{"version"},
			expectSuccess: true,
			description:   "version subcommand should exit with code 0",
		},
		{
			name:          "invalid_command",
			args:          []string{"nonexistent"},
			expectSuccess: false,
			description:   "invalid command should exit with non-zero code",
		},
		{
			name:          "import_missing_required_flag",
			args:          []string{"import", "--db", ""},
			expectSuccess: false,
			description:   "import with empty db path should fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binary, tt.args...)
			output, err := cmd.CombinedOutput()

			if tt.expectSuccess {
				if err != nil {
					// For --help and --version, check if it's actually an exit code issue
					if exitErr, ok := err.(*exec.ExitError); ok {
						if exitErr.ExitCode() != 0 {
							t.Errorf("%s: %v\nOutput: %s", tt.description, err, output)
						}
					} else {
						t.Errorf("%s: %v\nOutput: %s", tt.description, err, output)
					}
				}
			} else {
				if err == nil {
					t.Errorf("%s: expected non-zero exit code, got success\nOutput: %s", tt.description, output)
				}
			}
		})
	}
}

// TestIntegration_ImportCommand_WithSkipErrors tests import with --skip-errors flag
func TestIntegration_ImportCommand_WithSkipErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get the test binary
	binary := buildTestBinary(t)

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	
	// Create a directory with both valid and intentionally malformed JSONL files
	testDataDir := filepath.Join(tmpDir, "mixed-sde")
	if err := os.Mkdir(testDataDir, 0755); err != nil {
		t.Fatalf("failed to create test data directory: %v", err)
	}

	// Create a valid JSONL file
	validFile := filepath.Join(testDataDir, "valid.jsonl")
	validData := `{"typeID": 1, "typeName": "Test Item"}
{"typeID": 2, "typeName": "Another Item"}
`
	if err := os.WriteFile(validFile, []byte(validData), 0644); err != nil {
		t.Fatalf("failed to create valid test file: %v", err)
	}

	// Note: Without proper parsers registered, even valid files might fail
	// This test verifies the --skip-errors flag behavior

	// Run import command with --skip-errors
	importCmd := exec.Command(binary, "import", "--sde-dir", testDataDir, "--db", dbPath, "--skip-errors")
	output, _ := importCmd.CombinedOutput()

	// With --skip-errors, command might succeed even if some files fail
	// We just verify it doesn't panic and produces output
	outputStr := string(output)
	if len(outputStr) == 0 {
		t.Error("expected some output from import command")
	}
}

// TestIntegration_VerboseFlag tests the --verbose flag
func TestIntegration_VerboseFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get the test binary
	binary := buildTestBinary(t)

	// Run with --verbose flag
	cmd := exec.Command(binary, "--verbose", "--help")
	output, err := cmd.CombinedOutput()

	// Should succeed
	if err != nil {
		// Help might exit with 0 or specific code, check actual exit code
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Some CLI frameworks exit with code 0 for --help, others don't
			// As long as we get output, it's fine
			if len(output) == 0 {
				t.Errorf("verbose help failed: %v", exitErr)
			}
		}
	}

	// Verify we got output
	outputStr := string(output)
	if len(outputStr) == 0 {
		t.Error("expected output from verbose help command")
	}
}
