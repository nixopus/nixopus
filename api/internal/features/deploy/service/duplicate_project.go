package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// DuplicateProject creates a duplicate of an existing project with a different environment.
// It copies all configurations from the source project and creates a new project in draft status.
// Both projects are linked via a shared family_id.
func (s *DeployService) DuplicateProject(req *types.DuplicateProjectRequest, userID uuid.UUID, organizationID uuid.UUID) (shared_types.Application, error) {
	s.logger.Log(logger.Info, "duplicating project", "source_id: "+req.SourceProjectID.String())

	// Get the source project
	sourceProject, err := s.storage.GetApplicationById(req.SourceProjectID.String(), organizationID)
	if err != nil {
		s.logger.Log(logger.Error, "failed to get source project", err.Error())
		return shared_types.Application{}, types.ErrApplicationNotFound
	}

	// Check if trying to duplicate with the same environment
	if sourceProject.Environment == req.Environment {
		s.logger.Log(logger.Error, "cannot duplicate with same environment", "")
		return shared_types.Application{}, types.ErrSameEnvironmentAsDuplicate
	}

	// Determine the family_id
	var familyID uuid.UUID
	if sourceProject.FamilyID != nil {
		familyID = *sourceProject.FamilyID

		// Check if the environment already exists in the family
		exists, err := s.storage.IsEnvironmentInFamily(familyID, req.Environment)
		if err != nil {
			s.logger.Log(logger.Error, "failed to check environment in family", err.Error())
			return shared_types.Application{}, err
		}
		if exists {
			s.logger.Log(logger.Error, "environment already exists in family", "")
			return shared_types.Application{}, types.ErrEnvironmentAlreadyExistsInFamily
		}
	} else {
		// Create a new family_id for both projects
		familyID = uuid.New()

		// Update the source project with the new family_id
		if err := s.storage.UpdateApplicationFamilyID(sourceProject.ID, &familyID); err != nil {
			s.logger.Log(logger.Error, "failed to update source project family_id", err.Error())
			return shared_types.Application{}, err
		}
	}

	// Generate auto name based on source name and new environment
	newName := generateDuplicateName(sourceProject.Name, string(req.Environment))

	// Use provided branch if available, otherwise use source branch
	branch := sourceProject.Branch
	if strings.TrimSpace(req.Branch) != "" {
		branch = strings.TrimSpace(req.Branch)
	}

	now := time.Now()
	newProject := shared_types.Application{
		ID:                   uuid.New(),
		Name:                 newName,
		BuildVariables:       sourceProject.BuildVariables,
		EnvironmentVariables: sourceProject.EnvironmentVariables,
		Environment:          req.Environment,
		BuildPack:            sourceProject.BuildPack,
		Repository:           sourceProject.Repository,
		Branch:               branch,
		PreRunCommand:        sourceProject.PreRunCommand,
		PostRunCommand:       sourceProject.PostRunCommand,
		Port:                 sourceProject.Port,
		UserID:               userID,
		CreatedAt:            now,
		UpdatedAt:            now,
		DockerfilePath:       sourceProject.DockerfilePath,
		BasePath:             sourceProject.BasePath,
		OrganizationID:       organizationID,
		FamilyID:             &familyID,
		ProxyServer:          sourceProject.ProxyServer,
		Labels:               sourceProject.Labels,
	}

	// Save the new project
	if err := s.storage.AddApplication(&newProject); err != nil {
		s.logger.Log(logger.Error, "failed to create duplicate project", err.Error())
		return shared_types.Application{}, err
	}

	// Create application status with draft status
	appStatus := shared_types.ApplicationStatus{
		ID:            uuid.New(),
		ApplicationID: newProject.ID,
		Status:        shared_types.Draft,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.storage.AddApplicationStatus(&appStatus); err != nil {
		s.logger.Log(logger.Error, "failed to create application status", err.Error())
		return shared_types.Application{}, err
	}

	newProject.Status = &appStatus

	// Handle domains: require explicit domains when duplicating across environments
	domains := req.Domains
	if len(domains) == 0 {
		// Load domains from source project
		sourceDomains, err := s.storage.GetApplicationDomains(sourceProject.ID)
		if err != nil {
			s.logger.Log(logger.Error, "failed to load source domains", err.Error())
			return shared_types.Application{}, err
		}
		// When duplicating across environments, require explicit domains to avoid routing conflicts
		if sourceProject.Environment != req.Environment {
			s.logger.Log(logger.Error, "domains required when duplicating across environments", "")
			return shared_types.Application{}, types.ErrMissingDomain
		}
		// Same environment: copy domains from source
		for _, d := range sourceDomains {
			domains = append(domains, d.Domain)
		}
	}

	// Add domains to new project
	if len(domains) > 0 {
		if err := s.storage.AddApplicationDomains(newProject.ID, domains); err != nil {
			s.logger.Log(logger.Error, "failed to add domains to duplicate project", err.Error())
			return shared_types.Application{}, err
		}
		// Load domains into newProject for response
		domainsList, err := s.storage.GetApplicationDomains(newProject.ID)
		if err != nil {
			s.logger.Log(logger.Error, "failed to load domains after duplication", err.Error())
			return shared_types.Application{}, err
		}
		domainPtrs := make([]*shared_types.ApplicationDomain, len(domainsList))
		for i := range domainsList {
			domainPtrs[i] = &domainsList[i]
		}
		newProject.Domains = domainPtrs
	}

	s.logger.Log(logger.Info, "project duplicated successfully", "new_id: "+newProject.ID.String())
	return newProject, nil
}

// GetProjectFamily retrieves all projects that belong to a family.
func (s *DeployService) GetProjectFamily(familyID uuid.UUID, organizationID uuid.UUID) ([]shared_types.Application, error) {
	s.logger.Log(logger.Info, "getting project family", "family_id: "+familyID.String())

	projects, err := s.storage.GetProjectsByFamilyID(familyID, organizationID)
	if err != nil {
		s.logger.Log(logger.Error, "failed to get project family", err.Error())
		return nil, err
	}

	if len(projects) == 0 {
		return nil, types.ErrProjectFamilyNotFound
	}

	return projects, nil
}

// GetEnvironmentsInFamily retrieves all environments that exist in a project family.
func (s *DeployService) GetEnvironmentsInFamily(familyID uuid.UUID, organizationID uuid.UUID) ([]shared_types.Environment, error) {
	s.logger.Log(logger.Info, "getting environments in family", "family_id: "+familyID.String())

	environments, err := s.storage.GetEnvironmentsInFamily(familyID, organizationID)
	if err != nil {
		s.logger.Log(logger.Error, "failed to get environments in family", err.Error())
		return nil, err
	}

	return environments, nil
}

// generateDuplicateName creates a name for the duplicate project.
// It extracts the base name (removing any existing environment suffix) and appends the new environment.
func generateDuplicateName(sourceName string, newEnvironment string) string {
	// List of known environment suffixes to remove
	envSuffixes := []string{"-development", "-staging", "-production", "-dev", "-stage", "-prod"}

	baseName := sourceName
	for _, suffix := range envSuffixes {
		if strings.HasSuffix(strings.ToLower(baseName), suffix) {
			baseName = baseName[:len(baseName)-len(suffix)]
			break
		}
	}

	// Shorten environment name for the suffix
	envSuffix := newEnvironment
	switch newEnvironment {
	case "development":
		envSuffix = "dev"
	case "staging":
		envSuffix = "staging"
	case "production":
		envSuffix = "prod"
	}

	return fmt.Sprintf("%s-%s", baseName, envSuffix)
}
