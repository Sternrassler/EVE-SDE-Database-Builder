# Fuzz Testing für JSONL Parser

## Übersicht

Dieses Projekt verwendet Go's native Fuzz Testing (seit Go 1.18) zur Sicherstellung der Parser-Robustheit gegen fehlerhafte, ungültige und Edge-Case-Eingaben.

## Zielsetzung

- **Robustheit**: Parser soll bei allen Eingaben crash-frei bleiben
- **Sicherheit**: Keine Panics oder unerwartete Fehler bei malformed Input
- **Edge Cases**: Automatische Erkennung von Grenzfällen durch Coverage-Guided Fuzzing

## Implementierung

### Fuzz Test Funktionen

Das Projekt enthält drei spezialisierte Fuzz Tests in `internal/parser/parser_fuzz_test.go`:

1. **FuzzJSONLParser** - Allgemeiner JSONL Parser Test
   - Testet grundlegende JSONL Parsing-Logik
   - Seed Corpus: Valide JSON Objekte, leere Zeilen, Unicode, etc.
   - Fokus: JSON Syntax-Varianten

2. **FuzzJSONLParserNestedData** - Verschachtelte Strukturen
   - Testet Parser mit komplexen, verschachtelten JSON-Objekten
   - Seed Corpus: EVE SDE typische Strukturen (Maps, Arrays)
   - Fokus: Nested Objects und Type Safety

3. **FuzzJSONLParserLargeInput** - Große Eingaben
   - Testet Buffer-Handling und Performance
   - Seed Corpus: Progressiv größere JSON-Objekte
   - Fokus: Memory Safety und Scanner Buffer

### Seed Corpus

Initiale Test-Inputs aus realen EVE SDE Daten:

```
internal/parser/testdata/fuzz/
├── corpus1.jsonl          # agtAgentTypes (5 Zeilen)
├── corpus2.jsonl          # chrRaces (5 Zeilen)
├── corpus3.jsonl          # Einfache Test-Daten
├── corpus4_edge.jsonl     # Edge Cases (leere Zeilen)
└── corpus5_unicode.jsonl  # Unicode und Sonderzeichen
```

## Verwendung

### Schneller Test (Entwicklung)

```bash
make fuzz-quick
```

Führt Fuzz Tests für 5 Sekunden aus (~30k Iterationen).

### Voller Test (100k Iterationen)

```bash
make fuzz
```

Führt Fuzz Tests mit ca. 100.000 Iterationen aus (~20 Sekunden pro Test).

### Manuelle Ausführung

Einzelner Fuzz Test:

```bash
go test -v ./internal/parser -fuzz=FuzzJSONLParser -fuzztime=10s
```

Mit spezifischer Iterationsanzahl:

```bash
# Über Script (empfohlen)
bash scripts/run-fuzz-tests.sh 100000

# Mit Zeitvorgabe
FUZZ_TIME=30s bash scripts/run-fuzz-tests.sh
```

### Kontinuierliche Fuzzing

Für längere Fuzzing-Sessions (z.B. über Nacht):

```bash
go test -v ./internal/parser -fuzz=FuzzJSONLParser -fuzztime=1h
```

## Ergebnisse

### Testabdeckung

Die Fuzz Tests erreichen:
- **Basis Coverage**: 81-120 interessante Inputs durch Seed Corpus
- **Erweiterung**: +20-50 neue Inputs durch Coverage-Guided Mutations
- **Durchsatz**: ~5.000-10.000 Executions/Sekunde

### Bekannte Ergebnisse

**Status:** ✅ Alle Tests bestanden (crash-frei)

Typische Fuzz Test Ausgabe:

```
FuzzJSONLParser: 112,735 execs in 20s
  - New interesting: 37
  - Total coverage: 157 inputs
  - Status: PASS ✓

FuzzJSONLParserNestedData: 101,584 execs in 15s
  - New interesting: 110  
  - Total coverage: 113 inputs
  - Status: PASS ✓

FuzzJSONLParserLargeInput: 74,784 execs in 15s
  - New interesting: 122
  - Total coverage: 125 inputs
  - Status: PASS ✓
```

**Interpretation:**
- Parser ist robust gegen zufällige und malformed Eingaben
- Keine Panics oder Crashes bei 280k+ Fuzzing-Iterationen
- Coverage-Guided Fuzzing findet kontinuierlich neue Edge Cases

## Continuous Integration

Die Fuzz Tests sind für schnelle CI-Ausführung konfiguriert:

```yaml
# .github/workflows/test.yml (Beispiel)
- name: Run Fuzz Tests
  run: make fuzz-quick
```

Für nächtliche/wöchentliche umfassende Tests:

```yaml
# .github/workflows/fuzz-nightly.yml
- name: Extended Fuzz Testing
  run: make fuzz
```

## Fehleranalyse

### Crasher Corpus

Falls ein Fuzz Test einen Crash findet, wird der Input automatisch gespeichert:

```
internal/parser/testdata/fuzz/FuzzJSONLParser/
└── <hash> - Crasher Input
```

### Reproduktion

Crasher können deterministisch reproduziert werden:

```bash
go test -v ./internal/parser -run=FuzzJSONLParser/<hash>
```

## Best Practices

1. **Seed Corpus pflegen**: Regelmäßig reale EVE SDE Daten als Seeds hinzufügen
2. **Lange Sessions**: Wöchentlich längere Fuzz-Sessions (>1h) ausführen
3. **Crasher Review**: Alle gefundenen Crasher analysieren und fixen
4. **Coverage Monitoring**: Coverage-Metriken bei großen Parser-Änderungen überprüfen

## Technische Details

### Go Native Fuzzing

- **Engine**: libFuzzer-basiert (Coverage-Guided)
- **Mutationen**: Byte-Level Mutations mit Coverage Feedback
- **Parallelität**: 4 Workers (Standard), konfigurierbar via `-parallel`

### Ressourcen

- [Go Fuzzing Tutorial](https://go.dev/doc/tutorial/fuzz)
- [Go Fuzzing Documentation](https://go.dev/security/fuzz/)
- [ADR-003: JSONL Parser Architecture](../docs/adr/ADR-003-jsonl-parser-architecture.md)

## Zukunft

Geplante Erweiterungen:

- [ ] Strukturbasierte Fuzzing (JSON Schema Awareness)
- [ ] Property-Based Testing Integration
- [ ] Fuzz Corpus aus Production Logs
- [ ] Differential Fuzzing (Vergleich mit alternativen Parsern)

---

**Status:** ✅ Implementiert in v0.2.0  
**Maintainer:** EVE SDE Database Builder Team  
**Letzte Aktualisierung:** 2025-10-18
