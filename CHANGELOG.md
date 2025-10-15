# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Makefile vereinfacht: 15 → 9 Targets mit klarer Hierarchie (check-hooks, check-local, check-pr, check-ci)
- Alle Make Targets delegieren zu Scripts in `scripts/` (keine Code-Duplikation mehr)
- Git Hooks nutzen Make Targets statt direkte Script-Aufrufe
- Workflow `pr-quality-gates.yml` nutzt `make check-ci` statt gelöschtes `make pr-quality-gates-ci`

### Added

- Make Target `check-hooks`: Governance-Checks für Git Hooks (normative + adr)
- Make Target `security-blockers`: Prüft kritische Security Findings aus Trivy Report
- Atomare Targets: `normative-check`, `adr-check`, `adr-ref`, `commit-lint`, `release-check`
- Git Hook Pfad konfiguriert: `.githooks` als Standard

### Removed

- Verwaistes Script `scripts/workflows/pr-quality-gates/run.sh` (Logik in Makefile konsolidiert)
- Redundante Make Targets: `lint-ci`, `scan-json`, `pr-check`, `push-ci`, `ci-local`, `pr-quality-gates-ci`

## [0.1.0] - 2025-10-05

- Project initialization.
