package worker_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/database"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/worker"
	"github.com/jmoiron/sqlx"
)

// InvType represents a simplified EVE SDE invTypes record for testing
type InvType struct {
	TypeID        int      `json:"typeID"`
	TypeName      string   `json:"typeName"`
	GroupID       *int     `json:"groupID"`
	Description   *string  `json:"description"`
	Mass          *float64 `json:"mass"`
	Volume        *float64 `json:"volume"`
	Capacity      *float64 `json:"capacity"`
	PortionSize   *int     `json:"portionSize"`
	RaceID        *int     `json:"raceID"`
	BasePrice     *float64 `json:"basePrice"`
	Published     *int     `json:"published"`
	MarketGroupID *int     `json:"marketGroupID"`
	IconID        *int     `json:"iconID"`
	SoundID       *int     `json:"soundID"`
	GraphicID     *int     `json:"graphicID"`
}

// InvGroup represents a simplified EVE SDE invGroups record for testing
type InvGroup struct {
	GroupID              int    `json:"groupID"`
	CategoryID           *int   `json:"categoryID"`
	GroupName            string `json:"groupName"`
	IconID               *int   `json:"iconID"`
	UseBasePrice         *int   `json:"useBasePrice"`
	Anchored             *int   `json:"anchored"`
	Anchorable           *int   `json:"anchorable"`
	FittableNonSingleton *int   `json:"fittableNonSingleton"`
	Published            *int   `json:"published"`
}

// Blueprint represents a simplified EVE SDE blueprint record for testing
type Blueprint struct {
	BlueprintTypeID    int  `json:"blueprintTypeID"`
	MaxProductionLimit *int `json:"maxProductionLimit"`
}

// getRowCount is a helper to count rows in a table
func getRowCount(t *testing.T, db *sqlx.DB, tableName string) int {
	t.Helper()
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM " + tableName).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count rows in %s: %v", tableName, err)
	}
	return count
}

// TestOrchestrator_ImportAll tests E2E parallel import of 10 JSONL files
func TestOrchestrator_ImportAll(t *testing.T) {
	// Create test database with migrations
	db := database.NewTestDB(t)
	ctx := context.Background()

	// Get the testdata directory path
	testdataDir := filepath.Join("testdata", "sde")

	// Verify testdata directory exists
	if _, err := os.Stat(testdataDir); os.IsNotExist(err) {
		t.Fatalf("testdata directory does not exist: %s", testdataDir)
	}

	// Create parsers for all test files
	parsers := make(map[string]parser.Parser)

	// InvTypes parsers (3 files)
	invTypesColumns := []string{
		"typeID", "typeName", "groupID", "description", "mass", "volume",
		"capacity", "portionSize", "raceID", "basePrice", "published",
		"marketGroupID", "iconID", "soundID", "graphicID",
	}
	parsers[filepath.Join(testdataDir, "invTypes_1.jsonl")] = parser.NewJSONLParser[InvType]("invTypes", invTypesColumns)
	parsers[filepath.Join(testdataDir, "invTypes_2.jsonl")] = parser.NewJSONLParser[InvType]("invTypes", invTypesColumns)
	parsers[filepath.Join(testdataDir, "invTypes_3.jsonl")] = parser.NewJSONLParser[InvType]("invTypes", invTypesColumns)

	// InvGroups parsers (3 files)
	invGroupsColumns := []string{
		"groupID", "categoryID", "groupName", "iconID", "useBasePrice",
		"anchored", "anchorable", "fittableNonSingleton", "published",
	}
	parsers[filepath.Join(testdataDir, "invGroups_1.jsonl")] = parser.NewJSONLParser[InvGroup]("invGroups", invGroupsColumns)
	parsers[filepath.Join(testdataDir, "invGroups_2.jsonl")] = parser.NewJSONLParser[InvGroup]("invGroups", invGroupsColumns)
	parsers[filepath.Join(testdataDir, "invGroups_3.jsonl")] = parser.NewJSONLParser[InvGroup]("invGroups", invGroupsColumns)

	// IndustryBlueprints parsers (4 files)
	blueprintColumns := []string{"blueprintTypeID", "maxProductionLimit"}
	parsers[filepath.Join(testdataDir, "industryBlueprints_1.jsonl")] = parser.NewJSONLParser[Blueprint]("industryBlueprints", blueprintColumns)
	parsers[filepath.Join(testdataDir, "industryBlueprints_2.jsonl")] = parser.NewJSONLParser[Blueprint]("industryBlueprints", blueprintColumns)
	parsers[filepath.Join(testdataDir, "industryBlueprints_3.jsonl")] = parser.NewJSONLParser[Blueprint]("industryBlueprints", blueprintColumns)
	parsers[filepath.Join(testdataDir, "industryBlueprints_4.jsonl")] = parser.NewJSONLParser[Blueprint]("industryBlueprints", blueprintColumns)

	// Create worker pool with 4 workers (for parallel processing)
	pool := worker.NewPool(4)

	// Create orchestrator
	orch := worker.NewOrchestrator(db, pool, parsers)

	// Execute import
	progress, err := orch.ImportAll(ctx, testdataDir)
	if err != nil {
		t.Fatalf("ImportAll failed: %v", err)
	}

	// Verify progress tracking
	parsed, inserted, failed, total := progress.GetProgress()
	if total != 10 {
		t.Errorf("Expected 10 total files, got %d", total)
	}
	if parsed != 10 {
		t.Errorf("Expected 10 parsed files, got %d", parsed)
	}
	if failed != 0 {
		t.Errorf("Expected 0 failed files, got %d", failed)
	}
	if inserted != 10 {
		t.Errorf("Expected 10 inserted files, got %d", inserted)
	}

	// Verify database row counts
	// invTypes: 3 + 2 + 2 = 7 rows
	invTypesCount := getRowCount(t, db, "invTypes")
	expectedInvTypes := 7
	if invTypesCount != expectedInvTypes {
		t.Errorf("Expected %d rows in invTypes, got %d", expectedInvTypes, invTypesCount)
	}

	// invGroups: 3 + 2 + 2 = 7 rows
	invGroupsCount := getRowCount(t, db, "invGroups")
	expectedInvGroups := 7
	if invGroupsCount != expectedInvGroups {
		t.Errorf("Expected %d rows in invGroups, got %d", expectedInvGroups, invGroupsCount)
	}

	// industryBlueprints: 3 + 2 + 2 + 2 = 9 rows
	blueprintsCount := getRowCount(t, db, "industryBlueprints")
	expectedBlueprints := 9
	if blueprintsCount != expectedBlueprints {
		t.Errorf("Expected %d rows in industryBlueprints, got %d", expectedBlueprints, blueprintsCount)
	}

	// Verify some sample data integrity
	var typeName string
	err = db.QueryRow("SELECT typeName FROM invTypes WHERE typeID = 34").Scan(&typeName)
	if err != nil {
		t.Fatalf("Failed to query invTypes: %v", err)
	}
	if typeName != "Tritanium" {
		t.Errorf("Expected typeName 'Tritanium', got '%s'", typeName)
	}

	var groupName string
	err = db.QueryRow("SELECT groupName FROM invGroups WHERE groupID = 18").Scan(&groupName)
	if err != nil {
		t.Fatalf("Failed to query invGroups: %v", err)
	}
	if groupName != "Mineral" {
		t.Errorf("Expected groupName 'Mineral', got '%s'", groupName)
	}

	var blueprintTypeID int
	err = db.QueryRow("SELECT blueprintTypeID FROM industryBlueprints WHERE blueprintTypeID = 1000001").Scan(&blueprintTypeID)
	if err != nil {
		t.Fatalf("Failed to query industryBlueprints: %v", err)
	}
	if blueprintTypeID != 1000001 {
		t.Errorf("Expected blueprintTypeID 1000001, got %d", blueprintTypeID)
	}
}

// TestOrchestrator_ImportAll_ParallelPerformance verifies parallel processing benefits
func TestOrchestrator_ImportAll_ParallelPerformance(t *testing.T) {
	// This test verifies that the orchestrator can handle multiple files
	// in parallel without errors. Performance benchmarks are in orchestrator_test.go.

	db := database.NewTestDB(t)
	ctx := context.Background()
	testdataDir := filepath.Join("testdata", "sde")

	// Create parsers
	parsers := make(map[string]parser.Parser)
	invTypesColumns := []string{
		"typeID", "typeName", "groupID", "description", "mass", "volume",
		"capacity", "portionSize", "raceID", "basePrice", "published",
		"marketGroupID", "iconID", "soundID", "graphicID",
	}
	parsers[filepath.Join(testdataDir, "invTypes_1.jsonl")] = parser.NewJSONLParser[InvType]("invTypes", invTypesColumns)
	parsers[filepath.Join(testdataDir, "invTypes_2.jsonl")] = parser.NewJSONLParser[InvType]("invTypes", invTypesColumns)
	parsers[filepath.Join(testdataDir, "invTypes_3.jsonl")] = parser.NewJSONLParser[InvType]("invTypes", invTypesColumns)

	invGroupsColumns := []string{
		"groupID", "categoryID", "groupName", "iconID", "useBasePrice",
		"anchored", "anchorable", "fittableNonSingleton", "published",
	}
	parsers[filepath.Join(testdataDir, "invGroups_1.jsonl")] = parser.NewJSONLParser[InvGroup]("invGroups", invGroupsColumns)
	parsers[filepath.Join(testdataDir, "invGroups_2.jsonl")] = parser.NewJSONLParser[InvGroup]("invGroups", invGroupsColumns)
	parsers[filepath.Join(testdataDir, "invGroups_3.jsonl")] = parser.NewJSONLParser[InvGroup]("invGroups", invGroupsColumns)

	blueprintColumns := []string{"blueprintTypeID", "maxProductionLimit"}
	parsers[filepath.Join(testdataDir, "industryBlueprints_1.jsonl")] = parser.NewJSONLParser[Blueprint]("industryBlueprints", blueprintColumns)
	parsers[filepath.Join(testdataDir, "industryBlueprints_2.jsonl")] = parser.NewJSONLParser[Blueprint]("industryBlueprints", blueprintColumns)
	parsers[filepath.Join(testdataDir, "industryBlueprints_3.jsonl")] = parser.NewJSONLParser[Blueprint]("industryBlueprints", blueprintColumns)
	parsers[filepath.Join(testdataDir, "industryBlueprints_4.jsonl")] = parser.NewJSONLParser[Blueprint]("industryBlueprints", blueprintColumns)

	// Test with 4 workers (parallel processing)
	pool := worker.NewPool(4)
	orch := worker.NewOrchestrator(db, pool, parsers)

	progress, err := orch.ImportAll(ctx, testdataDir)
	if err != nil {
		t.Fatalf("ImportAll with 4 workers failed: %v", err)
	}

	// Verify all files were processed
	parsed, _, failed, _ := progress.GetProgress()
	if parsed != 10 {
		t.Errorf("Expected 10 parsed files, got %d", parsed)
	}
	if failed != 0 {
		t.Errorf("Expected 0 failed files, got %d", failed)
	}

	// Verify progress tracker provides detailed metrics
	detailed := progress.GetProgressDetailed()
	if detailed.ParsedFiles != 10 {
		t.Errorf("Expected ParsedFiles=10, got %d", detailed.ParsedFiles)
	}
	if detailed.TotalFiles != 10 {
		t.Errorf("Expected TotalFiles=10, got %d", detailed.TotalFiles)
	}
	if detailed.InsertedRows != 23 { // 7 + 7 + 9 = 23 total rows
		t.Errorf("Expected InsertedRows=23, got %d", detailed.InsertedRows)
	}
	if detailed.PercentFiles != 100.0 {
		t.Errorf("Expected PercentFiles=100.0, got %f", detailed.PercentFiles)
	}
}
