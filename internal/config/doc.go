// Package config verwaltet Konfiguration (TOML + Env + CLI Flags).
//
// Das config-Package implementiert ein kaskadierendes Konfigurationssystem für die
// EVE SDE Database Builder Anwendung mit Prioritätsfolge: TOML < Environment < CLI Flags.
//
// Siehe ADR-004: Configuration Format für Designentscheidungen.
//
// # Konfigurationsquellen
//
// Die Konfiguration wird in folgender Priorität geladen (spätere überschreiben frühere):
//
//  1. Default-Werte (hartcodiert)
//  2. TOML-Konfigurationsdatei
//  3. Environment Variables
//  4. CLI Flags (extern in main.go)
//
// # Grundlegende Verwendung
//
// Laden Sie die Konfiguration aus einer TOML-Datei:
//
//	cfg, err := config.Load("config.toml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Default-Konfiguration
//
// Erhalten Sie eine Konfiguration mit sinnvollen Standardwerten:
//
//	cfg := config.DefaultConfig()
//	// cfg enthält: SQLite DB Path, Worker-Count=CPU, Log-Level=info, etc.
//
// # Konfigurationsstruktur
//
// Die Hauptkonfiguration ist in logische Bereiche unterteilt:
//
//   - Database: SQLite-Pfad, Journal-Mode, Cache-Größe
//   - Import: SDE-Pfad, Sprache, Worker-Anzahl
//   - Logging: Log-Level, Ausgabeformat
//   - Update: Auto-Update-Einstellungen (optional, v2.0)
//
// # TOML-Beispiel
//
//	version = "1.0.0"
//
//	[database]
//	path = "./eve_sde.db"
//	journal_mode = "WAL"
//	cache_size_mb = 64
//
//	[import]
//	sde_path = "./sde-JSONL"
//	language = "en"
//	workers = 0  # 0 = Auto (CPU count)
//
//	[logging]
//	level = "info"
//	format = "text"
//
// # Environment Variables
//
// Folgende Environment Variables werden unterstützt:
//
//   - ESDEDB_DATABASE_PATH: Überschreibt database.path
//   - ESDEDB_SDE_PATH: Überschreibt import.sde_path
//   - ESDEDB_LANGUAGE: Überschreibt import.language
//   - ESDEDB_LOG_LEVEL: Überschreibt logging.level
//   - ESDEDB_LOGGING_FORMAT: Überschreibt logging.format
//   - ESDEDB_IMPORT_WORKER_COUNT: Überschreibt import.workers
//
// # Validierung
//
// Die Konfiguration wird beim Laden automatisch validiert:
//
//	cfg, err := config.Load("config.toml")
//	// err != nil wenn Validierung fehlschlägt
//
// Validierungsregeln:
//
//   - database.path ist erforderlich
//   - import.sde_path ist erforderlich
//   - import.workers: 0-32 (0 = Auto)
//   - import.language: en, de, fr, ja, ru, zh, es, ko
//   - logging.level: debug, info, warn, error
//
// # Worker-Auto-Konfiguration
//
// Wenn import.workers = 0, wird automatisch runtime.NumCPU() verwendet:
//
//	cfg := config.DefaultConfig()
//	// cfg.Import.Workers wird nach Validate() auf CPU-Anzahl gesetzt
//	cfg.Validate()
//
// # Integration mit anderen Packages
//
// Die Konfiguration wird typischerweise an andere Komponenten weitergegeben:
//
//	cfg, _ := config.Load("config.toml")
//	logger := logger.NewLogger(cfg.Logging.Level, cfg.Logging.Format)
//	// DB-Connection mit cfg.Database.Path, etc.
//
// # Best Practices
//
//   - Verwenden Sie config.toml für lokale Entwicklung
//   - Nutzen Sie Environment Variables für Container/Cloud-Deployments
//   - CLI Flags für einmalige Überschreibungen
//   - Committen Sie niemals Produktions-Credentials in TOML-Dateien
//   - Verwenden Sie config.toml.example als Template
//
// # Fehlerbehandlung
//
// Load() gibt einen Fehler zurück wenn:
//
//   - TOML-Datei existiert aber nicht geparst werden kann
//   - Validierung fehlschlägt
//
// Fehlende TOML-Datei ist kein Fehler (Default-Werte werden verwendet).
//
// Siehe auch:
//   - ADR-004: Configuration Format
//   - config.toml.example für Beispielkonfiguration
package config
