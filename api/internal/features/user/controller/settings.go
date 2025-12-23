package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

type UpdateFontRequest struct {
	FontFamily string `json:"font_family"`
	FontSize   int    `json:"font_size"`
}

func (c *UserController) UpdateFont(s fuego.ContextWithBody[UpdateFontRequest]) (*types.Response, error) {
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

	settings, err := c.service.UpdateFont(user.ID.String(), req.FontFamily, req.FontSize)
	if err != nil {
		c.logger.Log(logger.Error, "failed to update font settings", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.Response{
		Status:  "success",
		Message: "Font settings updated successfully",
		Data:    settings,
	}, nil
}

type UpdateThemeRequest struct {
	Theme string `json:"theme"`
}

func (c *UserController) UpdateTheme(s fuego.ContextWithBody[UpdateThemeRequest]) (*types.Response, error) {
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

	settings, err := c.service.UpdateTheme(user.ID.String(), req.Theme)
	if err != nil {
		c.logger.Log(logger.Error, "failed to update theme", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.Response{
		Status:  "success",
		Message: "Theme updated successfully",
		Data:    settings,
	}, nil
}

type UpdateLanguageRequest struct {
	Language string `json:"language"`
}

func (c *UserController) UpdateLanguage(s fuego.ContextWithBody[UpdateLanguageRequest]) (*types.Response, error) {
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

	settings, err := c.service.UpdateLanguage(user.ID.String(), req.Language)
	if err != nil {
		c.logger.Log(logger.Error, "failed to update language", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.Response{
		Status:  "success",
		Message: "Language updated successfully",
		Data:    settings,
	}, nil
}

type UpdateAutoUpdateRequest struct {
	AutoUpdate bool `json:"auto_update"`
}

func (c *UserController) UpdateAutoUpdate(s fuego.ContextWithBody[UpdateAutoUpdateRequest]) (*types.Response, error) {
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

	settings, err := c.service.UpdateAutoUpdate(user.ID.String(), req.AutoUpdate)
	if err != nil {
		c.logger.Log(logger.Error, "failed to update auto update setting", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.Response{
		Status:  "success",
		Message: "Auto update setting updated successfully",
		Data:    settings,
	}, nil
}

func (c *UserController) GetSettings(s fuego.ContextNoBody) (*types.Response, error) {
	w, r := s.Response(), s.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	settings, err := c.service.GetSettings(user.ID.String())
	if err != nil {
		c.logger.Log(logger.Error, "failed to get user settings", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.Response{
		Status:  "success",
		Message: "User settings fetched successfully",
		Data:    settings,
	}, nil
}
