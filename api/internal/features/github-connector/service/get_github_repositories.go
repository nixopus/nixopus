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

func (c *GithubConnectorService) GetGithubRepositories(user_id string) ([]shared_types.GithubRepository, error) {
	connectors, err := c.storage.GetAllConnectors(user_id)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}

	if len(connectors) == 0 {
		c.logger.Log(logger.Error, "No connectors found for user", user_id)
		return nil, nil
	}

	installation_id := connectors[0].InstallationID

	jwt := GenerateJwt(&connectors[0])

	accessToken, err := c.getInstallationToken(jwt, installation_id)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("Failed to get installation token: %s", err.Error()), "")
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.github.com/installation/repositories?per_page=500", nil)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "nixopus")

	resp, err := client.Do(req)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		c.logger.Log(logger.Error, fmt.Sprintf("GitHub API error: %s - %s", resp.Status, string(bodyBytes)), "")
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var response struct {
		Repositories []shared_types.GithubRepository `json:"repositories"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}

	return response.Repositories, nil
}

func (c *GithubConnectorService) getInstallationToken(jwt string, installation_id string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/app/installations/%s/access_tokens", installation_id)

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
