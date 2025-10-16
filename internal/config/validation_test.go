package config

import (
	"runtime"
	"testing"
)

// TestValidationValidConfig tests that a completely valid configuration passes validation
// This test covers the acceptance criterion: "Gültige Config → kein Fehler"
func TestValidationValidConfig(t *testing.T) {
	cfg := DefaultConfig()

	err := cfg.Validate()
	if err != nil {
		t.Errorf("valid default config should pass validation, got error: %v", err)
	}

	// Verify that workers=0 was auto-detected to NumCPU()
	expectedWorkers := runtime.NumCPU()
	if cfg.Import.Workers != expectedWorkers {
		t.Errorf("expected workers to be auto-detected to %d (NumCPU), got %d",
			expectedWorkers, cfg.Import.Workers)
	}
}

// TestValidationSDEPathNotExists tests handling of non-existent SDE paths
// This is an optional test as per issue requirements
// Note: Current implementation only validates that SDEPath is not empty,
// it does NOT check if the path exists on the filesystem
func TestValidationSDEPathNotExists(t *testing.T) {
	t.Run("Empty SDEPath fails validation", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Import.SDEPath = ""

		err := cfg.Validate()
		if err == nil {
			t.Fatal("expected validation error for empty SDE path")
		}

		expectedMsg := "import.sde_path is required"
		if err.Error() != expectedMsg {
			t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("Non-existent path passes validation (no filesystem check)", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Import.SDEPath = "/this/path/does/not/exist/anywhere"

		// Current implementation does NOT validate filesystem existence
		// This test documents that behavior
		err := cfg.Validate()
		if err != nil {
			t.Errorf("validation should not check filesystem existence, got unexpected error: %v", err)
		}
	})
}
