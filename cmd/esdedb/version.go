package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// buildTime wird beim Build gesetzt (siehe Makefile)
	buildTime = "unknown"
)

// VersionInfo enthält alle Versionsinformationen
type VersionInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
}

func newVersionCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long: `Version command zeigt erweiterte Versionsinformationen an.

Folgende Informationen werden angezeigt:
  - Version (aus VERSION Datei oder Build-Zeit)
  - Commit Hash (Git SHA)
  - Build Time (Zeitpunkt des Builds)

Die Ausgabe kann in verschiedenen Formaten erfolgen:
  - text (Standard): Menschenlesbare Ausgabe
  - json: JSON-Format für maschinelle Verarbeitung`,
		Example: `  # Standard Ausgabe (Text-Format)
  esdedb version

  # JSON-Format für maschinelle Verarbeitung
  esdedb version --format json

  # Kurze Version-Info via Root-Command
  esdedb --version`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVersionCmd(cmd, args, format)
		},
	}

	// Flags für Format-Option
	cmd.Flags().StringVar(&format, "format", "text", "Ausgabeformat (text oder json)")

	return cmd
}

func runVersionCmd(cmd *cobra.Command, args []string, format string) error {
	info := VersionInfo{
		Version:   version,
		Commit:    commit,
		BuildTime: buildTime,
	}

	switch format {
	case "json":
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(info)
	case "text":
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Version:    %s\n", info.Version)
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Commit:     %s\n", info.Commit)
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Build Time: %s\n", info.BuildTime)
		return nil
	default:
		return fmt.Errorf("unsupported format: %s (use 'text' or 'json')", format)
	}
}
