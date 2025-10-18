// Package golden provides utilities for golden file testing of parser output.
//
// Golden file tests verify that parser output remains stable and consistent
// across code changes by comparing actual output against expected "golden"
// reference files.
//
// # Overview
//
// This package implements the golden file testing pattern for EVE SDE parsers.
// Golden files are JSON-formatted reference files that contain the expected
// output of parser operations. Tests compare actual parser output against
// these golden files to detect unintended changes.
//
// # Features
//
//   - Automatic golden file generation with -update flag
//   - Byte-accurate JSON comparison
//   - Pretty-printed JSON output for readability
//   - Comprehensive test summary reporting
//   - Support for all 53 EVE SDE parsers
//
// # Usage
//
// ## Running Golden File Tests
//
// Run all golden file tests:
//
//	go test -v ./internal/parser/golden/...
//	make test-golden
//
// Run core parsers only (faster):
//
//	go test -v ./internal/parser/golden/... -run TestGoldenFiles_CoreParsers
//
// ## Updating Golden Files
//
// After intentional parser changes, regenerate golden files:
//
//	go test -v ./internal/parser/golden/... -update
//	make update-golden
//
// Always review changes before committing:
//
//	git diff testdata/golden/
//
// # Test Structure
//
// The package provides three main test functions:
//
//   - TestGoldenFiles_AllParsers: Tests all 53 registered parsers
//   - TestGoldenFiles_CoreParsers: Tests 7 core parsers (faster feedback)
//   - TestGoldenFiles_SingleParser: Example for focused testing
//
// # File Format
//
// Golden files are stored in testdata/golden/ with the following format:
//
//   - Filename: [table_name].golden.json
//   - Format: JSON array of parsed records
//   - Encoding: UTF-8
//   - Indentation: 2 spaces
//
// Example golden file (invCategories.golden.json):
//
//	[
//	  {
//	    "categoryID": 4,
//	    "categoryName": "Material",
//	    "iconID": null,
//	    "published": 1
//	  },
//	  {
//	    "categoryID": 6,
//	    "categoryName": "Ship",
//	    "iconID": null,
//	    "published": 1
//	  }
//	]
//
// # Core Functions
//
// ## CompareOrUpdate
//
// The main function for golden file testing:
//
//	func CompareOrUpdate(t *testing.T, tableName string, actual interface{}, update bool) bool
//
// Parameters:
//   - t: testing.T instance
//   - tableName: name of the table/parser being tested
//   - actual: actual parser output (will be JSON marshaled)
//   - update: if true, update the golden file instead of comparing
//
// Example usage:
//
//	results, err := parser.ParseFile(ctx, testFile)
//	if err != nil {
//	    t.Fatalf("ParseFile failed: %v", err)
//	}
//	golden.CompareOrUpdate(t, "invTypes", results, *updateFlag)
//
// ## Summary
//
// Track test results across multiple parsers:
//
//	summary := golden.NewSummary()
//	for tableName, parser := range allParsers {
//	    passed := golden.CompareOrUpdate(t, tableName, results, *updateFlag)
//	    summary.Record(passed, *updateFlag, !goldenExists)
//	}
//	t.Logf("%s", summary.String())
//
// # Workflow
//
// ## Development Workflow
//
//  1. Make parser changes
//  2. Run tests to see what changed:
//     go test -v ./internal/parser/golden/...
//  3. If changes are intentional, update golden files:
//     go test -v ./internal/parser/golden/... -update
//  4. Review changes:
//     git diff testdata/golden/
//  5. Commit both code and golden file updates:
//     git add internal/parser/ testdata/golden/
//     git commit -m "Update parser and golden files"
//
// ## CI/CD Integration
//
// In CI/CD pipelines, golden file tests run without -update flag:
//
//	- name: Run Golden File Tests
//	  run: go test -v ./internal/parser/golden/...
//
// If tests fail, developers must update golden files locally and commit.
//
// # Best Practices
//
// ## DO ✅
//
//   - Review golden file changes before committing
//   - Commit golden files with code changes
//   - Use descriptive commit messages explaining why output changed
//   - Run tests before updating golden files
//   - Keep test data minimal and focused
//
// ## DON'T ❌
//
//   - Blindly update golden files without review
//   - Commit generated files without verification
//   - Mix unrelated changes in one commit
//   - Update golden files in CI (only update locally)
//   - Use excessive test data
//
// # Troubleshooting
//
// ## Golden file test fails
//
//	--- FAIL: TestGoldenFiles_AllParsers/invTypes (0.00s)
//	    golden_test.go:45: Parser output mismatch for invTypes
//	        Golden file: testdata/golden/invTypes.golden.json
//	        Run with -update to update the golden file
//
// Solution:
//  1. Review the diff shown in test output
//  2. If change is intentional: go test -v ./internal/parser/golden/... -update
//  3. If change is not intentional: Fix your code
//
// ## Golden file missing
//
//	--- FAIL: TestGoldenFiles_AllParsers/newTable (0.00s)
//	    golden.go:85: Golden file does not exist: testdata/golden/newTable.golden.json
//	        Run with -update to create it
//
// Solution:
//
//	go test -v ./internal/parser/golden/... -update
//
// ## Test data file missing
//
//	--- SKIP: TestGoldenFiles_AllParsers/someTable (0.00s)
//	    golden_test.go:32: Test data file not found: testdata/sde/someTable.jsonl
//
// Solution:
//  1. Add test data file: testdata/sde/someTable.jsonl
//  2. Or skip the test if table is not yet implemented
//
// # See Also
//
//   - testdata/golden/README.md - Comprehensive golden file documentation
//   - internal/parser/parsers.go - Parser implementations
//   - internal/testutil/testutil.go - Test data utilities
//
// # References
//
//   - Golden File Testing Pattern: https://github.com/sebdah/goldie
//   - Go Testing: https://go.dev/doc/tutorial/add-a-test
//   - EVE SDE: https://developers.eveonline.com/
package golden
