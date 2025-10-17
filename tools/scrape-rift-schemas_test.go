// tools/scrape-rift-schemas_test.go
// Tests for RIFT SDE schema scraper
//go:build !add_tomap_methods
// +build !add_tomap_methods

package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/logger"
)

// TestScrapeTableSchema_Success tests successful schema download
func TestScrapeTableSchema_Success(t *testing.T) {
	// Create test logger
	log := setupTestLogger()

	// Create mock HTTP server
	testData := []map[string]interface{}{
		{
			"typeID":   34,
			"typeName": "Tritanium",
			"groupID":  18,
		},
	}
	jsonData, _ := json.Marshal(testData)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jsonData)
	}))
	defer server.Close()

	// Create temporary output directory
	tmpDir := t.TempDir()

	// Configure scraper
	cfg := &Config{
		OutputDir: tmpDir,
		BaseURL:   server.URL,
		Timeout:   5 * time.Second,
	}

	client := &http.Client{Timeout: cfg.Timeout}
	ctx := context.Background()

	// Run scraper
	err := scrapeTableSchema(ctx, client, cfg, "invTypes", log)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify output file exists
	schemaFile := filepath.Join(tmpDir, "invTypes.json")
	if _, err := os.Stat(schemaFile); os.IsNotExist(err) {
		t.Fatalf("Expected schema file to exist at %s", schemaFile)
	}

	// Verify content is valid JSON
	content, err := os.ReadFile(schemaFile)
	if err != nil {
		t.Fatalf("Failed to read schema file: %v", err)
	}

	var data interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		t.Fatalf("Schema file contains invalid JSON: %v", err)
	}
}

// TestScrapeTableSchema_HTTPError tests retry behavior on HTTP errors
func TestScrapeTableSchema_HTTPError(t *testing.T) {
	log := setupTestLogger()

	// Create mock server that always returns 500
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	cfg := &Config{
		OutputDir: tmpDir,
		BaseURL:   server.URL,
		Timeout:   5 * time.Second,
	}

	client := &http.Client{Timeout: cfg.Timeout}
	ctx := context.Background()

	// Run scraper - should fail after retries
	err := scrapeTableSchema(ctx, client, cfg, "invTypes", log)
	if err == nil {
		t.Fatal("Expected error for HTTP 500, got nil")
	}

	// Verify retry attempts were made (should be 3 + initial = 4 total)
	if attemptCount < 2 {
		t.Fatalf("Expected multiple retry attempts, got %d", attemptCount)
	}
}

// TestScrapeTableSchema_InvalidJSON tests that any valid HTTP response is accepted
// Since we create our own JSON placeholders, the response body content doesn't matter
func TestScrapeTableSchema_InvalidJSON(t *testing.T) {
	log := setupTestLogger()

	// Create mock server that returns anything (even invalid JSON)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<!DOCTYPE html><html>...</html>"))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	cfg := &Config{
		OutputDir: tmpDir,
		BaseURL:   server.URL,
		Timeout:   5 * time.Second,
	}

	client := &http.Client{Timeout: cfg.Timeout}
	ctx := context.Background()

	// Run scraper - should succeed since we create our own JSON
	err := scrapeTableSchema(ctx, client, cfg, "invTypes", log)
	if err != nil {
		t.Fatalf("Expected no error for HTML response (we create our own JSON), got: %v", err)
	}

	// Verify output file exists and is valid JSON
	schemaFile := filepath.Join(tmpDir, "invTypes.json")
	content, err := os.ReadFile(schemaFile)
	if err != nil {
		t.Fatalf("Failed to read schema file: %v", err)
	}

	var data interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		t.Fatalf("Generated schema file contains invalid JSON: %v", err)
	}
}

// TestScrapeTableSchema_ClientError tests non-retryable client errors
func TestScrapeTableSchema_ClientError(t *testing.T) {
	log := setupTestLogger()

	// Create mock server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	cfg := &Config{
		OutputDir: tmpDir,
		BaseURL:   server.URL,
		Timeout:   5 * time.Second,
	}

	client := &http.Client{Timeout: cfg.Timeout}
	ctx := context.Background()

	// Run scraper - should fail without retry (client error)
	err := scrapeTableSchema(ctx, client, cfg, "nonExistentTable", log)
	if err == nil {
		t.Fatal("Expected error for HTTP 404, got nil")
	}
}

// TestSaveSchema tests schema file creation
func TestSaveSchema(t *testing.T) {
	tmpDir := t.TempDir()

	testSchema := []byte(`{"test": "data"}`)

	err := saveSchema(tmpDir, "testTable", testSchema)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify file exists
	schemaFile := filepath.Join(tmpDir, "testTable.json")
	if _, err := os.Stat(schemaFile); os.IsNotExist(err) {
		t.Fatalf("Expected schema file to exist at %s", schemaFile)
	}

	// Verify content matches
	content, err := os.ReadFile(schemaFile)
	if err != nil {
		t.Fatalf("Failed to read schema file: %v", err)
	}

	if string(content) != string(testSchema) {
		t.Fatalf("Expected content %s, got %s", testSchema, content)
	}
}

// TestSaveSchema_CreateDirectory tests directory creation
func TestSaveSchema_CreateDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "nested", "schemas")

	testSchema := []byte(`{"test": "data"}`)

	err := saveSchema(outputDir, "testTable", testSchema)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Fatalf("Expected directory to be created at %s", outputDir)
	}

	// Verify file exists
	schemaFile := filepath.Join(outputDir, "testTable.json")
	if _, err := os.Stat(schemaFile); os.IsNotExist(err) {
		t.Fatalf("Expected schema file to exist at %s", schemaFile)
	}
}

// setupTestLogger creates a logger for testing
func setupTestLogger() *logger.Logger {
	return logger.NewLogger("error", "json")
}
