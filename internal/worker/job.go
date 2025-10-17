package worker

import (
	"context"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"
)

// JobExecutor repräsentiert eine ausführbare Aufgabe im Worker Pool.
//
// Dies ist eine Interface-basierte Abstraktion für typisierte Jobs, die
// eine Alternative zum funktionalen Job-Ansatz bietet. JobExecutor ermöglicht
// strukturierte Jobs mit Typsicherheit.
//
// Implementierende Typen (z.B. ParseJob, InsertJob) kapseln ihre
// Abhängigkeiten als Struct-Felder.
//
// Beispiel:
//
//	type CustomJob struct {
//	    Config Config
//	    Input  string
//	}
//
//	func (j *CustomJob) Execute(ctx context.Context) (JobResult, error) {
//	    result := processInput(j.Config, j.Input)
//	    return CustomResult{Data: result}, nil
//	}
type JobExecutor interface {
	Execute(ctx context.Context) (JobResult, error)
}

// JobResult repräsentiert das Ergebnis einer Job-Ausführung.
//
// Dies ist ein Marker-Interface für verschiedene Result-Typen.
// Implementierende Typen (z.B. ParseResult, InsertResult) können
// spezifische Daten zurückgeben, die vom Aufrufer typsicher
// ausgewertet werden können.
//
// Beispiel:
//
//	type CustomResult struct {
//	    ProcessedCount int
//	    Output         string
//	}
//
//	func (CustomResult) isJobResult() {}
type JobResult interface {
	isJobResult()
}

// ParseJob repräsentiert einen JSONL-Parse-Job.
//
// ParseJob wird für die parallele Phase 1 des EVE SDE Imports verwendet,
// bei der JSONL-Dateien parallel geparst werden. Der Job kapselt einen
// Parser und den Dateipfad.
//
// Beispiel:
//
//	job := &ParseJob{
//	    Parser:   parser.NewTypesParser(),
//	    FilePath: "types.jsonl",
//	}
//	result, err := job.Execute(ctx)
type ParseJob struct {
	Parser   parser.Parser // Parser-Implementierung für das File-Format
	FilePath string        // Pfad zur zu parsenden JSONL-Datei
}

// Execute führt den Parse-Job aus.
//
// Execute ruft die ParseFile-Methode des Parsers auf und gibt ein
// ParseResult mit den geparsten Items zurück. Context-Cancellation
// wird respektiert.
func (j *ParseJob) Execute(ctx context.Context) (JobResult, error) {
	items, err := j.Parser.ParseFile(ctx, j.FilePath)
	return ParseResult{Items: items}, err
}

// InsertJob repräsentiert einen Datenbank-Insert-Job.
//
// InsertJob wird für die sequentielle Phase 2 des EVE SDE Imports verwendet,
// bei der geparste Daten in die SQLite-Datenbank eingefügt werden.
// Diese Phase respektiert das 1-Writer-Constraint von SQLite.
//
// Hinweis: Die Implementierung ist derzeit ein Platzhalter für zukünftige
// Features (siehe TODO).
type InsertJob struct {
	Table string        // Ziel-Tabelle für Insert
	Rows  []interface{} // Einzufügende Zeilen
}

// Execute führt den Insert-Job aus.
//
// Hinweis: Dies ist eine Platzhalter-Implementierung für Phase 2 (DB-Insert).
// Die vollständige Implementierung wird in einem späteren Issue ergänzt.
//
// TODO: Implementierung der DB-Insert-Logik (sequenzieller Writer)
func (j *InsertJob) Execute(ctx context.Context) (JobResult, error) {
	// TODO: Implementierung der DB-Insert-Logik
	// Dies wird in einem späteren Issue implementiert (sequenzieller Writer)
	return InsertResult{RowsAffected: len(j.Rows)}, nil
}

// ParseResult repräsentiert das Ergebnis eines Parse-Jobs.
//
// ParseResult enthält die geparsten Items aus einer JSONL-Datei.
// Die Items sind als []interface{} repräsentiert, um verschiedene
// Parser-Typen zu unterstützen.
type ParseResult struct {
	Items []interface{} // Geparste Items aus der JSONL-Datei
}

// isJobResult markiert ParseResult als JobResult-Implementierung.
func (ParseResult) isJobResult() {}

// InsertResult repräsentiert das Ergebnis eines Insert-Jobs.
//
// InsertResult gibt Auskunft über die Anzahl der eingefügten Zeilen
// in der Datenbank.
type InsertResult struct {
	RowsAffected int // Anzahl der eingefügten Zeilen
}

// isJobResult markiert InsertResult als JobResult-Implementierung.
func (InsertResult) isJobResult() {}
