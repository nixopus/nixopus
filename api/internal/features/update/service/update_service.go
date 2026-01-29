package service

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	"github.com/raghavyuva/nixopus-api/internal/features/update/types"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

const (
	// VersionFilePath is the path to the version.txt file inside the container
	// The source is mounted at /etc/nixopus/source/ via docker-compose
	VersionFilePath = "/etc/nixopus/source/version.txt"
)

type Environment string

const (
	Production Environment = "production"
	Staging    Environment = "staging"
)

type UpdateService struct {
	storage *storage.App
	logger  *logger.Logger
	ctx     context.Context
	env     Environment
}

func NewUpdateService(storage *storage.App, logger *logger.Logger, ctx context.Context) *UpdateService {
	return &UpdateService{
		storage: storage,
		logger:  logger,
		ctx:     ctx,
		env:     getEnvironment(),
	}
}

func getEnvironment() Environment {
	if config.AppConfig.Docker.Context == "nixopus-staging" {
		return Staging
	}
	return Production
}

// CheckForUpdates checks for updates for the current environment and returns the current and latest version
func (s *UpdateService) CheckForUpdates() (*types.UpdateCheckResponse, error) {
	currentVersion, err := s.GetCurrentVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get current version: %w", err)
	}

	latestVersion, err := s.fetchLatestVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest version: %w", err)
	}

	updateAvailable := currentVersion != latestVersion && latestVersion != ""

	return &types.UpdateCheckResponse{
		CurrentVersion:  currentVersion,
		LatestVersion:   latestVersion,
		UpdateAvailable: updateAvailable,
		LastChecked:     time.Now(),
		Environment:     string(s.env),
	}, nil
}

// GetCurrentVersion gets the current version from config, or falls back to reading version.txt
// Returns an error if all sources fail instead of silently returning "unknown"
func (s *UpdateService) GetCurrentVersion() (string, error) {
	// First, try to get version from config
	if version := config.AppConfig.App.Version; version != "" {
		return version, nil
	}

	// Fallback: try to read from version.txt in the repository root
	version, err := s.readVersionFromFile()
	if err == nil {
		return version, nil
	}

	// If all sources fail, return an error instead of silently returning "unknown"
	return "", fmt.Errorf("failed to get version: APP_VERSION not set in config and version.txt not found in any expected location")
}

// readVersionFromFile attempts to read version.txt from various possible locations
func (s *UpdateService) readVersionFromFile() (string, error) {
	paths := s.getVersionFilePaths()

	for _, path := range paths {
		version, err := s.tryReadVersionFile(path)
		if err == nil {
			return version, nil
		}
	}

	return "", fmt.Errorf("version.txt not found in any expected location")
}

// tryReadVersionFile attempts to read and parse version.txt from the given path
func (s *UpdateService) tryReadVersionFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	version := strings.TrimSpace(string(data))
	version = strings.TrimPrefix(version, "v") // Remove 'v' prefix if present

	if version == "" {
		return "", fmt.Errorf("version.txt is empty")
	}

	s.logger.Log(logger.Info, "Read version from file", fmt.Sprintf("path: %s, version: %s", path, version))
	return version, nil
}

// getVersionFilePaths returns a list of potential paths to version.txt, ordered by priority
func (s *UpdateService) getVersionFilePaths() []string {
	var paths []string

	// 1. Check deployment source directories (highest priority for deployed instances)
	paths = append(paths, VersionFilePath)
	if s.env == Staging {
		paths = append(paths, "/etc/nixopus-staging/source/version.txt")
	}

	// 2. Try to find repository root by walking up from current working directory
	paths = append(paths, s.findVersionPathsFromCWD()...)

	// 3. Check paths relative to executable location
	paths = append(paths, s.findVersionPathsFromExecutable()...)

	return paths
}

// findVersionPathsFromCWD walks up from the current working directory to find version.txt
func (s *UpdateService) findVersionPathsFromCWD() []string {
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}

	var paths []string
	dir := cwd
	const maxDepth = 10

	for i := 0; i < maxDepth; i++ {
		paths = append(paths, filepath.Join(dir, "version.txt"))

		// Stop if we've reached the repository root
		if s.isRepositoryRoot(dir) {
			break
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached filesystem root
		}
		dir = parent
	}

	return paths
}

// findVersionPathsFromExecutable returns version.txt paths relative to the executable location
func (s *UpdateService) findVersionPathsFromExecutable() []string {
	execPath, err := os.Executable()
	if err != nil {
		return nil
	}

	execDir := filepath.Dir(execPath)
	var paths []string

	// Check executable directory and parent directory
	paths = append(paths, filepath.Join(execDir, "version.txt"))

	parent := filepath.Dir(execDir)
	if parent != execDir {
		paths = append(paths, filepath.Join(parent, "version.txt"))
	}

	return paths
}

// isRepositoryRoot checks if the given directory is likely the repository root
func (s *UpdateService) isRepositoryRoot(dir string) bool {
	// Check for common repository root markers
	markers := []string{".git", "api"}
	for _, marker := range markers {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
			return true
		}
	}
	return false
}

// fetchLatestVersion fetches the latest version from the appropriate branch from our repo
func (s *UpdateService) fetchLatestVersion() (string, error) {
	branch := s.getBranch()
	s.logger.Log(logger.Info, "Fetching latest version", fmt.Sprintf("Using branch: %s", branch))

	url := fmt.Sprintf("https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/%s/version.txt", branch)
	s.logger.Log(logger.Info, "Constructed version URL", url)

	resp, err := http.Get(url)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to fetch version", fmt.Sprintf("Error: %v", err))
		return "", err
	}
	defer resp.Body.Close()

	s.logger.Log(logger.Info, "Version fetch response", fmt.Sprintf("Status: %d", resp.StatusCode))
	if resp.StatusCode != http.StatusOK {
		s.logger.Log(logger.Error, "Failed to fetch version", fmt.Sprintf("Status code: %d", resp.StatusCode))
		return "", fmt.Errorf("failed to fetch version: status %d", resp.StatusCode)
	}

	versionBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to read version response", fmt.Sprintf("Error: %v", err))
		return "", err
	}

	version := strings.TrimSpace(string(versionBytes))
	s.logger.Log(logger.Info, "Successfully fetched version", version)
	return version, nil
}

func (s *UpdateService) getBranch() string {
	// TODO: add support for staging update
	// if s.env == Staging {
	// 	return "feat/develop"
	// }
	return "master"
}

// PerformUpdate performs an update by running the nixopus update CLI command via SSH on the host
// Requires organization context to get organization-specific SSH configuration
func (s *UpdateService) PerformUpdate(ctx context.Context) error {
	s.logger.Log(logger.Info, "Starting Nixopus update", "Connecting via SSH to run nixopus update")

	// Get organization ID from context
	orgIDAny := ctx.Value(shared_types.OrganizationIDKey)
	if orgIDAny == nil {
		return fmt.Errorf("organization ID not found in context - system updates require organization context")
	}

	var orgID uuid.UUID
	switch v := orgIDAny.(type) {
	case string:
		var err error
		orgID, err = uuid.Parse(v)
		if err != nil {
			return fmt.Errorf("invalid organization ID in context: %w", err)
		}
	case uuid.UUID:
		orgID = v
	default:
		return fmt.Errorf("unexpected organization ID type: %T", orgIDAny)
	}

	// Get organization-specific SSH manager
	manager, err := ssh.GetSSHManagerForOrganization(ctx, orgID)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to get SSH manager", err.Error())
		return fmt.Errorf("failed to get SSH manager: %w", err)
	}

	sshClient, err := manager.GetOrganizationSSH()
	if err != nil {
		s.logger.Log(logger.Error, "Failed to get SSH client", err.Error())
		return fmt.Errorf("failed to get SSH client: %w", err)
	}

	client, err := sshClient.Connect()
	if err != nil {
		s.logger.Log(logger.Error, "Failed to connect via SSH", err.Error())
		return fmt.Errorf("failed to connect via SSH: %w", err)
	}
	defer client.Close()

	s.logger.Log(logger.Info, "SSH connected", "Running nixopus update command")

	output, err := sshClient.RunCommand("nixopus update")
	if err != nil {
		s.logger.Log(logger.Error, "Update failed", fmt.Sprintf("error: %v, output: %s", err, output))
		return fmt.Errorf("failed to run nixopus update: %w (output: %s)", err, output)
	}

	s.logger.Log(logger.Info, "Update completed successfully", output)
	return nil
}

// GetUserAutoUpdatePreference gets the user's auto update preference from the database
func (s *UpdateService) GetUserAutoUpdatePreference(userID uuid.UUID) (bool, error) {
	var autoUpdate bool

	err := s.storage.Store.DB.NewSelect().
		TableExpr("user_settings").
		Column("auto_update").
		Where("user_id = ?", userID).
		Scan(s.ctx, &autoUpdate)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to get user settings: %w", err)
	}

	return autoUpdate, nil
}
