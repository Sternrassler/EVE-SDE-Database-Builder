# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Epic #5 Complete: CLI Interface Implementation** (PR #148-159)
  - **Phase 1: Core Commands** (Tasks #42, #43)
    - `esdedb import`: Vollständiger Import-Befehl mit Worker-Pool-Integration
      - File Discovery: Automatische JSONL-Datei-Erkennung im SDE-Verzeichnis
      - Progress Bar: Live-Metriken (Fortschritt, Geschwindigkeit, ETA, Worker-Status)
      - Error Handling: `--skip-errors` Flag für fehlertoleranten Import
      - Worker-Pool: Konfigurierbare Parallelität via `--workers` Flag
    - `esdedb validate`: Config-Validierung mit detailliertem Feedback
      - TOML-Syntax-Prüfung
      - Semantische Validierung (Pfade, Worker-Counts, Log-Levels)
      - Env-Var-Override-Unterstützung
  - **Phase 2: User Experience** (Tasks #44, #45, #46)
    - Progress Bar: Live-Updates mit Spinner, Prozent, Geschwindigkeit, ETA
    - Colored Output: Success/Error/Warning-Highlighting mit `--no-color` Flag
    - Help Text: Erweiterte Beispiele und Beschreibungen für alle Commands
  - **Phase 3: Extended Commands** (Tasks #47, #48)
    - `esdedb version`: Detaillierte Version-Info (Version, Go-Version, Commit, Build-Zeit)
      - `--json` Flag für maschinenlesbares Format
    - `esdedb stats`: Datenbank-Statistiken (Tabellen, Zeilen, Größe)
      - Detaillierte Auflistung aller Tabellen mit Zeilenzahlen
      - Gesamtstatistik: Anzahl Tabellen, Gesamtzeilen, DB-Dateigröße
  - **Phase 4: Shell Completion** (Task #49)
    - `esdedb completion bash/zsh/fish`: Shell-Completion-Generierung
    - Automatische Flag/Subcommand-Vervollständigung
    - Installation-Dokumentation für alle unterstützten Shells
  - **Phase 5: Configuration Management** (Task #50)
    - `esdedb config init`: Interaktive Config-Erstellung
    - `esdedb config convert`: Konvertierung vorhandener Configs
  - **Integration Tests** (Task #51)
    - 15+ Integration-Tests für alle Commands
    - Exit-Code-Validierung
    - Error-Handling-Tests
  - **Dokumentation** (Task #52)
    - Vollständige CLI-Dokumentation in `docs/cli/README.md`
    - Command-Referenz mit Beispielen
    - Installation- und Usage-Guides
- **CI/CD: Automated Quality Gates** (PR #149, commit 18993fe, 501df55)
  - GitHub Actions Lint Workflow
    - Automatische golangci-lint Ausführung bei jedem Push/PR
    - Security-gehärtet mit expliziten Permissions (CWE-275 fix)
    - Auto-Approval für established GitHub accounts (Copilot, Dependabot)
  - Dokumentation: `docs/ci-cd/bot-workflow-approval.md`
- **Code Quality: Comprehensive Lint Fixes** (commit a8fceb4)
  - 60+ errcheck Warnings behoben über 27 Files
  - Patterns: Explizite Error-Ignorierung (`_ =`), defer-Wrapper
  - Bereiche: cmd, internal/cli, config, database, parser, worker, tools
  - Alle Tests passing, Kompilierung clean
- **Epic #4 Complete: All 51 EVE SDE Parsers** (PR #135, PR #136)
  - **Phase 1 (Task #37)**: 7 zusätzliche Core Parsers
    - Inventory: InvCategoriesParser, InvMarketGroupsParser, InvMetaGroupsParser
    - Universe: MapStargatesParser, MapPlanetsParser
    - Character: ChrRacesParser, ChrFactionsParser
    - Gesamt Core Parsers: 17 (von geplanten 10-15)
  - **Phase 2 (Task #38)**: 34 Extended Parsers
    - Character/NPC: ancestries, bloodlines, attributes, npcCharacters, npcCorporations, npcCorporationDivisions, npcStations
    - Agents: agentTypes, agentsInSpace
    - Dogma Extended: attributeCategories, units, typeDogma, dynamicItemAttributes
    - Universe Extended: moons, stars, asteroidBelts, landmarks
    - Certificates/Skills: certificates, masteries
    - Skins: skins, skinLicenses, skinMaterials
    - Translation: translationLanguages
    - Station: operations, services, sovereigntyUpgrades
    - Miscellaneous: icons, graphics, contrabandTypes, controlTowerResources, corporationActivities, dbuffCollections, planetResources, planetSchematics, typeBonuses, _sde metadata
  - **Gesamt: 51 Parser** (100% aller EVE SDE Tabellen abgedeckt)
  - **171+ Tests** alle erfolgreich
  - **Code Organization**: Split in `parsers.go` (Core, 17 Parser) und `parsers_extended.go` (Extended, 36 Parser)
  - **JSON Schemas**: Alle 51 Schemas in `schemas/` Verzeichnis verfügbar
- **Developer Experience: Make Setup Target** (Commit cd0f404)
  - Neues `make setup` Target für komplette Projekt-Initialisierung
  - Automatische Installation von Go Dependencies
  - Automatische Installation von quicktype (wenn npm verfügbar)
  - Automatische Parser-Code-Generierung aus Schemas
  - Fix: `tools/generate-parsers.sh` verwendet korrekten Tool-Pfad
  - Dokumentation: README.md aktualisiert mit Setup-Anleitung nach `git clone`

### Changed

### Fixed

### Removed

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
