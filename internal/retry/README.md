# Retry Package

Das `retry` Package implementiert Exponential Backoff Retry-Logik für transiente Fehler.

## Übersicht

Das Package bietet:
- Exponential Backoff Algorithmus mit konfigurierbaren Parametern
- Jitter zur Vermeidung von Thundering Herd Problemen
- Context-aware Retry (Unterbrechung via `context.Context`)
- Intelligente Fehlerklassifikation (nur `ErrorTypeRetryable` wird wiederholt)
- Generische `DoWithResult` Funktion für Funktionen mit Rückgabewerten
- **Vordefinierte Policies** für Database, HTTP und File I/O
- **PolicyBuilder** für flexible Custom Policies
- **TOML Serialisierung** für Konfigurationsdateien

## Verwendung

### Vordefinierte Policies

#### DatabasePolicy
Optimiert für Datenbank-Operationen (5 retries, 50ms initial, 2s max):
```go
policy := retry.DatabasePolicy()
err := policy.Do(ctx, func() error {
    return db.Query(...)
})
```

**Wann nutzen:**
- Transaktions-Locks (SQLite BUSY)
- Connection Pool erschöpft
- Temporäre DB-Latenzen

#### HTTPPolicy
Optimiert für HTTP-Requests (3 retries, 100ms initial, 5s max):
```go
policy := retry.HTTPPolicy()
resp, err := retry.DoWithResult(ctx, policy, func() (*http.Response, error) {
    return http.Get(url)
})
```

**Wann nutzen:**
- API Rate Limits (429)
- Server Errors (5xx)
- Temporäre Netzwerkprobleme

#### FileIOPolicy
Optimiert für Datei-I/O (2 retries, 10ms initial, 500ms max):
```go
policy := retry.FileIOPolicy()
err := policy.Do(ctx, func() error {
    return os.WriteFile(path, data, 0644)
})
```

**Wann nutzen:**
- File Locking Konflikte
- Temporäre I/O-Fehler
- NFS/Netzwerk-Dateisysteme

### PolicyBuilder (Custom Policies)

```go
policy := retry.NewPolicyBuilder().
    WithMaxRetries(10).
    WithInitialDelay(200 * time.Millisecond).
    WithMaxDelay(30 * time.Second).
    WithMultiplier(3.0).
    WithJitter(false).
    Build()

err := policy.Do(ctx, myFunc)
```

### Einfaches Retry

```go
import (
    "context"
    "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/retry"
    "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"
)

// Verwende Default Policy (3 retries, 100ms initial, 5s max)
policy := retry.DefaultPolicy()

err := policy.Do(context.Background(), func() error {
    // Deine Logik hier
    return someOperation()
})
```

### Custom Policy

```go
// Custom Policy: 5 retries, 50ms initial delay, 10s max delay
policy := retry.NewPolicy(5, 50*time.Millisecond, 10*time.Second)

err := policy.Do(ctx, func() error {
    return databaseOperation()
})
```

### Mit Rückgabewert

```go
result, err := retry.DoWithResult(ctx, policy, func() (string, error) {
    return fetchData()
})
```

### Context Cancellation

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := policy.Do(ctx, func() error {
    return longRunningOperation()
})

// Wird automatisch abgebrochen wenn Context timeout erreicht
```

## Policy Konfiguration

### Struktur

```go
type Policy struct {
    MaxRetries   int           // Anzahl der Wiederholungen (Standard: 3)
    InitialDelay time.Duration // Initiale Verzögerung (Standard: 100ms)
    MaxDelay     time.Duration // Maximale Verzögerung (Standard: 5s)
    Multiplier   float64       // Multiplikator für Exponential Backoff (Standard: 2.0)
    Jitter       bool          // Jitter aktivieren (Standard: true)
}
```

### TOML Konfiguration

Policies können via TOML konfiguriert werden:

```toml
[retry.database]
max_retries = 5
initial_delay_ms = 50
max_delay_ms = 2000
multiplier = 2.0
jitter = true

[retry.http]
max_retries = 3
initial_delay_ms = 100
max_delay_ms = 5000
multiplier = 2.0
jitter = true
```

**Laden aus TOML:**
```go
var config retry.PolicyConfig
if _, err := toml.DecodeFile("config.toml", &config); err != nil {
    return err
}

policy := retry.FromConfig(config)
```

**Speichern als TOML:**
```go
policy := retry.DatabasePolicy()
config := policy.ToConfig()

var buf bytes.Buffer
encoder := toml.NewEncoder(&buf)
encoder.Encode(config)
```

### Backoff Berechnung

Die Verzögerung wird exponentiell berechnet:
```
delay = InitialDelay × Multiplier^attempt
```

Mit Jitter wird eine zufällige Variation von ±10% hinzugefügt.

**Beispiel (Default Policy):**
- Attempt 0: 100ms
- Attempt 1: 200ms (±10% mit Jitter)
- Attempt 2: 400ms (±10% mit Jitter)
- Attempt 3: 800ms (capped at MaxDelay = 5s)

## Fehlerklassifikation

Nur Fehler vom Typ `ErrorTypeRetryable` werden wiederholt:

```go
import "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"

// Wird wiederholt
err := errors.NewRetryable("temporary database lock", sqliteErr)

// Wird NICHT wiederholt
err := errors.NewFatal("database connection failed", dbErr)
err := errors.NewValidation("invalid input", nil)
```

## Use Cases

### SQLite BUSY Errors

```go
policy := retry.DefaultPolicy()

err := policy.Do(ctx, func() error {
    tx, err := db.Begin()
    if err != nil {
        // SQLite BUSY Fehler als Retryable markieren
        return errors.NewRetryable("database locked", err)
    }
    defer tx.Rollback()
    
    // ... transaction logic ...
    
    return tx.Commit()
})
```

### HTTP Requests (429, 503)

```go
policy := retry.NewPolicy(5, 1*time.Second, 30*time.Second)

data, err := retry.DoWithResult(ctx, policy, func() ([]byte, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, errors.NewRetryable("network error", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == 429 || resp.StatusCode == 503 {
        return nil, errors.NewRetryable("rate limited", nil)
    }
    
    if resp.StatusCode != 200 {
        return nil, errors.NewFatal("http error", fmt.Errorf("status %d", resp.StatusCode))
    }
    
    return io.ReadAll(resp.Body)
})
```

## Performance

Benchmark-Ergebnisse (AMD EPYC 7763):
- Erfolgreicher Aufruf: **~3.7 ns/op** (0 Allokationen)
- Ein Retry: **~580 ns/op** (6 Allokationen)
- Backoff Berechnung: **~15 ns/op** (0 Allokationen)

Overhead pro Retry-Versuch: **< 10μs** (erfüllt ADR-005 Anforderung)

## Referenzen

- **ADR-005:** Error Handling Strategy (Retry Pattern)
- [Exponential Backoff und Jitter (AWS)](https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/)
