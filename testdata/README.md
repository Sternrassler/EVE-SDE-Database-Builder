# Test Data Repository

This directory contains sample JSONL test data for all EVE SDE tables.

## Structure

```
testdata/
└── sde/
    ├── _sde.jsonl              # SDE metadata
    ├── agtAgents.jsonl         # Agent data
    ├── agtAgentTypes.jsonl     # Agent types
    ├── certCerts.jsonl         # Certificates
    ├── certMasteries.jsonl     # Certificate masteries
    ├── chrAncestries.jsonl     # Character ancestries
    ├── chrAttributes.jsonl     # Character attributes
    ├── chrBloodlines.jsonl     # Character bloodlines
    ├── chrFactions.jsonl       # Factions
    ├── chrNPCCharacters.jsonl  # NPC characters
    ├── chrRaces.jsonl          # Races
    ├── contrabandTypes.jsonl   # Contraband types
    ├── controlTowerResources.jsonl  # Control tower resources
    ├── crpActivities.jsonl     # Corporation activities
    ├── crpNPCCorporationDivisions.jsonl  # NPC corporation divisions
    ├── crpNPCCorporations.jsonl  # NPC corporations
    ├── dbuffCollections.jsonl  # Dogma buff collections
    ├── dogmaAttributeCategories.jsonl  # Dogma attribute categories
    ├── dogmaAttributes.jsonl   # Dogma attributes
    ├── dogmaEffects.jsonl      # Dogma effects
    ├── dogmaTypeAttributes.jsonl  # Type-attribute mappings
    ├── dogmaTypeEffects.jsonl  # Type-effect mappings
    ├── dogmaUnits.jsonl        # Dogma units
    ├── dynamicItemAttributes.jsonl  # Dynamic item attributes
    ├── eveGraphics.jsonl       # Graphics
    ├── eveIcons.jsonl          # Icons
    ├── industryBlueprints.jsonl  # Industry blueprints
    ├── invCategories.jsonl     # Inventory categories
    ├── invGroups.jsonl         # Inventory groups
    ├── invMarketGroups.jsonl   # Market groups
    ├── invMetaGroups.jsonl     # Meta groups
    ├── invTypes.jsonl          # Inventory types (items)
    ├── mapAsteroidBelts.jsonl  # Asteroid belts
    ├── mapConstellations.jsonl # Constellations
    ├── mapLandmarks.jsonl      # Landmarks
    ├── mapMoons.jsonl          # Moons
    ├── mapPlanets.jsonl        # Planets
    ├── mapRegions.jsonl        # Regions
    ├── mapSolarSystems.jsonl   # Solar systems
    ├── mapStargates.jsonl      # Stargates
    ├── mapStars.jsonl          # Stars
    ├── planetResources.jsonl   # Planet resources
    ├── planetSchematics.jsonl  # Planet schematics
    ├── skinLicenses.jsonl      # Skin licenses
    ├── skinMaterials.jsonl     # Skin materials
    ├── skins.jsonl             # Skins
    ├── sovereigntyUpgrades.jsonl  # Sovereignty upgrades
    ├── staOperations.jsonl     # Station operations
    ├── staServices.jsonl       # Station services
    ├── staStations.jsonl       # Stations
    ├── translationLanguages.jsonl  # Translation languages
    ├── typeBonuses.jsonl       # Type bonuses
    └── typeDogma.jsonl         # Type dogma data
```

## File Format

All files follow the JSONL (JSON Lines) format:
- Each line is a valid JSON object
- Each line represents one record
- Files contain 1-3 sample records for testing purposes
- Data is minimal but realistic, based on actual EVE Online SDE structure

## Usage

### Using the testutil Package

The `internal/testutil` package provides helper functions to load and work with test data:

```go
import (
    "testing"
    "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/testutil"
)

func TestExample(t *testing.T) {
    // Get path to testdata directory
    testDataPath := testutil.GetTestDataPath()
    
    // Get path to a specific test data file
    invTypesPath := testutil.GetTestDataFile("invTypes")
    
    // Load a JSONL file as raw lines
    lines := testutil.LoadJSONLFile(t, "invTypes")
    
    // Load and unmarshal JSONL into typed records
    type InvType struct {
        TypeID   int    `json:"typeID"`
        TypeName string `json:"typeName"`
        GroupID  *int   `json:"groupID"`
    }
    records := testutil.LoadJSONLFileAsRecords[InvType](t, "invTypes")
    
    // Get all available table names
    tables := testutil.TableNames()
}
```

### Direct File Access

You can also access test data files directly:

```go
import (
    "path/filepath"
    "runtime"
)

// Get project root
_, filename, _, _ := runtime.Caller(0)
projectRoot := filepath.Join(filepath.Dir(filename), "..")
testDataFile := filepath.Join(projectRoot, "testdata", "sde", "invTypes.jsonl")
```

## Data Contents

### Sample Records

Each file contains realistic sample data:

- **invTypes.jsonl**: Tritanium, Pyerite, Mexallon (basic minerals)
- **invGroups.jsonl**: Mineral, Frigate, Cruiser (basic groups)
- **mapSolarSystems.jsonl**: Tanoo, Jita, H-PA29 (various security spaces)
- **chrRaces.jsonl**: Caldari, Minmatar, Amarr (major races)
- And more...

### Data Relationships

Test data maintains referential integrity where appropriate:
- `invTypes.groupID` references `invGroups.groupID`
- `mapSolarSystems.regionID` references `mapRegions.regionID`
- `staStations.solarSystemID` references `mapSolarSystems.solarSystemID`
- etc.

## Maintenance

### Adding New Test Data

When adding new test data files:

1. Follow the JSONL format (one JSON object per line)
2. Include 1-3 realistic sample records
3. Update `internal/testutil/testutil.go` `TableNames()` function
4. Ensure field names match the actual SDE schema
5. Use realistic IDs that don't conflict with existing data

### Updating Existing Data

When updating test data:

1. Maintain JSONL format
2. Preserve referential integrity
3. Keep sample data minimal (1-3 records)
4. Run tests to verify changes: `go test ./internal/testutil/...`

## Testing

To verify test data integrity:

```bash
# Run testutil tests
go test -v ./internal/testutil/...

# Run all tests that use testdata
go test -v ./...
```

## Notes

- Test data is **not** production data
- Files are minimal to keep repository size small
- Data values are realistic but may not reflect current EVE Online state
- All JSON fields use proper types (`null` for nullable fields)
- Scientific notation is used for large numbers (e.g., `1.5e11`)

## References

- [EVE Online SDE Documentation](https://developers.eveonline.com/)
- [RIFT SDE Schema](https://sde.riftforeve.online/)
- [JSONL Format Specification](https://jsonlines.org/)
