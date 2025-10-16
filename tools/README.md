# EVE SDE Code Generation Tools

This directory contains tools for generating Go parsers from EVE SDE JSON schemas.

## Tools

### 1. RIFT Schema Scraper (`scrape-rift-schemas.go`)

Tool zum Scrapen aller Schema-Definitionen von der RIFT SDE API.

**ADR Reference:** ADR-003 (Full Code-Gen Approach)  
**Epic:** #3 JSONL Parser Migration  
**Source:** https://sde.riftforeve.online/

## Usage

### Basic Usage

```bash
go run tools/scrape-rift-schemas.go
```

This will verify access to 51 EVE SDE table schema pages and create placeholder JSON files in the `schemas/` directory.

### Options

```bash
go run tools/scrape-rift-schemas.go [flags]
```

Available flags:
- `-output <dir>`: Output directory for schema files (default: `schemas`)
- `-base-url <url>`: Base URL for RIFT SDE API (default: `https://sde.riftforeve.online`)
- `-timeout <duration>`: HTTP request timeout (default: `30s`)
- `-verbose`: Enable verbose/debug logging

### Examples

```bash
# Download schemas to custom directory
go run tools/scrape-rift-schemas.go -output /tmp/eve-schemas

# Use verbose logging
go run tools/scrape-rift-schemas.go -verbose

# Increase timeout for slow connections
go run tools/scrape-rift-schemas.go -timeout 60s
```

## Features

- **Automatic Retry**: HTTP requests use exponential backoff retry logic for transient errors
- **Error Handling**: Distinguishes between retryable (5xx, network) and non-retryable (4xx, validation) errors
- **51 Tables**: Verifies schema pages for all EVE SDE tables including:
  - Core tables (types, groups, categories, marketGroups, metaGroups)
  - Character/NPC data (ancestries, bloodlines, races, factions, etc.)
  - Industry/Blueprints
  - Dogma (ship fitting system attributes and effects)
  - Universe/Map data (regions, constellations, systems, etc.)
  - Certificates/Skills
  - Skins
  - And more...

## Output

Schema placeholder files are saved as JSON in the specified output directory:
- `schemas/types.json`
- `schemas/groups.json`
- `schemas/dogmaAttributes.json`
- etc.

Each file contains:
```json
{
  "_table": "types",
  "_source": "https://sde.riftforeve.online/schema/types/",
  "_status": "schema_page_verified"
}
```

## Current Implementation

The tool verifies that RIFT schema documentation pages are accessible and creates placeholder JSON files. This confirms the correct table names and validates network access to the RIFT documentation site.

**Future Enhancement:** Parse HTML schema documentation or download sample JSONL data from CCP to generate proper JSON schema files for use with code generation tools like `quicktype`.

## Testing

Run tests:

```bash
go test -v ./tools/...
```

Tests include:
- Mock HTTP client test for successful downloads
- Retry behavior on HTTP errors
- HTML response handling (creates valid JSON output)
- Client error (4xx) handling
- File I/O tests

## Implementation Details

- Uses `internal/retry` package for HTTP retry logic with exponential backoff
- Uses `internal/logger` package for structured logging
- Uses `internal/errors` package for typed error handling (Retryable, Fatal, Validation)
- HTTP timeout: 30s (configurable)
- Retry policy: 3 attempts with 100ms-5s backoff

## Table Names

The tool uses RIFT's naming convention for tables (e.g., `types`, `groups`, `dogmaAttributes`) which differs from the CCP database naming convention (e.g., `invTypes`, `invGroups`, `dogmaAttributes`).

## Next Steps

After running this tool, the placeholders confirm schema page accessibility. For actual code generation, use the `generate-parsers.sh` script (see below).

---

### 2. Parser Generator (`generate-parsers.sh`)

Batch code generation script for EVE SDE JSONL parsers using quicktype.

**ADR Reference:** ADR-003 (JSONL Parser Architecture)  
**Epic:** #3 JSONL Parser Migration

## Usage

### Prerequisites

Install quicktype:
```bash
npm install -g quicktype
```

### Basic Usage

```bash
# Using make target (recommended)
make generate-parsers

# Or run script directly
./tools/generate-parsers.sh
```

This will:
1. Check for quicktype installation
2. Find all JSON schemas in `schemas/` directory
3. Generate Go structs in `internal/parser/generated/`
4. Format code with `gofmt`
5. Verify compilation

### Environment Variables

Configure the script using environment variables:

```bash
# Custom schemas directory
SCHEMAS_DIR=./my-schemas make generate-parsers

# Custom output directory
OUTPUT_DIR=./custom-output make generate-parsers

# Custom package name
PACKAGE_NAME=mypackage make generate-parsers
```

### Examples

```bash
# Generate parsers with default settings
make generate-parsers

# Generate to custom location
OUTPUT_DIR=/tmp/parsers ./tools/generate-parsers.sh

# Use custom schemas directory
SCHEMAS_DIR=./test-schemas ./tools/generate-parsers.sh
```

## Features

- ✅ **Batch Processing**: Processes all JSON schemas in one run
- ✅ **CamelCase Naming**: Generates Go-idiomatic struct names
- ✅ **JSON Tags**: Automatically adds JSON struct tags
- ✅ **Error Handling**: Validates prerequisites and reports failures
- ✅ **Code Formatting**: Automatically runs gofmt on generated code
- ✅ **Compilation Check**: Verifies generated code compiles
- ✅ **Colored Output**: Easy-to-read logs with color coding

## Output Structure

Generated files are placed in `internal/parser/generated/` by default:

```
internal/parser/generated/
├── types.go          # TypeRow struct from schemas/types.json
├── groups.go         # GroupRow struct from schemas/groups.json
├── blueprints.go     # BlueprintRow struct from schemas/blueprints.json
└── ...               # One .go file per schema
```

Each generated file contains:
- Type-safe Go structs with JSON tags
- Unmarshal/Marshal helper functions
- "DO NOT EDIT" comment header

## Example Generated Code

Input schema (`schemas/types.json`):
```json
{
  "typeID": 34,
  "groupID": 18,
  "typeName": {
    "en": "Tritanium",
    "de": "Tritanium"
  },
  "mass": 0.01,
  "published": true
}
```

Generated Go code (`internal/parser/generated/types.go`):
```go
// Code generated from JSON Schema using quicktype. DO NOT EDIT.
package generated

type Types struct {
    TypeID    int64       `json:"typeID"`
    GroupID   int64       `json:"groupID"`
    TypeName  Description `json:"typeName"`
    Mass      float64     `json:"mass"`
    Published bool        `json:"published"`
}

type Description struct {
    En string `json:"en"`
    De string `json:"de"`
}
```

## Workflow

Typical workflow for generating parsers:

1. **Prepare Schemas**: Ensure `schemas/` directory contains JSON sample data
   ```bash
   # Option 1: Run schema scraper
   go run tools/scrape-rift-schemas.go
   
   # Option 2: Manually create sample JSON files
   # Place them in schemas/ directory
   ```

2. **Generate Parsers**:
   ```bash
   make generate-parsers
   ```

3. **Verify Output**:
   ```bash
   ls -l internal/parser/generated/
   go build ./internal/parser/generated/...
   ```

4. **Use Generated Structs** in your code:
   ```go
   import "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser/generated"
   
   data, _ := os.ReadFile("types.jsonl")
   types, err := generated.UnmarshalTypes(data)
   ```

## Error Handling

The script performs several validation checks:

- ✅ **quicktype installed**: Exits with error if not found
- ✅ **Schemas directory exists**: Exits if `schemas/` not found
- ✅ **JSON files present**: Exits if no `.json` files found
- ✅ **Compilation check**: Warns if generated code doesn't compile

Error messages are color-coded (red) for easy identification.

## Integration with CI/CD

The script can be integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions
- name: Install quicktype
  run: npm install -g quicktype

- name: Generate parsers
  run: make generate-parsers

- name: Verify generated code
  run: go build ./internal/parser/generated/...
```

## Notes

- Generated files are in `.gitignore` (regenerated on demand)
- Schemas directory should be committed to version control
- Uses `quicktype` for reliable, battle-tested code generation
- Supports 50+ EVE SDE tables when full schemas are available
