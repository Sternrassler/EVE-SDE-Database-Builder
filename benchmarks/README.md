# Benchmark Baselines

Dieses Verzeichnis enthält Baseline-Dateien für Performance-Regression Tests.

## Struktur

- `baseline.txt` - Aktuelle Benchmark-Baseline (alle Pakete)
- `baseline-worker.txt` - Worker Pool Benchmarks
- `baseline-parser.txt` - Parser Benchmarks
- `baseline-database.txt` - Database Benchmarks

## Verwendung

### Baseline erstellen

```bash
# Alle Benchmarks als Baseline speichern
make bench-baseline

# Oder manuell für einzelne Pakete
go test -bench=. -benchmem ./internal/worker/ > benchmarks/baseline-worker.txt
go test -bench=. -benchmem ./internal/parser/ > benchmarks/baseline-parser.txt
go test -bench=. -benchmem ./internal/database/ > benchmarks/baseline-database.txt
```

### Gegen Baseline vergleichen

```bash
# Vergleich mit benchstat (installiert via: go install golang.org/x/perf/cmd/benchstat@latest)
make bench-compare

# Oder manuell
go test -bench=. -benchmem ./internal/worker/ > /tmp/new.txt
benchstat benchmarks/baseline-worker.txt /tmp/new.txt
```

## CI Integration

Der GitHub Actions Workflow `.github/workflows/benchmark.yml` führt automatisch Benchmarks aus und vergleicht sie mit der Baseline. Bei signifikanten Regressionen (>10% langsamer) wird eine Warnung ausgegeben.

## Baseline aktualisieren

Die Baseline sollte aktualisiert werden:
- Nach Performance-Optimierungen
- Nach signifikanten Architektur-Änderungen
- Bei Hardware-Änderungen der CI-Umgebung

```bash
# Baseline aktualisieren und committen
make bench-baseline
git add benchmarks/
git commit -m "chore: Update benchmark baseline"
```

## Benchmark-Metriken

Die Baselines enthalten folgende Metriken:
- **ns/op**: Nanosekunden pro Operation
- **B/op**: Bytes alloziert pro Operation
- **allocs/op**: Anzahl Allocations pro Operation

## Regressions-Schwellwerte

- **Performance-Regression**: >10% langsamer (ns/op)
- **Memory-Regression**: >20% mehr Allocations (B/op oder allocs/op)
