package tasks

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/raghavyuva/caddygo"
	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"gopkg.in/yaml.v3"
)

// ComposeSourceType defines how the compose file is provided
type ComposeSourceType string

const (
	ComposeSourceRepository ComposeSourceType = "repository"
	ComposeSourceURL        ComposeSourceType = "url"
	ComposeSourcePath       ComposeSourceType = "path"
	ComposeSourceRaw        ComposeSourceType = "raw"
)

// ComposeConfig holds configuration for docker compose deployment
type ComposeConfig struct {
	shared_types.TaskPayload
	TaskContext     *TaskContext
	SourceType      ComposeSourceType
	ComposeFilePath string
	ComposeURL      string
	ComposeRaw      string
	EnvVars         map[string]string
}

// ComposeService represents a service defined in docker-compose.yml
type ComposeService struct {
	Image       string            `yaml:"image,omitempty"`
	Build       interface{}       `yaml:"build,omitempty"`
	Ports       []string          `yaml:"ports,omitempty"`
	Environment interface{}       `yaml:"environment,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	DependsOn   interface{}       `yaml:"depends_on,omitempty"`
	Volumes     []string          `yaml:"volumes,omitempty"`
	Networks    interface{}       `yaml:"networks,omitempty"`
}

// ComposeFile represents the structure of a docker-compose.yml file
type ComposeFile struct {
	Version  string                    `yaml:"version,omitempty"`
	Services map[string]ComposeService `yaml:"services"`
	Networks map[string]interface{}    `yaml:"networks,omitempty"`
	Volumes  map[string]interface{}    `yaml:"volumes,omitempty"`
}

// PrepareComposeFile prepares the compose file based on the source type
func (s *TaskService) PrepareComposeFile(cfg ComposeConfig) (string, error) {
	switch cfg.SourceType {
	case ComposeSourceRepository:
		return s.prepareComposeFromRepository(cfg)
	case ComposeSourceURL:
		return s.prepareComposeFromURL(cfg)
	case ComposeSourcePath:
		return s.prepareComposeFromPath(cfg)
	case ComposeSourceRaw:
		return s.prepareComposeFromRaw(cfg)
	default:
		return "", fmt.Errorf("unknown compose source type: %s", cfg.SourceType)
	}
}

// prepareComposeFromRepository clones the repository and returns path to compose file
func (s *TaskService) prepareComposeFromRepository(cfg ComposeConfig) (string, error) {
	cfg.TaskContext.AddLog("Cloning repository for Docker Compose deployment")

	repoPath, err := s.Clone(CloneConfig{
		TaskPayload:    cfg.TaskPayload,
		DeploymentType: string(shared_types.DeploymentTypeCreate),
		TaskContext:    cfg.TaskContext,
	})
	if err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	composeFilePath := cfg.ComposeFilePath
	if composeFilePath == "" {
		composeFilePath = "docker-compose.yml"
	}

	fullPath := filepath.Join(repoPath, composeFilePath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		alternates := []string{"docker-compose.yaml", "compose.yml", "compose.yaml"}
		found := false
		for _, alt := range alternates {
			altPath := filepath.Join(repoPath, alt)
			if _, err := os.Stat(altPath); err == nil {
				fullPath = altPath
				found = true
				break
			}
		}
		if !found {
			return "", types.ErrDockerComposeFileNotFound
		}
	}

	cfg.TaskContext.AddLog("Found compose file at: " + fullPath)
	return fullPath, nil
}

// prepareComposeFromURL downloads compose file from URL and saves it
func (s *TaskService) prepareComposeFromURL(cfg ComposeConfig) (string, error) {
	cfg.TaskContext.AddLog("Downloading Docker Compose file from URL: " + cfg.ComposeURL)

	resp, err := http.Get(cfg.ComposeURL)
	if err != nil {
		return "", fmt.Errorf("failed to download compose file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download compose file: status %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read compose file content: %w", err)
	}

	composeDir := s.getComposeDir(cfg.TaskPayload)
	if err := os.MkdirAll(composeDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create compose directory: %w", err)
	}

	composePath := filepath.Join(composeDir, "docker-compose.yml")
	if err := os.WriteFile(composePath, content, 0644); err != nil {
		return "", fmt.Errorf("failed to write compose file: %w", err)
	}

	cfg.TaskContext.AddLog("Compose file saved to: " + composePath)
	return composePath, nil
}

// prepareComposeFromPath validates and returns the provided path
func (s *TaskService) prepareComposeFromPath(cfg ComposeConfig) (string, error) {
	cfg.TaskContext.AddLog("Using compose file from path: " + cfg.ComposeFilePath)

	if _, err := os.Stat(cfg.ComposeFilePath); os.IsNotExist(err) {
		return "", types.ErrDockerComposeFileNotFound
	}

	return cfg.ComposeFilePath, nil
}

// prepareComposeFromRaw saves raw compose content to a file
func (s *TaskService) prepareComposeFromRaw(cfg ComposeConfig) (string, error) {
	cfg.TaskContext.AddLog("Creating compose file from raw content")

	if cfg.ComposeRaw == "" {
		return "", types.ErrDockerComposeInvalidConfig
	}

	var compose ComposeFile
	if err := yaml.Unmarshal([]byte(cfg.ComposeRaw), &compose); err != nil {
		return "", fmt.Errorf("invalid compose file content: %w", err)
	}

	if len(compose.Services) == 0 {
		return "", fmt.Errorf("compose file has no services defined")
	}

	composeDir := s.getComposeDir(cfg.TaskPayload)
	if err := os.MkdirAll(composeDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create compose directory: %w", err)
	}

	composePath := filepath.Join(composeDir, "docker-compose.yml")
	if err := os.WriteFile(composePath, []byte(cfg.ComposeRaw), 0644); err != nil {
		return "", fmt.Errorf("failed to write compose file: %w", err)
	}

	cfg.TaskContext.AddLog("Compose file saved to: " + composePath)
	return composePath, nil
}

// getComposeDir returns the directory for storing compose files
func (s *TaskService) getComposeDir(payload shared_types.TaskPayload) string {
	mountPath := os.Getenv("MOUNT_PATH")
	if mountPath == "" {
		mountPath = "/var/lib/nixopus"
	}
	return filepath.Join(mountPath, payload.Application.UserID.String(),
		string(payload.Application.Environment), payload.Application.ID.String())
}

// ComposeUp starts all services defined in the compose file
func (s *TaskService) ComposeUp(cfg ComposeConfig, composePath string) error {
	cfg.TaskContext.AddLog("Starting Docker Compose services")

	envVars := s.prepareComposeEnvVars(cfg)
	err := s.DockerRepo.ComposeUp(composePath, envVars)
	if err != nil {
		return fmt.Errorf("compose up failed: %w", err)
	}

	cfg.TaskContext.AddLog("Docker Compose services started successfully")
	return nil
}

// ComposeDown stops and removes all services
func (s *TaskService) ComposeDown(cfg ComposeConfig, composePath string) error {
	cfg.TaskContext.AddLog("Stopping Docker Compose services")

	err := s.DockerRepo.ComposeDown(composePath)
	if err != nil {
		return fmt.Errorf("compose down failed: %w", err)
	}

	cfg.TaskContext.AddLog("Docker Compose services stopped successfully")
	return nil
}

// ComposeBuild builds all services defined in the compose file
func (s *TaskService) ComposeBuild(cfg ComposeConfig, composePath string) error {
	cfg.TaskContext.AddLog("Building Docker Compose services")

	envVars := s.prepareComposeEnvVars(cfg)
	err := s.DockerRepo.ComposeBuild(composePath, envVars)
	if err != nil {
		return fmt.Errorf("compose build failed: %w", err)
	}

	cfg.TaskContext.AddLog("Docker Compose services built successfully")
	return nil
}

// prepareComposeEnvVars prepares environment variables for compose commands
func (s *TaskService) prepareComposeEnvVars(cfg ComposeConfig) map[string]string {
	envVars := make(map[string]string)

	appEnvVars := GetMapFromString(cfg.Application.EnvironmentVariables)
	for k, v := range appEnvVars {
		envVars[k] = v
	}

	buildVars := GetMapFromString(cfg.Application.BuildVariables)
	for k, v := range buildVars {
		envVars[k] = v
	}

	for k, v := range cfg.EnvVars {
		envVars[k] = v
	}

	return envVars
}

// AddComposeReverseProxy configures reverse proxy for compose services with domains
func (s *TaskService) AddComposeReverseProxy(cfg ComposeConfig, composePath string) error {
	if cfg.Application.Domain == "" {
		cfg.TaskContext.AddLog("No domain configured, skipping reverse proxy setup")
		return nil
	}

	cfg.TaskContext.AddLog("Configuring reverse proxy for domain: " + cfg.Application.Domain)

	composeContent, err := os.ReadFile(composePath)
	if err != nil {
		return fmt.Errorf("failed to read compose file: %w", err)
	}

	var compose ComposeFile
	if err := yaml.Unmarshal(composeContent, &compose); err != nil {
		return fmt.Errorf("failed to parse compose file: %w", err)
	}

	var primaryPort int
	for serviceName, service := range compose.Services {
		if len(service.Ports) > 0 {
			port, err := parseComposePort(service.Ports[0])
			if err != nil {
				s.Logger.Log(logger.Warning, "Failed to parse port for service "+serviceName, err.Error())
				continue
			}
			primaryPort = port
			cfg.TaskContext.AddLog(fmt.Sprintf("Found exposed port %d on service %s", port, serviceName))
			break
		}
	}

	if primaryPort == 0 {
		primaryPort = cfg.Application.Port
	}

	if primaryPort == 0 {
		cfg.TaskContext.AddLog("No exposed port found, skipping reverse proxy setup")
		return nil
	}

	client := GetCaddyClient()
	upstreamHost := config.AppConfig.SSH.Host

	err = client.AddDomainWithAutoTLS(cfg.Application.Domain, upstreamHost, primaryPort, caddygo.DomainOptions{})
	if err != nil {
		return fmt.Errorf("failed to add domain: %w", err)
	}

	client.Reload()
	cfg.TaskContext.AddLog("Reverse proxy configured successfully for " + cfg.Application.Domain)

	return nil
}

// parseComposePort extracts the host port from a docker-compose port mapping
// Formats: "8080", "8080:80", "127.0.0.1:8080:80"
func parseComposePort(portMapping string) (int, error) {
	parts := strings.Split(portMapping, ":")
	var portStr string

	switch len(parts) {
	case 1:
		portStr = parts[0]
	case 2:
		portStr = parts[0]
	case 3:
		portStr = parts[1]
	default:
		return 0, fmt.Errorf("invalid port mapping: %s", portMapping)
	}

	portStr = strings.Split(portStr, "/")[0]

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("invalid port number: %s", portStr)
	}

	return port, nil
}

// RemoveComposeReverseProxy removes the reverse proxy configuration for the application
func (s *TaskService) RemoveComposeReverseProxy(domain string) error {
	if domain == "" {
		return nil
	}

	client := GetCaddyClient()
	err := client.DeleteDomain(domain)
	if err != nil {
		s.Logger.Log(logger.Warning, "Failed to remove domain from reverse proxy", err.Error())
		return err
	}

	client.Reload()
	return nil
}

// GetComposeSourceType determines the source type from application config
func GetComposeSourceType(app shared_types.Application) ComposeSourceType {
	if app.ComposeFileContent != "" {
		return ComposeSourceRaw
	}

	if app.ComposeFileURL != "" {
		return ComposeSourceURL
	}

	if app.Repository != "" {
		if _, err := strconv.ParseInt(app.Repository, 10, 64); err == nil {
			return ComposeSourceRepository
		}
	}

	return ComposeSourceRepository
}

// ComposeRestart restarts all compose services (down + up)
func (s *TaskService) ComposeRestart(cfg ComposeConfig, composePath string) error {
	cfg.TaskContext.AddLog("Restarting Docker Compose services")

	if err := s.DockerRepo.ComposeDown(composePath); err != nil {
		cfg.TaskContext.AddLog("Warning: Failed to stop services: " + err.Error())
	}

	envVars := s.prepareComposeEnvVars(cfg)
	if err := s.DockerRepo.ComposeUp(composePath, envVars); err != nil {
		return fmt.Errorf("compose restart failed: %w", err)
	}

	cfg.TaskContext.AddLog("Docker Compose services restarted successfully")
	return nil
}

// ComposeDelete removes all compose resources and cleans up
func (s *TaskService) ComposeDelete(payload shared_types.TaskPayload) error {
	s.Logger.Log(logger.Info, "Deleting Docker Compose deployment", payload.Application.ID.String())

	composeDir := s.getComposeDir(payload)
	composePath := filepath.Join(composeDir, "docker-compose.yml")

	if _, err := os.Stat(composePath); err == nil {
		if err := s.DockerRepo.ComposeDown(composePath); err != nil {
			s.Logger.Log(logger.Warning, "Failed to stop compose services", err.Error())
		}
	}

	if payload.Application.Domain != "" {
		if err := s.RemoveComposeReverseProxy(payload.Application.Domain); err != nil {
			s.Logger.Log(logger.Warning, "Failed to remove reverse proxy", err.Error())
		}
	}

	if err := os.RemoveAll(composeDir); err != nil {
		s.Logger.Log(logger.Warning, "Failed to remove compose directory", err.Error())
	}

	s.Logger.Log(logger.Info, "Docker Compose deployment deleted", payload.Application.ID.String())
	return nil
}

// GetComposeFilePath returns the compose file path for an application
func (s *TaskService) GetComposeFilePath(payload shared_types.TaskPayload) string {
	composeDir := s.getComposeDir(payload)
	return filepath.Join(composeDir, "docker-compose.yml")
}

// IsComposeDeployment checks if the application uses docker-compose build pack
func IsComposeDeployment(app shared_types.Application) bool {
	return app.BuildPack == shared_types.DockerCompose
}
