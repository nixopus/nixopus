package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/user/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

type UpdateFontRequest struct {
	FontFamily string `json:"font_family" validate:"required" description:"Font family name for the editor" example:"JetBrains Mono"`
	FontSize   int    `json:"font_size" validate:"required,min=8,max=32" description:"Font size in pixels" example:"14"`
}

func (c *UserController) UpdateFont(s fuego.ContextWithBody[UpdateFontRequest]) (*types.UserSettingsResponse, error) {
	w, r := s.Response(), s.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		c.logUserDebug("UpdateFont", "authentication required", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	ctxStr := userRequestData(r, user)

	req, err := s.Body()
	if err != nil {
		c.logUserDebug("UpdateFont", fmt.Sprintf("parse body: %v", err), ctxStr)
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	settings, err := c.service.UpdateFont(user.ID.String(), req.FontFamily, req.FontSize)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("user: UpdateFont: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.UserSettingsResponse{
		Status:  "success",
		Message: "Font settings updated successfully",
		Data:    settings,
	}, nil
}

type UpdateThemeRequest struct {
	Theme string `json:"theme" validate:"required" description:"UI theme name" example:"dark"`
}

func (c *UserController) UpdateTheme(s fuego.ContextWithBody[UpdateThemeRequest]) (*types.UserSettingsResponse, error) {
	w, r := s.Response(), s.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		c.logUserDebug("UpdateTheme", "authentication required", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	ctxStr := userRequestData(r, user)

	req, err := s.Body()
	if err != nil {
		c.logUserDebug("UpdateTheme", fmt.Sprintf("parse body: %v", err), ctxStr)
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	settings, err := c.service.UpdateTheme(user.ID.String(), req.Theme)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("user: UpdateTheme: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.UserSettingsResponse{
		Status:  "success",
		Message: "Theme updated successfully",
		Data:    settings,
	}, nil
}

type UpdateLanguageRequest struct {
	Language string `json:"language" validate:"required" description:"Preferred language code" example:"en"`
}

func (c *UserController) UpdateLanguage(s fuego.ContextWithBody[UpdateLanguageRequest]) (*types.UserSettingsResponse, error) {
	w, r := s.Response(), s.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		c.logUserDebug("UpdateLanguage", "authentication required", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	ctxStr := userRequestData(r, user)

	req, err := s.Body()
	if err != nil {
		c.logUserDebug("UpdateLanguage", fmt.Sprintf("parse body: %v", err), ctxStr)
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	settings, err := c.service.UpdateLanguage(user.ID.String(), req.Language)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("user: UpdateLanguage: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.UserSettingsResponse{
		Status:  "success",
		Message: "Language updated successfully",
		Data:    settings,
	}, nil
}

type UpdateAutoUpdateRequest struct {
	AutoUpdate bool `json:"auto_update" description:"Whether to automatically apply updates" example:"true"`
}

func (c *UserController) UpdateAutoUpdate(s fuego.ContextWithBody[UpdateAutoUpdateRequest]) (*types.UserSettingsResponse, error) {
	w, r := s.Response(), s.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		c.logUserDebug("UpdateAutoUpdate", "authentication required", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	ctxStr := userRequestData(r, user)

	req, err := s.Body()
	if err != nil {
		c.logUserDebug("UpdateAutoUpdate", fmt.Sprintf("parse body: %v", err), ctxStr)
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	settings, err := c.service.UpdateAutoUpdate(user.ID.String(), req.AutoUpdate)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("user: UpdateAutoUpdate: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.UserSettingsResponse{
		Status:  "success",
		Message: "Auto update setting updated successfully",
		Data:    settings,
	}, nil
}

func (c *UserController) GetSettings(s fuego.ContextNoBody) (*types.UserSettingsResponse, error) {
	w, r := s.Response(), s.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		c.logUserDebug("GetSettings", "authentication required", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	ctxStr := userRequestData(r, user)

	settings, err := c.service.GetSettings(user.ID.String())
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("user: GetSettings: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.UserSettingsResponse{
		Status:  "success",
		Message: "User settings fetched successfully",
		Data:    settings,
	}, nil
}
