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

This will download all 50+ EVE SDE table schemas to the `schemas/` directory.

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
- **50+ Tables**: Scrapes complete EVE SDE schema including:
  - Inventory (invTypes, invGroups, etc.)
  - Industry/Blueprints
  - Dogma (ship fitting system)
  - Universe/Map data
  - Character/NPC data
  - And more...

## Output

Schema files are saved as JSON in the specified output directory:
- `schemas/invTypes.json`
- `schemas/invGroups.json`
- `schemas/dogmaAttributes.json`
- etc.

Each file contains a sample JSON response from the RIFT API that includes:
- Field names and types
- Example data structure
- Nested objects and arrays

## Testing

Run tests:

```bash
go test -v ./tools/...
```

Tests include:
- Mock HTTP client test for successful downloads
- Retry behavior on HTTP errors
- Invalid JSON handling
- Client error (4xx) handling
- File I/O tests

## Implementation Details

- Uses `internal/retry` package for HTTP retry logic with exponential backoff
- Uses `internal/logger` package for structured logging
- Uses `internal/errors` package for typed error handling (Retryable, Fatal, Validation)
- HTTP timeout: 30s (configurable)
- Retry policy: 3 attempts with 100ms-5s backoff

## Next Steps

After running this tool, the downloaded schemas can be used for:
1. Code generation with `quicktype` or similar tools
2. Type-safe struct generation for Go
3. Database schema validation
4. API contract documentation
