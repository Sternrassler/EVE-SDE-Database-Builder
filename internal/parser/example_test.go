package parser_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"
)

// TypeRow represents a simplified EVE SDE type record
type TypeRow struct {
	TypeID   int               `json:"typeID"`
	GroupID  int               `json:"groupID"`
	TypeName map[string]string `json:"typeName"`
	Mass     float64           `json:"mass,omitempty"`
	Volume   float64           `json:"volume,omitempty"`
}

// Validate implements the Validator interface for TypeRow
func (t TypeRow) Validate() error {
	if t.TypeID <= 0 {
		return fmt.Errorf("typeID must be positive, got %d", t.TypeID)
	}
	if len(t.TypeName) == 0 {
		return errors.New("typeName is required")
	}
	if _, ok := t.TypeName["en"]; !ok {
		return errors.New("typeName must contain English (en) translation")
	}
	if t.Mass < 0 {
		return fmt.Errorf("mass cannot be negative, got %f", t.Mass)
	}
	if t.Volume < 0 {
		return fmt.Errorf("volume cannot be negative, got %f", t.Volume)
	}
	return nil
}

// ExampleJSONLParser_ParseFile demonstrates basic JSONL file parsing
func ExampleJSONLParser_ParseFile() {
	// Create a temporary JSONL file for demonstration
	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "example_types.jsonl")

	content := `{"typeID":34,"groupID":18,"typeName":{"en":"Tritanium","de":"Tritanium"},"mass":0.01,"volume":0.01}
{"typeID":35,"groupID":18,"typeName":{"en":"Pyerite","de":"Pyerit"},"mass":0.01,"volume":0.01}
{"typeID":36,"groupID":18,"typeName":{"en":"Mexallon","de":"Mexallon"},"mass":0.01,"volume":0.01}
`
	_ = os.WriteFile(testFile, []byte(content), 0644)
	defer func() { _ = os.Remove(testFile) }()

	// Create a parser for TypeRow with table name and columns
	p := parser.NewJSONLParser[TypeRow](
		"invTypes",
		[]string{"typeID", "groupID", "typeName", "mass", "volume"},
	)

	// Parse the file
	ctx := context.Background()
	results, err := p.ParseFile(ctx, testFile)
	if err != nil {
		log.Fatal(err)
	}

	// Process the results
	for _, result := range results {
		if row, ok := result.(TypeRow); ok {
			fmt.Printf("Type %d: %s\n", row.TypeID, row.TypeName["en"])
		}
	}

	// Output:
	// Type 34: Tritanium
	// Type 35: Pyerite
	// Type 36: Mexallon
}

// ExampleJSONLParser_TableName demonstrates accessing parser metadata
func ExampleJSONLParser_TableName() {
	p := parser.NewJSONLParser[TypeRow](
		"invTypes",
		[]string{"typeID", "groupID"},
	)

	fmt.Println(p.TableName())
	// Output: invTypes
}

// ExampleJSONLParser_Columns demonstrates accessing column information
func ExampleJSONLParser_Columns() {
	p := parser.NewJSONLParser[TypeRow](
		"invTypes",
		[]string{"typeID", "groupID", "typeName"},
	)

	columns := p.Columns()
	for i, col := range columns {
		fmt.Printf("Column %d: %s\n", i+1, col)
	}

	// Output:
	// Column 1: typeID
	// Column 2: groupID
	// Column 3: typeName
}

// ExampleNewJSONLParser demonstrates creating a parser with different types
func ExampleNewJSONLParser() {
	// Create a parser for a specific data structure
	p := parser.NewJSONLParser[TypeRow](
		"invTypes", // database table name
		[]string{"typeID", "groupID", "typeName"}, // column names
	)

	fmt.Printf("Parser for table: %s\n", p.TableName())
	fmt.Printf("Number of columns: %d\n", len(p.Columns()))

	// Output:
	// Parser for table: invTypes
	// Number of columns: 3
}

// ExampleValidateBatch demonstrates validating parsed data
func ExampleValidateBatch() {
	// Sample data with some invalid entries
	items := []TypeRow{
		{
			TypeID:   34,
			GroupID:  18,
			TypeName: map[string]string{"en": "Tritanium", "de": "Tritanium"},
			Mass:     0.01,
			Volume:   0.01,
		},
		{
			TypeID:   -1, // Invalid: negative ID
			GroupID:  18,
			TypeName: map[string]string{"en": "Invalid"},
			Mass:     0.01,
		},
		{
			TypeID:   35,
			GroupID:  18,
			TypeName: map[string]string{"de": "Pyerit"}, // Invalid: missing English
			Mass:     0.01,
		},
		{
			TypeID:   36,
			GroupID:  18,
			TypeName: map[string]string{"en": "Mexallon"},
			Mass:     0.01,
			Volume:   0.01,
		},
	}

	// Validate the batch
	validItems, errs := parser.ValidateBatch(items)

	fmt.Printf("Valid items: %d\n", len(validItems))
	fmt.Printf("Invalid items: %d\n", len(errs))

	// Process valid items
	for _, item := range validItems {
		fmt.Printf("Type %d: %s\n", item.TypeID, item.TypeName["en"])
	}

	// Output:
	// Valid items: 2
	// Invalid items: 2
	// Type 34: Tritanium
	// Type 36: Mexallon
}
