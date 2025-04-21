package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/docker/docker/api/types/image"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/update/types"
	"github.com/raghavyuva/nixopus-api/internal/storage"
)

type UpdateService struct {
	storage *storage.App
	logger  *logger.Logger
	ctx     context.Context
	docker  *docker.DockerService
}

func NewUpdateService(storage *storage.App, logger *logger.Logger, ctx context.Context) *UpdateService {
	return &UpdateService{
		storage: storage,
		logger:  logger,
		ctx:     ctx,
		docker:  docker.NewDockerService(),
	}
}

type GitHubContainerVersion struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Metadata struct {
		Container struct {
			Tags []string `json:"tags"`
		} `json:"container"`
	} `json:"metadata"`
}

func (s *UpdateService) CheckForUpdates() (*types.UpdateCheckResponse, error) {
	images := s.docker.ListAllImages(image.ListOptions{})

	var currentVersion string
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == "nixopus-api:latest" {
				currentVersion = tag
				break
			}
		}
		if currentVersion != "" {
			break
		}
	}

	if currentVersion == "" {
		return nil, fmt.Errorf("failed to find current container version")
	}

	resp, err := http.Get("https://api.github.com/users/raghavyuva/packages/container/nixopus-api/versions")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch versions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch versions: status %d", resp.StatusCode)
	}

	var versions []GitHubContainerVersion
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, fmt.Errorf("failed to decode versions: %w", err)
	}

	latestVersion := currentVersion
	updateAvailable := false

	for _, version := range versions {
		for _, tag := range version.Metadata.Container.Tags {
			if tag == "latest" {
				continue
			}
			if tag > latestVersion {
				latestVersion = tag
				updateAvailable = true
			}
		}
	}

	return &types.UpdateCheckResponse{
		CurrentVersion:  currentVersion,
		LatestVersion:   latestVersion,
		UpdateAvailable: updateAvailable,
		LastChecked:     time.Now(),
	}, nil
}

// TODO: Implement update service
func (s *UpdateService) PerformUpdate() error {
	return nil
}

func (s *UpdateService) GetUserAutoUpdatePreference(userID uuid.UUID) (bool, error) {
	var autoUpdate bool

	err := s.storage.Store.DB.NewSelect().
		TableExpr("user_settings").
		Column("auto_update").
		Where("user_id = ?", userID).
		Scan(s.ctx, &autoUpdate)

	if err != nil {
		return false, fmt.Errorf("failed to get user settings: %w", err)
	}

	return autoUpdate, nil
}
