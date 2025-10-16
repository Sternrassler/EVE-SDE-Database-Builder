# JSONL Parser Package

This package provides a generic, type-safe parser interface for JSONL (JSON Lines) files used in the EVE SDE Database Builder.

## Features

- **Generic Parser Interface**: Type-safe parsing using Go generics
- **Line-by-Line Processing**: Memory-efficient parsing of large JSONL files
- **Context Support**: Cancellation and timeout support via `context.Context`
- **Error Handling**: Line-number-based error reporting for easy debugging
- **Large Line Support**: Handles JSON lines up to 10MB in size
- **Empty Line Handling**: Automatically skips empty lines
- **Data Validation**: Built-in validation interface for parsed data with batch processing
- **Streaming API**: Memory-efficient streaming parser for large files

## Architecture

The package implements the Parser interface defined in ADR-003 (JSONL Parser Architecture):

```go
type Parser interface {
    ParseFile(ctx context.Context, path string) ([]interface{}, error)
    TableName() string
    Columns() []string
}
```

## Usage

### Creating a Parser

Define a struct that matches your JSONL schema and create a parser:

```go
import "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/parser"

// Define your data structure
type TypeRow struct {
    TypeID   int               `json:"typeID"`
    GroupID  int               `json:"groupID"`
    TypeName map[string]string `json:"typeName"`
    Mass     float64           `json:"mass,omitempty"`
}

// Create a parser
p := parser.NewJSONLParser[TypeRow](
    "invTypes",                                   // database table name
    []string{"typeID", "groupID", "typeName"},    // column names
)
```

### Parsing a File

```go
ctx := context.Background()
results, err := p.ParseFile(ctx, "types.jsonl")
if err != nil {
    log.Fatal(err)
}

// Process results
for _, result := range results {
    if row, ok := result.(TypeRow); ok {
        fmt.Printf("Type %d: %s\n", row.TypeID, row.TypeName["en"])
    }
}
```

### With Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

results, err := p.ParseFile(ctx, "large-file.jsonl")
if errors.Is(err, context.DeadlineExceeded) {
    log.Println("Parsing timeout")
}
```

### Accessing Metadata

```go
tableName := p.TableName()  // "invTypes"
columns := p.Columns()      // ["typeID", "groupID", "typeName"]
```

## Data Validation

The parser package includes a `Validator` interface for validating parsed data with required fields, ranges, and format constraints.

### Implementing Validation

Implement the `Validator` interface on your data structures:

```go
type TypeRow struct {
    TypeID   int               `json:"typeID"`
    TypeName map[string]string `json:"typeName"`
    Mass     float64           `json:"mass,omitempty"`
}

// Validate implements the Validator interface
func (t TypeRow) Validate() error {
    if t.TypeID <= 0 {
        return fmt.Errorf("typeID must be positive, got %d", t.TypeID)
    }
    if len(t.TypeName) == 0 {
        return errors.New("typeName is required")
    }
    if _, ok := t.TypeName["en"]; !ok {
        return errors.New("typeName must contain English translation")
    }
    if t.Mass < 0 {
        return fmt.Errorf("mass cannot be negative, got %f", t.Mass)
    }
    return nil
}
```

### Batch Validation

Use `ValidateBatch` to filter invalid items and collect all validation errors:

```go
// Parse data
results, err := p.ParseFile(ctx, "types.jsonl")
if err != nil {
    log.Fatal(err)
}

// Convert to typed slice
items := make([]TypeRow, len(results))
for i, result := range results {
    items[i] = result.(TypeRow)
}

// Validate all items
validItems, errs := parser.ValidateBatch(items)

fmt.Printf("Valid: %d, Invalid: %d\n", len(validItems), len(errs))

// Process validation errors
for _, err := range errs {
    log.Printf("Validation error: %v\n", err)
}

// Use only valid items
for _, item := range validItems {
    // Process valid item
}
```

### Validation Features

- **Required Field Checks**: Validate presence and non-empty values
- **Range Validation**: Ensure numeric values are within valid ranges
- **Format Validation**: Check nested structures, maps, and custom formats
- **Batch Processing**: Validate all items and collect all errors at once
- **Error Context**: Each validation error includes the item index for debugging

## Error Handling

Errors include line numbers for easy debugging:

```jsonl
{"typeID":1,"name":"Valid"}
{"typeID":2,"name":"Invalid JSON
{"typeID":3,"name":"Valid"}
```

Will produce an error like:
```
line 2: failed to parse JSON: unexpected end of JSON input
```

## Testing

The package includes comprehensive tests:

```bash
# Run all tests
go test ./internal/parser/...

# Run with coverage
go test ./internal/parser/... -cover

# Run benchmarks
go test ./internal/parser/... -bench=.
```

## Performance

The parser is optimized for large files:

- **Buffer Size**: 1MB initial buffer, 10MB maximum line size
- **Memory**: Efficient line-by-line processing
- **Concurrency**: Safe for concurrent use (each parser maintains no state)

Benchmark results (100 lines):
```
BenchmarkJSONLParser_ParseFile_SmallFile-4   6033   201740 ns/op   1080560 B/op   713 allocs/op
```

## Integration with ADR-003

This implementation follows the architecture defined in `docs/adr/ADR-003-jsonl-parser-architecture.md`:

- ✅ Generic `Parser` interface
- ✅ Type-safe parsing with Go generics
- ✅ Line-by-line error handling
- ✅ Context support for cancellation
- ✅ Metadata methods (`TableName`, `Columns`)

## Future Enhancements

Potential improvements for future versions:

- Batch processing for memory-constrained environments
- Progress reporting for long-running parses
- Compressed file support (gzip, bzip2)
- Schema auto-generation from JSON Schema
- Advanced validation rules (cross-field validation, custom validators)

## References

- **ADR-003**: JSONL Parser Architecture (`docs/adr/ADR-003-jsonl-parser-architecture.md`)
- **JSON Lines**: https://jsonlines.org/
- **EVE SDE**: https://developers.eveonline.com/static-data
- **RIFT SDE Schema**: https://sde.riftforeve.online/

## See Also

- `internal/database` - Database layer for importing parsed data
- `internal/errors` - Error types for structured error handling
- `internal/logger` - Logging utilities for parser operations
