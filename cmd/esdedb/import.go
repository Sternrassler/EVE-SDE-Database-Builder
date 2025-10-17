package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/cli"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/database"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/logger"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/worker"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

var (
	sdeDir      string
	dbPath      string
	workerCount int
	skipErrors  bool
)

func newImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import EVE SDE JSONL files into SQLite database",
		Long: `Import command führt den 2-Phasen-Import von EVE SDE JSONL-Dateien aus:

Phase 1: Paralleles Parsing (Worker Pool)
  - Mehrere Worker-Threads parsen JSONL-Dateien gleichzeitig
  - Anzahl der Worker konfigurierbar (Standard: 4, Auto: -1 für NumCPU)

Phase 2: Sequenzielles Database-Insert (SQLite Single-Writer)
  - Geparste Daten werden in SQLite-Datenbank eingefügt
  - SQLite unterstützt nur einen Writer zur gleichen Zeit

Der Import zeigt einen Fortschrittsbalken mit Live-Metriken:
  - Anzahl verarbeiteter/fehlgeschlagener Dateien
  - Eingefügte Rows und Durchsatz (Rows/Sekunde)
  - Geschätzte verbleibende Zeit`,
		Example: `  # Import mit Standard-Einstellungen (4 Workers)
  esdedb import

  # Import mit benutzerdefiniertem SDE-Verzeichnis und Datenbank-Pfad
  esdedb import --sde-dir ./sde-JSONL --db ./eve-sde.db --workers 4

  # Automatische Worker-Anzahl basierend auf CPU-Kernen
  esdedb import --sde-dir ./sde-JSONL --db ./eve-sde.db --workers -1

  # Fehlerhafte Dateien überspringen und Import fortsetzen
  esdedb import --sde-dir ./sde-JSONL --db ./eve-sde.db --skip-errors

  # Import mit Verbose Logging (Debug-Level)
  esdedb --verbose import --sde-dir ./sde-JSONL`,
		RunE: runImportCmd,
	}

	// Flags
	cmd.Flags().StringVarP(&sdeDir, "sde-dir", "s", "./sde-JSONL", "Pfad zum Verzeichnis mit SDE JSONL-Dateien")
	cmd.Flags().StringVarP(&dbPath, "db", "d", "./eve-sde.db", "Pfad zur SQLite-Datenbank (wird erstellt falls nicht vorhanden)")
	cmd.Flags().IntVarP(&workerCount, "workers", "w", 4, "Anzahl paralleler Worker-Threads (-1 = Automatisch basierend auf CPU-Kernen)")
	cmd.Flags().BoolVar(&skipErrors, "skip-errors", false, "Überspringt fehlerhafte Dateien statt Import abzubrechen")

	return cmd
}

func runImportCmd(cmd *cobra.Command, args []string) error {
	log := logger.GetGlobalLogger()

	// Validate inputs
	if sdeDir == "" {
		return fmt.Errorf("--sde-dir darf nicht leer sein")
	}
	if dbPath == "" {
		return fmt.Errorf("--db darf nicht leer sein")
	}

	// Auto-detect worker count
	if workerCount == -1 {
		workerCount = runtime.NumCPU()
	}
	if workerCount <= 0 {
		workerCount = 1
	}

	log.Info("Starting EVE SDE Import",
		logger.Field{Key: "sde_dir", Value: sdeDir},
		logger.Field{Key: "db_path", Value: dbPath},
		logger.Field{Key: "workers", Value: workerCount},
		logger.Field{Key: "skip_errors", Value: skipErrors},
	)

	// Context mit Cancellation für Graceful Shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Signal Handling (SIGINT/SIGTERM)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Warn("Received interrupt signal, cancelling import...",
			logger.Field{Key: "signal", Value: sig.String()},
		)
		cancel()
	}()

	// Open Database
	db, err := database.NewDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Run Migrations (Schema Creation)
	// TODO: Implement schema migration
	// For now, we assume the schema exists or is created elsewhere

	// Create Worker Pool
	pool := worker.NewPool(workerCount)

	// Register Parsers (auto-register all available parsers)
	parsers := parser.RegisterParsers()

	log.Info("Registered parsers",
		logger.Field{Key: "parser_count", Value: len(parsers)},
	)

	// Create Orchestrator
	orch := worker.NewOrchestrator(db, pool, parsers)

	// Discover files first to set up progress bar
	files, err := worker.DiscoverJSONLFiles(sdeDir)
	if err != nil {
		return fmt.Errorf("failed to discover JSONL files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no JSONL files found in %s", sdeDir)
	}

	// Create enhanced progress bar with live metrics
	progressBar := cli.NewProgressBar(cli.ProgressBarConfig{
		Total:       len(files),
		Description: "Importing files",
		Width:       40,
		UpdateRate:  100 * time.Millisecond,
		ShowSpinner: false, // Spinner für einzelne Dateien kann optional aktiviert werden
	})

	// Start Import in background
	startTime := time.Now()
	done := make(chan struct{})
	var progress *worker.ProgressTracker
	var importErr error

	// Run import in goroutine
	go func() {
		progress, importErr = orch.ImportAll(ctx, sdeDir)
		close(done)
	}()

	// Create context for progress bar
	progressCtx, progressCancel := context.WithCancel(ctx)
	defer progressCancel()

	// Start progress bar updates in goroutine
	go func() {
		<-done
		progressCancel() // Signal progress bar to stop
	}()

	// Start progress bar (blocks until context is cancelled)
	progressBar.Start(progressCtx, progress)

	// Import finished
	progressBar.Finish()
	duration := time.Since(startTime)

	if importErr != nil {
		if importErr == context.Canceled {
			log.Warn("Import cancelled by user")
			return nil
		}
		return fmt.Errorf("import failed: %w", importErr)
	}

	// Report Results
	progressDetailed := progress.GetProgressDetailed()
	parsed := progressDetailed.ParsedFiles
	inserted := progressDetailed.InsertedFiles
	failed := progressDetailed.FailedFiles
	total := progressDetailed.TotalFiles

	log.Info("Import completed",
		logger.Field{Key: "total_files", Value: total},
		logger.Field{Key: "parsed_files", Value: int(parsed)},
		logger.Field{Key: "inserted_files", Value: int(inserted)},
		logger.Field{Key: "failed_files", Value: int(failed)},
		logger.Field{Key: "inserted_rows", Value: progressDetailed.InsertedRows},
		logger.Field{Key: "duration", Value: duration},
		logger.Field{Key: "rows_per_second", Value: progressDetailed.RowsPerSecond},
	)

	fmt.Printf("\n=== Import Summary ===\n")
	fmt.Printf("Files:     %d/%d parsed (%d failed)\n", parsed, total, failed)
	fmt.Printf("Rows:      %d inserted\n", progressDetailed.InsertedRows)
	fmt.Printf("Duration:  %v\n", duration)
	fmt.Printf("Throughput: %.0f rows/sec\n", progressDetailed.RowsPerSecond)
	fmt.Printf("\n")

	if failed > 0 {
		log.Warn("Some files failed to import",
			logger.Field{Key: "failed_count", Value: int(failed)},
		)
		if !skipErrors {
			return fmt.Errorf("%d files failed to import", failed)
		}
		cli.Warning("Warning: %d files failed to import (continuing due to --skip-errors)", failed)
	}

	return nil
}
