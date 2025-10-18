// Package testutil provides shared testing utilities for EVE SDE Database Builder tests.
package testutil

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

// RoundTripFunc is a function type that implements http.RoundTripper interface.
// This allows using functions directly as HTTP transport mocks.
type RoundTripFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements the http.RoundTripper interface for RoundTripFunc.
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// MockHTTPClient creates a mock HTTP client with a custom RoundTripper function.
// This is useful for testing code that makes HTTP requests without actually hitting the network.
//
// Example:
//
//	client := testutil.MockHTTPClient(func(req *http.Request) (*http.Response, error) {
//	    return &http.Response{
//	        StatusCode: 200,
//	        Body:       io.NopCloser(strings.NewReader(`{"result":"ok"}`)),
//	        Header:     make(http.Header),
//	    }, nil
//	})
func MockHTTPClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

// MockResponse creates a simple HTTP response with the given status code and body.
// The response includes standard headers and a properly closed body.
//
// Example:
//
//	resp := testutil.MockResponse(200, `{"id": 1, "name": "test"}`)
func MockResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

// MockJSONResponse creates an HTTP response with JSON content type and the given body.
// This is a convenience wrapper around MockResponse for JSON responses.
//
// Example:
//
//	resp := testutil.MockJSONResponse(200, `{"id": 1}`)
func MockJSONResponse(statusCode int, jsonBody string) *http.Response {
	resp := MockResponse(statusCode, jsonBody)
	resp.Header.Set("Content-Type", "application/json")
	return resp
}

// MockErrorResponse creates a failing HTTP response with an error.
// This is useful for testing error handling in HTTP client code.
//
// Example:
//
//	client := testutil.MockHTTPClient(func(req *http.Request) (*http.Response, error) {
//	    return testutil.MockErrorResponse(errors.New("network timeout"))
//	})
func MockErrorResponse(err error) (*http.Response, error) {
	return nil, err
}

// StaticMockClient creates a mock HTTP client that always returns the same response.
// This is useful for simple test cases that don't need dynamic behavior.
//
// Example:
//
//	client := testutil.StaticMockClient(200, `{"result":"ok"}`)
func StaticMockClient(statusCode int, body string) *http.Client {
	return MockHTTPClient(func(req *http.Request) (*http.Response, error) {
		return MockResponse(statusCode, body), nil
	})
}

// RequestRecorder records HTTP requests for later inspection in tests.
// This is useful for verifying that the correct requests were made.
type RequestRecorder struct {
	Requests []*http.Request
	Response *http.Response
	Error    error
}

// NewRequestRecorder creates a new RequestRecorder with the given response.
//
// Example:
//
//	recorder := testutil.NewRequestRecorder(testutil.MockResponse(200, "ok"))
//	client := recorder.Client()
//	// Make requests with client...
//	if len(recorder.Requests) != 1 {
//	    t.Error("expected 1 request")
//	}
func NewRequestRecorder(response *http.Response) *RequestRecorder {
	return &RequestRecorder{
		Requests: make([]*http.Request, 0),
		Response: response,
		Error:    nil,
	}
}

// Client returns an HTTP client that uses this recorder as transport.
func (r *RequestRecorder) Client() *http.Client {
	return MockHTTPClient(func(req *http.Request) (*http.Response, error) {
		// Clone the request to avoid modifications affecting the recorded version
		clonedReq := req.Clone(req.Context())

		// If request has a body, we need to read and restore it
		if req.Body != nil {
			bodyBytes, _ := io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			clonedReq.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		r.Requests = append(r.Requests, clonedReq)
		return r.Response, r.Error
	})
}

// Reset clears all recorded requests.
func (r *RequestRecorder) Reset() {
	r.Requests = make([]*http.Request, 0)
}

// RequestCount returns the number of recorded requests.
func (r *RequestRecorder) RequestCount() int {
	return len(r.Requests)
}

// LastRequest returns the most recently recorded request, or nil if no requests were made.
func (r *RequestRecorder) LastRequest() *http.Request {
	if len(r.Requests) == 0 {
		return nil
	}
	return r.Requests[len(r.Requests)-1]
}
