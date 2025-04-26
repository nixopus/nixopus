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
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	"github.com/raghavyuva/nixopus-api/internal/features/update/types"
	"github.com/raghavyuva/nixopus-api/internal/storage"
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
	docker  *docker.DockerService
	env     Environment
}

type PathConfig struct {
	SourceDir   string
	BackupDir   string
	ComposeFile string
}

func NewUpdateService(storage *storage.App, logger *logger.Logger, ctx context.Context) *UpdateService {
	return &UpdateService{
		storage: storage,
		logger:  logger,
		ctx:     ctx,
		docker:  docker.NewDockerService(),
		env:     getEnvironment(),
	}
}

func getEnvironment() Environment {
	if os.Getenv("DOCKER_CONTEXT") == "nixopus-staging" {
		return Staging
	}
	return Production
}

func (s *UpdateService) getPathConfig() PathConfig {
	baseDir := "/etc/nixopus"
	composeFile := "docker-compose.yml"
	if s.env == Staging {
		baseDir = "/etc/nixopus-staging"
		composeFile = "docker-compose-staging.yml"
	}
	return PathConfig{
		SourceDir:   baseDir + "/source",
		BackupDir:   fmt.Sprintf("%s/source_backups/backup-%s", baseDir, time.Now().Format("20060102-150405")),
		ComposeFile: composeFile,
	}
}

// CheckForUpdates checks for updates for the current environment and returns the current and latest version
func (s *UpdateService) CheckForUpdates() (*types.UpdateCheckResponse, error) {
	currentVersion, err := s.getCurrentVersion()
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

// getCurrentVersion gets the current version from the .env file
func (s *UpdateService) getCurrentVersion() (string, error) {
	version := os.Getenv("APP_VERSION")
	if version != "" {
		return version, nil
	}

	return "", fmt.Errorf("APP_VERSION not found in .env file")
}

// fetchLatestVersion fetches the latest version from the appropriate branch from our repo
func (s *UpdateService) fetchLatestVersion() (string, error) {
	branch := s.getBranch()
	url := fmt.Sprintf("https://raw.githubusercontent.com/raghavyuva/nixopus/refs/heads/%s/version.txt", branch)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch version: status %d", resp.StatusCode)
	}

	versionBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(versionBytes)), nil
}

func (s *UpdateService) getBranch() string {
	if s.env == Staging {
		return "feat/develop"
	}
	return "master"
}

// PerformUpdate performs an update for the current environment
func (s *UpdateService) PerformUpdate() error {
	ssh := ssh.NewSSH()
	client, err := ssh.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect via SSH: %w", err)
	}
	defer client.Close()

	paths := s.getPathConfig()
	updateSuccess := false

	defer func() {
		if !updateSuccess {
			s.handleRollback(ssh, paths)
		} else {
			s.cleanupBackup(ssh, paths)
		}
	}()

	if err := s.sourceCodeDirectoryCheck(ssh, paths.SourceDir); err != nil {
		return err
	}

	if err := s.createBackup(ssh, paths); err != nil {
		return err
	}

	if err := s.updateRepository(ssh, paths); err != nil {
		return err
	}

	if err := s.startContainers(ssh, paths); err != nil {
		return err
	}

	if err := s.verifyServices(ssh, paths); err != nil {
		return err
	}

	updateSuccess = true
	s.logger.Log(logger.Info, "Update completed successfully", "")
	return nil
}

func (s *UpdateService) sourceCodeDirectoryCheck(ssh *ssh.SSH, sourceDir string) error {
	dirCheckCmd := fmt.Sprintf("test -d %s && echo 'exists' || echo 'missing'", sourceDir)
	dirCheckOutput, err := ssh.RunCommand(dirCheckCmd)
	if err != nil {
		return fmt.Errorf("failed to check source directory: %w", err)
	}

	if strings.TrimSpace(dirCheckOutput) == "missing" {
		if _, err := ssh.RunCommand(fmt.Sprintf("mkdir -p %s", sourceDir)); err != nil {
			return fmt.Errorf("failed to create source directory: %w", err)
		}
	}
	return nil
}

func (s *UpdateService) createBackup(ssh *ssh.SSH, paths PathConfig) error {
	checkContainersCmd := fmt.Sprintf("cd %s && docker compose -f %s ps -q 2>/dev/null | wc -l", paths.SourceDir, paths.ComposeFile)
	containersOutput, err := ssh.RunCommand(checkContainersCmd)
	if err == nil && strings.TrimSpace(containersOutput) != "0" {
		s.logger.Log(logger.Info, "Creating backup of current deployment", paths.BackupDir)

		backupDirCmd := fmt.Sprintf("mkdir -p %s", paths.BackupDir)
		if _, err := ssh.RunCommand(backupDirCmd); err != nil {
			return fmt.Errorf("failed to create backup directory: %w", err)
		}

		backupCmd := fmt.Sprintf("cp -a %s %s", paths.SourceDir, paths.BackupDir)
		if _, err := ssh.RunCommand(backupCmd); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}
	return nil
}

func (s *UpdateService) updateRepository(ssh *ssh.SSH, paths PathConfig) error {
	gitDirCheckCmd := fmt.Sprintf("test -d %s/.git && echo 'exists' || echo 'missing'", paths.SourceDir)
	gitDirOutput, err := ssh.RunCommand(gitDirCheckCmd)
	if err != nil {
		return fmt.Errorf("failed to check git directory: %w", err)
	}

	if strings.TrimSpace(gitDirOutput) == "exists" {
		return s.updateExistingRepository(ssh, paths)
	}
	return s.cloneRepository(ssh, paths)
}

func (s *UpdateService) updateExistingRepository(ssh *ssh.SSH, paths PathConfig) error {
	s.logger.Log(logger.Info, "Updating existing repository", paths.SourceDir)
	branch := s.getBranch()
	fetchCmd := fmt.Sprintf("cd %s && git fetch --all && git reset --hard origin/%s && git checkout %s && git pull 2>&1",
		paths.SourceDir, branch, branch)
	fetchOutput, err := ssh.RunCommand(fetchCmd)
	if err != nil {
		s.logger.Log(logger.Error, "Git update failed", fmt.Sprintf("output: %s, error: %v", fetchOutput, err))
		return fmt.Errorf("failed to update repository: %w (output: %s)", err, fetchOutput)
	}
	return nil
}

func (s *UpdateService) cloneRepository(ssh *ssh.SSH, paths PathConfig) error {
	s.logger.Log(logger.Info, "Cloning repository", paths.SourceDir)
	if _, err := ssh.RunCommand(fmt.Sprintf("rm -rf %s/* %s/.[!.]*", paths.SourceDir, paths.SourceDir)); err != nil {
		return fmt.Errorf("failed to clean source directory: %w", err)
	}

	cloneCmd := fmt.Sprintf("cd %s && GIT_TERMINAL_PROMPT=0 git clone --verbose https://github.com/raghavyuva/nixopus.git . 2>&1", paths.SourceDir)
	cloneOutput, err := ssh.RunCommand(cloneCmd)
	if err != nil {
		s.logger.Log(logger.Error, "Git clone failed", fmt.Sprintf("output: %s, error: %v", cloneOutput, err))
		return fmt.Errorf("failed to clone repository: %w (output: %s)", err, cloneOutput)
	}

	branch := s.getBranch()
	checkoutCmd := fmt.Sprintf("cd %s && git checkout %s 2>&1", paths.SourceDir, branch)
	checkoutOutput, err := ssh.RunCommand(checkoutCmd)
	if err != nil {
		s.logger.Log(logger.Error, "Git checkout failed", fmt.Sprintf("output: %s, error: %v", checkoutOutput, err))
		return fmt.Errorf("failed to checkout %s branch: %w (output: %s)", branch, err, checkoutOutput)
	}
	return nil
}

func (s *UpdateService) startContainers(ssh *ssh.SSH, paths PathConfig) error {
	var startCmd string
	if s.env == Staging {
		startCmd = fmt.Sprintf("cd %s && DOCKER_HOST=unix:///var/run/docker.sock DOCKER_CONTEXT=nixopus-staging docker compose -f %s up -d --build 2>&1", paths.SourceDir, paths.ComposeFile)
	} else {
		startCmd = fmt.Sprintf("cd %s && DOCKER_HOST=unix:///var/run/docker.sock DOCKER_CONTEXT=nixopus docker compose -f %s up -d 2>&1", paths.SourceDir, paths.ComposeFile)
	}

	startOutput, err := ssh.RunCommand(startCmd)
	if err != nil {
		s.logger.Log(logger.Error, "Container start failed", fmt.Sprintf("output: %s, error: %v", startOutput, err))
		return fmt.Errorf("failed to start containers: %w (output: %s)", err, startOutput)
	}
	return nil
}

func (s *UpdateService) verifyServices(ssh *ssh.SSH, paths PathConfig) error {
	time.Sleep(10 * time.Second)
	checkCmd := fmt.Sprintf("cd %s && docker compose -f %s ps --format json 2>&1", paths.SourceDir, paths.ComposeFile)
	checkOutput, err := ssh.RunCommand(checkCmd)
	if err != nil {
		s.logger.Log(logger.Error, "Service verification failed", checkOutput)
		return fmt.Errorf("failed to verify services: %w", err)
	}
	return nil
}

func (s *UpdateService) handleRollback(ssh *ssh.SSH, paths PathConfig) {
	backupExistsCmd := fmt.Sprintf("test -d %s && echo 'exists' || echo 'missing'", paths.BackupDir)
	backupExists, _ := ssh.RunCommand(backupExistsCmd)

	if strings.TrimSpace(backupExists) == "exists" {
		s.logger.Log(logger.Warning, "Update failed, rolling back to previous version", "")

		restoreCmd := fmt.Sprintf("rm -rf %s && mv %s %s", paths.SourceDir, paths.BackupDir, paths.SourceDir)
		if _, err := ssh.RunCommand(restoreCmd); err != nil {
			s.logger.Log(logger.Error, "Failed to restore from backup", err.Error())
			return
		}

		if err := s.startContainers(ssh, paths); err != nil {
			s.logger.Log(logger.Error, "Failed to restart previous version", err.Error())
		} else {
			s.logger.Log(logger.Info, "Successfully rolled back to previous version", "")
		}
	}
}

func (s *UpdateService) cleanupBackup(ssh *ssh.SSH, paths PathConfig) {
	backupExistsCmd := fmt.Sprintf("test -d %s && echo 'exists' || echo 'missing'", paths.BackupDir)
	backupExists, _ := ssh.RunCommand(backupExistsCmd)
	if strings.TrimSpace(backupExists) == "exists" {
		if _, err := ssh.RunCommand(fmt.Sprintf("rm -rf %s", paths.BackupDir)); err != nil {
			s.logger.Log(logger.Warning, "Failed to remove backup directory", err.Error())
		}
	}
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
