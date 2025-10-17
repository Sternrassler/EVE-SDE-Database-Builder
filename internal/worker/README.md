# Worker Package

Worker Pool Implementierung für paralleles JSONL-Parsing gemäß ADR-006.

## Überblick

Das `worker` Package stellt einen konfigurierbaren Worker Pool zur Verfügung, der Jobs parallel mit einer definierten Anzahl von Workers verarbeitet. Es implementiert das Worker Pool Pattern mit Channel-basierter Job-Verteilung und Graceful Shutdown via Context.

**Neu in v0.1.0:** Import Orchestrator für 2-Phasen-Import (Parse parallel → Insert sequentiell)

## Features

- ✅ Konfigurierbare Worker-Anzahl
- ✅ Channel-basierte Job-Verteilung (buffered channels)
- ✅ Graceful Shutdown über `context.Context`
- ✅ Error Collection (alle Job-Fehler werden gesammelt)
- ✅ **Import Orchestrator (2-Phasen-Import)**
- ✅ **Progress Tracking (thread-safe)**
- ✅ 100% Test Coverage
- ✅ Thread-safe

## Verwendung

### Basis-Beispiel

```go
ctx := context.Background()
pool := worker.NewPool(4) // 4 parallele Workers
pool.Start(ctx)

// Jobs submitten
for i := 0; i < 10; i++ {
    jobID := i
    pool.Submit(worker.Job{
        ID: fmt.Sprintf("job-%d", jobID),
        Fn: func(ctx context.Context) (interface{}, error) {
            // Job-Logik hier
            return processData(jobID)
        },
    })
}

// Auf Completion warten
results, errors := pool.Wait()
```

### Context Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

pool := worker.NewPool(2)
pool.Start(ctx)

// Jobs submitten...

// Bei Bedarf canceln (z.B. Ctrl+C)
cancel() // Workers stoppen nach aktuellem Job

results, errors := pool.Wait()
```

### Error Handling

```go
results, errors := pool.Wait()

if len(errors) > 0 {
    for _, err := range errors {
        log.Printf("Job failed: %v", err)
    }
}

// Erfolgreiche Results verarbeiten
for _, result := range results {
    if result.Err == nil {
        // Success
        process(result.Data)
    }
}
```

## API

### `NewPool(workers int) *Pool`

Erstellt einen neuen Worker Pool mit der angegebenen Anzahl von Workers.

- `workers`: Anzahl paralleler Workers (Default: 1 bei workers ≤ 0)

### `Start(ctx context.Context)`

Startet die Worker Goroutines.

- `ctx`: Context für Cancellation

### `Submit(job Job)`

Fügt einen Job zur Warteschlange hinzu.

- `job`: Job mit ID und Fn

### `Wait() ([]Result, []error)`

Wartet bis alle Workers fertig sind und gibt Results und Errors zurück.

Returns:
- `[]Result`: Alle Job-Results (inkl. fehlgeschlagene)
- `[]error`: Nur die Fehler

## Typen

### Job Interface Pattern (Neu in v0.1.0)

#### `JobExecutor` Interface

```go
type JobExecutor interface {
    Execute(ctx context.Context) (JobResult, error)
}
```

Neue Interface-basierte Abstraktion für typisierte Jobs.

#### `ParseJob`

```go
type ParseJob struct {
    Parser   parser.Parser
    FilePath string
}
```

JSONL-Parse-Job für Phase 1 (paralleles Parsing).

**Beispiel:**

```go
job := &worker.ParseJob{
    Parser:   myParser,
    FilePath: "/data/types.jsonl",
}

result, err := job.Execute(ctx)
parseResult := result.(worker.ParseResult)
// parseResult.Items enthält geparste Daten
```

#### `InsertJob`

```go
type InsertJob struct {
    Table string
    Rows  []interface{}
}
```

Datenbank-Insert-Job für Phase 2 (sequenzieller Insert). *Hinweis: Aktuell Platzhalter-Implementierung.*

**Beispiel:**

```go
job := &worker.InsertJob{
    Table: "types",
    Rows:  parsedItems,
}

result, err := job.Execute(ctx)
insertResult := result.(worker.InsertResult)
// insertResult.RowsAffected enthält Anzahl eingefügter Rows
```

#### `JobResult` Interface

```go
type JobResult interface {
    isJobResult()
}
```

Marker-Interface für typisierte Job-Results.

**Implementierungen:**
- `ParseResult` - Enthält `Items []interface{}`
- `InsertResult` - Enthält `RowsAffected int`

### Legacy Job Pattern (Pool-Kompatibilität)

### `Job`

```go
type Job struct {
    ID string
    Fn func(context.Context) (interface{}, error)
}
```

Legacy-Job-Struktur für direkte Pool-Verwendung. Wird weiterhin von `Pool` unterstützt.

### `Result`

```go
type Result struct {
    JobID string
    Data  interface{}
    Err   error
}
```

## Architektur

Der Worker Pool folgt dem 2-Phasen-Import-Ansatz aus ADR-006:

**Phase 1:** JSONL Parse (parallel mit Worker Pool)  
**Phase 2:** DB Insert (sequentiell via SQLite Writer)

```
┌─────────────────────┐
│ Worker 1            │──┐
│ types.jsonl → []Row │  │
└─────────────────────┘  │
                         ├─────────► Results Channel ──► DB Writer
┌─────────────────────┐  │
│ Worker 2            │──┤
│ blueprints.jsonl    │  │
└─────────────────────┘  │
                         │
┌─────────────────────┐  │
│ Worker 3            │──┘
│ dogma.jsonl         │
└─────────────────────┘
```

## Best Practices

1. **Worker Count:** Verwenden Sie `runtime.NumCPU()` für optimale Auslastung
2. **Context:** Setzen Sie immer angemessene Timeouts
3. **Error Handling:** Prüfen Sie `errors` nach `Wait()`
4. **Job-Funktion:** Implementieren Sie Context-Checking in lang laufenden Jobs
5. **Resource Cleanup:** Verwenden Sie `defer cancel()` bei Context

## Tests

```bash
# Tests ausführen
go test ./internal/worker/...

# Mit Coverage
go test -cover ./internal/worker/...

# Beispiele anzeigen
go test -v ./internal/worker/... -run Example
```

## Siehe auch

- [ORCHESTRATOR.md](./ORCHESTRATOR.md) - **Import Orchestrator Dokumentation (2-Phasen-Import)**
- [ADR-006: Concurrency & Worker Pool Pattern](../../docs/adr/ADR-006-concurrency-worker-pool.md)
- [internal/parser](../parser/) - JSONL-Parsing
- [internal/database](../database/) - SQLite-Integration

## Schnellstart: Import Orchestrator

Für vollständige Dokumentation siehe [ORCHESTRATOR.md](./ORCHESTRATOR.md).

```go
// 1. Setup
db, _ := database.NewDB("eve_sde.db")
defer db.Close()

pool := worker.NewPool(4)
parsers := map[string]parser.Parser{
    "types.jsonl": myParser,
}

// 2. Import ausführen
orch := worker.NewOrchestrator(db, pool, parsers)
progress, err := orch.ImportAll(context.Background(), "/path/to/sde")

// 3. Ergebnisse prüfen
parsed, inserted, failed, total := progress.GetProgress()
log.Printf("Import: %d/%d parsed, %d inserted, %d failed", parsed, total, inserted, failed)
```
