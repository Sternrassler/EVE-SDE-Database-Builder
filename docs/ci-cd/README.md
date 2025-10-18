# CI/CD Pipeline Dokumentation

Umfassende Dokumentation der GitHub Actions Workflows für EVE SDE Database Builder.

## Übersicht

Das Projekt verwendet mehrere GitHub Actions Workflows für Continuous Integration und Continuous Deployment:

1. **PR Quality Check** - Qualitätsprüfungen für Pull Requests
2. **Benchmark** - Performance-Regression Tests
3. **Coverage** - Test Coverage Tracking
4. **Lint** - Code Linting
5. **Test** - Unit Tests mit Race Detector
6. **Release** - Automatisierte Release-Erstellung

---

## Workflows

### 1. PR Quality Check (`.github/workflows/pr-check.yml`)

**Trigger:** Pull Request auf `main` oder `master` Branch

**Jobs:**
- **Lint**: Führt golangci-lint aus
- **Test**: Führt alle Tests mit Coverage aus
- **Build**: Prüft ob das Projekt kompiliert

**Features:**
- ✅ Automatische Coverage-Berechnung
- ✅ PR-Kommentar mit Coverage-Prozentsatz
- ✅ Coverage-Report als Artifact

**Permissions:**
- `contents: read` - Repository-Inhalte lesen
- `pull-requests: write` - PR-Kommentare erstellen

**Nutzung:**
```bash
# Lokal vor dem Push testen:
make lint
make test
make build
```

---

### 2. Benchmark (`.github/workflows/benchmark.yml`)

**Trigger:**
- Pull Request auf `main` oder `master` Branch (bei Go-Datei-Änderungen)
- Workflow Dispatch (manueller Trigger)

**Features:**
- ✅ Performance-Regression Tests gegen Baseline
- ✅ Worker Pool, Parser und Database Benchmarks
- ✅ Automatischer Vergleich mit `benchstat`
- ✅ PR-Kommentar mit Benchmark-Ergebnissen
- ✅ Warnung bei >15% Performance-Regression

**Baseline-Dateien:**
- `benchmarks/baseline-worker.txt` - Worker Pool Benchmarks
- `benchmarks/baseline-parser.txt` - Parser Benchmarks
- `benchmarks/baseline-database.txt` - Database Benchmarks

**Nutzung:**
```bash
# Lokal Benchmarks ausführen:
make bench

# Neue Baseline erfassen:
make bench-baseline

# Gegen Baseline vergleichen:
make bench-compare
```

**Benchmark-Metriken:**
- `ns/op` - Nanosekunden pro Operation
- `B/op` - Bytes alloziert pro Operation
- `allocs/op` - Anzahl Allocations pro Operation

**Regression-Schwellwerte:**
- Performance: >15% langsamer → CI fails
- Memory: >20% mehr Allocations → Warnung

**Baseline aktualisieren:**
```bash
# Nach Performance-Optimierungen oder Architektur-Änderungen
make bench-baseline
git add benchmarks/
git commit -m "chore: Update benchmark baseline"
```

---

### 3. Coverage (`.github/workflows/coverage.yml`)

**Trigger:** 
- Push auf `main` oder `master` Branch
- Pull Request auf `main` oder `master` Branch

**Features:**
- ✅ Coverage-Report-Generierung
- ✅ Upload zu Codecov.io mit Trend Tracking
- ✅ HTML Coverage Report als Artifact
- ✅ Text Coverage Summary
- ✅ Codecov Badge im README mit Echtzeit-Prozentsatz

**Codecov.io Integration:**

Die Codecov-Integration ermöglicht:
- 📊 Visuelle Coverage-Dashboards
- 📈 Trend Tracking über Zeit
- 🎯 Coverage-Ziele und Thresholds
- 💬 Automatische PR-Kommentare mit Coverage-Diff
- 🚨 Warnung bei Coverage-Rückgang

**Konfiguration:** Siehe `codecov.yml` im Repository-Root

**Thresholds:**
- Project Coverage: Auto-Target mit 0.5% Threshold
- Patch Coverage: Auto-Target mit 0.5% Threshold
- Range: 70-100%
- Precision: 2 Dezimalstellen

**Ignorierte Pfade:**
- `tools/**` - Code-Generierungs-Tools
- `**/*_test.go` - Test-Dateien selbst
- `testdata/**` - Test-Fixtures
- `migrations/**` - SQL-Migrations
- `schemas/**` - JSON-Schemas

**Artifacts:**
- `coverage-report` - `coverage.out` und `coverage.txt`
- `coverage-html` - HTML-Report für Browser-Ansicht

**Nutzung:**
```bash
# Lokal Coverage generieren:
make coverage

# HTML Report anzeigen:
go tool cover -html=coverage.out
```

**Codecov Dashboard:**
- URL: https://codecov.io/gh/Sternrassler/EVE-SDE-Database-Builder
- Badge: ![codecov](https://codecov.io/gh/Sternrassler/EVE-SDE-Database-Builder/branch/main/graph/badge.svg)

---

### 4. Lint (`.github/workflows/lint.yml`)

**Trigger:**
- Push auf `main` oder `master` Branch
- Pull Request auf `main` oder `master` Branch

**Features:**
- ✅ golangci-lint mit 5 Minuten Timeout
- ✅ Automatische Go Version aus `go.mod`

**Nutzung:**
```bash
# Lokal ausführen:
make lint

# golangci-lint installieren (falls nicht vorhanden):
# https://golangci-lint.run/usage/install/
```

---

### 5. Test (`.github/workflows/test.yml`)

**Trigger:**
- Push auf `main` oder `master` Branch
- Pull Request auf `main` oder `master` Branch

**Features:**
- ✅ Tests mit Race Detector
- ✅ 10 Minuten Timeout
- ✅ Go Version aus `go.mod`

**Nutzung:**
```bash
# Lokal ausführen:
make test

# Mit Race Detector:
make test-race
```

---

### 6. Release (`.github/workflows/release.yml`)

**Trigger:** Git Tag mit Format `v*.*.*` (z.B. `v0.3.0`)

**Features:**
- ✅ Multi-Platform Builds
- ✅ Automatische Changelog-Extraktion
- ✅ GitHub Release mit Binaries
- ✅ Archive-Erstellung (tar.gz / zip)

**Unterstützte Plattformen:**
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

**Build-Artefakte:**
- `esdedb-{version}-{platform}.tar.gz` (Linux/macOS)
- `esdedb-{version}-{platform}.zip` (Windows)

**Nutzung:**

```bash
# 1. Version in VERSION-Datei aktualisieren
echo "0.3.0" > VERSION

# 2. CHANGELOG.md aktualisieren
# - Unreleased → [0.3.0] - YYYY-MM-DD
# - Änderungen dokumentieren

# 3. Commit und Tag erstellen
git add VERSION CHANGELOG.md
git commit -m "chore: Release v0.3.0"
git tag v0.3.0

# 4. Push mit Tags
git push origin main
git push origin v0.3.0

# 5. GitHub Actions erstellt automatisch:
#    - Release mit Changelog
#    - Binaries für alle Plattformen
```

**Release-Naming:**
- Tag: `v0.3.0`
- Release Name: `Release v0.3.0`
- Binaries: `esdedb-0.3.0-{platform}`

**Changelog-Extraktion:**

Der Workflow extrahiert automatisch den relevanten Abschnitt aus `CHANGELOG.md`:

1. Sucht nach `[{version}]` Sektion
2. Falls nicht gefunden, nutzt `[Unreleased]` Sektion
3. Extrahiert Content bis zur nächsten Version-Sektion

**Beispiel CHANGELOG.md Format:**

```markdown
## [Unreleased]

### Added
- Feature X

## [0.3.0] - 2025-10-18

### Added
- CI/CD Pipeline mit Release Automation
- Coverage Reports

### Fixed
- Bug Y
```

---

## Secrets & Konfiguration

### Erforderliche Secrets

**Codecov Integration:**
- `CODECOV_TOKEN` - Für Codecov.io Upload (Coverage Workflow)
  - Quelle: https://codecov.io/gh/Sternrassler/EVE-SDE-Database-Builder/settings
  - Scope: Repository-spezifischer Upload-Token
  - Setup: GitHub Settings → Secrets → Actions → New repository secret
  - Name: `CODECOV_TOKEN`
  - Value: Token von Codecov.io Dashboard

**Automatisch verfügbar:**
- `GITHUB_TOKEN` - Automatisch von GitHub Actions bereitgestellt

**Token Setup (Codecov.io):**

1. Account erstellen/einloggen auf https://codecov.io/
2. Repository autorisieren (via GitHub App)
3. Upload Token kopieren aus Repository Settings
4. Als GitHub Secret hinzufügen (siehe oben)
5. Nach erstem Coverage-Upload: Badge-Link verfügbar

**Hinweis:** Der Coverage-Workflow funktioniert auch ohne Token (mit reduzierter Funktionalität), aber für vollständige Features und private Repos ist der Token erforderlich.

### Branch Protection

Empfohlene Branch Protection Rules für `main`:

- ✅ Require pull request reviews before merging
- ✅ Require status checks to pass before merging:
  - `Lint`
  - `Test with Coverage`
  - `Build Check`
- ✅ Require branches to be up to date before merging
- ✅ Do not allow bypassing the above settings

---

## Monitoring & Debugging

### Workflow-Status prüfen

```bash
# GitHub CLI verwenden:
gh workflow list
gh run list
gh run view <run-id>
```

### Logs anzeigen

```bash
# Letzte Workflow-Runs anzeigen:
gh run list --workflow=pr-check.yml

# Logs eines spezifischen Runs:
gh run view <run-id> --log
```

### Artifacts herunterladen

```bash
# Alle Artifacts eines Runs:
gh run download <run-id>

# Spezifisches Artifact:
gh run download <run-id> -n coverage-report
```

---

## Lokale Validierung

Vor dem Push empfohlen:

```bash
# Kompletter PR-Check:
make lint
make test
make build

# Mit Coverage:
make coverage

# Performance Regression Check:
make bench-compare

# Optional: CI lokal simulieren mit act
# https://github.com/nektos/act
act pull_request
```

---

## Troubleshooting

### Build-Fehler

```bash
# Go Dependencies aktualisieren:
go mod tidy
go mod verify

# Clean Build:
make clean
make build
```

### Lint-Fehler

```bash
# Automatische Fixes:
make fmt
go mod tidy

# Detaillierte Lint-Ausgabe:
golangci-lint run --verbose
```

### Test-Fehler

```bash
# Verbose Test-Ausgabe:
go test -v ./...

# Spezifischen Test debuggen:
go test -v -run TestName ./path/to/package
```

### Coverage zu niedrig

```bash
# Fehlende Test Coverage identifizieren:
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -v 100.0%

# HTML Report für visuelle Analyse:
go tool cover -html=coverage.out
```

---

## Best Practices

### 1. Commit-Hygiene

- ✅ Kleine, fokussierte Commits
- ✅ Aussagekräftige Commit-Messages
- ✅ Tests vor Push ausführen

### 2. PR-Workflow

- ✅ Feature Branch von `main` erstellen
- ✅ Regelmäßig Tests lokal ausführen
- ✅ PR erst erstellen wenn alle Checks grün
- ✅ Review-Feedback zeitnah adressieren

### 3. Release-Workflow

- ✅ Version-Bump in `VERSION`-Datei
- ✅ `CHANGELOG.md` aktualisieren
- ✅ Tests vor Release ausführen
- ✅ Tag nach Schema `v*.*.*` erstellen

---

## Performance

### Durchschnittliche Workflow-Zeiten

- **Lint**: ~1-2 Minuten
- **Test**: ~2-5 Minuten
- **Coverage**: ~3-5 Minuten
- **Benchmark**: ~5-10 Minuten
- **Release**: ~5-10 Minuten (alle Plattformen)

### Optimierungen

- ✅ Go Module Caching aktiviert
- ✅ Parallelisierung von Jobs
- ✅ Matrix Builds für Release
- ✅ Artifact Retention (7-30 Tage)

---

## Referenzen

- [GitHub Actions Dokumentation](https://docs.github.com/en/actions)
- [golangci-lint](https://golangci-lint.run/)
- [Codecov](https://codecov.io/)
- [Keep a Changelog](https://keepachangelog.com/)
- [Semantic Versioning](https://semver.org/)
