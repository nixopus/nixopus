package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/config"
	s3store "github.com/nixopus/nixopus/api/internal/features/deploy/s3"
	"github.com/nixopus/nixopus/api/internal/features/deploy/types"
)

func (s *DeployService) ListArtifacts(applicationID uuid.UUID, organizationID uuid.UUID) ([]types.Artifact, error) {
	if !s3store.IsConfigured(config.AppConfig.S3) {
		return nil, types.ErrS3NotConfigured
	}

	app, err := s.storage.GetApplicationById(applicationID.String(), organizationID)
	if err != nil {
		return nil, types.ErrPermissionDenied
	}

	deployments, err := s.storage.GetApplicationDeployments(applicationID)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	var artifacts []types.Artifact
	for _, d := range deployments {
		if d.ImageS3Key == "" {
			continue
		}
		artifacts = append(artifacts, types.Artifact{
			DeploymentID:  d.ID.String(),
			ApplicationID: app.ID.String(),
			AppName:       app.Name,
			S3Key:         d.ImageS3Key,
			Size:          d.ImageSize,
			CreatedAt:     d.CreatedAt.Format(time.RFC3339),
		})
	}
	return artifacts, nil
}

func (s *DeployService) GetArtifactDownloadURL(deploymentID string, organizationID uuid.UUID) (string, error) {
	if !s3store.IsConfigured(config.AppConfig.S3) {
		return "", types.ErrS3NotConfigured
	}

	deployment, err := s.storage.GetApplicationDeploymentById(deploymentID)
	if err != nil {
		return "", fmt.Errorf("deployment not found: %w", err)
	}

	app, err := s.storage.GetApplicationById(deployment.ApplicationID.String(), organizationID)
	if err != nil {
		return "", types.ErrPermissionDenied
	}
	_ = app

	if deployment.ImageS3Key == "" {
		return "", fmt.Errorf("no artifact available for this deployment")
	}

	store, err := s3store.NewImageStore(config.AppConfig.S3)
	if err != nil {
		return "", fmt.Errorf("failed to create S3 client: %w", err)
	}

	url, err := store.PresignedDownloadURL(context.Background(), deployment.ImageS3Key, 15*time.Minute)
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}
	return url, nil
}

func (s *DeployService) DeleteArtifact(deploymentID string, organizationID uuid.UUID) error {
	if !s3store.IsConfigured(config.AppConfig.S3) {
		return types.ErrS3NotConfigured
	}

	deployment, err := s.storage.GetApplicationDeploymentById(deploymentID)
	if err != nil {
		return fmt.Errorf("deployment not found: %w", err)
	}

	app, err := s.storage.GetApplicationById(deployment.ApplicationID.String(), organizationID)
	if err != nil {
		return types.ErrPermissionDenied
	}
	_ = app

	if deployment.ImageS3Key == "" {
		return fmt.Errorf("no artifact to delete for this deployment")
	}

	store, err := s3store.NewImageStore(config.AppConfig.S3)
	if err != nil {
		return fmt.Errorf("failed to create S3 client: %w", err)
	}

	if err := store.DeleteImage(context.Background(), deployment.ImageS3Key); err != nil {
		return fmt.Errorf("failed to delete artifact from S3: %w", err)
	}

	deployment.ImageS3Key = ""
	deployment.ImageSize = 0
	return s.storage.UpdateApplicationDeployment(&deployment)
}
