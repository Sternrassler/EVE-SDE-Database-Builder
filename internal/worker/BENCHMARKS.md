# Worker Pool Benchmarks

Diese Datei dokumentiert die verfügbaren Benchmarks für den Worker Pool und wie diese ausgeführt werden.

## Übersicht

Der Worker Pool bietet verschiedene Benchmark-Suiten zur Performance-Analyse:

1. **CPU-Bound Benchmarks** - Simulieren CPU-intensive Operationen
2. **I/O-Bound Benchmarks** - Simulieren I/O-Wartezeiten (Datei-Lesen, DB-Queries)
3. **Mixed Load Benchmarks** - Kombinieren CPU- und I/O-Operationen
4. **Scaling Benchmarks** - Testen Skalierungsverhalten mit verschiedenen Job-Anzahlen

## Benchmark-Ausführung

### Alle Worker Pool Benchmarks ausführen

```bash
go test -bench=^BenchmarkPool -benchmem ./internal/worker/
```

### CPU-Bound Benchmarks (1, 2, 4, 8, 16 Workers)

```bash
go test -bench='^BenchmarkPool_[0-9]+Workers$' -benchmem ./internal/worker/
```

### I/O-Bound Benchmarks

```bash
go test -bench='IOBound' -benchmem ./internal/worker/
```

### Mixed Load Benchmarks

```bash
go test -bench='MixedLoad' -benchmem ./internal/worker/
```

### Spezifische Worker-Anzahl

```bash
# Nur 8 Workers
go test -bench='BenchmarkPool_8Workers' -benchmem ./internal/worker/

# 4 und 8 Workers
go test -bench='BenchmarkPool_[48]Workers$' -benchmem ./internal/worker/
```

## CPU & Memory Profiling

### CPU Profiling

```bash
# CPU Profile erstellen
go test -bench=BenchmarkPool_8Workers$ -cpuprofile=/tmp/cpu.prof ./internal/worker/

# Profile analysieren (interaktiv)
go tool pprof /tmp/cpu.prof

# Profile visualisieren (Web UI)
go tool pprof -http=:8080 /tmp/cpu.prof
```

### Memory Profiling

```bash
# Memory Profile erstellen
go test -bench=BenchmarkPool_8Workers$ -memprofile=/tmp/mem.prof ./internal/worker/

# Profile analysieren
go tool pprof /tmp/mem.prof

# Top Memory Allocations anzeigen
go tool pprof -top /tmp/mem.prof
```

### Kombiniertes Profiling

```bash
# CPU + Memory + Allocations
go test -bench=BenchmarkPool_8Workers$ \
  -cpuprofile=/tmp/cpu.prof \
  -memprofile=/tmp/mem.prof \
  -benchmem \
  ./internal/worker/
```

## Benchmark-Ergebnisse interpretieren

### Output Format

```
BenchmarkPool_8Workers-4       27980     83432 ns/op   24796 B/op     322 allocs/op
```

- **27980**: Anzahl der Iterationen
- **83432 ns/op**: Durchschnittliche Zeit pro Operation (Nanosekunden)
- **24796 B/op**: Bytes allokiert pro Operation
- **322 allocs/op**: Anzahl Allocations pro Operation
- **-4**: GOMAXPROCS (Anzahl verfügbarer CPU Cores)

### Vergleich zwischen Worker-Anzahlen

```bash
# Benchmark mit unterschiedlichen Worker-Anzahlen vergleichen
go test -bench='^BenchmarkPool_[0-9]+Workers$' -benchmem ./internal/worker/ | tee benchmark.txt

# Ergebnisse mit benchstat vergleichen (falls installiert)
benchstat benchmark.txt
```

## Erweiterte Optionen

### Längere Benchmark-Laufzeit

```bash
# 5 Sekunden pro Benchmark (für stabilere Ergebnisse)
go test -bench=BenchmarkPool_8Workers -benchtime=5s ./internal/worker/
```

### Mehrere Durchläufe für Stabilität

```bash
# 10 Durchläufe für statistische Signifikanz
go test -bench=BenchmarkPool_8Workers -count=10 ./internal/worker/
```

### Nur Kompilieren (keine Ausführung)

```bash
# Benchmarks kompilieren aber nicht ausführen
go test -c -o worker_bench.test ./internal/worker/
```

## Analyse der Ergebnisse

Siehe `docs/worker-pool-benchmark-analysis.md` für eine detaillierte Analyse der Benchmark-Ergebnisse und Empfehlungen zur optimalen Worker-Anzahl.

### Wichtigste Erkenntnisse

- **CPU-Bound**: 4-8 Workers optimal (entspricht CPU Cores)
- **I/O-Bound**: 8-16 Workers zeigt beste Performance (bis zu 14x Speedup)
- **Mixed Load**: 8 Workers als guter Kompromiss empfohlen
- **Memory Overhead**: Minimal (~16% mehr Allocations bei 16x Workers vs 1 Worker)

## Troubleshooting

### Benchmarks laufen zu lange

Reduziere `-benchtime`:
```bash
go test -bench=BenchmarkPool -benchtime=1s ./internal/worker/
```

### Inkonsistente Ergebnisse

Erhöhe `-count` für mehr Durchläufe:
```bash
go test -bench=BenchmarkPool_8Workers -count=20 ./internal/worker/
```

### System-Last beeinträchtigt Ergebnisse

Führe Benchmarks auf dediziertem System aus oder:
```bash
# Reduziere GOMAXPROCS
GOMAXPROCS=1 go test -bench=BenchmarkPool ./internal/worker/
```

## Kontinuierliche Performance-Überwachung

Zur Erkennung von Performance-Regressionen:

```bash
# Baseline erstellen
go test -bench=BenchmarkPool -benchmem ./internal/worker/ > baseline.txt

# Nach Änderungen: Vergleich
go test -bench=BenchmarkPool -benchmem ./internal/worker/ > new.txt
benchstat baseline.txt new.txt
```
