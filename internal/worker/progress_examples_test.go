// Package worker Progress Tracker - Usage Examples
//
// Dieses Dokument zeigt verschiedene Anwendungsfälle des erweiterten ProgressTracker.

package worker_test

import (
	"fmt"
	"time"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/worker"
)

// Example_progressTrackerBasic zeigt die grundlegende Verwendung
func Example_progressTrackerBasic() {
	// Erstelle Tracker für 10 Dateien
	tracker := worker.NewProgressTracker(10)
	tracker.SetTotalRows(1000)

	// Simuliere Fortschritt: 5 Dateien, 500 Zeilen
	tracker.Update(5, 500)

	// Hole detaillierten Fortschritt
	progress := tracker.GetProgressDetailed()

	fmt.Printf("Dateien: %d/%d (%.1f%%)\n",
		progress.ParsedFiles, progress.TotalFiles, progress.PercentFiles)
	fmt.Printf("Zeilen: %d/%d (%.1f%%)\n",
		progress.InsertedRows, progress.TotalRows, progress.PercentRows)

	// Output:
	// Dateien: 5/10 (50.0%)
	// Zeilen: 500/1000 (50.0%)
}

// Example_progressTrackerETA zeigt ETA-Berechnung
func Example_progressTrackerETA() {
	tracker := worker.NewProgressTracker(100)
	tracker.SetTotalRows(10000)

	// Simuliere 50% Fortschritt
	tracker.Update(50, 5000)

	progress := tracker.GetProgressDetailed()

	// ETA wird basierend auf aktuellem Durchsatz berechnet
	fmt.Printf("Fortschritt: %.0f%% (Zeilen)\n", progress.PercentRows)
	fmt.Printf("ETA verfügbar: %v\n", progress.ETA >= 0)

	// Output:
	// Fortschritt: 50% (Zeilen)
	// ETA verfügbar: true
}

// Example_progressTrackerConcurrent zeigt Thread-Safe Updates
func Example_progressTrackerConcurrent() {
	tracker := worker.NewProgressTracker(20)
	tracker.SetTotalRows(2000)

	// Simuliere 4 parallele Worker
	done := make(chan bool)
	for i := 0; i < 4; i++ {
		go func(workerID int) {
			for j := 0; j < 5; j++ {
				// Jeder Worker verarbeitet 5 Dateien à 100 Zeilen
				tracker.Update(1, 100)
				time.Sleep(1 * time.Millisecond)
			}
			done <- true
		}(i)
	}

	// Warte auf alle Worker
	for i := 0; i < 4; i++ {
		<-done
	}

	progress := tracker.GetProgressDetailed()
	fmt.Printf("Verarbeitet: %d Dateien, %d Zeilen\n",
		progress.ParsedFiles, progress.InsertedRows)

	// Output:
	// Verarbeitet: 20 Dateien, 2000 Zeilen
}

// Example_progressTrackerIncrementalUpdates zeigt schrittweise Updates
func Example_progressTrackerIncrementalUpdates() {
	tracker := worker.NewProgressTracker(5)

	// Datei 1: Parse + Insert erfolgreich
	tracker.IncrementParsed()
	tracker.AddInsertedRows(100)

	// Datei 2: Parse + Insert erfolgreich
	tracker.IncrementParsed()
	tracker.AddInsertedRows(150)

	// Datei 3: Parse erfolgreich, Insert fehlgeschlagen
	tracker.IncrementParsed()
	tracker.IncrementFailed()

	// Datei 4: Parse + Insert erfolgreich
	tracker.Update(1, 200)

	progress := tracker.GetProgressDetailed()

	fmt.Printf("Parsed: %d, Inserted: %d, Failed: %d\n",
		progress.ParsedFiles, progress.InsertedFiles, progress.FailedFiles)
	fmt.Printf("Total Rows Inserted: %d\n", progress.InsertedRows)

	// Output:
	// Parsed: 4, Inserted: 3, Failed: 1
	// Total Rows Inserted: 450
}

// Example_progressTrackerLegacyCompatibility zeigt Kompatibilität mit alter API
func Example_progressTrackerLegacyCompatibility() {
	tracker := worker.NewProgressTracker(10)

	// Legacy-Methoden (für Rückwärtskompatibilität)
	tracker.IncrementParsed()
	tracker.IncrementParsed()
	tracker.IncrementFailed()

	// Legacy GetProgress() Aufruf
	parsed, inserted, failed, total := tracker.GetProgress()

	fmt.Printf("Legacy API: %d/%d parsed, %d inserted, %d failed\n",
		parsed, total, inserted, failed)

	// Neue API liefert detailliertere Informationen
	progress := tracker.GetProgressDetailed()
	fmt.Printf("Enhanced API: %.1f%% complete, %d inserted files\n",
		progress.PercentFiles, progress.InsertedFiles)

	// Output:
	// Legacy API: 2/10 parsed, 1 inserted, 1 failed
	// Enhanced API: 20.0% complete, 1 inserted files
}

// Example_progressTrackerRealWorldScenario zeigt realistisches Szenario
func Example_progressTrackerRealWorldScenario() {
	// EVE SDE Import: 50 JSONL-Dateien, ~500k Zeilen total
	tracker := worker.NewProgressTracker(50)
	tracker.SetTotalRows(500000)

	// Simuliere Import-Fortschritt
	// Phase 1: Parsing (parallel)
	for i := 0; i < 25; i++ {
		tracker.IncrementParsed()
	}

	// Phase 2: Insert (sequentiell, mit Zeilen-Tracking)
	tracker.AddInsertedRows(250000) // 50% der Zeilen eingefügt

	progress := tracker.GetProgressDetailed()

	fmt.Printf("Import Status:\n")
	fmt.Printf("  Dateien: %d/%d (%.0f%%)\n",
		progress.ParsedFiles, progress.TotalFiles, progress.PercentFiles)
	fmt.Printf("  Zeilen: %d/%d (%.0f%%)\n",
		progress.InsertedRows, progress.TotalRows, progress.PercentRows)

	// Output:
	// Import Status:
	//   Dateien: 25/50 (50%)
	//   Zeilen: 250000/500000 (50%)
}
