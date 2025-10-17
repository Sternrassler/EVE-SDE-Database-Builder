package worker

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	apperrors "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"
)

// TestNewErrorCollector tests creating a new error collector
func TestNewErrorCollector(t *testing.T) {
	ec := NewErrorCollector()
	if ec == nil {
		t.Fatal("expected non-nil ErrorCollector")
	}
	if ec.errors == nil {
		t.Error("expected errors slice to be initialized")
	}
	if len(ec.errors) != 0 {
		t.Errorf("expected empty errors slice, got %d errors", len(ec.errors))
	}
}

// TestErrorCollector_Collect tests collecting errors
func TestErrorCollector_Collect(t *testing.T) {
	ec := NewErrorCollector()

	// Test collecting nil error (should be ignored)
	ec.Collect(nil)
	if len(ec.GetErrors()) != 0 {
		t.Error("expected nil errors to be ignored")
	}

	// Test collecting a single error
	err1 := errors.New("test error 1")
	ec.Collect(err1)

	errs := ec.GetErrors()
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if errs[0] != err1 {
		t.Error("collected error doesn't match original")
	}

	// Test collecting multiple errors
	err2 := errors.New("test error 2")
	err3 := errors.New("test error 3")
	ec.Collect(err2)
	ec.Collect(err3)

	errs = ec.GetErrors()
	if len(errs) != 3 {
		t.Fatalf("expected 3 errors, got %d", len(errs))
	}
}

// TestErrorCollector_GetErrors tests getting errors
func TestErrorCollector_GetErrors(t *testing.T) {
	ec := NewErrorCollector()

	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	ec.Collect(err1)
	ec.Collect(err2)

	// Get errors should return a copy
	errs1 := ec.GetErrors()
	errs2 := ec.GetErrors()

	if len(errs1) != 2 || len(errs2) != 2 {
		t.Error("expected both calls to return 2 errors")
	}

	// Modifying the returned slice should not affect the collector
	errs1[0] = errors.New("modified")
	errs3 := ec.GetErrors()

	if errs3[0].Error() == "modified" {
		t.Error("modifying returned slice should not affect collector")
	}
}

// TestErrorCollector_ConcurrentCollect tests thread-safety
func TestErrorCollector_ConcurrentCollect(t *testing.T) {
	ec := NewErrorCollector()
	const numGoroutines = 100
	const errorsPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Spawn multiple goroutines collecting errors concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < errorsPerGoroutine; j++ {
				err := fmt.Errorf("error from goroutine %d, iteration %d", goroutineID, j)
				ec.Collect(err)
			}
		}(i)
	}

	wg.Wait()

	// Verify all errors were collected
	errs := ec.GetErrors()
	expectedCount := numGoroutines * errorsPerGoroutine
	if len(errs) != expectedCount {
		t.Errorf("expected %d errors, got %d", expectedCount, len(errs))
	}
}

// TestErrorCollector_Summary tests generating error summary
func TestErrorCollector_Summary(t *testing.T) {
	ec := NewErrorCollector()

	// Create various error types
	err1 := apperrors.NewFatal("fatal error", nil).WithContext("file", "test1.jsonl")
	err2 := apperrors.NewRetryable("retryable error", nil).WithContext("file", "test2.jsonl")
	err3 := apperrors.NewValidation("validation error", nil).WithContext("table", "users")
	err4 := apperrors.NewSkippable("skippable error", nil).WithContext("file", "test1.jsonl")
	err5 := errors.New("plain error")

	ec.Collect(err1)
	ec.Collect(err2)
	ec.Collect(err3)
	ec.Collect(err4)
	ec.Collect(err5)

	summary := ec.Summary()

	// Test total count
	if summary.TotalErrors != 5 {
		t.Errorf("expected 5 total errors, got %d", summary.TotalErrors)
	}

	// Test by type
	if summary.ByType["Fatal"] != 1 {
		t.Errorf("expected 1 fatal error, got %d", summary.ByType["Fatal"])
	}
	if summary.ByType["Retryable"] != 1 {
		t.Errorf("expected 1 retryable error, got %d", summary.ByType["Retryable"])
	}
	if summary.ByType["Validation"] != 1 {
		t.Errorf("expected 1 validation error, got %d", summary.ByType["Validation"])
	}
	if summary.ByType["Skippable"] != 1 {
		t.Errorf("expected 1 skippable error, got %d", summary.ByType["Skippable"])
	}
	if summary.ByType["Other"] != 1 {
		t.Errorf("expected 1 other error, got %d", summary.ByType["Other"])
	}

	// Test by file
	if summary.ByFile["test1.jsonl"] != 2 {
		t.Errorf("expected 2 errors for test1.jsonl, got %d", summary.ByFile["test1.jsonl"])
	}
	if summary.ByFile["test2.jsonl"] != 1 {
		t.Errorf("expected 1 error for test2.jsonl, got %d", summary.ByFile["test2.jsonl"])
	}

	// Test by table
	if summary.ByTable["users"] != 1 {
		t.Errorf("expected 1 error for users table, got %d", summary.ByTable["users"])
	}

	// Test categorized errors
	if len(summary.Fatal) != 1 {
		t.Errorf("expected 1 fatal error in category, got %d", len(summary.Fatal))
	}
	if len(summary.Retryable) != 1 {
		t.Errorf("expected 1 retryable error in category, got %d", len(summary.Retryable))
	}
	if len(summary.Validation) != 1 {
		t.Errorf("expected 1 validation error in category, got %d", len(summary.Validation))
	}
	if len(summary.Skippable) != 1 {
		t.Errorf("expected 1 skippable error in category, got %d", len(summary.Skippable))
	}
	if len(summary.Other) != 1 {
		t.Errorf("expected 1 other error in category, got %d", len(summary.Other))
	}
}

// TestErrorCollector_SummaryEmpty tests summary with no errors
func TestErrorCollector_SummaryEmpty(t *testing.T) {
	ec := NewErrorCollector()
	summary := ec.Summary()

	if summary.TotalErrors != 0 {
		t.Errorf("expected 0 total errors, got %d", summary.TotalErrors)
	}

	if len(summary.ByType) != 0 {
		t.Error("expected empty ByType map")
	}

	if len(summary.ByFile) != 0 {
		t.Error("expected empty ByFile map")
	}

	if len(summary.ByTable) != 0 {
		t.Error("expected empty ByTable map")
	}
}

// TestErrorSummary_String tests the string representation
func TestErrorSummary_String(t *testing.T) {
	// Test empty summary
	ec := NewErrorCollector()
	summary := ec.Summary()
	str := summary.String()

	if str != "No errors collected" {
		t.Errorf("expected 'No errors collected', got '%s'", str)
	}

	// Test summary with errors
	err1 := apperrors.NewFatal("fatal error", nil).WithContext("file", "test.jsonl").WithContext("table", "items")
	err2 := apperrors.NewRetryable("retryable error", nil).WithContext("file", "test.jsonl")

	ec.Collect(err1)
	ec.Collect(err2)

	summary = ec.Summary()
	str = summary.String()

	// Verify the string contains expected information
	if str == "" {
		t.Error("expected non-empty string representation")
	}

	// Should contain total count
	expectedSubstrings := []string{
		"2 total errors",
		"By Type:",
		"By File:",
		"By Table:",
	}

	for _, substr := range expectedSubstrings {
		if !contains(str, substr) {
			t.Errorf("expected summary to contain '%s'", substr)
		}
	}
}

// TestErrorCollector_ConcurrentSummary tests concurrent access to Summary
func TestErrorCollector_ConcurrentSummary(t *testing.T) {
	ec := NewErrorCollector()

	// Add some errors
	for i := 0; i < 10; i++ {
		ec.Collect(fmt.Errorf("error %d", i))
	}

	var wg sync.WaitGroup
	const numGoroutines = 50

	// Concurrently read summaries while collecting more errors
	wg.Add(numGoroutines * 2)

	for i := 0; i < numGoroutines; i++ {
		// Read goroutines
		go func() {
			defer wg.Done()
			_ = ec.Summary()
		}()

		// Write goroutines
		go func(id int) {
			defer wg.Done()
			ec.Collect(fmt.Errorf("concurrent error %d", id))
		}(i)
	}

	wg.Wait()

	// Final check - should have 10 + numGoroutines errors
	summary := ec.Summary()
	expectedTotal := 10 + numGoroutines
	if summary.TotalErrors != expectedTotal {
		t.Errorf("expected %d total errors, got %d", expectedTotal, summary.TotalErrors)
	}
}

// TestErrorCollector_MultipleFilesSameError tests grouping by file
func TestErrorCollector_MultipleFilesSameError(t *testing.T) {
	ec := NewErrorCollector()

	files := []string{"file1.jsonl", "file2.jsonl", "file1.jsonl", "file3.jsonl", "file1.jsonl"}
	for _, file := range files {
		err := apperrors.NewSkippable("parse error", nil).WithContext("file", file)
		ec.Collect(err)
	}

	summary := ec.Summary()

	if summary.ByFile["file1.jsonl"] != 3 {
		t.Errorf("expected 3 errors for file1.jsonl, got %d", summary.ByFile["file1.jsonl"])
	}
	if summary.ByFile["file2.jsonl"] != 1 {
		t.Errorf("expected 1 error for file2.jsonl, got %d", summary.ByFile["file2.jsonl"])
	}
	if summary.ByFile["file3.jsonl"] != 1 {
		t.Errorf("expected 1 error for file3.jsonl, got %d", summary.ByFile["file3.jsonl"])
	}
}

// TestErrorCollector_MultipleTablesSameError tests grouping by table
func TestErrorCollector_MultipleTablesSameError(t *testing.T) {
	ec := NewErrorCollector()

	tables := []string{"users", "items", "users", "users", "items"}
	for _, table := range tables {
		err := apperrors.NewFatal("insert error", nil).WithContext("table", table)
		ec.Collect(err)
	}

	summary := ec.Summary()

	if summary.ByTable["users"] != 3 {
		t.Errorf("expected 3 errors for users table, got %d", summary.ByTable["users"])
	}
	if summary.ByTable["items"] != 2 {
		t.Errorf("expected 2 errors for items table, got %d", summary.ByTable["items"])
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
