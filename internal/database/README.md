# Database Layer

SQLite-basierte Datenbankschicht für den EVE SDE Database Builder gemäß ADR-001 (SQLite-Only Approach) und ADR-002 (Database Layer Design).

## Features

- ✅ **Optimierte SQLite-Verbindung** mit WAL-Modus und Performance-PRAGMAs
- ✅ **Batch Insert** für große Datenmengen (10k+ Rows in <2s)
- ✅ **Transaction Wrapper** mit automatischem Commit/Rollback
- ✅ **Query Helpers** mit Generics für typsichere Abfragen
- ✅ **Testing Utilities** für In-Memory-Tests mit automatischen Migrationen
- ✅ **Connection Pool** optimiert für SQLite (single writer)

## Architecture Decision Records

- [ADR-001: SQLite-Only Approach](../../docs/adr/ADR-001-sqlite-only-approach.md)
- [ADR-002: Database Layer Design](../../docs/adr/ADR-002-database-layer-design.md)

## Performance

**Benchmark-Ergebnisse (AMD EPYC 7763, In-Memory SQLite):**

| Rows   | Time (ms) | Rows/sec  | Memory (MB) |
|--------|-----------|-----------|-------------|
| 10k    | ~14       | ~714k     | ~6          |
| 50k    | ~67       | ~746k     | ~31         |
| 100k   | ~134      | ~746k     | ~62         |
| 500k   | ~664      | ~753k     | ~309        |

**Batch Size Optimization:**

| Batch Size | Time (ms) | Performance |
|------------|-----------|-------------|
| 100        | ~68       | Baseline    |
| 500        | ~67       | +1.5%       |
| 1000       | ~67       | +1.5%       |
| 5000       | ~67       | +1.5%       |

**Empfehlung:** Batch Size 1000 (guter Trade-off zwischen Performance und Speichernutzung)

## API Documentation

### Connection Management

#### NewDB

```go
func NewDB(path string) (*sqlx.DB, error)
```

Erstellt eine neue SQLite-Datenbankverbindung mit optimierten PRAGMAs:
- `journal_mode = WAL` (Write-Ahead Logging)
- `synchronous = NORMAL` (Balance zwischen Sicherheit und Performance)
- `foreign_keys = ON` (Referentielle Integrität)
- `cache_size = -64000` (64MB Cache)
- `temp_store = MEMORY` (Temporäre Tabellen im RAM)
- `busy_timeout = 5000` (5 Sekunden Wartezeit bei Lock)

**Parameter:**
- `path`: Dateipfad zur SQLite-Datenbank. `:memory:` für In-Memory-Datenbanken.

**Beispiel:**

```go
db, err := database.NewDB("eve_sde.db")
if err != nil {
    log.Fatal(err)
}
defer database.Close(db)
```

#### Close

```go
func Close(db *sqlx.DB) error
```

Schließt die Datenbankverbindung sicher.

### Batch Insert

#### BatchInsert

```go
func BatchInsert(ctx context.Context, db *sqlx.DB, table string, 
    columns []string, rows [][]interface{}, batchSize int) error
```

Führt optimierte Batch-Inserts durch. Teilt große Datenmengen automatisch in Batches auf und wickelt alle Operationen in einer Transaktion ab.

**Parameter:**
- `ctx`: Context für Timeout und Abbruch
- `db`: Datenbankverbindung
- `table`: Ziel-Tabellenname
- `columns`: Spaltennamen für Insert
- `rows`: Datenzeilen (jede Zeile muss gleiche Anzahl Werte wie Spalten haben)
- `batchSize`: Anzahl Zeilen pro INSERT-Statement (empfohlen: 1000)

**Beispiel:**

```go
columns := []string{"typeID", "typeName", "groupID"}
rows := [][]interface{}{
    {34, "Tritanium", 18},
    {35, "Pyerite", 18},
    {36, "Mexallon", 18},
}

ctx := context.Background()
err := database.BatchInsert(ctx, db, "invTypes", columns, rows, 1000)
if err != nil {
    log.Fatal(err)
}
```

#### BatchInsertWithProgress

```go
func BatchInsertWithProgress(ctx context.Context, db *sqlx.DB, table string,
    columns []string, rows [][]interface{}, batchSize int, 
    progressCallback ProgressCallback) error
```

Wie `BatchInsert`, aber mit optionalem Progress-Callback.

**Beispiel:**

```go
progressCallback := func(current, total int) {
    fmt.Printf("Imported %d/%d rows (%.1f%%)\n", 
        current, total, float64(current)/float64(total)*100)
}

err := database.BatchInsertWithProgress(ctx, db, "invTypes", 
    columns, rows, 1000, progressCallback)
```

### Transaction Wrapper

#### WithTransaction

```go
func WithTransaction(ctx context.Context, db *sqlx.DB, 
    fn func(*sqlx.Tx) error, opts ...TxOption) error
```

Führt eine Funktion innerhalb einer Transaktion aus. Automatisches Commit bei Erfolg, Rollback bei Fehler oder Panic.

**Parameter:**
- `ctx`: Context für Timeout und Abbruch
- `db`: Datenbankverbindung
- `fn`: Funktion, die innerhalb der Transaktion ausgeführt wird
- `opts`: Optionale Transaction-Konfiguration (Isolation Level, Read-Only)

**Beispiel:**

```go
err := database.WithTransaction(ctx, db, func(tx *sqlx.Tx) error {
    _, err := tx.Exec("INSERT INTO users (id, name) VALUES (?, ?)", 1, "Alice")
    if err != nil {
        return err // Rollback
    }
    
    _, err = tx.Exec("INSERT INTO roles (user_id, role) VALUES (?, ?)", 1, "admin")
    return err // Commit wenn kein Fehler
})
```

**Mit Optionen:**

```go
// Read-Only Transaction mit Serializable Isolation
err := database.WithTransaction(ctx, db, func(tx *sqlx.Tx) error {
    var count int
    return tx.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
}, database.WithReadOnly(), database.WithIsolationLevel(sql.LevelSerializable))
```

### Query Helpers

#### QueryRow

```go
func QueryRow[T any](ctx context.Context, db sqlx.QueryerContext, 
    query string, args ...interface{}) (T, error)
```

Führt eine Query aus, die genau eine Zeile zurückgibt, und scannt sie in den Typ T.

**Beispiel:**

```go
type User struct {
    ID   int    `db:"id"`
    Name string `db:"name"`
}

user, err := database.QueryRow[User](ctx, db, 
    "SELECT id, name FROM users WHERE id = ?", 1)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("User: %s\n", user.Name)
```

#### QueryAll

```go
func QueryAll[T any](ctx context.Context, db sqlx.QueryerContext, 
    query string, args ...interface{}) ([]T, error)
```

Führt eine Query aus, die mehrere Zeilen zurückgibt, und scannt sie in einen Slice.

**Beispiel:**

```go
users, err := database.QueryAll[User](ctx, db, 
    "SELECT id, name FROM users WHERE active = ?", true)
if err != nil {
    log.Fatal(err)
}

for _, user := range users {
    fmt.Printf("User: %s\n", user.Name)
}
```

#### Exists

```go
func Exists(ctx context.Context, db sqlx.QueryerContext, 
    query string, args ...interface{}) (bool, error)
```

Prüft, ob eine Query mindestens eine Zeile zurückgibt.

**Beispiel:**

```go
exists, err := database.Exists(ctx, db, 
    "SELECT 1 FROM users WHERE email = ?", "user@example.com")
if err != nil {
    log.Fatal(err)
}
if exists {
    fmt.Println("User mit dieser E-Mail existiert bereits")
}
```

### Testing Utilities

#### NewTestDB

```go
func NewTestDB(t *testing.T) *sqlx.DB
```

Erstellt eine In-Memory-Datenbank mit allen Migrationen für Tests. Automatisches Cleanup via `t.Cleanup()`.

**Beispiel:**

```go
func TestMyFeature(t *testing.T) {
    db := database.NewTestDB(t)
    // Datenbank ist automatisch aufgeräumt nach Test
    
    // Test-Daten einfügen
    _, err := db.Exec("INSERT INTO invTypes (typeID, typeName, groupID) VALUES (1, 'Test', 1)")
    if err != nil {
        t.Fatal(err)
    }
    
    // Tests durchführen...
}
```

#### ApplyMigrations

```go
func ApplyMigrations(db *sqlx.DB) error
```

Wendet alle Migrationen aus `migrations/sqlite/` auf die Datenbank an.

**Beispiel:**

```go
db, _ := database.NewDB(":memory:")
defer database.Close(db)

if err := database.ApplyMigrations(db); err != nil {
    log.Fatal(err)
}
```

## Migrations

### Migration Files

Alle Schema-Migrationen befinden sich in `migrations/sqlite/`:

| File | Description |
|------|-------------|
| `001_inv_types.sql` | invTypes (Haupt-Item-Tabelle) |
| `002_inv_groups.sql` | invGroups (Item-Kategorisierung) |
| `003_blueprints.sql` | Industry Blueprints (4 Tabellen) |
| `004_dogma.sql` | Dogma System (Attributes, Effects) |
| `005_universe.sql` | Universe Schema (Regions, Systems, etc.) |

### Make Targets

```bash
# Status anzeigen
make migrate-status

# Alle Migrationen anwenden
make migrate-up

# Alle Tabellen löschen (destructive)
make migrate-down

# Datenbankdatei löschen (destructive)
make migrate-clean

# Reset (clean + migrate-up)
make migrate-reset
```

### Custom Database File

```bash
# Andere Datenbankdatei verwenden
DB_FILE=custom.db make migrate-up
```

## Testing

### Unit Tests

```bash
# Alle Tests ausführen
go test -v ./internal/database/

# Spezifische Tests
go test -v ./internal/database/ -run TestBatchInsert
go test -v ./internal/database/ -run TestMigration
```

### Benchmarks

```bash
# Alle Benchmarks
go test -bench=. -benchmem ./internal/database/

# Spezifische Benchmarks
go test -bench=BenchmarkBatchInsert_10k -benchmem ./internal/database/
go test -bench=BenchmarkBatchInsert_100k -benchmem ./internal/database/
```

## Error Handling

Alle Fehler werden mit aussagekräftigen Fehlermeldungen zurückgegeben:

```go
// Validation-Fehler
if table == "" {
    return fmt.Errorf("table name cannot be empty")
}

// Datenbankfehler mit Kontext
if err := db.Ping(); err != nil {
    return fmt.Errorf("failed to ping database: %w", err)
}

// Transaction-Fehler mit Row-Position
if _, err := tx.ExecContext(ctx, sql, args...); err != nil {
    return fmt.Errorf("failed to insert batch at row %d: %w", i, err)
}
```

## Best Practices

### Connection Management

```go
// ✅ DO: Immer defer Close verwenden
db, err := database.NewDB("eve_sde.db")
if err != nil {
    return err
}
defer database.Close(db)

// ❌ DON'T: Vergessen zu schließen
db, _ := database.NewDB("eve_sde.db")
// ... Nutzung ohne Close
```

### Batch Insert

```go
// ✅ DO: Context mit Timeout verwenden
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

err := database.BatchInsert(ctx, db, table, columns, rows, 1000)

// ❌ DON'T: context.Background() für lange Operationen
err := database.BatchInsert(context.Background(), db, table, columns, rows, 1000)
```

### Transactions

```go
// ✅ DO: WithTransaction verwenden (automatisches Rollback)
err := database.WithTransaction(ctx, db, func(tx *sqlx.Tx) error {
    // Operationen
    return nil
})

// ❌ DON'T: Manuelle Transaction-Verwaltung
tx, _ := db.Begin()
// ... leicht vergessenes Rollback
tx.Commit()
```

### Testing

```go
// ✅ DO: NewTestDB verwenden (automatische Migrationen + Cleanup)
func TestFeature(t *testing.T) {
    db := database.NewTestDB(t)
    // Tests
}

// ❌ DON'T: Manuelle Setup/Teardown
func TestFeature(t *testing.T) {
    db, _ := database.NewDB(":memory:")
    defer database.Close(db)
    // Manuelle Migrationen...
}
```

## Implementation Notes

### SQLite Optimizations

Die folgenden PRAGMAs werden automatisch bei `NewDB()` gesetzt:

```sql
PRAGMA journal_mode = WAL;       -- Bessere Concurrency
PRAGMA synchronous = NORMAL;     -- Balance: Sicherheit/Performance
PRAGMA foreign_keys = ON;        -- Referentielle Integrität
PRAGMA cache_size = -64000;      -- 64MB Cache
PRAGMA temp_store = MEMORY;      -- Temp Tables im RAM
PRAGMA busy_timeout = 5000;      -- 5s Wartezeit bei Lock
```

### Connection Pool

```go
db.SetMaxOpenConns(1)  // SQLite: Nur 1 Writer
db.SetMaxIdleConns(1)  // Minimale Idle Connections
```

### Batch Insert Strategy

1. **Validation:** Prüft Eingabedaten vor Transaktion
2. **Transaction:** Eine Transaktion für alle Batches
3. **Batching:** Teilt Daten in Batches (default: 1000 Rows)
4. **Multi-Row INSERT:** `INSERT INTO table VALUES (?, ?), (?, ?), ...`
5. **Context Cancellation:** Prüft Context nach jedem Batch
6. **Rollback on Error:** Automatisches Rollback bei jedem Fehler

## Troubleshooting

### Database Locked

**Problem:** `database is locked`

**Lösung:**
- WAL-Modus ist aktiviert (sollte nicht passieren)
- Prüfen ob `busy_timeout` gesetzt ist (5000ms default)
- Prüfen ob alte Connections nicht geschlossen wurden

### Performance Issues

**Problem:** Langsame Batch-Inserts

**Diagnostik:**
```go
go test -bench=BenchmarkBatchInsert_10k -benchmem ./internal/database/
```

**Mögliche Ursachen:**
- Batch Size zu klein (< 100)
- Batch Size zu groß (> 5000)
- Keine Transaktion verwendet
- Foreign Key Constraints werden geprüft

### Migration Errors

**Problem:** Migration schlägt fehl

**Diagnostik:**
```bash
make migrate-status
sqlite3 eve_sde.db ".schema"
```

**Lösung:**
- Prüfen ob alle Migrationen idempotent sind (`IF NOT EXISTS`)
- Prüfen ob Foreign Keys korrekt referenziert werden

## License

Siehe [LICENSE](../../LICENSE)

## References

- [ADR-001: SQLite-Only Approach](../../docs/adr/ADR-001-sqlite-only-approach.md)
- [ADR-002: Database Layer Design](../../docs/adr/ADR-002-database-layer-design.md)
- [SQLite Documentation](https://www.sqlite.org/docs.html)
- [sqlx Documentation](https://jmoiron.github.io/sqlx/)
