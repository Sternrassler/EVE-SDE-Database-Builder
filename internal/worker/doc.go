// Package worker provides a worker pool implementation for parallel JSONL processing.
//
// Das worker-Package implementiert ein Worker Pool Pattern für die parallele Verarbeitung
// von JSONL-Dateien in der EVE SDE Database Builder Anwendung.
//
// # Grundkonzept
//
// Der Worker Pool verarbeitet Tasks parallel mit einer konfigurierbaren Anzahl von Workers.
// Dies ermöglicht effiziente CPU-Auslastung während der JSONL-Parse-Phase, bevor die Daten
// sequenziell in SQLite geschrieben werden (1-Writer-Constraint).
//
// # Architektur (2-Phasen-Import)
//
// Phase 1: JSONL Parse (parallel mit Worker Pool)
// Phase 2: DB Insert (sequenziell via SQLite Writer)
//
// # Pool-Erstellung
//
//	pool := worker.NewPool(4) // 4 parallele Workers
//	pool.Start(ctx)
//
// # Job-Submission
//
//	job := worker.Job{
//	    ID: "types.jsonl",
//	    Fn: func(ctx context.Context) (interface{}, error) {
//	        return parseFile(ctx, "types.jsonl")
//	    },
//	}
//	pool.Submit(job)
//
// # Graceful Shutdown
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	pool.Start(ctx)
//	// ... Submit jobs ...
//
//	// Signal Cancellation
//	cancel() // Workers stoppen nach aktuellem Job
//
// # Error Collection
//
// Der Pool sammelt Fehler von allen Jobs und gibt sie gesammelt zurück:
//
//	results, errors := pool.Wait()
//	if len(errors) > 0 {
//	    // Handle errors
//	}
//
// # Context-Unterstützung
//
// Alle Worker respektieren Context-Cancellation:
//   - Bei ctx.Done() brechen Workers nach aktuellem Job ab
//   - Keine Hard Aborts (kein Thread.Abort wie in VB.NET)
//   - Jobs können selbst auf Context-Cancellation reagieren
//
// # Channel-basierte Kommunikation
//
// Der Pool verwendet buffered Channels für Backpressure:
//   - jobs Channel: Verteilt Tasks an Workers
//   - results Channel: Sammelt Ergebnisse von Workers
//   - Buffering verhindert Blockierung bei Submit/Collect
//
// # Best Practices
//
//   - Verwenden Sie runtime.NumCPU() für optimale Worker-Anzahl
//   - Setzen Sie angemessene Context-Timeouts für lange Jobs
//   - Sammeln Sie Results kontinuierlich (non-blocking)
//   - Nutzen Sie Close() + Wait() für sauberes Shutdown
//
// Siehe auch:
//   - ADR-006: Concurrency & Worker Pool Pattern
//   - internal/parser für JSONL-Parsing
//   - internal/database für SQLite-Integration
package worker
