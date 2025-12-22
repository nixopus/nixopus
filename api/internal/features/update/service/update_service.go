package service

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	"github.com/raghavyuva/nixopus-api/internal/features/update/types"
	"github.com/raghavyuva/nixopus-api/internal/storage"
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

// GetCurrentVersion gets the current version from the version.txt file
// The file is mounted into the container at /etc/nixopus/source/version.txt
func (s *UpdateService) GetCurrentVersion() (string, error) {
	// First try to read from the mounted version.txt file (production)
	if versionBytes, err := os.ReadFile(VersionFilePath); err == nil {
		if version := strings.TrimSpace(string(versionBytes)); version != "" {
			return version, nil
		}
	}

	// Try local development path
	if versionBytes, err := os.ReadFile("../version.txt"); err == nil {
		if version := strings.TrimSpace(string(versionBytes)); version != "" {
			return version, nil
		}
	}

	// Fallback to environment variable if file read fails
	if version := config.AppConfig.App.Version; version != "" {
		return version, nil
	}

	return "unknown", nil
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
func (s *UpdateService) PerformUpdate() error {
	s.logger.Log(logger.Info, "Starting Nixopus update", "Connecting via SSH to run nixopus update")

	sshClient := ssh.NewSSH()
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
