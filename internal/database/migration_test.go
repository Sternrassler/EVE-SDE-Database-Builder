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
