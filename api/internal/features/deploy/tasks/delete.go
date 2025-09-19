package tasks

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/image"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// DeleteDeployment deletes a deployment and its associated resources.
// It stops and removes the service, image, and repository.
// It returns an error if any operation fails.
func (s *TaskService) DeleteDeployment(deployment *types.DeleteDeploymentRequest, userID uuid.UUID, organizationID uuid.UUID) error {
	application, err := s.Storage.GetApplicationById(deployment.ID.String(), organizationID)
	if err != nil {
		return fmt.Errorf("failed to get application details: %w", err)
	}

	domain := application.Domain

	services, err := s.DockerRepo.GetClusterServices()
	if err != nil {
		s.Logger.Log(logger.Error, "Failed to get services", err.Error())
	} else {
		for _, service := range services {
			if service.Spec.Annotations.Name == application.Name {
				s.Logger.Log(logger.Info, "Deleting service", service.ID)
				if err := s.DockerRepo.DeleteService(service.ID); err != nil {
					s.Logger.Log(logger.Error, "Failed to delete service", err.Error())
				} else {
					s.Logger.Log(logger.Info, "Service deleted successfully", service.ID)
				}
				break
			}
		}
	}

	deployments, err := s.Storage.GetApplicationDeployments(application.ID)
	if err != nil {
		s.Logger.Log(logger.Error, "Failed to get application deployments", err.Error())
	} else {
		for _, dep := range deployments {
			if dep.ContainerImage != "" {
				s.Logger.Log(logger.Info, "Removing image", dep.ContainerImage)
				if err := s.DockerRepo.RemoveImage(dep.ContainerImage, image.RemoveOptions{Force: true}); err != nil {
					s.Logger.Log(logger.Error, "Failed to remove image", err.Error())
				}
			}
		}
	}

	repoPath := filepath.Join(os.Getenv("MOUNT_PATH"), userID.String(), string(application.Environment), application.ID.String())
	s.Logger.Log(logger.Info, "Cleaning up repository directory", repoPath)

	err = s.Github_service.RemoveRepository(repoPath)
	if err != nil {
		s.Logger.Log(logger.Error, "Failed to remove repository", err.Error())
	}

	client := GetCaddyClient()
	err = client.DeleteDomain(domain)
	if err != nil {
		s.Logger.Log(logger.Error, "Failed to remove domain", err.Error())
	}
	client.Reload()

	return s.Storage.DeleteDeployment(deployment, userID)
}
