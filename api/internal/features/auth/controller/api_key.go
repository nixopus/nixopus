package auth

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
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

	// Get organization ID from header
	organizationIDStr := c.Request().Header.Get("X-Organization-Id")
	if organizationIDStr == "" {
		return nil, fuego.HTTPError{
			Err:    shared_types.ErrFailedToGetOrganizationFromContext,
			Status: http.StatusBadRequest,
		}
	}

	organizationID, err := uuid.Parse(organizationIDStr)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
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
