package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	productionServerURL = "http://localhost:8080"
	configFileName      = ".nixopus"
	authFileName        = "auth.json"
)

// AuthConfig represents global authentication configuration stored in user's home directory
type AuthConfig struct {
	AccessToken    string `json:"access_token,omitempty"`    // Bearer token for API authentication
	RefreshToken   string `json:"refresh_token,omitempty"`   // Optional refresh token
	OrganizationID string `json:"organization_id,omitempty"` // Active organization ID from Better Auth
}

// GetServerURL returns the server URL from environment variable or default
func GetServerURL() string {
	return productionServerURL
}

// Config represents the nixopus configuration stored in .nixopus file
type Config struct {
	Server       string            `json:"server,omitempty"`
	ProjectID    string            `json:"project_id,omitempty"`
	FamilyID     string            `json:"family_id,omitempty"`    // Family ID (group of apps)
	Applications map[string]string `json:"applications,omitempty"` // Map of app name -> application_id
	Sync         SyncConfig        `json:"sync"`
	EnvPath      string            `json:"env_path,omitempty"`
}

// SyncConfig represents sync-related configuration
type SyncConfig struct {
	DebounceMs int      `json:"debounce_ms"`
	Exclude    []string `json:"exclude"`
}

// getConfigPath returns the path to the .nixopus file in the project root
func getConfigPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Look for .git directory to find project root
	dir := cwd
	for {
		gitPath := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			// Found .git directory, use this as project root
			return filepath.Join(dir, configFileName), nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root, use current directory
			break
		}
		dir = parent
	}

	// If no .git found, use current directory
	return filepath.Join(cwd, configFileName), nil
}

// Load reads the configuration from .nixopus file
func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config not found. Run 'nixopus live' to initialize and start deployment")
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{
		Server: GetServerURL(),
		Sync: SyncConfig{
			DebounceMs: 300,
			Exclude: []string{
				"*.log",
				".git",
				"node_modules",
				"__pycache__",
				".env",
			},
		},
	}

	// Parse JSON config
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Ensure at least one application is present (either via ProjectID or Applications)
	if cfg.ProjectID == "" && len(cfg.Applications) == 0 {
		return nil, fmt.Errorf("no applications found in config. Run 'nixopus live' to initialize and start deployment")
	}

	// Ensure Server is set (from env or default)
	cfg.Server = GetServerURL()

	// Set defaults for sync if not present
	if cfg.Sync.DebounceMs == 0 {
		cfg.Sync.DebounceMs = 300
	}
	if len(cfg.Sync.Exclude) == 0 {
		cfg.Sync.Exclude = []string{
			"*.log",
			".git",
			"node_modules",
			"__pycache__",
			".env",
		}
	}

	// Ensure env path is not in exclude list if it's set
	if cfg.EnvPath != "" {
		cfg.Sync.Exclude = removeFromExcludes(cfg.Sync.Exclude, cfg.EnvPath)
	}

	return cfg, nil
}

// removeFromExcludes removes an item from the exclude list
func removeFromExcludes(excludes []string, item string) []string {
	result := make([]string, 0, len(excludes))
	cleanItem := filepath.Clean(item)
	baseName := filepath.Base(cleanItem)

	for _, exclude := range excludes {
		// Remove exact match
		if exclude == item || exclude == cleanItem {
			continue
		}
		// Remove base name match (e.g., if exclude is ".env" and item is ".env.production")
		if exclude == baseName {
			continue
		}
		result = append(result, exclude)
	}
	return result
}

// Save writes the configuration to .nixopus file
// Note: Server is not saved - it's always determined from defaultServerURL or API_URL env var
func (c *Config) Save() error {
	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Create a copy of config for saving (without Server and auth tokens)
	// Auth tokens are stored globally, not in project config
	saveConfig := &Config{
		FamilyID:     c.FamilyID,
		Applications: c.Applications,
		Sync:         c.Sync,
		EnvPath:      c.EnvPath,
		ProjectID:    c.ProjectID,
		// Auth tokens are stored globally, not in project config
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(saveConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetApplicationID returns the application ID for the given app name.
// If name is empty, returns the default application ID.
// Returns error if application not found.
func (c *Config) GetApplicationID(name string) (string, error) {
	// If name is empty, try to get default
	if name == "" {
		if appID, ok := c.Applications["default"]; ok && appID != "" {
			return appID, nil
		}
		return "", fmt.Errorf("no default application found. Use 'nixopus add' to add an application")
	}

	// Look up by name
	if appID, ok := c.Applications[name]; ok && appID != "" {
		return appID, nil
	}

	return "", fmt.Errorf("application '%s' not found. Use 'nixopus list' to see available applications", name)
}

// ValidateEnvPath validates that an env file path is safe and within project root
func ValidateEnvPath(envPath string) error {
	if envPath == "" {
		return nil // Empty is valid (optional)
	}

	// Prevent absolute paths outside project
	if filepath.IsAbs(envPath) {
		return fmt.Errorf("env path must be relative to project root, got absolute path: %s", envPath)
	}

	// Prevent path traversal attacks
	cleanPath := filepath.Clean(envPath)
	if strings.HasPrefix(cleanPath, "..") || strings.Contains(cleanPath, "..") {
		return fmt.Errorf("env path contains invalid path traversal: %s", envPath)
	}

	// Check if file exists
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	fullPath := filepath.Join(cwd, cleanPath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("env file not found: %s", envPath)
	}

	return nil
}

// getAuthPath returns the path to the global auth file in ~/.config/nixopus/auth.json
func getAuthPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "nixopus")
	authPath := filepath.Join(configDir, authFileName)

	return authPath, nil
}

// ensureAuthDir ensures the auth directory exists
func ensureAuthDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "nixopus")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create auth directory: %w", err)
	}

	return nil
}

// LoadAuth loads authentication tokens from global storage
func LoadAuth() (*AuthConfig, error) {
	authPath, err := getAuthPath()
	if err != nil {
		return nil, err
	}

	// Check if auth file exists
	if _, err := os.Stat(authPath); os.IsNotExist(err) {
		// Return empty auth config if file doesn't exist
		return &AuthConfig{}, nil
	}

	// Read auth file
	data, err := os.ReadFile(authPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read auth file: %w", err)
	}

	auth := &AuthConfig{}
	if err := json.Unmarshal(data, auth); err != nil {
		return nil, fmt.Errorf("failed to parse auth file: %w", err)
	}

	return auth, nil
}

// SaveAuth saves authentication tokens and organization ID to global storage
func SaveAuth(accessToken, refreshToken string) error {
	if err := ensureAuthDir(); err != nil {
		return err
	}

	authPath, err := getAuthPath()
	if err != nil {
		return err
	}

	// Load existing auth to preserve organization_id if not being changed
	existing, _ := LoadAuth()
	auth := &AuthConfig{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	if existing != nil && existing.OrganizationID != "" {
		auth.OrganizationID = existing.OrganizationID
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(auth, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal auth: %w", err)
	}

	// Write to file with restricted permissions (0600 = rw-------)
	if err := os.WriteFile(authPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write auth file: %w", err)
	}

	return nil
}

// SaveOrganizationID saves the organization ID to global auth storage
func SaveOrganizationID(organizationID string) error {
	auth, err := LoadAuth()
	if err != nil {
		auth = &AuthConfig{}
	}

	auth.OrganizationID = organizationID

	if err := ensureAuthDir(); err != nil {
		return err
	}

	authPath, err := getAuthPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(auth, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal auth: %w", err)
	}

	if err := os.WriteFile(authPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write auth file: %w", err)
	}

	return nil
}

// GetAccessToken returns the access token from global storage
func GetAccessToken() (string, error) {
	auth, err := LoadAuth()
	if err != nil {
		return "", err
	}

	if auth.AccessToken == "" {
		return "", fmt.Errorf("not authenticated. Please run 'nixopus login' first")
	}

	return auth.AccessToken, nil
}

// GetOrganizationID returns the organization ID from global storage
func GetOrganizationID() (string, error) {
	auth, err := LoadAuth()
	if err != nil {
		return "", err
	}

	if auth.OrganizationID == "" {
		return "", fmt.Errorf("no organization ID found. Please run 'nixopus login' again")
	}

	return auth.OrganizationID, nil
}

// ClearAuth removes authentication tokens from global storage
func ClearAuth() error {
	authPath, err := getAuthPath()
	if err != nil {
		return err
	}

	// Remove auth file if it exists
	if _, err := os.Stat(authPath); err == nil {
		if err := os.Remove(authPath); err != nil {
			return fmt.Errorf("failed to remove auth file: %w", err)
		}
	}

	return nil
}
