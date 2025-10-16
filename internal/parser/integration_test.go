package parser_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/database"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"
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

// toRow converts InvType to database row format
func invTypeToRow(item interface{}) []interface{} {
	inv := item.(InvType)
	return []interface{}{
		inv.TypeID,
		inv.TypeName,
		inv.GroupID,
		inv.Description,
		inv.Mass,
		inv.Volume,
		inv.Capacity,
		inv.PortionSize,
		inv.RaceID,
		inv.BasePrice,
		inv.Published,
		inv.MarketGroupID,
		inv.IconID,
		inv.SoundID,
		inv.GraphicID,
	}
}

// toRow converts InvGroup to database row format
func invGroupToRow(item interface{}) []interface{} {
	grp := item.(InvGroup)
	return []interface{}{
		grp.GroupID,
		grp.CategoryID,
		grp.GroupName,
		grp.IconID,
		grp.UseBasePrice,
		grp.Anchored,
		grp.Anchorable,
		grp.FittableNonSingleton,
		grp.Published,
	}
}

// toRow converts Blueprint to database row format
func blueprintToRow(item interface{}) []interface{} {
	bp := item.(Blueprint)
	return []interface{}{
		bp.BlueprintTypeID,
		bp.MaxProductionLimit,
	}
}

// convertToRows is a helper function to convert parsed items to database rows
func convertToRows(items []interface{}, converter func(interface{}) []interface{}) [][]interface{} {
	rows := make([][]interface{}, len(items))
	for i, item := range items {
		rows[i] = converter(item)
	}
	return rows
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

// TestParseAndInsert_InvTypes tests E2E flow: Parse invTypes JSONL → Insert to SQLite → Verify count
func TestParseAndInsert_InvTypes(t *testing.T) {
	// Create test database with migrations
	db := database.NewTestDB(t)
	ctx := context.Background()

	// Create testdata directory and sample JSONL file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invTypes.jsonl")

	// Create sample invTypes JSONL data
	testData := `{"typeID":34,"typeName":"Tritanium","groupID":18,"description":"Tritanium is the primary building block.","mass":0.01,"volume":0.01,"capacity":null,"portionSize":1,"raceID":null,"basePrice":5.0,"published":1,"marketGroupID":1857,"iconID":22,"soundID":null,"graphicID":null}
{"typeID":35,"typeName":"Pyerite","groupID":18,"description":"Pyerite is a common mineral.","mass":0.01,"volume":0.01,"capacity":null,"portionSize":1,"raceID":null,"basePrice":6.0,"published":1,"marketGroupID":1857,"iconID":22,"soundID":null,"graphicID":null}
{"typeID":36,"typeName":"Mexallon","groupID":18,"description":"Mexallon is a commonly used industrial mineral.","mass":0.01,"volume":0.01,"capacity":null,"portionSize":1,"raceID":null,"basePrice":60.0,"published":1,"marketGroupID":1857,"iconID":22,"soundID":null,"graphicID":null}
`
	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create parser
	invTypesParser := parser.NewJSONLParser[InvType]("invTypes", []string{
		"typeID", "typeName", "groupID", "description", "mass", "volume",
		"capacity", "portionSize", "raceID", "basePrice", "published",
		"marketGroupID", "iconID", "soundID", "graphicID",
	})

	// Parse file
	items, err := invTypesParser.ParseFile(ctx, testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}

	// Convert to database rows
	rows := convertToRows(items, invTypeToRow)

	// Batch insert into database
	err = database.BatchInsert(ctx, db, "invTypes", invTypesParser.Columns(), rows, 1000)
	if err != nil {
		t.Fatalf("BatchInsert failed: %v", err)
	}

	// Verify row count
	count := getRowCount(t, db, "invTypes")
	if count != len(items) {
		t.Errorf("Expected %d rows in invTypes, got %d", len(items), count)
	}

	// Verify data integrity - check one specific record
	var typeName string
	var basePrice float64
	err = db.QueryRow("SELECT typeName, basePrice FROM invTypes WHERE typeID = 34").Scan(&typeName, &basePrice)
	if err != nil {
		t.Fatalf("Failed to query invTypes: %v", err)
	}

	if typeName != "Tritanium" {
		t.Errorf("Expected typeName 'Tritanium', got '%s'", typeName)
	}
	if basePrice != 5.0 {
		t.Errorf("Expected basePrice 5.0, got %f", basePrice)
	}
}

// TestParseAndInsert_InvGroups tests E2E flow for invGroups table
func TestParseAndInsert_InvGroups(t *testing.T) {
	// Create test database with migrations
	db := database.NewTestDB(t)
	ctx := context.Background()

	// Create testdata directory and sample JSONL file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invGroups.jsonl")

	// Create sample invGroups JSONL data
	testData := `{"groupID":18,"categoryID":4,"groupName":"Mineral","iconID":22,"useBasePrice":1,"anchored":0,"anchorable":0,"fittableNonSingleton":0,"published":1}
{"groupID":25,"categoryID":6,"groupName":"Frigate","iconID":3,"useBasePrice":0,"anchored":0,"anchorable":0,"fittableNonSingleton":0,"published":1}
{"groupID":26,"categoryID":6,"groupName":"Cruiser","iconID":4,"useBasePrice":0,"anchored":0,"anchorable":0,"fittableNonSingleton":0,"published":1}
{"groupID":27,"categoryID":6,"groupName":"Battleship","iconID":5,"useBasePrice":0,"anchored":0,"anchorable":0,"fittableNonSingleton":0,"published":1}
`
	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create parser
	invGroupsParser := parser.NewJSONLParser[InvGroup]("invGroups", []string{
		"groupID", "categoryID", "groupName", "iconID", "useBasePrice",
		"anchored", "anchorable", "fittableNonSingleton", "published",
	})

	// Parse file
	items, err := invGroupsParser.ParseFile(ctx, testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(items) != 4 {
		t.Errorf("Expected 4 items, got %d", len(items))
	}

	// Convert to database rows
	rows := convertToRows(items, invGroupToRow)

	// Batch insert into database
	err = database.BatchInsert(ctx, db, "invGroups", invGroupsParser.Columns(), rows, 1000)
	if err != nil {
		t.Fatalf("BatchInsert failed: %v", err)
	}

	// Verify row count
	count := getRowCount(t, db, "invGroups")
	if count != len(items) {
		t.Errorf("Expected %d rows in invGroups, got %d", len(items), count)
	}

	// Verify data integrity - check one specific record
	var groupName string
	var categoryID int
	err = db.QueryRow("SELECT groupName, categoryID FROM invGroups WHERE groupID = 18").Scan(&groupName, &categoryID)
	if err != nil {
		t.Fatalf("Failed to query invGroups: %v", err)
	}

	if groupName != "Mineral" {
		t.Errorf("Expected groupName 'Mineral', got '%s'", groupName)
	}
	if categoryID != 4 {
		t.Errorf("Expected categoryID 4, got %d", categoryID)
	}
}

// TestParseAndInsert_Blueprints tests E2E flow for blueprints table
func TestParseAndInsert_Blueprints(t *testing.T) {
	// Create test database with migrations
	db := database.NewTestDB(t)
	ctx := context.Background()

	// Create testdata directory and sample JSONL file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "industryBlueprints.jsonl")

	// Create sample blueprint JSONL data
	testData := `{"blueprintTypeID":1000001,"maxProductionLimit":10}
{"blueprintTypeID":1000002,"maxProductionLimit":5}
{"blueprintTypeID":1000003,"maxProductionLimit":null}
`
	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create parser
	blueprintParser := parser.NewJSONLParser[Blueprint]("industryBlueprints", []string{
		"blueprintTypeID", "maxProductionLimit",
	})

	// Parse file
	items, err := blueprintParser.ParseFile(ctx, testFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}

	// Convert to database rows
	rows := convertToRows(items, blueprintToRow)

	// Batch insert into database
	err = database.BatchInsert(ctx, db, "industryBlueprints", blueprintParser.Columns(), rows, 1000)
	if err != nil {
		t.Fatalf("BatchInsert failed: %v", err)
	}

	// Verify row count
	count := getRowCount(t, db, "industryBlueprints")
	if count != len(items) {
		t.Errorf("Expected %d rows in industryBlueprints, got %d", len(items), count)
	}

	// Verify data integrity - check one specific record
	var blueprintTypeID int
	var maxProductionLimit *int
	err = db.QueryRow("SELECT blueprintTypeID, maxProductionLimit FROM industryBlueprints WHERE blueprintTypeID = 1000001").Scan(&blueprintTypeID, &maxProductionLimit)
	if err != nil {
		t.Fatalf("Failed to query industryBlueprints: %v", err)
	}

	if blueprintTypeID != 1000001 {
		t.Errorf("Expected blueprintTypeID 1000001, got %d", blueprintTypeID)
	}
	if maxProductionLimit == nil || *maxProductionLimit != 10 {
		t.Errorf("Expected maxProductionLimit 10, got %v", maxProductionLimit)
	}

	// Verify null handling - check record with null maxProductionLimit
	err = db.QueryRow("SELECT blueprintTypeID, maxProductionLimit FROM industryBlueprints WHERE blueprintTypeID = 1000003").Scan(&blueprintTypeID, &maxProductionLimit)
	if err != nil {
		t.Fatalf("Failed to query industryBlueprints: %v", err)
	}

	if maxProductionLimit != nil {
		t.Errorf("Expected null maxProductionLimit, got %v", maxProductionLimit)
	}
}
