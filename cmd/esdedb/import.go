package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/database"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/logger"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/worker"
	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
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
  Phase 2: Sequenzielles Database-Insert (SQLite Single-Writer)

Beispiel:
  esdedb import --sde-dir ./sde-JSONL --db ./eve-sde.db --workers 4
  esdedb import --sde-dir ./sde-JSONL --db ./eve-sde.db --workers -1  # Auto (NumCPU)
  esdedb import --sde-dir ./sde-JSONL --db ./eve-sde.db --skip-errors # Fehler überspringen`,
		RunE: runImportCmd,
	}

	// Flags
	cmd.Flags().StringVarP(&sdeDir, "sde-dir", "s", "./sde-JSONL", "Pfad zum SDE JSONL-Verzeichnis")
	cmd.Flags().StringVarP(&dbPath, "db", "d", "./eve-sde.db", "Pfad zur SQLite-Datenbank")
	cmd.Flags().IntVarP(&workerCount, "workers", "w", 4, "Anzahl Worker (-1 = Auto/NumCPU)")
	cmd.Flags().BoolVar(&skipErrors, "skip-errors", false, "Fehlerhafte Dateien überspringen (Standard: Abbruch bei Fehler)")

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
	defer db.Close()

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

	// Create progress bar
	bar := progressbar.NewOptions(len(files),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(40),
		progressbar.OptionSetDescription("[cyan]Importing files...[reset]"),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionShowIts(),
		progressbar.OptionSetItsString("files"),
	)

	// Start Import in background and update progress bar
	startTime := time.Now()

	// Channel to track progress updates
	done := make(chan struct{})
	var progress *worker.ProgressTracker
	var importErr error

	// Run import in goroutine
	go func() {
		progress, importErr = orch.ImportAll(ctx, sdeDir)
		close(done)
	}()

	// Update progress bar periodically
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	lastParsed := 0
	for {
		select {
		case <-done:
			// Import finished, update bar to completion
			if progress != nil {
				parsed, _, _, _ := progress.GetProgress()
				remaining := parsed - lastParsed
				for i := 0; i < remaining; i++ {
					_ = bar.Add(1)
				}
			}
			goto ImportDone
		case <-ticker.C:
			// Update progress bar based on parsed files
			if progress != nil {
				parsed, _, _, _ := progress.GetProgress()
				diff := parsed - lastParsed
				if diff > 0 {
					for i := 0; i < diff; i++ {
						_ = bar.Add(1)
					}
					lastParsed = parsed
				}
			}
		}
	}

ImportDone:
	duration := time.Since(startTime)
	fmt.Println() // New line after progress bar

	if importErr != nil {
		if importErr == context.Canceled {
			log.Warn("Import cancelled by user")
			return nil
		}
		return fmt.Errorf("import failed: %w", importErr)
	}

	// Report Results
	parsed, inserted, failed, total := progress.GetProgress()
	progressDetailed := progress.GetProgressDetailed()

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
		fmt.Printf("⚠️  Warning: %d files failed to import (continuing due to --skip-errors)\n", failed)
	}

	return nil
}
