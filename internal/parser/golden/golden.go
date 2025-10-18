// Package golden provides utilities for golden file testing of parser output.
// Golden file tests verify that parser output matches expected "golden" reference files.
// This package supports automatic update mode for regenerating golden files.
package golden

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// GetGoldenPath returns the absolute path to the golden files directory.
func GetGoldenPath() string {
	// Get the current source file directory
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get caller information")
	}

	// Navigate from internal/parser/golden to project root
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..", "..")
	goldenPath := filepath.Join(projectRoot, "testdata", "golden")

	return goldenPath
}

// GetGoldenFile returns the absolute path to a specific golden file.
func GetGoldenFile(tableName string) string {
	return filepath.Join(GetGoldenPath(), tableName+".golden.json")
}

// CompareOrUpdate compares the actual output with the golden file,
// or updates the golden file if update flag is set.
//
// Parameters:
//   - t: testing.T instance
//   - tableName: name of the table/parser being tested
//   - actual: actual parser output (will be JSON marshaled)
//   - update: if true, update the golden file instead of comparing
//
// Returns true if comparison passed or update was performed.
func CompareOrUpdate(t *testing.T, tableName string, actual interface{}, update bool) bool {
	t.Helper()

	goldenFile := GetGoldenFile(tableName)

	// Marshal actual output to pretty JSON
	actualJSON, err := json.MarshalIndent(actual, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal actual output for %s: %v", tableName, err)
		return false
	}

	if update {
		// Update mode: write actual output to golden file
		if err := os.MkdirAll(filepath.Dir(goldenFile), 0755); err != nil {
			t.Fatalf("failed to create golden directory: %v", err)
			return false
		}

		if err := os.WriteFile(goldenFile, actualJSON, 0644); err != nil {
			t.Fatalf("failed to write golden file %s: %v", goldenFile, err)
			return false
		}

		t.Logf("Updated golden file: %s", goldenFile)
		return true
	}

	// Compare mode: load golden file and compare
	expectedJSON, err := os.ReadFile(goldenFile)
	if err != nil {
		if os.IsNotExist(err) {
			t.Errorf("Golden file does not exist: %s\nRun with -update to create it", goldenFile)
			return false
		}
		t.Fatalf("failed to read golden file %s: %v", goldenFile, err)
		return false
	}

	// Compare JSON content
	if string(actualJSON) != string(expectedJSON) {
		t.Errorf("Parser output mismatch for %s\n"+
			"Golden file: %s\n"+
			"Run with -update to update the golden file\n"+
			"Expected:\n%s\n\nActual:\n%s",
			tableName, goldenFile, string(expectedJSON), string(actualJSON))
		return false
	}

	return true
}

// LoadGoldenFile loads and unmarshals a golden file into the provided type.
func LoadGoldenFile[T any](t *testing.T, tableName string) T {
	t.Helper()

	goldenFile := GetGoldenFile(tableName)
	data, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", goldenFile, err)
	}

	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal golden file %s: %v", goldenFile, err)
	}

	return result
}

// FileExists checks if a golden file exists.
func FileExists(tableName string) bool {
	goldenFile := GetGoldenFile(tableName)
	_, err := os.Stat(goldenFile)
	return err == nil
}

// Summary represents a summary of golden file test results.
type Summary struct {
	Total   int
	Passed  int
	Failed  int
	Updated int
	Missing int
}

// String returns a human-readable summary.
func (s Summary) String() string {
	return fmt.Sprintf(
		"Golden File Test Summary: Total=%d, Passed=%d, Failed=%d, Updated=%d, Missing=%d",
		s.Total, s.Passed, s.Failed, s.Updated, s.Missing,
	)
}

// NewSummary creates a new empty summary.
func NewSummary() *Summary {
	return &Summary{}
}

// Record records a test result in the summary.
func (s *Summary) Record(passed bool, updated bool, missing bool) {
	s.Total++
	if updated {
		s.Updated++
	} else if missing {
		s.Missing++
	} else if passed {
		s.Passed++
	} else {
		s.Failed++
	}
}
