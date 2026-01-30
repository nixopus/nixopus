package logincmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DeviceCodeRequest represents the request for device code
type DeviceCodeRequest struct {
	ClientID string `json:"client_id"`
	Scope    string `json:"scope,omitempty"`
}

// DeviceCodeResponse represents the response from device code endpoint
type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"` // seconds
	Interval                int    `json:"interval"`   // seconds
}

// TokenRequest represents the request for access token
type TokenRequest struct {
	GrantType  string `json:"grant_type"`
	DeviceCode string `json:"device_code"`
	ClientID   string `json:"client_id"`
}

// TokenResponse represents the response from token endpoint
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// TokenErrorResponse represents error response from token endpoint
type TokenErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// RequestDeviceCode requests a device code from Better Auth
func RequestDeviceCode(betterAuthURL, clientID, scope string) (*DeviceCodeResponse, error) {
	// Better Auth basePath is /api/auth, so endpoint is /api/auth/device/code
	// Ensure URL doesn't have trailing slash
	betterAuthURL = strings.TrimSuffix(betterAuthURL, "/")
	requestURL := fmt.Sprintf("%s/api/auth/device/code", betterAuthURL)

	reqBody := DeviceCodeRequest{
		ClientID: clientID,
		Scope:    scope,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Post(requestURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth server at %s: %w", requestURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Provide more helpful error message
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("device authorization endpoint not found (404). Make sure:\n  1. Better Auth service is running\n  2. Device Authorization plugin is enabled\n  3. Database migration has been run (bunx @better-auth/cli migrate)\n  4. Better Auth URL is correct: %s", betterAuthURL)
		}
		return nil, fmt.Errorf("device code request failed (status %d): %s", resp.StatusCode, string(body))
	}

	var deviceCodeResp DeviceCodeResponse
	if err := json.Unmarshal(body, &deviceCodeResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &deviceCodeResp, nil
}

// PollForToken polls for access token using device code
func PollForToken(betterAuthURL, deviceCode, clientID string, interval, expiresIn int) (string, string, error) {
	tokenURL := fmt.Sprintf("%s/api/auth/device/token", betterAuthURL)
	pollingInterval := time.Duration(interval) * time.Second
	deadline := time.Now().Add(time.Duration(expiresIn) * time.Second)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	for time.Now().Before(deadline) {
		// Use JSON format instead of form data (Better Auth expects JSON)
		tokenReq := TokenRequest{
			GrantType:  "urn:ietf:params:oauth:grant-type:device_code",
			DeviceCode: deviceCode,
			ClientID:   clientID,
		}

		jsonData, err := json.Marshal(tokenReq)
		if err != nil {
			return "", "", fmt.Errorf("failed to marshal request: %w", err)
		}

		req, err := http.NewRequest("POST", tokenURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return "", "", fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return "", "", fmt.Errorf("network error: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", "", fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode == http.StatusOK {
			var tokenResp TokenResponse
			if err := json.Unmarshal(body, &tokenResp); err != nil {
				return "", "", fmt.Errorf("failed to parse token response: %w", err)
			}
			return tokenResp.AccessToken, tokenResp.RefreshToken, nil
		}

		var errorResp TokenErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return "", "", fmt.Errorf("failed to parse error response (status %d): %w, body: %s", resp.StatusCode, err, string(body))
		}

		switch errorResp.Error {
		case "authorization_pending":
			// Continue polling
			time.Sleep(pollingInterval)
			continue
		case "slow_down":
			// Increase polling interval
			pollingInterval += 5 * time.Second
			time.Sleep(pollingInterval)
			continue
		case "expired_token":
			return "", "", fmt.Errorf("device code expired. Please run 'nixopus login' again")
		case "access_denied":
			return "", "", fmt.Errorf("access denied by user")
		default:
			errorMsg := errorResp.ErrorDescription
			if errorMsg == "" {
				errorMsg = errorResp.Error
			}
			if errorMsg == "" {
				errorMsg = fmt.Sprintf("unknown error (status %d)", resp.StatusCode)
			}
			return "", "", fmt.Errorf("token request failed: %s", errorMsg)
		}
	}

	return "", "", fmt.Errorf("polling timeout - device code expired")
}

// FormatUserCode formats user code with dash (e.g., ABCD-1234)
func FormatUserCode(userCode string) string {
	if len(userCode) >= 8 {
		return userCode[:4] + "-" + userCode[4:]
	}
	return userCode
}
