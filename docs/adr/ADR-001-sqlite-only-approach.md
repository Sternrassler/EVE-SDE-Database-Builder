# ADR-001: SQLite-Only Approach

**Status:** Accepted  
**Datum:** 2025-10-15  
**Entscheider:** Migration Team  
**Kontext:** VB.NET → Go Migration für EVE SDE Database Builder  
**SDE-Version:** 3060393 (15. Oktober 2025)

---

## Kontext & Problem

### EVE SDE Datenquellen

**Primärquellen:**

- [CCP Official SDE](https://developers.eveonline.com/static-data) - Original JSONL-Files
- [RIFT Enhanced SDE](https://sde.riftforeve.online/) - Schema-Dokumentation + Enhanced Version

**Datenformat:** JSONL (JSON Lines)

- 50+ Tabellen (z.B. `types.jsonl`, `blueprints.jsonl`, `dogmaAttributes.jsonl`)
- Enhanced SDE enthält zusätzliche Felder (z.B. Namen für Sterne/Planeten/Monde)
- Download-Größe: ~500 MB (komprimiert)

### Multi-DB Problem

Das bestehende VB.NET Tool unterstützt 6 Datenbanksysteme:

1. SQLite
2. MySQL
3. PostgreSQL
4. MS SQL Server
5. MS Access
6. CSV (Bulk-Insert-Mechanismus)

Jede Datenbank hat:

- Eigenen Driver
- SQL-Dialekt-Unterschiede
- Spezifische Konfigurationsparameter
- Unterschiedliche Performance-Charakteristiken
- Eigene Test-Infrastruktur-Anforderungen

**Problem:** Multi-DB-Support erhöht Komplexität, Entwicklungszeit und Wartungsaufwand erheblich.

---

## Entscheidung

Wir migrieren **ausschließlich SQLite** und entfernen Support für alle anderen Datenbanken.

### Begründung

**1. Nutzungsanalyse (Annahmen basierend auf Use-Case):**

- EVE SDE = statische Daten (kein Multiuser-Zugriff nötig)
- Typischer Workflow: Import → Lokal nutzen/exportieren → Fertig
- Keine verteilte Architektur erforderlich
- SQLite ist **ausreichend** für 99% der Use Cases

**2. Entwicklungsgeschwindigkeit:**

| Aspekt | Multi-DB | SQLite-only | Zeitersparnis |
|--------|----------|-------------|---------------|
| Database Layer | 4 Wochen | 1 Woche | **3 Wochen** |
| SQL-Dialekt-Handling | Komplex | Trivial | - |
| Testing (Integration) | 5x DB-Setups | 1x SQLite | **80% weniger** |
| CI/CD Pipeline | Multi-Container | Embedded | **Massiv** |

**3. Technische Vorteile SQLite:**

- ✅ **Zero-Config:** Keine Server-Installation nötig
- ✅ **Portabel:** Single-File DB, einfach zu teilen
- ✅ **Embedded:** Keine externe Prozesse
- ✅ **Performance:** Exzellent für Read-Heavy Workloads (SDE = statisch)
- ✅ **ACID:** Volle Transaktionsgarantien
- ✅ **Go-Support:** Mature Driver (`mattn/go-sqlite3`)

**4. Scope-Reduktion:**

- Keine Abstraction Layer nötig
- Kein ORM-Overhead (direkt `database/sql` + `sqlx`)
- Einfachere Migrations (nur SQLite-Schema)
- Weniger Dependencies

---

## Konsequenzen

### Positive Konsequenzen

1. **Schnellere Entwicklung:** ~3 Wochen Zeitersparnis
2. **Einfachere Architektur:** Kein Database Interface, keine Dialekt-Abstraction
3. **Bessere Testbarkeit:** In-Memory SQLite für Unit-Tests
4. **Geringere Komplexität:** Weniger Moving Parts
5. **Portabilität:** Single Binary + DB-File (kein Server-Setup)
6. **Performance-Optimierung:** SQLite-spezifische PRAGMAs nutzbar

### Negative Konsequenzen (Trade-offs)

1. **Breaking Change:** User mit MySQL/PostgreSQL/MSSQL müssen migrieren
2. **Kein Cloud-Native:** Keine Multi-Instanz-Szenarien
3. **Feature-Verlust:** MS Access Support entfällt (aber: war bereits Legacy)

### Risiken & Mitigationen

| Risiko | Wahrscheinlichkeit | Impact | Mitigation |
|--------|-------------------|--------|------------|
| User-Backlash (Multi-DB-Nutzer) | Niedrig | Niedrig | Projekt = Fork, keine Legacy-User-Base |
| Skalierungs-Grenzen | Niedrig | Niedrig | SDE-Daten < 2GB, SQLite-Limit = 281TB |
| Concurrent Writers | Niedrig | Niedrig | SDE = Read-Only nach Import, kein Problem |

---

## Alternativen (erwogen & verworfen)

### Alternative 1: Multi-DB mit ORM (GORM)

**Pro:**

- Feature-Parität mit VB.NET
- Abstraktion "kostenlos" via ORM

**Contra:**

- ❌ Entwicklungszeit: +4 Wochen
- ❌ ORM-Overhead (Performance-Penalty)
- ❌ Komplexität: Jede DB braucht eigene Tests
- ❌ Deployment: Multi-DB-CI-Pipeline nötig

**Entscheidung:** Verworfen (unnötige Komplexität für geringen Nutzen)

### Alternative 2: PostgreSQL-Only (statt SQLite)

**Pro:**

- Skaliert besser für sehr große Datenmengen
- Cloud-Native (RDS, Supabase, etc.)

**Contra:**

- ❌ Server-Setup nötig (nicht embedded)
- ❌ Komplexer für Endnutzer (Installation, Config)
- ❌ Overkill für SDE Use-Case (statische Daten, <2GB)

**Entscheidung:** Verworfen (SQLite ist ausreichend + einfacher)

### Alternative 3: DuckDB (statt SQLite)

**Pro:**

- Optimiert für Analytics (ähnlich SQLite)
- Sehr schnell für aggregierte Queries

**Contra:**

- ❌ Weniger verbreitet als SQLite
- ❌ Go-Driver weniger mature
- ❌ Keine signifikanten Vorteile für SDE-Workload

**Entscheidung:** Verworfen (SQLite ist Standard, besser unterstützt)

---

## Implementierungsdetails

### Database Layer (Technologie-Stack)

```go
// Kein ORM - direktes SQL mit sqlx
import (
    "database/sql"
    "encoding/json"
    _ "github.com/mattn/go-sqlite3"  // CGo-Driver
    "github.com/jmoiron/sqlx"        // Ergänzt sql.DB um Named Queries
)

type SQLiteDB struct {
    db *sqlx.DB
}

// JSONL Import (EVE SDE Format)
func (s *SQLiteDB) ImportJSONL(table string, jsonlFile string) error {
    tx, _ := s.db.Beginx()
    defer tx.Rollback()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        var record map[string]interface{}
        json.Unmarshal(scanner.Bytes(), &record)
        tx.NamedExec("INSERT INTO "+table+" (...) VALUES (:field1, :field2)", record)
    }

    return tx.Commit()
}
```

### Performance-Optimierung (SQLite PRAGMAs)

```sql
-- Bei DB-Erstellung
PRAGMA journal_mode = WAL;          -- Write-Ahead Logging (bessere Concurrency)
PRAGMA synchronous = NORMAL;        -- Balance: Safety vs. Speed
PRAGMA cache_size = -64000;         -- 64MB Cache
PRAGMA temp_store = MEMORY;         -- Temp Tables in RAM
PRAGMA mmap_size = 30000000000;     -- 30GB Memory-Mapped I/O
```

### Schema-Migrationen

```bash
# golang-migrate
migrate -path migrations/sqlite -database "sqlite3://eve_sde.db" up
```

---

## Compliance & Governance

### Normative Anforderungen (aus copilot-instructions.md)

- ✅ **MUST:** Keine Hardcodierung sensibler Parameter → SQLite-Pfad via Config
- ✅ **MUST:** Dependency Scans → `mattn/go-sqlite3` ist stabil, aktiv maintained
- ✅ **SHOULD:** Least Privilege → SQLite-File Permissions via OS
- ✅ **SHOULD:** Rollback-Fähigkeit → Schema-Migrationen mit `down`-Scripts

### ADR-Prozess

- Status: **Proposed** (wartet auf Approval)
- Review-Kriterien: Technische Machbarkeit, User-Impact, Zeitplan
- Supersession: Falls verworfen, neue ADR mit Begründung nötig

---

## Festlegungen

1. **Datenformat:** EVE SDE wird als JSONL (JSON Lines) geliefert
   - Quelle: [developers.eveonline.com/static-data](https://developers.eveonline.com/static-data)
   - Enhanced Version: [sde.riftforeve.online](https://sde.riftforeve.online/)
   - Import-Pipeline: JSONL → Parser → SQLite
   - Format: Eine JSON-Object pro Zeile (`.jsonl`-Files)
   - Kein CSV-Import nötig

2. **SQLite-Version:** Minimum 3.35+
   - Moderne Features: RETURNING, STRICT Tables, Generated Columns
   - Release-Datum: März 2021 (stabil, weitverbreitet)

3. **WAL-Mode:** Default aktiviert
   - Bessere Concurrency (Reads blockieren Writes nicht)
   - Performance-Vorteil bei Multi-Reader-Szenarien
   - Moderne Best Practice

4. **Datenbank-Layout:** Single-File
   - Eine `.db`-Datei für kompletten SDE-Import
   - Einfacher zu handhaben, teilen, backup-en

---

## Referenzen

**EVE SDE Datenquellen:**

- [CCP Static Data Export](https://developers.eveonline.com/static-data) - Offizielle JSONL-Downloads
- [RIFT SDE Schema Docs](https://sde.riftforeve.online/) - Automatisch generierte Schema-Dokumentation
- [EVE Developer Docs](https://developers.eveonline.com/docs/) - API & SDE Überblick

**SQLite & Go:**

- [SQLite Official Docs](https://www.sqlite.org/docs.html)
- [SQLite Performance Tuning](https://www.sqlite.org/pragma.html)
- [mattn/go-sqlite3 GitHub](https://github.com/mattn/go-sqlite3)
- [sqlx Documentation](https://jmoiron.github.io/sqlx/)
- [golang-migrate](https://github.com/golang-migrate/migrate)

**SDE Daten-Spezifikationen:**

- Format: JSONL (JSON Lines)
- Download-Größe: ~500 MB (komprimiert)
- Unkomprimiert: ~1.5 GB
- Tabellen: 50+ (types, blueprints, dogmaAttributes, mapSolarSystems, etc.)
- Aktuelle Version: 3060393 (15. Oktober 2025)

---

## Änderungshistorie

| Datum | Version | Änderung | Autor |
|-------|---------|----------|-------|
| 2025-10-15 | 1.0.0 | Status → Accepted | Migration Team |
| 2025-10-15 | 0.1.0 | Initial Draft | AI Copilot |

---

**Nächste Schritte:**

1. Review durch Team
2. Approval/Reject-Entscheidung
3. Bei Approval: Status → `Accepted`
4. Implementation gemäß Spec
