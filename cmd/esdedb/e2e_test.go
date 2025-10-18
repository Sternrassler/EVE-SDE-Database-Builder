package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/database"
	_ "github.com/mattn/go-sqlite3"
)

// TestE2E_FullImportWorkflow is a comprehensive end-to-end test that validates
// the complete import workflow:
// 1. CLI import command execution
// 2. Complete SDE import with test data
// 3. Database verification (tables populated correctly)
//
// This test must complete in <30s as per acceptance criteria.
func TestE2E_FullImportWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	startTime := time.Now()

	// Step 1: Build CLI binary
	binary := buildTestBinary(t)

	// Step 2: Setup test environment
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "e2e-test.db")

	// Use testdata/sde which contains sample JSONL files
	testDataDir := filepath.Join("..", "..", "testdata", "sde")
	absTestDataDir, err := filepath.Abs(testDataDir)
	if err != nil {
		t.Fatalf("failed to get absolute path for test data: %v", err)
	}

	// Verify test data directory exists
	if _, err := os.Stat(absTestDataDir); os.IsNotExist(err) {
		t.Fatalf("test data directory does not exist: %s", absTestDataDir)
	}

	// Step 3: Run full import command (with --skip-errors to handle files without parsers)
	t.Log("Running import command...")
	importCmd := exec.Command(binary, "import",
		"--sde-dir", absTestDataDir,
		"--db", dbPath,
		"--workers", "4",
		"--skip-errors",
	)
	output, err := importCmd.CombinedOutput()
	outputStr := string(output)

	// Note: Some test files may not have parsers, which is expected in test data
	if err != nil {
		t.Fatalf("import command failed: %v\nOutput: %s", err, outputStr)
	}

	// Step 4: Verify command output
	t.Run("verify_import_output", func(t *testing.T) {
		// Check for success indicators in output
		if !strings.Contains(outputStr, "Import Summary") {
			t.Error("expected 'Import Summary' in output")
		}
		if !strings.Contains(outputStr, "Import completed") {
			t.Error("expected 'Import completed' in output")
		}
		if !strings.Contains(outputStr, "parsed") {
			t.Error("expected 'parsed' files count in output")
		}
		// Note: Some test files may not have parsers - check for inserted rows instead
		if !strings.Contains(outputStr, "inserted") {
			t.Error("expected 'inserted' in output")
		}
	})

	// Step 5: Verify database was created
	t.Run("verify_database_exists", func(t *testing.T) {
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			t.Fatalf("database file was not created: %s", dbPath)
		}
	})

	// Step 6: Open database and verify schema
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer func() { _ = database.Close(db) }()

	// Step 7: Verify tables exist
	t.Run("verify_tables_exist", func(t *testing.T) {
		expectedTables := []string{
			"invTypes",
			"invGroups",
			"industryBlueprints",
			"dogmaAttributes",
			"dogmaEffects",
			"mapRegions",
		}

		for _, table := range expectedTables {
			var count int
			err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
			if err != nil {
				t.Errorf("failed to check table %s: %v", table, err)
			}
			if count == 0 {
				t.Errorf("expected table %s to exist", table)
			}
		}
	})

	// Step 8: Verify data was imported
	t.Run("verify_data_imported", func(t *testing.T) {
		// Check invTypes table has data
		var invTypesCount int
		err := db.QueryRow("SELECT COUNT(*) FROM invTypes").Scan(&invTypesCount)
		if err != nil {
			t.Fatalf("failed to count invTypes rows: %v", err)
		}
		if invTypesCount == 0 {
			t.Error("expected invTypes to have data")
		}
		t.Logf("invTypes has %d rows", invTypesCount)

		// Check invGroups table has data
		var invGroupsCount int
		err = db.QueryRow("SELECT COUNT(*) FROM invGroups").Scan(&invGroupsCount)
		if err != nil {
			t.Fatalf("failed to count invGroups rows: %v", err)
		}
		if invGroupsCount == 0 {
			t.Error("expected invGroups to have data")
		}
		t.Logf("invGroups has %d rows", invGroupsCount)

		// Verify specific data from testdata (Tritanium should exist)
		var typeName string
		err = db.QueryRow("SELECT typeName FROM invTypes WHERE typeID=34").Scan(&typeName)
		if err != nil {
			t.Errorf("failed to query Tritanium (typeID=34): %v", err)
		} else if typeName != "Tritanium" {
			t.Errorf("expected Tritanium, got %s", typeName)
		} else {
			t.Log("Successfully verified Tritanium data")
		}
	})

	// Step 9: Verify indexes were created
	t.Run("verify_indexes_exist", func(t *testing.T) {
		expectedIndexes := []string{
			"idx_invTypes_groupID",
			"idx_invGroups_categoryID",
		}

		for _, index := range expectedIndexes {
			var count int
			err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?", index).Scan(&count)
			if err != nil {
				t.Errorf("failed to check index %s: %v", index, err)
			}
			if count == 0 {
				t.Errorf("expected index %s to exist", index)
			}
		}
	})

	// Step 10: Verify database integrity
	t.Run("verify_database_integrity", func(t *testing.T) {
		var integrityCheck string
		err := db.QueryRow("PRAGMA integrity_check").Scan(&integrityCheck)
		if err != nil {
			t.Fatalf("failed to run integrity check: %v", err)
		}
		if integrityCheck != "ok" {
			t.Errorf("database integrity check failed: %s", integrityCheck)
		}
	})

	// Step 11: Verify performance (must complete in <30s)
	duration := time.Since(startTime)
	t.Run("verify_performance", func(t *testing.T) {
		maxDuration := 30 * time.Second
		if duration > maxDuration {
			t.Errorf("E2E test took %v, which exceeds the %v requirement", duration, maxDuration)
		}
		t.Logf("E2E test completed in %v (requirement: <%v)", duration, maxDuration)
	})

	t.Logf("Full E2E test completed successfully in %v", duration)
}

// TestE2E_ImportCommand_CLI tests the CLI import command with basic validation
func TestE2E_ImportCommand_CLI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	binary := buildTestBinary(t)
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "cli-test.db")

	testDataDir := filepath.Join("..", "..", "testdata", "sde")
	absTestDataDir, err := filepath.Abs(testDataDir)
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	// Test basic import (with --skip-errors for test data without parsers)
	cmd := exec.Command(binary, "import", "--sde-dir", absTestDataDir, "--db", dbPath, "--workers", "2", "--skip-errors")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("import command failed: %v\nOutput: %s", err, output)
	}

	// Verify output contains success message
	outputStr := string(output)
	if !strings.Contains(outputStr, "Import completed") {
		t.Errorf("expected 'Import completed' in output, got: %s", outputStr)
	}

	// Verify database exists and has tables
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer func() { _ = database.Close(db) }()

	var tableCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&tableCount)
	if err != nil {
		t.Fatalf("failed to count tables: %v", err)
	}
	if tableCount == 0 {
		t.Error("expected tables to be created")
	}
	t.Logf("Database has %d tables", tableCount)
}

// TestE2E_ImportCommand_WithWorkers tests import with different worker counts
func TestE2E_ImportCommand_WithWorkers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	binary := buildTestBinary(t)
	testDataDir := filepath.Join("..", "..", "testdata", "sde")
	absTestDataDir, err := filepath.Abs(testDataDir)
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	workerCounts := []struct {
		count int
		name  string
	}{
		{1, "workers_1"},
		{2, "workers_2"},
		{4, "workers_4"},
		{-1, "workers_auto"},
	}

	for _, tc := range workerCounts {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			dbPath := filepath.Join(tmpDir, "test.db")

			// Use fmt.Sprintf to properly convert int to string
			workerStr := strings.TrimPrefix(tc.name, "workers_")
			if workerStr == "auto" {
				workerStr = "-1"
			}

			cmd := exec.Command(binary, "import",
				"--sde-dir", absTestDataDir,
				"--db", dbPath,
				"--workers", workerStr,
				"--skip-errors",
			)

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("import with %d workers failed: %v\nOutput: %s", tc.count, err, output)
				return
			}

			// Verify database was created
			if _, err := os.Stat(dbPath); os.IsNotExist(err) {
				t.Errorf("database not created with %d workers", tc.count)
			}
		})
	}
}

// TestE2E_VerifyDatabaseSchema tests that the imported database has correct schema
func TestE2E_VerifyDatabaseSchema(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	binary := buildTestBinary(t)
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "schema-test.db")

	testDataDir := filepath.Join("..", "..", "testdata", "sde")
	absTestDataDir, err := filepath.Abs(testDataDir)
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	// Run import (with --skip-errors for test data without parsers)
	cmd := exec.Command(binary, "import", "--sde-dir", absTestDataDir, "--db", dbPath, "--skip-errors")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("import failed: %v\nOutput: %s", err, output)
	}

	// Open database
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer func() { _ = database.Close(db) }()

	// Verify invTypes columns
	t.Run("verify_invTypes_columns", func(t *testing.T) {
		expectedColumns := []string{
			"typeID", "typeName", "groupID", "description",
			"mass", "volume", "capacity", "portionSize",
		}

		rows, err := db.Query("PRAGMA table_info(invTypes)")
		if err != nil {
			t.Fatalf("failed to get table info: %v", err)
		}
		defer rows.Close()

		columnMap := make(map[string]bool)
		for rows.Next() {
			var cid int
			var name, colType string
			var notNull, pk int
			var dfltValue interface{}
			if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
				t.Fatalf("failed to scan column: %v", err)
			}
			columnMap[name] = true
		}

		for _, col := range expectedColumns {
			if !columnMap[col] {
				t.Errorf("expected column %s not found in invTypes", col)
			}
		}
	})
}
