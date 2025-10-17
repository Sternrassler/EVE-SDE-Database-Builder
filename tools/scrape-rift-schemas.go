// tools/scrape-rift-schemas.go
// Scrapes JSON schema definitions from RIFT SDE API
// ADR Reference: ADR-003 (Full Code-Gen Approach)
//go:build !add_tomap_methods
// +build !add_tomap_methods

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/logger"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/retry"
)

const (
	riftBaseURL    = "https://sde.riftforeve.online"
	defaultTimeout = 30 * time.Second
)

// List of EVE SDE tables to scrape (using RIFT naming convention)
var eveTables = []string{
	// Core tables
	"_sde",
	"types",
	"groups",
	"categories",
	"marketGroups",
	"metaGroups",

	// Character/NPC
	"ancestries",
	"bloodlines",
	"races",
	"factions",
	"characterAttributes",
	"npcCharacters",
	"npcCorporations",
	"npcCorporationDivisions",
	"npcStations",

	// Agents
	"agentTypes",
	"agentsInSpace",

	// Blueprints/Industry
	"blueprints",

	// Dogma (Ship fitting system)
	"dogmaAttributes",
	"dogmaAttributeCategories",
	"dogmaEffects",
	"dogmaUnits",
	"typeDogma",
	"dynamicItemAttributes",

	// Universe/Map
	"mapRegions",
	"mapConstellations",
	"mapSolarSystems",
	"mapStargates",
	"mapPlanets",
	"mapMoons",
	"mapStars",
	"mapAsteroidBelts",
	"landmarks",

	// Certificates/Skills
	"certificates",
	"masteries",

	// Skins
	"skins",
	"skinLicenses",
	"skinMaterials",

	// Translation/Localization
	"translationLanguages",

	// Station services
	"stationOperations",
	"stationServices",
	"sovereigntyUpgrades",

	// Miscellaneous
	"icons",
	"graphics",
	"contrabandTypes",
	"controlTowerResources",
	"corporationActivities",
	"dbuffCollections",
	"planetResources",
	"planetSchematics",
	"typeBonus",
}

type Config struct {
	OutputDir string
	BaseURL   string
	Timeout   time.Duration
	Verbose   bool
}

func main() {
	cfg := parseFlags()

	log := logger.NewLogger("info", "text")
	if cfg.Verbose {
		log = logger.NewLogger("debug", "text")
	}

	log.Info("Starting RIFT SDE Schema Scraper", logger.Field{Key: "tables", Value: len(eveTables)})

	ctx := context.Background()
	client := &http.Client{Timeout: cfg.Timeout}

	successCount := 0
	failCount := 0

	for _, table := range eveTables {
		log.Info("Scraping schema", logger.Field{Key: "table", Value: table})

		err := scrapeTableSchema(ctx, client, cfg, table, log)
		if err != nil {
			log.Error("Failed to scrape schema",
				logger.Field{Key: "table", Value: table},
				logger.Field{Key: "error", Value: err.Error()})
			failCount++
			continue
		}

		successCount++
	}

	log.Info("Schema scraping completed",
		logger.Field{Key: "success", Value: successCount},
		logger.Field{Key: "failed", Value: failCount},
		logger.Field{Key: "total", Value: len(eveTables)})

	if failCount > 0 {
		os.Exit(1)
	}
}

func parseFlags() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.OutputDir, "output", "schemas", "Output directory for schema files")
	flag.StringVar(&cfg.BaseURL, "base-url", riftBaseURL, "Base URL for RIFT SDE API")
	flag.DurationVar(&cfg.Timeout, "timeout", defaultTimeout, "HTTP request timeout")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Enable verbose logging")
	flag.Parse()
	return cfg
}

func scrapeTableSchema(ctx context.Context, client *http.Client, cfg *Config, table string, log *logger.Logger) error {
	// Use retry policy for HTTP requests
	policy := retry.HTTPPolicy()

	var schema []byte
	err := policy.Do(ctx, func() error {
		// Fetch schema from RIFT API
		// RIFT provides schema documentation pages at /schema/<table>/
		url := fmt.Sprintf("%s/schema/%s/", cfg.BaseURL, table)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return errors.NewFatal(fmt.Sprintf("failed to create request for %s", table), err)
		}

		resp, err := client.Do(req)
		if err != nil {
			// HTTP errors are typically retryable (network issues, timeouts)
			return errors.NewRetryable(fmt.Sprintf("HTTP request failed for %s", table), err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			if resp.StatusCode >= 500 {
				// Server errors are retryable
				return errors.NewRetryable(fmt.Sprintf("server error %d for %s", resp.StatusCode, table), nil)
			}
			// Client errors (4xx) are not retryable
			return errors.NewValidation(fmt.Sprintf("client error %d for %s", resp.StatusCode, table), nil)
		}

		// Read response body (HTML schema documentation)
		_, err = io.ReadAll(resp.Body)
		if err != nil {
			return errors.NewRetryable(fmt.Sprintf("failed to read response for %s", table), err)
		}

		// For now, create a minimal placeholder JSON that indicates the schema page exists
		// TODO: Future enhancement - parse HTML schema or download sample from CCP JSONL files
		// This placeholder confirms the schema page is accessible and lists the correct table name
		schemaObj := map[string]interface{}{
			"_table":  table,
			"_source": url,
			"_status": "schema_page_verified",
		}

		schemaJSON, err := json.MarshalIndent(schemaObj, "", "  ")
		if err != nil {
			return errors.NewFatal(fmt.Sprintf("failed to marshal placeholder JSON for %s", table), err)
		}

		schema = schemaJSON
		return nil
	})

	if err != nil {
		return err
	}

	// Save schema to file
	return saveSchema(cfg.OutputDir, table, schema)
}

func saveSchema(outputDir, table string, schema []byte) error {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return errors.NewFatal("failed to create output directory", err)
	}

	// Write schema file
	filename := filepath.Join(outputDir, fmt.Sprintf("%s.json", table))
	if err := os.WriteFile(filename, schema, 0644); err != nil {
		return errors.NewFatal(fmt.Sprintf("failed to write schema file %s", filename), err)
	}

	return nil
}
