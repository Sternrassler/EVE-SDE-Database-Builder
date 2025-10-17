package parser_test

import (
	"testing"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"
)

// TestRegisterParsers verifies that all parsers are registered correctly
// Epic #4: Complete - All 51 EVE SDE tables implemented
func TestRegisterParsers(t *testing.T) {
	parsers := parser.RegisterParsers()

	// Minimum expected parser count (51 EVE SDE tables)
	minExpectedParsers := 51

	// Verify we have at least 51 parsers
	if len(parsers) < minExpectedParsers {
		t.Errorf("Expected at least %d parsers, got %d", minExpectedParsers, len(parsers))
	}

	// Core parsers that must be registered (sanity check)
	requiredParsers := []string{
		// Core tables from Epic #4 Task #37
		"invTypes", "invGroups", "invCategories", "invMarketGroups", "invMetaGroups",
		"industryBlueprints",
		"dogmaAttributes", "dogmaEffects", "dogmaTypeAttributes", "dogmaTypeEffects",
		"mapRegions", "mapConstellations", "mapSolarSystems", "mapStargates", "mapPlanets",
		"chrRaces", "chrFactions",
		// Extended tables from Epic #4 Task #38 (sample)
		"chrAncestries", "chrBloodlines", "agtAgentTypes", "certCerts",
		"skins", "eveIcons", "_sde",
	}

	// Verify required parsers are registered
	for _, name := range requiredParsers {
		if _, ok := parsers[name]; !ok {
			t.Errorf("Required parser %s not registered", name)
		}
	}

	// Verify all parsers have correct table names
	for name, p := range parsers {
		if p.TableName() != name {
			t.Errorf("Parser %s has incorrect table name: %s", name, p.TableName())
		}
	}

	// Verify all parsers have non-empty columns
	for name, p := range parsers {
		if len(p.Columns()) == 0 {
			t.Errorf("Parser %s has empty columns", name)
		}
	}
}

// TestParserInstances verifies that parser instances are correctly configured
// Epic #4: Complete - Tests sample of all 51 parsers
func TestParserInstances(t *testing.T) {
	// Test representative sample from each category
	tests := []struct {
		name          string
		parser        parser.Parser
		expectedTable string
		minColumns    int // Minimum expected columns
	}{
		// Core Inventory & Market
		{"InvTypesParser", parser.InvTypesParser, "invTypes", 10},
		{"InvGroupsParser", parser.InvGroupsParser, "invGroups", 5},
		{"InvCategoriesParser", parser.InvCategoriesParser, "invCategories", 3},

		// Industry & Blueprints
		{"IndustryBlueprintsParser", parser.IndustryBlueprintsParser, "industryBlueprints", 2},

		// Dogma System (Core)
		{"DogmaAttributesParser", parser.DogmaAttributesParser, "dogmaAttributes", 5},
		{"DogmaEffectsParser", parser.DogmaEffectsParser, "dogmaEffects", 10},

		// Dogma System (Extended)
		{"DogmaAttributeCategoriesParser", parser.DogmaAttributeCategoriesParser, "dogmaAttributeCategories", 2},
		{"DogmaUnitsParser", parser.DogmaUnitsParser, "dogmaUnits", 2},

		// Universe/Map (Core)
		{"MapRegionsParser", parser.MapRegionsParser, "mapRegions", 4},
		{"MapSolarSystemsParser", parser.MapSolarSystemsParser, "mapSolarSystems", 5},

		// Universe/Map (Extended)
		{"MapMoonsParser", parser.MapMoonsParser, "mapMoons", 4},
		{"MapStarsParser", parser.MapStarsParser, "mapStars", 3},

		// Character/Faction (Core)
		{"ChrRacesParser", parser.ChrRacesParser, "chrRaces", 3},
		{"ChrFactionsParser", parser.ChrFactionsParser, "chrFactions", 5},

		// Character/NPC (Extended)
		{"ChrAncestriesParser", parser.ChrAncestriesParser, "chrAncestries", 3},
		{"ChrBloodlinesParser", parser.ChrBloodlinesParser, "chrBloodlines", 3},

		// Agents
		{"AgentTypesParser", parser.AgentTypesParser, "agtAgentTypes", 2},
		{"AgentsInSpaceParser", parser.AgentsInSpaceParser, "agtAgents", 5},

		// Certificates/Skills
		{"CertificatesParser", parser.CertificatesParser, "certCerts", 3},
		{"MasteriesParser", parser.MasteriesParser, "certMasteries", 2},

		// Skins
		{"SkinsParser", parser.SkinsParser, "skins", 3},

		// Station
		{"StationOperationsParser", parser.StationOperationsParser, "staOperations", 2},

		// Miscellaneous
		{"IconsParser", parser.IconsParser, "eveIcons", 2},
		{"SDEMetadataParser", parser.SDEMetadataParser, "_sde", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify table name
			if tt.parser.TableName() != tt.expectedTable {
				t.Errorf("%s.TableName() = %s, want %s", tt.name, tt.parser.TableName(), tt.expectedTable)
			}

			// Verify columns length (at least minimum)
			cols := tt.parser.Columns()
			if len(cols) < tt.minColumns {
				t.Errorf("%s.Columns() length = %d, want at least %d", tt.name, len(cols), tt.minColumns)
			}

			// Verify columns are not empty
			for i, col := range cols {
				if col == "" {
					t.Errorf("%s.Columns()[%d] is empty", tt.name, i)
				}
			}
		})
	}
}

// TestParserInterfaceCompliance verifies all parsers implement the Parser interface
// Epic #4: Complete - Validates all registered parsers via RegisterParsers()
func TestParserInterfaceCompliance(t *testing.T) {
	// Get all registered parsers dynamically
	allParsers := parser.RegisterParsers()

	// Verify we have a reasonable number of parsers
	if len(allParsers) < 51 {
		t.Fatalf("Expected at least 51 registered parsers, got %d", len(allParsers))
	}

	// Test each registered parser
	for name, p := range allParsers {
		t.Run(name, func(t *testing.T) {
			// Verify TableName returns non-empty string
			if p.TableName() == "" {
				t.Errorf("Parser %s returned empty TableName", name)
			}

			// Verify TableName matches registry key
			if p.TableName() != name {
				t.Errorf("Parser %s has mismatched TableName: got %s", name, p.TableName())
			}

			// Verify Columns returns non-empty slice
			if len(p.Columns()) == 0 {
				t.Errorf("Parser %s returned empty Columns", name)
			}

			// Verify all column names are non-empty
			for i, col := range p.Columns() {
				if col == "" {
					t.Errorf("Parser %s has empty column at index %d", name, i)
				}
			}
		})
	}
}
