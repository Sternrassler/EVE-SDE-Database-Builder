// Package main ist der Entry Point für das EVE SDE Database Builder CLI Toolpackage esdedb

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version wird beim Build gesetzt (siehe Makefile)
	version = "dev"
	commit  = "unknown"
)

var (
	configPath string
	verbose    bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "esdedb",
		Short:   "EVE SDE Database Builder - Import EVE Online Static Data Export to SQLite",
		Long:    `EVE SDE Database Builder (Go Edition) - CLI Tool für den Import von EVE Online SDE JSONL-Dateien in eine SQLite-Datenbank.`,
		Version: fmt.Sprintf("%s (commit: %s)", version, commit),
	}

	// Persistent Flags (global für alle Commands)
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "./config.toml", "Pfad zur Konfigurationsdatei")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose Logging (Debug Level)")

	// Subcommands (werden später implementiert)
	// rootCmd.AddCommand(newImportCmd())
	// rootCmd.AddCommand(newConfigCmd())
	// rootCmd.AddCommand(newVersionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
