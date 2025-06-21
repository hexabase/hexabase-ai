package proxmox

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockHTTPClient is a mock HTTP client for testing
type MockHTTPClient struct {
	t           *testing.T
	expectations []expectation
	callIndex    int
}

type expectation struct {
	method      string
	path        string
	requestBody interface{}
	headers     map[string]string
	response    *mockResponse
}

type mockResponse struct {
	statusCode int
	body       interface{}
	delay      time.Duration
}

// NewMockHTTPClient creates a new mock HTTP client
func NewMockHTTPClient(t *testing.T) *MockHTTPClient {
	return &MockHTTPClient{
		t:            t,
		expectations: make([]expectation, 0),
	}
}

// ExpectGet sets up an expectation for a GET request
func (m *MockHTTPClient) ExpectGet(path string) *expectationBuilder {
	return &expectationBuilder{
		mock: m,
		exp: expectation{
			method: "GET",
			path:   path,
		},
	}
}

// ExpectPost sets up an expectation for a POST request
func (m *MockHTTPClient) ExpectPost(path string) *expectationBuilder {
	return &expectationBuilder{
		mock: m,
		exp: expectation{
			method: "POST",
			path:   path,
		},
	}
}

// ExpectPut sets up an expectation for a PUT request
func (m *MockHTTPClient) ExpectPut(path string) *expectationBuilder {
	return &expectationBuilder{
		mock: m,
		exp: expectation{
			method: "PUT",
			path:   path,
		},
	}
}

// ExpectDelete sets up an expectation for a DELETE request
func (m *MockHTTPClient) ExpectDelete(path string) *expectationBuilder {
	return &expectationBuilder{
		mock: m,
		exp: expectation{
			method: "DELETE",
			path:   path,
		},
	}
}

// Do implements the http.Client interface
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.callIndex >= len(m.expectations) {
		m.t.Fatalf("unexpected request: %s %s", req.Method, req.URL.Path)
	}

	exp := m.expectations[m.callIndex]
	m.callIndex++

	// Verify method and path
	assert.Equal(m.t, exp.method, req.Method)
	assert.Equal(m.t, exp.path, req.URL.Path)

	// Verify headers if specified
	if exp.headers != nil {
		for key, value := range exp.headers {
			assert.Equal(m.t, value, req.Header.Get(key))
		}
	}

	// Verify request body if specified
	if exp.requestBody != nil && req.Body != nil {
		var actualBody interface{}
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&actualBody)
		assert.NoError(m.t, err)
		assert.Equal(m.t, exp.requestBody, actualBody)
	}

	// Simulate delay if specified
	if exp.response.delay > 0 {
		time.Sleep(exp.response.delay)
	}

	// Create response
	var bodyReader io.ReadCloser
	if exp.response.body != nil {
		bodyBytes, err := json.Marshal(exp.response.body)
		assert.NoError(m.t, err)
		bodyReader = io.NopCloser(bytes.NewReader(bodyBytes))
	} else {
		bodyReader = io.NopCloser(strings.NewReader(""))
	}

	return &http.Response{
		StatusCode: exp.response.statusCode,
		Body:       bodyReader,
		Header:     make(http.Header),
	}, nil
}

// AssertExpectations verifies all expectations were met
func (m *MockHTTPClient) AssertExpectations(t *testing.T) {
	assert.Equal(t, len(m.expectations), m.callIndex, "not all expectations were met")
}

// expectationBuilder helps build expectations fluently
type expectationBuilder struct {
	mock *MockHTTPClient
	exp  expectation
}

// WithJSON sets the expected request body
func (e *expectationBuilder) WithJSON(body interface{}) *expectationBuilder {
	e.exp.requestBody = body
	return e
}

// WithHeaders sets expected headers
func (e *expectationBuilder) WithHeaders(headers map[string]string) *expectationBuilder {
	e.exp.headers = headers
	return e
}

// RespondJSON sets the response
func (e *expectationBuilder) RespondJSON(statusCode int, body interface{}) {
	e.exp.response = &mockResponse{
		statusCode: statusCode,
		body:       body,
	}
	e.mock.expectations = append(e.mock.expectations, e.exp)
}

// RespondAfter sets a delayed response
func (e *expectationBuilder) RespondAfter(delay time.Duration, statusCode int, body interface{}) {
	e.exp.response = &mockResponse{
		statusCode: statusCode,
		body:       body,
		delay:      delay,
	}
	e.mock.expectations = append(e.mock.expectations, e.exp)
}