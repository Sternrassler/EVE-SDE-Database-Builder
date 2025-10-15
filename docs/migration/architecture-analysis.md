# Architecture Analysis: VB.NET → Go Migration

**Projekt:** EVE SDE Database Builder  
**Datum:** 2025-10-15  
**Status:** Initial Analysis  
**Ziel:** Vollständiges Refactoring von VB.NET/WinForms zu Go (CLI-basiert)

---

## Executive Summary

Das EVE SDE Database Builder Tool importiert EVE Online Static Data Export (SDE) YAML-Dateien in verschiedene Datenbanksysteme. Die bestehende VB.NET-Implementierung ist GUI-basiert und unterstützt 5 Datenbanktypen mit 70+ YAML-Parser-Klassen.

**Migrationsziel:** Modernes Go CLI-Tool mit **exklusivem SQLite-Support**, verbesserter Performance, Testbarkeit und Wartbarkeit.

**Scope-Reduktion:** Fokus auf SQLite eliminiert Multi-DB-Komplexität und beschleunigt Migration erheblich.

---

## 1. Bestehende Architektur (VB.NET)

### 1.1 Hauptkomponenten

```txt
EVE SDE Database Builder/
├── Forms (GUI Layer)
│   ├── frmMain.vb           # Hauptfenster (2237 Zeilen)
│   ├── frmAbout.vb          # About Dialog
│   ├── frmError.vb          # Error Handler
│   └── frmThreadSelect.vb   # Thread Configuration
│
├── Database Classes/
│   ├── DBFilesBase.vb       # Abstrakte Basis (204 Zeilen)
│   ├── SQLiteDB.vb          # SQLite Driver ⭐ **Migration Target**
│   ├── MySQLDB.vb           # MySQL Driver ❌ **Not Migrating**
│   ├── msSQLDB.vb           # MS SQL Server Driver ❌ **Not Migrating**
│   ├── postgreSQLDB.vb      # PostgreSQL Driver ❌ **Not Migrating**
│   ├── msAccessDB.vb        # MS Access Driver ❌ **Not Migrating**
│   └── CSVDB.vb             # CSV Export (Bulk Insert) ❌ **Not Needed (SQLite native)**
│
├── SDE YAML Classes/ (70+ Dateien)
│   ├── YAMLFilesBase.vb     # Parser Basis
│   ├── YAMLtypes.vb         # EVE Item Types (~460 Zeilen)
│   ├── YAMLblueprints.vb    # Blueprints
│   ├── YAMLdogmaEffects.vb  # Game Mechanics
│   ├── YAMLUniverse.vb      # Universe Data
│   └── ... (66 weitere Parser)
│
├── ESI.vb                   # EVE Swagger Interface Client
├── Globals.vb               # Global State & Enums
├── ProgramSettings.vb       # Configuration Manager
├── ProgramUpdater.vb        # Auto-Update Logic
└── ImportLanguage.vb        # i18n Support
```

### 1.2 Datenfluss

```txt
┌─────────────────┐
│  User (GUI)     │
│  File Selection │
└────────┬────────┘
         │
         ▼
┌────────────────────────┐
│  frmMain.vb            │
│  - Threading Control   │
│  - Progress Tracking   │
└───────┬────────────────┘
        │
        ▼
┌────────────────────────┐
│  YAMLParser Classes    │
│  - YAMLtypes           │
│  - YAMLblueprints      │
│  - ... (70+ Klassen)   │
└───────┬────────────────┘
        │
        ▼
┌────────────────────────┐
│  DBFilesBase           │
│  Database Abstraction  │
└───────┬────────────────┘
        │
        ├─► SQLiteDB
        ├─► MySQLDB
        ├─► PostgreSQLDB
        ├─► msSQLDB
        └─► msAccessDB
```

### 1.3 Externe Abhängigkeiten

| Dependency | Zweck | Go-Äquivalent | Status |
|-----------|-------|---------------|--------|
| `YamlDotNet` | YAML Parsing | `gopkg.in/yaml.v3` | ✅ Migrieren |
| `System.Data.SQLite` | SQLite Driver | `github.com/mattn/go-sqlite3` | ✅ Migrieren |
| `MySql.Data` | MySQL Driver | - | ❌ Out of Scope |
| `Npgsql` | PostgreSQL Driver | - | ❌ Out of Scope |
| `System.Data.SqlClient` | MS SQL Driver | - | ❌ Out of Scope |
| `System.Data.OleDb` | MS Access Driver | - | ❌ Out of Scope |

### 1.4 Kerndaten-Strukturen

**Datenbanktypen (Enum):**

```vb
Public Enum DatabaseType
    SQLite      ⭐ **Only Target**
    MySQL       ❌ Removed
    PostgreSQL  ❌ Removed
    MSSQL       ❌ Removed
    Access      ❌ Removed
    CSV         ❌ Removed (SQLite native transactions suffice)
End Enum
```

**Import-Parameter:**

```vb
Public Structure ImportParameters
    Dim RowLocation As Integer
    Dim InsertRecords As Boolean
    Dim ImportLanguageCode As LanguageCode
    Dim ReturnList As Boolean
End Structure
```

**Feldtypen:**

```vb
Public Enum FieldType
    int_type, bigint_type, bit_type
    char_type, varchar_type, nvarchar_type
    text_type, ntext_type
    real_type, float_type, double_type
    datetime_type, date_type
End Enum
```

---

## 2. Komplexitätsanalyse

### 2.1 Code-Statistiken

| Metrik | Anzahl | Bemerkung |
|--------|--------|-----------|
| VB.NET Dateien | ~89 | Ohne Designer-Dateien |
| YAML Parser Klassen | 70+ | Je eine pro SDE-Datei-Typ |
| Database Driver | ~~6~~ → **1** | **Nur SQLite** (5 Driver entfernt) |
| Zeilen in frmMain.vb | 2237 | Haupt-GUI-Logik → CLI |
| Zeilen in YAMLtypes.vb | ~460 | Komplexester Parser |
| Threaded Operations | Ja | Parallele YAML-Verarbeitung → Goroutines |

### 2.2 Kritische Pfade

1. **YAML Parsing:** 70+ Klassen mit individueller Logik (sehr heterogen)
2. ~~**Database Abstraction:** Jeder Driver hat eigene SQL-Dialekt-Quirks~~ → **Entfallen** (nur SQLite)
3. **Bulk Inserts:** ~~CSV-basiert~~ → SQLite Batch Transactions (Performance-kritisch)
4. **ESI Integration:** EVE Online API für Live-Daten-Sync
5. **Threading:** Manuelle Thread-Verwaltung → Goroutines (Worker Pools)

### 2.3 Risikobewertung

| Bereich | Risiko | Begründung | Mitigation |
|---------|--------|------------|------------|
| ~~MS Access Support~~ | ~~🔴 Hoch~~ | ~~Windows-only~~ | ✅ **Eliminiert** (Out of Scope) |
| YAML Schema Changes | 🟡 Mittel | CCP Games ändert SDE-Format | Versionierung + Schema-Tests |
| ~~SQL Dialekt-Unterschiede~~ | ~~🟡 Mittel~~ | ~~Jede DB hat Eigenheiten~~ | ✅ **Eliminiert** (nur SQLite) |
| Performance-Regression | 🟡 Mittel | VB.NET nutzt CSV-Bulk-Insert | SQLite Transactions + Benchmarks |
| ESI API Breaking Changes | 🟢 Niedrig | Stabile CCP API | Swagger Codegen |

---

## 3. Go-Architektur (Zielzustand)

### 3.1 Projektstruktur (Standard Go Layout)

```txt
translated/
├── cmd/
│   └── esdedb/
│       └── main.go              # CLI Entry Point
│
├── internal/
│   ├── database/
│   │   ├── sqlite.go            # SQLite Driver (Direct, no abstraction)
│   │   └── schema.go            # Schema Definitions
│   │
│   ├── parser/
│   │   ├── yaml/
│   │   │   ├── base.go          # Common Parser Logic
│   │   │   ├── types.go         # Types Parser
│   │   │   ├── blueprints.go
│   │   │   └── ... (70+ Dateien)
│   │   └── schema.go            # Schema Definitions
│   │
│   ├── esi/
│   │   └── client.go            # ESI API Client
│   │
│   └── config/
│       └── config.go            # Configuration Management
│
├── pkg/
│   └── models/
│       └── sde.go               # Shared Data Models
│
├── migrations/
│   └── sqlite/                  # SQLite Schema Migrations (nur 1 DB)
│
├── scripts/
│   └── generate-parsers.sh      # Code Generation (falls nötig)
│
├── go.mod
├── go.sum
└── README.md
```

### 3.2 Datenfluss (Go)

```txt
┌──────────────────┐
│  CLI (Cobra)     │
│  esdedb import   │
└────────┬─────────┘
         │
         ▼
┌────────────────────────┐
│  Parser Factory        │
│  - Erkennt YAML-Typ    │
│  - Lädt passenden Parser│
└───────┬────────────────┘
        │
        ▼
┌────────────────────────┐
│  YAML Parser           │
│  - gopkg.in/yaml.v3    │
│  - Struct Unmarshaling │
└───────┬────────────────┘
        │
        ▼
┌────────────────────────┐
│  SQLite Driver         │
│  - BatchInsert()       │
│  - Transaction Support│
│  - PRAGMA Optimization│
└───────┬────────────────┘
        │
        └─► SQLite (mattn/go-sqlite3)
```

### 3.3 Technologie-Stack

| Komponente | Bibliothek | Begründung |
|-----------|-----------|------------|
| CLI Framework | `github.com/spf13/cobra` | Standard, mächtig, gut dokumentiert |
| YAML Parsing | `gopkg.in/yaml.v3` | De-facto Standard für Go |
| Database Driver | `github.com/mattn/go-sqlite3` | CGo-basiert, stabil, performant |
| Database Layer | `database/sql` + `sqlx` | **Kein ORM nötig** (nur 1 DB, direktes SQL) |
| Migrations | `github.com/golang-migrate/migrate` | Versioniertes Schema-Management |
| Config Management | `github.com/spf13/viper` | YAML/TOML/ENV Support |
| Logging | `github.com/sirupsen/logrus` | Strukturiertes Logging |
| HTTP Client (ESI) | `github.com/go-resty/resty/v2` | Komfortabler als net/http |
| Progress Bars | `github.com/cheggaaa/pb/v3` | CLI Progress Tracking |
| Testing | `github.com/stretchr/testify` | Assertions + Mocking |

---

## 4. Migrationsstrategie

### 4.1 Phasen

**Phase 1: Foundation (Woche 1-2)**

- Go Module Setup
- CLI Framework (Cobra)
- Config Management
- Logging Infrastructure

**Phase 2: Database Layer (Woche 3)**

- SQLite Driver Implementation (direktes `database/sql` + `sqlx`)
- Migration Scripts (golang-migrate)
- Batch Insert + Transaction Optimization
- PRAGMA Performance Tuning

**Phase 3: YAML Parser Core (Woche 5-8)**

- Base Parser Implementation
- 10 kritischste Parser (types, blueprints, dogmaEffects, etc.)
- Schema Validation
- Unit Tests

**Phase 4: Remaining Parsers (Woche 9-12)**

- 60+ restliche Parser
- Code Generation (falls möglich)
- Integration Tests

**Phase 5: ESI Client (Woche 13)**

- API Client Implementation
- Swagger Codegen (falls verfügbar)
- Rate Limiting

**Phase 6: CLI Interface (Woche 14)**

- Command Structure
- Progress Tracking
- Error Handling
- Help/Documentation

**Phase 7: Testing & Performance (Woche 15-16)**

- E2E Tests
- Performance Benchmarks
- Memory Profiling
- Vergleich mit VB.NET Baseline

**Phase 8: Release (Woche 14-15)**

- Migration Guide (VB.NET → Go)
- Breaking Changes Documentation (Multi-DB → SQLite only)
- Binary Distribution (GitHub Releases)
- Docker Image (optional)

### 4.2 Kritische Entscheidungen (ADRs erforderlich)

1. **ADR-001: SQLite-Only Approach** ⭐ **NEW**
   - Begründung: Scope-Reduktion, Einfachheit, Portabilität
   - Trade-off: ~~Multi-DB~~ vs. Schnellere Entwicklung

2. **ADR-002: Database Layer Design**
   - ~~GORM~~ vs. **sqlx** vs. bare database/sql
   - Empfehlung: `database/sql` + `sqlx` (kein ORM-Overhead bei 1 DB)

3. **ADR-003: YAML Parser Architecture**
   - Code Generation vs. Hand-written
   - Reflection-basiert vs. Typed Structs

4. **ADR-004: Configuration Format**
   - YAML vs. TOML vs. JSON
   - Empfehlung: YAML (Konsistent mit SDE-Daten)

5. **ADR-005: Error Handling Pattern**
   - Wrapped Errors vs. Sentinel Errors
   - Logging-Strategie (stdout vs. file vs. syslog)

6. **ADR-006: Concurrency Model**
   - Worker Pool vs. Goroutines-per-File
   - Backpressure-Mechanismus

---

## 5. Kompatibilitäts-Matrix

### 5.1 Feature-Parität

| Feature | VB.NET | Go (Geplant) | Notes |
|---------|--------|--------------|-------|
| **SQLite Support** | ✅ | ✅ | **Primary Target** |
| ~~MySQL Support~~ | ✅ | ❌ | **Breaking Change** (Out of Scope) |
| ~~PostgreSQL Support~~ | ✅ | ❌ | **Breaking Change** (Out of Scope) |
| ~~MS SQL Support~~ | ✅ | ❌ | **Breaking Change** (Out of Scope) |
| ~~MS Access Support~~ | ✅ | ❌ | **Breaking Change** (Out of Scope) |
| ~~CSV Export~~ | ✅ | ❌ | **Not Needed** (SQLite native) |
| GUI | ✅ | ❌ | **Breaking Change** (CLI stattdessen) |
| Multithreading | ✅ | ✅ | Goroutines (bessere Performance erwartet) |
| ESI Integration | ✅ | ✅ | Full |
| Auto-Update | ✅ | ⚠️ | Optional (GitHub Releases) |
| i18n Support | ✅ | ⚠️ | Low Priority |

### 5.2 Breaking Changes

1. **GUI → CLI:** Keine grafische Oberfläche mehr
2. **Multi-DB → SQLite Only:** Alle Datenbanken außer SQLite entfernt
3. **CSV Export:** Entfernt (SQLite-Transaktionen ausreichend)
4. **Konfigurationsformat:** Neue Config-Struktur (YAML statt Registry/XML)
5. **Auto-Update:** Optional (via GitHub Releases statt custom updater)

---

## 6. Performance-Erwartungen

### 6.1 Benchmarks (Baseline: VB.NET)

| Operation | VB.NET (geschätzt) | Go (Ziel) | Verbesserung |
|-----------|-------------------|-----------|--------------|
| SQLite Bulk Insert (10k rows) | ~5s | ~2s | 2.5x |
| YAML Parse (types.yaml) | ~3s | ~1s | 3x |
| Gesamtimport (SDE Full) | ~15min | ~5min | 3x |
| Memory Footprint | ~500MB | ~100MB | 5x |

**Begründung:**

- Go: Native Compilation, bessere Memory Management
- Goroutines: Effizientere Concurrency als .NET Threads
- Batch Inserts: Optimierte DB-Transaktionen

### 6.2 Testplan

1. **Unit Tests:** Jeder Parser isoliert
2. **Integration Tests:** Database Layer + Parser
3. **E2E Tests:** Vollständiger SDE-Import (SQLite Testcase)
4. **Benchmark Tests:** Performance-Vergleich VB.NET vs. Go
5. **Fuzz Tests:** YAML Parser (ungültige Inputs)

---

## 7. Risiken & Mitigationen

| Risiko | Wahrscheinlichkeit | Impact | Mitigation |
|--------|-------------------|--------|------------|
| SDE Schema Breaking Change | Mittel | Hoch | Schema-Versionierung + Tests |
| Performance-Regression | Niedrig | Hoch | Frühzeitige Benchmarks |
| Scope Creep (70+ Parser) | Hoch | Mittel | Code Generation evaluieren |
| MS Access User Backlash | Niedrig | Niedrig | Migration Guide + Alternativen |
| ESI API Deprecation | Niedrig | Mittel | Swagger-basierte Codegen |

---

## 8. Offene Fragen

1. **Code Generation für Parser:**
   - Können wir aus SDE YAML-Schema Go-Structs generieren?
   - Tools: `go-yaml-tools`, `gojsonschema`

2. **Testing gegen echte SDE-Daten:**
   - Wo bekommen wir aktuelle SDE-Dumps für CI?
   - CCP Mirror? Cached Testdaten?

3. **Distribution:**
   - Binary Releases für Windows/Linux/macOS?
   - Docker Image?

4. **Backward Compatibility:**
   - Sollen alte VB.NET-generierte DBs migriert werden können?
   - Schema-Upgrade-Scripts?

---

## 9. Nächste Schritte

1. ✅ **Architecture Analysis** (dieses Dokument)
2. ⏳ **ADRs schreiben** (6 kritische Entscheidungen)
3. ⏳ **GitHub Project Setup** (Milestones, Labels, Epics)
4. ⏳ **Go Module Init** unter `translated/`
5. ⏳ **Prototyp:** SQLite + 1 Parser (YAMLtypes) als Proof-of-Concept

---

## Anhang A: YAML Parser Inventar (Top 20)

| Parser | Komplexität | Zeilen | Prio |
|--------|------------|--------|------|
| YAMLtypes | ⚠️ Hoch | ~460 | P0 |
| YAMLblueprints | ⚠️ Hoch | ? | P0 |
| YAMLdogmaEffects | ⚠️ Hoch | ? | P0 |
| YAMLdogmaAttributeTypes | ⚠️ Hoch | ? | P0 |
| YAMLgroups | 🟢 Mittel | ? | P1 |
| YAMLcategories | 🟢 Mittel | ? | P1 |
| YAMLmarketGroups | 🟢 Mittel | ? | P1 |
| YAMLUniverse | ⚠️ Hoch | ? | P0 |
| YAMLstaStations | 🟢 Mittel | ? | P2 |
| YAMLfactions | 🟢 Niedrig | ? | P2 |
| ... | | | |
| (60 weitere Parser) | 🟢 Niedrig | <200 | P3 |

**Priorisierung:**

- **P0:** Kritisch für Basis-Funktionalität (Types, Blueprints, Dogma)
- **P1:** Wichtig für vollständigen Import
- **P2:** Nice-to-have, später
- **P3:** Low Priority (selten genutzt)

---

**Autor:** AI Copilot  
**Review:** Pending  
**Version:** 0.1.0-draft
