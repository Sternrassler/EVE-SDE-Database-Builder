# ADR-002: Database Layer Design

**Status:** Accepted  
**Datum:** 2025-10-15  
**Entscheider:** Migration Team  
**Kontext:** VB.NET → Go Migration für EVE SDE Database Builder  
**Abhängig von:** ADR-001 (SQLite-Only Approach)

---

## Kontext & Problem

Nach Entscheidung für SQLite-Only (ADR-001) muss die **Database Layer Architektur** definiert werden.

### Anforderungen

**Funktional:**

- Import von JSONL-Daten (50+ Tabellen)
- Batch Inserts (Performance-kritisch: ~500k+ Zeilen pro Import)
- Schema Migrations (SDE-Versionen ändern sich)
- Transaktionale Sicherheit (Rollback bei Fehlern)
- Query-Support für Exporte/Analysen

**Nicht-Funktional:**

- Performance: Minimaler Overhead
- Maintainability: Klare, testbare Abstraktion
- Type Safety: Compile-Zeit-Fehler statt Runtime
- Migration Effort: Geringe Lernkurve

### VB.NET Status Quo

**Aktuelles Design:**

```vb
' SQLiteDB.vb - Direktes ADO.NET ohne Abstraction
Public Class SQLiteDB
    Private _connection As SQLiteConnection
    
    Public Sub BulkInsert(table As String, data As DataTable)
        Using transaction = _connection.BeginTransaction()
            Using cmd = New SQLiteCommand()
                ' ... Prepared Statements ...
            End Using
            transaction.Commit()
        End Using
    End Sub
End Class
```

**Charakteristik:** Thin Wrapper um ADO.NET, keine ORM-Features

---

## Entscheidung

Wir verwenden **`database/sql` + `sqlx`** ohne ORM.

### Architektur

```go
// internal/database/sqlite.go
package database

import (
    "database/sql"
    "github.com/jmoiron/sqlx"
    _ "github.com/mattn/go-sqlite3"
)

type DB struct {
    *sqlx.DB
}

// NewDB initialisiert SQLite-Connection mit Pragmas
func NewDB(path string) (*DB, error) {
    db, err := sqlx.Connect("sqlite3", path+"?_journal_mode=WAL")
    if err != nil {
        return nil, err
    }
    
    // Performance-Optimierungen
    db.MustExec("PRAGMA synchronous = NORMAL")
    db.MustExec("PRAGMA cache_size = -64000")  // 64MB
    db.MustExec("PRAGMA temp_store = MEMORY")
    db.MustExec("PRAGMA mmap_size = 30000000000")  // 30GB
    
    return &DB{db}, nil
}

// BatchInsert - Optimiert für JSONL-Import
func (db *DB) BatchInsert(table string, records []map[string]interface{}) error {
    tx, err := db.Beginx()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    stmt, err := tx.PrepareNamed(buildInsertQuery(table, records[0]))
    if err != nil {
        return err
    }
    
    for _, record := range records {
        if _, err := stmt.Exec(record); err != nil {
            return err
        }
    }
    
    return tx.Commit()
}

// Typed Query Beispiel (für Exporte)
type TypeRow struct {
    TypeID      int     `db:"typeID"`
    TypeName    string  `db:"typeName"`
    Description string  `db:"description"`
    Mass        float64 `db:"mass"`
}

func (db *DB) GetTypes() ([]TypeRow, error) {
    var types []TypeRow
    err := db.Select(&types, "SELECT * FROM invTypes LIMIT 100")
    return types, err
}
```

### Begründung

**1. Warum KEIN ORM (GORM, ent, etc.)?**

| Kriterium | GORM | sqlx | Bewertung |
|-----------|------|------|-----------|
| Bulk Insert Performance | ⚠️ Langsam (Reflections) | ✅ Schnell (Prepared Stmts) | **sqlx gewinnt** |
| Schema Migrations | ✅ Auto-Migrate | ❌ Extern (golang-migrate) | Neutral (wir brauchen eh Migrations) |
| Type Safety | ✅ Struct-basiert | ✅ Struct Tags | Gleich |
| Lernkurve | ⚠️ Steil (Magic Behavior) | ✅ Flach (SQL-nah) | **sqlx gewinnt** |
| Overhead | ❌ Hoch (Reflection) | ✅ Minimal | **sqlx gewinnt** |
| SQLite-spezifische PRAGMAs | ⚠️ Umständlich | ✅ Direkt | **sqlx gewint** |

**2. Warum NICHT bare `database/sql`?**

- `sqlx` ergänzt um Named Queries (`NamedExec`, `NamedQuery`)
- Struct Scanning (`Select`, `Get`) reduziert Boilerplate
- Kompatibel mit `database/sql` (kein Lock-In)
- Minimaler Overhead (~5% vs. bare sql)

**3. SQLite-Only Optimierung**

- Keine Multi-DB-Abstraktion nötig (ADR-001)
- Direkte PRAGMA-Kontrolle für Performance
- CGo-basierter Driver (`mattn/go-sqlite3`) erlaubt Custom Functions

---

## Konsequenzen

### Positive Konsequenzen

1. **Performance:** Optimale Batch-Insert-Geschwindigkeit
2. **Einfachheit:** Weniger Abstraktions-Layer als ORM
3. **SQL-Transparenz:** Queries sind explizit (keine Hidden Queries)
4. **Testing:** In-Memory SQLite für Unit-Tests trivial
5. **Migration Effort:** VB.NET → Go ähnlich (beide SQL-nah)

### Negative Konsequenzen

1. **Boilerplate:** Mehr Code als ORM für CRUD
2. **Schema-Evolution:** Manuelle Migrations (keine Auto-Migrate)
3. **Type Safety:** Weniger Compile-Zeit-Checks als ORMs mit Code-Generation

### Mitigationen

| Konsequenz | Mitigation |
|------------|------------|
| Boilerplate | Code-Generator für Standardfälle (INSERT/SELECT Queries) |
| Schema Migrations | `golang-migrate` mit Versionskontrolle |
| Type Safety | `sqlc` evaluieren (Query → Typsichere Go-Funktionen) |

---

## Alternativen (erwogen & verworfen)

### Alternative 1: GORM (Full ORM)

**Pro:**

- Auto-Migrations
- Associations (Foreign Keys) automatisch
- Weit verbreitet

**Contra:**

- ❌ Performance-Overhead (Reflection) für Bulk-Inserts kritisch
- ❌ "Magic Behavior" (implizite Queries)
- ❌ Overkill für Read-Heavy Workload (SDE = statisch)
- ❌ SQLite-spezifische Optimierungen schwieriger

**Entscheidung:** Verworfen (Performance-Anforderungen)

### Alternative 2: ent (Facebook's ORM)

**Pro:**

- Code-First Schema
- Type-Safe Queries
- Graph-basierte Queries

**Contra:**

- ❌ Noch steiler Learning Curve als GORM
- ❌ Schema-Generator nicht SDE-kompatibel (External Schema)
- ❌ Performance ähnlich wie GORM

**Entscheidung:** Verworfen (Komplexität)

### Alternative 3: sqlc (SQL → Go Code Generator)

**Pro:**

- ✅ Type-Safe Queries zur Compile-Zeit
- ✅ Null Overhead (generiert bare SQL-Code)
- ✅ SQL bleibt explizit

**Contra:**

- ⚠️ Code-Generation-Step nötig (Build-Komplexität)
- ⚠️ 50+ Tabellen = viel generierter Code

**Entscheidung:** **Evaluieren für v2.0** (gute Ergänzung zu sqlx, aber nicht v1 kritisch)

### Alternative 4: Bare `database/sql`

**Pro:**

- ✅ Null Dependencies (nur stdlib)
- ✅ Maximale Kontrolle

**Contra:**

- ❌ Zu viel Boilerplate für Struct Mapping
- ❌ Named Queries fehlen

**Entscheidung:** Verworfen (sqlx ist minimal und bewährt)

---

## Implementierungsdetails

### 1. Schema Migrations

**Tool:** `golang-migrate/migrate`

```bash
# Migrations erstellen
migrate create -ext sql -dir migrations/sqlite -seq create_inv_types

# Anwenden
migrate -path migrations/sqlite -database "sqlite3://eve_sde.db" up

# Rollback
migrate -path migrations/sqlite -database "sqlite3://eve_sde.db" down 1
```

**Beispiel Migration:**

```sql
-- migrations/sqlite/000001_create_inv_types.up.sql
CREATE TABLE IF NOT EXISTS invTypes (
    typeID INTEGER PRIMARY KEY,
    groupID INTEGER NOT NULL,
    typeName TEXT NOT NULL,
    description TEXT,
    mass REAL,
    volume REAL,
    capacity REAL,
    published INTEGER DEFAULT 0,
    FOREIGN KEY (groupID) REFERENCES invGroups(groupID)
) STRICT;

CREATE INDEX idx_invTypes_groupID ON invTypes(groupID);
CREATE INDEX idx_invTypes_published ON invTypes(published);

-- migrations/sqlite/000001_create_inv_types.down.sql
DROP TABLE IF EXISTS invTypes;
```

### 2. Performance-Optimierung: Batch Inserts

**Problem:** Einzelne Inserts = langsam

```go
// ❌ Langsam: ~5 Sekunden für 10k Rows
for _, record := range records {
    db.Exec("INSERT INTO types (...) VALUES (...)", record)
}

// ✅ Schnell: ~0.5 Sekunden für 10k Rows
tx, _ := db.Beginx()
stmt, _ := tx.PrepareNamed("INSERT INTO types (...) VALUES (:field1, :field2)")
for _, record := range records {
    stmt.Exec(record)
}
tx.Commit()
```

**Best Practice:** Batches von 1000-5000 Rows pro Transaktion

### 3. Testing-Strategie

```go
// internal/database/sqlite_test.go
func TestBatchInsert(t *testing.T) {
    // In-Memory DB für Tests
    db, _ := NewDB(":memory:")
    defer db.Close()
    
    // Schema laden
    db.MustExec(testSchema)
    
    // Test Data
    records := []map[string]interface{}{
        {"typeID": 34, "typeName": "Tritanium"},
        {"typeID": 35, "typeName": "Pyerite"},
    }
    
    err := db.BatchInsert("invTypes", records)
    assert.NoError(t, err)
    
    // Verify
    var count int
    db.Get(&count, "SELECT COUNT(*) FROM invTypes")
    assert.Equal(t, 2, count)
}
```

### 4. Connection Pool Konfiguration

```go
func NewDB(path string) (*DB, error) {
    db, err := sqlx.Connect("sqlite3", path+"?_journal_mode=WAL")
    if err != nil {
        return nil, err
    }
    
    // SQLite = Embedded, nur 1 Writer
    db.SetMaxOpenConns(1)
    db.SetMaxIdleConns(1)
    db.SetConnMaxLifetime(0)
    
    return &DB{db}, nil
}
```

---

## Migration von VB.NET

### Code-Vergleich

**VB.NET (alt):**

```vb
' SQLiteDB.vb
Public Sub BulkInsert(table As String, data As DataTable)
    Using transaction = _connection.BeginTransaction()
        Using cmd = New SQLiteCommand()
            cmd.Connection = _connection
            cmd.Transaction = transaction
            
            For Each row As DataRow In data.Rows
                cmd.CommandText = $"INSERT INTO {table} ..."
                cmd.ExecuteNonQuery()
            Next
        End Using
        transaction.Commit()
    End Using
End Sub
```

**Go (neu):**

```go
// internal/database/sqlite.go
func (db *DB) BatchInsert(table string, records []map[string]interface{}) error {
    tx, err := db.Beginx()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    stmt, err := tx.PrepareNamed(buildInsertQuery(table, records[0]))
    if err != nil {
        return err
    }
    
    for _, record := range records {
        if _, err := stmt.Exec(record); err != nil {
            return err
        }
    }
    
    return tx.Commit()
}
```

**Unterschiede:**

- ✅ Go: Defer für Cleanup (kein `Using`)
- ✅ Go: Error Handling explizit
- ✅ Go: Named Parameters via sqlx

---

## Compliance & Governance

### Normative Anforderungen (aus copilot-instructions.md)

- ✅ **MUST:** Tests grün → Unit-Tests für alle DB-Operationen
- ✅ **MUST:** Keine Hardcoded Secrets → DB-Pfad via Config
- ✅ **SHOULD:** Performance-Optimierung → Batching + Pragmas
- ✅ **SHOULD:** Rollback-Fähigkeit → Transaktionen + Migrations

### ADR-Abhängigkeiten

- **ADR-001:** SQLite-Only → Entscheidung basiert darauf
- **ADR-004:** Config Format → DB-Pfad aus Config lesen

---

## Referenzen

**Libraries:**

- [sqlx GitHub](https://github.com/jmoiron/sqlx)
- [database/sql Docs](https://pkg.go.dev/database/sql)
- [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)
- [golang-migrate](https://github.com/golang-migrate/migrate)

**Alternativen (für spätere Evaluation):**

- [sqlc](https://sqlc.dev/) - SQL → Type-Safe Go
- [GORM](https://gorm.io/)
- [ent](https://entgo.io/)

**Best Practices:**

- [SQLite Performance Tuning](https://www.sqlite.org/pragma.html)
- [Go database/sql Tutorial](https://go.dev/doc/database/index)

---

## Änderungshistorie

| Datum | Version | Änderung | Autor |
|-------|---------|----------|-------|
| 2025-10-15 | 1.0.0 | Status → Accepted | Migration Team |
| 2025-10-15 | 0.1.0 | Initial Draft | AI Copilot |
