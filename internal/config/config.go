// Package config verwaltet Konfiguration (TOML + Env + CLI Flags)package config

// Siehe ADR-004: Configuration Format
package config

import (
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/BurntSushi/toml"
)

// Config ist die Haupt-Konfigurationsstruktur
type Config struct {
	Version  string         `toml:"version"`
	Database DatabaseConfig `toml:"database"`
	Import   ImportConfig   `toml:"import"`
	Logging  LoggingConfig  `toml:"logging"`
	Update   UpdateConfig   `toml:"update"`
}

// DatabaseConfig konfiguriert SQLite-spezifische Parameter
type DatabaseConfig struct {
	Path        string `toml:"path"`
	JournalMode string `toml:"journal_mode"`
	CacheSizeMB int    `toml:"cache_size_mb"`
}

// ImportConfig konfiguriert den JSONL-Import
type ImportConfig struct {
	SDEPath  string `toml:"sde_path"`
	Language string `toml:"language"`
	Workers  int    `toml:"workers"`
}

// LoggingConfig konfiguriert Logging-Verhalten
type LoggingConfig struct {
	Level  string `toml:"level"`
	Format string `toml:"format"`
}

// UpdateConfig konfiguriert Auto-Update (optional, v2.0)
type UpdateConfig struct {
	Enabled  bool   `toml:"enabled"`
	CheckURL string `toml:"check_url"`
}

// Load lädt Config mit Prioritäts-Kaskade: TOML < Env < Flags
func Load(configPath string) (*Config, error) {
	// 1. Default Config
	cfg := DefaultConfig()

	// 2. TOML laden (falls vorhanden)
	if _, err := os.Stat(configPath); err == nil {
		if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// 3. Environment Variables überschreiben
	applyEnvVars(&cfg)

	// 4. CLI Flags (wird extern in main.go überschrieben)

	// 5. Validierung
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// DefaultConfig liefert sinnvolle Defaults
func DefaultConfig() Config {
	return Config{
		Version: "1.0.0",
		Database: DatabaseConfig{
			Path:        "./eve_sde.db",
			JournalMode: "WAL",
			CacheSizeMB: 64,
		},
		Import: ImportConfig{
			SDEPath:  "./sde-JSONL",
			Language: "en",
			Workers:  0, // 0 = Auto (runtime.NumCPU())
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
		Update: UpdateConfig{
			Enabled: false,
		},
	}
}

// Validate prüft Config-Constraints
func (c *Config) Validate() error {
	if c.Database.Path == "" {
		return fmt.Errorf("database.path is required")
	}
	if c.Import.SDEPath == "" {
		return fmt.Errorf("import.sde_path is required")
	}

	// Workers: 0 = Auto, 1-32 explizit
	if c.Import.Workers < 0 || c.Import.Workers > 32 {
		return fmt.Errorf("import.workers must be 0-32 (0=auto, got %d)", c.Import.Workers)
	}
	if c.Import.Workers == 0 {
		c.Import.Workers = runtime.NumCPU()
	}

	// Language Validation
	validLangs := map[string]bool{
		"en": true, "de": true, "fr": true, "ja": true,
		"ru": true, "zh": true, "es": true, "ko": true,
	}
	if !validLangs[c.Import.Language] {
		return fmt.Errorf("invalid language: %s (must be: en, de, fr, ja, ru, zh, es, ko)", c.Import.Language)
	}

	// Logging Level
	validLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("invalid logging.level: %s (must be: debug, info, warn, error)", c.Logging.Level)
	}

	return nil
}

// applyEnvVars überschreibt Config mit Environment Variables
func applyEnvVars(cfg *Config) {
	if dbPath := os.Getenv("ESDEDB_DATABASE_PATH"); dbPath != "" {
		cfg.Database.Path = dbPath
	}
	if sdePath := os.Getenv("ESDEDB_SDE_PATH"); sdePath != "" {
		cfg.Import.SDEPath = sdePath
	}
	if lang := os.Getenv("ESDEDB_LANGUAGE"); lang != "" {
		cfg.Import.Language = lang
	}
	if level := os.Getenv("ESDEDB_LOG_LEVEL"); level != "" {
		cfg.Logging.Level = level
	}
	if format := os.Getenv("ESDEDB_LOGGING_FORMAT"); format != "" {
		cfg.Logging.Format = format
	}
	if workers := os.Getenv("ESDEDB_IMPORT_WORKER_COUNT"); workers != "" {
		// Note: Type conversion handled here, validation happens in Validate()
		if w, err := strconv.Atoi(workers); err == nil {
			cfg.Import.Workers = w
		}
	}
}
