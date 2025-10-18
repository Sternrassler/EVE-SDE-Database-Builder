// Package testutil provides shared testing utilities for EVE SDE Database Builder tests.
package testutil

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// GetTestDataPath returns the absolute path to the testdata directory.
// It works from any package in the project by finding the project root.
func GetTestDataPath() string {
	// Get the current source file directory
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get caller information")
	}

	// Navigate from internal/testutil to project root
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	testDataPath := filepath.Join(projectRoot, "testdata", "sde")

	return testDataPath
}

// GetTestDataFile returns the absolute path to a specific test data file.
func GetTestDataFile(tableName string) string {
	return filepath.Join(GetTestDataPath(), tableName+".jsonl")
}

// LoadJSONLFile loads a JSONL file from testdata and returns the lines as a slice of strings.
func LoadJSONLFile(t *testing.T, tableName string) []string {
	t.Helper()

	filePath := GetTestDataFile(tableName)
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("failed to open test data file %s: %v", filePath, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("error reading test data file %s: %v", filePath, err)
	}

	return lines
}

// LoadJSONLFileAsRecords loads a JSONL file and unmarshals each line into the provided type T.
// Returns a slice of unmarshaled records.
func LoadJSONLFileAsRecords[T any](t *testing.T, tableName string) []T {
	t.Helper()

	lines := LoadJSONLFile(t, tableName)
	records := make([]T, 0, len(lines))

	for i, line := range lines {
		var record T
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("failed to unmarshal line %d in %s: %v", i+1, tableName, err)
		}
		records = append(records, record)
	}

	return records
}

// FileExists checks if a file exists at the given path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// CreateTempDir creates a temporary directory for tests.
func CreateTempDir(t *testing.T, pattern string) string {
	t.Helper()

	dir, err := os.MkdirTemp("", pattern)
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	return dir
}

// WriteJSONLFile writes a slice of records as JSONL to a file.
func WriteJSONLFile[T any](t *testing.T, filePath string, records []T) {
	t.Helper()

	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create file %s: %v", filePath, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for i, record := range records {
		data, err := json.Marshal(record)
		if err != nil {
			t.Fatalf("failed to marshal record %d: %v", i, err)
		}

		if _, err := writer.Write(data); err != nil {
			t.Fatalf("failed to write record %d: %v", i, err)
		}

		if _, err := writer.WriteString("\n"); err != nil {
			t.Fatalf("failed to write newline after record %d: %v", i, err)
		}
	}
}

// TableNames returns all available SDE table names from testdata.
func TableNames() []string {
	return []string{
		"_sde",
		"agtAgentTypes",
		"agtAgents",
		"certCerts",
		"certMasteries",
		"chrAncestries",
		"chrAttributes",
		"chrBloodlines",
		"chrFactions",
		"chrNPCCharacters",
		"chrRaces",
		"contrabandTypes",
		"controlTowerResources",
		"crpActivities",
		"crpNPCCorporationDivisions",
		"crpNPCCorporations",
		"dbuffCollections",
		"dogmaAttributeCategories",
		"dogmaAttributes",
		"dogmaEffects",
		"dogmaTypeAttributes",
		"dogmaTypeEffects",
		"dogmaUnits",
		"dynamicItemAttributes",
		"eveGraphics",
		"eveIcons",
		"industryBlueprints",
		"invCategories",
		"invGroups",
		"invMarketGroups",
		"invMetaGroups",
		"invTypes",
		"mapAsteroidBelts",
		"mapConstellations",
		"mapLandmarks",
		"mapMoons",
		"mapPlanets",
		"mapRegions",
		"mapSolarSystems",
		"mapStargates",
		"mapStars",
		"planetResources",
		"planetSchematics",
		"skinLicenses",
		"skinMaterials",
		"skins",
		"sovereigntyUpgrades",
		"staOperations",
		"staServices",
		"staStations",
		"translationLanguages",
		"typeBonuses",
		"typeDogma",
	}
}
