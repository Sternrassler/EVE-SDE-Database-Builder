package database

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/jmoiron/sqlx"
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

// TestMigration_002_InvGroups tests the 002_inv_groups.sql migration
func TestMigration_002_InvGroups(t *testing.T) {
	// Create in-memory database
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Read migration file
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "002_inv_groups.sql")
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
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='invGroups'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Table invGroups was not created: %v", err)
	}
	if tableName != "invGroups" {
		t.Errorf("Expected table name 'invGroups', got '%s'", tableName)
	}
}

// TestMigration_002_InvGroups_Schema verifies the schema structure
func TestMigration_002_InvGroups_Schema(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Read and execute migration
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "002_inv_groups.sql")
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
		"groupID", "categoryID", "groupName", "iconID", "useBasePrice",
		"anchored", "anchorable", "fittableNonSingleton", "published",
	}

	rows, err := db.Query("PRAGMA table_info(invGroups)")
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

		// Verify groupID is PRIMARY KEY
		if name == "groupID" && pk != 1 {
			t.Errorf("groupID should be PRIMARY KEY")
		}

		// Verify groupName is NOT NULL
		if name == "groupName" && notNull != 1 {
			t.Errorf("groupName should be NOT NULL")
		}
	}

	// Check all expected columns are present
	for _, col := range expectedColumns {
		if !columnMap[col] {
			t.Errorf("Expected column '%s' not found in table", col)
		}
	}
}

// TestMigration_002_InvGroups_Indexes verifies the indexes are created
func TestMigration_002_InvGroups_Indexes(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Read and execute migration
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "002_inv_groups.sql")
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
		"idx_invGroups_categoryID": false,
	}

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='invGroups' AND name NOT LIKE 'sqlite_%'")
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

// TestMigration_002_InvGroups_DataInsertion tests that data can be inserted
func TestMigration_002_InvGroups_DataInsertion(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Execute migration
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "002_inv_groups.sql")
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
		groupID    int
		groupName  string
		categoryID int
	}{
		{18, "Mineral", 4},
		{25, "Frigate", 6},
		{420, "Destroyer", 6},
	}

	for _, td := range testData {
		_, err := db.Exec(
			"INSERT INTO invGroups (groupID, groupName, categoryID) VALUES (?, ?, ?)",
			td.groupID, td.groupName, td.categoryID,
		)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// Verify data was inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM invGroups").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count rows: %v", err)
	}
	if count != len(testData) {
		t.Errorf("Expected %d rows, got %d", len(testData), count)
	}

	// Verify specific row
	var groupName string
	err = db.QueryRow("SELECT groupName FROM invGroups WHERE groupID = ?", 18).Scan(&groupName)
	if err != nil {
		t.Fatalf("Failed to query data: %v", err)
	}
	if groupName != "Mineral" {
		t.Errorf("Expected groupName 'Mineral', got '%s'", groupName)
	}
}

// TestMigration_002_InvGroups_IndexPerformance tests that indexes are used
func TestMigration_002_InvGroups_IndexPerformance(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Execute migration
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "002_inv_groups.sql")
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
			"INSERT INTO invGroups (groupID, groupName, categoryID) VALUES (?, ?, ?)",
			i, "Group"+string(rune(i)), i%10,
		)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// Test query using categoryID index
	rows, err := db.Query("SELECT groupID FROM invGroups WHERE categoryID = 5")
	if err != nil {
		t.Fatalf("Failed to query by categoryID: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	if count == 0 {
		t.Error("Expected at least one row with categoryID = 5")
	}
}

// TestMigration_002_InvGroups_Idempotence tests that migration can be run multiple times
func TestMigration_002_InvGroups_Idempotence(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Read migration file
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "002_inv_groups.sql")
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
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='invGroups'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Table invGroups does not exist after second migration: %v", err)
	}
}

// TestMigration_003_Blueprints tests the 003_blueprints.sql migration
func TestMigration_003_Blueprints(t *testing.T) {
	// Create in-memory database
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Read migration file
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "003_blueprints.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}

	// Execute migration
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}

	// Verify all tables exist
	expectedTables := []string{
		"industryBlueprints",
		"industryActivities",
		"industryActivityMaterials",
		"industryActivityProducts",
	}

	for _, table := range expectedTables {
		var tableName string
		err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&tableName)
		if err != nil {
			t.Fatalf("Table %s was not created: %v", table, err)
		}
		if tableName != table {
			t.Errorf("Expected table name '%s', got '%s'", table, tableName)
		}
	}
}

// TestMigration_003_Blueprints_Schema verifies the schema structure
func TestMigration_003_Blueprints_Schema(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Read and execute migration
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "003_blueprints.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}

	// Verify industryBlueprints schema
	t.Run("industryBlueprints", func(t *testing.T) {
		expectedColumns := []string{"blueprintTypeID", "maxProductionLimit"}
		rows, err := db.Query("PRAGMA table_info(industryBlueprints)")
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

			// Verify blueprintTypeID is PRIMARY KEY
			if name == "blueprintTypeID" && pk != 1 {
				t.Errorf("blueprintTypeID should be PRIMARY KEY")
			}
		}

		for _, col := range expectedColumns {
			if !columnMap[col] {
				t.Errorf("Expected column '%s' not found in table", col)
			}
		}
	})

	// Verify industryActivities schema
	t.Run("industryActivities", func(t *testing.T) {
		expectedColumns := []string{"blueprintTypeID", "activityID", "time"}
		rows, err := db.Query("PRAGMA table_info(industryActivities)")
		if err != nil {
			t.Fatalf("Failed to get table info: %v", err)
		}
		defer rows.Close()

		columnMap := make(map[string]bool)
		pkCount := 0
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

			// Count PRIMARY KEY columns
			if pk > 0 {
				pkCount++
			}
		}

		for _, col := range expectedColumns {
			if !columnMap[col] {
				t.Errorf("Expected column '%s' not found in table", col)
			}
		}

		// Verify composite PRIMARY KEY (blueprintTypeID, activityID)
		if pkCount != 2 {
			t.Errorf("Expected composite PRIMARY KEY with 2 columns, got %d", pkCount)
		}
	})

	// Verify industryActivityMaterials schema
	t.Run("industryActivityMaterials", func(t *testing.T) {
		expectedColumns := []string{"blueprintTypeID", "activityID", "materialTypeID", "quantity"}
		rows, err := db.Query("PRAGMA table_info(industryActivityMaterials)")
		if err != nil {
			t.Fatalf("Failed to get table info: %v", err)
		}
		defer rows.Close()

		columnMap := make(map[string]bool)
		pkCount := 0
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

			// Count PRIMARY KEY columns
			if pk > 0 {
				pkCount++
			}
		}

		for _, col := range expectedColumns {
			if !columnMap[col] {
				t.Errorf("Expected column '%s' not found in table", col)
			}
		}

		// Verify composite PRIMARY KEY (blueprintTypeID, activityID, materialTypeID)
		if pkCount != 3 {
			t.Errorf("Expected composite PRIMARY KEY with 3 columns, got %d", pkCount)
		}
	})

	// Verify industryActivityProducts schema
	t.Run("industryActivityProducts", func(t *testing.T) {
		expectedColumns := []string{"blueprintTypeID", "activityID", "productTypeID", "quantity"}
		rows, err := db.Query("PRAGMA table_info(industryActivityProducts)")
		if err != nil {
			t.Fatalf("Failed to get table info: %v", err)
		}
		defer rows.Close()

		columnMap := make(map[string]bool)
		pkCount := 0
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

			// Count PRIMARY KEY columns
			if pk > 0 {
				pkCount++
			}
		}

		for _, col := range expectedColumns {
			if !columnMap[col] {
				t.Errorf("Expected column '%s' not found in table", col)
			}
		}

		// Verify composite PRIMARY KEY (blueprintTypeID, activityID, productTypeID)
		if pkCount != 3 {
			t.Errorf("Expected composite PRIMARY KEY with 3 columns, got %d", pkCount)
		}
	})
}

// TestMigration_003_Blueprints_Indexes verifies the indexes are created
func TestMigration_003_Blueprints_Indexes(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Read and execute migration
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "003_blueprints.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}

	// Test indexes for industryActivities
	t.Run("industryActivities_indexes", func(t *testing.T) {
		expectedIndexes := map[string]bool{
			"idx_industryActivities_blueprintTypeID": false,
			"idx_industryActivities_activityID":      false,
		}

		rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='industryActivities' AND name NOT LIKE 'sqlite_%'")
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

		for indexName, found := range expectedIndexes {
			if !found {
				t.Errorf("Expected index '%s' not found", indexName)
			}
		}
	})

	// Test indexes for industryActivityMaterials
	t.Run("industryActivityMaterials_indexes", func(t *testing.T) {
		expectedIndexes := map[string]bool{
			"idx_industryActivityMaterials_blueprintTypeID": false,
			"idx_industryActivityMaterials_materialTypeID":  false,
		}

		rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='industryActivityMaterials' AND name NOT LIKE 'sqlite_%'")
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

		for indexName, found := range expectedIndexes {
			if !found {
				t.Errorf("Expected index '%s' not found", indexName)
			}
		}
	})

	// Test indexes for industryActivityProducts
	t.Run("industryActivityProducts_indexes", func(t *testing.T) {
		expectedIndexes := map[string]bool{
			"idx_industryActivityProducts_blueprintTypeID": false,
			"idx_industryActivityProducts_productTypeID":   false,
		}

		rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='industryActivityProducts' AND name NOT LIKE 'sqlite_%'")
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

		for indexName, found := range expectedIndexes {
			if !found {
				t.Errorf("Expected index '%s' not found", indexName)
			}
		}
	})
}

// TestMigration_003_Blueprints_DataInsertion tests that data can be inserted
func TestMigration_003_Blueprints_DataInsertion(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Execute migration
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "003_blueprints.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}

	// Insert test data into industryBlueprints
	_, err = db.Exec(
		"INSERT INTO industryBlueprints (blueprintTypeID, maxProductionLimit) VALUES (?, ?)",
		950, 1,
	)
	if err != nil {
		t.Fatalf("Failed to insert into industryBlueprints: %v", err)
	}

	// Insert test data into industryActivities
	_, err = db.Exec(
		"INSERT INTO industryActivities (blueprintTypeID, activityID, time) VALUES (?, ?, ?)",
		950, 1, 12000,
	)
	if err != nil {
		t.Fatalf("Failed to insert into industryActivities: %v", err)
	}

	// Insert test data into industryActivityMaterials
	_, err = db.Exec(
		"INSERT INTO industryActivityMaterials (blueprintTypeID, activityID, materialTypeID, quantity) VALUES (?, ?, ?, ?)",
		950, 1, 34, 1000,
	)
	if err != nil {
		t.Fatalf("Failed to insert into industryActivityMaterials: %v", err)
	}

	// Insert test data into industryActivityProducts
	_, err = db.Exec(
		"INSERT INTO industryActivityProducts (blueprintTypeID, activityID, productTypeID, quantity) VALUES (?, ?, ?, ?)",
		950, 1, 949, 1,
	)
	if err != nil {
		t.Fatalf("Failed to insert into industryActivityProducts: %v", err)
	}

	// Verify data was inserted into industryBlueprints
	var maxLimit int
	err = db.QueryRow("SELECT maxProductionLimit FROM industryBlueprints WHERE blueprintTypeID = ?", 950).Scan(&maxLimit)
	if err != nil {
		t.Fatalf("Failed to query industryBlueprints: %v", err)
	}
	if maxLimit != 1 {
		t.Errorf("Expected maxProductionLimit 1, got %d", maxLimit)
	}

	// Verify data was inserted into industryActivities
	var time int
	err = db.QueryRow("SELECT time FROM industryActivities WHERE blueprintTypeID = ? AND activityID = ?", 950, 1).Scan(&time)
	if err != nil {
		t.Fatalf("Failed to query industryActivities: %v", err)
	}
	if time != 12000 {
		t.Errorf("Expected time 12000, got %d", time)
	}

	// Verify data was inserted into industryActivityMaterials
	var quantity int
	err = db.QueryRow("SELECT quantity FROM industryActivityMaterials WHERE blueprintTypeID = ? AND activityID = ? AND materialTypeID = ?", 950, 1, 34).Scan(&quantity)
	if err != nil {
		t.Fatalf("Failed to query industryActivityMaterials: %v", err)
	}
	if quantity != 1000 {
		t.Errorf("Expected quantity 1000, got %d", quantity)
	}

	// Verify data was inserted into industryActivityProducts
	var productQuantity int
	err = db.QueryRow("SELECT quantity FROM industryActivityProducts WHERE blueprintTypeID = ? AND activityID = ? AND productTypeID = ?", 950, 1, 949).Scan(&productQuantity)
	if err != nil {
		t.Fatalf("Failed to query industryActivityProducts: %v", err)
	}
	if productQuantity != 1 {
		t.Errorf("Expected quantity 1, got %d", productQuantity)
	}
}

// TestMigration_003_Blueprints_IndexPerformance tests that indexes are used
func TestMigration_003_Blueprints_IndexPerformance(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Execute migration
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "003_blueprints.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}

	// Insert test data
	for i := 1; i <= 100; i++ {
		// Insert blueprint
		_, err := db.Exec(
			"INSERT INTO industryBlueprints (blueprintTypeID, maxProductionLimit) VALUES (?, ?)",
			i, 1,
		)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}

		// Insert activity
		activityID := (i % 5) + 1
		_, err = db.Exec(
			"INSERT INTO industryActivities (blueprintTypeID, activityID, time) VALUES (?, ?, ?)",
			i, activityID, i*100,
		)
		if err != nil {
			t.Fatalf("Failed to insert activity data: %v", err)
		}

		// Insert material
		materialTypeID := (i % 10) + 30
		_, err = db.Exec(
			"INSERT INTO industryActivityMaterials (blueprintTypeID, activityID, materialTypeID, quantity) VALUES (?, ?, ?, ?)",
			i, activityID, materialTypeID, i*10,
		)
		if err != nil {
			t.Fatalf("Failed to insert material data: %v", err)
		}

		// Insert product
		productTypeID := i + 1000
		_, err = db.Exec(
			"INSERT INTO industryActivityProducts (blueprintTypeID, activityID, productTypeID, quantity) VALUES (?, ?, ?, ?)",
			i, activityID, productTypeID, 1,
		)
		if err != nil {
			t.Fatalf("Failed to insert product data: %v", err)
		}
	}

	// Test query using blueprintTypeID index on industryActivities
	rows, err := db.Query("SELECT time FROM industryActivities WHERE blueprintTypeID = 50")
	if err != nil {
		t.Fatalf("Failed to query by blueprintTypeID: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	if count == 0 {
		t.Error("Expected at least one row with blueprintTypeID = 50")
	}

	// Test query using materialTypeID index
	rows, err = db.Query("SELECT quantity FROM industryActivityMaterials WHERE materialTypeID = 35")
	if err != nil {
		t.Fatalf("Failed to query by materialTypeID: %v", err)
	}
	defer rows.Close()

	count = 0
	for rows.Next() {
		count++
	}
	if count == 0 {
		t.Error("Expected at least one row with materialTypeID = 35")
	}

	// Test query using productTypeID index
	rows, err = db.Query("SELECT quantity FROM industryActivityProducts WHERE productTypeID = 1050")
	if err != nil {
		t.Fatalf("Failed to query by productTypeID: %v", err)
	}
	defer rows.Close()

	count = 0
	for rows.Next() {
		count++
	}
	if count == 0 {
		t.Error("Expected at least one row with productTypeID = 1050")
	}
}

// TestMigration_003_Blueprints_Idempotence tests that migration can be run multiple times
func TestMigration_003_Blueprints_Idempotence(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Read migration file
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "003_blueprints.sql")
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

	// Verify all tables still exist after second migration
	expectedTables := []string{
		"industryBlueprints",
		"industryActivities",
		"industryActivityMaterials",
		"industryActivityProducts",
	}

	for _, table := range expectedTables {
		var tableName string
		err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&tableName)
		if err != nil {
			t.Fatalf("Table %s does not exist after second migration: %v", table, err)
		}
	}

	// Verify indexes still exist after second migration
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name LIKE 'idx_industry%'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count indexes: %v", err)
	}
	if count != 6 {
		t.Errorf("Expected 6 indexes, got %d", count)
	}
}

// TestMigration_003_Blueprints_CompositeKeys tests composite primary keys work correctly
func TestMigration_003_Blueprints_CompositeKeys(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Execute migration
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "003_blueprints.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}

	// Test industryActivities composite key (blueprintTypeID, activityID)
	t.Run("industryActivities_composite_key", func(t *testing.T) {
		// Insert first row
		_, err := db.Exec(
			"INSERT INTO industryActivities (blueprintTypeID, activityID, time) VALUES (?, ?, ?)",
			100, 1, 1000,
		)
		if err != nil {
			t.Fatalf("Failed to insert first row: %v", err)
		}

		// Insert second row with different activityID (should succeed)
		_, err = db.Exec(
			"INSERT INTO industryActivities (blueprintTypeID, activityID, time) VALUES (?, ?, ?)",
			100, 2, 2000,
		)
		if err != nil {
			t.Fatalf("Failed to insert second row with different activityID: %v", err)
		}

		// Try to insert duplicate (should fail)
		_, err = db.Exec(
			"INSERT INTO industryActivities (blueprintTypeID, activityID, time) VALUES (?, ?, ?)",
			100, 1, 3000,
		)
		if err == nil {
			t.Error("Expected error when inserting duplicate composite key, got nil")
		}
	})

	// Test industryActivityMaterials composite key (blueprintTypeID, activityID, materialTypeID)
	t.Run("industryActivityMaterials_composite_key", func(t *testing.T) {
		// Insert first row
		_, err := db.Exec(
			"INSERT INTO industryActivityMaterials (blueprintTypeID, activityID, materialTypeID, quantity) VALUES (?, ?, ?, ?)",
			200, 1, 34, 100,
		)
		if err != nil {
			t.Fatalf("Failed to insert first row: %v", err)
		}

		// Insert second row with different materialTypeID (should succeed)
		_, err = db.Exec(
			"INSERT INTO industryActivityMaterials (blueprintTypeID, activityID, materialTypeID, quantity) VALUES (?, ?, ?, ?)",
			200, 1, 35, 200,
		)
		if err != nil {
			t.Fatalf("Failed to insert second row with different materialTypeID: %v", err)
		}

		// Try to insert duplicate (should fail)
		_, err = db.Exec(
			"INSERT INTO industryActivityMaterials (blueprintTypeID, activityID, materialTypeID, quantity) VALUES (?, ?, ?, ?)",
			200, 1, 34, 300,
		)
		if err == nil {
			t.Error("Expected error when inserting duplicate composite key, got nil")
		}
	})

	// Test industryActivityProducts composite key (blueprintTypeID, activityID, productTypeID)
	t.Run("industryActivityProducts_composite_key", func(t *testing.T) {
		// Insert first row
		_, err := db.Exec(
			"INSERT INTO industryActivityProducts (blueprintTypeID, activityID, productTypeID, quantity) VALUES (?, ?, ?, ?)",
			300, 1, 1000, 1,
		)
		if err != nil {
			t.Fatalf("Failed to insert first row: %v", err)
		}

		// Insert second row with different productTypeID (should succeed)
		_, err = db.Exec(
			"INSERT INTO industryActivityProducts (blueprintTypeID, activityID, productTypeID, quantity) VALUES (?, ?, ?, ?)",
			300, 1, 1001, 2,
		)
		if err != nil {
			t.Fatalf("Failed to insert second row with different productTypeID: %v", err)
		}

		// Try to insert duplicate (should fail)
		_, err = db.Exec(
			"INSERT INTO industryActivityProducts (blueprintTypeID, activityID, productTypeID, quantity) VALUES (?, ?, ?, ?)",
			300, 1, 1000, 3,
		)
		if err == nil {
			t.Error("Expected error when inserting duplicate composite key, got nil")
		}
	})
}

// TestMigration_004_Dogma tests the 004_dogma.sql migration
func TestMigration_004_Dogma(t *testing.T) {
	// Create in-memory database
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Read migration file
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "004_dogma.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}

	// Execute migration
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}

	// Verify all tables exist
	expectedTables := []string{
		"dogmaAttributes",
		"dogmaEffects",
		"dogmaTypeAttributes",
		"dogmaTypeEffects",
	}

	for _, table := range expectedTables {
		var tableName string
		err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&tableName)
		if err != nil {
			t.Fatalf("Table %s was not created: %v", table, err)
		}
		if tableName != table {
			t.Errorf("Expected table name '%s', got '%s'", table, tableName)
		}
	}
}

// TestMigration_004_Dogma_Schema verifies the schema structure
func TestMigration_004_Dogma_Schema(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Read and execute migration
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "004_dogma.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}

	// Verify dogmaAttributes schema
	t.Run("dogmaAttributes", func(t *testing.T) {
		expectedColumns := []string{
			"attributeID", "attributeName", "description", "iconID", "defaultValue",
			"published", "displayName", "unitID", "stackable", "highIsGood",
		}
		rows, err := db.Query("PRAGMA table_info(dogmaAttributes)")
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

			// Verify attributeID is PRIMARY KEY
			if name == "attributeID" && pk != 1 {
				t.Errorf("attributeID should be PRIMARY KEY")
			}
		}

		for _, col := range expectedColumns {
			if !columnMap[col] {
				t.Errorf("Expected column '%s' not found in table", col)
			}
		}
	})

	// Verify dogmaEffects schema
	t.Run("dogmaEffects", func(t *testing.T) {
		expectedColumns := []string{
			"effectID", "effectName", "effectCategory", "preExpression", "postExpression",
			"description", "guid", "iconID", "isOffensive", "isAssistance",
			"durationAttributeID", "trackingSpeedAttributeID", "dischargeAttributeID",
			"rangeAttributeID", "falloffAttributeID", "disallowAutoRepeat", "published",
			"displayName", "isWarpSafe", "rangeChance", "electronicChance",
			"propulsionChance", "distribution", "sfxName", "npcUsageChanceAttributeID",
			"npcActivationChanceAttributeID", "fittingUsageChanceAttributeID", "modifierInfo",
		}
		rows, err := db.Query("PRAGMA table_info(dogmaEffects)")
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

			// Verify effectID is PRIMARY KEY
			if name == "effectID" && pk != 1 {
				t.Errorf("effectID should be PRIMARY KEY")
			}
		}

		for _, col := range expectedColumns {
			if !columnMap[col] {
				t.Errorf("Expected column '%s' not found in table", col)
			}
		}
	})

	// Verify dogmaTypeAttributes schema
	t.Run("dogmaTypeAttributes", func(t *testing.T) {
		expectedColumns := []string{"typeID", "attributeID", "valueInt", "valueFloat"}
		rows, err := db.Query("PRAGMA table_info(dogmaTypeAttributes)")
		if err != nil {
			t.Fatalf("Failed to get table info: %v", err)
		}
		defer rows.Close()

		columnMap := make(map[string]bool)
		pkCount := 0
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

			// Count PRIMARY KEY columns
			if pk > 0 {
				pkCount++
			}
		}

		for _, col := range expectedColumns {
			if !columnMap[col] {
				t.Errorf("Expected column '%s' not found in table", col)
			}
		}

		// Verify composite PRIMARY KEY (typeID, attributeID)
		if pkCount != 2 {
			t.Errorf("Expected composite PRIMARY KEY with 2 columns, got %d", pkCount)
		}
	})

	// Verify dogmaTypeEffects schema
	t.Run("dogmaTypeEffects", func(t *testing.T) {
		expectedColumns := []string{"typeID", "effectID", "isDefault"}
		rows, err := db.Query("PRAGMA table_info(dogmaTypeEffects)")
		if err != nil {
			t.Fatalf("Failed to get table info: %v", err)
		}
		defer rows.Close()

		columnMap := make(map[string]bool)
		pkCount := 0
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

			// Count PRIMARY KEY columns
			if pk > 0 {
				pkCount++
			}
		}

		for _, col := range expectedColumns {
			if !columnMap[col] {
				t.Errorf("Expected column '%s' not found in table", col)
			}
		}

		// Verify composite PRIMARY KEY (typeID, effectID)
		if pkCount != 2 {
			t.Errorf("Expected composite PRIMARY KEY with 2 columns, got %d", pkCount)
		}
	})
}

// TestMigration_004_Dogma_Indexes verifies the indexes are created
func TestMigration_004_Dogma_Indexes(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Read and execute migration
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "004_dogma.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}

	// Test indexes for dogmaAttributes
	t.Run("dogmaAttributes_indexes", func(t *testing.T) {
		expectedIndexes := map[string]bool{
			"idx_dogmaAttributes_attributeName": false,
		}

		rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='dogmaAttributes' AND name NOT LIKE 'sqlite_%'")
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

		for indexName, found := range expectedIndexes {
			if !found {
				t.Errorf("Expected index '%s' not found", indexName)
			}
		}
	})

	// Test indexes for dogmaEffects
	t.Run("dogmaEffects_indexes", func(t *testing.T) {
		expectedIndexes := map[string]bool{
			"idx_dogmaEffects_effectName":     false,
			"idx_dogmaEffects_effectCategory": false,
		}

		rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='dogmaEffects' AND name NOT LIKE 'sqlite_%'")
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

		for indexName, found := range expectedIndexes {
			if !found {
				t.Errorf("Expected index '%s' not found", indexName)
			}
		}
	})

	// Test indexes for dogmaTypeAttributes
	t.Run("dogmaTypeAttributes_indexes", func(t *testing.T) {
		expectedIndexes := map[string]bool{
			"idx_dogmaTypeAttributes_typeID":      false,
			"idx_dogmaTypeAttributes_attributeID": false,
		}

		rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='dogmaTypeAttributes' AND name NOT LIKE 'sqlite_%'")
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

		for indexName, found := range expectedIndexes {
			if !found {
				t.Errorf("Expected index '%s' not found", indexName)
			}
		}
	})

	// Test indexes for dogmaTypeEffects
	t.Run("dogmaTypeEffects_indexes", func(t *testing.T) {
		expectedIndexes := map[string]bool{
			"idx_dogmaTypeEffects_typeID":   false,
			"idx_dogmaTypeEffects_effectID": false,
		}

		rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='dogmaTypeEffects' AND name NOT LIKE 'sqlite_%'")
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

		for indexName, found := range expectedIndexes {
			if !found {
				t.Errorf("Expected index '%s' not found", indexName)
			}
		}
	})
}

// TestMigration_004_Dogma_DataInsertion tests that data can be inserted
func TestMigration_004_Dogma_DataInsertion(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Execute migration
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "004_dogma.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}

	// Insert test data into dogmaAttributes
	_, err = db.Exec(
		"INSERT INTO dogmaAttributes (attributeID, attributeName, defaultValue, published) VALUES (?, ?, ?, ?)",
		4, "agility", 0.0, 1,
	)
	if err != nil {
		t.Fatalf("Failed to insert into dogmaAttributes: %v", err)
	}

	// Insert test data into dogmaEffects
	_, err = db.Exec(
		"INSERT INTO dogmaEffects (effectID, effectName, effectCategory, published) VALUES (?, ?, ?, ?)",
		11, "loPower", 0, 1,
	)
	if err != nil {
		t.Fatalf("Failed to insert into dogmaEffects: %v", err)
	}

	// Insert test data into dogmaTypeAttributes
	_, err = db.Exec(
		"INSERT INTO dogmaTypeAttributes (typeID, attributeID, valueFloat) VALUES (?, ?, ?)",
		34, 4, 5.0,
	)
	if err != nil {
		t.Fatalf("Failed to insert into dogmaTypeAttributes: %v", err)
	}

	// Insert test data into dogmaTypeEffects
	_, err = db.Exec(
		"INSERT INTO dogmaTypeEffects (typeID, effectID, isDefault) VALUES (?, ?, ?)",
		34, 11, 1,
	)
	if err != nil {
		t.Fatalf("Failed to insert into dogmaTypeEffects: %v", err)
	}

	// Verify data was inserted into dogmaAttributes
	var attributeName string
	err = db.QueryRow("SELECT attributeName FROM dogmaAttributes WHERE attributeID = ?", 4).Scan(&attributeName)
	if err != nil {
		t.Fatalf("Failed to query dogmaAttributes: %v", err)
	}
	if attributeName != "agility" {
		t.Errorf("Expected attributeName 'agility', got '%s'", attributeName)
	}

	// Verify data was inserted into dogmaEffects
	var effectName string
	err = db.QueryRow("SELECT effectName FROM dogmaEffects WHERE effectID = ?", 11).Scan(&effectName)
	if err != nil {
		t.Fatalf("Failed to query dogmaEffects: %v", err)
	}
	if effectName != "loPower" {
		t.Errorf("Expected effectName 'loPower', got '%s'", effectName)
	}

	// Verify data was inserted into dogmaTypeAttributes
	var valueFloat float64
	err = db.QueryRow("SELECT valueFloat FROM dogmaTypeAttributes WHERE typeID = ? AND attributeID = ?", 34, 4).Scan(&valueFloat)
	if err != nil {
		t.Fatalf("Failed to query dogmaTypeAttributes: %v", err)
	}
	if valueFloat != 5.0 {
		t.Errorf("Expected valueFloat 5.0, got %f", valueFloat)
	}

	// Verify data was inserted into dogmaTypeEffects
	var isDefault int
	err = db.QueryRow("SELECT isDefault FROM dogmaTypeEffects WHERE typeID = ? AND effectID = ?", 34, 11).Scan(&isDefault)
	if err != nil {
		t.Fatalf("Failed to query dogmaTypeEffects: %v", err)
	}
	if isDefault != 1 {
		t.Errorf("Expected isDefault 1, got %d", isDefault)
	}
}

// TestMigration_004_Dogma_IndexPerformance tests that indexes are used
func TestMigration_004_Dogma_IndexPerformance(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Execute migration
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "004_dogma.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}

	// Insert test data into dogmaAttributes
	for i := 1; i <= 100; i++ {
		_, err := db.Exec(
			"INSERT INTO dogmaAttributes (attributeID, attributeName, defaultValue, published) VALUES (?, ?, ?, ?)",
			i, "attribute_"+string(rune(i)), float64(i), i%2,
		)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// Insert test data into dogmaEffects
	for i := 1; i <= 100; i++ {
		_, err := db.Exec(
			"INSERT INTO dogmaEffects (effectID, effectName, effectCategory, published) VALUES (?, ?, ?, ?)",
			i, "effect_"+string(rune(i)), i%5, i%2,
		)
		if err != nil {
			t.Fatalf("Failed to insert effect data: %v", err)
		}
	}

	// Insert test data into dogmaTypeAttributes
	for i := 1; i <= 100; i++ {
		typeID := i
		attributeID := (i % 10) + 1
		_, err := db.Exec(
			"INSERT INTO dogmaTypeAttributes (typeID, attributeID, valueFloat) VALUES (?, ?, ?)",
			typeID, attributeID, float64(i*10),
		)
		if err != nil {
			t.Fatalf("Failed to insert type attribute data: %v", err)
		}
	}

	// Insert test data into dogmaTypeEffects
	for i := 1; i <= 100; i++ {
		typeID := i
		effectID := (i % 10) + 1
		_, err := db.Exec(
			"INSERT INTO dogmaTypeEffects (typeID, effectID, isDefault) VALUES (?, ?, ?)",
			typeID, effectID, i%2,
		)
		if err != nil {
			t.Fatalf("Failed to insert type effect data: %v", err)
		}
	}

	// Test query using attributeName index
	rows, err := db.Query("SELECT attributeID FROM dogmaAttributes WHERE attributeName LIKE 'attribute_%'")
	if err != nil {
		t.Fatalf("Failed to query by attributeName: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	if count == 0 {
		t.Error("Expected at least one row with attributeName like 'attribute_%'")
	}

	// Test query using effectCategory index
	rows, err = db.Query("SELECT effectID FROM dogmaEffects WHERE effectCategory = 2")
	if err != nil {
		t.Fatalf("Failed to query by effectCategory: %v", err)
	}
	defer rows.Close()

	count = 0
	for rows.Next() {
		count++
	}
	if count == 0 {
		t.Error("Expected at least one row with effectCategory = 2")
	}

	// Test query using typeID index on dogmaTypeAttributes
	rows, err = db.Query("SELECT attributeID FROM dogmaTypeAttributes WHERE typeID = 50")
	if err != nil {
		t.Fatalf("Failed to query by typeID: %v", err)
	}
	defer rows.Close()

	count = 0
	for rows.Next() {
		count++
	}
	if count == 0 {
		t.Error("Expected at least one row with typeID = 50")
	}

	// Test query using effectID index on dogmaTypeEffects
	rows, err = db.Query("SELECT typeID FROM dogmaTypeEffects WHERE effectID = 5")
	if err != nil {
		t.Fatalf("Failed to query by effectID: %v", err)
	}
	defer rows.Close()

	count = 0
	for rows.Next() {
		count++
	}
	if count == 0 {
		t.Error("Expected at least one row with effectID = 5")
	}
}

// TestMigration_004_Dogma_Idempotence tests that migration can be run multiple times
func TestMigration_004_Dogma_Idempotence(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Read migration file
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "004_dogma.sql")
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

	// Verify all tables still exist after second migration
	expectedTables := []string{
		"dogmaAttributes",
		"dogmaEffects",
		"dogmaTypeAttributes",
		"dogmaTypeEffects",
	}

	for _, table := range expectedTables {
		var tableName string
		err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&tableName)
		if err != nil {
			t.Fatalf("Table %s does not exist after second migration: %v", table, err)
		}
	}

	// Verify indexes still exist after second migration
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name LIKE 'idx_dogma%'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count indexes: %v", err)
	}
	if count != 7 {
		t.Errorf("Expected 7 indexes, got %d", count)
	}
}

// TestMigration_004_Dogma_CompositeKeys tests composite primary keys work correctly
func TestMigration_004_Dogma_CompositeKeys(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Execute migration
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "004_dogma.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}

	// Test dogmaTypeAttributes composite key (typeID, attributeID)
	t.Run("dogmaTypeAttributes_composite_key", func(t *testing.T) {
		// Insert first row
		_, err := db.Exec(
			"INSERT INTO dogmaTypeAttributes (typeID, attributeID, valueFloat) VALUES (?, ?, ?)",
			100, 1, 10.0,
		)
		if err != nil {
			t.Fatalf("Failed to insert first row: %v", err)
		}

		// Insert second row with different attributeID (should succeed)
		_, err = db.Exec(
			"INSERT INTO dogmaTypeAttributes (typeID, attributeID, valueFloat) VALUES (?, ?, ?)",
			100, 2, 20.0,
		)
		if err != nil {
			t.Fatalf("Failed to insert second row with different attributeID: %v", err)
		}

		// Try to insert duplicate (should fail)
		_, err = db.Exec(
			"INSERT INTO dogmaTypeAttributes (typeID, attributeID, valueFloat) VALUES (?, ?, ?)",
			100, 1, 30.0,
		)
		if err == nil {
			t.Error("Expected error when inserting duplicate composite key, got nil")
		}
	})

	// Test dogmaTypeEffects composite key (typeID, effectID)
	t.Run("dogmaTypeEffects_composite_key", func(t *testing.T) {
		// Insert first row
		_, err := db.Exec(
			"INSERT INTO dogmaTypeEffects (typeID, effectID, isDefault) VALUES (?, ?, ?)",
			200, 1, 1,
		)
		if err != nil {
			t.Fatalf("Failed to insert first row: %v", err)
		}

		// Insert second row with different effectID (should succeed)
		_, err = db.Exec(
			"INSERT INTO dogmaTypeEffects (typeID, effectID, isDefault) VALUES (?, ?, ?)",
			200, 2, 0,
		)
		if err != nil {
			t.Fatalf("Failed to insert second row with different effectID: %v", err)
		}

		// Try to insert duplicate (should fail)
		_, err = db.Exec(
			"INSERT INTO dogmaTypeEffects (typeID, effectID, isDefault) VALUES (?, ?, ?)",
			200, 1, 0,
		)
		if err == nil {
			t.Error("Expected error when inserting duplicate composite key, got nil")
		}
	})
}

// TestMigration_005_Universe tests the 005_universe.sql migration
func TestMigration_005_Universe(t *testing.T) {
	// Create in-memory database
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Read migration file
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "005_universe.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}

	// Execute migration
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}

	// Verify all tables exist
	expectedTables := []string{
		"mapRegions",
		"mapConstellations",
		"mapSolarSystems",
		"mapStargates",
		"mapPlanets",
	}

	for _, table := range expectedTables {
		var tableName string
		err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&tableName)
		if err != nil {
			t.Fatalf("Table %s was not created: %v", table, err)
		}
		if tableName != table {
			t.Errorf("Expected table name '%s', got '%s'", table, tableName)
		}
	}
}

// TestMigration_005_Universe_Schema verifies the schema structure
func TestMigration_005_Universe_Schema(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Read and execute migration
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "005_universe.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}

	// Verify mapRegions schema
	t.Run("mapRegions", func(t *testing.T) {
		expectedColumns := []string{
			"regionID", "regionName", "x", "y", "z", "factionID",
		}
		rows, err := db.Query("PRAGMA table_info(mapRegions)")
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

			// Verify regionID is PRIMARY KEY
			if name == "regionID" && pk != 1 {
				t.Errorf("regionID should be PRIMARY KEY")
			}
		}

		for _, col := range expectedColumns {
			if !columnMap[col] {
				t.Errorf("Expected column '%s' not found in table", col)
			}
		}
	})

	// Verify mapConstellations schema
	t.Run("mapConstellations", func(t *testing.T) {
		expectedColumns := []string{
			"constellationID", "constellationName", "regionID", "x", "y", "z", "factionID",
		}
		rows, err := db.Query("PRAGMA table_info(mapConstellations)")
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

			// Verify constellationID is PRIMARY KEY
			if name == "constellationID" && pk != 1 {
				t.Errorf("constellationID should be PRIMARY KEY")
			}
		}

		for _, col := range expectedColumns {
			if !columnMap[col] {
				t.Errorf("Expected column '%s' not found in table", col)
			}
		}
	})

	// Verify mapSolarSystems schema
	t.Run("mapSolarSystems", func(t *testing.T) {
		expectedColumns := []string{
			"solarSystemID", "solarSystemName", "regionID", "constellationID",
			"x", "y", "z", "security", "securityClass",
		}
		rows, err := db.Query("PRAGMA table_info(mapSolarSystems)")
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

			// Verify solarSystemID is PRIMARY KEY
			if name == "solarSystemID" && pk != 1 {
				t.Errorf("solarSystemID should be PRIMARY KEY")
			}
		}

		for _, col := range expectedColumns {
			if !columnMap[col] {
				t.Errorf("Expected column '%s' not found in table", col)
			}
		}
	})

	// Verify mapStargates schema
	t.Run("mapStargates", func(t *testing.T) {
		expectedColumns := []string{
			"stargateID", "solarSystemID", "destinationID",
		}
		rows, err := db.Query("PRAGMA table_info(mapStargates)")
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

			// Verify stargateID is PRIMARY KEY
			if name == "stargateID" && pk != 1 {
				t.Errorf("stargateID should be PRIMARY KEY")
			}
		}

		for _, col := range expectedColumns {
			if !columnMap[col] {
				t.Errorf("Expected column '%s' not found in table", col)
			}
		}
	})

	// Verify mapPlanets schema
	t.Run("mapPlanets", func(t *testing.T) {
		expectedColumns := []string{
			"planetID", "planetName", "solarSystemID", "typeID", "x", "y", "z",
		}
		rows, err := db.Query("PRAGMA table_info(mapPlanets)")
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

			// Verify planetID is PRIMARY KEY
			if name == "planetID" && pk != 1 {
				t.Errorf("planetID should be PRIMARY KEY")
			}
		}

		for _, col := range expectedColumns {
			if !columnMap[col] {
				t.Errorf("Expected column '%s' not found in table", col)
			}
		}
	})
}

// TestMigration_005_Universe_Indexes verifies the indexes are created
func TestMigration_005_Universe_Indexes(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Read and execute migration
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "005_universe.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}

	// Test indexes for mapConstellations
	t.Run("mapConstellations_indexes", func(t *testing.T) {
		expectedIndexes := map[string]bool{
			"idx_mapConstellations_regionID": false,
		}

		rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='mapConstellations' AND name NOT LIKE 'sqlite_%'")
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

		for indexName, found := range expectedIndexes {
			if !found {
				t.Errorf("Expected index '%s' not found", indexName)
			}
		}
	})

	// Test indexes for mapSolarSystems
	t.Run("mapSolarSystems_indexes", func(t *testing.T) {
		expectedIndexes := map[string]bool{
			"idx_mapSolarSystems_regionID":        false,
			"idx_mapSolarSystems_constellationID": false,
		}

		rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='mapSolarSystems' AND name NOT LIKE 'sqlite_%'")
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

		for indexName, found := range expectedIndexes {
			if !found {
				t.Errorf("Expected index '%s' not found", indexName)
			}
		}
	})

	// Test indexes for mapStargates
	t.Run("mapStargates_indexes", func(t *testing.T) {
		expectedIndexes := map[string]bool{
			"idx_mapStargates_solarSystemID": false,
			"idx_mapStargates_destinationID": false,
		}

		rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='mapStargates' AND name NOT LIKE 'sqlite_%'")
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

		for indexName, found := range expectedIndexes {
			if !found {
				t.Errorf("Expected index '%s' not found", indexName)
			}
		}
	})

	// Test indexes for mapPlanets
	t.Run("mapPlanets_indexes", func(t *testing.T) {
		expectedIndexes := map[string]bool{
			"idx_mapPlanets_solarSystemID": false,
			"idx_mapPlanets_typeID":        false,
		}

		rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='mapPlanets' AND name NOT LIKE 'sqlite_%'")
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

		for indexName, found := range expectedIndexes {
			if !found {
				t.Errorf("Expected index '%s' not found", indexName)
			}
		}
	})
}

// TestMigration_005_Universe_DataInsertion tests that data can be inserted
func TestMigration_005_Universe_DataInsertion(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Execute migration
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "005_universe.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}

	// Insert test data into mapRegions
	_, err = db.Exec(
		"INSERT INTO mapRegions (regionID, regionName, x, y, z, factionID) VALUES (?, ?, ?, ?, ?, ?)",
		10000001, "Derelik", -1.234e16, 5.678e16, -9.012e16, 500001,
	)
	if err != nil {
		t.Fatalf("Failed to insert into mapRegions: %v", err)
	}

	// Insert test data into mapConstellations
	_, err = db.Exec(
		"INSERT INTO mapConstellations (constellationID, constellationName, regionID, x, y, z, factionID) VALUES (?, ?, ?, ?, ?, ?, ?)",
		20000001, "Anbald", 10000001, -1.234e16, 5.678e16, -9.012e16, 500001,
	)
	if err != nil {
		t.Fatalf("Failed to insert into mapConstellations: %v", err)
	}

	// Insert test data into mapSolarSystems
	_, err = db.Exec(
		"INSERT INTO mapSolarSystems (solarSystemID, solarSystemName, regionID, constellationID, x, y, z, security, securityClass) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		30000001, "Tanoo", 10000001, 20000001, -1.234e16, 5.678e16, -9.012e16, 0.8, "B",
	)
	if err != nil {
		t.Fatalf("Failed to insert into mapSolarSystems: %v", err)
	}

	// Insert test data into mapStargates
	_, err = db.Exec(
		"INSERT INTO mapStargates (stargateID, solarSystemID, destinationID) VALUES (?, ?, ?)",
		50000001, 30000001, 30000002,
	)
	if err != nil {
		t.Fatalf("Failed to insert into mapStargates: %v", err)
	}

	// Insert test data into mapPlanets
	_, err = db.Exec(
		"INSERT INTO mapPlanets (planetID, planetName, solarSystemID, typeID, x, y, z) VALUES (?, ?, ?, ?, ?, ?, ?)",
		40000001, "Tanoo I", 30000001, 2016, -1.0e11, 2.0e11, -3.0e11,
	)
	if err != nil {
		t.Fatalf("Failed to insert into mapPlanets: %v", err)
	}

	// Verify data was inserted into mapRegions
	var regionName string
	err = db.QueryRow("SELECT regionName FROM mapRegions WHERE regionID = ?", 10000001).Scan(&regionName)
	if err != nil {
		t.Fatalf("Failed to query mapRegions: %v", err)
	}
	if regionName != "Derelik" {
		t.Errorf("Expected regionName 'Derelik', got '%s'", regionName)
	}

	// Verify data was inserted into mapConstellations
	var constellationName string
	err = db.QueryRow("SELECT constellationName FROM mapConstellations WHERE constellationID = ?", 20000001).Scan(&constellationName)
	if err != nil {
		t.Fatalf("Failed to query mapConstellations: %v", err)
	}
	if constellationName != "Anbald" {
		t.Errorf("Expected constellationName 'Anbald', got '%s'", constellationName)
	}

	// Verify data was inserted into mapSolarSystems
	var solarSystemName string
	err = db.QueryRow("SELECT solarSystemName FROM mapSolarSystems WHERE solarSystemID = ?", 30000001).Scan(&solarSystemName)
	if err != nil {
		t.Fatalf("Failed to query mapSolarSystems: %v", err)
	}
	if solarSystemName != "Tanoo" {
		t.Errorf("Expected solarSystemName 'Tanoo', got '%s'", solarSystemName)
	}

	// Verify data was inserted into mapStargates
	var destinationID int
	err = db.QueryRow("SELECT destinationID FROM mapStargates WHERE stargateID = ?", 50000001).Scan(&destinationID)
	if err != nil {
		t.Fatalf("Failed to query mapStargates: %v", err)
	}
	if destinationID != 30000002 {
		t.Errorf("Expected destinationID 30000002, got %d", destinationID)
	}

	// Verify data was inserted into mapPlanets
	var planetName string
	err = db.QueryRow("SELECT planetName FROM mapPlanets WHERE planetID = ?", 40000001).Scan(&planetName)
	if err != nil {
		t.Fatalf("Failed to query mapPlanets: %v", err)
	}
	if planetName != "Tanoo I" {
		t.Errorf("Expected planetName 'Tanoo I', got '%s'", planetName)
	}
}

// TestMigration_005_Universe_Idempotence tests that migration can be run multiple times
func TestMigration_005_Universe_Idempotence(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer Close(db)

	// Read migration file
	migrationPath := filepath.Join("..", "..", "migrations", "sqlite", "005_universe.sql")
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

	// Verify all tables still exist after second migration
	expectedTables := []string{
		"mapRegions",
		"mapConstellations",
		"mapSolarSystems",
		"mapStargates",
		"mapPlanets",
	}

	for _, table := range expectedTables {
		var tableName string
		err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&tableName)
		if err != nil {
			t.Fatalf("Table %s does not exist after second migration: %v", table, err)
		}
	}

	// Verify indexes still exist after second migration
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name LIKE 'idx_map%'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count indexes: %v", err)
	}
	if count != 7 {
		t.Errorf("Expected 7 indexes, got %d", count)
	}
}

// =============================================================================
// Integration Tests for All Migrations
// =============================================================================

// TestMigrationsApply tests that all migrations can be applied in correct order
// and verifies that all expected tables and indexes exist.
//
// This is the main integration test that validates:
// - All 5 migrations execute successfully
// - All tables are created
// - All indexes are created
// - Schema structure is correct
func TestMigrationsApply(t *testing.T) {
	// Create in-memory database using NewTestDB which automatically applies all migrations
	db := NewTestDB(t)

	// Verify all expected tables exist
	expectedTables := []string{
		// Migration 001: invTypes
		"invTypes",
		// Migration 002: invGroups
		"invGroups",
		// Migration 003: Blueprints
		"industryBlueprints",
		"industryActivities",
		"industryActivityMaterials",
		"industryActivityProducts",
		// Migration 004: Dogma
		"dogmaAttributes",
		"dogmaEffects",
		"dogmaTypeAttributes",
		"dogmaTypeEffects",
		// Migration 005: Universe
		"mapRegions",
		"mapConstellations",
		"mapSolarSystems",
		"mapStargates",
		"mapPlanets",
	}

	for _, table := range expectedTables {
		exists, err := TableExists(db, table)
		if err != nil {
			t.Errorf("Error checking table %s: %v", table, err)
		}
		if !exists {
			t.Errorf("Table %s should exist after migrations", table)
		}
	}

	// Verify all expected indexes exist
	expectedIndexes := []string{
		// Migration 001: invTypes indexes
		"idx_invTypes_groupID",
		"idx_invTypes_marketGroupID",
		// Migration 002: invGroups indexes
		"idx_invGroups_categoryID",
		// Migration 003: Blueprints indexes
		"idx_industryActivities_blueprintTypeID",
		"idx_industryActivities_activityID",
		"idx_industryActivityMaterials_blueprintTypeID",
		"idx_industryActivityMaterials_materialTypeID",
		"idx_industryActivityProducts_blueprintTypeID",
		"idx_industryActivityProducts_productTypeID",
		// Migration 004: Dogma indexes
		"idx_dogmaAttributes_attributeName",
		"idx_dogmaEffects_effectName",
		"idx_dogmaEffects_effectCategory",
		"idx_dogmaTypeAttributes_typeID",
		"idx_dogmaTypeAttributes_attributeID",
		"idx_dogmaTypeEffects_typeID",
		"idx_dogmaTypeEffects_effectID",
		// Migration 005: Universe indexes
		"idx_mapConstellations_regionID",
		"idx_mapSolarSystems_regionID",
		"idx_mapSolarSystems_constellationID",
		"idx_mapStargates_solarSystemID",
		"idx_mapStargates_destinationID",
		"idx_mapPlanets_solarSystemID",
		"idx_mapPlanets_typeID",
	}

	for _, index := range expectedIndexes {
		exists, err := IndexExists(db, index)
		if err != nil {
			t.Errorf("Error checking index %s: %v", index, err)
		}
		if !exists {
			t.Errorf("Index %s should exist after migrations", index)
		}
	}
}

// TestMigrationsApply_Idempotence tests that migrations can be applied multiple times
// without errors (idempotence).
//
// This validates that:
// - CREATE TABLE IF NOT EXISTS works correctly
// - CREATE INDEX IF NOT EXISTS works correctly
// - Double execution doesn't corrupt the schema
func TestMigrationsApply_Idempotence(t *testing.T) {
	// Create in-memory database
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer Close(db)

	// Apply migrations first time
	if err := ApplyMigrations(db); err != nil {
		t.Fatalf("Failed to apply migrations first time: %v", err)
	}

	// Verify tables exist after first application
	tables := []string{"invTypes", "invGroups", "dogmaAttributes"}
	for _, table := range tables {
		exists, err := TableExists(db, table)
		if err != nil {
			t.Fatalf("Error checking table %s after first migration: %v", table, err)
		}
		if !exists {
			t.Fatalf("Table %s should exist after first migration", table)
		}
	}

	// Apply migrations second time (idempotence test)
	if err := ApplyMigrations(db); err != nil {
		t.Fatalf("Failed to apply migrations second time (idempotence test): %v", err)
	}

	// Verify tables still exist after second application
	for _, table := range tables {
		exists, err := TableExists(db, table)
		if err != nil {
			t.Fatalf("Error checking table %s after second migration: %v", table, err)
		}
		if !exists {
			t.Fatalf("Table %s should still exist after second migration", table)
		}
	}

	// Verify indexes still exist
	indexes := []string{
		"idx_invTypes_groupID",
		"idx_invGroups_categoryID",
		"idx_dogmaAttributes_attributeName",
	}
	for _, index := range indexes {
		exists, err := IndexExists(db, index)
		if err != nil {
			t.Fatalf("Error checking index %s after second migration: %v", index, err)
		}
		if !exists {
			t.Fatalf("Index %s should still exist after second migration", index)
		}
	}

	// Verify no duplicate tables were created
	var tableCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('invTypes', 'invGroups', 'dogmaAttributes')").Scan(&tableCount)
	if err != nil {
		t.Fatalf("Failed to count tables: %v", err)
	}
	if tableCount != 3 {
		t.Errorf("Expected 3 tables, got %d (possible duplicates)", tableCount)
	}
}

// TestMigrationsApply_CorrectOrder tests that migrations are applied in the correct order
// by verifying the sorted file names.
func TestMigrationsApply_CorrectOrder(t *testing.T) {
	// Read migration files
	migrationsDir := filepath.Join("..", "..", "migrations", "sqlite")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("Failed to read migrations directory: %v", err)
	}

	// Filter SQL files
	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}

	// Verify we have exactly 5 migration files
	if len(migrationFiles) != 5 {
		t.Errorf("Expected 5 migration files, got %d", len(migrationFiles))
	}

	// Verify correct order (should be sorted numerically)
	expectedOrder := []string{
		"001_inv_types.sql",
		"002_inv_groups.sql",
		"003_blueprints.sql",
		"004_dogma.sql",
		"005_universe.sql",
	}

	// Sort the files (as ApplyMigrations does)
	sort.Strings(migrationFiles)

	// Verify order matches expected
	for i, expected := range expectedOrder {
		if i >= len(migrationFiles) {
			t.Errorf("Missing migration file: %s", expected)
			continue
		}
		if migrationFiles[i] != expected {
			t.Errorf("Migration order mismatch at position %d: expected %s, got %s", i, expected, migrationFiles[i])
		}
	}
}

// TestMigrationsApply_SchemaValidation performs detailed schema validation
// for critical tables to ensure correct structure.
func TestMigrationsApply_SchemaValidation(t *testing.T) {
	db := NewTestDB(t)

	// Test invTypes table structure
	t.Run("invTypes_schema", func(t *testing.T) {
		// Verify primary key
		var pkCount int
		err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('invTypes') WHERE pk = 1").Scan(&pkCount)
		if err != nil {
			t.Fatalf("Failed to check primary key: %v", err)
		}
		if pkCount != 1 {
			t.Errorf("invTypes should have 1 primary key column, got %d", pkCount)
		}

		// Verify NOT NULL constraints
		var notNullCount int
		err = db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('invTypes') WHERE name = 'typeName' AND \"notnull\" = 1").Scan(&notNullCount)
		if err != nil {
			t.Fatalf("Failed to check NOT NULL constraint: %v", err)
		}
		if notNullCount != 1 {
			t.Errorf("typeName should be NOT NULL")
		}
	})

	// Test dogmaTypeAttributes composite key
	t.Run("dogmaTypeAttributes_composite_key", func(t *testing.T) {
		var pkCount int
		err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('dogmaTypeAttributes') WHERE pk > 0").Scan(&pkCount)
		if err != nil {
			t.Fatalf("Failed to check composite key: %v", err)
		}
		if pkCount != 2 {
			t.Errorf("dogmaTypeAttributes should have 2 primary key columns (composite), got %d", pkCount)
		}
	})

	// Test industryActivities composite key
	t.Run("industryActivities_composite_key", func(t *testing.T) {
		var pkCount int
		err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('industryActivities') WHERE pk > 0").Scan(&pkCount)
		if err != nil {
			t.Fatalf("Failed to check composite key: %v", err)
		}
		if pkCount != 2 {
			t.Errorf("industryActivities should have 2 primary key columns (composite), got %d", pkCount)
		}
	})
}

// TestMigrationsApply_DataInsertion tests that data can be inserted into all tables
// after migrations are applied.
func TestMigrationsApply_DataInsertion(t *testing.T) {
	db := NewTestDB(t)

	// Test insert into invTypes
	t.Run("insert_invTypes", func(t *testing.T) {
		_, err := db.Exec("INSERT INTO invTypes (typeID, typeName, groupID) VALUES (?, ?, ?)", 1, "Test Item", 10)
		if err != nil {
			t.Errorf("Failed to insert into invTypes: %v", err)
		}
	})

	// Test insert into invGroups
	t.Run("insert_invGroups", func(t *testing.T) {
		_, err := db.Exec("INSERT INTO invGroups (groupID, groupName, categoryID) VALUES (?, ?, ?)", 10, "Test Group", 1)
		if err != nil {
			t.Errorf("Failed to insert into invGroups: %v", err)
		}
	})

	// Test insert into dogmaAttributes
	t.Run("insert_dogmaAttributes", func(t *testing.T) {
		_, err := db.Exec("INSERT INTO dogmaAttributes (attributeID, attributeName, defaultValue, published) VALUES (?, ?, ?, ?)", 1, "testAttr", 1.0, 1)
		if err != nil {
			t.Errorf("Failed to insert into dogmaAttributes: %v", err)
		}
	})

	// Test insert into mapRegions
	t.Run("insert_mapRegions", func(t *testing.T) {
		_, err := db.Exec("INSERT INTO mapRegions (regionID, regionName, x, y, z) VALUES (?, ?, ?, ?, ?)", 1, "Test Region", 0.0, 0.0, 0.0)
		if err != nil {
			t.Errorf("Failed to insert into mapRegions: %v", err)
		}
	})

	// Test insert into industryBlueprints
	t.Run("insert_industryBlueprints", func(t *testing.T) {
		_, err := db.Exec("INSERT INTO industryBlueprints (blueprintTypeID, maxProductionLimit) VALUES (?, ?)", 1, 10)
		if err != nil {
			t.Errorf("Failed to insert into industryBlueprints: %v", err)
		}
	})
}

// =============================================================================
// Helper Functions
// =============================================================================

// TableExists checks if a table exists in the SQLite database.
func TableExists(db *sqlx.DB, tableName string) (bool, error) {
	var name string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&name)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// IndexExists checks if an index exists in the SQLite database.
func IndexExists(db *sqlx.DB, indexName string) (bool, error) {
	var name string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name=?", indexName).Scan(&name)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
