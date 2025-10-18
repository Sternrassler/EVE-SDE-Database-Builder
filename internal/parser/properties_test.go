package parser_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// SimpleTestRecord for property testing
type SimpleTestRecord struct {
	ID    int    `json:"id"`
	Value string `json:"value"`
}

// TestProperties_Parser_EmptyInputProducesEmptyOutput tests that parsing empty input always yields empty results
func TestProperties_Parser_EmptyInputProducesEmptyOutput(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("empty input produces empty output", prop.ForAll(
		func() bool {
			p := parser.NewJSONLParser[SimpleTestRecord]("test_table", []string{"id", "value"})
			ctx := context.Background()

			// We need to use ParseFile with a temp file since parseReader is not exported
			// Instead, let's test with an empty temp file
			tmpFile := t.TempDir() + "/empty.jsonl"
			if err := createTempJSONLFile(tmpFile, ""); err != nil {
				return false
			}

			results, err := p.ParseFile(ctx, tmpFile)
			if err != nil {
				return false
			}

			return len(results) == 0
		},
	))

	properties.TestingRun(t)
}

// TestProperties_Parser_LineCountPreservation tests that the number of valid lines equals the number of results
func TestProperties_Parser_LineCountPreservation(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("line count equals result count", prop.ForAll(
		func(records []SimpleTestRecord) bool {
			if len(records) == 0 {
				return true // Skip empty case
			}

			// Create JSONL content from records
			var buf bytes.Buffer
			for _, record := range records {
				data, _ := json.Marshal(record)
				buf.Write(data)
				buf.WriteString("\n")
			}

			tmpFile := t.TempDir() + "/test.jsonl"
			if err := createTempJSONLFile(tmpFile, buf.String()); err != nil {
				return false
			}

			p := parser.NewJSONLParser[SimpleTestRecord]("test_table", []string{"id", "value"})
			ctx := context.Background()
			results, err := p.ParseFile(ctx, tmpFile)

			if err != nil {
				return false
			}

			return len(results) == len(records)
		},
		genSimpleTestRecords(1, 100),
	))

	properties.TestingRun(t)
}

// TestProperties_Parser_RoundtripPreservesData tests that parse → serialize → parse preserves data
func TestProperties_Parser_RoundtripPreservesData(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("roundtrip preserves data", prop.ForAll(
		func(records []SimpleTestRecord) bool {
			if len(records) == 0 {
				return true // Skip empty case
			}

			// First serialization
			var buf bytes.Buffer
			for _, record := range records {
				data, _ := json.Marshal(record)
				buf.Write(data)
				buf.WriteString("\n")
			}

			tmpFile1 := t.TempDir() + "/roundtrip1.jsonl"
			if err := createTempJSONLFile(tmpFile1, buf.String()); err != nil {
				return false
			}

			// First parse
			p := parser.NewJSONLParser[SimpleTestRecord]("test_table", []string{"id", "value"})
			ctx := context.Background()
			results1, err := p.ParseFile(ctx, tmpFile1)
			if err != nil {
				return false
			}

			// Second serialization
			var buf2 bytes.Buffer
			for _, result := range results1 {
				record, ok := result.(SimpleTestRecord)
				if !ok {
					return false
				}
				data, _ := json.Marshal(record)
				buf2.Write(data)
				buf2.WriteString("\n")
			}

			tmpFile2 := t.TempDir() + "/roundtrip2.jsonl"
			if err := createTempJSONLFile(tmpFile2, buf2.String()); err != nil {
				return false
			}

			// Second parse
			results2, err := p.ParseFile(ctx, tmpFile2)
			if err != nil {
				return false
			}

			// Compare results
			if len(results1) != len(results2) {
				return false
			}

			for i := range results1 {
				r1, ok1 := results1[i].(SimpleTestRecord)
				r2, ok2 := results2[i].(SimpleTestRecord)
				if !ok1 || !ok2 {
					return false
				}
				if r1.ID != r2.ID || r1.Value != r2.Value {
					return false
				}
			}

			return true
		},
		genSimpleTestRecords(1, 50),
	))

	properties.TestingRun(t)
}

// TestProperties_Parser_EmptyLinesIgnored tests that empty lines don't affect result count
func TestProperties_Parser_EmptyLinesIgnored(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("empty lines are ignored", prop.ForAll(
		func(records []SimpleTestRecord, emptyLineCount int) bool {
			if len(records) == 0 {
				return true // Skip empty case
			}

			// Create JSONL content with empty lines interspersed
			var buf bytes.Buffer
			for i, record := range records {
				data, _ := json.Marshal(record)
				buf.Write(data)
				buf.WriteString("\n")
				
				// Add some empty lines
				if i < emptyLineCount {
					buf.WriteString("\n")
				}
			}

			tmpFile := t.TempDir() + "/emptylines.jsonl"
			if err := createTempJSONLFile(tmpFile, buf.String()); err != nil {
				return false
			}

			p := parser.NewJSONLParser[SimpleTestRecord]("test_table", []string{"id", "value"})
			ctx := context.Background()
			results, err := p.ParseFile(ctx, tmpFile)

			if err != nil {
				return false
			}

			// Result count should equal record count, not affected by empty lines
			return len(results) == len(records)
		},
		genSimpleTestRecords(1, 50),
		gen.IntRange(0, 10),
	))

	properties.TestingRun(t)
}

// TestProperties_Parser_TableNamePreserved tests that table name is always preserved
func TestProperties_Parser_TableNamePreserved(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("table name is preserved", prop.ForAll(
		func(tableName string) bool {
			if len(tableName) == 0 {
				return true // Skip empty table names
			}

			p := parser.NewJSONLParser[SimpleTestRecord](tableName, []string{"id", "value"})
			return p.TableName() == tableName
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	properties.TestingRun(t)
}

// TestProperties_Parser_ColumnsPreserved tests that columns are always preserved
func TestProperties_Parser_ColumnsPreserved(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("columns are preserved", prop.ForAll(
		func(columns []string) bool {
			if len(columns) == 0 {
				return true // Skip empty columns
			}

			p := parser.NewJSONLParser[SimpleTestRecord]("test_table", columns)
			resultColumns := p.Columns()

			if len(resultColumns) != len(columns) {
				return false
			}

			for i, col := range columns {
				if resultColumns[i] != col {
					return false
				}
			}

			return true
		},
		gen.SliceOf(gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 })),
	))

	properties.TestingRun(t)
}

// Generator helpers

func genSimpleTestRecords(minLen, maxLen int) gopter.Gen {
	return gen.SliceOfN(
		maxLen,
		gen.Struct(reflect.TypeOf(SimpleTestRecord{}), map[string]gopter.Gen{
			"ID":    gen.IntRange(1, 10000),
			"Value": gen.AlphaString(),
		}),
	).SuchThat(func(v interface{}) bool {
		slice := v.([]SimpleTestRecord)
		return len(slice) >= minLen
	})
}

// Helper function to create temporary JSONL files
func createTempJSONLFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
