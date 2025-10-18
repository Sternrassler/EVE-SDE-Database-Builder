package golden_test

import (
	"context"
	"flag"
	"testing"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser/golden"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/testutil"
)

// update flag for regenerating golden files
// Usage: go test -v ./internal/parser/golden/... -update
var update = flag.Bool("update", false, "update golden files")

// TestGoldenFiles_AllParsers tests all registered parsers against their golden files.
// This test verifies that parser output remains stable across changes.
//
// To update golden files after intentional changes:
//   go test -v ./internal/parser/golden/... -update
//
// To run the test:
//   go test -v ./internal/parser/golden/...
func TestGoldenFiles_AllParsers(t *testing.T) {
	// Get all registered parsers
	allParsers := parser.RegisterParsers()

	// Create a summary to track results
	summary := golden.NewSummary()

	// Test each parser
	for tableName, p := range allParsers {
		t.Run(tableName, func(t *testing.T) {
			// Check if test data file exists
			testDataFile := testutil.GetTestDataFile(tableName)
			if !testutil.FileExists(testDataFile) {
				t.Skipf("Test data file not found: %s", testDataFile)
				return
			}

			// Parse the test data file
			ctx := context.Background()
			results, err := p.ParseFile(ctx, testDataFile)
			if err != nil {
				t.Fatalf("ParseFile failed for %s: %v", tableName, err)
			}

			// Check if we got any results
			if len(results) == 0 {
				t.Logf("Warning: Parser %s returned 0 results from %s", tableName, testDataFile)
			}

			// Compare or update golden file
			goldenExists := golden.FileExists(tableName)
			passed := golden.CompareOrUpdate(t, tableName, results, *update)

			// Record result in summary
			summary.Record(passed, *update, !goldenExists && !*update)
		})
	}

	// Print summary
	t.Logf("\n%s", summary.String())

	// Fail if any tests failed (but not in update mode)
	if !*update && summary.Failed > 0 {
		t.Errorf("%d parser(s) failed golden file comparison", summary.Failed)
	}

	// Warn if golden files are missing (but not in update mode)
	if !*update && summary.Missing > 0 {
		t.Logf("Warning: %d golden file(s) missing - run with -update to create them", summary.Missing)
	}

	// Report updates
	if *update && summary.Updated > 0 {
		t.Logf("Updated %d golden file(s)", summary.Updated)
	}
}

// TestGoldenFiles_SingleParser demonstrates testing a single parser.
// This can be used as a template for focused testing during development.
func TestGoldenFiles_SingleParser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping single parser test in short mode")
	}

	tableName := "invCategories"
	testDataFile := testutil.GetTestDataFile(tableName)

	// Skip if test data doesn't exist
	if !testutil.FileExists(testDataFile) {
		t.Skipf("Test data file not found: %s", testDataFile)
	}

	// Parse the test data
	p := parser.InvCategoriesParser
	ctx := context.Background()
	results, err := p.ParseFile(ctx, testDataFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Verify we got some results
	if len(results) == 0 {
		t.Fatal("Expected at least one result")
	}

	// Compare or update golden file
	golden.CompareOrUpdate(t, tableName, results, *update)
}

// TestGoldenFiles_CoreParsers tests a subset of core parsers for faster feedback.
// This is useful for CI/CD pipelines where testing all parsers might be too slow.
func TestGoldenFiles_CoreParsers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping core parsers test in short mode")
	}

	// Core tables that should always work
	coreTables := []string{
		"invTypes",
		"invGroups",
		"invCategories",
		"dogmaAttributes",
		"dogmaEffects",
		"mapSolarSystems",
		"chrRaces",
	}

	allParsers := parser.RegisterParsers()
	summary := golden.NewSummary()

	for _, tableName := range coreTables {
		t.Run(tableName, func(t *testing.T) {
			p, ok := allParsers[tableName]
			if !ok {
				t.Fatalf("Parser not found: %s", tableName)
			}

			testDataFile := testutil.GetTestDataFile(tableName)
			if !testutil.FileExists(testDataFile) {
				t.Skipf("Test data file not found: %s", testDataFile)
				return
			}

			ctx := context.Background()
			results, err := p.ParseFile(ctx, testDataFile)
			if err != nil {
				t.Fatalf("ParseFile failed: %v", err)
			}

			goldenExists := golden.FileExists(tableName)
			passed := golden.CompareOrUpdate(t, tableName, results, *update)
			summary.Record(passed, *update, !goldenExists && !*update)
		})
	}

	t.Logf("\n%s", summary.String())
}

// BenchmarkGoldenFileComparison benchmarks the golden file comparison operation.
func BenchmarkGoldenFileComparison(b *testing.B) {
	tableName := "invCategories"
	testDataFile := testutil.GetTestDataFile(tableName)

	if !testutil.FileExists(testDataFile) {
		b.Skipf("Test data file not found: %s", testDataFile)
	}

	p := parser.InvCategoriesParser
	ctx := context.Background()
	results, err := p.ParseFile(ctx, testDataFile)
	if err != nil {
		b.Fatalf("ParseFile failed: %v", err)
	}

	// Only benchmark if golden file exists
	if !golden.FileExists(tableName) {
		b.Skip("Golden file not found - run with -update first")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a dummy testing.T to satisfy the interface
		t := &testing.T{}
		golden.CompareOrUpdate(t, tableName, results, false)
	}
}
