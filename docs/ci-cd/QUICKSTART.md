# CI/CD Quick Start Guide

Schnelleinstieg in die CI/CD Pipeline für EVE SDE Database Builder.

## Für Entwickler

### Vor dem Push

```bash
# Alle wichtigen Checks lokal ausführen:
make lint          # Code-Qualität prüfen
make test          # Tests ausführen
make coverage      # Coverage-Report generieren
make build         # Kompilierung prüfen
```

### Pull Request erstellen

1. Branch erstellen:
   ```bash
   git checkout -b feat/my-feature
   ```

2. Code ändern und committen:
   ```bash
   git add .
   git commit -m "feat: Add new feature"
   ```

3. Push:
   ```bash
   git push origin feat/my-feature
   ```

4. PR erstellen auf GitHub

5. **Automatisch ausgeführt:**
   - ✅ Lint Check
   - ✅ Test mit Coverage
   - ✅ Build Check
   - ✅ Coverage-Kommentar im PR

### Nach dem Review

- PR wird gemerged → Workflows laufen auf `main`
- Coverage wird getrackt
- Status Badges werden aktualisiert

---

## Für Maintainer

### Release erstellen

1. **Version aktualisieren:**
   ```bash
   echo "0.3.0" > VERSION
   ```

2. **CHANGELOG.md aktualisieren:**
   ```markdown
   ## [Unreleased]
   
   ## [0.3.0] - 2025-10-18
   
   ### Added
   - CI/CD Pipeline mit Release Automation
   - Coverage Reports
   
   ### Changed
   - ...
   
   ### Fixed
   - ...
   ```

3. **Release-Bereitschaft prüfen:**
   ```bash
   make release-check
   ```

4. **Commit und Tag erstellen:**
   ```bash
   git add VERSION CHANGELOG.md
   git commit -m "chore: Release v0.3.0"
   git tag v0.3.0
   ```

5. **Push mit Tag:**
   ```bash
   git push origin main
   git push origin v0.3.0
   ```

6. **Automatisch erstellt:**
   - ✅ Multi-Platform Binaries (Linux, macOS, Windows)
   - ✅ GitHub Release mit Changelog
   - ✅ Download-Links für alle Plattformen

### Release-Artefakte

Nach dem Release verfügbar auf: `https://github.com/Sternrassler/EVE-SDE-Database-Builder/releases`

**Dateien:**
- `esdedb-{version}-linux-amd64.tar.gz`
- `esdedb-{version}-linux-arm64.tar.gz`
- `esdedb-{version}-darwin-amd64.tar.gz`
- `esdedb-{version}-darwin-arm64.tar.gz`
- `esdedb-{version}-windows-amd64.zip`

---

## Troubleshooting

### Workflow schlägt fehl

**Lint-Fehler:**
```bash
# Lokal ausführen:
make lint

# Automatische Fixes:
make fmt
go mod tidy
```

**Test-Fehler:**
```bash
# Verbose Test-Ausgabe:
go test -v ./...

# Spezifischen Test:
go test -v -run TestName ./path/to/package
```

**Build-Fehler:**
```bash
# Dependencies aktualisieren:
go mod tidy
go mod verify

# Clean Build:
make clean
make build
```

### Release schlägt fehl

1. **Version-Format prüfen:**
   - Tag muss `v*.*.*` Format haben (z.B. `v0.3.0`)

2. **CHANGELOG prüfen:**
   - Version muss in CHANGELOG.md vorhanden sein

3. **Logs anzeigen:**
   ```bash
   gh run list --workflow=release.yml
   gh run view <run-id> --log
   ```

---

## Status anzeigen

### GitHub CLI

```bash
# Alle Workflows:
gh workflow list

# Letzte Runs:
gh run list

# Spezifischen Run anzeigen:
gh run view <run-id>

# Logs anzeigen:
gh run view <run-id> --log

# Artifacts herunterladen:
gh run download <run-id>
```

### GitHub Web

- **Workflows:** `https://github.com/Sternrassler/EVE-SDE-Database-Builder/actions`
- **Releases:** `https://github.com/Sternrassler/EVE-SDE-Database-Builder/releases`
- **Coverage:** Artifacts in Workflow-Runs

---

## Best Practices

### 1. Vor jedem Push

```bash
make lint test build
```

### 2. Vor jedem Release

```bash
make release-check
```

### 3. Coverage beobachten

- Coverage-Reports in PR-Kommentaren prüfen
- HTML-Report bei Bedarf von Artifacts herunterladen

### 4. Dependencies aktualisieren

```bash
go get -u ./...
go mod tidy
make test
```

---

## Weitere Informationen

- **Vollständige Dokumentation:** [README.md](README.md)
- **Workflow-Details:** Siehe `.github/workflows/*.yml`
- **Contributing:** [.github/copilot-instructions.md](../../.github/copilot-instructions.md)
