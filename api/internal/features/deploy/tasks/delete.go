package tasks

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/image"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// DeleteDeployment deletes a deployment and its associated resources.
// It stops and removes the service, image, and repository.
// It returns an error if any operation fails.
func (s *TaskService) DeleteDeployment(deployment *types.DeleteDeploymentRequest, userID uuid.UUID, organizationID uuid.UUID) error {
	application, err := s.Storage.GetApplicationById(deployment.ID.String(), organizationID)
	if err != nil {
		return fmt.Errorf("failed to get application details: %w", err)
	}

	// Load domains if not already loaded
	if application.Domains == nil || len(application.Domains) == 0 {
		domainsList, err := s.Storage.GetApplicationDomains(application.ID)
		if err == nil {
			domainPtrs := make([]*shared_types.ApplicationDomain, len(domainsList))
			for i := range domainsList {
				domainPtrs[i] = &domainsList[i]
			}
			application.Domains = domainPtrs
		}
	}

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

	// Remove all domains from Caddy
	if len(application.Domains) > 0 {
		client := GetCaddyClient()
		if client == nil {
			s.Logger.Log(logger.Warning, "Caddy client not configured", "")
		} else {
			for _, appDomain := range application.Domains {
				if appDomain.Domain != "" {
					err = client.DeleteDomain(appDomain.Domain)
					if err != nil {
						s.Logger.Log(logger.Error, "Failed to remove domain", err.Error())
					}
				}
			}
			client.Reload()
		}
	}

	// Handle family cleanup: if this project belongs to a family,
	// check if only one member remains and clear its family_id
	if application.FamilyID != nil {
		s.Logger.Log(logger.Info, "Checking family cleanup", application.FamilyID.String())
		if err := s.Storage.ClearFamilyIDIfSingleMember(*application.FamilyID); err != nil {
			s.Logger.Log(logger.Error, "Failed to cleanup family", err.Error())
		}
	}

	return s.Storage.DeleteDeployment(deployment, userID)
}
