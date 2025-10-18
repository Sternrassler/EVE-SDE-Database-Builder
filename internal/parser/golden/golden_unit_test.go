package golden

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetGoldenPath(t *testing.T) {
	path := GetGoldenPath()

	// Verify path is absolute
	if !filepath.IsAbs(path) {
		t.Errorf("GetGoldenPath() returned relative path: %s", path)
	}

	// Verify path ends with testdata/golden
	if !filepath.IsAbs(path) || filepath.Base(path) != "golden" {
		t.Errorf("GetGoldenPath() does not end with 'golden': %s", path)
	}

	parent := filepath.Base(filepath.Dir(path))
	if parent != "testdata" {
		t.Errorf("GetGoldenPath() parent is not 'testdata': %s", parent)
	}
}

func TestGetGoldenFile(t *testing.T) {
	tableName := "test_table"
	goldenFile := GetGoldenFile(tableName)

	// Verify path is absolute
	if !filepath.IsAbs(goldenFile) {
		t.Errorf("GetGoldenFile() returned relative path: %s", goldenFile)
	}

	// Verify file has correct name
	expectedName := "test_table.golden.json"
	if filepath.Base(goldenFile) != expectedName {
		t.Errorf("GetGoldenFile() = %s, want filename %s", goldenFile, expectedName)
	}
}

func TestFileExists(t *testing.T) {
	// Test with non-existent file
	if FileExists("nonexistent_table_xyz123") {
		t.Error("FileExists() returned true for non-existent file")
	}

	// Create a temporary golden file
	tableName := "temp_test"

	// Temporarily override GetGoldenPath for this test
	originalPath := GetGoldenPath()
	defer func() {
		// Note: Can't actually override GetGoldenPath easily without refactoring,
		// so we'll just create file in the real golden path for this test
	}()

	// Create a test file
	testFile := filepath.Join(originalPath, tableName+".golden.json")
	if err := os.MkdirAll(filepath.Dir(testFile), 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Write test file
	if err := os.WriteFile(testFile, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	defer func() {
		// Clean up
		_ = os.Remove(testFile)
	}()

	// Now it should exist
	if !FileExists(tableName) {
		t.Error("FileExists() returned false for existing file")
	}
}

func TestSummary(t *testing.T) {
	s := NewSummary()

	if s.Total != 0 || s.Passed != 0 || s.Failed != 0 {
		t.Error("NewSummary() should return zeroed summary")
	}

	// Test passed record
	s.Record(true, false, false)
	if s.Total != 1 || s.Passed != 1 {
		t.Errorf("Record(passed) failed: Total=%d, Passed=%d", s.Total, s.Passed)
	}

	// Test failed record
	s.Record(false, false, false)
	if s.Total != 2 || s.Failed != 1 {
		t.Errorf("Record(failed) failed: Total=%d, Failed=%d", s.Total, s.Failed)
	}

	// Test updated record
	s.Record(false, true, false)
	if s.Total != 3 || s.Updated != 1 {
		t.Errorf("Record(updated) failed: Total=%d, Updated=%d", s.Total, s.Updated)
	}

	// Test missing record
	s.Record(false, false, true)
	if s.Total != 4 || s.Missing != 1 {
		t.Errorf("Record(missing) failed: Total=%d, Missing=%d", s.Total, s.Missing)
	}

	// Test String method
	str := s.String()
	if str == "" {
		t.Error("Summary.String() returned empty string")
	}

	// Verify summary contains expected values
	expected := "Total=4"
	if !contains(str, expected) {
		t.Errorf("Summary.String() does not contain %s: %s", expected, str)
	}
}

func TestCompareOrUpdate_UpdateMode(t *testing.T) {
	// Temporarily change golden path for this test
	// Note: This is a limitation - we can't easily override GetGoldenPath
	// In a real scenario, we'd refactor to inject the path

	testData := map[string]interface{}{
		"id":   1,
		"name": "test",
	}

	// Create a mock testing.T to capture behavior
	mockT := &testing.T{}

	// Test update mode
	tableName := "test_update_table"
	result := CompareOrUpdate(mockT, tableName, testData, true)

	if !result {
		t.Error("CompareOrUpdate in update mode should return true")
	}

	// Verify file was created
	if !FileExists(tableName) {
		t.Error("CompareOrUpdate in update mode did not create golden file")
	}

	// Clean up
	goldenFile := GetGoldenFile(tableName)
	_ = os.Remove(goldenFile)
}

func TestCompareOrUpdate_CompareMode(t *testing.T) {
	// Create a temporary golden file
	tableName := "test_compare_table"
	testData := map[string]interface{}{
		"id":   1,
		"name": "test",
	}

	// First, create the golden file
	mockT1 := &testing.T{}
	CompareOrUpdate(mockT1, tableName, testData, true)

	// Now compare with same data
	mockT2 := &testing.T{}
	result := CompareOrUpdate(mockT2, tableName, testData, false)

	if !result {
		t.Error("CompareOrUpdate with matching data should return true")
	}

	// Clean up
	goldenFile := GetGoldenFile(tableName)
	_ = os.Remove(goldenFile)
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
