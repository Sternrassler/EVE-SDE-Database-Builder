# Database Migrations

## Übersicht

Dieses Verzeichnis enthält SQL-Migrationsskripte für die EVE SDE SQLite-Datenbank.

## Migrationen

| Nr. | Datei | Beschreibung | Status |
|-----|-------|--------------|--------|
| 001 | `001_inv_types.sql` | invTypes Tabelle (häufigste SDE-Tabelle) | ✅ Implementiert |
| 002 | `002_inv_groups.sql` | invGroups Tabelle (Item Groups Kategorisierung) | ✅ Implementiert |
| 003 | `003_blueprints.sql` | Industry Blueprints (Blueprints, Activities, Materials, Products) | ✅ Implementiert |
| 004 | `004_dogma.sql` | Dogma System (Attributes, Effects, Type Attributes/Effects) | ✅ Implementiert |

## Migration-Format

Jede Migration folgt diesem Format:

```sql
-- Migration: <nummer>_<name>.sql
-- Description: <Kurze Beschreibung>
-- Source: RIFT SDE Schema (https://sde.riftforeve.online/)
-- ADR Reference: <relevante ADRs>

CREATE TABLE IF NOT EXISTS <table_name> (
    -- Spalten mit Constraints
);

CREATE INDEX IF NOT EXISTS <index_name> ON <table_name>(<column>);
```

## Ausführung

### Manuell mit SQLite CLI

```bash
sqlite3 eve_sde.db < migrations/sqlite/001_inv_types.sql
```

### Programmatisch (Go)

```go
db, _ := database.NewDB("eve_sde.db")
migrationSQL, _ := os.ReadFile("migrations/sqlite/001_inv_types.sql")
db.Exec(string(migrationSQL))
```

### Idempotenz

Alle Migrationen sind idempotent (können mehrfach ausgeführt werden):

- `CREATE TABLE IF NOT EXISTS`
- `CREATE INDEX IF NOT EXISTS`

## Tests

Migration-Tests befinden sich in `internal/database/migration_test.go`:

```bash
# Alle Migration-Tests ausführen
go test -v ./internal/database -run TestMigration

# Spezifische Migration testen
go test -v ./internal/database -run TestMigration_001
```

## Schema-Quelle

- **Primär:** [RIFT SDE Schema](https://sde.riftforeve.online/)
- **Fallback:** [CCP Official SDE](https://developers.eveonline.com/static-data)

## ADR Referenzen

- [ADR-001: SQLite-Only Approach](../../docs/adr/ADR-001-sqlite-only-approach.md)
- [ADR-002: Database Layer Design](../../docs/adr/ADR-002-database-layer-design.md)

## Naming Convention

Format: `<nummer>_<table_name>.sql`

- Nummer: 3-stellig, aufsteigend (001, 002, ...)
- Table Name: Lowercase, Unterstriche für Wörter
- Beispiel: `001_inv_types.sql`

## Best Practices

1. **Idempotenz:** Immer `IF NOT EXISTS` verwenden
2. **Dokumentation:** Jede Migration mit Header-Kommentar
3. **Tests:** Für jede Migration entsprechende Tests schreiben
4. **ADR:** Bei Architektur-relevanten Änderungen ADR referenzieren
5. **Reihenfolge:** Migrationen sequentiell nummerieren
6. **Foreign Keys:** Optional, nach ADR-001 (SQLite FK enforcement ist enabled)
