package auth

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	deploy_service "github.com/raghavyuva/nixopus-api/internal/features/deploy/service"
	deploy_storage "github.com/raghavyuva/nixopus-api/internal/features/deploy/storage"
	deploy_types "github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// CLIInitRequest represents the request for CLI init
type CLIInitRequest struct {
	Name                 string            `json:"name"`
	Repository           string            `json:"repository"`
	Branch               string            `json:"branch,omitempty"`
	Domains              []string          `json:"domains,omitempty"`
	EnvironmentVariables map[string]string `json:"environment_variables,omitempty"`
}

// CLIInitResponse represents the response from CLI init
type CLIInitResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	ProjectID string `json:"project_id"`
	FamilyID  string `json:"family_id"`
	Domain    string `json:"domain,omitempty"`
}

// HandleCLIInit handles CLI init - creates a draft project
func (ar *AuthController) HandleCLIInit(c fuego.ContextWithBody[CLIInitRequest]) (*CLIInitResponse, error) {
	req, err := c.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	if req.Name == "" {
		return nil, fuego.HTTPError{
			Err:    fmt.Errorf("project name is required"),
			Status: http.StatusBadRequest,
		}
	}

	if req.Repository == "" {
		return nil, fuego.HTTPError{
			Err:    fmt.Errorf("repository is required"),
			Status: http.StatusBadRequest,
		}
	}

	// Get user from request context
	user := utils.GetUser(c.Response(), c.Request())
	if user == nil {
		ar.logger.Log(logger.Error, "user not found", "")
		return nil, fuego.HTTPError{
			Err:    fmt.Errorf("user not found"),
			Status: http.StatusUnauthorized,
		}
	}

	// Get organization ID from request context
	organizationID := utils.GetOrganizationID(c.Request())
	if organizationID == uuid.Nil {
		ar.logger.Log(logger.Error, "organization not found", "")
		return nil, fuego.HTTPError{
			Err:    fmt.Errorf("organization not found"),
			Status: http.StatusUnauthorized,
		}
	}

	// Create deploy service to use CreateProject function
	deployStorage := &deploy_storage.DeployStorage{DB: ar.store.DB, Ctx: ar.ctx}
	deployService := deploy_service.NewDeployService(ar.store, ar.ctx, ar.logger, deployStorage)

	// Create project request with defaults
	environment := shared_types.Development
	if req.Branch == "" {
		req.Branch = "main" // Default branch
	}

	createProjectReq := &deploy_types.CreateProjectRequest{
		Name:                 req.Name,
		Repository:           req.Repository,
		Branch:               req.Branch,
		Domains:              req.Domains,
		Environment:          environment,
		BuildPack:            shared_types.DockerFile, // Default build pack
		EnvironmentVariables: req.EnvironmentVariables,
		BasePath:             "/", // CLI init always creates app at repo root
	}

	// Create project using internal function
	application, err := deployService.CreateProject(createProjectReq, user.ID, organizationID)
	if err != nil {
		ar.logger.Log(logger.Error, fmt.Sprintf("Failed to create project: %v", err), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	// Determine domain: use first custom domain if available, otherwise generate default domain
	var domain string
	if len(application.Domains) > 0 && application.Domains[0] != nil {
		domain = application.Domains[0].Domain
	} else {
		// Generate default domain: {first-8-chars}.nixopus.com (without protocol)
		appIDStr := application.ID.String()
		if len(appIDStr) >= 8 {
			domain = fmt.Sprintf("%s.nixopus.com", appIDStr[:8])
		}
	}

	familyID := ""
	if application.FamilyID != nil {
		familyID = application.FamilyID.String()
	}

	return &CLIInitResponse{
		Status:    "success",
		Message:   "Project created successfully",
		ProjectID: application.ID.String(),
		FamilyID:  familyID,
		Domain:    domain,
	}, nil
}
