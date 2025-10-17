package worker

import (
	"context"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"
)

// JobExecutor repräsentiert eine ausführbare Aufgabe im Worker Pool
// Dies ist die neue Interface-basierte Abstraktion für typisierte Jobs
type JobExecutor interface {
	Execute(ctx context.Context) (JobResult, error)
}

// JobResult repräsentiert das Ergebnis einer Job-Ausführung
// Dies ist ein Marker-Interface für verschiedene Result-Typen
type JobResult interface {
	isJobResult()
}

// ParseJob repräsentiert einen JSONL-Parse-Job
type ParseJob struct {
	Parser   parser.Parser
	FilePath string
}

// Execute führt den Parse-Job aus
func (j *ParseJob) Execute(ctx context.Context) (JobResult, error) {
	items, err := j.Parser.ParseFile(ctx, j.FilePath)
	return ParseResult{Items: items}, err
}

// InsertJob repräsentiert einen Datenbank-Insert-Job
type InsertJob struct {
	Table string
	Rows  []interface{}
}

// Execute führt den Insert-Job aus
// Hinweis: Dies ist eine Platzhalter-Implementierung für Phase 2 (DB-Insert)
func (j *InsertJob) Execute(ctx context.Context) (JobResult, error) {
	// TODO: Implementierung der DB-Insert-Logik
	// Dies wird in einem späteren Issue implementiert (sequenzieller Writer)
	return InsertResult{RowsAffected: len(j.Rows)}, nil
}

// ParseResult repräsentiert das Ergebnis eines Parse-Jobs
type ParseResult struct {
	Items []interface{}
}

// isJobResult markiert ParseResult als JobResult-Implementierung
func (ParseResult) isJobResult() {}

// InsertResult repräsentiert das Ergebnis eines Insert-Jobs
type InsertResult struct {
	RowsAffected int
}

// isJobResult markiert InsertResult als JobResult-Implementierung
func (InsertResult) isJobResult() {}
