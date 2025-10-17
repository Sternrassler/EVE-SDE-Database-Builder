package worker_test

import (
	"fmt"

	apperrors "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/worker"
)

// Example_errorCollectorBasic demonstrates basic error collection
func Example_errorCollectorBasic() {
	ec := worker.NewErrorCollector()

	// Collect various errors
	err1 := apperrors.NewFatal("database connection failed", nil)
	err2 := apperrors.NewRetryable("temporary network error", nil)
	err3 := apperrors.NewSkippable("invalid record", nil).WithContext("file", "data.jsonl")

	ec.Collect(err1)
	ec.Collect(err2)
	ec.Collect(err3)

	// Get summary
	summary := ec.Summary()
	fmt.Printf("Total errors: %d\n", summary.TotalErrors)
	fmt.Printf("Fatal errors: %d\n", len(summary.Fatal))
	fmt.Printf("Retryable errors: %d\n", len(summary.Retryable))
	fmt.Printf("Skippable errors: %d\n", len(summary.Skippable))

	// Output:
	// Total errors: 3
	// Fatal errors: 1
	// Retryable errors: 1
	// Skippable errors: 1
}

// Example_errorCollectorGrouping demonstrates error grouping by file and table
func Example_errorCollectorGrouping() {
	ec := worker.NewErrorCollector()

	// Collect errors with file context
	for i := 0; i < 3; i++ {
		err := apperrors.NewSkippable("parse error", nil).WithContext("file", "users.jsonl")
		ec.Collect(err)
	}

	for i := 0; i < 2; i++ {
		err := apperrors.NewSkippable("parse error", nil).WithContext("file", "items.jsonl")
		ec.Collect(err)
	}

	// Get summary
	summary := ec.Summary()
	fmt.Printf("Errors in users.jsonl: %d\n", summary.ByFile["users.jsonl"])
	fmt.Printf("Errors in items.jsonl: %d\n", summary.ByFile["items.jsonl"])

	// Output:
	// Errors in users.jsonl: 3
	// Errors in items.jsonl: 2
}

// Example_errorCollectorSummaryReport demonstrates the summary report
func Example_errorCollectorSummaryReport() {
	ec := worker.NewErrorCollector()

	// Collect errors with both file and table context
	err1 := apperrors.NewFatal("insert failed", nil).
		WithContext("file", "agents.jsonl").
		WithContext("table", "agents")
	err2 := apperrors.NewRetryable("timeout", nil).
		WithContext("file", "types.jsonl").
		WithContext("table", "types")

	ec.Collect(err1)
	ec.Collect(err2)

	// Get summary
	summary := ec.Summary()

	// Print in deterministic order
	fmt.Printf("Total errors: %d\n", summary.TotalErrors)
	fmt.Printf("Fatal: %d, Retryable: %d\n", len(summary.Fatal), len(summary.Retryable))
	fmt.Printf("Files with errors: %d\n", len(summary.ByFile))
	fmt.Printf("Tables with errors: %d\n", len(summary.ByTable))

	// Output:
	// Total errors: 2
	// Fatal: 1, Retryable: 1
	// Files with errors: 2
	// Tables with errors: 2
}
