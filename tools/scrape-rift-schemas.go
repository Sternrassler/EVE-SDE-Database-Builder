// tools/scrape-rift-schemas.go
// Scrapes JSON schema definitions from RIFT SDE API
// ADR Reference: ADR-003 (Full Code-Gen Approach)
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

// List of EVE SDE tables to scrape
var eveTables = []string{
	// Inventory
	"invTypes",
	"invGroups",
	"invCategories",
	"invMarketGroups",
	"invMetaTypes",
	"invMetaGroups",
	"invTraits",
	"invTypeMaterials",
	"invTypeReactions",
	"invContrabandTypes",
	"invFlags",

	// Industry/Blueprints
	"industryBlueprints",
	"industryActivity",
	"industryActivityMaterials",
	"industryActivityProducts",
	"industryActivityProbabilities",
	"industryActivitySkills",

	// Dogma (Ship fitting system)
	"dogmaAttributes",
	"dogmaEffects",
	"dogmaTypeAttributes",
	"dogmaTypeEffects",
	"dogmaAttributeCategories",
	"dogmaAttributeTypes",
	"dogmaExpressions",
	"dogmaUnits",

	// Universe/Map
	"mapRegions",
	"mapConstellations",
	"mapSolarSystems",
	"mapSolarSystemJumps",
	"mapDenormalize",
	"mapJumps",
	"mapLocationWormholeClasses",
	"mapLocationScenes",
	"mapCelestialStatistics",
	"mapUniverse",
	"mapStargates",
	"mapPlanets",

	// Character/NPC
	"chrFactions",
	"chrRaces",
	"chrAncestries",
	"chrBloodlines",
	"chrAttributes",
	"agtAgents",
	"agtAgentTypes",
	"agtResearchAgents",

	// NPC Corporations
	"crpNPCCorporations",
	"crpNPCCorporationDivisions",
	"crpNPCCorporationTrades",
	"crpActivities",

	// Certificates/Skills
	"certCerts",
	"certMasteries",
	"certSkills",

	// Translation/Localization
	"translationTables",

	// Skins
	"skinLicenses",
	"skinMaterials",
	"skinShip",

	// Research
	"ramActivities",
	"ramAssemblyLineTypes",
	"ramAssemblyLineTypeDetailPerCategory",
	"ramAssemblyLineTypeDetailPerGroup",
	"ramInstallationTypeContents",

	// Station services
	"staOperations",
	"staOperationServices",
	"staServices",
	"staStations",
	"staStationTypes",

	// Miscellaneous
	"eveUnits",
	"eveGraphics",
	"eveIcons",
	"planetSchematics",
	"planetSchematicsPinMap",
	"planetSchematicsTypeMap",
	"warCombatZones",
	"warCombatZoneSystems",
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
		// RIFT provides schema information via the API endpoint for each table
		url := fmt.Sprintf("%s/%s?limit=1", cfg.BaseURL, table)

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

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.NewRetryable(fmt.Sprintf("failed to read response for %s", table), err)
		}

		// Parse JSON to validate it's well-formed
		var data interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			return errors.NewValidation(fmt.Sprintf("invalid JSON response for %s", table), err)
		}

		schema = body
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
