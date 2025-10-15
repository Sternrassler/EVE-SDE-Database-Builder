# EVE SDE Database Builder (Go Edition)

**Status:** 🚧 In Development (Architecture Phase Complete)

Modernes CLI-Tool für den Import von EVE Online Static Data Export (SDE) JSONL-Dateien in eine SQLite-Datenbank. Komplettes Refactoring der VB.NET Legacy-Version mit Fokus auf Performance, Wartbarkeit und Testbarkeit.

> **Legacy VB.NET Version:** Siehe Branch `legacy-vbnet` (wenn vorhanden)

---

## Features

- ✅ **SQLite-Only:** Einfach, portabel, keine externen Datenbankserver
- ✅ **JSONL Support:** Native Unterstützung für CCP's JSONL-Format
- ✅ **Parallel Processing:** Worker Pool für schnellen Import (~3.5min statt ~8min)
- ✅ **Type-Safe:** Full Code Generation für alle 50+ SDE-Tabellen
- ✅ **Structured Logging:** zerolog (JSON/Text)
- ✅ **Resilient:** Retry-Pattern für transiente Fehler
- ✅ **Configuration:** TOML + Environment Variables + CLI Flags

---

## Architecture

Vollständige Architektur-Dokumentation: [docs/migration/architecture-analysis.md](docs/migration/architecture-analysis.md)

### ADRs (Architecture Decision Records)

- [ADR-001: SQLite-Only Approach](docs/adr/ADR-001-sqlite-only-approach.md)
- [ADR-002: Database Layer Design](docs/adr/ADR-002-database-layer-design.md)
- [ADR-003: JSONL Parser Architecture](docs/adr/ADR-003-jsonl-parser-architecture.md)
- [ADR-004: Configuration Format](docs/adr/ADR-004-configuration-format.md)
- [ADR-005: Error Handling Strategy](docs/adr/ADR-005-error-handling-strategy.md)
- [ADR-006: Concurrency & Worker Pool](docs/adr/ADR-006-concurrency-worker-pool.md)

### Tech Stack

| Komponente | Library |
|-----------|---------|
| CLI Framework | `github.com/spf13/cobra` |
| Database | `github.com/mattn/go-sqlite3` + `sqlx` |
| Config | `github.com/BurntSushi/toml` |
| Logging | `github.com/rs/zerolog` |
| JSON Parsing | `encoding/json` (stdlib) |

---

## Development

### Prerequisites

- Go 1.21+
- SQLite 3.35+
- Make

### Setup

```bash
# Install Dependencies
go mod download

# Run Tests (when implemented)
make test

# Lint
make lint

# Security Scan
make scan
```

### Project Structure

```
.
├── cmd/esdedb/              # CLI Entry Point
├── internal/
│   ├── config/              # Configuration Management
│   ├── database/            # SQLite Driver
│   ├── parser/              # JSONL Parser
│   │   └── generated/       # Generated Structs (50+ files)
│   ├── errors/              # Custom Error Types
│   ├── logger/              # Logging
│   ├── retry/               # Retry Pattern
│   └── worker/              # Worker Pool
├── migrations/sqlite/       # Schema Migrations
├── tools/                   # Code Generation Tools
└── docs/                    # Documentation & ADRs
```

---

## Contributing

Siehe [.github/copilot-instructions.md](.github/copilot-instructions.md) für Engineering-Richtlinien.

**Workflow:**

1. Issue erstellen
2. Branch: `feat/<slug>` oder `fix/<slug>`
3. Tests schreiben (TDD)
4. PR mit Issue-Referenz
5. CI Gates (Lint + Tests + Security)

---

## License

Siehe [LICENSE](LICENSE)

---

## Credits

- **EVE Online SDE:** [CCP Games](https://developers.eveonline.com/)
- **RIFT SDE Schema:** [sde.riftforeve.online](https://sde.riftforeve.online/)

---

**Status:** 🚧 v0.1.0-dev (Architecture Complete, Implementation in Progress)
