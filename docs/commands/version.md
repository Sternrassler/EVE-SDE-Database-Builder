# Version Command Documentation

## Übersicht

Der `version` Command zeigt erweiterte Versionsinformationen für das EVE SDE Database Builder Tool an.

## Verwendung

```bash
esdedb version [flags]
```

## Flags

- `--format string`: Ausgabeformat (text oder json) (default: "text")
- `-h, --help`: Hilfe für version Command anzeigen

## Ausgabeinformationen

Der Command zeigt folgende Informationen an:

1. **Version**: Aus der `VERSION` Datei oder zur Build-Zeit gesetzt
2. **Commit**: Git Commit Hash (SHA)
3. **Build Time**: Zeitpunkt des Builds (ISO 8601 Format)

## Ausgabeformate

### Text Format (Standard)

Menschenlesbare Ausgabe mit klarer Formatierung:

```bash
$ esdedb version
Version:    0.2.0
Commit:     4ea694374865e69a3505b1851ec2c7f3e1c92ff8
Build Time: 2025-10-17T16:20:42Z
```

### JSON Format

Maschinenlesbare JSON-Ausgabe für CI/CD oder Automatisierung:

```bash
$ esdedb version --format json
{
  "version": "0.2.0",
  "commit": "4ea694374865e69a3505b1851ec2c7f3e1c92ff8",
  "buildTime": "2025-10-17T16:20:42Z"
}
```

## Beispiele

### Standard Text-Ausgabe

```bash
esdedb version
```

### JSON-Format für maschinelle Verarbeitung

```bash
esdedb version --format json
```

### In CI/CD Pipeline verwenden

```bash
# Version in Variable speichern
VERSION=$(esdedb version --format json | jq -r '.version')
echo "Building with version: $VERSION"
```

### Mit jq parsen

```bash
# Alle Felder einzeln extrahieren
esdedb version --format json | jq -r '.version, .commit, .buildTime'
```

## Build-Zeit Injection

### Development Builds

Ohne spezielle Build-Flags werden Default-Werte verwendet:

```bash
go build -o esdedb ./cmd/esdedb/
```

Ausgabe:
- Version: `dev`
- Commit: `unknown`
- Build Time: `unknown`

### Production Builds

Für Production Builds sollten die Werte zur Build-Zeit injiziert werden:

```bash
VERSION=$(cat VERSION | tr -d '\n')
COMMIT=$(git log -1 --pretty=format:"%H")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

go build -ldflags "\
  -X main.version=${VERSION} \
  -X main.commit=${COMMIT} \
  -X main.buildTime=${BUILD_TIME}" \
  -o esdedb ./cmd/esdedb/
```

### Makefile Integration

Empfohlene Erweiterung des Makefile `build` Targets:

```makefile
build: ## Build the project with version information
	@VERSION=$$(cat VERSION | tr -d '\n'); \
	COMMIT=$$(git log -1 --pretty=format:"%H"); \
	BUILD_TIME=$$(date -u +"%Y-%m-%dT%H:%M:%SZ"); \
	go build -ldflags "\
	  -X main.version=$${VERSION} \
	  -X main.commit=$${COMMIT} \
	  -X main.buildTime=$${BUILD_TIME}" \
	  -o esdedb ./cmd/esdedb/
```

### GitHub Actions Integration

Beispiel für GitHub Actions Workflow:

```yaml
- name: Build with version info
  run: |
    VERSION=$(cat VERSION | tr -d '\n')
    COMMIT=${{ github.sha }}
    BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    go build -ldflags "\
      -X main.version=${VERSION} \
      -X main.commit=${COMMIT} \
      -X main.buildTime=${BUILD_TIME}" \
      -o esdedb ./cmd/esdedb/
```

## Fehlerbehandlung

### Ungültiges Format

Wenn ein nicht unterstütztes Format angegeben wird, gibt der Command einen Fehler zurück:

```bash
$ esdedb version --format xml
Error: unsupported format: xml (use 'text' or 'json')
```

Exit Code: `1`

## Tests

Der Command ist vollständig getestet:

```bash
# Alle Version Command Tests ausführen
go test -v ./cmd/esdedb/ -run TestVersion

# Spezifische Tests
go test -v ./cmd/esdedb/ -run TestVersionCmd_TextFormat
go test -v ./cmd/esdedb/ -run TestVersionCmd_JSONFormat
```

## Siehe auch

- `esdedb --version`: Zeigt kurze Version im Root Command Format
- `esdedb --help`: Allgemeine Hilfe
- [VERSION Datei](../../VERSION): Single Source of Truth für Version
