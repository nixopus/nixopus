package initcmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ValidateAPIKeyRequest represents the request body for API key validation
type ValidateAPIKeyRequest struct {
	APIKey string `json:"api_key"`
}

// ValidateAPIKeyResponse represents the response from API key validation endpoint
type ValidateAPIKeyResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Valid   bool   `json:"valid"`
}

// CreateProjectRequest represents the request body for creating a project
type CreateProjectRequest struct {
	APIKey               string            `json:"api_key"`
	Name                 string            `json:"name"`
	Repository           string            `json:"repository"`
	Branch               string            `json:"branch,omitempty"`
	EnvironmentVariables map[string]string `json:"environment_variables,omitempty"`
}

// CreateProjectResponse represents the response from project creation endpoint
type CreateProjectResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	ProjectID string `json:"project_id"`
	FamilyID  string `json:"family_id"`
}

// baseHTTPClient provides a reusable HTTP client for making requests
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

// post makes a POST request
func (c *baseHTTPClient) post(url string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest("POST", url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

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

// getGitBranch gets the current git branch
func getGitBranch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// getGitInfo gets the git repository name and remote URL
func getGitInfo() (string, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	repoName := filepath.Base(cwd)

	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	output, err := cmd.Output()
	if err != nil {
		return repoName, repoName, nil
	}

	repoURL := strings.TrimSpace(string(output))
	if repoURL == "" {
		return repoName, repoName, nil
	}

	parts := strings.Split(repoURL, "/")
	if len(parts) > 0 {
		lastPart := strings.TrimSuffix(parts[len(parts)-1], ".git")
		if lastPart != "" {
			repoName = lastPart
		}
	}

	return repoName, repoURL, nil
}

// ValidateAPIKey validates an API key by calling the validation endpoint
func ValidateAPIKey(serverURL, apiKey string) error {
	client := newBaseHTTPClient()
	url := buildURL(serverURL, "/api/v1/auth/validate-api-key")

	reqBody := ValidateAPIKeyRequest{
		APIKey: apiKey,
	}

	resp, err := client.post(url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	bodyBytes, err := readResponseBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return handleErrorResponse(resp, bodyBytes, "API key validation failed")
	}

	var validationResp ValidateAPIKeyResponse
	if err := parseJSONResponse(bodyBytes, &validationResp); err != nil {
		return err
	}

	if !validationResp.Valid {
		if validationResp.Message != "" {
			return fmt.Errorf("API key validation failed: %s", validationResp.Message)
		}
		return fmt.Errorf("API key validation failed")
	}

	return nil
}

// CreateProject creates a draft project on the server using the CLI init endpoint
// Returns: projectID, familyID, error
func CreateProject(serverURL, apiKey string, envVars map[string]string) (string, string, error) {
	repoName, repoURL, err := getGitInfo()
	if err != nil {
		return "", "", fmt.Errorf("failed to get git info: %w", err)
	}

	branch := getGitBranch()

	client := newBaseHTTPClient()
	url := buildURL(serverURL, "/api/v1/auth/cli-init")

	reqBody := CreateProjectRequest{
		APIKey:               apiKey,
		Name:                 repoName,
		Repository:           repoURL,
		Branch:               branch,
		EnvironmentVariables: envVars,
	}

	resp, err := client.post(url, reqBody)
	if err != nil {
		return "", "", fmt.Errorf("failed to connect to server: %w", err)
	}

	bodyBytes, err := readResponseBody(resp)
	if err != nil {
		return "", "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", "", handleErrorResponse(resp, bodyBytes, "failed to create project")
	}

	var projectResp CreateProjectResponse
	if err := parseJSONResponse(bodyBytes, &projectResp); err != nil {
		return "", "", err
	}

	if projectResp.ProjectID == "" {
		return "", "", fmt.Errorf("project ID not found in response")
	}

	if projectResp.FamilyID == "" {
		return "", "", fmt.Errorf("family ID not found in response")
	}

	return projectResp.ProjectID, projectResp.FamilyID, nil
}
