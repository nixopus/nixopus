package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/update/service"
	"github.com/raghavyuva/nixopus-api/internal/features/update/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
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
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	// If the environment is development, return current version but skip remote check
	if config.AppConfig.App.Environment == "development" {
		currentVersion, err := c.service.GetCurrentVersion()
		if err != nil {
			// In development, log the error but don't fail the request
			c.logger.Log(logger.Warning, "Failed to get current version in development", err.Error())
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
		c.logger.Log(logger.Error, "failed to check for updates", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	// If update is available and user has auto update enabled, perform the update
	if response.UpdateAvailable {
		autoUpdate, err := c.service.GetUserAutoUpdatePreference(user.ID)
		if err != nil {
			c.logger.Log(logger.Error, "failed to get user auto-update preference", err.Error())
			return response, nil
		}

		if autoUpdate {
			go func() {
				if err := c.service.PerformUpdate(); err != nil {
					c.logger.Log(logger.Error, "failed to perform auto-update", err.Error())
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
		return &types.UpdateResponse{
			Success: true,
			Message: "Update completed successfully",
		}, nil
	}

	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	req, err := s.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	updateInfo, err := c.service.CheckForUpdates()
	if err != nil {
		c.logger.Log(logger.Error, "failed to check for updates", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	if !updateInfo.UpdateAvailable && !req.Force {
		return &types.UpdateResponse{
			Success: false,
			Message: "No updates available",
		}, nil
	}

	if err := c.service.PerformUpdate(); err != nil {
		c.logger.Log(logger.Error, "failed to perform update", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.UpdateResponse{
		Success: true,
		Message: "Update completed successfully",
	}, nil
}
