package service

import (
	"strings"
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

	// For compose apps, extract domain names from ComposeDomains if Domains is empty.
	// Service linkage is deferred until compose services are discovered during deploy.
	domains := req.Domains
	if len(domains) == 0 && len(req.ComposeDomains) > 0 {
		for _, cd := range req.ComposeDomains {
			d := strings.TrimSpace(cd.Domain)
			if d != "" {
				domains = append(domains, d)
			}
		}
	}

	// Set BasePath default to "/" if empty (for CLI init, always root)
	basePath := req.BasePath
	if basePath == "" {
		basePath = "/"
	}

	// Create a new family_id for this application
	// This allows grouping multiple apps (monorepo) or environments (duplicates)
	familyID := uuid.New()

	source := req.Source
	if source == "" {
		source = shared_types.SourceGithub
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
		UserID:               userID,
		CreatedAt:            now,
		UpdatedAt:            now,
		DockerfilePath:       req.DockerfilePath,
		BasePath:             basePath,
		OrganizationID:       organizationID,
		FamilyID:             &familyID,
		Source:               source,
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

	if len(req.ComposeServices) > 0 {
		if err := s.persistComposeServicesAndLinkDomains(application.ID, req.ComposeServices, req.ComposeDomains); err != nil {
			s.logger.Log(logger.Error, "failed to persist compose services", err.Error())
		}
	}

	domainsList, err := s.storage.GetApplicationDomains(application.ID)
	if err != nil {
		s.logger.Log(logger.Error, "failed to load domains after creation", err.Error())
		return shared_types.Application{}, err
	}
	domainPtrs := make([]*shared_types.ApplicationDomain, len(domainsList))
	for i := range domainsList {
		domainPtrs[i] = &domainsList[i]
	}
	application.Domains = domainPtrs
	application.Status = &appStatus

	s.logger.Log(logger.Info, "project created successfully", "id: "+application.ID.String())
	return application, nil
}

func (s *DeployService) persistComposeServicesAndLinkDomains(appID uuid.UUID, previewServices []types.PreviewComposeService, composeDomains []types.ComposeDomain) error {
	var services []shared_types.ComposeService
	for _, ps := range previewServices {
		services = append(services, shared_types.ComposeService{
			ServiceName: ps.ServiceName,
			Port:        ps.Port,
		})
	}
	if err := s.storage.UpsertComposeServices(appID, services); err != nil {
		return err
	}

	if len(composeDomains) == 0 {
		return nil
	}

	persisted, err := s.storage.GetComposeServices(appID)
	if err != nil {
		return err
	}
	serviceByName := make(map[string]*shared_types.ComposeService, len(persisted))
	for i := range persisted {
		serviceByName[persisted[i].ServiceName] = &persisted[i]
	}

	for _, cd := range composeDomains {
		domain := strings.TrimSpace(cd.Domain)
		if domain == "" || cd.ServiceName == "" {
			continue
		}
		svc, ok := serviceByName[cd.ServiceName]
		if !ok {
			continue
		}
		port := cd.Port
		if port == 0 {
			port = svc.Port
		}
		if err := s.storage.UpdateApplicationDomainService(appID, domain, &svc.ID, &port); err != nil {
			s.logger.Log(logger.Warning, "failed to link domain "+domain+" to service "+cd.ServiceName, err.Error())
		}
	}

	return nil
}
