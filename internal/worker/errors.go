package worker

import (
	"fmt"
	"sync"

	apperrors "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"
)

// ErrorCollector collects errors from multiple workers in a thread-safe manner.
//
// ErrorCollector ermöglicht das sichere Sammeln von Fehlern aus parallelen
// Worker-Goroutines. Alle Operationen sind Thread-Safe und können ohne
// externe Synchronisierung aus mehreren Goroutines aufgerufen werden.
//
// ErrorCollector unterstützt erweiterte Fehleranalyse über Summary(),
// das Fehler nach Typ, Datei und Tabelle gruppiert.
//
// Beispiel:
//
//	ec := worker.NewErrorCollector()
//	var wg sync.WaitGroup
//	for i := 0; i < 10; i++ {
//	    wg.Add(1)
//	    go func() {
//	        defer wg.Done()
//	        if err := processItem(); err != nil {
//	            ec.Collect(err)
//	        }
//	    }()
//	}
//	wg.Wait()
//	errors := ec.GetErrors()
type ErrorCollector struct {
	errors []error
	mu     sync.Mutex
}

// NewErrorCollector creates a new ErrorCollector.
//
// Der zurückgegebene ErrorCollector ist sofort einsatzbereit und
// Thread-Safe für parallele Collect-Aufrufe.
func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{
		errors: make([]error, 0),
	}
}

// Collect adds an error to the collection in a thread-safe manner.
//
// nil-Fehler werden automatisch ignoriert. Collect kann sicher aus
// mehreren Goroutines gleichzeitig aufgerufen werden.
//
// Beispiel:
//
//	if err := processFile(file); err != nil {
//	    ec.Collect(err)
//	}
func (ec *ErrorCollector) Collect(err error) {
	if err == nil {
		return
	}
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.errors = append(ec.errors, err)
}

// GetErrors returns all collected errors.
//
// GetErrors gibt eine Kopie des internen Error-Slices zurück, um
// externe Modifikationen zu verhindern. Die Methode ist Thread-Safe.
//
// Rückgabewert:
//   - []error: Kopie aller gesammelten Fehler (leeres Slice wenn keine Fehler)
func (ec *ErrorCollector) GetErrors() []error {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	// Return a copy to prevent external modification
	result := make([]error, len(ec.errors))
	copy(result, ec.errors)
	return result
}

// ErrorSummary provides a summary of collected errors grouped by type and context.
//
// ErrorSummary aggregiert Fehler nach verschiedenen Kriterien und ermöglicht
// detaillierte Fehleranalyse. Die Gruppierung basiert auf AppError-Metadaten
// (Type, Context).
//
// Felder:
//   - TotalErrors: Gesamtzahl aller gesammelten Fehler
//   - ByType: Fehler gruppiert nach ErrorType (Fatal, Retryable, etc.)
//   - ByFile: Fehler gruppiert nach betroffener Datei (aus Context["file"])
//   - ByTable: Fehler gruppiert nach betroffener Tabelle (aus Context["table"])
//   - Fatal, Retryable, Validation, Skippable: Fehler nach Kategorie
//   - Other: Nicht-AppError Fehler
type ErrorSummary struct {
	TotalErrors int            // Gesamtzahl aller Fehler
	ByType      map[string]int // Fehler nach ErrorType
	ByFile      map[string]int // Fehler nach Datei
	ByTable     map[string]int // Fehler nach Tabelle
	Fatal       []error        // Fatal-Fehler (kritisch)
	Retryable   []error        // Retryable-Fehler (temporär)
	Validation  []error        // Validation-Fehler (Datenproblem)
	Skippable   []error        // Skippable-Fehler (überspringbar)
	Other       []error        // Sonstige Fehler (nicht-AppError)
}

// Summary generates an ErrorSummary from collected errors.
//
// Summary analysiert alle gesammelten Fehler und erstellt eine
// strukturierte Zusammenfassung mit Gruppierungen und Kategorisierungen.
// Die Methode ist Thread-Safe.
//
// Beispiel:
//
//	summary := ec.Summary()
//	log.Printf("Total: %d, Fatal: %d, Retryable: %d",
//	    summary.TotalErrors, len(summary.Fatal), len(summary.Retryable))
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

// String returns a human-readable summary report.
//
// String formatiert den ErrorSummary als mehrzeiligen Text-Report
// mit Gruppierungen nach Typ, Datei und Tabelle. Ideal für Logging
// und Fehlerberichte.
//
// Beispiel Output:
//
//	Error Summary: 5 total errors
//
//	By Type:
//	  Fatal: 1
//	  Retryable: 2
//	  Skippable: 2
//
//	By File:
//	  types.jsonl: 3 errors
//	  agents.jsonl: 2 errors
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
