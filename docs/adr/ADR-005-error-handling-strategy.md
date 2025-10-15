# ADR-005: Error Handling Strategy

**Status:** Accepted  
**Datum:** 2025-10-15  
**Entscheider:** Migration Team  
**Kontext:** VB.NET → Go Migration für EVE SDE Database Builder  
**Abhängig von:** ADR-002 (Database Layer), ADR-003 (JSONL Parser), ADR-004 (Config)

---

## Kontext & Problem

### Anforderungen

Die Anwendung muss verschiedene Fehlersituationen robust behandeln:

**Kritische Fehler (Import muss abbrechen):**

- Datenbank-Connection-Fehler
- Korrupte JSONL-Dateien (ungültiges JSON)
- Disk-Full während Batch-Insert
- Schema-Mismatch (SDE-Version inkompatibel)

**Wiederherstellbare Fehler (Retry/Skip möglich):**

- Einzelne JSONL-Zeile fehlerhaft (→ Skip + Log)
- Transiente DB-Fehler (Lock, Timeout → Retry)
- Fehlende optionale Felder (→ NULL + Log)

**User-Fehler (Validierung):**

- Fehlende Config-Werte
- Ungültiger DB-Pfad
- Nicht-existierendes SDE-Verzeichnis

**Betriebliche Anforderungen:**

- Strukturiertes Logging (JSON für Monitoring)
- Error-Recovery (Restart ab letztem Checkpoint)
- Debugging-Informationen (Stack Traces)

### VB.NET Status Quo

**Aktueller Ansatz:**

```vb
' Globals.vb
Public Sub ShowErrorMessage(ex As Exception)
    Dim msg As String = ex.Message
    If Not CancelImport Then
        If Not IsNothing(ex.InnerException) Then
            msg &= vbCrLf & vbCrLf & "Inner Exception: " & ex.InnerException.ToString
        End If
        Call MsgBox(msg, vbExclamation, Application.ProductName)
    End If
End Sub

Public Sub WriteMsgToErrorLog(ByVal ErrorMsg As String)
    Call OutputToFile("Errors.log", ErrorMsg)
End Sub

' frmMain.vb
Try
    ' File Operations
    Dim BSD_DI As New DirectoryInfo(UserApplicationSettings.SDEDirectory & BSDPath)
    Dim BSD_FilesList As FileInfo() = BSD_DI.GetFiles()
    ' ...
Catch ex As Exception
    Call ShowErrorMessage(ex)
End Try

' Legacy Error Handling (stellenweise)
On Error Resume Next
' ... Code ...
On Error GoTo 0
```

**Charakteristik:**

- Mix aus Try/Catch und `On Error Resume Next` (Legacy VB)
- Fehler werden via MsgBox angezeigt (GUI-abhängig)
- Einfache Textdatei-Logs (`Errors.log`, `OutputLog.log`)
- Keine strukturierte Fehlerklassifikation
- Keine Retry-Logik

### Herausforderung

**Go Error Handling Anforderungen:**

1. **Explizite Errors:** Go nutzt `error` return values (kein Exception-System)
2. **Error Wrapping:** Context hinzufügen via `fmt.Errorf("... %w", err)`
3. **Error Types:** Unterscheidbare Fehlertypen für Recovery
4. **Structured Logging:** JSON-Format für Aggregation/Monitoring
5. **CLI-Kontext:** Keine GUI → Errors zu stdout/stderr + Logfiles

---

## Entscheidung

Wir verwenden **Custom Error Types** + **Error Wrapping** + **Structured Logging (zerolog)** + **Retry-Pattern** für transiente Fehler.

### Architektur

**Error Typen (Custom):**

```go
// internal/errors/errors.go
package errors

import (
    "errors"
    "fmt"
)

// ErrorType klassifiziert Fehler für Recovery-Entscheidungen
type ErrorType int

const (
    ErrorTypeUnknown ErrorType = iota
    ErrorTypeFatal               // Unrecoverable (abort import)
    ErrorTypeRetryable          // Transient (retry with backoff)
    ErrorTypeValidation         // User input error
    ErrorTypeSkippable          // Record-level error (skip + log)
)

// AppError ist unser Custom Error Type
type AppError struct {
    Type    ErrorType
    Message string
    Err     error  // Wrapped original error
    Context map[string]interface{}  // Structured context
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Err)
    }
    return e.Message
}

func (e *AppError) Unwrap() error {
    return e.Err
}

// Helper Constructors
func Fatal(msg string, err error) *AppError {
    return &AppError{
        Type:    ErrorTypeFatal,
        Message: msg,
        Err:     err,
        Context: make(map[string]interface{}),
    }
}

func Retryable(msg string, err error) *AppError {
    return &AppError{
        Type:    ErrorTypeRetryable,
        Message: msg,
        Err:     err,
        Context: make(map[string]interface{}),
    }
}

func Validation(msg string) *AppError {
    return &AppError{
        Type:    ErrorTypeValidation,
        Message: msg,
        Context: make(map[string]interface{}),
    }
}

func Skippable(msg string, err error) *AppError {
    return &AppError{
        Type:    ErrorTypeSkippable,
        Message: msg,
        Err:     err,
        Context: make(map[string]interface{}),
    }
}

// WithContext fügt Kontext hinzu
func (e *AppError) WithContext(key string, value interface{}) *AppError {
    e.Context[key] = value
    return e
}

// IsType prüft ErrorType
func IsType(err error, t ErrorType) bool {
    var appErr *AppError
    if errors.As(err, &appErr) {
        return appErr.Type == t
    }
    return false
}
```

**Structured Logging (zerolog):**

```go
// internal/logger/logger.go
package logger

import (
    "os"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

func Init(level string, format string) {
    // Log Level
    switch level {
    case "debug":
        zerolog.SetGlobalLevel(zerolog.DebugLevel)
    case "info":
        zerolog.SetGlobalLevel(zerolog.InfoLevel)
    case "warn":
        zerolog.SetGlobalLevel(zerolog.WarnLevel)
    case "error":
        zerolog.SetGlobalLevel(zerolog.ErrorLevel)
    default:
        zerolog.SetGlobalLevel(zerolog.InfoLevel)
    }
    
    // Format
    if format == "text" {
        log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
    } else {
        // JSON (default für Production)
        log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
    }
}

// LogError loggt AppError mit vollem Kontext
func LogError(err error) {
    var appErr *errors.AppError
    if errors.As(err, &appErr) {
        event := log.Error().
            Str("type", appErr.Type.String()).
            Str("message", appErr.Message)
        
        // Kontext hinzufügen
        for k, v := range appErr.Context {
            event = event.Interface(k, v)
        }
        
        // Original Error
        if appErr.Err != nil {
            event = event.Err(appErr.Err)
        }
        
        event.Msg("application error")
    } else {
        log.Error().Err(err).Msg("unknown error")
    }
}
```

**Retry-Pattern (für transiente Fehler):**

```go
// internal/retry/retry.go
package retry

import (
    "context"
    "time"
    "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"
    "github.com/rs/zerolog/log"
)

type Config struct {
    MaxAttempts int
    InitialDelay time.Duration
    MaxDelay     time.Duration
    Multiplier   float64
}

func DefaultConfig() Config {
    return Config{
        MaxAttempts:  3,
        InitialDelay: 100 * time.Millisecond,
        MaxDelay:     5 * time.Second,
        Multiplier:   2.0,
    }
}

// Do führt Funktion mit Exponential Backoff aus
func Do(ctx context.Context, cfg Config, fn func() error) error {
    var lastErr error
    delay := cfg.InitialDelay
    
    for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
        err := fn()
        
        // Success
        if err == nil {
            return nil
        }
        
        lastErr = err
        
        // Nicht-Retryable Errors → sofort abbrechen
        if !errors.IsType(err, errors.ErrorTypeRetryable) {
            return err
        }
        
        // Letzter Versuch → kein Sleep
        if attempt == cfg.MaxAttempts {
            break
        }
        
        log.Warn().
            Int("attempt", attempt).
            Dur("delay", delay).
            Err(err).
            Msg("retrying operation")
        
        // Backoff
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(delay):
        }
        
        // Exponential Backoff
        delay = time.Duration(float64(delay) * cfg.Multiplier)
        if delay > cfg.MaxDelay {
            delay = cfg.MaxDelay
        }
    }
    
    return lastErr
}
```

**Error Handling in Practice (Parser):**

```go
// internal/parser/jsonl.go
func ParseJSONL[T any](r io.Reader) ([]T, error) {
    var results []T
    scanner := bufio.NewScanner(r)
    scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)
    
    lineNum := 0
    var skippedLines []int
    
    for scanner.Scan() {
        lineNum++
        var item T
        if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
            // Skippable Error: Einzelne Zeile fehlerhaft
            appErr := errors.Skippable("invalid JSON line", err).
                WithContext("line", lineNum).
                WithContext("content", string(scanner.Bytes()[:100]))  // Erste 100 Bytes
            
            logger.LogError(appErr)
            skippedLines = append(skippedLines, lineNum)
            continue  // Skip + Log, aber nicht abbrechen
        }
        results = append(results, item)
    }
    
    if err := scanner.Err(); err != nil {
        // Fatal: Scanner-Fehler = File corrupted
        return nil, errors.Fatal("scanner error", err)
    }
    
    if len(skippedLines) > 0 {
        log.Warn().
            Ints("skipped_lines", skippedLines).
            Int("total_lines", lineNum).
            Msg("completed with skipped lines")
    }
    
    return results, nil
}
```

**Error Handling in Database Layer:**

```go
// internal/database/sqlite.go
func (db *DB) BatchInsert(table string, records []map[string]interface{}) error {
    return retry.Do(context.Background(), retry.DefaultConfig(), func() error {
        tx, err := db.Beginx()
        if err != nil {
            // Retryable: Connection Fehler
            return errors.Retryable("failed to begin transaction", err)
        }
        defer tx.Rollback()
        
        stmt, err := tx.PrepareNamed(buildInsertQuery(table, records[0]))
        if err != nil {
            // Fatal: Query-Syntax-Fehler
            return errors.Fatal("failed to prepare statement", err).
                WithContext("table", table)
        }
        defer stmt.Close()
        
        for i, record := range records {
            if _, err := stmt.Exec(record); err != nil {
                // Fatal: Data-Constraint-Violation
                return errors.Fatal("failed to insert record", err).
                    WithContext("table", table).
                    WithContext("record_index", i)
            }
        }
        
        if err := tx.Commit(); err != nil {
            // Retryable: Lock Timeout
            return errors.Retryable("failed to commit transaction", err)
        }
        
        return nil
    })
}
```

**CLI Error Handling:**

```go
// cmd/esdedb/main.go
func runImport(cmd *cobra.Command, args []string) error {
    cfg, err := config.Load(configPath)
    if err != nil {
        // Validation Error → User-freundlich anzeigen
        if errors.IsType(err, errors.ErrorTypeValidation) {
            return fmt.Errorf("configuration error: %w", err)
        }
        return err
    }
    
    db, err := database.NewDB(cfg.Database.Path)
    if err != nil {
        // Fatal → Exit mit Fehler
        logger.LogError(errors.Fatal("database connection failed", err))
        return err
    }
    defer db.Close()
    
    // Import
    if err := doImport(cfg, db); err != nil {
        logger.LogError(err)
        return err
    }
    
    log.Info().Msg("import completed successfully")
    return nil
}
```

### Begründung

**Warum Custom Error Types?**

| Ansatz | Pro | Contra | Bewertung |
|--------|-----|--------|-----------|
| Bare `error` | ✅ Einfach, Go-Standard | ❌ Keine Klassifikation | Unzureichend |
| `errors.Is/As` | ✅ Sentinel Errors | ⚠️ Nur für bekannte Types | Ergänzend verwendet |
| Custom `AppError` | ✅ Type-basierte Recovery | ⚠️ Boilerplate | **Gewählt** |
| `pkg/errors` | ✅ Stack Traces | ❌ Archived (no longer maintained) | Verworfen |

**Warum zerolog statt log/slog?**

| Feature | zerolog | slog (Go 1.21+) | log (stdlib) |
|---------|---------|-----------------|--------------|
| Structured Logging | ✅ Zero-Alloc | ✅ Ja | ❌ Nein |
| Performance | ✅ Sehr schnell | ✅ Schnell | ⚠️ Langsam |
| JSON Support | ✅ Native | ✅ Native | ❌ Nein |
| Context Fields | ✅ Chainable | ✅ Ja | ❌ Nein |
| Maturity | ✅ Seit 2017 | ⚠️ Neu (2023) | ✅ Stable |

**Entscheidung:** `zerolog` (bewährt, performant, idiomatisch)

**Warum Retry-Pattern?**

- ✅ SQLite Locks sind transient (100-500ms)
- ✅ Batch-Inserts profitieren von Auto-Retry
- ✅ Exponential Backoff verhindert Thundering Herd

---

## Konsequenzen

### Positive Konsequenzen

1. **Type-Safe Recovery:** `IsType()` ermöglicht robuste Error-Handling-Pfade
2. **Debugging:** Strukturierte Logs + Context → schnellere Root-Cause-Analyse
3. **Resilience:** Retry-Pattern für transiente Fehler (SQLite Locks)
4. **Production-Ready:** JSON-Logs → Aggregation via ELK/Loki
5. **CLI-freundlich:** Errors zu stderr, keine GUI-Abhängigkeit
6. **Testing:** Mock-freundlich (`error` interface bleibt unverändert)

### Negative Konsequenzen

1. **Boilerplate:** Custom Error Types erfordern Constructor-Funktionen
2. **Learning Curve:** Team muss Error-Klassifikation lernen
3. **Overhead:** Error Wrapping + Context = mehr Allokationen (minimal)

### Mitigationen

| Konsequenz | Mitigation |
|------------|------------|
| Boilerplate | Helper-Funktionen (`Fatal()`, `Retryable()`, etc.) |
| Learning Curve | Beispiele in ADR + Code-Reviews |
| Overhead | Akzeptabel (< 1% Performance-Impact bei I/O-dominiertem Workload) |

---

## Alternativen (erwogen & verworfen)

### Alternative 1: Panic/Recover statt Error Returns

**Pro:**

- ✅ Weniger Boilerplate (kein `if err != nil`)

**Contra:**

- ❌ Unidiomatisch in Go (Panics nur für echte Programmer Errors)
- ❌ Schwierig zu testen
- ❌ Versteckt Control Flow

**Entscheidung:** Verworfen (Go-Konvention: Errors as Values)

### Alternative 2: `github.com/pkg/errors` (Stack Traces)

**Pro:**

- ✅ Stack Traces für Debugging

**Contra:**

- ❌ Projekt archived (nicht mehr maintained)
- ❌ Go 1.13+ `errors.Is/As` ist bessere Lösung
- ❌ Performance-Overhead (Stack Capture)

**Entscheidung:** Verworfen (stdlib `errors` package ausreichend)

### Alternative 3: `slog` (Go 1.21+ stdlib)

**Pro:**

- ✅ Stdlib (keine externe Dependency)
- ✅ Structured Logging

**Contra:**

- ⚠️ Neu (weniger Battle-tested als zerolog)
- ⚠️ Etwas langsamer (Benchmarks)

**Entscheidung:** Evaluieren für v2.0 (zerolog für v1.0 bewährt)

### Alternative 4: Keine Retry-Logik

**Pro:**

- ✅ Einfacher Code

**Contra:**

- ❌ SQLite Lock-Timeouts führen zu fehlgeschlagenen Imports
- ❌ User muss manuell neu starten

**Entscheidung:** Verworfen (Retry ist essential für Robustheit)

---

## Implementierungsdetails

### 1. Error Type Enum

```go
// internal/errors/types.go
package errors

func (t ErrorType) String() string {
    switch t {
    case ErrorTypeFatal:
        return "fatal"
    case ErrorTypeRetryable:
        return "retryable"
    case ErrorTypeValidation:
        return "validation"
    case ErrorTypeSkippable:
        return "skippable"
    default:
        return "unknown"
    }
}
```

### 2. Logger Integration

```go
// cmd/esdedb/main.go
func main() {
    rootCmd := &cobra.Command{
        Use:   "esdedb",
        Short: "EVE SDE Database Builder",
        PersistentPreRun: func(cmd *cobra.Command, args []string) {
            // Logger initialisieren aus Config
            cfg, _ := config.Load(configPath)
            logger.Init(cfg.Logging.Level, cfg.Logging.Format)
        },
    }
    
    if err := rootCmd.Execute(); err != nil {
        // Top-Level Error Handling
        logger.LogError(err)
        os.Exit(1)
    }
}
```

### 3. Testing-Strategie

```go
// internal/errors/errors_test.go
func TestErrorType_Classification(t *testing.T) {
    err := errors.Fatal("db error", sql.ErrNoRows)
    
    assert.True(t, errors.IsType(err, errors.ErrorTypeFatal))
    assert.False(t, errors.IsType(err, errors.ErrorTypeRetryable))
}

func TestRetry_Backoff(t *testing.T) {
    attempts := 0
    err := retry.Do(context.Background(), retry.DefaultConfig(), func() error {
        attempts++
        if attempts < 3 {
            return errors.Retryable("transient", nil)
        }
        return nil
    })
    
    assert.NoError(t, err)
    assert.Equal(t, 3, attempts)
}
```

### 4. Context Propagation

```go
// internal/database/sqlite.go
func (db *DB) BatchInsertWithContext(ctx context.Context, table string, records []map[string]interface{}) error {
    return retry.Do(ctx, retry.DefaultConfig(), func() error {
        // Check Context Cancellation
        if ctx.Err() != nil {
            return errors.Fatal("context cancelled", ctx.Err())
        }
        
        // ... Insert Logic ...
        return nil
    })
}
```

---

## Migration von VB.NET

### Code-Vergleich

**VB.NET (MsgBox + TextLog):**

```vb
' Globals.vb
Public Sub ShowErrorMessage(ex As Exception)
    Dim msg As String = ex.Message
    If Not IsNothing(ex.InnerException) Then
        msg &= vbCrLf & "Inner Exception: " & ex.InnerException.ToString
    End If
    Call MsgBox(msg, vbExclamation, Application.ProductName)
End Sub

Public Sub WriteMsgToErrorLog(ByVal ErrorMsg As String)
    Call OutputToFile("Errors.log", ErrorMsg)
End Sub

' frmMain.vb
Try
    Dim BSD_DI As New DirectoryInfo(UserApplicationSettings.SDEDirectory & BSDPath)
    Dim BSD_FilesList As FileInfo() = BSD_DI.GetFiles()
Catch ex As Exception
    Call ShowErrorMessage(ex)
End Try
```

**Go (Structured + Typed):**

```go
// internal/parser/files.go
func LoadFiles(dir string) ([]string, error) {
    files, err := os.ReadDir(dir)
    if err != nil {
        return nil, errors.Fatal("failed to read directory", err).
            WithContext("dir", dir)
    }
    
    var result []string
    for _, f := range files {
        result = append(result, f.Name())
    }
    
    log.Info().
        Int("file_count", len(result)).
        Str("dir", dir).
        Msg("loaded files")
    
    return result, nil
}

// Aufruf
files, err := LoadFiles(cfg.Import.SDEPath)
if err != nil {
    logger.LogError(err)  // Structured JSON Log
    return err
}
```

**Unterschiede:**

- ✅ Go: Keine GUI-Abhängigkeit (MsgBox → stderr + Log)
- ✅ Go: Strukturierte Logs (JSON) statt Plain Text
- ✅ Go: Error Wrapping mit Kontext
- ✅ Go: Type-basierte Recovery-Entscheidungen

---

## Compliance & Governance

### Normative Anforderungen

- ✅ **MUST:** Keine Panics für erwartbare Fehler (nur Programmer Errors)
- ✅ **MUST:** Strukturierte Logs (JSON für Production)
- ✅ **SHOULD:** Error Wrapping mit Kontext (`fmt.Errorf("%w", err)`)
- ✅ **SHOULD:** Retry für transiente Fehler (DB Locks)

### ADR-Abhängigkeiten

- **ADR-002:** Database Layer → Retry-Pattern für SQLite Locks
- **ADR-003:** JSONL Parser → Skippable Errors für einzelne Zeilen
- **ADR-004:** Config → Validation Errors

---

## Referenzen

**Libraries:**

- [rs/zerolog](https://github.com/rs/zerolog) - Structured Logging (Zero-Alloc)
- [errors (stdlib)](https://pkg.go.dev/errors) - Error Wrapping (`Is`, `As`, `Unwrap`)
- [context (stdlib)](https://pkg.go.dev/context) - Context Propagation

**Patterns:**

- [Exponential Backoff](https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/)
- [Go Error Handling Best Practices](https://go.dev/blog/error-handling-and-go)

**Alternativen (nicht gewählt):**

- [pkg/errors](https://github.com/pkg/errors) - Archived, nicht mehr maintained
- [slog](https://pkg.go.dev/log/slog) - Go 1.21+ stdlib (für v2.0 evaluieren)

---

## Änderungshistorie

| Datum | Version | Änderung | Autor |
|-------|---------|----------|-------|
| 2025-10-15 | 0.1.0 | Initial Draft | AI Copilot |
| 2025-10-15 | 1.0.0 | Status → Accepted (Custom Errors + zerolog + Retry) | Migration Team |

---

**Nächste Schritte:**

1. ✅ ~~Review durch Team~~ (Accepted)
2. `internal/errors/errors.go` implementieren
3. `internal/logger/logger.go` implementieren
4. `internal/retry/retry.go` implementieren
5. Integration in Parser + Database Layer
6. Tests für Error-Klassifikation + Retry-Logik
7. ✅ ~~Bei Erfolg: Status → `Accepted`~~
