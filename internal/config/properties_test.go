package config

import (
	"runtime"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperties_ConfigValidation_WorkerCountNormalization tests that worker count 0 is always normalized to runtime.NumCPU()
func TestProperties_ConfigValidation_WorkerCountNormalization(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("worker count 0 normalizes to NumCPU", prop.ForAll(
		func(dbPath, sdePath, lang, logLevel string) bool {
			cfg := Config{
				Version: "1.0.0",
				Database: DatabaseConfig{
					Path:        dbPath,
					JournalMode: "WAL",
					CacheSizeMB: 64,
				},
				Import: ImportConfig{
					SDEPath:  sdePath,
					Language: lang,
					Workers:  0, // Always test with 0
				},
				Logging: LoggingConfig{
					Level:  logLevel,
					Format: "text",
				},
			}

			err := cfg.Validate()
			if err != nil {
				return false
			}

			// After validation, worker count should be normalized to runtime.NumCPU()
			return cfg.Import.Workers == runtime.NumCPU()
		},
		genNonEmptyString(),      // dbPath
		genNonEmptyString(),      // sdePath
		genValidLanguage(),       // lang
		genValidLoggingLevel(),   // logLevel
	))

	properties.TestingRun(t)
}

// TestProperties_ConfigValidation_ValidConfigAlwaysValid tests that valid configs always pass validation
func TestProperties_ConfigValidation_ValidConfigAlwaysValid(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("valid config always passes validation", prop.ForAll(
		func(workers int, lang, logLevel string) bool {
			cfg := Config{
				Version: "1.0.0",
				Database: DatabaseConfig{
					Path:        "./test.db",
					JournalMode: "WAL",
					CacheSizeMB: 64,
				},
				Import: ImportConfig{
					SDEPath:  "./test-sde",
					Language: lang,
					Workers:  workers,
				},
				Logging: LoggingConfig{
					Level:  logLevel,
					Format: "text",
				},
			}

			err := cfg.Validate()
			return err == nil
		},
		gen.IntRange(1, 32),    // workers
		genValidLanguage(),     // lang
		genValidLoggingLevel(), // logLevel
	))

	properties.TestingRun(t)
}

// TestProperties_ConfigValidation_InvalidWorkerCountFails tests that invalid worker counts fail validation
func TestProperties_ConfigValidation_InvalidWorkerCountFails(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("invalid worker count fails validation", prop.ForAll(
		func(workers int) bool {
			cfg := Config{
				Version: "1.0.0",
				Database: DatabaseConfig{
					Path:        "./test.db",
					JournalMode: "WAL",
					CacheSizeMB: 64,
				},
				Import: ImportConfig{
					SDEPath:  "./test-sde",
					Language: "en",
					Workers:  workers,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "text",
				},
			}

			err := cfg.Validate()
			// Should fail for values < 0 or > 32
			return err != nil
		},
		gen.OneGenOf(
			gen.IntRange(-100, -1),  // Negative values
			gen.IntRange(33, 100),   // Too high values
		),
	))

	properties.TestingRun(t)
}

// TestProperties_ConfigValidation_EmptyPathsFail tests that empty required paths fail validation
func TestProperties_ConfigValidation_EmptyPathsFail(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("empty database path fails validation", prop.ForAll(
		func() bool {
			cfg := Config{
				Version: "1.0.0",
				Database: DatabaseConfig{
					Path:        "", // Empty path
					JournalMode: "WAL",
					CacheSizeMB: 64,
				},
				Import: ImportConfig{
					SDEPath:  "./test-sde",
					Language: "en",
					Workers:  4,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "text",
				},
			}

			err := cfg.Validate()
			return err != nil
		},
	))

	properties.Property("empty SDE path fails validation", prop.ForAll(
		func() bool {
			cfg := Config{
				Version: "1.0.0",
				Database: DatabaseConfig{
					Path:        "./test.db",
					JournalMode: "WAL",
					CacheSizeMB: 64,
				},
				Import: ImportConfig{
					SDEPath:  "", // Empty path
					Language: "en",
					Workers:  4,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "text",
				},
			}

			err := cfg.Validate()
			return err != nil
		},
	))

	properties.TestingRun(t)
}

// TestProperties_ConfigValidation_InvalidLanguageFails tests that invalid languages fail validation
func TestProperties_ConfigValidation_InvalidLanguageFails(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("invalid language fails validation", prop.ForAll(
		func(lang string) bool {
			cfg := Config{
				Version: "1.0.0",
				Database: DatabaseConfig{
					Path:        "./test.db",
					JournalMode: "WAL",
					CacheSizeMB: 64,
				},
				Import: ImportConfig{
					SDEPath:  "./test-sde",
					Language: lang,
					Workers:  4,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "text",
				},
			}

			err := cfg.Validate()
			return err != nil
		},
		genInvalidLanguage(),
	))

	properties.TestingRun(t)
}

// TestProperties_ConfigValidation_InvalidLogLevelFails tests that invalid log levels fail validation
func TestProperties_ConfigValidation_InvalidLogLevelFails(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("invalid log level fails validation", prop.ForAll(
		func(level string) bool {
			cfg := Config{
				Version: "1.0.0",
				Database: DatabaseConfig{
					Path:        "./test.db",
					JournalMode: "WAL",
					CacheSizeMB: 64,
				},
				Import: ImportConfig{
					SDEPath:  "./test-sde",
					Language: "en",
					Workers:  4,
				},
				Logging: LoggingConfig{
					Level:  level,
					Format: "text",
				},
			}

			err := cfg.Validate()
			return err != nil
		},
		genInvalidLogLevel(),
	))

	properties.TestingRun(t)
}

// Generator helpers

func genNonEmptyString() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0
	})
}

func genValidLanguage() gopter.Gen {
	return gen.OneConstOf("en", "de", "fr", "ja", "ru", "zh", "es", "ko")
}

func genInvalidLanguage() gopter.Gen {
	validLangs := map[string]bool{
		"en": true, "de": true, "fr": true, "ja": true,
		"ru": true, "zh": true, "es": true, "ko": true,
	}

	return gen.AlphaString().SuchThat(func(s string) bool {
		return !validLangs[s] && len(s) > 0
	})
}

func genValidLoggingLevel() gopter.Gen {
	return gen.OneConstOf("debug", "info", "warn", "error")
}

func genInvalidLogLevel() gopter.Gen {
	validLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}

	return gen.AlphaString().SuchThat(func(s string) bool {
		return !validLevels[s] && len(s) > 0
	})
}
