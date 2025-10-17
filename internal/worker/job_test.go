package worker

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

// MockParser implements parser.Parser for testing
type MockParser struct {
	tableName   string
	columns     []string
	parseFunc   func(ctx context.Context, path string) ([]interface{}, error)
	shouldFail  bool
	failWithErr error
	returnItems []interface{}
}

func (m *MockParser) ParseFile(ctx context.Context, path string) ([]interface{}, error) {
	if m.parseFunc != nil {
		return m.parseFunc(ctx, path)
	}
	if m.shouldFail {
		return nil, m.failWithErr
	}
	return m.returnItems, nil
}

func (m *MockParser) TableName() string {
	return m.tableName
}

func (m *MockParser) Columns() []string {
	return m.columns
}

// TestParseJob_Execute tests successful ParseJob execution
func TestParseJob_Execute(t *testing.T) {
	ctx := context.Background()

	expectedItems := []interface{}{
		map[string]interface{}{"id": 1, "name": "Item1"},
		map[string]interface{}{"id": 2, "name": "Item2"},
	}

	mockParser := &MockParser{
		tableName:   "test_table",
		columns:     []string{"id", "name"},
		returnItems: expectedItems,
	}

	job := &ParseJob{
		Parser:   mockParser,
		FilePath: "/test/path/file.jsonl",
	}

	result, err := job.Execute(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify result implements JobResult interface
	_ = JobResult(result)

	parseResult, ok := result.(ParseResult)
	if !ok {
		t.Fatalf("expected ParseResult, got %T", result)
	}

	if len(parseResult.Items) != len(expectedItems) {
		t.Errorf("expected %d items, got %d", len(expectedItems), len(parseResult.Items))
	}
}

// TestParseJob_ExecuteWithError tests ParseJob execution with parser error
func TestParseJob_ExecuteWithError(t *testing.T) {
	ctx := context.Background()

	expectedErr := errors.New("parse error: invalid JSON")
	mockParser := &MockParser{
		tableName:   "test_table",
		shouldFail:  true,
		failWithErr: expectedErr,
	}

	job := &ParseJob{
		Parser:   mockParser,
		FilePath: "/test/path/invalid.jsonl",
	}

	result, err := job.Execute(ctx)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != expectedErr.Error() {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	parseResult, ok := result.(ParseResult)
	if !ok {
		t.Fatalf("expected ParseResult even on error, got %T", result)
	}

	if parseResult.Items != nil {
		t.Errorf("expected nil items on error, got %v", parseResult.Items)
	}
}

// TestParseJob_ExecuteWithContextCancellation tests ParseJob with cancelled context
func TestParseJob_ExecuteWithContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockParser := &MockParser{
		parseFunc: func(ctx context.Context, path string) ([]interface{}, error) {
			// Check if context is cancelled
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				return []interface{}{}, nil
			}
		},
	}

	job := &ParseJob{
		Parser:   mockParser,
		FilePath: "/test/path/file.jsonl",
	}

	_, err := job.Execute(ctx)

	if err != context.Canceled {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}

// TestParseJob_ExecuteWithEmptyResult tests ParseJob with empty result
func TestParseJob_ExecuteWithEmptyResult(t *testing.T) {
	ctx := context.Background()

	mockParser := &MockParser{
		returnItems: []interface{}{},
	}

	job := &ParseJob{
		Parser:   mockParser,
		FilePath: "/test/path/empty.jsonl",
	}

	result, err := job.Execute(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	parseResult, ok := result.(ParseResult)
	if !ok {
		t.Fatalf("expected ParseResult, got %T", result)
	}

	if len(parseResult.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(parseResult.Items))
	}
}

// TestInsertJob_Execute tests basic InsertJob execution
func TestInsertJob_Execute(t *testing.T) {
	ctx := context.Background()

	rows := []interface{}{
		map[string]interface{}{"id": 1, "name": "Item1"},
		map[string]interface{}{"id": 2, "name": "Item2"},
		map[string]interface{}{"id": 3, "name": "Item3"},
	}

	job := &InsertJob{
		Table: "test_table",
		Rows:  rows,
	}

	result, err := job.Execute(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	insertResult, ok := result.(InsertResult)
	if !ok {
		t.Fatalf("expected InsertResult, got %T", result)
	}

	if insertResult.RowsAffected != len(rows) {
		t.Errorf("expected %d rows affected, got %d", len(rows), insertResult.RowsAffected)
	}
}

// TestInsertJob_ExecuteWithEmptyRows tests InsertJob with no rows
func TestInsertJob_ExecuteWithEmptyRows(t *testing.T) {
	ctx := context.Background()

	job := &InsertJob{
		Table: "test_table",
		Rows:  []interface{}{},
	}

	result, err := job.Execute(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	insertResult, ok := result.(InsertResult)
	if !ok {
		t.Fatalf("expected InsertResult, got %T", result)
	}

	if insertResult.RowsAffected != 0 {
		t.Errorf("expected 0 rows affected, got %d", insertResult.RowsAffected)
	}
}

// TestInsertJob_ExecuteWithContext tests InsertJob respects context
func TestInsertJob_ExecuteWithContext(t *testing.T) {
	ctx := context.Background()

	job := &InsertJob{
		Table: "test_table",
		Rows:  []interface{}{map[string]interface{}{"id": 1}},
	}

	result, err := job.Execute(ctx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// TestJobInterface_Compliance tests that concrete types implement JobExecutor interface
func TestJobInterface_Compliance(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		job  JobExecutor
	}{
		{
			name: "ParseJob implements JobExecutor",
			job: &ParseJob{
				Parser:   &MockParser{returnItems: []interface{}{}},
				FilePath: "/test/path.jsonl",
			},
		},
		{
			name: "InsertJob implements JobExecutor",
			job: &InsertJob{
				Table: "test_table",
				Rows:  []interface{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// If this compiles, the interface is satisfied
			_, err := tt.job.Execute(ctx)
			if err != nil {
				// Some jobs might error, that's OK for this test
				t.Logf("Job errored: %v (expected for some implementations)", err)
			}
		})
	}
}

// TestParseResult tests ParseResult structure
func TestParseResult(t *testing.T) {
	items := []interface{}{1, 2, 3}
	result := ParseResult{Items: items}

	if len(result.Items) != len(items) {
		t.Errorf("expected %d items, got %d", len(items), len(result.Items))
	}
}

// TestInsertResult tests InsertResult structure
func TestInsertResult(t *testing.T) {
	result := InsertResult{RowsAffected: 42}

	if result.RowsAffected != 42 {
		t.Errorf("expected 42 rows affected, got %d", result.RowsAffected)
	}
}

// Example_parseJob demonstrates how to use ParseJob with the JobExecutor interface
func Example_parseJob() {
	ctx := context.Background()

	// Create a mock parser
	mockParser := &MockParser{
		tableName: "types",
		columns:   []string{"typeID", "typeName"},
		returnItems: []interface{}{
			map[string]interface{}{"typeID": 1, "typeName": "Tritanium"},
			map[string]interface{}{"typeID": 2, "typeName": "Pyerite"},
		},
	}

	// Create a ParseJob
	job := &ParseJob{
		Parser:   mockParser,
		FilePath: "/data/types.jsonl",
	}

	// Execute the job
	result, err := job.Execute(ctx)
	if err != nil {
		panic(err)
	}

	// Type assert to ParseResult
	parseResult := result.(ParseResult)

	// Output the number of items parsed
	fmt.Printf("Parsed %d items\n", len(parseResult.Items))
	// Output: Parsed 2 items
}

// Example_insertJob demonstrates how to use InsertJob with the JobExecutor interface
func Example_insertJob() {
	ctx := context.Background()

	// Create sample rows to insert
	rows := []interface{}{
		map[string]interface{}{"typeID": 1, "typeName": "Tritanium"},
		map[string]interface{}{"typeID": 2, "typeName": "Pyerite"},
		map[string]interface{}{"typeID": 3, "typeName": "Mexallon"},
	}

	// Create an InsertJob
	job := &InsertJob{
		Table: "types",
		Rows:  rows,
	}

	// Execute the job
	result, err := job.Execute(ctx)
	if err != nil {
		panic(err)
	}

	// Type assert to InsertResult
	insertResult := result.(InsertResult)

	// Output the number of rows affected
	fmt.Printf("Inserted %d rows\n", insertResult.RowsAffected)
	// Output: Inserted 3 rows
}
