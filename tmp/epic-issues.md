# Epic Issue Templates für GitHub

## Epic 1: Foundation & Project Setup

**Titel:** Epic: Foundation & Project Setup

**Labels:** `epic: foundation`, `type: feature`, `area: go-migration`, `priority: high`

**Body:**

```markdown
## Ziel

Basis-Infrastruktur für Go-Migration etablieren: Core Libraries, Logging, Error Handling, Configuration Management.

## Kontext

- **ADR-004:** Configuration Format (TOML + Env + Flags)
- **ADR-005:** Error Handling Strategy (Custom Types + zerolog + Retry)

## Akzeptanzkriterien

- [ ] Logger initialisierbar (zerolog, JSON/Text Format)
- [ ] Config-Loading aus TOML + Env + CLI Flags
- [ ] Custom Error Types (Fatal/Retryable/Validation/Skippable)
- [ ] Retry-Pattern für transiente Fehler
- [ ] Unit Tests (>80% Coverage)

## Tasks

### Logging (`internal/logger/`)
- [ ] #ISSUE_NR: Implementiere zerolog Wrapper
- [ ] #ISSUE_NR: Logger Initialization (Level, Format)
- [ ] #ISSUE_NR: Structured Logging Helper (LogError, LogInfo, etc.)
- [ ] #ISSUE_NR: Tests für Logger

### Error Handling (`internal/errors/`)
- [ ] #ISSUE_NR: Custom Error Types (AppError, ErrorType enum)
- [ ] #ISSUE_NR: Helper Constructors (Fatal, Retryable, Validation, Skippable)
- [ ] #ISSUE_NR: Error Context Management (WithContext)
- [ ] #ISSUE_NR: Tests für Error Classification

### Retry Pattern (`internal/retry/`)
- [ ] #ISSUE_NR: Exponential Backoff Implementation
- [ ] #ISSUE_NR: Context Cancellation Support
- [ ] #ISSUE_NR: Configurable Retry Policy
- [ ] #ISSUE_NR: Tests für Retry Logic

### Configuration (bereits vorhanden, Tests fehlen)
- [ ] #ISSUE_NR: Unit Tests für Config Loading
- [ ] #ISSUE_NR: Env Var Override Tests
- [ ] #ISSUE_NR: Validation Tests

## Abhängigkeiten

Keine (Foundation Epic)

## Schätzung

**Aufwand:** ~3-5 Tage  
**Komplexität:** Mittel
```

---

## Epic 2: Database Layer (SQLite)

**Titel:** Epic: Database Layer Implementation (SQLite)

**Labels:** `epic: database-layer`, `component: database`, `type: feature`, `priority: critical`

**Body:**

```markdown
## Ziel

SQLite Database Layer mit Batch Insert, Migrations, Performance-Optimierung implementieren.

## Kontext

- **ADR-001:** SQLite-Only Approach
- **ADR-002:** Database Layer Design (database/sql + sqlx, kein ORM)

## Akzeptanzkriterien

- [ ] SQLite Connection mit WAL Mode + PRAGMAs
- [ ] Batch Insert für 10k+ Rows (< 2s)
- [ ] Transaction Support + Rollback
- [ ] Schema Migrations (golang-migrate)
- [ ] Unit Tests mit In-Memory SQLite

## Tasks

### Core Driver (`internal/database/`)
- [ ] #ISSUE_NR: SQLite Connection Setup (NewDB, PRAGMA Optimierungen)
- [ ] #ISSUE_NR: Batch Insert Implementation
- [ ] #ISSUE_NR: Transaction Wrapper (BeginTx, Commit, Rollback)
- [ ] #ISSUE_NR: Query Helper (Select, Get via sqlx)
- [ ] #ISSUE_NR: Connection Pool Config (SetMaxOpenConns = 1)

### Schema Migrations (`migrations/sqlite/`)
- [ ] #ISSUE_NR: Migration Files (invTypes, invGroups, invCategories)
- [ ] #ISSUE_NR: Migration Files (blueprints, dogmaAttributes, dogmaEffects)
- [ ] #ISSUE_NR: Migration Files (universe: solarSystems, regions, etc.)
- [ ] #ISSUE_NR: Migration Automation (make migrate-up, migrate-down)

### Testing
- [ ] #ISSUE_NR: In-Memory SQLite Tests
- [ ] #ISSUE_NR: Batch Insert Benchmarks (10k, 100k Rows)
- [ ] #ISSUE_NR: Transaction Rollback Tests
- [ ] #ISSUE_NR: Schema Migration Integration Tests

## Abhängigkeiten

- **Epic 1:** Foundation (Error Handling, Retry Pattern)

## Schätzung

**Aufwand:** ~5-7 Tage  
**Komplexität:** Hoch (Performance-kritisch)
```

---

## Epic 3: JSONL Parser Migration

**Titel:** Epic: JSONL Parser Migration (50+ Tables)

**Labels:** `epic: parser-migration`, `component: parser`, `type: feature`, `priority: critical`

**Body:**

```markdown
## Ziel

Code-Generation für alle 50+ SDE-Tabellen via RIFT Schema + quicktype. Generic JSONL-Parser implementieren.

## Kontext

- **ADR-003:** JSONL Parser Architecture (Full Code-Gen)
- **RIFT SDE:** https://sde.riftforeve.online/ (Schema-Quelle)

## Akzeptanzkriterien

- [ ] Code-Gen Tools (scrape-rift-schemas.go, add-tomap-methods.go)
- [ ] 50+ Generated Structs in `internal/parser/generated/`
- [ ] Generic ParseJSONL[T] Funktion
- [ ] ToMap() Methods für DB-Insert
- [ ] Unit Tests für Top-10 Parser

## Tasks

### Code-Gen Tools (`tools/`)
- [ ] #ISSUE_NR: scrape-rift-schemas.go (RIFT HTML → JSON Schema)
- [ ] #ISSUE_NR: add-tomap-methods.go (Post-Processing)
- [ ] #ISSUE_NR: Makefile Target `make generate-parsers`
- [ ] #ISSUE_NR: CI Check (generierter Code aktuell?)

### Generic Parser (`internal/parser/`)
- [ ] #ISSUE_NR: ParseJSONL[T] Implementation
- [ ] #ISSUE_NR: ParseJSONLStream (Streaming Variant)
- [ ] #ISSUE_NR: Error Handling (Skippable Line Errors)
- [ ] #ISSUE_NR: Context Cancellation Support

### Generated Structs (`internal/parser/generated/`)
- [ ] #ISSUE_NR: Execute Code-Gen (50+ Files)
- [ ] #ISSUE_NR: Review Top-10 Structs (types, blueprints, dogma)
- [ ] #ISSUE_NR: Manual Fixes (falls nötig)

### Testing
- [ ] #ISSUE_NR: Unit Tests für types.jsonl
- [ ] #ISSUE_NR: Unit Tests für blueprints.jsonl
- [ ] #ISSUE_NR: Integration Test (JSONL → SQLite)
- [ ] #ISSUE_NR: Error Handling Tests (malformed JSONL)

## Abhängigkeiten

- **Epic 2:** Database Layer (für ToMap() → BatchInsert)

## Schätzung

**Aufwand:** ~7-10 Tage  
**Komplexität:** Sehr Hoch (50+ Tabellen, Code-Gen)
```

---

## Epic 4: Worker Pool & Concurrency

**Titel:** Epic: Worker Pool & Parallel Import

**Labels:** `epic: foundation`, `component: worker`, `type: feature`, `area: performance`, `priority: high`

**Body:**

```markdown
## Ziel

Worker Pool Pattern für parallelen JSONL-Import implementieren. 2-Phase Import (Parse || Insert).

## Kontext

- **ADR-006:** Concurrency & Worker Pool Pattern
- **Performance-Ziel:** ~3.5 min (vs VB.NET ~8 min)

## Akzeptanzkriterien

- [ ] Worker Pool (N Workers, M Tasks)
- [ ] Buffered Channels (Backpressure)
- [ ] Context Cancellation (Graceful Shutdown)
- [ ] 2-Phase Import (Parse parallel → Insert sequential)
- [ ] Benchmark: <4 min für Full SDE Import

## Tasks

### Worker Pool (`internal/worker/`)
- [ ] #ISSUE_NR: Pool Implementation (Start, Submit, Wait)
- [ ] #ISSUE_NR: Worker Goroutines (Task Processing)
- [ ] #ISSUE_NR: Buffered Channels (Task Queue, Result Queue)
- [ ] #ISSUE_NR: Context Cancellation (SIGINT/SIGTERM)

### Import Orchestrator (`cmd/esdedb/`)
- [ ] #ISSUE_NR: 2-Phase Import Logic
- [ ] #ISSUE_NR: File Discovery (sde-JSONL Directory Scan)
- [ ] #ISSUE_NR: Progress Tracking (optional)
- [ ] #ISSUE_NR: Error Aggregation (alle Worker-Fehler sammeln)

### Testing
- [ ] #ISSUE_NR: Unit Tests für Worker Pool
- [ ] #ISSUE_NR: Cancellation Tests (Context)
- [ ] #ISSUE_NR: Integration Test (End-to-End Import)
- [ ] #ISSUE_NR: Benchmarks (Workers=1 vs 4 vs 8)

## Abhängigkeiten

- **Epic 2:** Database Layer
- **Epic 3:** JSONL Parser

## Schätzung

**Aufwand:** ~4-6 Tage  
**Komplexität:** Hoch (Concurrency-Patterns)
```

---

## Epic 5: CLI Interface

**Titel:** Epic: CLI Interface & Commands

**Labels:** `epic: cli-interface`, `component: cli`, `type: feature`, `priority: medium`

**Body:**

```markdown
## Ziel

Vollständiges CLI Interface mit Commands (import, config, version) und User-Feedback.

## Kontext

- **CLI Framework:** cobra
- **Commands:** import, config init/convert, version

## Akzeptanzkriterien

- [ ] `esdedb import` Command (vollständig)
- [ ] `esdedb config init` (generate config.toml)
- [ ] `esdedb config convert` (XML → TOML)
- [ ] `esdedb version` (Version + Commit Hash)
- [ ] Progress Bar (optional)

## Tasks

### Commands (`cmd/esdedb/`)
- [ ] #ISSUE_NR: Import Command (CLI Flags, Config Loading)
- [ ] #ISSUE_NR: Config Init Command
- [ ] #ISSUE_NR: Config Convert Command (XML → TOML Migration)
- [ ] #ISSUE_NR: Version Command

### User Feedback
- [ ] #ISSUE_NR: Progress Bar (pb/v3) für Import
- [ ] #ISSUE_NR: Structured Output (Table für Summary)
- [ ] #ISSUE_NR: Error Messages (User-freundlich)

### Testing
- [ ] #ISSUE_NR: CLI Integration Tests
- [ ] #ISSUE_NR: Flag Parsing Tests
- [ ] #ISSUE_NR: Help Text Validation

## Abhängigkeiten

- **Epic 2:** Database Layer
- **Epic 3:** Parser
- **Epic 4:** Worker Pool

## Schätzung

**Aufwand:** ~3-4 Tage  
**Komplexität:** Mittel
```

---

## Epic 6: Testing & Quality

**Titel:** Epic: Testing Strategy & Quality Gates

**Labels:** `epic: testing`, `type: test`, `area: performance`, `priority: high`

**Body:**

```markdown
## Ziel

Umfassende Testing-Strategie: Unit, Integration, E2E, Benchmarks, Fuzz Tests.

## Kontext

- **Test Coverage:** >80% (kritische Pfade: 100%)
- **Benchmarks:** Performance-Regression verhindern

## Akzeptanzkriterien

- [ ] Unit Tests (alle Module)
- [ ] Integration Tests (JSONL → SQLite)
- [ ] E2E Tests (Full SDE Import)
- [ ] Benchmarks (Parser, Database, Worker Pool)
- [ ] Fuzz Tests (JSONL Parser)

## Tasks

### Unit Tests
- [ ] #ISSUE_NR: Config Tests
- [ ] #ISSUE_NR: Logger Tests
- [ ] #ISSUE_NR: Error Handling Tests
- [ ] #ISSUE_NR: Retry Pattern Tests

### Integration Tests
- [ ] #ISSUE_NR: Parser → Database Integration
- [ ] #ISSUE_NR: Worker Pool Integration
- [ ] #ISSUE_NR: CLI Integration

### E2E Tests
- [ ] #ISSUE_NR: Full Import Test (sample SDE)
- [ ] #ISSUE_NR: Config File Loading Test
- [ ] #ISSUE_NR: Error Recovery Test

### Benchmarks
- [ ] #ISSUE_NR: JSONL Parsing Benchmarks
- [ ] #ISSUE_NR: Batch Insert Benchmarks
- [ ] #ISSUE_NR: Worker Pool Scaling Benchmarks

### Fuzz Tests
- [ ] #ISSUE_NR: JSONL Parser Fuzz Tests (malformed JSON)

## Abhängigkeiten

- Alle anderen Epics (Testing nach Implementierung)

## Schätzung

**Aufwand:** ~5-7 Tage  
**Komplexität:** Hoch (vollständige Coverage)
```

---

## Epic 7: Documentation & Migration Guide

**Titel:** Epic: Documentation & VB.NET Migration Guide

**Labels:** `type: docs`, `area: go-migration`, `priority: medium`

**Body:**

```markdown
## Ziel

Vollständige Dokumentation für User und Entwickler. Migration Guide für VB.NET → Go Transition.

## Akzeptanzkriterien

- [ ] README.md komplett
- [ ] Migration Guide (VB.NET → Go)
- [ ] API Documentation (godoc)
- [ ] Deployment Guide (Binaries, Docker)
- [ ] Troubleshooting Guide

## Tasks

### User Documentation
- [ ] #ISSUE_NR: README Update (Installation, Usage)
- [ ] #ISSUE_NR: Configuration Guide (TOML Syntax)
- [ ] #ISSUE_NR: CLI Command Reference
- [ ] #ISSUE_NR: FAQ

### Developer Documentation
- [ ] #ISSUE_NR: Architecture Documentation (Update)
- [ ] #ISSUE_NR: Code-Gen Documentation
- [ ] #ISSUE_NR: Testing Guide
- [ ] #ISSUE_NR: Contributing Guide

### Migration Guide
- [ ] #ISSUE_NR: VB.NET → Go Migration Guide
- [ ] #ISSUE_NR: Breaking Changes Documentation
- [ ] #ISSUE_NR: Config Conversion Guide (XML → TOML)
- [ ] #ISSUE_NR: Performance Comparison

## Abhängigkeiten

- Alle anderen Epics (Dokumentation nach Implementierung)

## Schätzung

**Aufwand:** ~3-4 Tage  
**Komplexität:** Niedrig
```
