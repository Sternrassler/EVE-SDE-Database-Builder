package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// newConfigCmd erstellt das config Command mit Subcommands
func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management commands",
		Long: `Config command bietet Verwaltungsfunktionen für TOML-Konfigurationsdateien.

Subcommands:
  init    - Generiert eine neue config.toml mit Standard-Einstellungen
  convert - Konvertiert alte VB.NET XML-Config zu TOML-Format

Die Konfigurationsdatei steuert:
  - Datenbank-Einstellungen (Pfad, SQLite-Pragmas)
  - Import-Einstellungen (SDE-Verzeichnis, Worker-Count, Sprache)
  - Logging-Einstellungen (Level, Format)`,
		Example: `  # Neue Konfiguration erstellen
  esdedb config init

  # Neue Konfiguration mit benutzerdefiniertem Pfad
  esdedb config init --output ./my-config.toml

  # VB.NET XML-Config zu TOML konvertieren
  esdedb config convert --input ApplicationSettings.xml --output config.toml`,
	}

	// Subcommands hinzufügen
	cmd.AddCommand(newConfigInitCmd())
	cmd.AddCommand(newConfigConvertCmd())

	return cmd
}

// newConfigInitCmd erstellt das 'config init' Subcommand
func newConfigInitCmd() *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generate a new config.toml with default settings",
		Long: `Init command generiert eine neue TOML-Konfigurationsdatei mit Standard-Einstellungen.

Die generierte Datei enthält:
  - Standard-Pfade für Datenbank und SDE-Verzeichnis
  - Empfohlene SQLite-Pragmas (WAL-Modus, Cache-Größe)
  - Standard-Import-Einstellungen (Worker-Count, Sprache)
  - Logging-Konfiguration (Level, Format)

Alle Werte können nachträglich angepasst werden.`,
		Example: `  # Standard config.toml erstellen (im aktuellen Verzeichnis)
  esdedb config init

  # Config in benutzerdefiniertem Pfad erstellen
  esdedb config init --output ./configs/production.toml

  # Bestehende Datei überschreiben (mit Warnung)
  esdedb config init --output ./config.toml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigInit(outputPath)
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "./config.toml", "Ausgabe-Pfad für config.toml")

	return cmd
}

// newConfigConvertCmd erstellt das 'config convert' Subcommand
func newConfigConvertCmd() *cobra.Command {
	var inputPath string
	var outputPath string

	cmd := &cobra.Command{
		Use:   "convert",
		Short: "Convert VB.NET XML config to TOML format",
		Long: `Convert command konvertiert alte VB.NET XML-Konfigurationen zu TOML-Format.

Dieser Befehl ist für die Migration von der Legacy VB.NET Version gedacht.
Er liest eine ApplicationSettings.xml und konvertiert die bekannten Felder:
  - SelectedDB → Ignoriert (Go-Version nur SQLite)
  - SelectedLanguage → import.language
  - SQLiteDBPath → database.path
  - SDEPath → import.sde_path
  - ThreadCount → import.workers

Neue Felder (z.B. logging.level) werden mit Standardwerten gefüllt.`,
		Example: `  # XML zu TOML konvertieren
  esdedb config convert --input ApplicationSettings.xml --output config.toml

  # Alte Config überschreiben
  esdedb config convert -i old.xml -o config.toml

  # Mit Verbose Logging für Debug-Ausgaben
  esdedb --verbose config convert -i ApplicationSettings.xml -o config.toml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigConvert(inputPath, outputPath)
		},
	}

	cmd.Flags().StringVarP(&inputPath, "input", "i", "ApplicationSettings.xml", "Pfad zur XML-Konfigurationsdatei")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "./config.toml", "Ausgabe-Pfad für config.toml")

	return cmd
}

// runConfigInit generiert eine neue config.toml mit Standardwerten
func runConfigInit(outputPath string) error {
	// Prüfen ob Datei bereits existiert
	if _, err := os.Stat(outputPath); err == nil {
		fmt.Printf("⚠️  Warning: File %s already exists. Overwriting...\n", outputPath)
	}

	// Standard TOML Content
	defaultConfig := `# EVE SDE Database Builder Configuration
version = "1.0.0"

[database]
path = "./eve_sde.db"
journal_mode = "WAL"
cache_size_mb = 64

[import]
sde_path = "./sde-JSONL"
language = "en"  # en, de, fr, ja, ru, zh, es, ko
workers = 4      # 0 = auto (runtime.NumCPU())

[logging]
level = "info"   # debug, info, warn, error
format = "text"  # text, json

[update]
enabled = false
check_url = ""
`

	// Verzeichnis erstellen falls nicht vorhanden
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Datei schreiben
	if err := os.WriteFile(outputPath, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✅ Successfully created config.toml at %s\n", outputPath)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Edit %s to customize settings\n", outputPath)
	fmt.Printf("  2. Validate config: esdedb validate --config %s\n", outputPath)
	fmt.Printf("  3. Run import: esdedb import --config %s\n", outputPath)

	return nil
}

// XMLSettings repräsentiert die VB.NET XML-Konfiguration
type XMLSettings struct {
	XMLName          xml.Name `xml:"Settings"`
	SelectedDB       string   `xml:"SelectedDB"`
	SelectedLanguage string   `xml:"SelectedLanguage"`
	SQLiteDBPath     string   `xml:"SQLiteDBPath"`
	SDEPath          string   `xml:"SDEPath"`
	ThreadCount      int      `xml:"ThreadCount"`
}

// runConfigConvert konvertiert XML zu TOML
func runConfigConvert(inputPath, outputPath string) error {
	// Prüfen ob Input existiert
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputPath)
	}

	// XML lesen
	xmlData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read XML file: %w", err)
	}

	// XML parsen
	var settings XMLSettings
	if err := xml.Unmarshal(xmlData, &settings); err != nil {
		return fmt.Errorf("failed to parse XML: %w", err)
	}

	// Sprache konvertieren (English → en, etc.)
	language := convertLanguage(settings.SelectedLanguage)

	// Default-Werte setzen falls nicht vorhanden
	dbPath := settings.SQLiteDBPath
	if dbPath == "" {
		dbPath = "./eve_sde.db"
	}

	sdePath := settings.SDEPath
	if sdePath == "" {
		sdePath = "./sde-JSONL"
	}

	workers := settings.ThreadCount
	if workers <= 0 {
		workers = 4
	}

	// TOML Content generieren
	tomlContent := fmt.Sprintf(`# EVE SDE Database Builder Configuration
# Converted from VB.NET XML config: %s
version = "1.0.0"

[database]
path = "%s"
journal_mode = "WAL"
cache_size_mb = 64

[import]
sde_path = "%s"
language = "%s"
workers = %d

[logging]
level = "info"   # debug, info, warn, error
format = "text"  # text, json

[update]
enabled = false
check_url = ""
`, inputPath, dbPath, sdePath, language, workers)

	// Output-Verzeichnis erstellen falls nicht vorhanden
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// TOML-Datei schreiben
	if err := os.WriteFile(outputPath, []byte(tomlContent), 0644); err != nil {
		return fmt.Errorf("failed to write TOML file: %w", err)
	}

	fmt.Printf("✅ Successfully converted XML to TOML\n")
	fmt.Printf("\nConversion Summary:\n")
	fmt.Printf("  Input:      %s\n", inputPath)
	fmt.Printf("  Output:     %s\n", outputPath)
	fmt.Printf("  Database:   %s\n", dbPath)
	fmt.Printf("  SDE Path:   %s\n", sdePath)
	fmt.Printf("  Language:   %s (from %s)\n", language, settings.SelectedLanguage)
	fmt.Printf("  Workers:    %d\n", workers)
	fmt.Printf("\nNote: SelectedDB field (%s) was ignored (Go version is SQLite-only)\n", settings.SelectedDB)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Review %s for correctness\n", outputPath)
	fmt.Printf("  2. Validate: esdedb validate --config %s\n", outputPath)

	return nil
}

// convertLanguage konvertiert VB.NET Sprach-Namen zu ISO-Codes
func convertLanguage(vbLang string) string {
	languageMap := map[string]string{
		"English":  "en",
		"German":   "de",
		"French":   "fr",
		"Japanese": "ja",
		"Russian":  "ru",
		"Chinese":  "zh",
		"Spanish":  "es",
		"Korean":   "ko",
	}

	if code, ok := languageMap[vbLang]; ok {
		return code
	}

	// Fallback: lowercase first 2 chars oder default "en"
	if len(vbLang) >= 2 {
		return vbLang[:2]
	}
	return "en"
}
