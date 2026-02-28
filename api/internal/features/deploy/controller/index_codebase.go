package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/live"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *DeployController) IndexCodebase(f fuego.ContextNoBody) (*types.IndexCodebaseResponse, error) {
	w, r := f.Response(), f.Request()

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.HTTPError{Status: http.StatusUnauthorized}
	}

	organizationID := utils.GetOrganizationID(r)
	if organizationID == uuid.Nil {
		return nil, fuego.HTTPError{
			Err:    errors.New("organization ID is required"),
			Status: http.StatusBadRequest,
		}
	}

	applicationID := f.QueryParam("application_id")
	if applicationID == "" {
		return nil, fuego.HTTPError{
			Err:    errors.New("application_id is required"),
			Status: http.StatusBadRequest,
		}
	}

	appID, err := uuid.Parse(applicationID)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    fmt.Errorf("invalid application_id: %w", err),
			Status: http.StatusBadRequest,
		}
	}

	application, err := c.storage.GetApplicationById(applicationID, organizationID)
	if err != nil {
		c.logger.Log(logger.Error, "index: application not found", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusNotFound,
		}
	}

	stagingPath, err := c.resolveStagingPath(r.Context(), appID, user.ID, organizationID, application.BasePath, string(application.Environment))
	if err != nil {
		c.logger.Log(logger.Error, "index: failed to resolve staging path", err.Error())
		return nil, fuego.HTTPError{
			Err:    fmt.Errorf("failed to resolve staging path: %w", err),
			Status: http.StatusInternalServerError,
		}
	}

	orgCtx := context.WithValue(r.Context(), shared_types.OrganizationIDKey, organizationID.String())
	result, err := live.IndexFromStaging(orgCtx, c.store, stagingPath, appID, organizationID)
	if err != nil {
		c.logger.Log(logger.Error, "index: indexing failed", err.Error())
		return nil, fuego.HTTPError{
			Err:    fmt.Errorf("indexing failed: %w", err),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.IndexCodebaseResponse{
		Status:  "success",
		Message: "Codebase indexed successfully",
		Data:    result,
	}, nil
}

func (c *DeployController) resolveStagingPath(ctx context.Context, applicationID, userID, organizationID uuid.UUID, basePath, environment string) (string, error) {
	orgCtx := context.WithValue(ctx, shared_types.OrganizationIDKey, organizationID.String())

	stagingPath, _, err := c.githubService.GetClonePath(orgCtx, userID.String(), environment, applicationID.String())
	if err != nil {
		if config.AppConfig.App.Environment == "development" || config.AppConfig.App.Environment == "dev" {
			localPath := filepath.Join(os.TempDir(), "nixopus-staging", userID.String(), environment, applicationID.String())
			if basePath != "" && basePath != "/" {
				localPath = filepath.Join(localPath, basePath)
			}
			if mkErr := os.MkdirAll(localPath, 0755); mkErr != nil {
				return "", mkErr
			}
			return localPath, nil
		}
		return "", err
	}

	if basePath != "" && basePath != "/" {
		stagingPath = filepath.Join(stagingPath, basePath)
	}

	return stagingPath, nil
}
