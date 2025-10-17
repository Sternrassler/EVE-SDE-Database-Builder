package main

import (
	"fmt"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/cli"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/config"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/logger"
	"github.com/spf13/cobra"
)

func newValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration file",
		Long: `Validate command prüft eine TOML-Konfigurationsdatei auf Gültigkeit.

Exit Codes:
  0 - Konfiguration ist gültig
  1 - Konfiguration ist ungültig oder Fehler beim Laden

Beispiel:
  esdedb validate --config ./config.toml
  esdedb validate -c /etc/esdedb/config.toml`,
		RunE: runValidateCmd,
	}

	return cmd
}

func runValidateCmd(cmd *cobra.Command, args []string) error {
	log := logger.GetGlobalLogger()

	log.Info("Validating configuration",
		logger.Field{Key: "config_path", Value: configPath},
	)

	// Load and validate config
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Error("Configuration validation failed",
			logger.Field{Key: "config_path", Value: configPath},
			logger.Field{Key: "error", Value: err.Error()},
		)
		cli.Error("Configuration validation failed: %v", err)
		return err
	}

	// If we get here, validation passed
	log.Info("Configuration validation successful",
		logger.Field{Key: "config_path", Value: configPath},
		logger.Field{Key: "database_path", Value: cfg.Database.Path},
		logger.Field{Key: "sde_path", Value: cfg.Import.SDEPath},
		logger.Field{Key: "workers", Value: cfg.Import.Workers},
		logger.Field{Key: "language", Value: cfg.Import.Language},
	)

	cli.Success("Configuration is valid")
	fmt.Printf("\nConfiguration Summary:\n")
	fmt.Printf("  Database Path: %s\n", cfg.Database.Path)
	fmt.Printf("  SDE Path:      %s\n", cfg.Import.SDEPath)
	fmt.Printf("  Workers:       %d\n", cfg.Import.Workers)
	fmt.Printf("  Language:      %s\n", cfg.Import.Language)
	fmt.Printf("  Log Level:     %s\n", cfg.Logging.Level)
	fmt.Printf("  Log Format:    %s\n", cfg.Logging.Format)

	return nil
}
