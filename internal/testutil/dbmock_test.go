package testutil_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/database"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/testutil"
)

func TestMockDB_Exec(t *testing.T) {
	db := testutil.NewMockDB()

	expectedResult := testutil.NewMockResult(123, 5)
	db.ExecFunc = func(query string, args ...interface{}) (sql.Result, error) {
		return expectedResult, nil
	}

	query := "INSERT INTO test (name) VALUES (?)"
	result, err := db.Exec(query, "test-value")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify result
	lastID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("unexpected error getting last insert id: %v", err)
	}
	if lastID != 123 {
		t.Errorf("expected last insert id 123, got %d", lastID)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("unexpected error getting rows affected: %v", err)
	}
	if rowsAffected != 5 {
		t.Errorf("expected 5 rows affected, got %d", rowsAffected)
	}

	// Verify call was recorded
	if len(db.ExecCalls) != 1 {
		t.Fatalf("expected 1 exec call, got %d", len(db.ExecCalls))
	}

	call := db.ExecCalls[0]
	if call.Query != query {
		t.Errorf("expected query %q, got %q", query, call.Query)
	}

	if len(call.Args) != 1 || call.Args[0] != "test-value" {
		t.Errorf("expected args [test-value], got %v", call.Args)
	}
}

func TestMockDB_ExecError(t *testing.T) {
	db := testutil.NewMockDB()
	expectedErr := errors.New("exec failed")

	db.ExecFunc = func(query string, args ...interface{}) (sql.Result, error) {
		return testutil.MockResult{}, expectedErr
	}

	_, err := db.Exec("UPDATE test SET name = ?", "value")
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	// Call should still be recorded even with error
	if len(db.ExecCalls) != 1 {
		t.Errorf("expected 1 exec call even with error, got %d", len(db.ExecCalls))
	}
}

func TestMockDB_Query(t *testing.T) {
	db := testutil.NewMockDB()

	query := "SELECT * FROM test WHERE id = ?"
	_, err := db.Query(query, 123)

	// Default implementation returns error
	if err == nil {
		t.Error("expected error from default query implementation")
	}

	// Verify call was recorded
	if len(db.QueryCalls) != 1 {
		t.Fatalf("expected 1 query call, got %d", len(db.QueryCalls))
	}

	call := db.QueryCalls[0]
	if call.Query != query {
		t.Errorf("expected query %q, got %q", query, call.Query)
	}

	if len(call.Args) != 1 || call.Args[0] != 123 {
		t.Errorf("expected args [123], got %v", call.Args)
	}
}

func TestMockDB_QueryRow(t *testing.T) {
	db := testutil.NewMockDB()

	query := "SELECT name FROM test WHERE id = ?"
	row := db.QueryRow(query, 456)

	// Default implementation returns nil
	if row != nil {
		t.Error("expected nil from default query row implementation")
	}

	// Verify call was recorded
	if len(db.QueryRowCalls) != 1 {
		t.Fatalf("expected 1 query row call, got %d", len(db.QueryRowCalls))
	}

	call := db.QueryRowCalls[0]
	if call.Query != query {
		t.Errorf("expected query %q, got %q", query, call.Query)
	}

	if len(call.Args) != 1 || call.Args[0] != 456 {
		t.Errorf("expected args [456], got %v", call.Args)
	}
}

func TestMockDB_Prepare(t *testing.T) {
	db := testutil.NewMockDB()

	query := "SELECT * FROM test WHERE id = ?"
	_, err := db.Prepare(query)

	// Default implementation returns error
	if err == nil {
		t.Error("expected error from default prepare implementation")
	}

	// Verify call was recorded
	if len(db.PrepareCalls) != 1 {
		t.Fatalf("expected 1 prepare call, got %d", len(db.PrepareCalls))
	}

	if db.PrepareCalls[0] != query {
		t.Errorf("expected query %q, got %q", query, db.PrepareCalls[0])
	}
}

func TestMockDB_Begin(t *testing.T) {
	db := testutil.NewMockDB()

	_, err := db.Begin()

	// Default implementation returns error
	if err == nil {
		t.Error("expected error from default begin implementation")
	}
}

func TestMockDB_Close(t *testing.T) {
	db := testutil.NewMockDB()

	if db.Closed {
		t.Error("expected db not to be closed initially")
	}

	err := db.Close()
	if err != nil {
		t.Errorf("unexpected error from close: %v", err)
	}

	if !db.Closed {
		t.Error("expected db to be marked as closed")
	}
}

func TestMockDB_Ping(t *testing.T) {
	db := testutil.NewMockDB()

	// Default implementation returns nil
	err := db.Ping()
	if err != nil {
		t.Errorf("unexpected error from default ping: %v", err)
	}

	// Test custom ping error
	expectedErr := errors.New("connection lost")
	db.PingFunc = func() error {
		return expectedErr
	}

	err = db.Ping()
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestMockDB_Reset(t *testing.T) {
	db := testutil.NewMockDB()

	// Make some calls
	db.Exec("INSERT INTO test VALUES (?)", 1)
	db.Query("SELECT * FROM test")
	db.QueryRow("SELECT * FROM test WHERE id = ?", 1)
	db.Prepare("SELECT * FROM test")
	db.Close()

	// Verify calls were recorded
	if len(db.ExecCalls) != 1 {
		t.Fatalf("expected 1 exec call before reset, got %d", len(db.ExecCalls))
	}
	if len(db.QueryCalls) != 1 {
		t.Fatalf("expected 1 query call before reset, got %d", len(db.QueryCalls))
	}
	if len(db.QueryRowCalls) != 1 {
		t.Fatalf("expected 1 query row call before reset, got %d", len(db.QueryRowCalls))
	}
	if len(db.PrepareCalls) != 1 {
		t.Fatalf("expected 1 prepare call before reset, got %d", len(db.PrepareCalls))
	}
	if !db.Closed {
		t.Fatal("expected db to be closed before reset")
	}

	// Reset
	db.Reset()

	// Verify everything was cleared
	if len(db.ExecCalls) != 0 {
		t.Errorf("expected 0 exec calls after reset, got %d", len(db.ExecCalls))
	}
	if len(db.QueryCalls) != 0 {
		t.Errorf("expected 0 query calls after reset, got %d", len(db.QueryCalls))
	}
	if len(db.QueryRowCalls) != 0 {
		t.Errorf("expected 0 query row calls after reset, got %d", len(db.QueryRowCalls))
	}
	if len(db.PrepareCalls) != 0 {
		t.Errorf("expected 0 prepare calls after reset, got %d", len(db.PrepareCalls))
	}
	if db.Closed {
		t.Error("expected db not to be closed after reset")
	}
}

func TestMockDB_MultipleOperations(t *testing.T) {
	db := testutil.NewMockDB()

	// Configure exec to return specific result
	db.ExecFunc = func(query string, args ...interface{}) (sql.Result, error) {
		return testutil.NewMockResult(100, 10), nil
	}

	// Perform multiple operations
	queries := []string{
		"INSERT INTO test VALUES (1)",
		"UPDATE test SET name = 'a'",
		"DELETE FROM test WHERE id = 1",
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			t.Fatalf("unexpected error for query %q: %v", query, err)
		}
	}

	// Verify all calls were recorded
	if len(db.ExecCalls) != len(queries) {
		t.Errorf("expected %d exec calls, got %d", len(queries), len(db.ExecCalls))
	}

	// Verify each query
	for i, expectedQuery := range queries {
		if db.ExecCalls[i].Query != expectedQuery {
			t.Errorf("call %d: expected query %q, got %q", i, expectedQuery, db.ExecCalls[i].Query)
		}
	}
}

func TestMockResult(t *testing.T) {
	result := testutil.NewMockResult(42, 7)

	lastID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lastID != 42 {
		t.Errorf("expected last insert id 42, got %d", lastID)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rowsAffected != 7 {
		t.Errorf("expected 7 rows affected, got %d", rowsAffected)
	}
}

func TestMockResultWithError(t *testing.T) {
	expectedErr := errors.New("database error")
	result := testutil.NewMockResultWithError(expectedErr)

	lastID, err := result.LastInsertId()
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
	if lastID != 0 {
		t.Errorf("expected last insert id 0 with error, got %d", lastID)
	}

	rowsAffected, err := result.RowsAffected()
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
	if rowsAffected != 0 {
		t.Errorf("expected 0 rows affected with error, got %d", rowsAffected)
	}
}

func TestSQLXAdapter(t *testing.T) {
	// Create a real test database
	db := database.NewTestDB(t)

	// Create adapter
	adapter := testutil.NewSQLXAdapter(db)

	// Test ping
	err := adapter.Ping()
	if err != nil {
		t.Fatalf("unexpected error from ping: %v", err)
	}

	// Test exec (create table)
	_, err = adapter.Exec("CREATE TABLE IF NOT EXISTS test_adapter (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("unexpected error from exec: %v", err)
	}

	// Test insert
	result, err := adapter.Exec("INSERT INTO test_adapter (id, name) VALUES (?, ?)", 1, "test")
	if err != nil {
		t.Fatalf("unexpected error from insert: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("unexpected error getting rows affected: %v", err)
	}
	if rowsAffected != 1 {
		t.Errorf("expected 1 row affected, got %d", rowsAffected)
	}

	// Test query row
	row := adapter.QueryRow("SELECT name FROM test_adapter WHERE id = ?", 1)
	if row == nil {
		t.Fatal("expected non-nil row")
	}

	var name string
	err = row.Scan(&name)
	if err != nil {
		t.Fatalf("unexpected error scanning row: %v", err)
	}

	if name != "test" {
		t.Errorf("expected name 'test', got %q", name)
	}

	// Test close
	err = adapter.Close()
	if err != nil {
		t.Errorf("unexpected error from close: %v", err)
	}
}
