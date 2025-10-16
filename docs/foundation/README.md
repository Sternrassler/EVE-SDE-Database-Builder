# Foundation Packages

Die Foundation Packages bilden die technische Grundlage der EVE SDE Database Builder Anwendung. Sie stellen wiederverwendbare, robuste Infrastrukturkomponenten für Logging, Fehlerbehandlung, Retry-Mechanismen und Konfigurationsverwaltung bereit.

## Übersicht

| Package | Beschreibung | Hauptfunktionalität |
|---------|--------------|---------------------|
| **logger** | Strukturiertes Logging | Zerolog-Wrapper, Log-Levels, Formate (JSON/Text), Context-Logging |
| **errors** | Typisierte Fehler | Fehlerklassifizierung (Fatal/Retryable/Validation/Skippable), Context, Error Wrapping |
| **retry** | Retry-Mechanismen | Exponential Backoff, vordefinierte Policies, Context-Support |
| **config** | Konfigurationsverwaltung | TOML + Env + Flags, Validierung, Defaults |

## Package-Details

### logger

**Zweck**: Einheitliche Logging-Schnittstelle für die gesamte Anwendung

**Kernfeatures**:
- Strukturiertes Logging mit zerolog
- Konfigurierbare Log-Levels (debug, info, warn, error, fatal)
- Ausgabeformate: JSON (Produktion), Text (Entwicklung)
- Context-basiertes Logging (Request IDs, User IDs)
- Spezialisierte Helper (HTTP, DB, App-Lifecycle)
- Globaler Logger für einfache Nutzung

**Verwendung**:
```go
import "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/logger"

log := logger.NewLogger("info", "json")
log.Info("Application started", logger.Field{Key: "version", Value: "1.0.0"})
```

**Dokumentation**: `go doc internal/logger`

---

### errors

**Zweck**: Präzise Fehlerklassifizierung für robuste Fehlerbehandlung

**Kernfeatures**:
- Vier Fehlertypen: Fatal, Retryable, Validation, Skippable
- Strukturierte Context-Informationen
- Error Wrapping (Go 1.13+ `errors.Is`, `errors.As`)
- Integration mit retry-Package
- Type-Checker-Funktionen

**Verwendung**:
```go
import "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"

err := errors.NewRetryable("API request failed", originalErr)
err = err.WithContext("endpoint", "/api/v1/data")

if errors.IsRetryable(err) {
    // Retry-Logik
}
```

**Dokumentation**: `go doc internal/errors`

---

### retry

**Zweck**: Robuste Retry-Mechanismen mit exponentiellem Backoff

**Kernfeatures**:
- Exponentieller Backoff mit Jitter
- Vordefinierte Policies (Default, Database, HTTP, FileIO)
- Policy Builder für Custom-Konfiguration
- Context-Support (Cancellation, Timeouts)
- Integration mit errors-Package (nur Retryable werden wiederholt)
- Generische `DoWithResult[T]` für Funktionen mit Rückgabewert
- TOML-Konfiguration

**Verwendung**:
```go
import "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/retry"

policy := retry.DefaultPolicy()
err := policy.Do(ctx, func() error {
    return someOperation()
})
```

**Dokumentation**: `go doc internal/retry`

---

### config

**Zweck**: Kaskadierende Konfigurationsverwaltung (TOML → Env → Flags)

**Kernfeatures**:
- Multi-Source-Konfiguration mit Priorität
- TOML-Datei-Support
- Environment Variable Override
- Automatische Validierung
- Auto-Konfiguration (Worker-Count = CPU)
- Konfigurationsbereiche: Database, Import, Logging, Update

**Verwendung**:
```go
import "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/config"

cfg, err := config.Load("config.toml")
if err != nil {
    log.Fatal(err)
}
```

**Dokumentation**: `go doc internal/config`

## Design-Entscheidungen

### ADR-Referenzen

- **ADR-004**: Configuration Format (TOML + Environment + Flags)
- Weitere relevante ADRs in `docs/adr/`

### Designprinzipien

1. **Minimale Abhängigkeiten**: Nur essenzielle externe Libraries (zerolog, TOML)
2. **Go-Idiomatisch**: Nutzung von Standard-Patterns (errors.Is/As, Context)
3. **Testbarkeit**: Alle Packages vollständig getestet, Examples als Living Documentation
4. **Fehlerbehandlung**: Explizite Fehlerklassifizierung statt impliziter Annahmen
5. **Konfigurierbarkeit**: Sinnvolle Defaults, aber alles überschreibbar

## Verwendung

### Initialisierung (Typischer Startup)

```go
// 1. Konfiguration laden
cfg, err := config.Load("config.toml")
if err != nil {
    log.Fatal(err)
}

// 2. Logger initialisieren
log := logger.NewLogger(cfg.Logging.Level, cfg.Logging.Format)
logger.SetGlobalLogger(log)

// 3. Application Start loggen
logger.LogAppStart("1.0.0", "abc123")

// 4. Retry-Policy für DB-Operationen
retryPolicy := retry.DatabasePolicy()
```

### Fehlerbehandlung mit Retry

```go
policy := retry.HTTPPolicy()
data, err := retry.DoWithResult(ctx, policy, func() (Data, error) {
    result, err := fetchFromAPI()
    if err != nil {
        // Temporärer Fehler → Retry
        if isTemporary(err) {
            return Data{}, errors.NewRetryable("API temporary error", err)
        }
        // Permanenter Fehler → Kein Retry
        return Data{}, errors.NewFatal("API permanent error", err)
    }
    return result, nil
})

if err != nil {
    logger.LogAppError(err)
    // Fehlerbehandlung
}
```

### Context-basiertes Logging

```go
ctx := context.WithValue(ctx, logger.RequestIDKey, "req-123")
ctxLogger := logger.GetGlobalLogger().WithContext(ctx)

ctxLogger.Info("Processing request") // RequestID automatisch enthalten
```

## Tests & Beispiele

### Tests ausführen

```bash
# Alle Foundation-Tests
make test

# Spezifisches Package
go test -v ./internal/logger
go test -v ./internal/errors
go test -v ./internal/retry
go test -v ./internal/config

# Mit Coverage
go test -cover ./internal/...
```

### Beispiele ausführen

```bash
# Logger-Beispiele
go test -v ./internal/logger -run Example

# Errors-Beispiele
go test -v ./internal/errors -run Example

# Retry-Beispiele
go test -v ./internal/retry -run Example

# Config-Beispiele
go test -v ./internal/config -run Example
```

### Dokumentation anzeigen

```bash
# Package-Dokumentation
go doc internal/logger
go doc internal/errors
go doc internal/retry
go doc internal/config

# Spezifische Funktion
go doc internal/retry.DefaultPolicy
go doc internal/errors.NewRetryable

# Lokal godoc Server starten
godoc -http=:6060
# Dann: http://localhost:6060/pkg/github.com/Sternrassler/EVE-SDE-Database-Builder/internal/
```

## Best Practices

### Logging

- **Entwicklung**: `NewLogger("debug", "text")` für lesbare Konsolen-Ausgabe
- **Produktion**: `NewLogger("info", "json")` für strukturierte Log-Aggregation
- **Sensitive Daten**: Niemals Credentials oder personenbezogene Daten loggen
- **Context**: Nutzen Sie Context-Logger für Request-Tracing

### Fehlerbehandlung

- **Klassifizierung**: Wählen Sie den richtigen ErrorType (Fatal/Retryable/Validation/Skippable)
- **Context**: Fügen Sie relevante Debug-Informationen mit `WithContext()` hinzu
- **Wrapping**: Wrappen Sie ursprüngliche Fehler, um die Fehlerkette zu erhalten
- **Type-Checks**: Nutzen Sie `errors.IsRetryable()` statt String-Vergleichen

### Retry

- **Policy-Auswahl**: Verwenden Sie passende vordefinierte Policies
- **Jitter**: Aktiviert halten für parallele Requests (vermeidet Thundering Herd)
- **Timeouts**: Setzen Sie Context-Timeouts für Gesamtoperationen
- **Rate Limits**: Beachten Sie externe API Rate Limits bei Policy-Konfiguration

### Konfiguration

- **Entwicklung**: `config.toml` im Repository-Root (nicht committen!)
- **Produktion**: Environment Variables für Container/Cloud
- **Secrets**: Niemals in TOML-Dateien (nutzen Sie Secret Manager)
- **Template**: `config.toml.example` als Vorlage für neue Umgebungen

## Integration mit anderen Komponenten

Die Foundation Packages werden von allen anderen Komponenten genutzt:

- **DB-Import**: Nutzt config für Paths/Workers, logger für Fortschritt, retry für DB-Transaktionen
- **HTTP-Client**: Nutzt retry.HTTPPolicy(), errors.NewRetryable(), logger für Requests
- **CLI**: Nutzt config.Load() für Konfiguration, logger für Ausgabe

## Erweiterung

### Neue Retry-Policy hinzufügen

```go
// In internal/retry/policies.go
func CustomPolicy() *Policy {
    return NewPolicy(
        maxRetries,
        initialDelay,
        maxDelay,
    )
}
```

### Neuer ErrorType

Falls neue Fehlerklassifizierung benötigt wird:

1. Neue Konstante in `internal/errors/errors.go` hinzufügen
2. Constructor-Funktion erstellen
3. Type-Checker-Funktion hinzufügen
4. Tests und Beispiele ergänzen

### Logger Helper

Für domänenspezifisches Logging:

```go
// In internal/logger/helpers.go
func (l *Logger) LogCustomEvent(data CustomData) {
    l.Info("Custom event", 
        Field{Key: "field1", Value: data.Field1},
        Field{Key: "field2", Value: data.Field2},
    )
}
```

## Versionierung

Foundation Packages folgen der Repository-Versionierung (siehe `VERSION`).

Breaking Changes in Foundation-APIs erfordern:
- Major Version Bump
- Migration Guide
- Deprecation Warnings in vorheriger Minor Version

## Lizenz

Siehe `LICENSE` im Repository-Root.

## Weitere Ressourcen

- **ADR-Dokumentation**: `docs/adr/`
- **Changelog**: `CHANGELOG.md`
- **Beispielkonfiguration**: `config.toml.example`
- **Projektziele**: Issue #1 (Foundation & Project Setup)
