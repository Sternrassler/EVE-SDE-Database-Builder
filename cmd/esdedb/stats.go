package main

import (
	"fmt"
	"os"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/database"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/logger"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

var (
	statsDBPath string
)

// TableStats enthält Statistiken für eine einzelne Tabelle
type TableStats struct {
	Name     string
	RowCount int64
}

// DatabaseStats enthält Gesamtstatistiken der Datenbank
type DatabaseStats struct {
	Tables    []TableStats
	DBSize    int64
	TotalRows int64
}

func newStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show database statistics",
		Long: `Stats command zeigt Statistiken über die SQLite-Datenbank an.

Folgende Informationen werden angezeigt:
  - Liste aller Tabellen in der Datenbank
  - Anzahl der Zeilen pro Tabelle
  - Gesamtgröße der Datenbankdatei

Die Statistiken helfen bei der Überwachung des Datenbank-Inhalts
und können zur Diagnose von Import-Problemen verwendet werden.`,
		Example: `  # Statistiken für Standard-Datenbank anzeigen
  esdedb stats --db ./eve-sde.db

  # Statistiken für benutzerdefinierte Datenbank
  esdedb stats --db /path/to/custom.db

  # Mit Verbose Logging
  esdedb --verbose stats --db ./eve-sde.db`,
		RunE: runStatsCmd,
	}

	// Flags
	cmd.Flags().StringVarP(&statsDBPath, "db", "d", "./eve-sde.db", "Pfad zur SQLite-Datenbank")

	return cmd
}

func runStatsCmd(cmd *cobra.Command, args []string) error {
	log := logger.GetGlobalLogger()

	// Validate inputs
	if statsDBPath == "" {
		return fmt.Errorf("--db darf nicht leer sein")
	}

	// Check if database file exists
	if _, err := os.Stat(statsDBPath); os.IsNotExist(err) {
		return fmt.Errorf("database file does not exist: %s", statsDBPath)
	}

	log.Info("Collecting database statistics",
		logger.Field{Key: "db_path", Value: statsDBPath},
	)

	// Open Database (read-only)
	db, err := database.NewDB(statsDBPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Collect statistics
	stats, err := collectStats(db)
	if err != nil {
		return fmt.Errorf("failed to collect statistics: %w", err)
	}

	// Display statistics
	displayStats(stats, statsDBPath)

	log.Info("Statistics collection completed",
		logger.Field{Key: "table_count", Value: len(stats.Tables)},
		logger.Field{Key: "total_rows", Value: stats.TotalRows},
		logger.Field{Key: "db_size_bytes", Value: stats.DBSize},
	)

	return nil
}

func collectStats(db *sqlx.DB) (*DatabaseStats, error) {
	stats := &DatabaseStats{
		Tables: make([]TableStats, 0),
	}

	// Get database file size
	fileInfo, err := os.Stat(statsDBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	stats.DBSize = fileInfo.Size()

	// Get list of all tables
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tableNames []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tableNames = append(tableNames, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating table names: %w", err)
	}

	// Get row count for each table
	for _, tableName := range tableNames {
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
		var count int64
		if err := db.QueryRow(countQuery).Scan(&count); err != nil {
			return nil, fmt.Errorf("failed to count rows in table %s: %w", tableName, err)
		}

		stats.Tables = append(stats.Tables, TableStats{
			Name:     tableName,
			RowCount: count,
		})
		stats.TotalRows += count
	}

	return stats, nil
}

func displayStats(stats *DatabaseStats, dbPath string) {
	fmt.Printf("\n=== Database Statistics ===\n")
	fmt.Printf("Database: %s\n", dbPath)
	fmt.Printf("Size:     %s\n", formatBytes(stats.DBSize))
	fmt.Printf("\n")

	if len(stats.Tables) == 0 {
		fmt.Printf("No tables found in database\n")
		return
	}

	fmt.Printf("Tables:   %d\n", len(stats.Tables))
	fmt.Printf("Total Rows: %d\n", stats.TotalRows)
	fmt.Printf("\n")

	fmt.Printf("%-30s %15s\n", "Table Name", "Row Count")
	fmt.Printf("%-30s %15s\n", "----------", "---------")
	for _, table := range stats.Tables {
		fmt.Printf("%-30s %15d\n", table.Name, table.RowCount)
	}
	fmt.Printf("\n")
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
