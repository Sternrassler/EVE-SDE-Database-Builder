# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

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
