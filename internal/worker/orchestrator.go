package worker

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/database"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"
	"github.com/jmoiron/sqlx"
)

// ProgressTracker verfolgt den Fortschritt des Imports
type ProgressTracker struct {
	parsed   atomic.Int32
	inserted atomic.Int32
	failed   atomic.Int32
	total    int
}

// NewProgressTracker erstellt einen neuen ProgressTracker
func NewProgressTracker(total int) *ProgressTracker {
	return &ProgressTracker{total: total}
}

// IncrementParsed erhöht den Parse-Counter
func (p *ProgressTracker) IncrementParsed() {
	p.parsed.Add(1)
}

// IncrementInserted erhöht den Insert-Counter
func (p *ProgressTracker) IncrementInserted() {
	p.inserted.Add(1)
}

// IncrementFailed erhöht den Failed-Counter
func (p *ProgressTracker) IncrementFailed() {
	p.failed.Add(1)
}

// GetProgress gibt die aktuellen Zähler zurück
func (p *ProgressTracker) GetProgress() (parsed, inserted, failed, total int) {
	return int(p.parsed.Load()), int(p.inserted.Load()), int(p.failed.Load()), p.total
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

		progress.IncrementInserted()
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
