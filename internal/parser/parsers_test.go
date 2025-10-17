package parser_test

import (
	"testing"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"
)

// TestRegisterParsers verifies that all 10 parsers are registered correctly
func TestRegisterParsers(t *testing.T) {
	parsers := parser.RegisterParsers()

	expectedParsers := []string{
		"invTypes",
		"invGroups",
		"industryBlueprints",
		"dogmaAttributes",
		"mapSolarSystems",
		"dogmaEffects",
		"dogmaTypeAttributes",
		"dogmaTypeEffects",
		"mapRegions",
		"mapConstellations",
	}

	// Verify count
	if len(parsers) != len(expectedParsers) {
		t.Errorf("Expected %d parsers, got %d", len(expectedParsers), len(parsers))
	}

	// Verify each parser is registered
	for _, name := range expectedParsers {
		if _, ok := parsers[name]; !ok {
			t.Errorf("Parser %s not registered", name)
		}
	}

	// Verify all parsers have correct table names
	for name, p := range parsers {
		if p.TableName() != name {
			t.Errorf("Parser %s has incorrect table name: %s", name, p.TableName())
		}
	}
}

// TestParserInstances verifies that parser instances are correctly configured
func TestParserInstances(t *testing.T) {
	tests := []struct {
		name           string
		parser         parser.Parser
		expectedTable  string
		expectedColLen int
	}{
		{
			name:           "InvTypesParser",
			parser:         parser.InvTypesParser,
			expectedTable:  "invTypes",
			expectedColLen: 15,
		},
		{
			name:           "InvGroupsParser",
			parser:         parser.InvGroupsParser,
			expectedTable:  "invGroups",
			expectedColLen: 9,
		},
		{
			name:           "IndustryBlueprintsParser",
			parser:         parser.IndustryBlueprintsParser,
			expectedTable:  "industryBlueprints",
			expectedColLen: 2,
		},
		{
			name:           "DogmaAttributesParser",
			parser:         parser.DogmaAttributesParser,
			expectedTable:  "dogmaAttributes",
			expectedColLen: 10,
		},
		{
			name:           "MapSolarSystemsParser",
			parser:         parser.MapSolarSystemsParser,
			expectedTable:  "mapSolarSystems",
			expectedColLen: 9,
		},
		{
			name:           "DogmaEffectsParser",
			parser:         parser.DogmaEffectsParser,
			expectedTable:  "dogmaEffects",
			expectedColLen: 28,
		},
		{
			name:           "DogmaTypeAttributesParser",
			parser:         parser.DogmaTypeAttributesParser,
			expectedTable:  "dogmaTypeAttributes",
			expectedColLen: 4,
		},
		{
			name:           "DogmaTypeEffectsParser",
			parser:         parser.DogmaTypeEffectsParser,
			expectedTable:  "dogmaTypeEffects",
			expectedColLen: 3,
		},
		{
			name:           "MapRegionsParser",
			parser:         parser.MapRegionsParser,
			expectedTable:  "mapRegions",
			expectedColLen: 6,
		},
		{
			name:           "MapConstellationsParser",
			parser:         parser.MapConstellationsParser,
			expectedTable:  "mapConstellations",
			expectedColLen: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify table name
			if tt.parser.TableName() != tt.expectedTable {
				t.Errorf("%s.TableName() = %s, want %s", tt.name, tt.parser.TableName(), tt.expectedTable)
			}

			// Verify columns length
			cols := tt.parser.Columns()
			if len(cols) != tt.expectedColLen {
				t.Errorf("%s.Columns() length = %d, want %d", tt.name, len(cols), tt.expectedColLen)
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
func TestParserInterfaceCompliance(t *testing.T) {
	parsers := []parser.Parser{
		parser.InvTypesParser,
		parser.InvGroupsParser,
		parser.IndustryBlueprintsParser,
		parser.DogmaAttributesParser,
		parser.MapSolarSystemsParser,
		parser.DogmaEffectsParser,
		parser.DogmaTypeAttributesParser,
		parser.DogmaTypeEffectsParser,
		parser.MapRegionsParser,
		parser.MapConstellationsParser,
	}

	for _, p := range parsers {
		// Verify TableName returns non-empty string
		if p.TableName() == "" {
			t.Errorf("Parser %T returned empty TableName", p)
		}

		// Verify Columns returns non-empty slice
		if len(p.Columns()) == 0 {
			t.Errorf("Parser %T returned empty Columns", p)
		}
	}
}
