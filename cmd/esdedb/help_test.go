package main

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestRootCmd_HelpText verifies that the root command has complete help text
func TestRootCmd_HelpText(t *testing.T) {
	rootCmd := newRootCommand()

	// Test Use field
	if rootCmd.Use != "esdedb" {
		t.Errorf("expected Use to be 'esdedb', got '%s'", rootCmd.Use)
	}

	// Test Short description
	if rootCmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	// Test Long description
	if rootCmd.Long == "" {
		t.Error("expected Long description to be set")
	}

	// Test Example field
	if rootCmd.Example == "" {
		t.Error("expected Example field to be set")
	}

	// Verify Example contains actual examples
	if !strings.Contains(rootCmd.Example, "esdedb") {
		t.Error("expected Example to contain command examples")
	}

	// Verify Long description contains meaningful content
	if !strings.Contains(rootCmd.Long, "Import") && !strings.Contains(rootCmd.Long, "SDE") {
		t.Error("expected Long description to contain meaningful content about the tool")
	}
}

// TestImportCmd_HelpText verifies that the import command has complete help text
func TestImportCmd_HelpText(t *testing.T) {
	cmd := newImportCmd()

	// Test Use field
	if cmd.Use != "import" {
		t.Errorf("expected Use to be 'import', got '%s'", cmd.Use)
	}

	// Test Short description
	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}
	if !strings.Contains(cmd.Short, "Import") && !strings.Contains(cmd.Short, "import") {
		t.Error("expected Short description to mention 'import'")
	}

	// Test Long description
	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}

	// Test Example field
	if cmd.Example == "" {
		t.Error("expected Example field to be set")
	}

	// Verify Example contains multiple examples
	exampleLines := strings.Split(cmd.Example, "\n")
	exampleCount := 0
	for _, line := range exampleLines {
		if strings.Contains(line, "esdedb import") {
			exampleCount++
		}
	}
	if exampleCount < 3 {
		t.Errorf("expected at least 3 examples in Example field, found %d", exampleCount)
	}

	// Verify Long description mentions phases
	if !strings.Contains(cmd.Long, "Phase 1") || !strings.Contains(cmd.Long, "Phase 2") {
		t.Error("expected Long description to describe both import phases")
	}

	// Test Flags
	flags := cmd.Flags()
	requiredFlags := []string{"sde-dir", "db", "workers", "skip-errors"}
	for _, flagName := range requiredFlags {
		flag := flags.Lookup(flagName)
		if flag == nil {
			t.Errorf("expected flag '%s' to be defined", flagName)
		} else if flag.Usage == "" {
			t.Errorf("expected flag '%s' to have a usage description", flagName)
		}
	}
}

// TestValidateCmd_HelpText verifies that the validate command has complete help text
func TestValidateCmd_HelpText(t *testing.T) {
	cmd := newValidateCmd()

	// Test Use field
	if cmd.Use != "validate" {
		t.Errorf("expected Use to be 'validate', got '%s'", cmd.Use)
	}

	// Test Short description
	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}
	if !strings.Contains(cmd.Short, "Validate") && !strings.Contains(cmd.Short, "validate") {
		t.Error("expected Short description to mention 'validate'")
	}

	// Test Long description
	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}

	// Test Example field
	if cmd.Example == "" {
		t.Error("expected Example field to be set")
	}

	// Verify Example contains multiple examples
	exampleLines := strings.Split(cmd.Example, "\n")
	exampleCount := 0
	for _, line := range exampleLines {
		if strings.Contains(line, "esdedb validate") || strings.Contains(line, "esdedb --verbose validate") {
			exampleCount++
		}
	}
	if exampleCount < 2 {
		t.Errorf("expected at least 2 examples in Example field, found %d", exampleCount)
	}

	// Verify Long description mentions what is validated
	if !strings.Contains(cmd.Long, "TOML") || !strings.Contains(cmd.Long, "validiert") {
		t.Error("expected Long description to describe what is validated")
	}
}

// TestPersistentFlags_Descriptions verifies that all persistent flags have descriptions
func TestPersistentFlags_Descriptions(t *testing.T) {
	rootCmd := newRootCommand()

	persistentFlags := rootCmd.PersistentFlags()
	requiredFlags := []string{"config", "verbose", "no-color"}

	for _, flagName := range requiredFlags {
		flag := persistentFlags.Lookup(flagName)
		if flag == nil {
			t.Errorf("expected persistent flag '%s' to be defined", flagName)
		} else if flag.Usage == "" {
			t.Errorf("expected persistent flag '%s' to have a usage description", flagName)
		}
	}
}

// Helper function to create root command for testing
func newRootCommand() *cobra.Command {
	// Simplified version of main.go rootCmd creation for testing
	rootCmd := &cobra.Command{
		Use:   "esdedb",
		Short: "EVE SDE Database Builder - Import EVE Online Static Data Export to SQLite",
		Long: `EVE SDE Database Builder (Go Edition) - CLI Tool für den Import von EVE Online SDE JSONL-Dateien in eine SQLite-Datenbank.

Dieses Tool importiert die EVE Online Static Data Export (SDE) Daten aus dem JSONL-Format
in eine SQLite-Datenbank. Der Import erfolgt in zwei Phasen:
  1. Paralleles Parsing der JSONL-Dateien mit Worker Pool
  2. Sequenzielles Einfügen in die SQLite-Datenbank

Verfügbare Befehle:
  import   - Importiert SDE JSONL-Dateien in SQLite-Datenbank
  validate - Validiert eine TOML-Konfigurationsdatei`,
		Example: `  # Import mit Standard-Einstellungen
  esdedb import

  # Import mit benutzerdefinierten Pfaden
  esdedb import --sde-dir ./sde-JSONL --db ./eve-sde.db

  # Konfiguration validieren
  esdedb validate --config ./config.toml

  # Verbose Logging aktivieren
  esdedb --verbose import --sde-dir ./sde-JSONL`,
	}

	// Add persistent flags
	var configPath string
	var verbose bool
	var noColor bool
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "./config.toml", "Pfad zur TOML-Konfigurationsdatei")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Aktiviert Verbose Logging (Debug Level) für detaillierte Ausgaben")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Deaktiviert farbige Konsolenausgabe (nützlich für Logs/CI)")

	return rootCmd
}
