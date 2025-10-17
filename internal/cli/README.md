# Progress Bar Implementation

## Übersicht

Das `internal/cli` Package stellt eine erweiterte Progress Bar für Import-Operationen bereit.

## Features

### 1. Live-Update (Zeilen/Sekunde)
Die Progress Bar zeigt in Echtzeit die Anzahl der verarbeiteten Zeilen pro Sekunde an:
```
[cyan]Importing[reset] [yellow]2500 rows/s[reset] [blue]ETA: 2m 30s[reset]
```

### 2. ETA Anzeige
Die geschätzte verbleibende Zeit (Estimated Time to Arrival) wird basierend auf dem aktuellen Durchsatz berechnet und angezeigt.

### 3. Spinner für einzelne Dateien
Optional kann ein animierter Spinner für die Verarbeitung einzelner Dateien aktiviert werden:
```go
pb := cli.NewProgressBar(cli.ProgressBarConfig{
    Total:       100,
    ShowSpinner: true,
})

pb.StartSpinner("processing file.jsonl")
// ... process file ...
pb.StopSpinner()
```

## Verwendung

### Basis-Beispiel

```go
import (
    "context"
    "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/cli"
    "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/worker"
)

// Create progress bar
pb := cli.NewProgressBar(cli.ProgressBarConfig{
    Total:       100,
    Description: "Importing files",
    Width:       40,
    UpdateRate:  100 * time.Millisecond,
})

// Create progress tracker
tracker := worker.NewProgressTracker(100)
tracker.SetTotalRows(10000)

// Start progress bar in background
ctx, cancel := context.WithCancel(context.Background())
go pb.Start(ctx, tracker)

// ... perform import operations ...
// Update tracker with: tracker.Update(filesProcessed, rowsInserted)

// Stop progress bar
cancel()
pb.Finish()
```

### Integration im Import Command

Die Progress Bar ist bereits im `esdedb import` Command integriert:

```bash
esdedb import --sde-dir ./sde-JSONL --db ./eve-sde.db --workers 4
```

Die Ausgabe zeigt:
- Anzahl verarbeiteter Dateien
- Aktuelle Verarbeitungsgeschwindigkeit (Zeilen/Sekunde)
- Geschätzte verbleibende Zeit (ETA)
- Fortschrittsbalken mit Prozentanzeige

## Konfiguration

### ProgressBarConfig Optionen

| Option | Typ | Standard | Beschreibung |
|--------|-----|----------|--------------|
| `Total` | `int` | - | Gesamtzahl der zu verarbeitenden Elemente (erforderlich) |
| `Description` | `string` | "Processing" | Beschreibung der Progress Bar |
| `Width` | `int` | 40 | Breite der Progress Bar in Zeichen |
| `UpdateRate` | `time.Duration` | 100ms | Rate, mit der die Progress Bar aktualisiert wird |
| `ShowSpinner` | `bool` | false | Aktiviert Spinner für einzelne Dateien |
| `Output` | `io.Writer` | os.Stdout | Ausgabeziel für die Progress Bar |

## Implementierungsdetails

### Architektur

Die Progress Bar besteht aus zwei Hauptkomponenten:

1. **ProgressBar**: Hauptkomponente, die die Anzeige verwaltet
   - Nutzt `github.com/schollz/progressbar/v3` als Basis
   - Überwacht einen `worker.ProgressTracker`
   - Aktualisiert die Anzeige periodisch mit Live-Metriken

2. **Spinner**: Optional aktivierbare Komponente für Datei-spezifische Anzeigen
   - Rotierender Unicode-Spinner
   - Zeigt aktuell verarbeitete Datei an
   - Unabhängig von der Haupt-Progress Bar

### Thread-Safety

- Alle Komponenten sind Thread-Safe
- Die Progress Bar läuft in einer eigenen Goroutine
- Der Spinner läuft in einer separaten Goroutine
- Synchronisation erfolgt über Channels und Context

### Performance

- Minimaler Overhead durch atomic Operations im ProgressTracker
- Konfigurierbare Update-Rate zur Balance zwischen Responsiveness und CPU-Last
- Benchmarks verfügbar in `progress_test.go`

## Tests

Vollständige Test-Suite verfügbar:

```bash
go test -v ./internal/cli/...
```

Enthält:
- Unit Tests für alle Komponenten
- Concurrency Tests
- Example Tests zur Demonstration der Verwendung
- Benchmarks für Performance-Validierung

## Beispiele

Siehe `internal/cli/example_test.go` für vollständige Beispiele:

- `Example_progressBarBasicUsage`: Basis-Verwendung
- `Example_progressBarWithSpinner`: Mit Datei-Spinner
- `Example_progressBarLiveMetrics`: Mit Live-Metriken

## Zukünftige Erweiterungen

Mögliche Verbesserungen:
- [ ] Multi-Progress-Bar-Support für parallele Operationen
- [ ] Customizable Themes
- [ ] Pausieren/Fortsetzen der Anzeige
- [ ] Export von Metriken (JSON, CSV)
