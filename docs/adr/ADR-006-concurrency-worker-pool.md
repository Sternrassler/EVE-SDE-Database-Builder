# ADR-006: Concurrency & Worker Pool Pattern

**Status:** Accepted  
**Datum:** 2025-10-15  
**Entscheider:** Migration Team  
**Kontext:** VB.NET → Go Migration für EVE SDE Database Builder  
**Abhängig von:** ADR-003 (JSONL Parser), ADR-005 (Error Handling)

---

## Kontext & Problem

### Anforderungen

**Parallele JSONL-Verarbeitung:**

- 50+ JSONL-Dateien (types.jsonl ~500k Zeilen, blueprints.jsonl ~20k Zeilen, etc.)
- Gesamtgröße: ~1.5 GB unkomprimiert
- Import-Dauer (Single-Threaded): ~8-10 Minuten (geschätzt)
- Ziel: < 3 Minuten via Parallelisierung

**Ressourcen-Constraints:**

- **SQLite:** Nur **1 Writer** zur Zeit (WAL Mode erlaubt parallele Reads)
- **CPU:** Multi-Core ausnutzen (4-16 Cores typisch)
- **Memory:** Streaming-Parser (kein Full-File-Load)

**User-Konfiguration:**

- Worker Count konfigurierbar (1-32, Default: 4)
- "Max Threads" Option (= Runtime.NumCPU())

**Fehlerbehandlung:**

- Worker-Crash darf Import nicht komplett abbrechen
- Graceful Shutdown bei Cancel (Ctrl+C)

### VB.NET Status Quo

**Threading-Modell:**

```vb
' frmMain.vb (~2237 Zeilen)
Public SelectedThreads As Integer  ' User-konfiguriert

' Thread-Verwaltung
Dim ThreadsArray As New List(Of ThreadList)

For Each YAMLFile In ImportFileList
    Select Case .FileName
        Case YAMLagents.agentsFile
            Dim Agents As New YAMLagents(...)
            TempThreadList.T = New Thread(AddressOf Agents.ImportFile)
            Call ThreadsArray.Add(TempThreadList)
    End Select
Next

' Thread-Start-Logik
If SelectedThreads = -1 Then
    ' Max Threads: Alle sofort starten
    For i = 0 To ThreadsArray.Count - 1
        Call ImportFile(ThreadsArray(i).T, ThreadsArray(i).Params)
    Next
Else
    ' Limited Threads: Polling-basierte Limitierung
    Do
        ActiveThreads = 0
        For Each Th In ThreadsArray
            If Th.T.IsAlive Then
                ActiveThreads += 1
            End If
        Next
        
        If ActiveThreads <= SelectedThreads Then
            Call ImportFile(ThreadsArray(i).T, ThreadsArray(i).Params)
        Else
            ' Warten via DoEvents (Busy-Wait!)
        End If
        Application.DoEvents()
    Loop Until ThreadStarted
End If

' Thread Cleanup
Private Sub KillThreads(ByRef ListofThreads As List(Of ThreadList))
    For i = 0 To ListofThreads.Count - 1
        If ListofThreads(i).T.IsAlive Then
            ListofThreads(i).T.Abort()  ' Hard Abort!
        End If
    Next
End Sub
```

**Charakteristik:**

- Native .NET `Thread` (kein ThreadPool)
- Busy-Wait Polling (`Application.DoEvents()` Loop)
- Hard Thread Abort (`Thread.Abort()`)
- 1 Thread pro JSONL-Datei (50+ Threads!)
- Keine Worker Pool Abstraktion

**Problem:** VB.NET-Ansatz ist ineffizient (Busy-Wait, Hard Abort, keine Backpressure)

### Herausforderung

**Go Concurrency Patterns:**

1. **Goroutines:** Leichtgewichtig, aber SQLite ist 1-Writer-only
2. **Worker Pool:** N Goroutines verarbeiten M Aufgaben (M >> N)
3. **Channels:** Queue für Tasks + Kommunikation
4. **Context:** Graceful Cancellation (kein Hard Abort)
5. **sync.WaitGroup:** Warten auf Completion

**Design-Frage:** Wie parallelisieren bei 1-Writer-Constraint?

---

## Entscheidung

Wir verwenden **Worker Pool Pattern** mit **Buffered Channels** + **Context Cancellation** + **2-Phasen-Import** (Parse parallel, DB-Insert sequentiell).

### Architektur

**2-Phasen-Ansatz:**

```
Phase 1: JSONL Parse (parallel)    Phase 2: DB Insert (sequentiell)
┌─────────────────────┐            ┌──────────────────┐
│ Worker 1            │            │                  │
│ types.jsonl → []Row │──┐         │                  │
└─────────────────────┘  │         │                  │
                         ├─────────► SQLite Writer    │
┌─────────────────────┐  │         │ (1 Goroutine)    │
│ Worker 2            │  │         │                  │
│ blueprints.jsonl →  │──┤         │ Batch Insert     │
└─────────────────────┘  │         │ (Transactions)   │
                         │         │                  │
┌─────────────────────┐  │         │                  │
│ Worker 3            │──┘         │                  │
│ dogma.jsonl → []Row │            │                  │
└─────────────────────┘            └──────────────────┘
```

**Worker Pool Implementation:**

```go
// internal/worker/pool.go
package worker

import (
    "context"
    "fmt"
    "sync"
    
    "github.com/rs/zerolog/log"
)

// Task repräsentiert eine Import-Aufgabe
type Task struct {
    File   string
    Parser func(context.Context) ([]interface{}, error)
}

// Result repräsentiert das Ergebnis einer Task
type Result struct {
    File    string
    Records []interface{}
    Err     error
}

// Pool ist ein Worker Pool für parallele JSONL-Verarbeitung
type Pool struct {
    workers int
    tasks   chan Task
    results chan Result
    wg      sync.WaitGroup
}

// NewPool erstellt einen Worker Pool
func NewPool(workers int) *Pool {
    return &Pool{
        workers: workers,
        tasks:   make(chan Task, workers*2),  // Buffered für Backpressure
        results: make(chan Result, workers*2),
    }
}

// Start startet Worker Goroutines
func (p *Pool) Start(ctx context.Context) {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go p.worker(ctx, i)
    }
}

// worker verarbeitet Tasks aus dem Channel
func (p *Pool) worker(ctx context.Context, id int) {
    defer p.wg.Done()
    
    log.Debug().Int("worker_id", id).Msg("worker started")
    
    for {
        select {
        case <-ctx.Done():
            log.Debug().Int("worker_id", id).Msg("worker cancelled")
            return
            
        case task, ok := <-p.tasks:
            if !ok {
                log.Debug().Int("worker_id", id).Msg("worker finished (channel closed)")
                return
            }
            
            log.Info().
                Int("worker_id", id).
                Str("file", task.File).
                Msg("processing file")
            
            // Task ausführen
            records, err := task.Parser(ctx)
            
            // Result senden
            p.results <- Result{
                File:    task.File,
                Records: records,
                Err:     err,
            }
            
            if err != nil {
                log.Error().
                    Err(err).
                    Str("file", task.File).
                    Msg("task failed")
            } else {
                log.Info().
                    Int("record_count", len(records)).
                    Str("file", task.File).
                    Msg("task completed")
            }
        }
    }
}

// Submit fügt Task zur Queue hinzu
func (p *Pool) Submit(task Task) error {
    select {
    case p.tasks <- task:
        return nil
    default:
        return fmt.Errorf("task queue full")
    }
}

// Close schließt Task-Channel (keine neuen Tasks mehr)
func (p *Pool) Close() {
    close(p.tasks)
}

// Wait wartet auf Worker-Completion
func (p *Pool) Wait() {
    p.wg.Wait()
    close(p.results)
}

// Results liefert Result-Channel
func (p *Pool) Results() <-chan Result {
    return p.results
}
```

**Import Orchestrator (2-Phasen):**

```go
// cmd/esdedb/import.go
package main

import (
    "context"
    "os"
    "os/signal"
    "runtime"
    "syscall"
    
    "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/worker"
    "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"
    "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/database"
    "github.com/rs/zerolog/log"
)

func runImport(cfg *config.Config, db *database.DB) error {
    // Context für Cancellation (Ctrl+C)
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Signal Handling
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-sigChan
        log.Warn().Msg("received interrupt, cancelling import...")
        cancel()
    }()
    
    // Worker Count bestimmen
    workers := cfg.Import.Workers
    if workers <= 0 {
        workers = runtime.NumCPU()
    }
    
    log.Info().
        Int("workers", workers).
        Str("sde_path", cfg.Import.SDEPath).
        Msg("starting import")
    
    // === Phase 1: Parallel Parsing ===
    pool := worker.NewPool(workers)
    pool.Start(ctx)
    
    // Tasks submiten
    files, err := discoverJSONLFiles(cfg.Import.SDEPath)
    if err != nil {
        return err
    }
    
    for _, file := range files {
        task := worker.Task{
            File: file,
            Parser: func(ctx context.Context) ([]interface{}, error) {
                return parseFile(ctx, file)
            },
        }
        if err := pool.Submit(task); err != nil {
            return err
        }
    }
    
    pool.Close()  // Keine neuen Tasks
    
    // Results sammeln (non-blocking)
    go pool.Wait()
    
    // === Phase 2: Sequential DB Insert ===
    insertedFiles := 0
    failedFiles := 0
    
    for result := range pool.Results() {
        if result.Err != nil {
            log.Error().
                Err(result.Err).
                Str("file", result.File).
                Msg("parse failed, skipping insert")
            failedFiles++
            continue
        }
        
        // DB Insert (sequentiell, da SQLite 1-Writer)
        table := inferTableName(result.File)
        if err := db.BatchInsert(table, toMapSlice(result.Records)); err != nil {
            log.Error().
                Err(err).
                Str("file", result.File).
                Str("table", table).
                Msg("insert failed")
            failedFiles++
        } else {
            insertedFiles++
        }
        
        // Ctx Cancellation prüfen
        if ctx.Err() != nil {
            log.Warn().Msg("import cancelled")
            return ctx.Err()
        }
    }
    
    log.Info().
        Int("inserted", insertedFiles).
        Int("failed", failedFiles).
        Msg("import completed")
    
    return nil
}
```

**Graceful Shutdown:**

```go
// cmd/esdedb/main.go
func main() {
    rootCmd := &cobra.Command{
        Use:   "esdedb",
        Short: "EVE SDE Database Builder",
        RunE: func(cmd *cobra.Command, args []string) error {
            cfg, _ := config.Load(configPath)
            db, _ := database.NewDB(cfg.Database.Path)
            defer db.Close()
            
            return runImport(cfg, db)
        },
    }
    
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### Begründung

**Warum Worker Pool statt 1 Goroutine pro Datei?**

| Ansatz | VB.NET (50+ Threads) | Go (1 Goroutine/File) | Go (Worker Pool) |
|--------|----------------------|-----------------------|------------------|
| Overhead | ⚠️ Hoch (OS Threads) | ✅ Niedrig (Goroutines) | ✅ Optimal (Bounded) |
| Resource Control | ❌ Keine Limits | ⚠️ 50+ Goroutines | ✅ Config-basiert |
| Backpressure | ❌ Nein (Busy-Wait) | ❌ Nein | ✅ Buffered Channel |
| Cancellation | ❌ Hard Abort | ✅ Context | ✅ Context |
| SQLite Contention | ⚠️ Viele Writer-Konflikte | ⚠️ Viele Writer-Konflikte | ✅ Sequential Insert |

**Warum 2-Phasen (Parse || Insert)?**

- ✅ **Parse:** CPU-bound, parallelisierbar (JSON Decode)
- ✅ **Insert:** I/O-bound, SQLite 1-Writer (sequentiell optimal)
- ✅ **Memory:** Parsed Daten in Channel (Backpressure via Buffering)

**Warum Context Cancellation statt Thread.Abort?**

- ✅ **Graceful:** Worker beenden Task, dann Exit (kein Partial State)
- ✅ **Safe:** Keine Data-Race (vs. VB.NET Hard Abort mid-operation)

---

## Konsequenzen

### Positive Konsequenzen

1. **Performance:** ~3x schneller als VB.NET (Parallel Parse + Optimized Insert)
2. **Resource Control:** Bounded Worker Count (kein Ressourcen-Explosion)
3. **Graceful Shutdown:** Context Cancellation (Ctrl+C safe)
4. **Backpressure:** Buffered Channels verhindern Memory-Overflow
5. **Testbarkeit:** Worker Pool ist isoliert testbar (Mock Tasks)
6. **SQLite-freundlich:** 1 Writer → keine Lock Contention

### Negative Konsequenzen

1. **Komplexität:** 2-Phasen-Ansatz vs. Naive 1-Goroutine/File
2. **Memory:** Parsed Records in Channel (Peak: Workers * Avg File Size)
3. **Latency:** Kleine Dateien müssen auf große warten (Sequential Insert)

### Mitigationen

| Konsequenz | Mitigation |
|------------|------------|
| Komplexität | Worker Pool als Modul (`internal/worker/`) |
| Memory | Buffered Channel Size = `workers * 2` (Bounded) |
| Latency | Kleine Files zuerst sortieren (Optional: Weighted Queue) |

---

## Alternativen (erwogen & verworfen)

### Alternative 1: 1 Goroutine pro File (Naive Parallel)

**Pro:**

- ✅ Einfach (kein Worker Pool Code)

**Contra:**

- ❌ 50+ Goroutines → SQLite Lock Contention
- ❌ Keine Backpressure (Memory-Explosion möglich)
- ❌ Schwierig zu limitieren (User Config?)

**Entscheidung:** Verworfen (SQLite 1-Writer Constraint)

### Alternative 2: Sequential Import (kein Parallelismus)

**Pro:**

- ✅ Maximal einfach
- ✅ SQLite-optimal (1 Writer)

**Contra:**

- ❌ Langsam (~8-10 Minuten statt <3 Minuten)
- ❌ CPU-Kerne ungenutzt

**Entscheidung:** Verworfen (Performance-Anforderung)

### Alternative 3: Pipeline Pattern (Parse → Transform → Insert)

**Pro:**

- ✅ Flexible Stage-Pipeline
- ✅ Backpressure via Channel Buffering

**Contra:**

- ⚠️ Komplexer als Worker Pool (3+ Stages)
- ⚠️ Overkill für 2-Phasen-Problem

**Entscheidung:** Evaluieren für v2.0 (Worker Pool ausreichend für v1.0)

### Alternative 4: External Queue (Redis, RabbitMQ)

**Pro:**

- ✅ Distributed Processing möglich

**Contra:**

- ❌ Overkill (kein Distributed Use-Case)
- ❌ Externe Dependency (Redis/RabbitMQ)

**Entscheidung:** Verworfen (Scope: Single-Machine Tool)

---

## Implementierungsdetails

### 1. Worker Count Konfiguration

```toml
# config.toml
[import]
workers = 4  # 0 = auto (runtime.NumCPU())
```

```go
// internal/config/config.go
type ImportConfig struct {
    Workers int `toml:"workers"`  // 0 = Auto
}

func (c *Config) Validate() error {
    if c.Import.Workers < 0 || c.Import.Workers > 32 {
        return fmt.Errorf("workers must be 0-32 (0=auto)")
    }
    return nil
}
```

### 2. Context Propagation

```go
// internal/parser/jsonl.go
func ParseJSONL[T any](ctx context.Context, r io.Reader) ([]T, error) {
    var results []T
    scanner := bufio.NewScanner(r)
    
    for scanner.Scan() {
        // Context Cancellation Check
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
        }
        
        var item T
        if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
            // ... Error Handling ...
        }
        results = append(results, item)
    }
    
    return results, scanner.Err()
}
```

### 3. File Discovery

```go
// internal/importer/files.go
func discoverJSONLFiles(dir string) ([]string, error) {
    var files []string
    
    err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() && strings.HasSuffix(info.Name(), ".jsonl") {
            files = append(files, path)
        }
        return nil
    })
    
    return files, err
}

// inferTableName extrahiert Tabellenname aus Filename
func inferTableName(file string) string {
    base := filepath.Base(file)
    return strings.TrimSuffix(base, ".jsonl")
}
```

### 4. Progress Tracking (Optional)

```go
// internal/worker/progress.go
type ProgressTracker struct {
    total     int
    completed int
    mu        sync.Mutex
}

func (p *ProgressTracker) Increment() {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.completed++
    
    pct := (p.completed * 100) / p.total
    log.Info().
        Int("completed", p.completed).
        Int("total", p.total).
        Int("percent", pct).
        Msg("import progress")
}
```

### 5. Testing-Strategie

```go
// internal/worker/pool_test.go
func TestPool_Cancellation(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    pool := worker.NewPool(2)
    pool.Start(ctx)
    
    // Submit 10 Tasks
    for i := 0; i < 10; i++ {
        pool.Submit(worker.Task{
            File: fmt.Sprintf("test%d.jsonl", i),
            Parser: func(ctx context.Context) ([]interface{}, error) {
                time.Sleep(100 * time.Millisecond)
                return []interface{}{}, nil
            },
        })
    }
    
    // Cancel nach 50ms
    time.AfterFunc(50*time.Millisecond, cancel)
    
    pool.Close()
    pool.Wait()
    
    // Verify: Nicht alle Tasks completed
    completed := 0
    for range pool.Results() {
        completed++
    }
    
    assert.Less(t, completed, 10)  // Einige wurden cancelled
}
```

---

## Migration von VB.NET

### Code-Vergleich

**VB.NET (Polling + Hard Abort):**

```vb
' frmMain.vb
Dim ThreadsArray As New List(Of ThreadList)

' Thread-Start mit Busy-Wait
Do
    ActiveThreads = 0
    For Each Th In ThreadsArray
        If Th.T.IsAlive Then
            ActiveThreads += 1
        End If
    Next
    
    If ActiveThreads <= SelectedThreads Then
        Call ImportFile(ThreadsArray(i).T, ThreadsArray(i).Params)
    End If
    Application.DoEvents()  ' Busy-Wait!
Loop Until ThreadStarted

' Thread Cleanup (Hard Abort!)
Private Sub KillThreads(ByRef ListofThreads As List(Of ThreadList))
    For i = 0 To ListofThreads.Count - 1
        If ListofThreads(i).T.IsAlive Then
            ListofThreads(i).T.Abort()  ' Dangerous!
        End If
    Next
End Sub
```

**Go (Worker Pool + Context):**

```go
// cmd/esdedb/import.go
pool := worker.NewPool(workers)
pool.Start(ctx)  // Workers starten

// Tasks submiten (non-blocking)
for _, file := range files {
    pool.Submit(worker.Task{...})
}

pool.Close()  // Keine neuen Tasks
pool.Wait()   // Warten auf Completion

// Graceful Shutdown via Context
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go func() {
    <-sigChan
    cancel()  // Workers beenden gracefully
}()
```

**Unterschiede:**

- ✅ Go: Channel-basiert (kein Polling)
- ✅ Go: Context Cancellation (kein Hard Abort)
- ✅ Go: Bounded Worker Count (kein Ressourcen-Explosion)
- ✅ Go: 2-Phasen (SQLite-optimal)

---

## Performance-Erwartungen

### Benchmark-Ziele (50+ JSONL Files, ~1.5 GB)

| Operation | VB.NET (Threads) | Go (Sequential) | Go (Workers=4) | Go (Workers=8) |
|-----------|------------------|-----------------|----------------|----------------|
| Parse Phase | ~6 min | ~8 min | ~2 min | ~1.5 min |
| Insert Phase | ~2 min | ~2 min | ~2 min | ~2 min |
| **Total** | **~8 min** | **~10 min** | **~4 min** | **~3.5 min** |

**Begründung:**

- Parse: Parallelisiert (Linear Speedup bis CPU-Limit)
- Insert: Sequential (SQLite 1-Writer)

---

## Compliance & Governance

### Normative Anforderungen

- ✅ **MUST:** Graceful Shutdown via Context (kein Hard Abort)
- ✅ **MUST:** Bounded Concurrency (User-konfigurierbar)
- ✅ **SHOULD:** Progress Tracking (für User-Feedback)
- ✅ **MAY:** Weighted Task Queue (Optimize Small Files First)

### ADR-Abhängigkeiten

- **ADR-003:** JSONL Parser → Context-aware Parsing
- **ADR-005:** Error Handling → Worker Errors nicht fatal

---

## Referenzen

**Go Concurrency Patterns:**

- [Go Concurrency Patterns (Rob Pike)](https://go.dev/talks/2012/concurrency.slide)
- [context Package](https://pkg.go.dev/context)
- [sync.WaitGroup](https://pkg.go.dev/sync#WaitGroup)

**Worker Pool Implementations:**

- [gammazero/workerpool](https://github.com/gammazero/workerpool)
- [alitto/pond](https://github.com/alitto/pond)

**SQLite Concurrency:**

- [SQLite WAL Mode](https://www.sqlite.org/wal.html)
- [SQLite Single-Writer](https://www.sqlite.org/lockingv3.html)

---

## Änderungshistorie

| Datum | Version | Änderung | Autor |
|-------|---------|----------|-------|
| 2025-10-15 | 0.1.0 | Initial Draft | AI Copilot |
| 2025-10-15 | 1.0.0 | Status → Accepted (Worker Pool + 2-Phasen-Import) | Migration Team |

---

**Nächste Schritte:**

1. ✅ ~~Review durch Team~~ (Accepted)
2. `internal/worker/pool.go` implementieren
3. `cmd/esdedb/import.go` 2-Phasen-Orchestrator
4. Context Cancellation in Parser integrieren
5. Progress Tracking (optional)
6. Benchmarks: Sequential vs. Workers=4 vs. Workers=8
7. ✅ ~~Bei Erfolg: Status → `Accepted`~~
