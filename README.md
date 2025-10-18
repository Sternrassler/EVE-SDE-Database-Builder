# EVE SDE Database Builder (Go Edition)

[![Test](https://github.com/Sternrassler/EVE-SDE-Database-Builder/workflows/Test/badge.svg)](https://github.com/Sternrassler/EVE-SDE-Database-Builder/actions/workflows/test.yml)
[![Lint](https://github.com/Sternrassler/EVE-SDE-Database-Builder/workflows/Lint/badge.svg)](https://github.com/Sternrassler/EVE-SDE-Database-Builder/actions/workflows/lint.yml)
[![Coverage](https://github.com/Sternrassler/EVE-SDE-Database-Builder/workflows/Coverage/badge.svg)](https://github.com/Sternrassler/EVE-SDE-Database-Builder/actions/workflows/coverage.yml)
[![Release](https://img.shields.io/github/v/release/Sternrassler/EVE-SDE-Database-Builder)](https://github.com/Sternrassler/EVE-SDE-Database-Builder/releases/latest)
[![License](https://img.shields.io/github/license/Sternrassler/EVE-SDE-Database-Builder)](LICENSE)

**Status:** 🚀 Core Implementation Complete (51/51 Parsers Ready)

Modernes CLI-Tool für den Import von EVE Online Static Data Export (SDE) JSONL-Dateien in eine SQLite-Datenbank. Komplettes Refactoring der VB.NET Legacy-Version mit Fokus auf Performance, Wartbarkeit und Testbarkeit.

> **Legacy VB.NET Version:** Siehe Branch `legacy-vbnet` (wenn vorhanden)

---

## Features

- ✅ **SQLite-Only:** Einfach, portabel, keine externen Datenbankserver
- ✅ **JSONL Support:** Native Unterstützung für CCP's JSONL-Format
- ✅ **51 Parser Implementations:** Vollständige Abdeckung aller EVE SDE Tabellen
- ✅ **Type-Safe:** Generic parser mit compile-time type checks
- ✅ **Structured Logging:** zerolog (JSON/Text)
- ✅ **Resilient:** Retry-Pattern für transiente Fehler
- ✅ **Configuration:** TOML + Environment Variables + CLI Flags
- ✅ **Shell Completion:** Unterstützung für bash, zsh und fish
- 🚧 **Parallel Processing:** Worker Pool (in Entwicklung - Epic #5)

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
- Node.js + npm (für Code-Generierung)

### Setup

```bash
# Komplettes Setup (Dependencies + Code-Generierung)
make setup

# Oder manuell Schritt für Schritt:

# 1. Install Go Dependencies
go mod download

# 2. Install Code Generation Tools
npm install -g quicktype

# 3. Generate Parser Code from Schemas
make generate-parsers

# Run Tests
make test

# Lint
make lint

# Security Scan
make scan

# Run Benchmarks
make bench

# Performance Regression Testing
make bench-baseline  # Capture baseline
make bench-compare   # Compare against baseline
```

**Nach `git clone`:** Führe `make setup` aus, um alle Dependencies zu installieren und Parser-Code zu generieren.

### Database Migrations

```bash
# Show migration status
make migrate-status

# Apply all migrations (creates eve_sde.db)
make migrate-up

# Drop all tables (destructive - requires confirmation)
make migrate-down

# Delete database file (destructive - requires confirmation)
make migrate-clean

# Reset database (clean + migrate-up)
make migrate-reset
```

**Database File:** `eve_sde.db` (default, can be overridden with `DB_FILE=custom.db make migrate-up`)

**Performance:**
- 10k rows: ~14ms
- 100k rows: ~134ms
- 500k rows: ~664ms

### Shell Completion

Das CLI unterstützt Shell-Completion für bash, zsh und fish:

```bash
# Bash - Für aktuelle Shell-Session
source <(esdedb completion bash)

# Bash - Permanent (Linux)
esdedb completion bash > /etc/bash_completion.d/esdedb

# Bash - Permanent (macOS)
esdedb completion bash > $(brew --prefix)/etc/bash_completion.d/esdedb

# Zsh - Für aktuelle Shell-Session
source <(esdedb completion zsh)

# Zsh - Permanent (Linux)
esdedb completion zsh > "${fpath[1]}/_esdedb"

# Zsh - Permanent (macOS)
esdedb completion zsh > $(brew --prefix)/share/zsh/site-functions/_esdedb

# Fish - Für aktuelle Shell-Session
esdedb completion fish | source

# Fish - Permanent
esdedb completion fish > ~/.config/fish/completions/esdedb.fish
```

**Hinweis:** Nach der Installation der Completion-Scripte muss eine neue Shell gestartet werden.

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

## CI/CD Pipeline

Vollständig automatisierte CI/CD-Pipeline mit GitHub Actions:

- **🔍 Lint:** Automatische Code-Qualitätsprüfung mit golangci-lint
- **✅ Test:** Alle Tests mit Race Detector auf PRs und Main
- **📊 Coverage:** Automatische Coverage-Reports und Trend-Tracking
- **🏁 Benchmark:** Performance-Regression Tests gegen Baseline
- **🚀 Release:** Multi-Platform Builds (Linux, macOS, Windows) bei Git Tags

**Workflows:**
- [`pr-check.yml`](.github/workflows/pr-check.yml) - PR Quality Gates (Lint + Test + Coverage)
- [`benchmark.yml`](.github/workflows/benchmark.yml) - Performance Regression Detection
- [`coverage.yml`](.github/workflows/coverage.yml) - Detaillierte Coverage-Reports
- [`release.yml`](.github/workflows/release.yml) - Automatisierte Release-Erstellung

**Dokumentation:** Siehe [docs/ci-cd/README.md](docs/ci-cd/README.md)

**Lokale Validierung:**
```bash
# Vor dem Push ausführen:
make lint          # Code-Qualität
make test          # Unit Tests
make coverage      # Coverage Report
make build         # Kompilierung

# Release-Bereitschaft prüfen:
make release-check
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

## Recent Milestones

- ✅ **Epic #3 Complete** - Parser Core Infrastructure (Generic JSONL parser, validation, streaming)
- ✅ **Epic #4 Complete** - Full Parser Migration (51/51 EVE SDE tables implemented)
- 🚧 **Epic #5 Next** - Worker Pool & Parallel Processing

**Status:** 🚀 v0.2.0 (Core Implementation Complete, Worker Pool in Development)
