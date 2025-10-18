# Mock & Stub Utilities

Dieses Dokument beschreibt die Mock- und Stub-Utilities im `testutil` Package, die zur Test-Isolation und Vereinfachung von Unit-Tests entwickelt wurden.

## Übersicht

Das `testutil` Package bietet drei Hauptkategorien von Test-Utilities:

1. **HTTP Client Mocks** - Für Tests, die HTTP-Requests simulieren müssen
2. **Database Mocks** - Für Tests ohne echte Datenbankverbindung
3. **Logger Stubs** - Für Tests mit kontrollierter Log-Ausgabe

## HTTP Client Mocks

### MockHTTPClient

Erstellt einen HTTP-Client mit vollständig konfigurierbarem Verhalten.

```go
client := testutil.MockHTTPClient(func(req *http.Request) (*http.Response, error) {
    return testutil.MockResponse(200, `{"status":"ok"}`), nil
})
```

### StaticMockClient

Für einfache Test-Szenarien mit statischen Antworten.

```go
client := testutil.StaticMockClient(200, `{"data":"test"}`)
```

### RequestRecorder

Zeichnet alle HTTP-Requests auf zur späteren Verifikation.

```go
recorder := testutil.NewRequestRecorder(testutil.MockResponse(200, "ok"))
client := recorder.Client()

// Requests durchführen...
client.Get("http://example.com/api")

// Verifikation
fmt.Printf("Requests: %d\n", recorder.RequestCount())
lastReq := recorder.LastRequest()
```

### Helper-Funktionen

- `MockResponse(statusCode, body)` - Erstellt eine einfache HTTP-Response
- `MockJSONResponse(statusCode, jsonBody)` - Response mit JSON Content-Type
- `MockErrorResponse(err)` - Simuliert Request-Fehler

## Database Mocks

### MockDB

Mock-Implementierung des `DBInterface` für Datenbankoperationen.

```go
db := testutil.NewMockDB()

// Verhalten konfigurieren
db.ExecFunc = func(query string, args ...interface{}) (sql.Result, error) {
    return testutil.NewMockResult(1, 1), nil
}

// Verwenden
result, err := db.Exec("INSERT INTO users VALUES (?)", "Alice")

// Verifikation
fmt.Printf("Exec calls: %d\n", len(db.ExecCalls))
fmt.Printf("Query: %s\n", db.ExecCalls[0].Query)
```

### MockResult

Mock für `sql.Result` mit konfigurierbaren Rückgabewerten.

```go
result := testutil.NewMockResult(lastInsertID, rowsAffected)
errorResult := testutil.NewMockResultWithError(err)
```

### SQLXAdapter

Adapter um echte `*sqlx.DB` Instanzen durch das `DBInterface` zu verwenden.

```go
realDB := database.NewTestDB(t)
adapter := testutil.NewSQLXAdapter(realDB)

// adapter implementiert DBInterface
result, err := adapter.Exec("INSERT INTO test VALUES (?)", 1)
```

### Aufgezeichnete Daten

- `ExecCalls` - Alle Exec-Aufrufe mit Query und Argumenten
- `QueryCalls` - Alle Query-Aufrufe
- `QueryRowCalls` - Alle QueryRow-Aufrufe
- `PrepareCalls` - Alle Prepare-Aufrufe
- `Closed` - Ob Close aufgerufen wurde

## Logger Stubs

### LoggerStub

Zeichnet Log-Nachrichten auf zur Verifikation im Test.

```go
log := testutil.NewLoggerStub()

// Logging
log.Info("Application started")
log.Error("Connection failed", logger.Field{Key: "error", Value: "timeout"})

// Verifikation
fmt.Printf("Total messages: %d\n", log.MessageCount())
fmt.Printf("Errors: %d\n", log.ErrorCount())

if log.HasMessage("Application started") {
    // ...
}
```

### NewSilentLogger

Erstellt einen Logger ohne Aufzeichnung (für Performance).

```go
log := testutil.NewSilentLogger()

// Logging hat keinen Overhead
for i := 0; i < 10000; i++ {
    log.Info("Processing")
}

// Keine Aufzeichnung
fmt.Printf("Recorded: %d\n", log.MessageCount()) // 0
```

### Verfügbare Methoden

**Logging:**
- `Debug(msg, fields...)`
- `Info(msg, fields...)`
- `Warn(msg, fields...)`
- `Error(msg, fields...)`
- `Fatal(msg, fields...)` - Beendet NICHT das Programm im Test

**Verifikation:**
- `MessageCount()` - Gesamtzahl der Nachrichten
- `DebugCount()`, `InfoCount()`, `WarnCount()`, `ErrorCount()`, `FatalCount()`
- `Messages()` - Alle Nachrichten
- `MessagesAtLevel(level)` - Nachrichten eines Levels
- `HasMessage(text)` - Prüft ob Text vorkommt
- `HasMessageAtLevel(level, text)` - Prüft Text bei Level
- `ContainsMessage(substring)` - Prüft auf Teilstring
- `LastMessage()` - Letzte Nachricht
- `String()` - Lesbare Darstellung aller Logs
- `Reset()` - Löscht alle aufgezeichneten Nachrichten

## Verwendungsbeispiele

### Beispiel 1: HTTP Client Test

```go
func TestFetchData(t *testing.T) {
    // Setup mock client
    client := testutil.MockHTTPClient(func(req *http.Request) (*http.Response, error) {
        if req.URL.Path == "/data" {
            return testutil.MockJSONResponse(200, `{"id":1}`), nil
        }
        return testutil.MockResponse(404, "Not Found"), nil
    })

    // Test code that uses client
    fetcher := NewDataFetcher(client)
    data, err := fetcher.Fetch("/data")
    
    // Assertions
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    // ... weitere Prüfungen
}
```

### Beispiel 2: Database Test

```go
func TestInsertUser(t *testing.T) {
    // Setup mock database
    db := testutil.NewMockDB()
    
    var capturedQuery string
    db.ExecFunc = func(query string, args ...interface{}) (sql.Result, error) {
        capturedQuery = query
        return testutil.NewMockResult(42, 1), nil
    }

    // Test
    repo := NewUserRepository(db)
    id, err := repo.InsertUser("Alice")
    
    // Verify
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if id != 42 {
        t.Errorf("expected ID 42, got %d", id)
    }
    if !strings.Contains(capturedQuery, "INSERT") {
        t.Error("expected INSERT query")
    }
}
```

### Beispiel 3: Logger Test

```go
func TestProcessWithLogging(t *testing.T) {
    // Setup logger stub
    log := testutil.NewLoggerStub()

    // Test
    processor := NewProcessor(log)
    processor.Process()

    // Verify logging behavior
    if !log.HasMessage("Processing started") {
        t.Error("expected start message")
    }
    if log.ErrorCount() > 0 {
        t.Errorf("unexpected errors: %s", log.String())
    }
}
```

### Beispiel 4: Integration Test mit allen Mocks

```go
func TestServiceIntegration(t *testing.T) {
    // Setup all mocks
    httpClient := testutil.StaticMockClient(200, `{"status":"ok"}`)
    
    db := testutil.NewMockDB()
    db.ExecFunc = func(query string, args ...interface{}) (sql.Result, error) {
        return testutil.NewMockResult(1, 1), nil
    }
    
    log := testutil.NewLoggerStub()

    // Create service with mocked dependencies
    service := NewService(httpClient, db, log)
    
    // Test
    err := service.ProcessData()
    
    // Verify
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(db.ExecCalls) == 0 {
        t.Error("expected database call")
    }
    if log.InfoCount() == 0 {
        t.Error("expected info logs")
    }
}
```

## Best Practices

### 1. Mock-Konfiguration vor Verwendung

Konfiguriere Mock-Verhalten immer vor der Verwendung im Test:

```go
db := testutil.NewMockDB()
db.ExecFunc = func(...) { /* ... */ }  // DANN verwenden
```

### 2. Explizite Verifikation

Prüfe explizit, dass erwartete Aufrufe erfolgt sind:

```go
if len(db.ExecCalls) != 1 {
    t.Errorf("expected 1 exec call, got %d", len(db.ExecCalls))
}
```

### 3. Reset bei mehreren Subtests

Nutze `Reset()` zwischen Subtests:

```go
t.Run("Test1", func(t *testing.T) {
    log.Info("test")
    // ... Assertions
})

log.Reset()  // Zurücksetzen

t.Run("Test2", func(t *testing.T) {
    // ... nächster Test
})
```

### 4. Silent Logger für Performance-Tests

Verwende `NewSilentLogger()` wenn keine Log-Verifikation nötig ist:

```go
func BenchmarkProcess(b *testing.B) {
    log := testutil.NewSilentLogger()  // Kein Recording-Overhead
    processor := NewProcessor(log)
    
    for i := 0; i < b.N; i++ {
        processor.Process()
    }
}
```

### 5. RequestRecorder für API-Tests

Nutze `RequestRecorder` zur Verifikation von HTTP-Anfragen:

```go
recorder := testutil.NewRequestRecorder(testutil.MockResponse(200, "ok"))
client := recorder.Client()

// ... Teste Code der client verwendet

// Verify
if recorder.RequestCount() != expectedCount {
    t.Errorf("unexpected request count")
}
lastReq := recorder.LastRequest()
if lastReq.Method != "POST" {
    t.Error("expected POST request")
}
```

## Integration mit bestehendem Code

Die Mock-Utilities sind so konzipiert, dass sie nahtlos mit bestehendem Code funktionieren:

- **HTTP Client**: Jeder Code, der `*http.Client` akzeptiert, kann gemockt werden
- **Database**: Code muss `DBInterface` statt `*sql.DB` verwenden (oder `SQLXAdapter` nutzen)
- **Logger**: LoggerStub implementiert dasselbe Interface wie `logger.Logger`

## Weitere Informationen

- Siehe `example_mocks_test.go` für vollständige Beispiele
- Siehe Test-Dateien (`*_test.go`) für detaillierte Verwendung
- Package-Dokumentation: `go doc github.com/Sternrassler/EVE-SDE-Database-Builder/internal/testutil`
