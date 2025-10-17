package worker

import (
	"fmt"
	"sync"

	apperrors "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"
)

// ErrorCollector collects errors from multiple workers in a thread-safe manner
type ErrorCollector struct {
	errors []error
	mu     sync.Mutex
}

// NewErrorCollector creates a new ErrorCollector
func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{
		errors: make([]error, 0),
	}
}

// Collect adds an error to the collection in a thread-safe manner
func (ec *ErrorCollector) Collect(err error) {
	if err == nil {
		return
	}
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.errors = append(ec.errors, err)
}

// GetErrors returns all collected errors
func (ec *ErrorCollector) GetErrors() []error {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	// Return a copy to prevent external modification
	result := make([]error, len(ec.errors))
	copy(result, ec.errors)
	return result
}

// ErrorSummary provides a summary of collected errors grouped by type and context
type ErrorSummary struct {
	TotalErrors int
	ByType      map[string]int
	ByFile      map[string]int
	ByTable     map[string]int
	Fatal       []error
	Retryable   []error
	Validation  []error
	Skippable   []error
	Other       []error
}

// Summary generates an ErrorSummary from collected errors
func (ec *ErrorCollector) Summary() ErrorSummary {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	summary := ErrorSummary{
		TotalErrors: len(ec.errors),
		ByType:      make(map[string]int),
		ByFile:      make(map[string]int),
		ByTable:     make(map[string]int),
		Fatal:       make([]error, 0),
		Retryable:   make([]error, 0),
		Validation:  make([]error, 0),
		Skippable:   make([]error, 0),
		Other:       make([]error, 0),
	}

	for _, err := range ec.errors {
		// Classify by error type
		if appErr, ok := err.(*apperrors.AppError); ok {
			errType := appErr.Type.String()
			summary.ByType[errType]++

			// Group by error type
			switch appErr.Type {
			case apperrors.ErrorTypeFatal:
				summary.Fatal = append(summary.Fatal, err)
			case apperrors.ErrorTypeRetryable:
				summary.Retryable = append(summary.Retryable, err)
			case apperrors.ErrorTypeValidation:
				summary.Validation = append(summary.Validation, err)
			case apperrors.ErrorTypeSkippable:
				summary.Skippable = append(summary.Skippable, err)
			}

			// Extract file context if available
			if file, ok := appErr.Context["file"].(string); ok {
				summary.ByFile[file]++
			}

			// Extract table context if available
			if table, ok := appErr.Context["table"].(string); ok {
				summary.ByTable[table]++
			}
		} else {
			// Non-AppError
			summary.ByType["Other"]++
			summary.Other = append(summary.Other, err)
		}
	}

	return summary
}

// String returns a human-readable summary report
func (es ErrorSummary) String() string {
	if es.TotalErrors == 0 {
		return "No errors collected"
	}

	report := fmt.Sprintf("Error Summary: %d total errors\n\n", es.TotalErrors)

	// By Type
	if len(es.ByType) > 0 {
		report += "By Type:\n"
		for errType, count := range es.ByType {
			report += fmt.Sprintf("  %s: %d\n", errType, count)
		}
		report += "\n"
	}

	// By File
	if len(es.ByFile) > 0 {
		report += "By File:\n"
		for file, count := range es.ByFile {
			report += fmt.Sprintf("  %s: %d errors\n", file, count)
		}
		report += "\n"
	}

	// By Table
	if len(es.ByTable) > 0 {
		report += "By Table:\n"
		for table, count := range es.ByTable {
			report += fmt.Sprintf("  %s: %d errors\n", table, count)
		}
		report += "\n"
	}

	return report
}
