# Logger Package

A structured logging wrapper around [zerolog](https://github.com/rs/zerolog) with configurable log levels and output formats.

## Features

- Configurable log levels (Debug, Info, Warn, Error, Fatal)
- Format selection (JSON for production, Text for development)
- Context-aware logging with support for RequestID and UserID
- Global logger singleton or dependency injection
- Structured logging with custom fields
- High test coverage (>86%)

## Installation

The logger package is part of the EVE-SDE-Database-Builder internal packages and uses zerolog v1.34.0.

## Usage

### Basic Usage

```go
package main

import (
    "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/logger"
)

func main() {
    // Create a logger with JSON format and info level
    log := logger.NewLogger("info", "json")
    log.Info("Application started", logger.Field{Key: "version", Value: "0.1.0"})
}
```

### Development vs Production

```go
// Development: Use text format for human-readable output
devLogger := logger.NewLogger("debug", "text")
devLogger.Debug("Debugging info", logger.Field{Key: "module", Value: "auth"})

// Production: Use JSON format for structured logging
prodLogger := logger.NewLogger("info", "json")
prodLogger.Info("User logged in", logger.Field{Key: "user_id", Value: 12345})
```

### Using the Global Logger

```go
// Set up global logger once at application startup
logger.SetGlobalLogger(logger.NewLogger("info", "json"))

// Use it anywhere in your application
log := logger.GetGlobalLogger()
log.Info("Processing request")
```

### Context-Aware Logging

```go
import (
    "context"
    "github.com/Sternrassler/EVE-SDE-Database-Builder/internal/logger"
)

// Add context values
ctx := context.Background()
ctx = context.WithValue(ctx, logger.RequestIDKey, "req-12345")
ctx = context.WithValue(ctx, logger.UserIDKey, "user-67890")

// Create a logger with context
log := logger.NewLogger("info", "json")
ctxLog := log.WithContext(ctx)

// All logs will include RequestID and UserID
ctxLog.Info("Request completed", logger.Field{Key: "duration_ms", Value: 150})
// Output: {"level":"info","RequestID":"req-12345","UserID":"user-67890","duration_ms":150,"time":"...","message":"Request completed"}
```

### Multiple Fields

```go
log.Info("User action",
    logger.Field{Key: "user", Value: "admin"},
    logger.Field{Key: "action", Value: "delete"},
    logger.Field{Key: "resource", Value: "user-123"},
)
```

## Log Levels

The following log levels are supported:
- `debug` - Detailed information for debugging
- `info` - General informational messages
- `warn` or `warning` - Warning messages
- `error` - Error messages
- `fatal` - Fatal errors (will exit the program)

## Output Formats

### JSON Format
```json
{"level":"info","version":"0.1.0","time":"2025-10-15T19:56:16Z","message":"Application started"}
```

### Text Format
```
7:56PM INF Application started version=0.1.0
```

## Testing

Run the tests with:
```bash
make test
```

Check test coverage:
```bash
make coverage
```

## Linting

Run the linter:
```bash
make lint
```

## Example

See `cmd/example/main.go` for a complete example demonstrating all features.

Run it with:
```bash
go run cmd/example/main.go
```
