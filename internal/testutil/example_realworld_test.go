package testutil_test

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/logger"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/testutil"
)

// RealWorldAPIClient demonstrates a typical HTTP API client that needs mocking.
type RealWorldAPIClient struct {
	httpClient *http.Client
	baseURL    string
	log        *testutil.LoggerStub
}

func NewRealWorldAPIClient(client *http.Client, baseURL string, log *testutil.LoggerStub) *RealWorldAPIClient {
	return &RealWorldAPIClient{
		httpClient: client,
		baseURL:    baseURL,
		log:        log,
	}
}

func (c *RealWorldAPIClient) FetchUserData(userID int) (string, error) {
	url := fmt.Sprintf("%s/users/%d", c.baseURL, userID)
	c.log.Debug("Fetching user data", logger.Field{Key: "url", Value: url})

	resp, err := c.httpClient.Get(url)
	if err != nil {
		c.log.Error("HTTP request failed", logger.Field{Key: "error", Value: err.Error()})
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		c.log.Warn("Non-200 status", logger.Field{Key: "status", Value: resp.StatusCode})
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Error("Failed to read response", logger.Field{Key: "error", Value: err.Error()})
		return "", err
	}

	c.log.Info("User data fetched successfully", logger.Field{Key: "user_id", Value: userID})
	return string(body), nil
}

// Example_realWorldAPIClientTesting demonstrates testing an HTTP API client with mocks.
func Example_realWorldAPIClientTesting() {
	// Setup: Create a mock HTTP client that returns user data
	httpClient := testutil.MockHTTPClient(func(req *http.Request) (*http.Response, error) {
		// Verify request path
		if req.URL.Path == "/users/123" {
			return testutil.MockJSONResponse(200, `{"id":123,"name":"Alice"}`), nil
		}
		return testutil.MockResponse(404, "Not Found"), nil
	})

	// Setup: Create a logger stub to verify logging behavior
	log := testutil.NewLoggerStub()

	// Setup: Create client under test
	client := NewRealWorldAPIClient(httpClient, "http://api.example.com", log)

	// Test: Fetch user data
	data, err := client.FetchUserData(123)
	if err != nil {
		panic(err)
	}

	// Verify: Check response
	fmt.Printf("Received data: %s\n", data)

	// Verify: Check logging behavior
	fmt.Printf("Debug messages: %d\n", log.DebugCount())
	fmt.Printf("Info messages: %d\n", log.InfoCount())

	if log.HasMessage("User data fetched successfully") {
		fmt.Println("Success message logged")
	}

	// Output:
	// Received data: {"id":123,"name":"Alice"}
	// Debug messages: 1
	// Info messages: 1
	// Success message logged
}

// RealWorldRepository demonstrates a typical database repository that needs mocking.
type RealWorldRepository struct {
	db  testutil.DBInterface
	log *testutil.LoggerStub
}

func NewRealWorldRepository(db testutil.DBInterface, log *testutil.LoggerStub) *RealWorldRepository {
	return &RealWorldRepository{
		db:  db,
		log: log,
	}
}

func (r *RealWorldRepository) InsertUser(name, email string) (int64, error) {
	r.log.Debug("Inserting user", logger.Field{Key: "name", Value: name})

	result, err := r.db.Exec(
		"INSERT INTO users (name, email) VALUES (?, ?)",
		name, email,
	)
	if err != nil {
		r.log.Error("Insert failed", logger.Field{Key: "error", Value: err.Error()})
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	r.log.Info("User inserted", logger.Field{Key: "id", Value: id})
	return id, nil
}

func (r *RealWorldRepository) GetUserCount() (int, error) {
	r.log.Debug("Getting user count")

	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		r.log.Error("Query failed", logger.Field{Key: "error", Value: err.Error()})
		return 0, err
	}

	return count, nil
}

// Example_realWorldRepositoryTesting demonstrates testing a database repository with mocks.
func Example_realWorldRepositoryTesting() {
	// Setup: Create a mock database
	db := testutil.NewMockDB()

	// Setup: Configure mock behavior for Exec
	db.ExecFunc = func(_ string, _ ...interface{}) (sql.Result, error) {
		// Simulate successful insert with ID 42
		return testutil.NewMockResult(42, 1), nil
	}

	// Setup: Create logger stub
	log := testutil.NewLoggerStub()

	// Setup: Create repository under test
	repo := NewRealWorldRepository(db, log)

	// Test: Insert user
	id, err := repo.InsertUser("Alice", "alice@example.com")
	if err != nil {
		panic(err)
	}

	// Verify: Check returned ID
	fmt.Printf("Inserted user ID: %d\n", id)

	// Verify: Check database was called correctly
	if len(db.ExecCalls) != 1 {
		panic("Expected 1 Exec call")
	}

	call := db.ExecCalls[0]
	fmt.Printf("Query: %s\n", call.Query)
	fmt.Printf("Name arg: %v\n", call.Args[0])
	fmt.Printf("Email arg: %v\n", call.Args[1])

	// Verify: Check logging
	fmt.Printf("Debug logs: %d\n", log.DebugCount())
	fmt.Printf("Info logs: %d\n", log.InfoCount())

	// Output:
	// Inserted user ID: 42
	// Query: INSERT INTO users (name, email) VALUES (?, ?)
	// Name arg: Alice
	// Email arg: alice@example.com
	// Debug logs: 1
	// Info logs: 1
}

// RealWorldService demonstrates a service that combines multiple dependencies.
type RealWorldService struct {
	apiClient *RealWorldAPIClient
	repo      *RealWorldRepository
	log       *testutil.LoggerStub
}

func NewRealWorldService(apiClient *RealWorldAPIClient, repo *RealWorldRepository, log *testutil.LoggerStub) *RealWorldService {
	return &RealWorldService{
		apiClient: apiClient,
		repo:      repo,
		log:       log,
	}
}

func (s *RealWorldService) SyncUser(_ context.Context, userID int) error {
	s.log.Info("Starting user sync", logger.Field{Key: "user_id", Value: userID})

	// Fetch from API
	userData, err := s.apiClient.FetchUserData(userID)
	if err != nil {
		s.log.Error("Failed to fetch user", logger.Field{Key: "error", Value: err.Error()})
		return err
	}

	// Store in database (simplified - in real code would parse JSON)
	_, err = s.repo.InsertUser("user", "user@example.com")
	if err != nil {
		s.log.Error("Failed to store user", logger.Field{Key: "error", Value: err.Error()})
		return err
	}

	s.log.Info("User sync completed", logger.Field{Key: "user_id", Value: userID}, logger.Field{Key: "data_size", Value: len(userData)})
	return nil
}

// Example_realWorldServiceIntegrationTesting demonstrates integration testing with all mocks.
func Example_realWorldServiceIntegrationTesting() {
	// Setup HTTP mock
	httpClient := testutil.MockHTTPClient(func(_ *http.Request) (*http.Response, error) {
		return testutil.MockJSONResponse(200, `{"id":123,"name":"Alice","email":"alice@example.com"}`), nil
	})

	// Setup DB mock
	db := testutil.NewMockDB()
	db.ExecFunc = func(_ string, _ ...interface{}) (sql.Result, error) {
		return testutil.NewMockResult(1, 1), nil
	}

	// Setup logger
	log := testutil.NewLoggerStub()

	// Create dependencies
	apiClient := NewRealWorldAPIClient(httpClient, "http://api.example.com", log)
	repo := NewRealWorldRepository(db, log)

	// Create service
	service := NewRealWorldService(apiClient, repo, log)

	// Test
	err := service.SyncUser(context.Background(), 123)
	if err != nil {
		panic(err)
	}

	// Verify: Check all components worked together
	fmt.Printf("API client logged debug: %v\n", log.ContainsMessage("Fetching user data"))
	fmt.Printf("Repository logged debug: %v\n", log.ContainsMessage("Inserting user"))
	fmt.Printf("Service logged completion: %v\n", log.ContainsMessage("User sync completed"))

	// Verify: Database operations
	fmt.Printf("Database inserts: %d\n", len(db.ExecCalls))

	// Verify: Final state
	fmt.Printf("Total info logs: %d\n", log.InfoCount())
	fmt.Printf("Total error logs: %d\n", log.ErrorCount())

	// Output:
	// API client logged debug: true
	// Repository logged debug: true
	// Service logged completion: true
	// Database inserts: 1
	// Total info logs: 4
	// Total error logs: 0
}

// Example_realWorldErrorHandling demonstrates testing error scenarios with mocks.
func Example_realWorldErrorHandling() {
	// Setup: HTTP client that fails
	httpClient := testutil.MockHTTPClient(func(_ *http.Request) (*http.Response, error) {
		return testutil.MockResponse(500, "Internal Server Error"), nil
	})

	// Setup: DB mock
	db := testutil.NewMockDB()

	// Setup: Logger
	log := testutil.NewLoggerStub()

	// Create components
	apiClient := NewRealWorldAPIClient(httpClient, "http://api.example.com", log)
	repo := NewRealWorldRepository(db, log)
	service := NewRealWorldService(apiClient, repo, log)

	// Test: This should fail due to HTTP 500
	err := service.SyncUser(context.Background(), 123)

	// Verify: Error was returned
	fmt.Printf("Error occurred: %v\n", err != nil)

	// Verify: Warning was logged for non-200 status
	fmt.Printf("Warning logs: %d\n", log.WarnCount())

	// Verify: Error was logged
	fmt.Printf("Error logs: %d\n", log.ErrorCount())

	// Verify: Database was NOT called (due to API failure)
	fmt.Printf("Database inserts: %d\n", len(db.ExecCalls))

	// Output:
	// Error occurred: true
	// Warning logs: 1
	// Error logs: 1
	// Database inserts: 0
}

// Example_realWorldPerformanceTesting demonstrates performance testing with silent logger.
func Example_realWorldPerformanceTesting() {
	// Setup: Use silent logger for performance tests (no overhead)
	log := testutil.NewSilentLogger()

	// Setup: Fast mock responses
	httpClient := testutil.StaticMockClient(200, `{"data":"ok"}`)
	
	db := testutil.NewMockDB()
	db.ExecFunc = func(_ string, _ ...interface{}) (sql.Result, error) {
		return testutil.NewMockResult(1, 1), nil
	}

	// Create components
	apiClient := NewRealWorldAPIClient(httpClient, "http://api.example.com", log)
	repo := NewRealWorldRepository(db, log)
	service := NewRealWorldService(apiClient, repo, log)

	// Simulate performance test (many operations)
	for i := 0; i < 100; i++ {
		_ = service.SyncUser(context.Background(), i)
	}

	// Verify: Operations completed
	fmt.Printf("Operations completed: 100\n")
	fmt.Printf("Database operations: %d\n", len(db.ExecCalls))

	// Verify: No logs recorded (silent logger)
	fmt.Printf("Logs recorded: %d\n", log.MessageCount())

	// Output:
	// Operations completed: 100
	// Database operations: 100
	// Logs recorded: 0
}
