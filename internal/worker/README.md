# Worker Package

Worker Pool Implementierung für paralleles JSONL-Parsing gemäß ADR-006.

## Überblick

Das `worker` Package stellt einen konfigurierbaren Worker Pool zur Verfügung, der Jobs parallel mit einer definierten Anzahl von Workers verarbeitet. Es implementiert das Worker Pool Pattern mit Channel-basierter Job-Verteilung und Graceful Shutdown via Context.

## Features

- ✅ Konfigurierbare Worker-Anzahl
- ✅ Channel-basierte Job-Verteilung (buffered channels)
- ✅ Graceful Shutdown über `context.Context`
- ✅ Error Collection (alle Job-Fehler werden gesammelt)
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

### `Job`

```go
type Job struct {
    ID string
    Fn func(context.Context) (interface{}, error)
}
```

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

- [ADR-006: Concurrency & Worker Pool Pattern](../../docs/adr/ADR-006-concurrency-worker-pool.md)
- [internal/parser](../parser/) - JSONL-Parsing
- [internal/database](../database/) - SQLite-Integration
