# Import Orchestrator Documentation

## Overview

The Import Orchestrator implements the 2-phase import architecture defined in ADR-006. It coordinates the parallel parsing of JSONL files and sequential database insertion, optimized for SQLite's single-writer constraint.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Import Orchestrator                      │
└─────────────────────────────────────────────────────────────┘
                             │
                ┌────────────┴────────────┐
                │                         │
        ┌───────▼───────┐         ┌──────▼──────┐
        │   Phase 1:    │         │  Phase 2:   │
        │   Parallel    │────────▶│  Sequential │
        │   Parsing     │         │   Insert    │
        └───────────────┘         └─────────────┘
                │                         │
        ┌───────┴────────┐       ┌────────┴────────┐
        │  Worker Pool   │       │ Database Batch  │
        │  (N workers)   │       │ Insert (1 conn) │
        └────────────────┘       └─────────────────┘
```

## Components

### 1. Orchestrator Struct

```go
type Orchestrator struct {
    db      *sqlx.DB
    pool    *Pool
    parsers map[string]parser.Parser
}
```

**Purpose**: Coordinates database connection, worker pool, and registered parsers.

**Fields**:
- `db`: SQLite database connection
- `pool`: Worker pool for parallel parsing
- `parsers`: Map of file names to parser implementations

### 2. Progress Tracker

```go
type ProgressTracker struct {
    parsed   atomic.Int32
    inserted atomic.Int32
    failed   atomic.Int32
    total    int
}
```

**Purpose**: Thread-safe progress monitoring during import.

**Methods**:
- `IncrementParsed()`: Increment parsed file counter
- `IncrementInserted()`: Increment successfully inserted counter
- `IncrementFailed()`: Increment failed operations counter
- `GetProgress()`: Get current counters (parsed, inserted, failed, total)

## Usage

### Basic Import Flow

```go
// 1. Setup database
db, err := database.NewDB("eve_sde.db")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// 2. Create worker pool (4 workers)
pool := worker.NewPool(4)

// 3. Register parsers
parsers := map[string]parser.Parser{
    "types.jsonl":      parser.NewJSONLParser[TypeInfo]("invTypes", []string{"typeID", "typeName"}),
    "groups.jsonl":     parser.NewJSONLParser[GroupInfo]("invGroups", []string{"groupID", "groupName"}),
    "blueprints.jsonl": parser.NewJSONLParser[Blueprint]("blueprints", []string{"blueprintID", "name"}),
}

// 4. Create orchestrator
orch := worker.NewOrchestrator(db, pool, parsers)

// 5. Execute import
ctx := context.Background()
progress, err := orch.ImportAll(ctx, "/path/to/sde/fsd")
if err != nil {
    log.Fatalf("Import failed: %v", err)
}

// 6. Check results
parsed, inserted, failed, total := progress.GetProgress()
log.Printf("Import complete: %d/%d parsed, %d inserted, %d failed", parsed, total, inserted, failed)
```

### With Context Cancellation

```go
// Create cancellable context
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Setup signal handling for Ctrl+C
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
go func() {
    <-sigChan
    log.Println("Interrupt received, cancelling import...")
    cancel()
}()

// Execute import (will stop gracefully on cancel)
progress, err := orch.ImportAll(ctx, sdeDir)
if err == context.Canceled {
    log.Println("Import cancelled by user")
}
```

### With Progress Monitoring

```go
// Start import in goroutine
progressChan := make(chan *worker.ProgressTracker)
go func() {
    progress, _ := orch.ImportAll(ctx, sdeDir)
    progressChan <- progress
}()

// Monitor progress (polling)
ticker := time.NewTicker(1 * time.Second)
defer ticker.Stop()

for {
    select {
    case <-ticker.C:
        parsed, inserted, failed, total := progress.GetProgress()
        pct := (parsed * 100) / total
        log.Printf("Progress: %d%% (%d/%d files)", pct, parsed, total)
        
    case finalProgress := <-progressChan:
        parsed, inserted, failed, total := finalProgress.GetProgress()
        log.Printf("Final: %d/%d parsed, %d inserted, %d failed", parsed, total, inserted, failed)
        return
    }
}
```

## Phase Details

### Phase 1: Parallel Parsing

**Goal**: Parse JSONL files in parallel using worker pool.

**Process**:
1. Create parse tasks for each registered parser
2. Submit tasks to worker pool
3. Workers parse files concurrently
4. Results collected in channel

**Characteristics**:
- CPU-bound operation
- Benefits from parallelization
- N workers process M files (M >> N typical)
- Context cancellation supported

**Performance**: ~2-4x speedup with 4 workers (depends on CPU cores)

### Phase 2: Sequential Insert

**Goal**: Insert parsed data into SQLite database.

**Process**:
1. Process results from Phase 1 sequentially
2. Convert parsed records to database rows
3. Batch insert using transactions
4. Track success/failure

**Characteristics**:
- I/O-bound operation
- Sequential due to SQLite single-writer constraint
- Uses batch inserts (1000 rows/batch) for efficiency
- Transactional (all-or-nothing per file)

**Performance**: SQLite-optimal (no lock contention)

## Error Handling

### Parse Errors

**Behavior**: Parse failures don't stop the import.

```go
// Failed parse increments failed counter
if result.Err != nil {
    progress.IncrementFailed()
    continue // Skip to next file
}
```

**Use Cases**:
- Malformed JSON in JSONL file
- Missing required fields
- Type conversion errors

### Insert Errors

**Behavior**: Insert failures don't stop the import.

```go
// Failed insert increments failed counter
err = database.BatchInsert(ctx, o.db, table, columns, rows, 1000)
if err != nil {
    progress.IncrementFailed()
    continue // Skip to next file
}
```

**Use Cases**:
- Table doesn't exist
- Schema mismatch
- Constraint violations

### Context Cancellation

**Behavior**: Import stops gracefully, no partial data.

```go
select {
case <-ctx.Done():
    return progress, ctx.Err()
default:
}
```

**Guarantees**:
- Current batch completes (transactional)
- No orphaned goroutines
- Progress reflects actual state

## Performance Characteristics

### Benchmarks

Based on test data (10 files):

```
BenchmarkOrchestrator_2Workers_10Files   100   11245 ns/op
BenchmarkOrchestrator_4Workers_10Files   150    7893 ns/op
```

**Speedup**: ~1.4x from 2 to 4 workers

### Real-World Estimates

EVE SDE (50+ files, ~1.5 GB):

| Configuration | Parse Time | Insert Time | Total  |
|---------------|------------|-------------|--------|
| Sequential    | ~8 min     | ~2 min      | ~10 min|
| 2 Workers     | ~4 min     | ~2 min      | ~6 min |
| 4 Workers     | ~2 min     | ~2 min      | ~4 min |
| 8 Workers     | ~1.5 min   | ~2 min      | ~3.5 min|

**Note**: Insert time constant (SQLite single-writer)

## Testing

### Unit Tests

15 test cases cover:
- Progress tracker (creation, increment, concurrency)
- Orchestrator (creation, task creation, row conversion)
- Import flow (empty, single file, multiple files)
- Error handling (parse errors, mixed results)
- Context cancellation

### Example Tests

3 example tests demonstrate:
- `Example_orchestratorBasicUsage`: Basic import flow
- `Example_orchestratorWithContextCancellation`: Graceful shutdown
- `Example_orchestratorProgressTracking`: Progress monitoring

### Running Tests

```bash
# All worker tests
make test

# Only orchestrator tests
go test -v ./internal/worker/... -run "Orchestrator"

# Only examples
go test -v ./internal/worker/... -run "Example"

# Benchmarks
go test -bench=. ./internal/worker/...
```

## ADR Compliance

This implementation follows **ADR-006: Concurrency & Worker Pool Pattern**.

**Key Requirements**:
- ✅ 2-phase import (parallel parse, sequential insert)
- ✅ Worker pool pattern (bounded concurrency)
- ✅ Context cancellation (graceful shutdown)
- ✅ SQLite single-writer optimization
- ✅ Progress tracking
- ✅ Error resilience

## Future Enhancements

### Planned (v1.1)
- File discovery (scan SDE directory for .jsonl files)
- Record-to-row conversion (automatic mapping)
- Progress callbacks (real-time updates)

### Considered (v2.0)
- Weighted task queue (prioritize small files)
- Pipeline pattern (parse → transform → insert)
- Retry logic (transient errors)

## Related

- **ADR-006**: Concurrency & Worker Pool Pattern
- **internal/worker/pool.go**: Worker pool implementation
- **internal/parser/parser.go**: Parser interface
- **internal/database/batch.go**: Batch insert functionality
