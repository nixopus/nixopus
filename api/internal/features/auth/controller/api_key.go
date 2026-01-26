package auth

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	deploy_service "github.com/raghavyuva/nixopus-api/internal/features/deploy/service"
	deploy_storage "github.com/raghavyuva/nixopus-api/internal/features/deploy/storage"
	deploy_types "github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
	"github.com/uptrace/bun"
)

// CreateAPIKeyResponse represents the response when creating an API key
type CreateAPIKeyResponse struct {
	Status  string           `json:"status"`
	Message string           `json:"message"`
	Data    CreateAPIKeyData `json:"data"`
}

type CreateAPIKeyData struct {
	APIKey *shared_types.APIKey `json:"api_key"`
	Key    string               `json:"key"` // Only returned once
}

// ListAPIKeysResponse represents the response when listing API keys
type ListAPIKeysResponse struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Data    []*shared_types.APIKey `json:"data"`
}

// MessageResponse represents a generic message response
type MessageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Member represents the member table from Better Auth schema
type Member struct {
	bun.BaseModel  `bun:"table:member,alias:m"`
	ID             uuid.UUID `bun:"id,pk,type:uuid"`
	OrganizationID uuid.UUID `bun:"organization_id,type:uuid,notnull"`
	UserID         uuid.UUID `bun:"user_id,type:uuid,notnull"`
	Role           string    `bun:"role,type:text,notnull,default:'member'"`
	CreatedAt      time.Time `bun:"created_at,type:timestamp,notnull"`
}

// getOrganizationIDWithFallback gets organization ID from multiple sources with fallbacks:
// 1. Context (set by middleware)
// 2. Header (X-Organization-Id)
// 3. Better Auth session
// 4. User's first organization from member table
func (ar *AuthController) getOrganizationIDWithFallback(ctx context.Context, r *http.Request, userID uuid.UUID) (uuid.UUID, error) {
	// 1. Try to get from context (set by middleware for non-auth endpoints)
	if orgID := utils.GetOrganizationID(r); orgID != uuid.Nil {
		return orgID, nil
	}

	// 2. Try to get from header
	organizationIDStr := r.Header.Get("X-Organization-Id")
	if organizationIDStr != "" {
		orgID, err := uuid.Parse(organizationIDStr)
		if err == nil {
			return orgID, nil
		}
	}

	// 3. Try to get from Better Auth session
	orgIDStr, err := utils.GetOrganizationIDFromBetterAuth(r)
	if err == nil && orgIDStr != "" {
		orgID, err := uuid.Parse(orgIDStr)
		if err == nil {
			return orgID, nil
		}
	}

	// 4. Fallback: Get user's first organization from member table
	var member Member
	err = ar.store.DB.NewSelect().
		Model(&member).
		Where("user_id = ?", userID).
		Order("created_at ASC").
		Limit(1).
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, fmt.Errorf("user has no organizations")
		}
		return uuid.Nil, fmt.Errorf("failed to query member table: %w", err)
	}

	return member.OrganizationID, nil
}

// CreateAPIKey creates a new API key for the authenticated user
func (ar *AuthController) CreateAPIKey(c fuego.ContextWithBody[types.CreateAPIKeyRequest]) (*CreateAPIKeyResponse, error) {
	req, err := c.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	if err := ar.validator.ValidateRequest(&req); err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	// Get user from context
	user := utils.GetUser(c.Response(), c.Request())
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    shared_types.ErrFailedToGetUserFromContext,
			Status: http.StatusUnauthorized,
		}
	}

	// Get organization ID with fallbacks (context, header, session, or first org)
	organizationID, err := ar.getOrganizationIDWithFallback(c.Request().Context(), c.Request(), user.ID)
	if err != nil {
		ar.logger.Log(logger.Warning, fmt.Sprintf("Failed to get organization ID for user %s: %v", user.ID, err), "")
		return nil, fuego.HTTPError{
			Err:    fmt.Errorf("failed to determine organization: %w", err),
			Status: http.StatusBadRequest,
		}
	}

	key, apiKey, err := ar.apiKeyService.GenerateAPIKey(user.ID, organizationID, req.Name, req.ExpiresInDays)
	if err != nil {
		ar.logger.Log(logger.Error, fmt.Sprintf("Failed to create API key: %v", err), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &CreateAPIKeyResponse{
		Status:  "success",
		Message: "API key created successfully",
		Data: CreateAPIKeyData{
			APIKey: apiKey,
			Key:    key,
		},
	}, nil
}

// ListAPIKeys lists all API keys for the authenticated user
func (ar *AuthController) ListAPIKeys(c fuego.ContextNoBody) (*ListAPIKeysResponse, error) {
	// Get user from context
	user := utils.GetUser(c.Response(), c.Request())
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    shared_types.ErrFailedToGetUserFromContext,
			Status: http.StatusUnauthorized,
		}
	}

	apiKeys, err := ar.apiKeyService.ListAPIKeys(user.ID)
	if err != nil {
		ar.logger.Log(logger.Error, fmt.Sprintf("Failed to list API keys: %v", err), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &ListAPIKeysResponse{
		Status:  "success",
		Message: "API keys retrieved successfully",
		Data:    apiKeys,
	}, nil
}

// RevokeAPIKey revokes an API key
func (ar *AuthController) RevokeAPIKey(c fuego.ContextNoBody) (*MessageResponse, error) {
	keyIDStr := c.PathParam("id")
	if keyIDStr == "" {
		return nil, fuego.HTTPError{
			Err:    fmt.Errorf("API key ID is required"),
			Status: http.StatusBadRequest,
		}
	}

	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	// Get user from context
	user := utils.GetUser(c.Response(), c.Request())
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    shared_types.ErrFailedToGetUserFromContext,
			Status: http.StatusUnauthorized,
		}
	}

	// Verify the API key belongs to the user
	apiKeys, err := ar.apiKeyService.ListAPIKeys(user.ID)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	found := false
	for _, ak := range apiKeys {
		if ak.ID == keyID {
			found = true
			break
		}
	}

	if !found {
		return nil, fuego.HTTPError{
			Err:    fmt.Errorf("API key not found"),
			Status: http.StatusNotFound,
		}
	}

	if err := ar.apiKeyService.RevokeAPIKey(keyID); err != nil {
		ar.logger.Log(logger.Error, fmt.Sprintf("Failed to revoke API key: %v", err), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &MessageResponse{
		Status:  "success",
		Message: "API key revoked successfully",
	}, nil
}

// ValidateAPIKeyRequest represents the request to validate an API key
type ValidateAPIKeyRequest struct {
	APIKey string `json:"api_key"`
}

// ValidateAPIKeyResponse represents the response from validating an API key
type ValidateAPIKeyResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Valid   bool   `json:"valid"`
}

// ValidateAPIKey validates an API key without requiring authentication
// This is a public endpoint used by CLI tools to validate API keys before initialization
func (ar *AuthController) ValidateAPIKey(c fuego.ContextWithBody[ValidateAPIKeyRequest]) (*ValidateAPIKeyResponse, error) {
	req, err := c.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	if req.APIKey == "" {
		return nil, fuego.HTTPError{
			Err:    fmt.Errorf("API key is required"),
			Status: http.StatusBadRequest,
		}
	}

	// Verify the API key
	apiKey, err := ar.apiKeyService.VerifyAPIKey(req.APIKey)
	if err != nil {
		ar.logger.Log(logger.Info, fmt.Sprintf("API key validation failed: %v", err), "")
		return &ValidateAPIKeyResponse{
			Status:  "error",
			Message: "Invalid API key",
			Valid:   false,
		}, nil
	}

	// Check if key is valid (not revoked or expired)
	if !apiKey.IsValid() {
		return &ValidateAPIKeyResponse{
			Status:  "error",
			Message: "API key is revoked or expired",
			Valid:   false,
		}, nil
	}

	return &ValidateAPIKeyResponse{
		Status:  "success",
		Message: "API key is valid",
		Valid:   true,
	}, nil
}

// CLIInitRequest represents the request for CLI init
type CLIInitRequest struct {
	APIKey               string            `json:"api_key"`
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

// HandleCLIInit handles CLI init - validates API key and creates a draft project
func (ar *AuthController) HandleCLIInit(c fuego.ContextWithBody[CLIInitRequest]) (*CLIInitResponse, error) {
	req, err := c.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	if req.APIKey == "" {
		return nil, fuego.HTTPError{
			Err:    fmt.Errorf("API key is required"),
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

	// Verify the API key to get user and organization info
	apiKey, err := ar.apiKeyService.VerifyAPIKey(req.APIKey)
	if err != nil {
		ar.logger.Log(logger.Info, fmt.Sprintf("API key validation failed: %v", err), "")
		return nil, fuego.HTTPError{
			Err:    fmt.Errorf("invalid API key"),
			Status: http.StatusUnauthorized,
		}
	}

	// Check if key is valid (not revoked or expired)
	if !apiKey.IsValid() {
		return nil, fuego.HTTPError{
			Err:    fmt.Errorf("API key is revoked or expired"),
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
	application, err := deployService.CreateProject(createProjectReq, apiKey.UserID, apiKey.OrganizationID)
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
