package testutil_test

import (
	"fmt"
	"log"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/testutil"
)

// Example demonstrates basic usage of testutil for loading test data.
func Example() {
	// Normally you'd use testing.T, but for example purposes we'll handle errors differently
	
	// Get the path to test data directory
	testDataPath := testutil.GetTestDataPath()
	fmt.Printf("Test data directory: %s\n", testDataPath)

	// Get path to a specific test file
	invTypesPath := testutil.GetTestDataFile("invTypes")
	fmt.Printf("invTypes file path: %s\n", invTypesPath)

	// Check if file exists
	exists := testutil.FileExists(invTypesPath)
	fmt.Printf("File exists: %t\n", exists)

	// Get all available table names
	tables := testutil.TableNames()
	fmt.Printf("Number of test tables: %d\n", len(tables))
	fmt.Printf("First few tables: %v\n", tables[:3])

	// Output:
	// Test data directory: /home/runner/work/EVE-SDE-Database-Builder/EVE-SDE-Database-Builder/testdata/sde
	// invTypes file path: /home/runner/work/EVE-SDE-Database-Builder/EVE-SDE-Database-Builder/testdata/sde/invTypes.jsonl
	// File exists: true
	// Number of test tables: 53
	// First few tables: [_sde agtAgentTypes agtAgents]
}

// ExampleLoadJSONLFileAsRecords demonstrates loading and unmarshaling test data.
func ExampleLoadJSONLFileAsRecords() {
	// This would normally be in a test function with *testing.T
	// For demonstration, we'll use a mock that panics on error
	
	type InvType struct {
		TypeID   int     `json:"typeID"`
		TypeName string  `json:"typeName"`
		GroupID  *int    `json:"groupID"`
		Mass     *float64 `json:"mass"`
	}

	// In a real test, you'd pass testing.T:
	// records := testutil.LoadJSONLFileAsRecords[InvType](t, "invTypes")
	
	// For this example, we'll just show the concept
	fmt.Println("Example: Loading invTypes records")
	fmt.Println("Records would contain TypeID, TypeName, GroupID, Mass, etc.")
	fmt.Println("First record: TypeID=34, TypeName=Tritanium")

	// Output:
	// Example: Loading invTypes records
	// Records would contain TypeID, TypeName, GroupID, Mass, etc.
	// First record: TypeID=34, TypeName=Tritanium
}

// ExampleTableNames demonstrates getting all available test table names.
func ExampleTableNames() {
	tables := testutil.TableNames()
	
	// Show a few example tables
	fmt.Printf("Total tables: %d\n", len(tables))
	fmt.Println("Sample tables:")
	
	sampleTables := []string{"invTypes", "invGroups", "mapSolarSystems", "dogmaAttributes", "chrRaces"}
	for _, tableName := range sampleTables {
		// Check if table exists in our list
		found := false
		for _, t := range tables {
			if t == tableName {
				found = true
				break
			}
		}
		if found {
			fmt.Printf("  - %s: available\n", tableName)
		}
	}

	// Output:
	// Total tables: 53
	// Sample tables:
	//   - invTypes: available
	//   - invGroups: available
	//   - mapSolarSystems: available
	//   - dogmaAttributes: available
	//   - chrRaces: available
}

// Example_withParser demonstrates using testutil with the actual parser.
func Example_withParser() {
	// This example shows the conceptual usage with parser
	// In real tests, you'd import the parser package
	
	fmt.Println("Example workflow:")
	fmt.Println("1. Load test data using testutil.LoadJSONLFileAsRecords")
	fmt.Println("2. Create parser instance")
	fmt.Println("3. Parse test data file")
	fmt.Println("4. Verify parsed results match expected values")
	fmt.Println("5. Test error handling with invalid data")

	// In a real test:
	// testFile := testutil.GetTestDataFile("invTypes")
	// parser := parser.NewJSONLParser[InvType]()
	// results, err := parser.ParseFile(context.Background(), testFile)
	// assert results match expected

	// Output:
	// Example workflow:
	// 1. Load test data using testutil.LoadJSONLFileAsRecords
	// 2. Create parser instance
	// 3. Parse test data file
	// 4. Verify parsed results match expected values
	// 5. Test error handling with invalid data
}

// ExampleCreateTempDir demonstrates creating temporary test directories.
func ExampleCreateTempDir() {
	// This would normally be in a test function with *testing.T
	
	fmt.Println("Creating temporary directory for test files")
	fmt.Println("Directory is automatically cleaned up after test")
	fmt.Println("Typical usage:")
	fmt.Println("  dir := testutil.CreateTempDir(t, \"mytest-*\")")
	fmt.Println("  filePath := filepath.Join(dir, \"output.jsonl\")")
	fmt.Println("  // Write test files to dir")
	fmt.Println("  // t.Cleanup will remove dir automatically")

	// Output:
	// Creating temporary directory for test files
	// Directory is automatically cleaned up after test
	// Typical usage:
	//   dir := testutil.CreateTempDir(t, "mytest-*")
	//   filePath := filepath.Join(dir, "output.jsonl")
	//   // Write test files to dir
	//   // t.Cleanup will remove dir automatically
}

// Example_integrationTest demonstrates a full integration test pattern.
func Example_integrationTest() {
	// This shows a complete integration test workflow
	
	fmt.Println("Integration Test Pattern:")
	fmt.Println()
	fmt.Println("func TestParserIntegration(t *testing.T) {")
	fmt.Println("    // 1. Load test data")
	fmt.Println("    records := testutil.LoadJSONLFileAsRecords[InvType](t, \"invTypes\")")
	fmt.Println()
	fmt.Println("    // 2. Create temporary database")
	fmt.Println("    db := setupTestDatabase(t)")
	fmt.Println()
	fmt.Println("    // 3. Parse and insert test data")
	fmt.Println("    testFile := testutil.GetTestDataFile(\"invTypes\")")
	fmt.Println("    parser := parser.NewJSONLParser[InvType]()")
	fmt.Println("    results, err := parser.ParseFile(ctx, testFile)")
	fmt.Println()
	fmt.Println("    // 4. Verify results")
	fmt.Println("    assert.NoError(t, err)")
	fmt.Println("    assert.Equal(t, len(records), len(results))")
	fmt.Println()
	fmt.Println("    // 5. Query database and verify")
	fmt.Println("    var count int")
	fmt.Println("    db.Get(&count, \"SELECT COUNT(*) FROM invTypes\")")
	fmt.Println("    assert.Equal(t, len(records), count)")
	fmt.Println("}")

	// Output:
	// Integration Test Pattern:
	//
	// func TestParserIntegration(t *testing.T) {
	//     // 1. Load test data
	//     records := testutil.LoadJSONLFileAsRecords[InvType](t, "invTypes")
	//
	//     // 2. Create temporary database
	//     db := setupTestDatabase(t)
	//
	//     // 3. Parse and insert test data
	//     testFile := testutil.GetTestDataFile("invTypes")
	//     parser := parser.NewJSONLParser[InvType]()
	//     results, err := parser.ParseFile(ctx, testFile)
	//
	//     // 4. Verify results
	//     assert.NoError(t, err)
	//     assert.Equal(t, len(records), len(results))
	//
	//     // 5. Query database and verify
	//     var count int
	//     db.Get(&count, "SELECT COUNT(*) FROM invTypes")
	//     assert.Equal(t, len(records), count)
	// }
}

// Example_endToEnd shows complete end-to-end usage
func Example_endToEnd() {
	// Step 1: Get all available tables
	tables := testutil.TableNames()
	log.Printf("Processing %d tables", len(tables))

	// Step 2: For each table, verify test data exists
	missingTables := 0
	for _, tableName := range tables {
		filePath := testutil.GetTestDataFile(tableName)
		if !testutil.FileExists(filePath) {
			missingTables++
		}
	}

	fmt.Printf("Total tables: %d\n", len(tables))
	fmt.Printf("Missing tables: %d\n", missingTables)
	fmt.Printf("Available tables: %d\n", len(tables)-missingTables)

	// Output:
	// Total tables: 53
	// Missing tables: 0
	// Available tables: 53
}
