# Worker Pool Benchmark Analyse

**Datum:** 2025-10-17  
**System:** AMD EPYC 7763 64-Core Processor (4 Cores verfügbar)  
**Go Version:** 1.24.7

## Zusammenfassung der Ergebnisse

### CPU-Bound Workloads (100 Jobs)

| Workers | ns/op    | Speedup | B/op    | allocs/op |
|---------|----------|---------|---------|-----------|
| 1       | 84,837   | 1.00x   | 24,116  | 315       |
| 2       | 89,203   | 0.95x   | 24,216  | 316       |
| 4       | 81,477   | 1.04x   | 24,409  | 318       |
| 8       | 81,030   | 1.05x   | 24,798  | 322       |
| 16      | 84,489   | 1.00x   | 25,703  | 330       |

**Erkenntnis:** Bei CPU-bound Workloads zeigt sich nur minimaler Performancevorteil durch zusätzliche Worker (ca. 4-5% bei 4-8 Workers). Dies ist erwartbar, da die Testumgebung 4 Cores hat und die simulierte CPU-Arbeit minimal ist.

### I/O-Bound Workloads (100 Jobs, 10ms I/O-Delay pro Job)

| Workers | ns/op         | Speedup | B/op    | allocs/op |
|---------|---------------|---------|---------|-----------|
| 1       | 1,019,243,513 | 1.00x   | 26,152  | 226       |
| 2       | 509,934,838   | 2.00x   | 25,404  | 219       |
| 4       | 255,123,561   | 4.00x   | 25,654  | 222       |
| 8       | 132,392,637   | 7.70x   | 26,479  | 231       |
| 16      | 71,419,460    | 14.27x  | 28,032  | 246       |

**Erkenntnis:** Bei I/O-bound Workloads zeigt sich **exzellente Skalierung**:
- 2 Workers: ~2x schneller
- 4 Workers: ~4x schneller  
- 8 Workers: ~7.7x schneller
- 16 Workers: ~14.3x schneller

Dies demonstriert den großen Vorteil von Worker Pools bei I/O-lastigen Operationen wie Datei-Parsing oder Datenbank-Queries.

## Empfehlungen für EVE SDE Database Builder

### Optimale Worker-Anzahl

Basierend auf den Benchmarks und dem Use Case (Parsing großer JSONL-Dateien):

1. **Für CPU-intensive Operationen (Parsing, Transformation):**
   - **Empfehlung: 4-8 Workers**
   - Entspricht typischerweise `runtime.NumCPU()` oder `runtime.NumCPU() * 2`
   - Minimale Overhead-Zunahme bei solider Parallelität

2. **Für I/O-intensive Operationen (Datei-Lesen, DB-Writes):**
   - **Empfehlung: 8-16 Workers**
   - Maximiert Durchsatz bei I/O-Wartezeiten
   - 16 Workers zeigt 14x Speedup gegenüber single-threaded

3. **Für gemischte Workloads (typischer Use Case):**
   - **Empfehlung: 8 Workers als Default**
   - Guter Kompromiss zwischen CPU- und I/O-Performance
   - Skaliert gut auf multi-core Systemen

### Implementierungshinweise

```go
// Empfohlene Standard-Konfiguration
func DefaultWorkerCount() int {
    numCPU := runtime.NumCPU()
    
    // Für I/O-lastige Workloads: 2x CPU Cores
    // Für CPU-lastige Workloads: 1x CPU Cores
    return numCPU * 2
}

// Für große Import-Operationen
func OptimalWorkerCount() int {
    numCPU := runtime.NumCPU()
    
    // Min 4, Max 16, Default 2x Cores
    workers := numCPU * 2
    if workers < 4 {
        workers = 4
    }
    if workers > 16 {
        workers = 16
    }
    return workers
}
```

## Memory Profiling

Die Benchmarks zeigen moderate Memory-Nutzung:
- **1 Worker:** ~24KB pro Operation, 315 Allocations
- **16 Workers:** ~28KB pro Operation, 330 Allocations

**Erkenntnis:** Memory Overhead durch zusätzliche Workers ist minimal (~16% mehr Allocations bei 16x Workers). Der Worker Pool ist memory-effizient.

## CPU Profiling

Für detailliertes CPU Profiling kann folgender Befehl verwendet werden:

```bash
# CPU Profile erstellen
go test -bench=BenchmarkPool_8Workers$ -cpuprofile=/tmp/cpu.prof ./internal/worker/

# Profile analysieren
go tool pprof -http=:8080 /tmp/cpu.prof
```

## Fazit

Der Worker Pool zeigt **exzellente Performance-Charakteristiken** für I/O-bound Workloads mit nahezu linearer Skalierung bis 16 Workers. Für den EVE SDE Database Builder Use Case (paralleles Parsing und Import von JSONL-Dateien) wird eine **Standard-Konfiguration von 8 Workers** empfohlen, mit der Möglichkeit, dies basierend auf `runtime.NumCPU()` anzupassen.

## Nächste Schritte

- [ ] CPU Profiling mit `pprof` für detaillierte Analyse
- [ ] Memory Profiling mit `-memprofile` 
- [ ] Benchmarks mit realen EVE SDE Dateien (z.B. invTypes mit ~20k Einträgen)
- [ ] Load Testing mit simulierten Production Workloads
