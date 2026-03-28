package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/file-manager/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/ssh"
)

type CopyDirectory struct {
	FromPath string `json:"from_path"`
	ToPath   string `json:"to_path"`
}

func (c *FileManagerController) CopyDirectory(f fuego.ContextWithBody[CopyDirectory]) (*types.MessageResponse, error) {
	request, err := f.Body()

	if err != nil {
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	err = c.service.CopyDirectory(f.Request().Context(), request.FromPath, request.ToPath)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		if ssh.IsNoDefaultServerError(err) {
			return nil, fuego.HTTPError{Status: http.StatusServiceUnavailable, Detail: err.Error()}
		}
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.MessageResponse{
		Status:  "success",
		Message: "Directory copied successfully",
	}, nil
}
