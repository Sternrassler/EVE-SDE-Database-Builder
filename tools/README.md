# RIFT Schema Scraper Tool

## Purpose

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

After running this tool, the placeholders confirm schema page accessibility. For actual code generation:
1. Parse RIFT HTML schema documentation to extract field types and structures
2. OR download CCP JSONL data and extract sample objects
3. Use `quicktype` or similar tools to generate Go structs from sample JSON
4. Implement JSONL parser using generated types
