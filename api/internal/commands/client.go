package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// BaseHTTPClient provides a reusable HTTP client for making requests
// without automatic authentication (useful for public endpoints)
type BaseHTTPClient struct {
	client *http.Client
}

// NewBaseHTTPClient creates a new base HTTP client with optimized settings
func NewBaseHTTPClient() *BaseHTTPClient {
	return &BaseHTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
				DisableCompression:  false,
			},
		},
	}
}

// BuildURL constructs a URL from base server URL and path
// Automatically handles trailing slashes
func BuildURL(serverURL, path string) string {
	serverURL = strings.TrimSuffix(serverURL, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return serverURL + path
}

// BuildRequestBody marshals the request body to JSON
func BuildRequestBody(body interface{}) ([]byte, error) {
	if body == nil {
		return nil, nil
	}
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	return jsonData, nil
}

// CreateRequest creates a new HTTP request with the given method, URL, and body
func (c *BaseHTTPClient) CreateRequest(method, url string, body interface{}) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := BuildRequestBody(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set Content-Type for JSON requests if body is provided
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// Do executes an HTTP request
func (c *BaseHTTPClient) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, sanitizeHTTPError(err)
	}
	return resp, nil
}

// Post makes a POST request
func (c *BaseHTTPClient) Post(url string, body interface{}) (*http.Response, error) {
	req, err := c.CreateRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Get makes a GET request
func (c *BaseHTTPClient) Get(url string) (*http.Response, error) {
	req, err := c.CreateRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Put makes a PUT request
func (c *BaseHTTPClient) Put(url string, body interface{}) (*http.Response, error) {
	req, err := c.CreateRequest("PUT", url, body)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Delete makes a DELETE request
func (c *BaseHTTPClient) Delete(url string) (*http.Response, error) {
	req, err := c.CreateRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// ErrorResponse represents the standard error response structure from the API
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

// ReadResponseBody reads the response body and returns it as bytes
func ReadResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return bodyBytes, nil
}

// ParseErrorResponse attempts to parse an error response from the API
func ParseErrorResponse(bodyBytes []byte) *ErrorResponse {
	var errorResp ErrorResponse
	if err := json.Unmarshal(bodyBytes, &errorResp); err == nil {
		return &errorResp
	}
	return nil
}

// HandleErrorResponse checks the status code and returns an appropriate error
// If the response contains a parseable error structure, it extracts the message
func HandleErrorResponse(resp *http.Response, bodyBytes []byte, defaultMessage string) error {
	errorResp := ParseErrorResponse(bodyBytes)
	if errorResp != nil {
		if errorResp.Message != "" {
			return fmt.Errorf("%s: %s (status: %d)", defaultMessage, errorResp.Message, resp.StatusCode)
		}
		if errorResp.Error != "" {
			return fmt.Errorf("%s: %s (status: %d)", defaultMessage, errorResp.Error, resp.StatusCode)
		}
	}
	return fmt.Errorf("%s (status: %d)", defaultMessage, resp.StatusCode)
}

// ParseJSONResponse parses a JSON response into the provided target
func ParseJSONResponse(bodyBytes []byte, target interface{}) error {
	if err := json.Unmarshal(bodyBytes, target); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	return nil
}

// sanitizeHTTPError converts technical HTTP errors into user-friendly messages
func sanitizeHTTPError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Remove POST/GET/etc method prefixes and URLs from error messages
	if strings.Contains(errStr, "unsupported protocol scheme") {
		return fmt.Errorf("invalid server URL format. Please include http:// or https://")
	}

	if strings.Contains(errStr, "no such host") || strings.Contains(errStr, "unknown host") {
		return fmt.Errorf("server hostname not found")
	}

	if strings.Contains(errStr, "connection refused") {
		return fmt.Errorf("connection refused. Please check if the server is running")
	}

	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded") {
		return fmt.Errorf("connection timeout. Please check your network connection")
	}

	if strings.Contains(errStr, "network is unreachable") {
		return fmt.Errorf("network unreachable. Please check your network connection")
	}

	// For other errors, return a generic message without exposing technical details
	return fmt.Errorf("connection failed: %w", err)
}
