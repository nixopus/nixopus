package secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// SecretManagerType represents the type of secret manager
type SecretManagerType string

const (
	SecretManagerNone      SecretManagerType = "none"
	SecretManagerInfisical SecretManagerType = "infisical"
)

// SecretManagerConfig holds configuration for secret managers
type SecretManagerConfig struct {
	Type           SecretManagerType
	Enabled        bool
	ProjectID      string
	Environment    string
	SecretPath     string // Path where secrets are stored in Infisical (e.g., "/" or "/api")
	ServiceName    string
	InfisicalURL   string
	InfisicalToken string
}

// SecretManager interface for fetching secrets
type SecretManager interface {
	GetSecret(ctx context.Context, key string) (string, error)
	GetSecrets(ctx context.Context, prefix string) (map[string]string, error)
}

// LoadSecretManagerConfig loads secret manager configuration from environment variables
func LoadSecretManagerConfig(serviceName string) *SecretManagerConfig {
	managerType := SecretManagerType(strings.ToLower(os.Getenv("SECRET_MANAGER_TYPE")))
	if managerType == "" {
		managerType = SecretManagerNone
	}

	enabled := os.Getenv("SECRET_MANAGER_ENABLED") == "true"
	if !enabled && managerType == SecretManagerNone {
		return &SecretManagerConfig{
			Type:    SecretManagerNone,
			Enabled: false,
		}
	}

	config := &SecretManagerConfig{
		Type:           managerType,
		Enabled:        enabled,
		ProjectID:      os.Getenv("SECRET_MANAGER_PROJECT_ID"),
		Environment:    getEnvOrDefault("SECRET_MANAGER_ENVIRONMENT", "prod"),
		SecretPath:     getEnvOrDefault("SECRET_MANAGER_SECRET_PATH", "/"),
		ServiceName:    serviceName,
		InfisicalURL:   getEnvOrDefault("INFISICAL_URL", "https://app.infisical.com"),
		InfisicalToken: os.Getenv("INFISICAL_TOKEN"),
	}

	return config
}

// NewSecretManager creates a new secret manager instance based on configuration
func NewSecretManager(config *SecretManagerConfig) (SecretManager, error) {
	if !config.Enabled || config.Type == SecretManagerNone {
		return &NoOpSecretManager{}, nil
	}

	switch config.Type {
	case SecretManagerInfisical:
		if config.InfisicalToken == "" {
			return nil, fmt.Errorf("INFISICAL_TOKEN is required when using Infisical")
		}
		return NewInfisicalManager(config), nil
	default:
		return &NoOpSecretManager{}, nil
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// normalizeEnvironmentName converts common environment names to Infisical slug format
func normalizeEnvironmentName(env string) string {
	env = strings.ToLower(strings.TrimSpace(env))
	// Map common variations to Infisical slug format
	// Note: Infisical uses "prod" not "production", "dev" not "development"
	switch env {
	case "dev", "development":
		return "dev"
	case "staging", "stage":
		return "staging"
	case "prod", "production":
		return "prod"
	default:
		return env
	}
}

// NoOpSecretManager is a no-op implementation that returns empty values
type NoOpSecretManager struct{}

func (n *NoOpSecretManager) GetSecret(ctx context.Context, key string) (string, error) {
	return "", fmt.Errorf("secret manager not configured")
}

func (n *NoOpSecretManager) GetSecrets(ctx context.Context, prefix string) (map[string]string, error) {
	return make(map[string]string), nil
}

// InfisicalManager implements SecretManager for Infisical
type InfisicalManager struct {
	config     *SecretManagerConfig
	httpClient *http.Client
	baseURL    string
}

func NewInfisicalManager(config *SecretManagerConfig) *InfisicalManager {
	return &InfisicalManager{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: config.InfisicalURL,
	}
}

type InfisicalSecret struct {
	SecretKey   string `json:"secretKey"`
	SecretValue string `json:"secretValue"`
}

type InfisicalSecretsResponse struct {
	Secrets []InfisicalSecret `json:"secrets"`
}

func (i *InfisicalManager) GetSecret(ctx context.Context, key string) (string, error) {
	secrets, err := i.GetSecrets(ctx, "")
	if err != nil {
		return "", err
	}

	value, exists := secrets[key]
	if !exists {
		return "", fmt.Errorf("secret %s not found", key)
	}
	return value, nil
}

func (i *InfisicalManager) GetSecrets(ctx context.Context, prefix string) (map[string]string, error) {
	url := fmt.Sprintf("%s/api/v3/secrets/raw", i.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", i.config.InfisicalToken))
	req.Header.Set("Content-Type", "application/json")

	// Add query parameters
	q := req.URL.Query()
	if i.config.ProjectID != "" {
		q.Add("workspaceId", i.config.ProjectID)
	}

	// Environment is required by Infisical API
	envSlug := strings.ToLower(i.config.Environment)
	if envSlug == "" {
		return nil, fmt.Errorf("environment is required but not set in SECRET_MANAGER_ENVIRONMENT")
	}

	// Normalize common environment names to Infisical slug format
	envSlug = normalizeEnvironmentName(envSlug)
	q.Add("environment", envSlug)

	// Add secret path (defaults to "/" for root)
	secretPath := i.config.SecretPath
	if secretPath == "" {
		secretPath = "/"
	}
	q.Add("secretPath", secretPath)

	// Set recursive to true to fetch secrets from subfolders if needed
	q.Add("recursive", "true")

	req.URL.RawQuery = q.Encode()

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch secrets: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		// Handle 404 gracefully - secrets might not exist yet
		if resp.StatusCode == http.StatusNotFound {
			return make(map[string]string), nil
		}
		return nil, fmt.Errorf("failed to fetch secrets: status %d, body: %s", resp.StatusCode, string(body))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try to parse as structured response first
	var secretsResponse InfisicalSecretsResponse
	if err := json.Unmarshal(bodyBytes, &secretsResponse); err == nil && secretsResponse.Secrets != nil {
		result := make(map[string]string)
		for _, secret := range secretsResponse.Secrets {
			if prefix == "" || strings.HasPrefix(secret.SecretKey, prefix) {
				result[secret.SecretKey] = secret.SecretValue
			}
		}
		return result, nil
	}

	// Fallback: try parsing as flat key-value object
	var flatSecrets map[string]string
	if err := json.Unmarshal(bodyBytes, &flatSecrets); err == nil {
		result := make(map[string]string)
		for key, value := range flatSecrets {
			if prefix == "" || strings.HasPrefix(key, prefix) {
				result[key] = value
			}
		}
		return result, nil
	}

	return nil, fmt.Errorf("failed to parse secrets response: invalid JSON format")
}

// LoadSecretsIntoEnv loads secrets from the secret manager into environment variables
// This is useful for services that expect environment variables
func LoadSecretsIntoEnv(manager SecretManager, prefix string) error {
	if manager == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	secrets, err := manager.GetSecrets(ctx, prefix)
	if err != nil {
		return fmt.Errorf("failed to load secrets: %w", err)
	}

	for key, value := range secrets {
		if err := os.Setenv(key, value); err != nil {
			log.Printf("Warning: Failed to set environment variable %s: %v", key, err)
		}
	}

	return nil
}
