package worker

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/database"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"
	"github.com/jmoiron/sqlx"
)

// Progress repräsentiert den aktuellen Import-Fortschritt mit detaillierten Metriken
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

// ProgressTracker verfolgt den Fortschritt des Imports mit atomaren Zählern
type ProgressTracker struct {
	parsedFiles  atomic.Int64
	insertedRows atomic.Int64
	failed       atomic.Int64
	totalFiles   int64
	totalRows    atomic.Int64
	startTime    time.Time
}

// NewProgressTracker erstellt einen neuen ProgressTracker
func NewProgressTracker(total int) *ProgressTracker {
	return &ProgressTracker{
		totalFiles: int64(total),
		startTime:  time.Now(),
	}
}

// SetTotalRows setzt die Gesamtzahl der zu verarbeitenden Zeilen
func (p *ProgressTracker) SetTotalRows(total int64) {
	p.totalRows.Store(total)
}

// IncrementParsed erhöht den Parse-Counter
func (p *ProgressTracker) IncrementParsed() {
	p.parsedFiles.Add(1)
}

// IncrementInserted erhöht den Insert-Counter (deprecated - verwende AddInsertedRows)
func (p *ProgressTracker) IncrementInserted() {
	// Legacy-Kompatibilität: Do nothing, da neue Implementierung rows zählt, nicht files
	// Wird nur für Tests benötigt
}

// IncrementFailed erhöht den Failed-Counter
func (p *ProgressTracker) IncrementFailed() {
	p.failed.Add(1)
}

// Update aktualisiert den Fortschritt mit Anzahl geparster Dateien und eingefügter Zeilen
// parsed: Anzahl neu geparster Dateien (typischerweise 1)
// inserted: Anzahl neu eingefügter Zeilen
func (p *ProgressTracker) Update(parsed int, inserted int) {
	if parsed > 0 {
		p.parsedFiles.Add(int64(parsed))
	}
	if inserted > 0 {
		p.insertedRows.Add(int64(inserted))
	}
}

// AddInsertedRows fügt die Anzahl eingefügter Zeilen hinzu
func (p *ProgressTracker) AddInsertedRows(count int64) {
	p.insertedRows.Add(count)
}

// GetProgress gibt die aktuellen Zähler zurück (für Rückwärtskompatibilität)
func (p *ProgressTracker) GetProgress() (parsed, inserted, failed, total int) {
	parsedFiles := p.parsedFiles.Load()
	failedFiles := p.failed.Load()

	// Legacy-Kompatibilität: inserted entspricht parsedFiles minus failed
	insertedFiles := parsedFiles - failedFiles

	return int(parsedFiles), int(insertedFiles), int(failedFiles), int(p.totalFiles)
}

// GetProgressDetailed gibt detaillierte Fortschrittsinformationen zurück
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

// Orchestrator koordiniert Parser → Worker Pool → Database
type Orchestrator struct {
	db      *sqlx.DB
	pool    *Pool
	parsers map[string]parser.Parser
}

// NewOrchestrator erstellt einen neuen Orchestrator
func NewOrchestrator(db *sqlx.DB, pool *Pool, parsers map[string]parser.Parser) *Orchestrator {
	return &Orchestrator{
		db:      db,
		pool:    pool,
		parsers: parsers,
	}
}

// ParseTask repräsentiert eine Parse-Aufgabe mit Dateinamen
type ParseTask struct {
	File   string
	Parser parser.Parser
}

// ParseResult repräsentiert das Ergebnis eines Parse-Vorgangs
type ParseResultData struct {
	File    string
	Table   string
	Columns []string
	Records []interface{}
	Err     error
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
func (o *Orchestrator) convertToRows(records []interface{}, columnCount int) ([][]interface{}, error) {
	if len(records) == 0 {
		return [][]interface{}{}, nil
	}

	rows := make([][]interface{}, len(records))
	for i := range records {
		// Each record should be a map or struct that can be converted to a row
		// For now, we create a placeholder row
		row := make([]interface{}, columnCount)
		// TODO: Implement actual conversion logic based on record type
		// This is a placeholder that needs proper implementation
		rows[i] = row
	}

	return rows, nil
}
