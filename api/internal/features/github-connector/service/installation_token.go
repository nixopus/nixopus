package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/raghavyuva/nixopus-api/internal/config"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *GithubConnectorService) getInstallationToken(jwt string, installation_id string) (string, error) {
	url := fmt.Sprintf("%s/app/installations/%s/access_tokens", githubAPIBaseURL, installation_id)

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "nixopus")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)
		
		// Handle specific error cases
		if resp.StatusCode == http.StatusNotFound {
			return "", fmt.Errorf("installation not found: the GitHub installation ID '%s' is invalid or the app does not have access to it. Please reconnect your GitHub account", installation_id)
		}
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return "", fmt.Errorf("authentication failed: the GitHub App credentials are invalid or expired. Please check your app configuration")
		}
		
		errMsg := fmt.Sprintf("Failed to get installation token: %s - %s", resp.Status, bodyStr)
		return "", errors.New(errMsg)
	}

	var tokenResp struct {
		Token string `json:"token"`
	}

	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return "", err
	}

	return tokenResp.Token, nil
}

func GenerateJwt(app_credentials *shared_types.GithubConnector) string {
	var pem string
	var appID string

	// Use connector credentials if available, otherwise use shared config
	if app_credentials != nil && app_credentials.Pem != "" && app_credentials.AppID != "" {
		pem = app_credentials.Pem
		appID = app_credentials.AppID
	} else {
		// Use shared GitHub App credentials from config
		githubConfig := config.AppConfig.GitHub
		if githubConfig.Pem == "" || githubConfig.AppID == "" {
			fmt.Println("Error: GitHub App credentials not configured")
			return ""
		}
		pem = githubConfig.Pem
		appID = githubConfig.AppID
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(pem))
	if err != nil {
		fmt.Println("Error parsing private key:", err)
		return ""
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(time.Minute * 10).Unix(),
		"iss": fmt.Sprintf("%v", appID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		fmt.Println("Error signing token:", err)
		return ""
	}

	return tokenString
}
