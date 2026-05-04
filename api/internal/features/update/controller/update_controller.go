package controller

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/config"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/update/service"
	"github.com/nixopus/nixopus/api/internal/features/update/types"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

type UpdateController struct {
	service *service.UpdateService
	logger  *logger.Logger
}

func NewUpdateController(service *service.UpdateService, logger *logger.Logger) *UpdateController {
	return &UpdateController{
		service: service,
		logger:  logger,
	}
}

func (c *UpdateController) CheckForUpdates(s fuego.ContextNoBody) (*types.UpdateCheckResponse, error) {
	w, r := s.Response(), s.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		c.logUpdateDebug("CheckForUpdates", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	ctxStr := updateRequestData(r, user)

	// If the environment is development, return current version but skip remote check
	if config.AppConfig.App.Environment == "development" {
		currentVersion, err := c.service.GetCurrentVersion()
		if err != nil {
			// In development, log the error but don't fail the request
			c.logger.Log(logger.Warning, fmt.Sprintf("update: CheckForUpdates: get current version (dev): %v", err), ctxStr)
			currentVersion = "unknown"
		}
		return &types.UpdateCheckResponse{
			CurrentVersion:  currentVersion,
			LatestVersion:   currentVersion,
			UpdateAvailable: false,
			Environment:     "development",
		}, nil
	}

	response, err := c.service.CheckForUpdates()
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("update: CheckForUpdates: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Debug, "update: CheckForUpdates ok", fmt.Sprintf("%s update_available=%v current=%s latest=%s", ctxStr, response.UpdateAvailable, response.CurrentVersion, response.LatestVersion))

	// If update is available and user has auto update enabled, perform the update
	if response.UpdateAvailable {
		autoUpdate, err := c.service.GetUserAutoUpdatePreference(user.ID)
		if err != nil {
			c.logger.Log(logger.Error, fmt.Sprintf("update: CheckForUpdates: auto-update preference: %v", err), ctxStr)
			return response, nil
		}

		if autoUpdate {
			go func() {
				// Get organization ID from request context
				orgID := utils.GetOrganizationID(r)
				if orgID == uuid.Nil {
					c.logger.Log(logger.Error, "update: CheckForUpdates: auto-update skipped: organization ID required", ctxStr)
					return
				}
				orgCtx := r.Context()
				if orgCtx == nil {
					orgCtx = context.Background()
				}
				orgCtx = context.WithValue(orgCtx, shared_types.OrganizationIDKey, orgID.String())
				autoData := fmt.Sprintf("%s org_id=%s", ctxStr, orgID)
				if err := c.service.PerformUpdate(orgCtx); err != nil {
					c.logger.Log(logger.Error, fmt.Sprintf("update: CheckForUpdates: auto-update: %v", err), autoData)
				}
			}()
		}
	}

	return response, nil
}

func (c *UpdateController) PerformUpdate(s fuego.ContextWithBody[types.UpdateRequest]) (*types.UpdateResponse, error) {
	w, r := s.Response(), s.Request()
	user := utils.GetUser(w, r)

	// If the environment is development, we will not perform updates
	if config.AppConfig.App.Environment == "development" {
		c.logUpdateDebug("PerformUpdate", "skipped in development", updateRequestData(r, user))
		return &types.UpdateResponse{
			Success: true,
			Message: "Update completed successfully",
		}, nil
	}

	if user == nil {
		c.logUpdateDebug("PerformUpdate", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	ctxStr := updateRequestData(r, user)

	req, err := s.Body()
	if err != nil {
		c.logUpdateDebug("PerformUpdate", fmt.Sprintf("parse body: %v", err), ctxStr)
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	updateInfo, err := c.service.CheckForUpdates()
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("update: PerformUpdate: check: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	if !updateInfo.UpdateAvailable && !req.Force {
		c.logger.Log(logger.Info, "update: PerformUpdate: no update available", fmt.Sprintf("%s force=%v", ctxStr, req.Force))
		return &types.UpdateResponse{
			Success: false,
			Message: "No updates available",
		}, nil
	}

	// Get organization ID from request context
	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logUpdateDebug("PerformUpdate", "organization ID required", ctxStr)
		return nil, fuego.BadRequestError{Detail: "organization ID not found in context", Err: fmt.Errorf("organization ID not found in context")}
	}
	orgCtx := r.Context()
	orgCtx = context.WithValue(orgCtx, shared_types.OrganizationIDKey, orgID.String())
	execCtx := fmt.Sprintf("%s org_id=%s force=%v", ctxStr, orgID, req.Force)
	if err := c.service.PerformUpdate(orgCtx); err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("update: PerformUpdate: %v", err), execCtx)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "update: PerformUpdate completed", execCtx)

	return &types.UpdateResponse{
		Success: true,
		Message: "Update completed successfully",
	}, nil
}
