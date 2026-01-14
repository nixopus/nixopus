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

	domains := req.Domains

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
		UserID:               userID,
		CreatedAt:            now,
		UpdatedAt:            now,
		DockerfilePath:       req.DockerfilePath,
		BasePath:             req.BasePath,
		OrganizationID:       organizationID,
	}

	// Begin transaction for atomicity
	tx, err := s.store.DB.BeginTx(s.Ctx, nil)
	if err != nil {
		s.logger.Log(logger.Error, "failed to begin transaction", err.Error())
		return shared_types.Application{}, err
	}
	defer tx.Rollback()

	// Save the application to the database
	if _, err := tx.NewInsert().Model(&application).Exec(s.Ctx); err != nil {
		s.logger.Log(logger.Error, "failed to create application", err.Error())
		return shared_types.Application{}, err
	}

	// Add domains to application_domains table within transaction
	if len(domains) > 0 {
		// Use transaction for domain operations
		for _, domain := range domains {
			if domain == "" {
				continue
			}
			appDomain := &shared_types.ApplicationDomain{
				ID:            uuid.New(),
				ApplicationID: application.ID,
				Domain:        domain,
				CreatedAt:     now,
			}
			if _, err := tx.NewInsert().Model(appDomain).Exec(s.Ctx); err != nil {
				s.logger.Log(logger.Error, "failed to add domain", err.Error())
				return shared_types.Application{}, err
			}
		}
	}

	// Create an application status with "draft" status
	appStatus := shared_types.ApplicationStatus{
		ID:            uuid.New(),
		ApplicationID: application.ID,
		Status:        shared_types.Draft,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if _, err := tx.NewInsert().Model(&appStatus).Exec(s.Ctx); err != nil {
		s.logger.Log(logger.Error, "failed to create application status", err.Error())
		return shared_types.Application{}, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		s.logger.Log(logger.Error, "failed to commit transaction", err.Error())
		return shared_types.Application{}, err
	}

	// Load domains into application for response (after successful commit)
	if len(domains) > 0 {
		domainsList, err := s.storage.GetApplicationDomains(application.ID)
		if err != nil {
			s.logger.Log(logger.Error, "failed to load domains after creation", err.Error())
			return shared_types.Application{}, err
		}
		// Convert []ApplicationDomain to []*ApplicationDomain
		domainPtrs := make([]*shared_types.ApplicationDomain, len(domainsList))
		for i := range domainsList {
			domainPtrs[i] = &domainsList[i]
		}
		application.Domains = domainPtrs
	}

	// Attach the status to the application for the response
	application.Status = &appStatus

	s.logger.Log(logger.Info, "project created successfully", "id: "+application.ID.String())
	return application, nil
}
