// Package main ist der Entry Point für das EVE SDE Database Builder CLI Toolpackage esdedb

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/logger"
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
	// Initialize logger
	logLevel := "info"
	logFormat := "json"

	log := logger.NewLogger(logLevel, logFormat)
	logger.SetGlobalLogger(log)

	// Log application start
	logger.LogAppStart(version, commit)

	// Setup signal handler for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.LogAppShutdown(fmt.Sprintf("received signal: %v", sig))
		os.Exit(0)
	}()

	rootCmd := &cobra.Command{
		Use:     "esdedb",
		Short:   "EVE SDE Database Builder - Import EVE Online Static Data Export to SQLite",
		Long:    `EVE SDE Database Builder (Go Edition) - CLI Tool für den Import von EVE Online SDE JSONL-Dateien in eine SQLite-Datenbank.`,
		Version: fmt.Sprintf("%s (commit: %s)", version, commit),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Adjust log level based on verbose flag
			if verbose {
				log := logger.NewLogger("debug", logFormat)
				logger.SetGlobalLogger(log)
			}
		},
	}

	// Persistent Flags (global für alle Commands)
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "./config.toml", "Pfad zur Konfigurationsdatei")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose Logging (Debug Level)")

	// Subcommands (werden später implementiert)
	// rootCmd.AddCommand(newImportCmd())
	// rootCmd.AddCommand(newConfigCmd())
	// rootCmd.AddCommand(newVersionCmd())

	if err := rootCmd.Execute(); err != nil {
		logger.LogAppShutdown(fmt.Sprintf("error: %v", err))
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Log normal shutdown
	logger.LogAppShutdown("normal exit")
}
