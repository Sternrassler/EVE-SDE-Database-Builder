# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2025-10-17

### Added

- **Pre-Push Git Hook** (PR #c6f53ba)
  - Automatische Ausführung von `make lint` vor jedem Push
  - Automatische Ausführung von `make test` vor jedem Push
  - Verhindert Push von Code mit Linting-Fehlern oder fehlschlagenden Tests
  - Hook-Datei: `.githooks/pre-push`
- **Code Generation Tools** (PR #134)
  - Tool: `add-tomap-methods` - Fügt ToMap() Methoden zu generierten Structs hinzu
  - Tool: `scrape-rift-schemas` - Scrapt YAML-Schemas von RIFT EVE Schema Browser
  - Separate Package-Struktur: `tools/add-tomap-methods/` und `tools/scrape-rift-schemas/`
  - Tests für beide Tools mit vollständiger Abdeckung
  - Behebung: "main redeclared" und "Config redeclared" Kompilierungsfehler
- **Parser Package Documentation** (PR #133)
  - Package-Level Dokumentation für alle Parser-Komponenten
  - Behebung: SA1012 Lint-Fehler (nil context → context.Background())
  - Verbesserte Godoc-Kommentare für öffentliche APIs
- **Error Recovery Strategies** (PR #132)
  - Dokumentation: Umfassende Error Recovery Patterns
  - Beispiele für Skip-Mode, FailFast-Mode und Error-Threshold-Handling
  - Integration mit JSONL Parser für robuste Fehlerbehandlung
  - Fehler-Statistiken und Reporting (ParseResult mit ErrorSummary)
- **Core Parsers Implementation** (PR #131)
  - Parser für invTypes mit vollständiger Feldabdeckung
  - Parser für invGroups mit Gruppen-Hierarchie
  - Behebung aller errcheck Lint-Warnungen
  - Registrierung aller Parser im zentralen Registry
- **Parser Performance Benchmarks** (PR #130)
  - Benchmarks für 1k, 10k, 50k, 100k und 500k Zeilen
  - Memory-Effizienz Tests für große Dateien
  - Streaming Performance Validation
  - Backpressure Handling Tests
- **Parser Integration Tests (E2E: JSONL → DB)**
  - Integration Test für invTypes (Parse JSONL → Insert → Verify Row Count)
  - Integration Test für invGroups (Parse JSONL → Insert → Verify Row Count)
  - Integration Test für industryBlueprints (Parse JSONL → Insert → Verify Row Count)
  - Helper-Funktionen für Row-Konvertierung (invTypeToRow, invGroupToRow, blueprintToRow)
  - Test-Datenstrukturen mit korrekter JSON-Mapping und Nullable-Feldern
  - Datei: `internal/parser/integration_test.go`
- **Database Layer Implementation (Epic abgeschlossen)**
  - Migration Automation Make Targets (`migrate-up`, `migrate-down`, `migrate-status`, `migrate-clean`, `migrate-reset`)
  - Umfassende README für `internal/database/` mit API-Dokumentation
  - Performance-Dokumentation (10k Rows in ~14ms, 100k in ~134ms, 500k in ~664ms)
  - Beispiele für Batch Insert, Transactions, Query Helpers
  - Best Practices und Troubleshooting Guide
- Migration für `industryBlueprints` Schema (`migrations/sqlite/003_blueprints.sql`)
  - CREATE TABLE für industryBlueprints (blueprintTypeID, maxProductionLimit)
  - CREATE TABLE für industryActivities (blueprintTypeID, activityID, time)
  - CREATE TABLE für industryActivityMaterials (blueprintTypeID, activityID, materialTypeID, quantity)
  - CREATE TABLE für industryActivityProducts (blueprintTypeID, activityID, productTypeID, quantity)
  - Composite PRIMARY KEYs für Activities, Materials und Products
  - 6 Indizes für performante Lookups (blueprintTypeID, activityID, materialTypeID, productTypeID)
  - Idempotente Migration (CREATE IF NOT EXISTS)
- Umfassende Tests für Blueprints Migration
  - Schema-Validierung (alle Tabellen, Spalten, Constraints)
  - Composite PRIMARY KEY Tests
  - Index-Überprüfung (6 Indizes)
  - Datenoperationen (Insert, Query für alle 4 Tabellen)
  - Index-Performance-Tests
  - Idempotenz-Test (wiederholbare Ausführung)
- Migration für `invTypes` Tabelle (`migrations/sqlite/001_inv_types.sql`)
  - CREATE TABLE Statement mit allen Spalten aus RIFT SDE Schema
  - Indizes für typeID (PRIMARY KEY), groupID, marketGroupID
  - Idempotente Migration (CREATE IF NOT EXISTS)
- Umfassende Tests für invTypes Migration
  - Schema-Validierung (alle Spalten, Constraints)
  - Index-Überprüfung
  - Datenoperationen (Insert, Query)
  - Idempotenz-Test (wiederholbare Ausführung)
- Integration Tests für Foundation-Komponenten (Logger, Errors, Retry)
  - Scenario: Retry mit Logging bei jedem Versuch
  - Scenario: Error Context wird geloggt
  - Scenario: Fatal Error → Panic Recovery
  - Scenario: Retryable Error → Retry mit Backoff → Success
  - Scenario: Context Cancellation
  - Scenario: Multiple Error Types mit Retry-Policy
  - Scenario: Logger mit Context Values
  - Scenario: Error Chain Logging
- Make Target `check-hooks`: Governance-Checks für Git Hooks (normative + adr)
- Make Target `security-blockers`: Prüft kritische Security Findings aus Trivy Report
- Make Target `secrets-check`: Professionelles Secret-Scanning mit Gitleaks
- Atomare Targets: `normative-check`, `adr-check`, `adr-ref`, `commit-lint`, `release-check`
- Script `check-secrets.sh`: Gitleaks-Integration mit automatischer Installation
- Git Hook Pfad konfiguriert: `.githooks` als Standard

### Changed

- Makefile vereinfacht: 15 → 9 Targets mit klarer Hierarchie (check-hooks, check-local, check-pr, check-ci)
- Alle Make Targets delegieren zu Scripts in `scripts/` (keine Code-Duplikation mehr)
- Git Hooks nutzen Make Targets statt direkte Script-Aufrufe
- Workflow `pr-quality-gates.yml` nutzt `make check-ci` statt gelöschtes `make pr-quality-gates-ci`
- Secret-Scanning: Professionelles Gitleaks statt einfacher Grep-Heuristik
- Pre-Commit Hook nutzt ausschließlich Make Targets (`check-hooks`, `secrets-check`)

### Removed

- Verwaistes Script `scripts/workflows/pr-quality-gates/run.sh` (Logik in Makefile konsolidiert)
- Redundante Make Targets: `lint-ci`, `scan-json`, `pr-check`, `push-ci`, `ci-local`, `pr-quality-gates-ci`
- Einfache Grep-basierte Secret-Heuristik im Pre-Commit Hook

## [0.1.0] - 2025-10-05

- Project initialization.
