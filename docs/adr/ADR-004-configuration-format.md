# ADR-004: Configuration Format & Management

**Status:** Accepted  
**Datum:** 2025-10-15  
**Entscheider:** Migration Team  
**Kontext:** VB.NET → Go Migration für EVE SDE Database Builder  
**Abhängig von:** ADR-001 (SQLite-Only), ADR-002 (Database Layer)

---

## Kontext & Problem

### Anforderungen

Die Go-Anwendung benötigt Konfiguration für:

**Runtime-Parameter:**

- SQLite DB-Pfad (Output)
- SDE Source-Pfad (Input: JSONL-Verzeichnis)
- Thread/Worker Count (Concurrency)
- Import-Sprache (English, German, etc.)
- Logging-Level (Debug, Info, Warn, Error)

**Build/Deployment-Parameter:**

- Version
- Update-URL (optional, für späteres Auto-Update)

### VB.NET Status Quo

**Aktueller Ansatz:**

```vb
' ProgramSettings.vb (~321 Zeilen)
Public Class ProgramSettings
    Private Const AppSettingsFileName As String = "ApplicationSettings"
    Private Const XMLfileType As String = ".xml"
    
    Public Sub New()
        FullAppSettingsFileName = AppSettingsFileName & XMLfileType
    End Sub
    
    Private Function GetSettingValue(FileName As String, ...) As Object
        Dim m_xmld As New XmlDocument
        m_xmld.Load(FileName & XMLfileType)
        ' ... XPath-basiertes Auslesen ...
    End Function
End Class
```

**Beispiel `ApplicationSettings.xml`:**

```xml
<?xml version="1.0" encoding="utf-8"?>
<Settings>
  <SelectedDB>SQLite</SelectedDB>
  <SelectedLanguage>English</SelectedLanguage>
  <SQLiteDBPath>C:\EVE\Database.db</SQLiteDBPath>
  <SDEPath>C:\EVE\sde</SDEPath>
  <ThreadCount>4</ThreadCount>
</Settings>
```

**Charakteristik:**

- Custom XML-Parser (XmlDocument)
- ~321 Zeilen Boilerplate für XML-Handling
- Keine Validierung (alle Werte als `Object`)
- Type-Casting zur Runtime

### Herausforderung

**Welches Format für Go?**

1. **XML beibehalten?** → Kompatibilität mit VB.NET, aber ungewöhnlich in Go-Ökosystem
2. **YAML?** → Weit verbreitet (Kubernetes, Docker-Compose), aber externe Dependency
3. **TOML?** → Go-Community-Standard (z.B. `Cargo.toml`, `Hugo config`)
4. **JSON?** → Stdlib Support, aber weniger Human-Readable
5. **Environment Variables?** → 12-Factor App Prinzip, aber unübersichtlich bei vielen Parametern
6. **CLI Flags only?** → Einfach, aber keine Persistenz

---

## Entscheidung

Wir verwenden **TOML** für persistente Konfiguration + **CLI Flags** für Overrides + **Environment Variables** für Secrets.

**Hybridansatz:**

```txt
Priorität (niedrig → hoch):
1. TOML Config File (Defaults & User Preferences)
2. Environment Variables (Deployment-spezifisch)
3. CLI Flags (Session-Override)
```

### Architektur

**Beispiel `config.toml`:**

```toml
# EVE SDE Database Builder Configuration
version = "1.0.0"

[database]
path = "./eve_sde.db"
# SQLite-spezifische Pragmas (siehe ADR-002)
journal_mode = "WAL"
cache_size_mb = 64

[import]
sde_path = "./sde-JSONL"
language = "en"  # en, de, fr, ja, ru, zh, es, ko
workers = 4      # Concurrent JSONL Parsers

[logging]
level = "info"   # debug, info, warn, error
format = "json"  # json, text

[update]
# Optional: Auto-Update (für spätere Implementierung)
enabled = false
check_url = "https://example.com/latest-version.json"
```

**Go Implementation:**

```go
// internal/config/config.go
package config

import (
    "os"
    "github.com/BurntSushi/toml"
    "github.com/spf13/pflag"
)

type Config struct {
    Version  string
    Database DatabaseConfig `toml:"database"`
    Import   ImportConfig   `toml:"import"`
    Logging  LoggingConfig  `toml:"logging"`
    Update   UpdateConfig   `toml:"update"`
}

type DatabaseConfig struct {
    Path          string `toml:"path"`
    JournalMode   string `toml:"journal_mode"`
    CacheSizeMB   int    `toml:"cache_size_mb"`
}

type ImportConfig struct {
    SDEPath  string `toml:"sde_path"`
    Language string `toml:"language"`
    Workers  int    `toml:"workers"`
}

type LoggingConfig struct {
    Level  string `toml:"level"`
    Format string `toml:"format"`
}

type UpdateConfig struct {
    Enabled  bool   `toml:"enabled"`
    CheckURL string `toml:"check_url"`
}

// Load lädt Config mit Prioritäts-Kaskade
func Load(configPath string) (*Config, error) {
    // 1. TOML File laden (Defaults)
    var cfg Config
    if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
        if !os.IsNotExist(err) {
            return nil, err
        }
        // Fallback zu Defaults wenn keine Config-Datei
        cfg = DefaultConfig()
    }
    
    // 2. Environment Variables überschreiben
    if dbPath := os.Getenv("ESDEDB_DATABASE_PATH"); dbPath != "" {
        cfg.Database.Path = dbPath
    }
    if sdePath := os.Getenv("ESDEDB_SDE_PATH"); sdePath != "" {
        cfg.Import.SDEPath = sdePath
    }
    
    // 3. CLI Flags überschreiben (höchste Prio)
    applyFlags(&cfg)
    
    // 4. Validierung
    if err := cfg.Validate(); err != nil {
        return nil, err
    }
    
    return &cfg, nil
}

// Validate prüft Config-Constraints
func (c *Config) Validate() error {
    if c.Database.Path == "" {
        return fmt.Errorf("database.path is required")
    }
    if c.Import.SDEPath == "" {
        return fmt.Errorf("import.sde_path is required")
    }
    if c.Import.Workers < 1 || c.Import.Workers > 32 {
        return fmt.Errorf("import.workers must be between 1 and 32")
    }
    // Language validation
    validLangs := map[string]bool{
        "en": true, "de": true, "fr": true, "ja": true,
        "ru": true, "zh": true, "es": true, "ko": true,
    }
    if !validLangs[c.Import.Language] {
        return fmt.Errorf("invalid language: %s", c.Import.Language)
    }
    return nil
}

// DefaultConfig liefert sinnvolle Defaults
func DefaultConfig() Config {
    return Config{
        Version: "1.0.0",
        Database: DatabaseConfig{
            Path:        "./eve_sde.db",
            JournalMode: "WAL",
            CacheSizeMB: 64,
        },
        Import: ImportConfig{
            SDEPath:  "./sde-JSONL",
            Language: "en",
            Workers:  4,
        },
        Logging: LoggingConfig{
            Level:  "info",
            Format: "text",
        },
        Update: UpdateConfig{
            Enabled: false,
        },
    }
}
```

**CLI Integration (cobra/pflag):**

```go
// cmd/esdedb/main.go
package main

import (
    "github.com/spf13/cobra"
    "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/config"
)

var (
    configPath string
    dbPath     string
    sdePath    string
    workers    int
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "esdedb",
        Short: "EVE SDE Database Builder",
        RunE:  run,
    }
    
    // Persistent Flags (global)
    rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "./config.toml", "Config file path")
    
    // Import Command
    importCmd := &cobra.Command{
        Use:   "import",
        Short: "Import SDE data into SQLite",
        RunE:  runImport,
    }
    importCmd.Flags().StringVar(&dbPath, "db", "", "SQLite database path (overrides config)")
    importCmd.Flags().StringVar(&sdePath, "sde", "", "SDE JSONL directory (overrides config)")
    importCmd.Flags().IntVar(&workers, "workers", 0, "Worker count (overrides config)")
    
    rootCmd.AddCommand(importCmd)
    rootCmd.Execute()
}

func runImport(cmd *cobra.Command, args []string) error {
    // Config laden
    cfg, err := config.Load(configPath)
    if err != nil {
        return err
    }
    
    // CLI Flags überschreiben (höchste Prio)
    if dbPath != "" {
        cfg.Database.Path = dbPath
    }
    if sdePath != "" {
        cfg.Import.SDEPath = sdePath
    }
    if workers > 0 {
        cfg.Import.Workers = workers
    }
    
    // Import ausführen
    return doImport(cfg)
}
```

### Begründung

**Warum TOML?**

| Kriterium | XML | YAML | TOML | JSON | Env Vars | Bewertung |
|-----------|-----|------|------|------|----------|-----------|
| Human-Readable | ⚠️ Verbose | ✅ Gut | ✅ Sehr gut | ⚠️ OK | ❌ Schlecht (bei >10 Vars) | **TOML gewinnt** |
| Stdlib Support | ❌ Nein (encoding/xml unergonomisch) | ❌ Nein | ❌ Nein | ✅ Ja | ✅ Ja | Neutral (alle brauchen Libs) |
| Go-Community | ❌ Unüblich | ✅ Üblich | ✅ Standard | ✅ Üblich | ✅ 12-Factor | **TOML/YAML/Env gleichwertig** |
| Type Safety | ⚠️ Strings | ⚠️ Implicit Typing | ✅ Explicit | ✅ Explicit | ❌ Nur Strings | **TOML gewinnt** |
| Comments | ✅ Ja | ✅ Ja | ✅ Ja | ❌ Nein | ❌ Nein | TOML/YAML/XML |
| Nesting | ✅ Ja | ✅ Ja (zu flexibel) | ✅ Ja (Tables) | ✅ Ja | ⚠️ Umständlich | **TOML gewinnt** (Struktur ohne YAML-Chaos) |
| VB.NET Migration | ✅ Identisch | ❌ Neu | ❌ Neu | ❌ Neu | ❌ Neu | Irrelevant (Clean Slate) |

**Warum Hybrid (TOML + Env + Flags)?**

- ✅ **TOML:** User-freundlich für lokale Dev/Testing
- ✅ **Env Vars:** Container/Cloud-Deployment (12-Factor App)
- ✅ **CLI Flags:** Quick Overrides ohne Config-Edit

---

## Konsequenzen

### Positive Konsequenzen

1. **Einfachheit:** TOML ist kompakt & lesbar (vs. XML's 321 Zeilen Boilerplate)
2. **Type Safety:** Struct Tags → Compile-Zeit-Validierung
3. **Flexibilität:** 3-Stufen-Override (TOML < Env < Flags)
4. **Go-Idiomatisch:** Standard in Go-Community (Hugo, Grafana, etc.)
5. **Testing:** In-Memory Config via Struct-Literals (kein File I/O nötig)
6. **12-Factor Compliant:** Env Vars für Container-Deployments

### Negative Konsequenzen

1. **Breaking Change:** VB.NET User müssen von XML zu TOML migrieren
2. **Library Dependency:** `github.com/BurntSushi/toml` (statt Stdlib)
3. **Migration Effort:** Bestehende XML-Configs müssen konvertiert werden

### Mitigationen

| Konsequenz | Mitigation |
|------------|------------|
| Breaking Change | Migration-Tool `convert-config.sh` (XML → TOML) |
| Dependency | BurntSushi/toml ist mature & stabil (seit 2013) |
| Migration | Beispiel-Configs + Dokumentation im README |

---

## Alternativen (erwogen & verworfen)

### Alternative 1: XML beibehalten

**Pro:**

- ✅ Kompatibel mit VB.NET
- ✅ Keine User-Migration nötig

**Contra:**

- ❌ Unüblich in Go-Ökosystem
- ❌ Verbose (viel Boilerplate)
- ❌ `encoding/xml` ist unergonomisch (Struct Tags komplex)

**Entscheidung:** Verworfen (Go-Idiomatik wichtiger als VB.NET-Kompatibilität)

### Alternative 2: Pure YAML

**Pro:**

- ✅ Weit verbreitet (Kubernetes, Docker-Compose)
- ✅ Flexibel

**Contra:**

- ⚠️ Zu flexibel (Indentation-Fehler, implizite Typen)
- ⚠️ Keine Multi-Line-Strings ohne Tricks
- ⚠️ Whitespace-sensitiv → fehleranfällig

**Entscheidung:** Verworfen (TOML ist klarer & sicherer)

### Alternative 3: Pure Environment Variables

**Pro:**

- ✅ 12-Factor App Prinzip
- ✅ Container-freundlich
- ✅ Keine Config-Datei nötig

**Contra:**

- ❌ Unübersichtlich bei >10 Variablen
- ❌ Keine Nested Structures (nur Flat Keys)
- ❌ Schwierig für lokale Entwicklung

**Entscheidung:** Verworfen als **alleinige** Lösung, aber **inkludiert** im Hybrid-Ansatz

### Alternative 4: Pure CLI Flags

**Pro:**

- ✅ Explizit & transparent
- ✅ `--help` dokumentiert automatisch

**Contra:**

- ❌ Keine Persistenz (User muss immer alles angeben)
- ❌ Unhandlich bei vielen Parametern

**Entscheidung:** Verworfen als **alleinige** Lösung, aber **inkludiert** für Overrides

### Alternative 5: JSON

**Pro:**

- ✅ Stdlib Support (`encoding/json`)
- ✅ Weit verbreitet

**Contra:**

- ❌ Keine Comments (→ keine Inline-Doku)
- ❌ Trailing Commas = Parse Error (fehleranfällig)
- ❌ Weniger Human-Readable als TOML

**Entscheidung:** Verworfen (TOML ist User-freundlicher)

---

## Implementierungsdetails

### 1. Config-File Hierarchie

```
1. /etc/esdedb/config.toml        (System-wide)
2. ~/.config/esdedb/config.toml   (User-specific)
3. ./config.toml                  (Working Directory)
4. --config <path>                (CLI Override)
```

**Ladereihenfolge:** 1 → 2 → 3 → 4 (spätere überschreiben frühere)

### 2. Environment Variable Mapping

```bash
# Konvention: ESDEDB_<SECTION>_<KEY>
export ESDEDB_DATABASE_PATH="/var/lib/esdedb/eve.db"
export ESDEDB_IMPORT_SDE_PATH="/mnt/sde"
export ESDEDB_IMPORT_WORKERS="8"
export ESDEDB_LOGGING_LEVEL="debug"
```

**Implementierung:**

```go
func applyEnvVars(cfg *Config) {
    if v := os.Getenv("ESDEDB_DATABASE_PATH"); v != "" {
        cfg.Database.Path = v
    }
    if v := os.Getenv("ESDEDB_IMPORT_SDE_PATH"); v != "" {
        cfg.Import.SDEPath = v
    }
    if v := os.Getenv("ESDEDB_IMPORT_WORKERS"); v != "" {
        if workers, err := strconv.Atoi(v); err == nil {
            cfg.Import.Workers = workers
        }
    }
    // ... weitere Mappings
}
```

### 3. Config Generation

```bash
# Initial Config erstellen
esdedb config init --output ./config.toml

# Mit Custom Werten
esdedb config init \
    --db-path ./custom.db \
    --sde-path /opt/sde \
    --workers 8
```

**Implementierung:**

```go
func generateConfig(output string, overrides map[string]interface{}) error {
    cfg := config.DefaultConfig()
    
    // Overrides anwenden
    if dbPath, ok := overrides["db-path"].(string); ok {
        cfg.Database.Path = dbPath
    }
    // ... weitere Overrides
    
    // TOML schreiben
    f, _ := os.Create(output)
    defer f.Close()
    return toml.NewEncoder(f).Encode(cfg)
}
```

### 4. Migration Tool (XML → TOML)

```bash
# Convert VB.NET XML Config
esdedb config convert \
    --input ApplicationSettings.xml \
    --output config.toml
```

**Beispiel-Konvertierung:**

```xml
<!-- ApplicationSettings.xml -->
<Settings>
  <SQLiteDBPath>C:\EVE\Database.db</SQLiteDBPath>
  <SDEPath>C:\EVE\sde</SDEPath>
  <ThreadCount>4</ThreadCount>
</Settings>
```

→

```toml
# config.toml
[database]
path = "C:/EVE/Database.db"  # Windows-Pfad normalisiert

[import]
sde_path = "C:/EVE/sde"
workers = 4
language = "en"  # Default (nicht in XML vorhanden)
```

### 5. Testing-Strategie

```go
// internal/config/config_test.go
func TestLoad_DefaultConfig(t *testing.T) {
    // In-Memory Config (kein File I/O)
    cfg := config.DefaultConfig()
    
    assert.Equal(t, "./eve_sde.db", cfg.Database.Path)
    assert.Equal(t, 4, cfg.Import.Workers)
}

func TestLoad_EnvOverride(t *testing.T) {
    os.Setenv("ESDEDB_DATABASE_PATH", "/tmp/test.db")
    defer os.Unsetenv("ESDEDB_DATABASE_PATH")
    
    cfg, _ := config.Load("testdata/config.toml")
    
    assert.Equal(t, "/tmp/test.db", cfg.Database.Path)
}

func TestValidate_InvalidWorkers(t *testing.T) {
    cfg := config.DefaultConfig()
    cfg.Import.Workers = 100  // > 32
    
    err := cfg.Validate()
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "workers must be between")
}
```

---

## Migration von VB.NET

### Code-Vergleich

**VB.NET (XML):**

```vb
' ProgramSettings.vb
Private Function GetSettingValue(...) As Object
    Dim m_xmld As New XmlDocument
    m_xmld.Load(FileName & XMLfileType)
    m_nodelist = m_xmld.SelectNodes("/" & RootElement & "/" & ElementString)
    
    If Not IsNothing(m_nodelist.Item(0)) Then
        TempValue = m_nodelist.Item(0).InnerText
        ' ... Type Casting zur Runtime ...
        Select Case ObjectType
            Case SettingTypes.TypeBoolean
                Return CBool(TempValue)
            Case SettingTypes.TypeString
                Return TempValue
        End Select
    End If
End Function
```

**Go (TOML):**

```go
// internal/config/config.go
var cfg Config
toml.DecodeFile("config.toml", &cfg)  // One-Liner!

// Type Safety zur Compile-Zeit
dbPath := cfg.Database.Path  // string
workers := cfg.Import.Workers  // int
```

**Unterschiede:**

- ✅ Go: Deklarativ (Struct Tags), VB.NET: Imperativ (XPath + Casting)
- ✅ Go: Type-Safe, VB.NET: Runtime-Casting mit `Object`
- ✅ Go: ~50 Zeilen, VB.NET: ~321 Zeilen

---

## Compliance & Governance

### Normative Anforderungen

- ✅ **MUST:** Keine Hardcoded Secrets → DB-Pfad aus Config/Env
- ✅ **SHOULD:** Validierung → `Validate()` Methode
- ✅ **MAY:** Auto-Update Support → Config-Sektion vorbereitet

### ADR-Abhängigkeiten

- **ADR-001:** SQLite-Only → Vereinfacht Config (keine Multi-DB-Switches)
- **ADR-002:** Database Layer → DB-Pfad aus Config gelesen

---

## Referenzen

**Libraries:**

- [BurntSushi/toml](https://github.com/BurntSushi/toml) - TOML Parser (v1.3+)
- [spf13/cobra](https://github.com/spf13/cobra) - CLI Framework
- [spf13/pflag](https://github.com/spf13/pflag) - POSIX-style Flags
- [spf13/viper](https://github.com/spf13/viper) - Alternative (Multi-Format Config Library)

**Standards:**

- [TOML Spec v1.0.0](https://toml.io/en/v1.0.0)
- [12-Factor App Config](https://12factor.net/config)

**Alternativen (nicht gewählt):**

- [gopkg.in/yaml.v3](https://github.com/go-yaml/yaml) - YAML Parser
- [encoding/json](https://pkg.go.dev/encoding/json) - JSON (Stdlib)

---

## Änderungshistorie

| Datum | Version | Änderung | Autor |
|-------|---------|----------|-------|
| 2025-10-15 | 0.1.0 | Initial Draft | AI Copilot |
| 2025-10-15 | 1.0.0 | Status → Accepted (TOML + Env + Flags Hybrid) | Migration Team |

---

**Nächste Schritte:**

1. ✅ ~~Review durch Team~~ (Accepted)
2. `internal/config/config.go` implementieren
3. `esdedb config init` Command implementieren
4. `convert-config.sh` Tool (XML → TOML) erstellen
5. Beispiel-Configs in `examples/` ablegen
6. ✅ ~~Bei Erfolg: Status → `Accepted`~~
