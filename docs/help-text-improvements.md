# Help Text Verbesserungen - Dokumentation

## Übersicht

Dieses Dokument zeigt die Verbesserungen am Help-Text für alle Cobra Commands im EVE SDE Database Builder.

## Änderungen

### 1. Root Command (`esdedb`)

**Neu hinzugefügt:**
- Erweiterte `Long` Beschreibung mit Details zum 2-Phasen-Import
- `Example` Feld mit 4 praktischen Beispielen
- Verbesserte Flag-Beschreibungen

**Beispiele:**
```bash
# Import mit Standard-Einstellungen
esdedb import

# Import mit benutzerdefinierten Pfaden
esdedb import --sde-dir ./sde-JSONL --db ./eve-sde.db

# Konfiguration validieren
esdedb validate --config ./config.toml

# Verbose Logging aktivieren
esdedb --verbose import --sde-dir ./sde-JSONL
```

### 2. Import Command (`esdedb import`)

**Neu hinzugefügt:**
- Detaillierte Beschreibung der beiden Import-Phasen
- Erklärung des Fortschrittsbalkens und Live-Metriken
- `Example` Feld mit 5 praktischen Beispielen
- Verbesserte Flag-Beschreibungen

**Beispiele:**
```bash
# Import mit Standard-Einstellungen (4 Workers)
esdedb import

# Import mit benutzerdefiniertem SDE-Verzeichnis und Datenbank-Pfad
esdedb import --sde-dir ./sde-JSONL --db ./eve-sde.db --workers 4

# Automatische Worker-Anzahl basierend auf CPU-Kernen
esdedb import --sde-dir ./sde-JSONL --db ./eve-sde.db --workers -1

# Fehlerhafte Dateien überspringen und Import fortsetzen
esdedb import --sde-dir ./sde-JSONL --db ./eve-sde.db --skip-errors

# Import mit Verbose Logging (Debug-Level)
esdedb --verbose import --sde-dir ./sde-JSONL
```

### 3. Validate Command (`esdedb validate`)

**Neu hinzugefügt:**
- Detaillierte Liste der Validierungsaspekte
- `Example` Feld mit 4 praktischen Beispielen
- Klarere Beschreibung des Validierungsprozesses

**Beispiele:**
```bash
# Konfigurationsdatei validieren (Standard: ./config.toml)
esdedb validate

# Spezifische Konfigurationsdatei validieren
esdedb validate --config ./config.toml

# Produktions-Konfiguration validieren
esdedb validate -c /etc/esdedb/config.toml

# Mit Verbose Logging für detaillierte Ausgabe
esdedb --verbose validate --config ./config.toml
```

## Flag-Beschreibungen

### Globale Flags (Persistent)
- `--config, -c`: Pfad zur TOML-Konfigurationsdatei
- `--verbose, -v`: Aktiviert Verbose Logging (Debug Level) für detaillierte Ausgaben
- `--no-color`: Deaktiviert farbige Konsolenausgabe (nützlich für Logs/CI)

### Import Command Flags
- `--sde-dir, -s`: Pfad zum Verzeichnis mit SDE JSONL-Dateien
- `--db, -d`: Pfad zur SQLite-Datenbank (wird erstellt falls nicht vorhanden)
- `--workers, -w`: Anzahl paralleler Worker-Threads (-1 = Automatisch basierend auf CPU-Kernen)
- `--skip-errors`: Überspringt fehlerhafte Dateien statt Import abzubrechen

## Tests

Neue Tests in `cmd/esdedb/help_test.go`:
- `TestRootCmd_HelpText`: Verifiziert Root Command Help-Text
- `TestImportCmd_HelpText`: Verifiziert Import Command Help-Text
- `TestValidateCmd_HelpText`: Verifiziert Validate Command Help-Text
- `TestPersistentFlags_Descriptions`: Verifiziert Global Flag Beschreibungen

Alle Tests prüfen:
- Vorhandensein von `Use`, `Short`, `Long` und `Example` Feldern
- Mindestanzahl von Beispielen im `Example` Feld
- Inhaltliche Qualität der Beschreibungen
- Vollständigkeit der Flag-Beschreibungen

## Verifizierung

Um die verbesserten Help-Texte zu sehen:

```bash
# Root Command
esdedb --help

# Import Command
esdedb help import
esdedb import --help

# Validate Command
esdedb help validate
esdedb validate --help
```

## Definition of Done

✅ Alle Akzeptanzkriterien erfüllt:
- [x] Cobra Long Description für alle Commands
- [x] Beispiele für jeden Command (via Example-Feld)
- [x] Flag-Beschreibungen vollständig und aussagekräftig
- [x] `esdedb help <command>` vollständig und informativ
- [x] Tests validieren Help-Text-Qualität
