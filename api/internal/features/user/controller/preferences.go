package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/user/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// GetUserPreferences retrieves user preferences
func (c *UserController) GetUserPreferences(s fuego.ContextNoBody) (*types.UserPreferencesResponse, error) {
	w, r := s.Response(), s.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	prefs, err := c.service.GetUserPreferences(user.ID.String())
	if err != nil {
		c.logger.Log(logger.Error, "failed to get user preferences", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
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

	prefs, err := c.service.UpdateUserPreferences(user.ID.String(), req)
	if err != nil {
		c.logger.Log(logger.Error, "failed to update user preferences", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.UserPreferencesResponse{
		Status:  "success",
		Message: "User preferences updated successfully",
		Data:    prefs,
	}, nil
}
