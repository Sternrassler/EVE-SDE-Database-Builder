# EVE SDE Database Builder - CLI Dokumentation

## Übersicht

Das EVE SDE Database Builder CLI (`esdedb`) ist ein Kommandozeilen-Tool für den Import von EVE Online Static Data Export (SDE) JSONL-Dateien in eine SQLite-Datenbank. Das Tool bietet eine moderne, benutzerfreundliche Oberfläche mit strukturiertem Logging, Fortschrittsanzeigen und umfangreichen Konfigurationsmöglichkeiten.

## Installation

```bash
# Von Source bauen
make build

# Direkter Go Build
go build -o esdedb ./cmd/esdedb/
```

## Schnellstart

```bash
# Import mit Standard-Einstellungen
esdedb import

# Import mit benutzerdefinierten Pfaden
esdedb import --sde-dir ./sde-JSONL --db ./eve-sde.db

# Konfiguration validieren
esdedb validate --config ./config.toml

# Datenbank-Statistiken anzeigen
esdedb stats --db ./eve-sde.db

# Version anzeigen
esdedb version
```

## Verfügbare Commands

### Haupt-Commands

| Command | Beschreibung | Referenz |
|---------|-------------|----------|
| `import` | Importiert SDE JSONL-Dateien in SQLite-Datenbank | [Import Command](#import-command) |
| `validate` | Validiert eine TOML-Konfigurationsdatei | [Validate Command](#validate-command) |
| `version` | Zeigt erweiterte Versionsinformationen an | [Version Command](#version-command) |
| `stats` | Zeigt Datenbank-Statistiken an | [Stats Command](#stats-command) |
| `completion` | Generiert Shell-Completion-Scripte | [Completion Command](#completion-command) |

### Utility Commands

| Command | Beschreibung |
|---------|-------------|
| `help` | Hilfe für jeden Command |

## Globale Flags

Diese Flags sind für alle Commands verfügbar:

| Flag | Shorthand | Default | Beschreibung |
|------|-----------|---------|--------------|
| `--config` | `-c` | `./config.toml` | Pfad zur TOML-Konfigurationsdatei |
| `--verbose` | `-v` | `false` | Aktiviert Verbose Logging (Debug Level) |
| `--no-color` | - | `false` | Deaktiviert farbige Konsolenausgabe |
| `--help` | `-h` | - | Zeigt Hilfe an |
| `--version` | - | - | Zeigt Version an (nur Root Command) |

### Globale Flags - Details

#### `--config` / `-c`

Gibt den Pfad zur TOML-Konfigurationsdatei an. Die Konfigurationsdatei kann alle CLI-Parameter enthalten und dient als zentrale Konfigurationsquelle.

```bash
esdedb --config /etc/esdedb/config.toml import
```

**Siehe auch:** [ADR-004: Configuration Format](../adr/ADR-004-configuration-format.md)

#### `--verbose` / `-v`

Aktiviert detailliertes Logging auf Debug-Level. Nützlich für Troubleshooting und Entwicklung.

```bash
esdedb --verbose import --sde-dir ./sde-JSONL
```

**Log Level:**
- Standard: `info` (strukturierte JSON Logs)
- Mit `--verbose`: `debug` (detaillierte Trace-Informationen)

#### `--no-color`

Deaktiviert farbige Terminal-Ausgabe. Nützlich für Log-Dateien, CI/CD-Pipelines oder Terminals ohne Farbunterstützung.

```bash
esdedb --no-color import | tee import.log
```

## Import Command

Der `import` Command führt den 2-Phasen-Import von EVE SDE JSONL-Dateien in eine SQLite-Datenbank aus.

### Verwendung

```bash
esdedb import [flags]
```

### Import-Phasen

#### Phase 1: Paralleles Parsing (Worker Pool)

- Mehrere Worker-Threads parsen JSONL-Dateien gleichzeitig
- Anzahl der Worker konfigurierbar (Standard: 4, Auto: -1 für NumCPU)
- Parallel Processing für maximale Performance
- Progress Tracking für jede Datei

#### Phase 2: Sequenzielles Database-Insert (SQLite Single-Writer)

- Geparste Daten werden in SQLite-Datenbank eingefügt
- SQLite unterstützt nur einen Writer zur gleichen Zeit
- Transaktions-basierte Inserts für Konsistenz
- Retry-Mechanismus für transiente Fehler

**Siehe auch:** [ADR-006: Concurrency & Worker Pool](../adr/ADR-006-concurrency-worker-pool.md)

### Flags

| Flag | Shorthand | Default | Beschreibung |
|------|-----------|---------|--------------|
| `--sde-dir` | `-s` | `./sde-JSONL` | Pfad zum Verzeichnis mit SDE JSONL-Dateien |
| `--db` | `-d` | `./eve-sde.db` | Pfad zur SQLite-Datenbank (wird erstellt falls nicht vorhanden) |
| `--workers` | `-w` | `4` | Anzahl paralleler Worker-Threads (-1 = Automatisch basierend auf CPU-Kernen) |
| `--skip-errors` | - | `false` | Überspringt fehlerhafte Dateien statt Import abzubrechen |

### Fortschrittsanzeige

Der Import zeigt einen Live-Fortschrittsbalken mit folgenden Metriken:

- **Anzahl verarbeiteter Dateien** (Parsed/Total)
- **Fehlgeschlagene Dateien** (Failed)
- **Eingefügte Rows** (Total)
- **Durchsatz** (Rows/Sekunde)
- **Geschätzte verbleibende Zeit** (ETA)

### Beispiele

#### Standard-Import (4 Workers)

```bash
esdedb import
```

#### Import mit benutzerdefinierten Pfaden

```bash
esdedb import --sde-dir ./sde-JSONL --db ./eve-sde.db --workers 4
```

#### Automatische Worker-Anzahl basierend auf CPU-Kernen

```bash
esdedb import --sde-dir ./sde-JSONL --db ./eve-sde.db --workers -1
```

#### Fehlerhafte Dateien überspringen

```bash
esdedb import --sde-dir ./sde-JSONL --db ./eve-sde.db --skip-errors
```

#### Import mit Verbose Logging

```bash
esdedb --verbose import --sde-dir ./sde-JSONL
```

#### Import mit Log-Datei

```bash
esdedb --no-color import 2>&1 | tee import.log
```

### Zusammenfassung nach Import

Nach erfolgreichem Import wird eine Zusammenfassung angezeigt:

```
=== Import Summary ===
Files:      150/150 parsed (0 failed)
Rows:       1234567 inserted
Duration:   2m15s
Throughput: 9134 rows/sec
```

### Fehlerbehandlung

#### Fehlende Verzeichnisse

```bash
$ esdedb import --sde-dir ./nicht-existent
Error: no JSONL files found in ./nicht-existent
```

#### Datenbankfehler

Bei Datenbankfehlern wird ein detaillierter Fehler ausgegeben mit:
- Fehlertyp
- Betroffene Datei
- Stack Trace (bei `--verbose`)

#### Graceful Shutdown

Der Import kann jederzeit mit `Ctrl+C` (SIGINT) oder `SIGTERM` sauber abgebrochen werden:

```
^C
Received interrupt signal, cancelling import...
Import cancelled by user
```

## Validate Command

Der `validate` Command prüft eine TOML-Konfigurationsdatei auf Gültigkeit.

### Verwendung

```bash
esdedb validate [flags]
```

### Validierungen

Der Command validiert folgende Aspekte:

1. **TOML-Syntax**: Datei muss gültiges TOML sein
2. **Erforderliche Felder**: `database.path`, `import.sde_path`
3. **Wertebereichsprüfungen**:
   - `workers`: 1-32
   - `language`: en/de/fr/ja/ru/zh/es/ko
4. **Logik-Konsistenz**: z.B. gültige Pfade, sinnvolle Werte

### Ausgabe bei erfolgreicher Validierung

```
Configuration is valid

Configuration Summary:
  Database Path: ./eve-sde.db
  SDE Path:      ./sde-JSONL
  Workers:       4
  Language:      en
  Log Level:     info
  Log Format:    json
```

### Beispiele

#### Standard-Konfiguration validieren

```bash
esdedb validate
```

#### Spezifische Konfigurationsdatei validieren

```bash
esdedb validate --config ./config.toml
```

#### Produktions-Konfiguration validieren

```bash
esdedb validate -c /etc/esdedb/config.toml
```

#### Mit Verbose Logging

```bash
esdedb --verbose validate --config ./config.toml
```

### Fehlerbehandlung

Bei Validierungsfehlern wird ein detaillierter Fehler ausgegeben:

```bash
$ esdedb validate --config invalid.toml
Configuration validation failed: invalid value for workers: 100 (must be between 1 and 32)
```

## Version Command

Der `version` Command zeigt erweiterte Versionsinformationen an.

### Verwendung

```bash
esdedb version [flags]
```

### Flags

| Flag | Default | Beschreibung |
|------|---------|--------------|
| `--format` | `text` | Ausgabeformat (text oder json) |

### Ausgabeinformationen

1. **Version**: Aus der `VERSION` Datei oder zur Build-Zeit gesetzt
2. **Commit**: Git Commit Hash (SHA)
3. **Build Time**: Zeitpunkt des Builds (ISO 8601 Format)

### Ausgabeformate

#### Text Format (Standard)

```bash
$ esdedb version
Version:    0.2.0
Commit:     4ea694374865e69a3505b1851ec2c7f3e1c92ff8
Build Time: 2025-10-17T16:20:42Z
```

#### JSON Format

```bash
$ esdedb version --format json
{
  "version": "0.2.0",
  "commit": "4ea694374865e69a3505b1851ec2c7f3e1c92ff8",
  "buildTime": "2025-10-17T16:20:42Z"
}
```

### Beispiele

#### Standard Text-Ausgabe

```bash
esdedb version
```

#### JSON-Format für maschinelle Verarbeitung

```bash
esdedb version --format json
```

#### In CI/CD Pipeline verwenden

```bash
# Version in Variable speichern
VERSION=$(esdedb version --format json | jq -r '.version')
echo "Building with version: $VERSION"
```

#### Mit jq parsen

```bash
# Alle Felder einzeln extrahieren
esdedb version --format json | jq -r '.version, .commit, .buildTime'
```

### Kurze Version via Root Command

```bash
$ esdedb --version
esdedb version dev (commit: unknown)
```

**Siehe auch:** [Version Command Dokumentation](../commands/version.md)

## Stats Command

Der `stats` Command zeigt Statistiken über die SQLite-Datenbank an.

### Verwendung

```bash
esdedb stats [flags]
```

### Flags

| Flag | Shorthand | Default | Beschreibung |
|------|-----------|---------|--------------|
| `--db` | `-d` | `./eve-sde.db` | Pfad zur SQLite-Datenbank |

### Ausgabeinformationen

Der Command zeigt folgende Informationen an:

1. **Datenbank-Pfad**: Pfad zur Datenbankdatei
2. **Dateigröße**: Größe der Datenbankdatei (formatiert)
3. **Tabellen-Anzahl**: Anzahl der Tabellen in der Datenbank
4. **Gesamtzahl Rows**: Summe aller Zeilen über alle Tabellen
5. **Tabellen-Details**: Name und Zeilenanzahl für jede Tabelle

### Beispiel-Ausgabe

```
=== Database Statistics ===
Database: ./eve-sde.db
Size:     256.4 MiB

Tables:   51
Total Rows: 1234567

Table Name                     Row Count
----------                     ---------
agents                              4321
blueprints                         12345
certificates                         123
...
```

### Beispiele

#### Statistiken für Standard-Datenbank anzeigen

```bash
esdedb stats --db ./eve-sde.db
```

#### Statistiken für benutzerdefinierte Datenbank

```bash
esdedb stats --db /path/to/custom.db
```

#### Mit Verbose Logging

```bash
esdedb --verbose stats --db ./eve-sde.db
```

### Fehlerbehandlung

#### Datenbank existiert nicht

```bash
$ esdedb stats --db non-existent.db
Error: database file does not exist: non-existent.db
```

## Completion Command

Der `completion` Command generiert Shell-Completion-Scripte für verschiedene Shells.

### Verwendung

```bash
esdedb completion [bash|zsh|fish|powershell]
```

### Unterstützte Shells

| Shell | Subcommand |
|-------|-----------|
| Bash | `esdedb completion bash` |
| Zsh | `esdedb completion zsh` |
| Fish | `esdedb completion fish` |
| PowerShell | `esdedb completion powershell` |

### Installation

#### Bash

```bash
# Für aktuelle Shell-Session
source <(esdedb completion bash)

# Permanent (Linux)
esdedb completion bash > /etc/bash_completion.d/esdedb

# Permanent (macOS)
esdedb completion bash > $(brew --prefix)/etc/bash_completion.d/esdedb
```

#### Zsh

```bash
# Für aktuelle Shell-Session
source <(esdedb completion zsh)

# Permanent (Linux)
esdedb completion zsh > "${fpath[1]}/_esdedb"

# Permanent (macOS)
esdedb completion zsh > $(brew --prefix)/share/zsh/site-functions/_esdedb
```

#### Fish

```bash
# Für aktuelle Shell-Session
esdedb completion fish | source

# Permanent
esdedb completion fish > ~/.config/fish/completions/esdedb.fish
```

#### PowerShell

```powershell
# Für aktuelle Shell-Session
esdedb completion powershell | Out-String | Invoke-Expression

# Permanent (zu PowerShell-Profil hinzufügen)
esdedb completion powershell >> $PROFILE
```

### Hinweise

- Nach der Installation der Completion-Scripte muss eine neue Shell gestartet werden
- Completion funktioniert für Commands, Flags und Flag-Werte
- Bei Problemen: `--help` Flag verwenden für Hinweise

## Flag-Referenz

### Vollständige Flag-Übersicht

| Command | Flag | Shorthand | Type | Default | Beschreibung |
|---------|------|-----------|------|---------|--------------|
| (Global) | `--config` | `-c` | string | `./config.toml` | Pfad zur TOML-Konfigurationsdatei |
| (Global) | `--verbose` | `-v` | bool | `false` | Aktiviert Verbose Logging (Debug Level) |
| (Global) | `--no-color` | - | bool | `false` | Deaktiviert farbige Konsolenausgabe |
| (Global) | `--help` | `-h` | bool | `false` | Zeigt Hilfe an |
| (Global) | `--version` | - | bool | `false` | Zeigt Version an (nur Root) |
| import | `--sde-dir` | `-s` | string | `./sde-JSONL` | Pfad zum SDE-Verzeichnis |
| import | `--db` | `-d` | string | `./eve-sde.db` | Pfad zur SQLite-Datenbank |
| import | `--workers` | `-w` | int | `4` | Anzahl paralleler Worker (-1 = auto) |
| import | `--skip-errors` | - | bool | `false` | Fehlerhafte Dateien überspringen |
| stats | `--db` | `-d` | string | `./eve-sde.db` | Pfad zur SQLite-Datenbank |
| version | `--format` | - | string | `text` | Ausgabeformat (text oder json) |

### Flag-Typen

- **string**: Textuelle Werte (Pfade, Namen, etc.)
- **bool**: Boolean-Werte (true/false)
- **int**: Ganzzahlige Werte

### Flag-Priorität

Die Konfiguration wird in folgender Reihenfolge angewendet (höchste Priorität zuerst):

1. **CLI Flags**: Explizit übergebene Flags
2. **Environment Variables**: Umgebungsvariablen (siehe ADR-004)
3. **Config File**: Werte aus TOML-Konfigurationsdatei
4. **Defaults**: Eingebaute Standard-Werte

## Häufige Workflows

### Erstmaliger Import

```bash
# 1. Konfiguration erstellen
cp config.toml.example config.toml

# 2. Konfiguration anpassen und validieren
esdedb validate --config ./config.toml

# 3. Import durchführen
esdedb import --sde-dir ./sde-JSONL --db ./eve-sde.db

# 4. Ergebnis prüfen
esdedb stats --db ./eve-sde.db
```

### Update bestehender Datenbank

```bash
# 1. Backup erstellen
cp eve-sde.db eve-sde.db.backup

# 2. Neue SDE-Daten importieren
esdedb import --sde-dir ./sde-JSONL-new --db ./eve-sde.db

# 3. Statistiken vergleichen
esdedb stats --db ./eve-sde.db
```

### Troubleshooting

```bash
# 1. Verbose Logging aktivieren
esdedb --verbose import --sde-dir ./sde-JSONL 2>&1 | tee debug.log

# 2. Fehlerhafte Dateien überspringen
esdedb import --skip-errors --sde-dir ./sde-JSONL

# 3. Worker-Anzahl reduzieren
esdedb import --workers 1 --sde-dir ./sde-JSONL
```

### CI/CD Integration

```bash
# Version in CI/CD Pipeline abrufen
VERSION=$(esdedb version --format json | jq -r '.version')

# Import in CI/CD mit Logging
esdedb --no-color import --sde-dir /data/sde-JSONL --db /data/eve-sde.db 2>&1 | tee /logs/import-${VERSION}.log

# Exit Code prüfen
if [ $? -eq 0 ]; then
  echo "Import successful"
  esdedb stats --db /data/eve-sde.db
else
  echo "Import failed"
  exit 1
fi
```

## Exit Codes

| Exit Code | Bedeutung |
|-----------|-----------|
| `0` | Erfolgreiche Ausführung |
| `1` | Allgemeiner Fehler (Command fehlgeschlagen) |

## Logging

### Log-Formate

Das Tool unterstützt strukturiertes Logging mit zwei Formaten:

- **JSON** (Standard): Maschinenlesbare Logs für Analyse-Tools
- **Text**: Menschenlesbare Logs für Entwicklung

### Log-Levels

- **error**: Kritische Fehler, die die Ausführung verhindern
- **warn**: Warnungen, die Aufmerksamkeit erfordern
- **info**: Informative Meldungen über den Programmfluss
- **debug**: Detaillierte Debug-Informationen (nur mit `--verbose`)

### Log-Beispiele

#### JSON Format (Standard)

```json
{"level":"info","version":"0.2.0","commit":"abc123","time":"2025-10-17T17:28:15Z","message":"Application started"}
{"level":"info","sde_dir":"./sde-JSONL","db_path":"./eve-sde.db","workers":4,"skip_errors":false,"time":"2025-10-17T17:28:15Z","message":"Starting EVE SDE Import"}
{"level":"info","total_files":150,"parsed_files":150,"inserted_files":150,"failed_files":0,"inserted_rows":1234567,"duration":"2m15s","rows_per_second":9134.2,"time":"2025-10-17T17:30:30Z","message":"Import completed"}
```

#### Text Format (mit --verbose)

```
2025-10-17T17:28:15Z INFO Application started version=0.2.0 commit=abc123
2025-10-17T17:28:15Z INFO Starting EVE SDE Import sde_dir=./sde-JSONL db_path=./eve-sde.db workers=4 skip_errors=false
2025-10-17T17:30:30Z INFO Import completed total_files=150 parsed_files=150 inserted_files=150 failed_files=0 inserted_rows=1234567 duration=2m15s rows_per_second=9134.2
```

## Umgebungsvariablen

Das Tool unterstützt Konfiguration über Umgebungsvariablen (siehe ADR-004):

| Variable | Entspricht Flag | Beispiel |
|----------|----------------|----------|
| `ESDEDB_CONFIG` | `--config` | `ESDEDB_CONFIG=/etc/esdedb/config.toml` |
| `ESDEDB_SDE_DIR` | `--sde-dir` | `ESDEDB_SDE_DIR=/data/sde-JSONL` |
| `ESDEDB_DB_PATH` | `--db` | `ESDEDB_DB_PATH=/data/eve-sde.db` |
| `ESDEDB_WORKERS` | `--workers` | `ESDEDB_WORKERS=8` |
| `ESDEDB_VERBOSE` | `--verbose` | `ESDEDB_VERBOSE=true` |
| `ESDEDB_NO_COLOR` | `--no-color` | `ESDEDB_NO_COLOR=true` |

### Beispiel

```bash
export ESDEDB_SDE_DIR=/data/sde-JSONL
export ESDEDB_DB_PATH=/data/eve-sde.db
export ESDEDB_WORKERS=8

esdedb import
```

## Weitere Ressourcen

### Architecture Decision Records

- [ADR-001: SQLite-Only Approach](../adr/ADR-001-sqlite-only-approach.md)
- [ADR-002: Database Layer Design](../adr/ADR-002-database-layer-design.md)
- [ADR-003: JSONL Parser Architecture](../adr/ADR-003-jsonl-parser-architecture.md)
- [ADR-004: Configuration Format](../adr/ADR-004-configuration-format.md)
- [ADR-005: Error Handling Strategy](../adr/ADR-005-error-handling-strategy.md)
- [ADR-006: Concurrency & Worker Pool](../adr/ADR-006-concurrency-worker-pool.md)

### Weitere Dokumentation

- [Version Command Details](../commands/version.md)
- [Project README](../../README.md)
- [Contributing Guidelines](../../.github/copilot-instructions.md)

## Support & Feedback

Bei Fragen, Problemen oder Feature-Wünschen:

1. **Issues**: [GitHub Issues](https://github.com/Sternrassler/EVE-SDE-Database-Builder/issues)
2. **Pull Requests**: Contributions willkommen (siehe Contributing Guidelines)
3. **Help**: `esdedb --help` oder `esdedb <command> --help`
