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

	// Collect domains from both old Domain field (for backward compatibility) and new Domains array
	domains := req.Domains
	if len(domains) == 0 && req.Domain != "" {
		domains = []string{req.Domain}
	}

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
		Domain:               req.Domain, // Keep for backward compatibility
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

	// Add domains to application_domains table
	if len(domains) > 0 {
		if err := s.storage.AddApplicationDomains(application.ID, domains); err != nil {
			s.logger.Log(logger.Error, "failed to add domains", err.Error())
			return shared_types.Application{}, err
		}
		// Load domains into application for response
		domainsList, _ := s.storage.GetApplicationDomains(application.ID)
		// Convert []ApplicationDomain to []*ApplicationDomain
		domainPtrs := make([]*shared_types.ApplicationDomain, len(domainsList))
		for i := range domainsList {
			domainPtrs[i] = &domainsList[i]
		}
		application.Domains = domainPtrs
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
