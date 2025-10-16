package database

import (
	"testing"
)

// TestNewTestDB_Creation tests that NewTestDB creates a valid in-memory database
func TestNewTestDB_Creation(t *testing.T) {
	db := NewTestDB(t)

	// Verify database is not nil
	if db == nil {
		t.Fatal("NewTestDB returned nil database")
	}

	// Verify connection is working
	if err := db.Ping(); err != nil {
		t.Errorf("Database ping failed: %v", err)
	}
}

// TestNewTestDB_MigrationsApplied tests that migrations are automatically applied
func TestNewTestDB_MigrationsApplied(t *testing.T) {
	db := NewTestDB(t)

	// Verify that expected tables exist (from migrations)
	expectedTables := []string{
		"invTypes",
		"invGroups",
		"industryBlueprints",
		"dogmaAttributes",
		"mapRegions",
	}

	for _, tableName := range expectedTables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&name)
		if err != nil {
			t.Errorf("Expected table '%s' not found: %v", tableName, err)
		}
		if name != tableName {
			t.Errorf("Expected table '%s', got '%s'", tableName, name)
		}
	}
}

// TestNewTestDB_Cleanup tests that cleanup is properly registered
func TestNewTestDB_Cleanup(t *testing.T) {
	// This test verifies that NewTestDB registers cleanup by checking
	// that multiple calls work independently (each gets its own cleanup)

	db1 := NewTestDB(t)
	db2 := NewTestDB(t)

	// Both databases should be independent and functional
	if err := db1.Ping(); err != nil {
		t.Errorf("Database 1 ping failed: %v", err)
	}
	if err := db2.Ping(); err != nil {
		t.Errorf("Database 2 ping failed: %v", err)
	}

	// Cleanup will be verified when the test function exits
	// If cleanup is not properly registered, we may get resource leaks
	// but the test will still pass (cleanup is implicit)
}

// TestNewTestDB_IndependentInstances tests that each test gets its own isolated database
func TestNewTestDB_IndependentInstances(t *testing.T) {
	db1 := NewTestDB(t)
	db2 := NewTestDB(t)

	// Insert data into db1
	_, err := db1.Exec(`
		INSERT INTO invTypes (typeID, typeName, groupID)
		VALUES (1, 'Test Type 1', 1)
	`)
	if err != nil {
		t.Fatalf("Failed to insert into db1: %v", err)
	}

	// Verify data exists in db1
	var count1 int
	err = db1.QueryRow("SELECT COUNT(*) FROM invTypes WHERE typeID = 1").Scan(&count1)
	if err != nil {
		t.Fatalf("Failed to query db1: %v", err)
	}
	if count1 != 1 {
		t.Errorf("Expected 1 row in db1, got %d", count1)
	}

	// Verify data does NOT exist in db2 (independent instance)
	var count2 int
	err = db2.QueryRow("SELECT COUNT(*) FROM invTypes WHERE typeID = 1").Scan(&count2)
	if err != nil {
		t.Fatalf("Failed to query db2: %v", err)
	}
	if count2 != 0 {
		t.Errorf("Expected 0 rows in db2 (independent instance), got %d", count2)
	}
}

// TestNewTestDB_PragmasApplied tests that PRAGMAs are correctly set
func TestNewTestDB_PragmasApplied(t *testing.T) {
	db := NewTestDB(t)

	// Verify foreign keys are enabled
	var foreignKeys string
	err := db.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys)
	if err != nil {
		t.Fatalf("Failed to query foreign_keys pragma: %v", err)
	}
	if foreignKeys != "1" {
		t.Errorf("Expected foreign_keys = 1, got %s", foreignKeys)
	}

	// Verify cache size is set
	var cacheSize string
	err = db.QueryRow("PRAGMA cache_size").Scan(&cacheSize)
	if err != nil {
		t.Fatalf("Failed to query cache_size pragma: %v", err)
	}
	if cacheSize != "-64000" {
		t.Errorf("Expected cache_size = -64000, got %s", cacheSize)
	}
}

// TestApplyMigrations_AllMigrationsExecuted tests that all migration files are applied
func TestApplyMigrations_AllMigrationsExecuted(t *testing.T) {
	// Create fresh in-memory database without NewTestDB
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer Close(db)

	// Apply migrations manually
	if err := ApplyMigrations(db); err != nil {
		t.Fatalf("ApplyMigrations failed: %v", err)
	}

	// Verify all expected tables exist
	expectedTables := map[string]bool{
		// From 001_inv_types.sql
		"invTypes": false,
		// From 002_inv_groups.sql
		"invGroups": false,
		// From 003_blueprints.sql
		"industryBlueprints":        false,
		"industryActivities":        false,
		"industryActivityMaterials": false,
		"industryActivityProducts":  false,
		// From 004_dogma.sql
		"dogmaAttributes":     false,
		"dogmaEffects":        false,
		"dogmaTypeAttributes": false,
		"dogmaTypeEffects":    false,
		// From 005_universe.sql
		"mapRegions":        false,
		"mapConstellations": false,
		"mapSolarSystems":   false,
		"mapStargates":      false,
		"mapPlanets":        false,
	}

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		t.Fatalf("Failed to query tables: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			t.Fatalf("Failed to scan table name: %v", err)
		}
		if _, exists := expectedTables[tableName]; exists {
			expectedTables[tableName] = true
		}
	}

	// Check that all expected tables were created
	for tableName, found := range expectedTables {
		if !found {
			t.Errorf("Expected table '%s' was not created by migrations", tableName)
		}
	}
}

// TestApplyMigrations_Idempotent tests that migrations can be applied multiple times
func TestApplyMigrations_Idempotent(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer Close(db)

	// Apply migrations first time
	if err := ApplyMigrations(db); err != nil {
		t.Fatalf("ApplyMigrations failed on first run: %v", err)
	}

	// Apply migrations second time (should not fail)
	if err := ApplyMigrations(db); err != nil {
		t.Fatalf("ApplyMigrations failed on second run (not idempotent): %v", err)
	}

	// Verify tables still exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='invTypes'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query table after second migration: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 invTypes table, got %d", count)
	}
}

// TestApplyMigrations_IndexesCreated tests that indexes are created by migrations
func TestApplyMigrations_IndexesCreated(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer Close(db)

	if err := ApplyMigrations(db); err != nil {
		t.Fatalf("ApplyMigrations failed: %v", err)
	}

	// Verify that indexes exist (sample check)
	expectedIndexes := []string{
		"idx_invTypes_groupID",
		"idx_invTypes_marketGroupID",
		"idx_invGroups_categoryID",
	}

	for _, indexName := range expectedIndexes {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name=?", indexName).Scan(&name)
		if err != nil {
			t.Errorf("Expected index '%s' not found: %v", indexName, err)
		}
		if name != indexName {
			t.Errorf("Expected index '%s', got '%s'", indexName, name)
		}
	}
}

// TestApplyMigrations_DataInsertion tests that data can be inserted after migrations
func TestApplyMigrations_DataInsertion(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer Close(db)

	if err := ApplyMigrations(db); err != nil {
		t.Fatalf("ApplyMigrations failed: %v", err)
	}

	// Test inserting data into various tables
	testCases := []struct {
		table      string
		insertSQL  string
		selectSQL  string
		expectedID int
	}{
		{
			table:      "invTypes",
			insertSQL:  "INSERT INTO invTypes (typeID, typeName, groupID) VALUES (1, 'Test Type', 1)",
			selectSQL:  "SELECT typeID FROM invTypes WHERE typeID = 1",
			expectedID: 1,
		},
		{
			table:      "invGroups",
			insertSQL:  "INSERT INTO invGroups (groupID, groupName, categoryID) VALUES (1, 'Test Group', 1)",
			selectSQL:  "SELECT groupID FROM invGroups WHERE groupID = 1",
			expectedID: 1,
		},
		{
			table:      "industryBlueprints",
			insertSQL:  "INSERT INTO industryBlueprints (blueprintTypeID, maxProductionLimit) VALUES (1, 10)",
			selectSQL:  "SELECT blueprintTypeID FROM industryBlueprints WHERE blueprintTypeID = 1",
			expectedID: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.table, func(t *testing.T) {
			// Insert test data
			_, err := db.Exec(tc.insertSQL)
			if err != nil {
				t.Fatalf("Failed to insert into %s: %v", tc.table, err)
			}

			// Verify data was inserted
			var id int
			err = db.QueryRow(tc.selectSQL).Scan(&id)
			if err != nil {
				t.Fatalf("Failed to query %s: %v", tc.table, err)
			}
			if id != tc.expectedID {
				t.Errorf("Expected ID %d, got %d", tc.expectedID, id)
			}
		})
	}
}

// TestNewTestDB_UsageExample demonstrates typical usage pattern
func TestNewTestDB_UsageExample(t *testing.T) {
	// Create test database with all migrations applied
	db := NewTestDB(t)

	// Insert test data
	_, err := db.Exec(`
		INSERT INTO invTypes (typeID, typeName, groupID, description, published)
		VALUES (34, 'Tritanium', 18, 'A basic mineral', 1)
	`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Query and verify
	var typeName string
	err = db.QueryRow("SELECT typeName FROM invTypes WHERE typeID = 34").Scan(&typeName)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if typeName != "Tritanium" {
		t.Errorf("Expected typeName 'Tritanium', got '%s'", typeName)
	}

	// Database will be automatically cleaned up when test completes
}
