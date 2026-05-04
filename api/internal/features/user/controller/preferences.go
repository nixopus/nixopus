package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/user/types"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

// GetUserPreferences retrieves user preferences
func (c *UserController) GetUserPreferences(s fuego.ContextNoBody) (*types.UserPreferencesResponse, error) {
	w, r := s.Response(), s.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		c.logUserDebug("GetUserPreferences", "authentication required", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	ctxStr := userRequestData(r, user)

	prefs, err := c.service.GetUserPreferences(user.ID.String())
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("user: GetUserPreferences: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.UserPreferencesResponse{
		Status:  "success",
		Message: "User preferences fetched successfully",
		Data:    prefs,
	}, nil
}

// UpdateUserPreferences updates user preferences with the provided data
func (c *UserController) UpdateUserPreferences(s fuego.ContextWithBody[shared_types.UserPreferencesData]) (*types.UserPreferencesResponse, error) {
	w, r := s.Response(), s.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		c.logUserDebug("UpdateUserPreferences", "authentication required", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	ctxStr := userRequestData(r, user)

	req, err := s.Body()
	if err != nil {
		c.logUserDebug("UpdateUserPreferences", fmt.Sprintf("parse body: %v", err), ctxStr)
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	prefs, err := c.service.UpdateUserPreferences(user.ID.String(), req)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("user: UpdateUserPreferences: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.UserPreferencesResponse{
		Status:  "success",
		Message: "User preferences updated successfully",
		Data:    prefs,
	}, nil
}
