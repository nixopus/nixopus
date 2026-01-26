package removecmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DeleteApplicationRequest represents the request body for deleting an application
type DeleteApplicationRequest struct {
	ID uuid.UUID `json:"id"`
}

// DeleteApplicationResponse represents the response
type DeleteApplicationResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// baseHTTPClient provides a reusable HTTP client
type baseHTTPClient struct {
	client *http.Client
}

// newBaseHTTPClient creates a new base HTTP client
func newBaseHTTPClient() *baseHTTPClient {
	return &baseHTTPClient{
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
			},
		},
	}
}

// buildURL constructs a URL from base server URL and path
func buildURL(serverURL, path string) string {
	serverURL = strings.TrimSuffix(serverURL, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return serverURL + path
}

// delete makes a DELETE request with API key authentication
func (c *baseHTTPClient) delete(url string, body interface{}, apiKey string) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest("DELETE", url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, sanitizeHTTPError(err)
	}
	return resp, nil
}

// readResponseBody reads the response body and returns it as bytes
func readResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return bodyBytes, nil
}

// parseJSONResponse parses a JSON response into the provided target
func parseJSONResponse(bodyBytes []byte, target interface{}) error {
	if err := json.Unmarshal(bodyBytes, target); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	return nil
}

// errorResponse represents the standard error response structure from the API
type errorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

// handleErrorResponse checks the status code and returns an appropriate error
func handleErrorResponse(resp *http.Response, bodyBytes []byte, defaultMessage string) error {
	var errorResp errorResponse
	if err := json.Unmarshal(bodyBytes, &errorResp); err == nil {
		if errorResp.Message != "" {
			return fmt.Errorf("%s: %s (status: %d)", defaultMessage, errorResp.Message, resp.StatusCode)
		}
		if errorResp.Error != "" {
			return fmt.Errorf("%s: %s (status: %d)", defaultMessage, errorResp.Error, resp.StatusCode)
		}
	}
	return fmt.Errorf("%s (status: %d)", defaultMessage, resp.StatusCode)
}

// sanitizeHTTPError converts technical HTTP errors into user-friendly messages
func sanitizeHTTPError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()
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
	return fmt.Errorf("connection failed: %w", err)
}

// deleteApplication calls the API to delete an application
func deleteApplication(serverURL, apiKey, applicationID string) error {
	// Parse application ID UUID
	appUUID, err := uuid.Parse(applicationID)
	if err != nil {
		return fmt.Errorf("invalid application_id: %w", err)
	}

	client := newBaseHTTPClient()
	url := buildURL(serverURL, "/api/v1/deploy/application")

	reqBody := DeleteApplicationRequest{
		ID: appUUID,
	}

	resp, err := client.delete(url, reqBody, apiKey)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	bodyBytes, err := readResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return handleErrorResponse(resp, bodyBytes, "failed to delete application")
	}

	var deleteResp DeleteApplicationResponse
	if err := parseJSONResponse(bodyBytes, &deleteResp); err != nil {
		// If response is empty (204 No Content), that's also fine
		if resp.StatusCode == http.StatusNoContent {
			return nil
		}
		return err
	}

	return nil
}
