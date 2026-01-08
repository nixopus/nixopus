package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
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
		errMsg := fmt.Sprintf("Failed to get installation token: %s - %s", resp.Status, string(bodyBytes))
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
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(app_credentials.Pem))
	if err != nil {
		fmt.Println("Error parsing private key:", err)
		return ""
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(time.Minute * 10).Unix(),
		"iss": fmt.Sprintf("%v", app_credentials.AppID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		fmt.Println("Error signing token:", err)
		return ""
	}

	return tokenString
}

// getAccessTokenForUser retrieves an installation access token for the user.
// It uses the connectorID if provided, otherwise finds a connector with valid installation_id.
// This is a shared helper function used by multiple service methods.
func (c *GithubConnectorService) getAccessTokenForUser(userID string, connectorID string) (string, error) {
	connectors, err := c.storage.GetAllConnectors(userID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return "", err
	}

	if len(connectors) == 0 {
		c.logger.Log(logger.Error, "No connectors found for user", userID)
		return "", fmt.Errorf("no connectors found")
	}

	var connectorToUse *shared_types.GithubConnector

	// If connectorID is provided, find that specific connector
	if connectorID != "" {
		for i := range connectors {
			if connectors[i].ID.String() == connectorID {
				connectorToUse = &connectors[i]
				break
			}
		}
		if connectorToUse == nil {
			c.logger.Log(logger.Error, fmt.Sprintf("Connector with id %s not found for user", connectorID), userID)
			return "", fmt.Errorf("connector not found")
		}
	} else {
		// Find connector with valid installation_id (not empty)
		for i := range connectors {
			if connectors[i].InstallationID != "" && connectors[i].InstallationID != " " {
				connectorToUse = &connectors[i]
				break
			}
		}
		if connectorToUse == nil {
			c.logger.Log(logger.Error, "No connector with valid installation_id found for user", userID)
			return "", fmt.Errorf("no connector with valid installation found")
		}
	}

	// Validate installation_id is not empty
	if connectorToUse.InstallationID == "" || connectorToUse.InstallationID == " " {
		c.logger.Log(logger.Error, fmt.Sprintf("Connector %s has empty installation_id", connectorToUse.ID.String()), userID)
		return "", fmt.Errorf("connector has no installation_id")
	}

	installationID := connectorToUse.InstallationID
	jwt := GenerateJwt(connectorToUse)
	if jwt == "" {
		c.logger.Log(logger.Error, "Failed to generate app JWT", "")
		return "", fmt.Errorf("failed to generate app JWT")
	}

	accessToken, err := c.getInstallationToken(jwt, installationID)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("Failed to get installation token: %s", err.Error()), "")
		return "", err
	}

	return accessToken, nil
}
