package testutil_test

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/testutil"
)

func TestMockHTTPClient(t *testing.T) {
	t.Parallel()
	expectedBody := `{"test": "data"}`

	client := testutil.MockHTTPClient(func(_ *http.Request) (*http.Response, error) {
		return testutil.MockResponse(200, expectedBody), nil
	})

	resp, err := client.Get("http://example.com/api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	if string(body) != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, string(body))
	}
}

func TestMockResponse(t *testing.T) {
	t.Parallel()
	resp := testutil.MockResponse(404, "not found")

	if resp.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", resp.StatusCode)
	}

	if resp.Status != "Not Found" {
		t.Errorf("expected status text 'Not Found', got %q", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	if string(body) != "not found" {
		t.Errorf("expected body 'not found', got %q", string(body))
	}
}

func TestMockJSONResponse(t *testing.T) {
	t.Parallel()
	jsonBody := `{"id": 123, "name": "test"}`
	resp := testutil.MockJSONResponse(200, jsonBody)

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", contentType)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	if string(body) != jsonBody {
		t.Errorf("expected body %q, got %q", jsonBody, string(body))
	}
}

func TestMockErrorResponse(t *testing.T) {
	t.Parallel()
	expectedErr := errors.New("connection timeout")

	client := testutil.MockHTTPClient(func(_ *http.Request) (*http.Response, error) {
		return testutil.MockErrorResponse(expectedErr)
	})

	resp, err := client.Get("http://example.com/api")
	if err == nil {
		t.Error("expected error, got nil")
	}

	if resp != nil {
		t.Error("expected nil response, got non-nil")
	}

	// Error should be wrapped but contain the original error
	if !strings.Contains(err.Error(), expectedErr.Error()) {
		t.Errorf("expected error to contain %q, got %v", expectedErr.Error(), err)
	}
}

func TestStaticMockClient(t *testing.T) {
	t.Parallel()
	expectedBody := "static response"
	client := testutil.StaticMockClient(201, expectedBody)

	// Make multiple requests to verify static behavior
	for i := 0; i < 3; i++ {
		resp, err := client.Get("http://example.com/api")
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i, err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != 201 {
			t.Errorf("request %d: expected status 201, got %d", i, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("request %d: failed to read body: %v", i, err)
		}

		if string(body) != expectedBody {
			t.Errorf("request %d: expected body %q, got %q", i, expectedBody, string(body))
		}
	}
}

func TestRequestRecorder_Basic(t *testing.T) {
	t.Parallel()
	recorder := testutil.NewRequestRecorder(testutil.MockResponse(200, "ok"))
	client := recorder.Client()

	// Make a request
	resp, err := client.Get("http://example.com/api/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Verify request was recorded
	if recorder.RequestCount() != 1 {
		t.Errorf("expected 1 request, got %d", recorder.RequestCount())
	}

	lastReq := recorder.LastRequest()
	if lastReq == nil {
		t.Fatal("expected non-nil last request")
	}

	if lastReq.URL.String() != "http://example.com/api/test" {
		t.Errorf("expected URL 'http://example.com/api/test', got %q", lastReq.URL.String())
	}

	if lastReq.Method != "GET" {
		t.Errorf("expected method 'GET', got %q", lastReq.Method)
	}
}

func TestRequestRecorder_MultipleRequests(t *testing.T) {
	t.Parallel()
	recorder := testutil.NewRequestRecorder(testutil.MockResponse(200, "ok"))
	client := recorder.Client()

	// Make multiple requests
	urls := []string{
		"http://example.com/api/one",
		"http://example.com/api/two",
		"http://example.com/api/three",
	}

	for _, url := range urls {
		resp, err := client.Get(url)
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", url, err)
		}
		_ = resp.Body.Close()
	}

	// Verify all requests were recorded
	if recorder.RequestCount() != len(urls) {
		t.Errorf("expected %d requests, got %d", len(urls), recorder.RequestCount())
	}

	// Verify last request
	lastReq := recorder.LastRequest()
	if lastReq == nil {
		t.Fatal("expected non-nil last request")
	}

	expectedLastURL := urls[len(urls)-1]
	if lastReq.URL.String() != expectedLastURL {
		t.Errorf("expected last URL %q, got %q", expectedLastURL, lastReq.URL.String())
	}
}

func TestRequestRecorder_WithBody(t *testing.T) {
	t.Parallel()
	recorder := testutil.NewRequestRecorder(testutil.MockResponse(201, "created"))
	client := recorder.Client()

	// Make POST request with body
	requestBody := `{"name": "test", "value": 123}`
	resp, err := client.Post("http://example.com/api/create", "application/json", strings.NewReader(requestBody))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Verify request was recorded
	if recorder.RequestCount() != 1 {
		t.Fatalf("expected 1 request, got %d", recorder.RequestCount())
	}

	lastReq := recorder.LastRequest()
	if lastReq == nil {
		t.Fatal("expected non-nil last request")
	}

	// Verify method and URL
	if lastReq.Method != "POST" {
		t.Errorf("expected method 'POST', got %q", lastReq.Method)
	}

	// Verify body was preserved
	if lastReq.Body == nil {
		t.Fatal("expected non-nil request body")
	}

	recordedBody, err := io.ReadAll(lastReq.Body)
	if err != nil {
		t.Fatalf("failed to read recorded body: %v", err)
	}

	if string(recordedBody) != requestBody {
		t.Errorf("expected body %q, got %q", requestBody, string(recordedBody))
	}
}

func TestRequestRecorder_Reset(t *testing.T) {
	t.Parallel()
	recorder := testutil.NewRequestRecorder(testutil.MockResponse(200, "ok"))
	client := recorder.Client()

	// Make some requests
	for i := 0; i < 3; i++ {
		resp, err := client.Get("http://example.com/api")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_ = resp.Body.Close()
	}

	if recorder.RequestCount() != 3 {
		t.Fatalf("expected 3 requests before reset, got %d", recorder.RequestCount())
	}

	// Reset recorder
	recorder.Reset()

	if recorder.RequestCount() != 0 {
		t.Errorf("expected 0 requests after reset, got %d", recorder.RequestCount())
	}

	if recorder.LastRequest() != nil {
		t.Error("expected nil last request after reset")
	}
}

func TestRequestRecorder_NoRequests(t *testing.T) {
	t.Parallel()
	recorder := testutil.NewRequestRecorder(testutil.MockResponse(200, "ok"))

	if recorder.RequestCount() != 0 {
		t.Errorf("expected 0 requests initially, got %d", recorder.RequestCount())
	}

	if recorder.LastRequest() != nil {
		t.Error("expected nil last request when no requests made")
	}
}

func TestRequestRecorder_WithError(t *testing.T) {
	t.Parallel()
	expectedErr := errors.New("mock error")
	recorder := testutil.NewRequestRecorder(nil)
	recorder.Error = expectedErr

	client := recorder.Client()

	resp, err := client.Get("http://example.com/api")
	if err == nil {
		t.Error("expected error, got nil")
	}

	if resp != nil {
		t.Error("expected nil response with error")
	}

	// Error should be wrapped but contain the original error
	if !strings.Contains(err.Error(), expectedErr.Error()) {
		t.Errorf("expected error to contain %q, got %v", expectedErr.Error(), err)
	}

	// Error should still record the request
	if recorder.RequestCount() != 1 {
		t.Errorf("expected 1 request even with error, got %d", recorder.RequestCount())
	}
}
