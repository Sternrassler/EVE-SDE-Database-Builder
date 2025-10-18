package testutil_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/logger"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/testutil"
)

// Example_httpMockBasic demonstrates basic HTTP client mocking.
func Example_httpMockBasic() {
	// Create a mock HTTP client that returns a fixed response
	client := testutil.StaticMockClient(200, `{"status":"ok"}`)

	resp, err := client.Get("http://api.example.com/health")
	if err != nil {
		panic(err)
	}
	defer func() { _ = resp.Body.Close() }()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	// Output: Status: 200
}

// Example_httpMockDynamic demonstrates dynamic HTTP response mocking.
func Example_httpMockDynamic() {
	// Create a mock client with custom logic
	client := testutil.MockHTTPClient(func(req *http.Request) (*http.Response, error) {
		// Respond differently based on URL path
		switch req.URL.Path {
		case "/users":
			return testutil.MockJSONResponse(200, `[{"id":1,"name":"Alice"}]`), nil
		case "/error":
			return testutil.MockResponse(500, "Internal Server Error"), nil
		default:
			return testutil.MockResponse(404, "Not Found"), nil
		}
	})

	// Test different endpoints
	resp1, err := client.Get("http://api.example.com/users")
	if err != nil {
		panic(err)
	}
	defer func() { _ = resp1.Body.Close() }()
	fmt.Printf("Users endpoint: %d\n", resp1.StatusCode)

	resp2, err := client.Get("http://api.example.com/error")
	if err != nil {
		panic(err)
	}
	defer func() { _ = resp2.Body.Close() }()
	fmt.Printf("Error endpoint: %d\n", resp2.StatusCode)

	resp3, err := client.Get("http://api.example.com/unknown")
	if err != nil {
		panic(err)
	}
	defer func() { _ = resp3.Body.Close() }()
	fmt.Printf("Unknown endpoint: %d\n", resp3.StatusCode)

	// Output:
	// Users endpoint: 200
	// Error endpoint: 500
	// Unknown endpoint: 404
}

// Example_httpMockRecorder demonstrates request recording for verification.
func Example_httpMockRecorder() {
	// Create a recorder that captures requests
	recorder := testutil.NewRequestRecorder(testutil.MockResponse(200, "ok"))
	client := recorder.Client()

	// Make multiple requests
	_, _ = client.Get("http://api.example.com/endpoint1")
	_, _ = client.Get("http://api.example.com/endpoint2")
	_, _ = client.Post("http://api.example.com/data", "application/json", nil)

	// Verify requests were made
	fmt.Printf("Total requests: %d\n", recorder.RequestCount())

	lastReq := recorder.LastRequest()
	fmt.Printf("Last method: %s\n", lastReq.Method)
	fmt.Printf("Last path: %s\n", lastReq.URL.Path)

	// Output:
	// Total requests: 3
	// Last method: POST
	// Last path: /data
}

// Example_dbMock demonstrates database mocking.
func Example_dbMock() {
	// Create a mock database
	db := testutil.NewMockDB()

	// Configure mock behavior for Exec
	db.ExecFunc = func(_ string, _ ...interface{}) (sql.Result, error) {
		return testutil.NewMockResult(1, 1), nil
	}

	// Use the mock database
	result, err := db.Exec("INSERT INTO users (name) VALUES (?)", "Alice")
	if err != nil {
		panic(err)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Rows affected: %d\n", rowsAffected)

	// Verify the query was called with correct arguments
	if len(db.ExecCalls) > 0 {
		call := db.ExecCalls[0]
		fmt.Printf("Query called: %s\n", call.Query)
		fmt.Printf("Arguments: %v\n", call.Args)
	}

	// Output:
	// Rows affected: 1
	// Query called: INSERT INTO users (name) VALUES (?)
	// Arguments: [Alice]
}

// Example_dbMockErrors demonstrates mocking database errors.
func Example_dbMockErrors() {
	db := testutil.NewMockDB()

	// Configure mock to return an error
	expectedErr := errors.New("unique constraint violation")
	db.ExecFunc = func(_ string, _ ...interface{}) (sql.Result, error) {
		return nil, expectedErr
	}

	// Try to execute a query
	_, err := db.Exec("INSERT INTO users (email) VALUES (?)", "duplicate@example.com")

	if err != nil {
		fmt.Printf("Error occurred: %s\n", err.Error())
	}

	// Output:
	// Error occurred: unique constraint violation
}

// Example_loggerStub demonstrates logger stubbing for tests.
func Example_loggerStub() {
	// Create a logger stub that captures messages
	log := testutil.NewLoggerStub()

	// Use the logger in your code
	log.Info("Application started")
	log.Debug("Processing request", logger.Field{Key: "request_id", Value: "123"})
	log.Error("Failed to connect", logger.Field{Key: "host", Value: "db.example.com"})

	// Verify logged messages
	fmt.Printf("Total messages: %d\n", log.MessageCount())
	fmt.Printf("Info messages: %d\n", log.InfoCount())
	fmt.Printf("Error messages: %d\n", log.ErrorCount())

	// Check specific messages
	if log.HasMessage("Application started") {
		fmt.Println("Found startup message")
	}

	// Output:
	// Total messages: 3
	// Info messages: 1
	// Error messages: 1
	// Found startup message
}

// Example_loggerSilent demonstrates silent logger for tests that don't verify logs.
func Example_loggerSilent() {
	// Create a silent logger (no overhead for recording)
	log := testutil.NewSilentLogger()

	// Log many messages without performance impact
	for i := 0; i < 1000; i++ {
		log.Info("Processing item")
		log.Debug("Details")
	}

	// Silent logger doesn't record messages
	fmt.Printf("Messages recorded: %d\n", log.MessageCount())

	// Output:
	// Messages recorded: 0
}

// ServiceWithDependencies demonstrates testing a service with multiple dependencies.
type ServiceWithDependencies struct {
	httpClient *http.Client
	db         testutil.DBInterface
	log        *testutil.LoggerStub
}

func (s *ServiceWithDependencies) FetchAndStore(_ context.Context, url string) error {
	// Fetch data from HTTP
	resp, err := s.httpClient.Get(url)
	if err != nil {
		s.log.Error("HTTP request failed", logger.Field{Key: "error", Value: err.Error()})
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		s.log.Warn("Non-200 status", logger.Field{Key: "status", Value: resp.StatusCode})
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	// Store in database
	_, err = s.db.Exec("INSERT INTO data (url, status) VALUES (?, ?)", url, resp.StatusCode)
	if err != nil {
		s.log.Error("Database insert failed", logger.Field{Key: "error", Value: err.Error()})
		return err
	}

	s.log.Info("Successfully fetched and stored", logger.Field{Key: "url", Value: url})
	return nil
}

// Example_integrationWithMocks demonstrates testing with all mocks together.
func Example_integrationWithMocks() {
	// Setup all mocks
	httpClient := testutil.MockHTTPClient(func(_ *http.Request) (*http.Response, error) {
		return testutil.MockJSONResponse(200, `{"data":"test"}`), nil
	})

	db := testutil.NewMockDB()
	db.ExecFunc = func(_ string, _ ...interface{}) (sql.Result, error) {
		return testutil.NewMockResult(1, 1), nil
	}

	log := testutil.NewLoggerStub()

	// Create service with mocked dependencies
	service := &ServiceWithDependencies{
		httpClient: httpClient,
		db:         db,
		log:        log,
	}

	// Test the service
	err := service.FetchAndStore(context.Background(), "http://api.example.com/data")
	if err != nil {
		panic(err)
	}

	// Verify behavior
	fmt.Printf("DB calls: %d\n", len(db.ExecCalls))
	fmt.Printf("Log info messages: %d\n", log.InfoCount())

	if log.HasMessage("Successfully fetched and stored") {
		fmt.Println("Success message logged")
	}

	// Output:
	// DB calls: 1
	// Log info messages: 1
	// Success message logged
}

// Example_errorScenarioTesting demonstrates testing error scenarios.
func Example_errorScenarioTesting() {
	// Test HTTP error
	httpClient := testutil.MockHTTPClient(func(_ *http.Request) (*http.Response, error) {
		return nil, errors.New("network timeout")
	})

	db := testutil.NewMockDB()
	log := testutil.NewLoggerStub()

	service := &ServiceWithDependencies{
		httpClient: httpClient,
		db:         db,
		log:        log,
	}

	err := service.FetchAndStore(context.Background(), "http://api.example.com/data")

	fmt.Printf("Error occurred: %v\n", err != nil)
	fmt.Printf("Error logs: %d\n", log.ErrorCount())

	// Verify database was not called (due to HTTP error)
	fmt.Printf("DB calls: %d\n", len(db.ExecCalls))

	// Output:
	// Error occurred: true
	// Error logs: 1
	// DB calls: 0
}
