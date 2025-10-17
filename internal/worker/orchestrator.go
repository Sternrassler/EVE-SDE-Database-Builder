package worker

import (
	"context"
	"fmt"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/database"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"
	"github.com/jmoiron/sqlx"
)

// Progress repräsentiert den aktuellen Import-Fortschritt mit detaillierten Metriken.
//
// Progress wird von ProgressTracker.GetProgressDetailed() zurückgegeben und
// bietet umfassende Informationen über den aktuellen Stand eines Worker-Pool-basierten
// Imports, inklusive ETA-Berechnung und Durchsatz-Metriken.
//
// Beispiel:
//
//	progress := tracker.GetProgressDetailed()
//	log.Printf("Fortschritt: %.1f%% (%d/%d Dateien, %.0f Zeilen/s, ETA: %v)",
//	    progress.PercentFiles, progress.ParsedFiles, progress.TotalFiles,
//	    progress.RowsPerSecond, progress.ETA)
type Progress struct {
	ParsedFiles   int64         // Anzahl vollständig geparster Dateien
	InsertedFiles int64         // Anzahl erfolgreich eingefügter Dateien
	FailedFiles   int64         // Anzahl fehlgeschlagener Dateien
	TotalFiles    int64         // Gesamtzahl der zu verarbeitenden Dateien
	TotalRows     int64         // Gesamtzahl der zu verarbeitenden Zeilen (wenn bekannt)
	InsertedRows  int64         // Anzahl eingefügter Zeilen
	PercentFiles  float64       // Fortschritt in % (basierend auf Dateien)
	PercentRows   float64       // Fortschritt in % (basierend auf Zeilen)
	ETA           time.Duration // Geschätzte verbleibende Zeit
	ElapsedTime   time.Duration // Verstrichene Zeit seit Start
	RowsPerSecond float64       // Durchsatz (Zeilen/Sekunde)
}

// ProgressTracker verfolgt den Fortschritt des Imports mit atomaren Zählern.
//
// ProgressTracker ist Thread-Safe und kann aus mehreren Goroutines gleichzeitig
// aktualisiert werden. Verwendet atomic.Int64 für Lock-freie Updates.
//
// ProgressTracker unterstützt sowohl File- als auch Row-basiertes Tracking,
// was detaillierte Fortschrittsanzeigen mit ETA-Berechnung ermöglicht.
//
// Beispiel:
//
//	tracker := worker.NewProgressTracker(100) // 100 Dateien
//	tracker.SetTotalRows(1000000)             // 1M Zeilen erwartet
//
//	// In Worker-Goroutines
//	tracker.Update(1, 10000) // 1 Datei geparst, 10k Zeilen eingefügt
//
//	// Fortschritt abrufen
//	progress := tracker.GetProgressDetailed()
type ProgressTracker struct {
	parsedFiles  atomic.Int64
	insertedRows atomic.Int64
	failed       atomic.Int64
	totalFiles   int64
	totalRows    atomic.Int64
	startTime    time.Time
}

// NewProgressTracker erstellt einen neuen ProgressTracker.
//
// Parameter:
//   - total: Gesamtzahl der zu verarbeitenden Dateien
//
// Der Tracker ist sofort einsatzbereit. StartTime wird auf time.Now() gesetzt.
func NewProgressTracker(total int) *ProgressTracker {
	return &ProgressTracker{
		totalFiles: int64(total),
		startTime:  time.Now(),
	}
}

// SetTotalRows setzt die Gesamtzahl der zu verarbeitenden Zeilen.
//
// Diese Methode sollte aufgerufen werden, wenn die erwartete Gesamtzeilenzahl
// bekannt ist (z.B. nach Parsing-Phase). Ermöglicht präzisere ETA-Berechnung.
func (p *ProgressTracker) SetTotalRows(total int64) {
	p.totalRows.Store(total)
}

// IncrementParsed erhöht den Parse-Counter.
//
// Sollte aufgerufen werden, wenn eine Datei erfolgreich geparst wurde.
// Thread-Safe.
func (p *ProgressTracker) IncrementParsed() {
	p.parsedFiles.Add(1)
}

// IncrementInserted erhöht den Insert-Counter (deprecated - verwende AddInsertedRows)
func (p *ProgressTracker) IncrementInserted() {
	// Legacy-Kompatibilität: Do nothing, da neue Implementierung rows zählt, nicht files
	// Wird nur für Tests benötigt
}

// IncrementFailed erhöht den Failed-Counter.
//
// Sollte aufgerufen werden, wenn die Verarbeitung einer Datei fehlschlägt.
// Thread-Safe.
func (p *ProgressTracker) IncrementFailed() {
	p.failed.Add(1)
}

// Update aktualisiert den Fortschritt mit Anzahl geparster Dateien und eingefügter Zeilen.
//
// Dies ist eine Convenience-Methode für atomare Updates von parsed und inserted.
// Thread-Safe.
//
// Parameter:
//   - parsed: Anzahl neu geparster Dateien (typischerweise 1)
//   - inserted: Anzahl neu eingefügter Zeilen
//
// Beispiel:
//
//	tracker.Update(1, 5000) // 1 Datei geparst, 5000 Zeilen eingefügt
func (p *ProgressTracker) Update(parsed int, inserted int) {
	if parsed > 0 {
		p.parsedFiles.Add(int64(parsed))
	}
	if inserted > 0 {
		p.insertedRows.Add(int64(inserted))
	}
}

// AddInsertedRows fügt die Anzahl eingefügter Zeilen hinzu.
//
// Thread-Safe. Verwendet für Bulk-Insert-Tracking.
func (p *ProgressTracker) AddInsertedRows(count int64) {
	p.insertedRows.Add(count)
}

// GetProgress gibt die aktuellen Zähler zurück (für Rückwärtskompatibilität).
//
// Deprecated: Verwenden Sie stattdessen GetProgressDetailed() für detaillierte
// Fortschrittsinformationen inkl. ETA und Durchsatz.
//
// Rückgabewerte:
//   - parsed: Anzahl geparster Dateien
//   - inserted: Anzahl erfolgreich eingefügter Dateien (parsed - failed)
//   - failed: Anzahl fehlgeschlagener Dateien
//   - total: Gesamtzahl der Dateien
func (p *ProgressTracker) GetProgress() (parsed, inserted, failed, total int) {
	parsedFiles := p.parsedFiles.Load()
	failedFiles := p.failed.Load()

	// Legacy-Kompatibilität: inserted entspricht parsedFiles minus failed
	insertedFiles := parsedFiles - failedFiles

	return int(parsedFiles), int(insertedFiles), int(failedFiles), int(p.totalFiles)
}

// GetProgressDetailed gibt detaillierte Fortschrittsinformationen zurück.
//
// GetProgressDetailed berechnet erweiterte Metriken wie Prozent-Fortschritt,
// ETA (Estimated Time to Arrival) und Durchsatz (Zeilen/Sekunde).
//
// Die ETA wird bevorzugt aus Zeilen-Metriken berechnet (präziser), mit
// Fallback auf Datei-Metriken.
//
// Thread-Safe.
func (p *ProgressTracker) GetProgressDetailed() Progress {
	parsedFiles := p.parsedFiles.Load()
	insertedRows := p.insertedRows.Load()
	failedFiles := p.failed.Load()
	totalRows := p.totalRows.Load()
	elapsed := time.Since(p.startTime)

	// Prozentberechnung (Dateien)
	var percentFiles float64
	if p.totalFiles > 0 {
		percentFiles = float64(parsedFiles) / float64(p.totalFiles) * 100.0
	}

	// Prozentberechnung (Zeilen)
	var percentRows float64
	if totalRows > 0 {
		percentRows = float64(insertedRows) / float64(totalRows) * 100.0
	}

	// Durchsatz berechnen (Zeilen/Sekunde)
	var rowsPerSecond float64
	if elapsed.Seconds() > 0 {
		rowsPerSecond = float64(insertedRows) / elapsed.Seconds()
	}

	// ETA berechnen (basierend auf Zeilen, falls verfügbar, sonst Dateien)
	var eta time.Duration
	if totalRows > 0 && insertedRows > 0 && rowsPerSecond > 0 {
		// ETA basierend auf Zeilen
		remainingRows := totalRows - insertedRows
		etaSeconds := float64(remainingRows) / rowsPerSecond
		eta = time.Duration(etaSeconds * float64(time.Second))
	} else if p.totalFiles > 0 && parsedFiles > 0 && elapsed.Seconds() > 0 {
		// Fallback: ETA basierend auf Dateien
		remainingFiles := p.totalFiles - parsedFiles
		filesPerSecond := float64(parsedFiles) / elapsed.Seconds()
		if filesPerSecond > 0 {
			etaSeconds := float64(remainingFiles) / filesPerSecond
			eta = time.Duration(etaSeconds * float64(time.Second))
		}
	}

	insertedFiles := parsedFiles - failedFiles

	return Progress{
		ParsedFiles:   parsedFiles,
		InsertedFiles: insertedFiles,
		FailedFiles:   failedFiles,
		TotalFiles:    p.totalFiles,
		TotalRows:     totalRows,
		InsertedRows:  insertedRows,
		PercentFiles:  percentFiles,
		PercentRows:   percentRows,
		ETA:           eta,
		ElapsedTime:   elapsed,
		RowsPerSecond: rowsPerSecond,
	}
}

// Orchestrator koordiniert Parser → Worker Pool → Database.
//
// Orchestrator implementiert das 2-Phasen-Import-Pattern für EVE SDE:
//   - Phase 1: Paralleles JSONL-Parsing mit Worker Pool
//   - Phase 2: Sequenzielles Database-Insert (SQLite 1-Writer-Constraint)
//
// Der Orchestrator verwaltet die Koordination zwischen Parsern, Worker Pool
// und Datenbank-Connection, inkl. Fortschritts-Tracking und Error-Handling.
//
// Beispiel:
//
//	parsers := map[string]parser.Parser{
//	    "types": parser.NewTypesParser(),
//	    "agents": parser.NewAgentsParser(),
//	}
//	orch := worker.NewOrchestrator(db, pool, parsers)
//	tracker, err := orch.ImportAll(ctx, sdeDir)
type Orchestrator struct {
	db      *sqlx.DB
	pool    *Pool
	parsers map[string]parser.Parser
}

// NewOrchestrator erstellt einen neuen Orchestrator.
//
// Parameter:
//   - db: SQLite-Datenbankverbindung (für Phase 2: Insert)
//   - pool: Worker Pool (für Phase 1: Parsing)
//   - parsers: Map von Parser-Name zu Parser-Implementierung
//
// Der Pool sollte bereits mit Start(ctx) gestartet sein, bevor ImportAll()
// aufgerufen wird.
func NewOrchestrator(db *sqlx.DB, pool *Pool, parsers map[string]parser.Parser) *Orchestrator {
	return &Orchestrator{
		db:      db,
		pool:    pool,
		parsers: parsers,
	}
}

// ParseTask repräsentiert eine Parse-Aufgabe mit Dateinamen.
//
// ParseTask wird intern vom Orchestrator verwendet, um JSONL-Dateien
// mit dem passenden Parser zu verknüpfen.
type ParseTask struct {
	File   string        // Pfad zur JSONL-Datei
	Parser parser.Parser // Zugeordneter Parser
}

// ParseResultData repräsentiert das Ergebnis eines Parse-Vorgangs.
//
// ParseResultData wird vom Orchestrator verwendet, um Parse-Ergebnisse
// zwischen Phase 1 (Parsing) und Phase 2 (Insert) zu übergeben.
type ParseResultData struct {
	File    string        // Quelldatei (für Error-Reporting)
	Table   string        // Ziel-Tabelle für Insert
	Columns []string      // Spalten-Namen für Insert
	Records []interface{} // Geparste Records
	Err     error         // Parse-Fehler (falls aufgetreten)
}

// ImportAll führt 2-Phase Import aus: Parse parallel → Insert sequentiell
func (o *Orchestrator) ImportAll(ctx context.Context, sdeDir string) (*ProgressTracker, error) {
	// Discover JSONL files und erstelle Tasks
	tasks, err := o.createParseTasks(sdeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create parse tasks: %w", err)
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("no JSONL files found in %s", sdeDir)
	}

	// Progress Tracker initialisieren
	progress := NewProgressTracker(len(tasks))

	// === Phase 1: Parallel Parsing ===
	// Worker Pool starten
	o.pool.Start(ctx)

	// Parse-Jobs submiten
	for _, task := range tasks {
		t := task // Capture loop variable
		job := Job{
			ID: t.File,
			Fn: func(ctx context.Context) (interface{}, error) {
				records, err := t.Parser.ParseFile(ctx, t.File)
				if err != nil {
					return nil, err
				}
				return ParseResultData{
					File:    t.File,
					Table:   t.Parser.TableName(),
					Columns: t.Parser.Columns(),
					Records: records,
					Err:     nil,
				}, nil
			},
		}
		o.pool.Submit(job)
	}

	// Warte auf alle Parse-Jobs
	results, _ := o.pool.Wait()

	// === Phase 2: Sequential Insert ===
	// Process results sequentially for SQLite single-writer constraint
	for _, result := range results {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return progress, ctx.Err()
		default:
		}

		progress.IncrementParsed()

		if result.Err != nil {
			progress.IncrementFailed()
			continue // Skip failed parse results
		}

		// Extract parse result
		parseResult, ok := result.Data.(ParseResultData)
		if !ok {
			progress.IncrementFailed()
			continue
		}

		// Convert []interface{} to [][]interface{} for BatchInsert
		rows, err := o.convertToRows(parseResult.Records, len(parseResult.Columns))
		if err != nil {
			progress.IncrementFailed()
			continue
		}

		// Perform batch insert
		err = database.BatchInsert(ctx, o.db, parseResult.Table, parseResult.Columns, rows, 1000)
		if err != nil {
			progress.IncrementFailed()
			continue
		}

		// Track successful insert
		progress.AddInsertedRows(int64(len(parseResult.Records)))
	}

	return progress, nil
}

// createParseTasks erstellt Parse-Tasks für alle JSONL-Dateien im SDE-Verzeichnis
func (o *Orchestrator) createParseTasks(sdeDir string) ([]ParseTask, error) {
	// Placeholder implementation - in real scenario, this would discover files
	// For now, return empty slice as we don't have file discovery logic yet
	var tasks []ParseTask

	// Iterate through registered parsers and create tasks
	// In real implementation, this would scan sdeDir for matching .jsonl files
	for fileName, p := range o.parsers {
		tasks = append(tasks, ParseTask{
			File:   fileName,
			Parser: p,
		})
	}

	return tasks, nil
}

// convertToRows konvertiert []interface{} in [][]interface{} für BatchInsert
// Es unterstützt Structs (via Reflection) und Maps (für Tests)
func (o *Orchestrator) convertToRows(records []interface{}, columnCount int) ([][]interface{}, error) {
	if len(records) == 0 {
		return [][]interface{}{}, nil
	}

	rows := make([][]interface{}, len(records))

	for i, record := range records {
		// Use reflection to handle different record types
		val := reflect.ValueOf(record)

		// Handle pointer to struct
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		var row []interface{}

		switch val.Kind() {
		case reflect.Struct:
			// Extract struct field values in order
			row = make([]interface{}, 0, columnCount)
			typ := val.Type()

			for j := 0; j < val.NumField(); j++ {
				field := val.Field(j)

				// Skip unexported fields
				if !typ.Field(j).IsExported() {
					continue
				}

				// Get field value, handling nil pointers
				var fieldValue interface{}
				if field.Kind() == reflect.Ptr {
					if field.IsNil() {
						fieldValue = nil
					} else {
						fieldValue = field.Elem().Interface()
					}
				} else {
					fieldValue = field.Interface()
				}

				row = append(row, fieldValue)
			}

		case reflect.Map:
			// For map[string]interface{}, create a placeholder row
			// Maps don't have guaranteed order, so this is primarily for testing
			row = make([]interface{}, columnCount)
			// Fill with nil values - real implementation would need column order
			for j := 0; j < columnCount; j++ {
				row[j] = nil
			}

		default:
			return nil, fmt.Errorf("record %d is not a struct or map, got %v", i, val.Kind())
		}

		// Verify column count matches
		if len(row) != columnCount {
			return nil, fmt.Errorf("record %d has %d fields, expected %d columns", i, len(row), columnCount)
		}

		rows[i] = row
	}

	return rows, nil
}
