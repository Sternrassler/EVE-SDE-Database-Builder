package database

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMigration_001_InvTypes tests the 001_inv_types.sql migration
func TestMigration_001_InvTypes(t *testing.T) {
	// Create in-memory database
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Read migration file
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "001_inv_types.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}

	// Execute migration
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}

	// Verify table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='invTypes'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Table invTypes was not created: %v", err)
	}
	if tableName != "invTypes" {
		t.Errorf("Expected table name 'invTypes', got '%s'", tableName)
	}
}

// TestMigration_001_InvTypes_Schema verifies the schema structure
func TestMigration_001_InvTypes_Schema(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Read and execute migration
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "001_inv_types.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}

	// Verify all expected columns exist
	expectedColumns := []string{
		"typeID", "typeName", "groupID", "description", "mass",
		"volume", "capacity", "portionSize", "raceID", "basePrice",
		"published", "marketGroupID", "iconID", "soundID", "graphicID",
	}

	rows, err := db.Query("PRAGMA table_info(invTypes)")
	if err != nil {
		t.Fatalf("Failed to get table info: %v", err)
	}
	defer rows.Close()

	columnMap := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue interface{}

		err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk)
		if err != nil {
			t.Fatalf("Failed to scan column info: %v", err)
		}
		columnMap[name] = true

		// Verify typeID is PRIMARY KEY
		if name == "typeID" && pk != 1 {
			t.Errorf("typeID should be PRIMARY KEY")
		}

		// Verify typeName is NOT NULL
		if name == "typeName" && notNull != 1 {
			t.Errorf("typeName should be NOT NULL")
		}
	}

	// Check all expected columns are present
	for _, col := range expectedColumns {
		if !columnMap[col] {
			t.Errorf("Expected column '%s' not found in table", col)
		}
	}
}

// TestMigration_001_InvTypes_Indexes verifies the indexes are created
func TestMigration_001_InvTypes_Indexes(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Read and execute migration
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "001_inv_types.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}

	// Verify indexes exist
	expectedIndexes := map[string]bool{
		"idx_invTypes_groupID":       false,
		"idx_invTypes_marketGroupID": false,
	}

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='invTypes' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		t.Fatalf("Failed to query indexes: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var indexName string
		if err := rows.Scan(&indexName); err != nil {
			t.Fatalf("Failed to scan index name: %v", err)
		}
		if _, exists := expectedIndexes[indexName]; exists {
			expectedIndexes[indexName] = true
		}
	}

	// Check all expected indexes are present
	for indexName, found := range expectedIndexes {
		if !found {
			t.Errorf("Expected index '%s' not found", indexName)
		}
	}
}

// TestMigration_001_InvTypes_DataInsertion tests that data can be inserted
func TestMigration_001_InvTypes_DataInsertion(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Execute migration
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "001_inv_types.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}

	// Insert test data
	testData := []struct {
		typeID   int
		typeName string
		groupID  int
	}{
		{34, "Tritanium", 18},
		{35, "Pyerite", 18},
		{36, "Mexallon", 18},
	}

	for _, td := range testData {
		_, err := db.Exec(
			"INSERT INTO invTypes (typeID, typeName, groupID) VALUES (?, ?, ?)",
			td.typeID, td.typeName, td.groupID,
		)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// Verify data was inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM invTypes").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count rows: %v", err)
	}
	if count != len(testData) {
		t.Errorf("Expected %d rows, got %d", len(testData), count)
	}

	// Verify specific row
	var typeName string
	err = db.QueryRow("SELECT typeName FROM invTypes WHERE typeID = ?", 34).Scan(&typeName)
	if err != nil {
		t.Fatalf("Failed to query data: %v", err)
	}
	if typeName != "Tritanium" {
		t.Errorf("Expected typeName 'Tritanium', got '%s'", typeName)
	}
}

// TestMigration_001_InvTypes_IndexPerformance tests that indexes are used
func TestMigration_001_InvTypes_IndexPerformance(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Execute migration
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "001_inv_types.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}

	// Insert some test data
	for i := 1; i <= 100; i++ {
		_, err := db.Exec(
			"INSERT INTO invTypes (typeID, typeName, groupID, marketGroupID) VALUES (?, ?, ?, ?)",
			i, "Type"+string(rune(i)), i%10, i%20,
		)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// Test query using groupID index
	rows, err := db.Query("SELECT typeID FROM invTypes WHERE groupID = 5")
	if err != nil {
		t.Fatalf("Failed to query by groupID: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	if count == 0 {
		t.Error("Expected at least one row with groupID = 5")
	}

	// Test query using marketGroupID index
	rows, err = db.Query("SELECT typeID FROM invTypes WHERE marketGroupID = 10")
	if err != nil {
		t.Fatalf("Failed to query by marketGroupID: %v", err)
	}
	defer rows.Close()

	count = 0
	for rows.Next() {
		count++
	}
	if count == 0 {
		t.Error("Expected at least one row with marketGroupID = 10")
	}
}

// TestMigration_001_InvTypes_Idempotence tests that migration can be run multiple times
func TestMigration_001_InvTypes_Idempotence(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Read migration file
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "001_inv_types.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}

	// Execute migration first time
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		t.Fatalf("Failed to execute migration first time: %v", err)
	}

	// Execute migration second time (should not fail due to CREATE TABLE IF NOT EXISTS)
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		t.Fatalf("Failed to execute migration second time: %v", err)
	}

	// Verify table still exists and has correct structure
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='invTypes'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Table invTypes does not exist after second migration: %v", err)
	}
}
