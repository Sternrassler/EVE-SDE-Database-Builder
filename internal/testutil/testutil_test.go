package testutil_test

import (
	"bufio"
	"os"
	"path/filepath"
	"testing"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/testutil"
)

func TestGetTestDataPath(t *testing.T) {
	path := testutil.GetTestDataPath()

	if path == "" {
		t.Fatal("GetTestDataPath returned empty string")
	}

	if !filepath.IsAbs(path) {
		t.Errorf("GetTestDataPath should return absolute path, got: %s", path)
	}

	if !testutil.FileExists(path) {
		t.Errorf("testdata path does not exist: %s", path)
	}
}

func TestGetTestDataFile(t *testing.T) {
	filePath := testutil.GetTestDataFile("invTypes")

	if filePath == "" {
		t.Fatal("GetTestDataFile returned empty string")
	}

	if !filepath.IsAbs(filePath) {
		t.Errorf("GetTestDataFile should return absolute path, got: %s", filePath)
	}

	expectedSuffix := "invTypes.jsonl"
	if filepath.Base(filePath) != expectedSuffix {
		t.Errorf("expected file name to be %s, got: %s", expectedSuffix, filepath.Base(filePath))
	}
}

func TestLoadJSONLFile(t *testing.T) {
	lines := testutil.LoadJSONLFile(t, "invTypes")

	if len(lines) == 0 {
		t.Fatal("LoadJSONLFile returned no lines")
	}

	// Check that we got at least 3 lines (our test data has 3)
	if len(lines) < 3 {
		t.Errorf("expected at least 3 lines, got %d", len(lines))
	}

	// Check that each line is not empty
	for i, line := range lines {
		if line == "" {
			t.Errorf("line %d is empty", i)
		}
	}
}

func TestLoadJSONLFileAsRecords(t *testing.T) {
	type InvType struct {
		TypeID   int      `json:"typeID"`
		TypeName string   `json:"typeName"`
		GroupID  *int     `json:"groupID"`
		Mass     *float64 `json:"mass"`
	}

	records := testutil.LoadJSONLFileAsRecords[InvType](t, "invTypes")

	if len(records) == 0 {
		t.Fatal("LoadJSONLFileAsRecords returned no records")
	}

	// Check that we got at least 3 records
	if len(records) < 3 {
		t.Errorf("expected at least 3 records, got %d", len(records))
	}

	// Check first record
	if records[0].TypeID != 34 {
		t.Errorf("expected first record TypeID to be 34, got %d", records[0].TypeID)
	}

	if records[0].TypeName != "Tritanium" {
		t.Errorf("expected first record TypeName to be Tritanium, got %s", records[0].TypeName)
	}
}

func TestFileExists(t *testing.T) {
	existingFile := testutil.GetTestDataFile("invTypes")
	if !testutil.FileExists(existingFile) {
		t.Errorf("FileExists should return true for existing file: %s", existingFile)
	}

	nonExistingFile := "/path/to/nonexistent/file.jsonl"
	if testutil.FileExists(nonExistingFile) {
		t.Errorf("FileExists should return false for non-existing file: %s", nonExistingFile)
	}
}

func TestCreateTempDir(t *testing.T) {
	dir := testutil.CreateTempDir(t, "test-*")

	if dir == "" {
		t.Fatal("CreateTempDir returned empty string")
	}

	if !testutil.FileExists(dir) {
		t.Errorf("temp directory does not exist: %s", dir)
	}

	// The directory should be cleaned up automatically by t.Cleanup
}

func TestWriteJSONLFile(t *testing.T) {
	type TestRecord struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	records := []TestRecord{
		{ID: 1, Name: "Test 1"},
		{ID: 2, Name: "Test 2"},
		{ID: 3, Name: "Test 3"},
	}

	dir := testutil.CreateTempDir(t, "write-test-*")
	filePath := filepath.Join(dir, "test.jsonl")

	testutil.WriteJSONLFile(t, filePath, records)

	if !testutil.FileExists(filePath) {
		t.Errorf("file was not created: %s", filePath)
	}

	// Read back and verify by opening the file directly
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("failed to open written file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}

	if lineCount != len(records) {
		t.Errorf("expected %d lines, got %d", len(records), lineCount)
	}
}

func TestTableNames(t *testing.T) {
	tables := testutil.TableNames()

	if len(tables) == 0 {
		t.Fatal("TableNames returned empty slice")
	}

	// Check that we have at least 50 tables (we have 51 including _sde)
	if len(tables) < 50 {
		t.Errorf("expected at least 50 tables, got %d", len(tables))
	}

	// Check for a few expected tables
	expectedTables := []string{"invTypes", "invGroups", "mapSolarSystems", "dogmaAttributes"}
	for _, expected := range expectedTables {
		found := false
		for _, table := range tables {
			if table == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected table %s not found in TableNames", expected)
		}
	}
}
