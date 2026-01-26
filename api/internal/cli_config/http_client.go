package cli_config

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

// HTTPClient is a helper client that automatically adds API key authentication
// to all requests. It loads the config once and reuses it for all requests.
type HTTPClient struct {
	config *Config
	client *http.Client
}

// NewHTTPClient creates a new HTTP client with automatic API key authentication
func NewHTTPClient() (*HTTPClient, error) {
	config, err := Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &HTTPClient{
		config: config,
		client: newOptimizedHTTPClient(),
	}, nil
}

// NewHTTPClientWithConfig creates a new HTTP client with the provided config
func NewHTTPClientWithConfig(config *Config) *HTTPClient {
	return &HTTPClient{
		config: config,
		client: newOptimizedHTTPClient(),
	}
}

// newOptimizedHTTPClient creates an HTTP client with connection pooling and timeouts
// for better performance and resource management
func newOptimizedHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second, // Request timeout
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        100,              // Maximum idle connections
			MaxIdleConnsPerHost: 10,               // Maximum idle connections per host
			IdleConnTimeout:     90 * time.Second, // Idle connection timeout
			TLSHandshakeTimeout: 10 * time.Second,
			DisableCompression:  false, // Enable compression for better performance
		},
	}
}

// GetConfig returns the loaded configuration
func (c *HTTPClient) GetConfig() *Config {
	return c.config
}

// GetAPIKey returns the API key from the config
func (c *HTTPClient) GetAPIKey() string {
	return c.config.APIKey
}

// GetServerURL returns the server URL from the config
func (c *HTTPClient) GetServerURL() string {
	return c.config.Server
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
	return fmt.Errorf("connection failed")
}

// NewRequest creates a new HTTP request with automatic API key authentication
func (c *HTTPClient) NewRequest(method, path string, body interface{}) (*http.Request, error) {
	url := c.config.Server + path

	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, sanitizeHTTPError(err)
	}

	// Automatically add API key to Authorization header
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	// Set Content-Type for JSON requests if body is provided
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// Do executes an HTTP request with automatic API key authentication
func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, sanitizeHTTPError(err)
	}
	return resp, nil
}

// Post is a convenience method for POST requests
func (c *HTTPClient) Post(path string, body interface{}) (*http.Response, error) {
	req, err := c.NewRequest("POST", path, body)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Get is a convenience method for GET requests
func (c *HTTPClient) Get(path string) (*http.Response, error) {
	req, err := c.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Delete is a convenience method for DELETE requests
func (c *HTTPClient) Delete(path string) (*http.Response, error) {
	req, err := c.NewRequest("DELETE", path, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Put is a convenience method for PUT requests
func (c *HTTPClient) Put(path string, body interface{}) (*http.Response, error) {
	req, err := c.NewRequest("PUT", path, body)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Patch is a convenience method for PATCH requests
func (c *HTTPClient) Patch(path string, body interface{}) (*http.Response, error) {
	req, err := c.NewRequest("PATCH", path, body)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}
