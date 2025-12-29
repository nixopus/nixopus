package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/tasks"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// CreateProject creates a new application without triggering deployment.
// The application is saved with a "draft" status and can be deployed later.
func (s *DeployService) CreateProject(req *types.CreateProjectRequest, userID uuid.UUID, organizationID uuid.UUID) (shared_types.Application, error) {
	s.logger.Log(logger.Info, "creating project without deployment", "name: "+req.Name)

	now := time.Now()
	application := shared_types.Application{
		ID:                   uuid.New(),
		Name:                 req.Name,
		BuildVariables:       tasks.GetStringFromMap(req.BuildVariables),
		EnvironmentVariables: tasks.GetStringFromMap(req.EnvironmentVariables),
		Environment:          req.Environment,
		BuildPack:            req.BuildPack,
		Repository:           req.Repository,
		Branch:               req.Branch,
		PreRunCommand:        req.PreRunCommand,
		PostRunCommand:       req.PostRunCommand,
		Port:                 req.Port,
		Domain:               req.Domain,
		UserID:               userID,
		CreatedAt:            now,
		UpdatedAt:            now,
		DockerfilePath:       req.DockerfilePath,
		BasePath:             req.BasePath,
		OrganizationID:       organizationID,
	}

	// Save the application to the database
	if err := s.storage.AddApplication(&application); err != nil {
		s.logger.Log(logger.Error, "failed to create application", err.Error())
		return shared_types.Application{}, err
	}

	// Create an application status with "draft" status
	appStatus := shared_types.ApplicationStatus{
		ID:            uuid.New(),
		ApplicationID: application.ID,
		Status:        shared_types.Draft,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.storage.AddApplicationStatus(&appStatus); err != nil {
		s.logger.Log(logger.Error, "failed to create application status", err.Error())
		return shared_types.Application{}, err
	}

	// Attach the status to the application for the response
	application.Status = &appStatus

	s.logger.Log(logger.Info, "project created successfully", "id: "+application.ID.String())
	return application, nil
}
