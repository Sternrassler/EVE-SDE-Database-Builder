# JSONL Parser Package

This package provides a generic, type-safe parser interface for JSONL (JSON Lines) files used in the EVE SDE Database Builder.

## Features

- **Generic Parser Interface**: Type-safe parsing using Go generics
- **Line-by-Line Processing**: Memory-efficient parsing of large JSONL files
- **Context Support**: Cancellation and timeout support via `context.Context`
- **Error Handling**: Line-number-based error reporting for easy debugging
- **Large Line Support**: Handles JSON lines up to 10MB in size
- **Empty Line Handling**: Automatically skips empty lines

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

- Streaming mode for extremely large files (callback-based)
- Batch processing for memory-constrained environments
- Schema validation against JSON Schema
- Progress reporting for long-running parses
- Compressed file support (gzip, bzip2)

## References

- **ADR-003**: JSONL Parser Architecture (`docs/adr/ADR-003-jsonl-parser-architecture.md`)
- **JSON Lines**: https://jsonlines.org/
- **EVE SDE**: https://developers.eveonline.com/static-data
- **RIFT SDE Schema**: https://sde.riftforeve.online/

## See Also

- `internal/database` - Database layer for importing parsed data
- `internal/errors` - Error types for structured error handling
- `internal/logger` - Logging utilities for parser operations
