package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/database"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

func TestStatsCmd_ValidDatabase(t *testing.T) {
	// Create a temporary database with some test tables
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create and populate test database
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	// Create test tables
	_, err = db.Exec(`CREATE TABLE test_table1 (id INTEGER PRIMARY KEY, name TEXT)`)
	if err != nil {
		t.Fatalf("failed to create test_table1: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE test_table2 (id INTEGER PRIMARY KEY, value INTEGER)`)
	if err != nil {
		t.Fatalf("failed to create test_table2: %v", err)
	}

	// Insert test data
	_, err = db.Exec(`INSERT INTO test_table1 (name) VALUES ('test1'), ('test2'), ('test3')`)
	if err != nil {
		t.Fatalf("failed to insert into test_table1: %v", err)
	}

	_, err = db.Exec(`INSERT INTO test_table2 (value) VALUES (1), (2)`)
	if err != nil {
		t.Fatalf("failed to insert into test_table2: %v", err)
	}

	db.Close()

	// Set up the command with proper args
	cmd := newStatsCmd()
	cmd.SetArgs([]string{"--db", dbPath})

	// Execute the command
	err = cmd.Execute()
	if err != nil {
		t.Errorf("expected no error for valid database, got: %v", err)
	}
}

func TestStatsCmd_NonExistentDatabase(t *testing.T) {
	// Point to a non-existent database file
	nonExistentPath := "/tmp/does-not-exist-stats-test.db"

	// Make sure the file doesn't exist
	os.Remove(nonExistentPath)

	// Set up the command
	cmd := newStatsCmd()
	cmd.SetArgs([]string{"--db", nonExistentPath})

	// Execute the command - should return error
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for non-existent database, got nil")
	}
}

func TestStatsCmd_EmptyDatabase(t *testing.T) {
	// Create an empty database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "empty.db")

	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create empty database: %v", err)
	}
	db.Close()

	// Set up the command
	cmd := newStatsCmd()
	cmd.SetArgs([]string{"--db", dbPath})

	// Execute the command - should succeed even with no tables
	err = cmd.Execute()
	if err != nil {
		t.Errorf("expected no error for empty database, got: %v", err)
	}
}

func TestStatsCmd_EmptyDBPath(t *testing.T) {
	// Set up the command with empty db path
	cmd := newStatsCmd()
	cmd.SetArgs([]string{"--db", ""})

	// Execute the command - should return error
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for empty db path, got nil")
	}

	expectedErr := "--db darf nicht leer sein"
	if err != nil && err.Error() != expectedErr {
		t.Errorf("expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestStatsCmd_Help(t *testing.T) {
	// Test that help text is available
	cmd := newStatsCmd()

	if cmd.Use != "stats" {
		t.Errorf("expected Use to be 'stats', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}
}

func TestStatsCmd_Integration(t *testing.T) {
	// Create a temporary database with realistic data
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "integration.db")

	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	// Create tables similar to EVE SDE structure
	_, err = db.Exec(`
		CREATE TABLE invTypes (
			typeID INTEGER PRIMARY KEY,
			typeName TEXT,
			groupID INTEGER
		)
	`)
	if err != nil {
		t.Fatalf("failed to create invTypes: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE invGroups (
			groupID INTEGER PRIMARY KEY,
			groupName TEXT
		)
	`)
	if err != nil {
		t.Fatalf("failed to create invGroups: %v", err)
	}

	// Insert some test data
	for i := 1; i <= 100; i++ {
		_, err = db.Exec(`INSERT INTO invTypes (typeID, typeName, groupID) VALUES (?, ?, ?)`, i, "Type"+string(rune(i)), i%10)
		if err != nil {
			t.Fatalf("failed to insert into invTypes: %v", err)
		}
	}

	for i := 1; i <= 10; i++ {
		_, err = db.Exec(`INSERT INTO invGroups (groupID, groupName) VALUES (?, ?)`, i, "Group"+string(rune(i)))
		if err != nil {
			t.Fatalf("failed to insert into invGroups: %v", err)
		}
	}

	db.Close()

	// Simulate CLI execution
	rootCmd := &cobra.Command{Use: "esdedb"}
	statsCmd := newStatsCmd()
	rootCmd.AddCommand(statsCmd)

	// Execute with proper args
	rootCmd.SetArgs([]string{"stats", "--db", dbPath})
	err = rootCmd.Execute()
	if err != nil {
		t.Errorf("integration test failed: %v", err)
	}
}

func TestCollectStats(t *testing.T) {
	// Create a test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "collect.db")

	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	// Create test table
	_, err = db.Exec(`CREATE TABLE test_count (id INTEGER PRIMARY KEY)`)
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}

	// Insert 5 rows
	for i := 1; i <= 5; i++ {
		_, err = db.Exec(`INSERT INTO test_count (id) VALUES (?)`, i)
		if err != nil {
			t.Fatalf("failed to insert row: %v", err)
		}
	}

	// Set the global statsDBPath for collectStats (needed for fileInfo)
	oldPath := statsDBPath
	statsDBPath = dbPath
	defer func() { statsDBPath = oldPath }()

	// Collect stats
	stats, err := collectStats(db)
	if err != nil {
		t.Fatalf("collectStats failed: %v", err)
	}

	// Verify results
	if len(stats.Tables) != 1 {
		t.Errorf("expected 1 table, got %d", len(stats.Tables))
	}

	if stats.Tables[0].Name != "test_count" {
		t.Errorf("expected table name 'test_count', got '%s'", stats.Tables[0].Name)
	}

	if stats.Tables[0].RowCount != 5 {
		t.Errorf("expected 5 rows, got %d", stats.Tables[0].RowCount)
	}

	if stats.TotalRows != 5 {
		t.Errorf("expected total rows 5, got %d", stats.TotalRows)
	}

	if stats.DBSize == 0 {
		t.Error("expected non-zero database size")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{1048576, "1.0 MiB"},
		{1572864, "1.5 MiB"},
		{1073741824, "1.0 GiB"},
	}

	for _, tt := range tests {
		result := formatBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("formatBytes(%d) = %s, expected %s", tt.bytes, result, tt.expected)
		}
	}
}
