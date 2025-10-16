package parser_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"
)

// ValidatableTestRow is a test structure with validation
type ValidatableTestRow struct {
	ID    int
	Name  string
	Value float64
}

// Validate implements the Validator interface for ValidatableTestRow
func (v ValidatableTestRow) Validate() error {
	if v.ID <= 0 {
		return fmt.Errorf("ID must be positive, got %d", v.ID)
	}
	if v.Name == "" {
		return errors.New("name is required")
	}
	if v.Value < 0 {
		return fmt.Errorf("value must be non-negative, got %f", v.Value)
	}
	return nil
}

// ValidatableNestedRow is a test structure with nested validation
type ValidatableNestedRow struct {
	TypeID    int
	TypeName  map[string]string
	Mass      float64
	Volume    float64
	Published bool
}

// Validate implements complex validation rules
func (v ValidatableNestedRow) Validate() error {
	// Required field check
	if v.TypeID <= 0 {
		return fmt.Errorf("typeID must be positive, got %d", v.TypeID)
	}

	// Nested field validation
	if len(v.TypeName) == 0 {
		return errors.New("typeName is required and must contain at least one localization")
	}

	// Check for required English name
	if enName, ok := v.TypeName["en"]; !ok || enName == "" {
		return errors.New("typeName must contain a non-empty English (en) translation")
	}

	// Range validation
	if v.Mass < 0 {
		return fmt.Errorf("mass must be non-negative, got %f", v.Mass)
	}
	if v.Volume < 0 {
		return fmt.Errorf("volume must be non-negative, got %f", v.Volume)
	}

	return nil
}

// ProductRow for testing edge cases
type ProductRow struct {
	ID          int
	Name        string
	Price       float64
	Quantity    int
	Description string
}

func (p ProductRow) Validate() error {
	var errs []string

	if p.ID <= 0 {
		errs = append(errs, fmt.Sprintf("invalid ID: %d", p.ID))
	}
	if p.Name == "" {
		errs = append(errs, "name is required")
	}
	if p.Price < 0 {
		errs = append(errs, fmt.Sprintf("price must be non-negative: %f", p.Price))
	}
	if p.Quantity < 0 {
		errs = append(errs, fmt.Sprintf("quantity must be non-negative: %d", p.Quantity))
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// TestValidatableTestRow_Validate tests the basic validation logic
func TestValidatableTestRow_Validate(t *testing.T) {
	tests := []struct {
		name    string
		row     ValidatableTestRow
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Valid row",
			row:     ValidatableTestRow{ID: 1, Name: "Test", Value: 10.5},
			wantErr: false,
		},
		{
			name:    "Zero ID",
			row:     ValidatableTestRow{ID: 0, Name: "Test", Value: 10.5},
			wantErr: true,
			errMsg:  "ID must be positive",
		},
		{
			name:    "Negative ID",
			row:     ValidatableTestRow{ID: -1, Name: "Test", Value: 10.5},
			wantErr: true,
			errMsg:  "ID must be positive",
		},
		{
			name:    "Empty name",
			row:     ValidatableTestRow{ID: 1, Name: "", Value: 10.5},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name:    "Negative value",
			row:     ValidatableTestRow{ID: 1, Name: "Test", Value: -5.0},
			wantErr: true,
			errMsg:  "value must be non-negative",
		},
		{
			name:    "Zero value is valid",
			row:     ValidatableTestRow{ID: 1, Name: "Test", Value: 0},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.row.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

// TestValidatableNestedRow_Validate tests complex nested validation
func TestValidatableNestedRow_Validate(t *testing.T) {
	tests := []struct {
		name    string
		row     ValidatableNestedRow
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid row",
			row: ValidatableNestedRow{
				TypeID:    34,
				TypeName:  map[string]string{"en": "Tritanium", "de": "Tritanium"},
				Mass:      0.01,
				Volume:    0.01,
				Published: true,
			},
			wantErr: false,
		},
		{
			name: "Invalid TypeID",
			row: ValidatableNestedRow{
				TypeID:   0,
				TypeName: map[string]string{"en": "Test"},
				Mass:     0.01,
				Volume:   0.01,
			},
			wantErr: true,
			errMsg:  "typeID must be positive",
		},
		{
			name: "Empty TypeName map",
			row: ValidatableNestedRow{
				TypeID:   34,
				TypeName: map[string]string{},
				Mass:     0.01,
				Volume:   0.01,
			},
			wantErr: true,
			errMsg:  "typeName is required",
		},
		{
			name: "Missing English translation",
			row: ValidatableNestedRow{
				TypeID:   34,
				TypeName: map[string]string{"de": "Tritanium"},
				Mass:     0.01,
				Volume:   0.01,
			},
			wantErr: true,
			errMsg:  "English (en) translation",
		},
		{
			name: "Empty English translation",
			row: ValidatableNestedRow{
				TypeID:   34,
				TypeName: map[string]string{"en": ""},
				Mass:     0.01,
				Volume:   0.01,
			},
			wantErr: true,
			errMsg:  "English (en) translation",
		},
		{
			name: "Negative mass",
			row: ValidatableNestedRow{
				TypeID:   34,
				TypeName: map[string]string{"en": "Test"},
				Mass:     -0.01,
				Volume:   0.01,
			},
			wantErr: true,
			errMsg:  "mass must be non-negative",
		},
		{
			name: "Negative volume",
			row: ValidatableNestedRow{
				TypeID:   34,
				TypeName: map[string]string{"en": "Test"},
				Mass:     0.01,
				Volume:   -0.01,
			},
			wantErr: true,
			errMsg:  "volume must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.row.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

// TestValidateBatch_AllValid tests ValidateBatch with all valid items
func TestValidateBatch_AllValid(t *testing.T) {
	items := []ValidatableTestRow{
		{ID: 1, Name: "Item 1", Value: 10.0},
		{ID: 2, Name: "Item 2", Value: 20.0},
		{ID: 3, Name: "Item 3", Value: 30.0},
	}

	validItems, errs := parser.ValidateBatch(items)

	if len(errs) != 0 {
		t.Errorf("Expected no errors, got %d: %v", len(errs), errs)
	}

	if len(validItems) != 3 {
		t.Errorf("Expected 3 valid items, got %d", len(validItems))
	}

	// Verify items are unchanged
	for i, item := range validItems {
		if item.ID != items[i].ID || item.Name != items[i].Name || item.Value != items[i].Value {
			t.Errorf("Item %d was modified: expected %+v, got %+v", i, items[i], item)
		}
	}
}

// TestValidateBatch_SomeInvalid tests ValidateBatch with mixed valid/invalid items
func TestValidateBatch_SomeInvalid(t *testing.T) {
	items := []ValidatableTestRow{
		{ID: 1, Name: "Valid 1", Value: 10.0},    // valid
		{ID: -1, Name: "Invalid 1", Value: 20.0}, // invalid ID
		{ID: 2, Name: "", Value: 30.0},           // invalid name
		{ID: 3, Name: "Valid 2", Value: 40.0},    // valid
		{ID: 4, Name: "Invalid 2", Value: -10.0}, // invalid value
	}

	validItems, errs := parser.ValidateBatch(items)

	// Should have 2 valid items (index 0 and 3)
	if len(validItems) != 2 {
		t.Errorf("Expected 2 valid items, got %d", len(validItems))
	}

	// Should have 3 errors (index 1, 2, 4)
	if len(errs) != 3 {
		t.Fatalf("Expected 3 errors, got %d: %v", len(errs), errs)
	}

	// Verify error messages contain item indices
	if !strings.Contains(errs[0].Error(), "item 1") {
		t.Errorf("First error should mention 'item 1', got: %v", errs[0])
	}
	if !strings.Contains(errs[1].Error(), "item 2") {
		t.Errorf("Second error should mention 'item 2', got: %v", errs[1])
	}
	if !strings.Contains(errs[2].Error(), "item 4") {
		t.Errorf("Third error should mention 'item 4', got: %v", errs[2])
	}

	// Verify valid items are the expected ones
	if validItems[0].ID != 1 || validItems[0].Name != "Valid 1" {
		t.Errorf("First valid item incorrect: got %+v", validItems[0])
	}
	if validItems[1].ID != 3 || validItems[1].Name != "Valid 2" {
		t.Errorf("Second valid item incorrect: got %+v", validItems[1])
	}
}

// TestValidateBatch_AllInvalid tests ValidateBatch with all invalid items
func TestValidateBatch_AllInvalid(t *testing.T) {
	items := []ValidatableTestRow{
		{ID: -1, Name: "Test", Value: 10.0}, // invalid ID
		{ID: 1, Name: "", Value: 20.0},      // invalid name
		{ID: 2, Name: "Test", Value: -5.0},  // invalid value
	}

	validItems, errs := parser.ValidateBatch(items)

	if len(validItems) != 0 {
		t.Errorf("Expected 0 valid items, got %d", len(validItems))
	}

	if len(errs) != 3 {
		t.Errorf("Expected 3 errors, got %d", len(errs))
	}
}

// TestValidateBatch_EmptySlice tests ValidateBatch with empty input
func TestValidateBatch_EmptySlice(t *testing.T) {
	items := []ValidatableTestRow{}

	validItems, errs := parser.ValidateBatch(items)

	if len(validItems) != 0 {
		t.Errorf("Expected 0 valid items, got %d", len(validItems))
	}

	if len(errs) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(errs))
	}
}

// TestValidateBatch_NestedStructure tests ValidateBatch with complex nested structures
func TestValidateBatch_NestedStructure(t *testing.T) {
	items := []ValidatableNestedRow{
		{
			TypeID:    34,
			TypeName:  map[string]string{"en": "Tritanium", "de": "Tritanium"},
			Mass:      0.01,
			Volume:    0.01,
			Published: true,
		},
		{
			TypeID:    35,
			TypeName:  map[string]string{"de": "Pyerit"}, // Missing English
			Mass:      0.01,
			Volume:    0.01,
			Published: true,
		},
		{
			TypeID:    36,
			TypeName:  map[string]string{"en": "Mexallon"},
			Mass:      -0.01, // Invalid mass
			Volume:    0.01,
			Published: false,
		},
		{
			TypeID:    37,
			TypeName:  map[string]string{"en": "Isogen", "de": "Isogen"},
			Mass:      0.01,
			Volume:    0.01,
			Published: true,
		},
	}

	validItems, errs := parser.ValidateBatch(items)

	// Should have 2 valid items (index 0 and 3)
	if len(validItems) != 2 {
		t.Errorf("Expected 2 valid items, got %d", len(validItems))
	}

	// Should have 2 errors (index 1 and 2)
	if len(errs) != 2 {
		t.Fatalf("Expected 2 errors, got %d: %v", len(errs), errs)
	}

	// Verify the valid items
	if validItems[0].TypeID != 34 {
		t.Errorf("First valid item should have TypeID 34, got %d", validItems[0].TypeID)
	}
	if validItems[1].TypeID != 37 {
		t.Errorf("Second valid item should have TypeID 37, got %d", validItems[1].TypeID)
	}
}

// TestValidateBatch_MultipleErrors tests that all errors are collected
func TestValidateBatch_MultipleErrors(t *testing.T) {
	items := []ProductRow{
		{ID: 1, Name: "Valid", Price: 10.0, Quantity: 5},
		{ID: -1, Name: "", Price: -5.0, Quantity: -10}, // Multiple errors
		{ID: 2, Name: "Valid 2", Price: 20.0, Quantity: 10},
	}

	validItems, errs := parser.ValidateBatch(items)

	if len(validItems) != 2 {
		t.Errorf("Expected 2 valid items, got %d", len(validItems))
	}

	if len(errs) != 1 {
		t.Fatalf("Expected 1 error (with multiple validation failures), got %d", len(errs))
	}

	// The error should contain information about all failed validations
	errStr := errs[0].Error()
	expectedSubstrings := []string{"item 1", "invalid ID", "name is required", "price", "quantity"}
	for _, substr := range expectedSubstrings {
		if !strings.Contains(errStr, substr) {
			t.Errorf("Error should contain %q, got: %s", substr, errStr)
		}
	}
}

// TestValidateBatch_LargeDataset tests performance with many items
func TestValidateBatch_LargeDataset(t *testing.T) {
	items := make([]ValidatableTestRow, 10000)
	for i := 0; i < 10000; i++ {
		items[i] = ValidatableTestRow{
			ID:    i + 1,
			Name:  fmt.Sprintf("Item %d", i),
			Value: float64(i) * 1.5,
		}
	}

	// Add some invalid items
	items[100].ID = -1     // invalid
	items[500].Name = ""   // invalid
	items[999].Value = -10 // invalid

	validItems, errs := parser.ValidateBatch(items)

	if len(validItems) != 9997 {
		t.Errorf("Expected 9997 valid items, got %d", len(validItems))
	}

	if len(errs) != 3 {
		t.Errorf("Expected 3 errors, got %d", len(errs))
	}
}

// BenchmarkValidateBatch benchmarks the ValidateBatch function
func BenchmarkValidateBatch(b *testing.B) {
	items := make([]ValidatableTestRow, 1000)
	for i := 0; i < 1000; i++ {
		items[i] = ValidatableTestRow{
			ID:    i + 1,
			Name:  fmt.Sprintf("Item %d", i),
			Value: float64(i) * 1.5,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.ValidateBatch(items)
	}
}

// BenchmarkValidateBatch_WithErrors benchmarks ValidateBatch with some invalid items
func BenchmarkValidateBatch_WithErrors(b *testing.B) {
	items := make([]ValidatableTestRow, 1000)
	for i := 0; i < 1000; i++ {
		items[i] = ValidatableTestRow{
			ID:    i + 1,
			Name:  fmt.Sprintf("Item %d", i),
			Value: float64(i) * 1.5,
		}
	}

	// Make every 10th item invalid
	for i := 0; i < 1000; i += 10 {
		items[i].ID = -1
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.ValidateBatch(items)
	}
}

// Example demonstrating ValidateBatch usage
func ExampleValidateBatch() {
	items := []ValidatableTestRow{
		{ID: 1, Name: "Valid Item", Value: 10.0},
		{ID: -1, Name: "Invalid Item", Value: 20.0}, // Invalid ID
		{ID: 2, Name: "Another Valid", Value: 30.0},
	}

	validItems, errs := parser.ValidateBatch(items)

	fmt.Printf("Valid items: %d\n", len(validItems))
	fmt.Printf("Errors: %d\n", len(errs))

	// Output:
	// Valid items: 2
	// Errors: 1
}

// Test to ensure Validator interface compatibility
func TestValidator_InterfaceCompatibility(t *testing.T) {
	var _ parser.Validator = (*ValidatableTestRow)(nil)
	var _ parser.Validator = (*ValidatableNestedRow)(nil)
	var _ parser.Validator = (*ProductRow)(nil)
}
