package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/config"
	s3store "github.com/nixopus/nixopus/api/internal/features/deploy/s3"
	"github.com/nixopus/nixopus/api/internal/features/deploy/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

func (s *DeployService) ListArtifacts(applicationID uuid.UUID, organizationID uuid.UUID) ([]types.Artifact, error) {
	app, err := s.storage.GetApplicationBasicById(applicationID.String(), organizationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "not found") {
			return nil, types.ErrPermissionDenied
		}
		s.logger.Log(logger.Error, "failed to get application for artifact listing", err.Error())
		return nil, err
	}

	deployments, err := s.storage.GetApplicationDeployments(applicationID)
	if err != nil {
		s.logger.Log(logger.Error, "failed to list deployments", err.Error())
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	artifacts := make([]types.Artifact, 0, len(deployments))
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

func (s *DeployService) GetArtifactDownloadURL(ctx context.Context, deploymentID string, organizationID uuid.UUID) (string, error) {
	if !s3store.IsConfigured(config.AppConfig.S3) {
		return "", types.ErrS3NotConfigured
	}

	deployment, err := s.storage.GetApplicationDeploymentById(deploymentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("deployment not found")
		}
		s.logger.Log(logger.Error, "failed to get deployment for download", err.Error())
		return "", err
	}

	_, err = s.storage.GetApplicationBasicById(deployment.ApplicationID.String(), organizationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "not found") {
			return "", types.ErrPermissionDenied
		}
		s.logger.Log(logger.Error, "failed to verify app ownership for download", err.Error())
		return "", err
	}

	if deployment.ImageS3Key == "" {
		return "", fmt.Errorf("no artifact available for this deployment")
	}

	store, err := s3store.NewImageStore(config.AppConfig.S3)
	if err != nil {
		s.logger.Log(logger.Error, "failed to create S3 client", err.Error())
		return "", fmt.Errorf("failed to create S3 client: %w", err)
	}

	exists, err := store.ObjectExists(ctx, deployment.ImageS3Key)
	if err != nil {
		s.logger.Log(logger.Error, "failed to check artifact existence", err.Error())
		return "", fmt.Errorf("failed to check artifact existence: %w", err)
	}
	if !exists {
		return "", fmt.Errorf("no artifact available for this deployment")
	}

	url, err := store.PresignedDownloadURL(ctx, deployment.ImageS3Key, 15*time.Minute)
	if err != nil {
		s.logger.Log(logger.Error, "failed to generate download URL", err.Error())
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}
	return url, nil
}

func (s *DeployService) DeleteArtifact(ctx context.Context, deploymentID string, organizationID uuid.UUID) error {
	if !s3store.IsConfigured(config.AppConfig.S3) {
		return types.ErrS3NotConfigured
	}

	deployment, err := s.storage.GetApplicationDeploymentById(deploymentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("deployment not found")
		}
		s.logger.Log(logger.Error, "failed to get deployment for delete", err.Error())
		return err
	}

	_, err = s.storage.GetApplicationBasicById(deployment.ApplicationID.String(), organizationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "not found") {
			return types.ErrPermissionDenied
		}
		s.logger.Log(logger.Error, "failed to verify app ownership for delete", err.Error())
		return err
	}

	if deployment.ImageS3Key == "" {
		return fmt.Errorf("no artifact to delete for this deployment")
	}

	store, err := s3store.NewImageStore(config.AppConfig.S3)
	if err != nil {
		s.logger.Log(logger.Error, "failed to create S3 client", err.Error())
		return fmt.Errorf("failed to create S3 client: %w", err)
	}

	if err := store.DeleteImage(ctx, deployment.ImageS3Key); err != nil {
		s.logger.Log(logger.Error, "failed to delete artifact from S3", err.Error())
		return fmt.Errorf("failed to delete artifact from S3: %w", err)
	}

	if err := s.storage.ClearDeploymentArtifactFields(deployment.ID); err != nil {
		s.logger.Log(logger.Error, "failed to clear artifact metadata", err.Error())
		return fmt.Errorf("failed to clear artifact metadata: %w", err)
	}
	return nil
}
