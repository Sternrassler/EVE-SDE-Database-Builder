package parser_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"
)

// FuzzJSONLParser tests the JSONL parser with randomly generated input
// to ensure it handles malformed, edge case, and unexpected data gracefully
// without crashing.
func FuzzJSONLParser(f *testing.F) {
	// Seed corpus with valid JSONL examples
	f.Add([]byte(`{"id":1,"name":"test"}`))
	f.Add([]byte(`{"id":1,"name":"test"}
{"id":2,"name":"test2"}`))
	f.Add([]byte(`{"typeID":34,"typeName":{"en":"Tritanium"},"mass":0.01}`))
	
	// Empty lines and whitespace
	f.Add([]byte(`{"id":1,"name":"test"}

{"id":2,"name":"test2"}`))
	
	// Edge cases
	f.Add([]byte(``))                           // Empty file
	f.Add([]byte(`{}`))                         // Empty JSON object
	f.Add([]byte(`{"id":0}`))                   // Minimal valid JSON
	f.Add([]byte(`{"name":""}`))                // Empty string value
	f.Add([]byte(`{"value":null}`))             // Null value
	f.Add([]byte(`{"numbers":[1,2,3]}`))        // Array value
	f.Add([]byte(`{"nested":{"key":"value"}}`)) // Nested object
	
	// Unicode and special characters
	f.Add([]byte(`{"name":"Tritanium™"}`))
	f.Add([]byte(`{"text":"日本語"}`))
	f.Add([]byte(`{"emoji":"🚀"}`))
	
	// Large numbers
	f.Add([]byte(`{"id":9223372036854775807}`)) // Max int64
	f.Add([]byte(`{"value":1.7976931348623157e+308}`)) // Large float
	
	f.Fuzz(func(t *testing.T, data []byte) {
		// Create temporary file for fuzzing
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "fuzz.jsonl")
		
		if err := os.WriteFile(testFile, data, 0644); err != nil {
			t.Skip("Failed to write test file")
		}
		
		// Create parser - using TestRow from parser_test.go
		p := parser.NewJSONLParser[TestRow]("fuzz_table", []string{"id", "name"})
		
		// Parse with context - should not panic
		ctx := context.Background()
		_, err := p.ParseFile(ctx, testFile)
		
		// The parser may return an error for invalid JSON, but it must not panic
		// We just want to ensure robustness, not that all inputs are valid
		_ = err
		
		// If we get here without panic, the test passes
	})
}

// FuzzJSONLParserNestedData tests the parser with more complex nested structures
func FuzzJSONLParserNestedData(f *testing.F) {
	// Seed corpus with nested structure examples
	f.Add([]byte(`{"typeID":34,"typeName":{"en":"Tritanium","de":"Tritanium"},"mass":0.01}`))
	f.Add([]byte(`{"typeID":35,"typeName":{},"mass":0.0}`))
	f.Add([]byte(`{"typeID":1,"typeName":{"en":"Test","de":"Test","fr":"Test"},"mass":1.0}`))
	
	f.Fuzz(func(t *testing.T, data []byte) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "fuzz_nested.jsonl")
		
		if err := os.WriteFile(testFile, data, 0644); err != nil {
			t.Skip("Failed to write test file")
		}
		
		// Using TestNestedRow from parser_test.go
		p := parser.NewJSONLParser[TestNestedRow]("fuzz_nested", []string{"typeID", "typeName", "mass"})
		
		ctx := context.Background()
		_, err := p.ParseFile(ctx, testFile)
		
		// Should not panic regardless of input
		_ = err
	})
}

// FuzzJSONLParserLargeInput tests the parser with large inputs to ensure
// buffer handling is robust
func FuzzJSONLParserLargeInput(f *testing.F) {
	// Seed with progressively larger inputs
	f.Add([]byte(`{"name":"x"}`))
	f.Add([]byte(`{"name":"` + string(make([]byte, 100)) + `"}`))
	f.Add([]byte(`{"name":"` + string(make([]byte, 1000)) + `"}`))
	
	f.Fuzz(func(t *testing.T, data []byte) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "fuzz_large.jsonl")
		
		if err := os.WriteFile(testFile, data, 0644); err != nil {
			t.Skip("Failed to write test file")
		}
		
		p := parser.NewJSONLParser[TestRow]("fuzz_large", []string{"id", "name"})
		
		ctx := context.Background()
		_, err := p.ParseFile(ctx, testFile)
		
		// Should handle large inputs without panic
		_ = err
	})
}
