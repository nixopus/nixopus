package gh

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/nixopus/nixopus/api/internal/config"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/sirupsen/logrus"
)

// InstallationToken exchanges a GitHub App JWT for an installation access token.
func InstallationToken(jwtStr string, installationID string) (string, error) {
	u := fmt.Sprintf("%s/app/installations/%s/access_tokens", APIBaseURL, installationID)

	client := &http.Client{}
	req, err := http.NewRequest("POST", u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtStr))
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
		if resp.StatusCode == http.StatusNotFound {
			return "", fmt.Errorf("installation not found: the GitHub installation ID '%s' is invalid or the app does not have access to it. Please reconnect your GitHub account", installationID)
		}
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return "", fmt.Errorf("authentication failed: the GitHub App credentials are invalid or expired. Please check your app configuration")
		}
		return "", fmt.Errorf("Failed to get installation token: %s - %s", resp.Status, bodyStr)
	}

	var tokenResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}
	return tokenResp.Token, nil
}

// GenerateJwt returns a GitHub App JWT from connector credentials or shared config.
func GenerateJwt(appCredentials *shared_types.GithubConnector) string {
	var pem string
	var appID string
	if appCredentials != nil && appCredentials.Pem != "" && appCredentials.AppID != "" {
		pem = appCredentials.Pem
		appID = appCredentials.AppID
	} else {
		githubConfig := config.AppConfig.GitHub
		if githubConfig.Pem == "" || githubConfig.AppID == "" {
			logrus.Error("github connector service: GenerateJwt GitHub App credentials not configured")
			return ""
		}
		pem = githubConfig.Pem
		appID = githubConfig.AppID
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(pem))
	if err != nil {
		logrus.Errorf("github connector service: GenerateJwt parse private key: %v", err)
		return ""
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(10 * time.Minute).Unix(),
		"iss": fmt.Sprintf("%v", appID),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		logrus.Errorf("github connector service: GenerateJwt sign token: %v", err)
		return ""
	}
	return tokenString
}
