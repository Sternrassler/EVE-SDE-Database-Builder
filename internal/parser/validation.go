// Package parser provides JSONL parsing interfaces and implementations
// for EVE SDE data files.
package parser

import (
	"fmt"
)

// Validator defines the interface for validating parsed data structures.
// Types implementing this interface can be validated for required fields,
// ranges, and format constraints.
type Validator interface {
	// Validate returns an error if the data fails validation checks.
	// Returns nil if all validation checks pass.
	Validate() error
}

// ValidateBatch validates a batch of items that implement the Validator interface.
// It filters out invalid items and returns only valid items along with all validation errors.
//
// The function processes all items even if some fail validation, collecting all errors.
// This allows callers to see all validation issues at once rather than stopping at the first error.
//
// Parameters:
//   - items: A slice of items implementing the Validator interface
//
// Returns:
//   - []T: A slice containing only items that passed validation
//   - []error: A slice of all validation errors encountered (empty if all items valid)
//
// Example usage:
//
//	type Product struct {
//	    ID    int
//	    Name  string
//	    Price float64
//	}
//
//	func (p Product) Validate() error {
//	    if p.ID <= 0 {
//	        return fmt.Errorf("invalid ID: %d", p.ID)
//	    }
//	    if p.Name == "" {
//	        return errors.New("name is required")
//	    }
//	    if p.Price < 0 {
//	        return fmt.Errorf("invalid price: %f", p.Price)
//	    }
//	    return nil
//	}
//
//	items := []Product{{ID: 1, Name: "Valid", Price: 10}, {ID: -1, Name: "Invalid"}}
//	validItems, errs := ValidateBatch(items)
//	// validItems contains only the first item
//	// errs contains the validation error for the second item
func ValidateBatch[T Validator](items []T) ([]T, []error) {
	var validItems []T
	var errors []error

	for i, item := range items {
		if err := item.Validate(); err != nil {
			// Wrap error with index for better context
			errors = append(errors, fmt.Errorf("item %d: %w", i, err))
		} else {
			validItems = append(validItems, item)
		}
	}

	return validItems, errors
}
