package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/file-manager/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/ssh"
)

type CreateDirectoryRequest struct {
	Path string `json:"path"`
}

func (c *FileManagerController) CreateDirectory(f fuego.ContextWithBody[CreateDirectoryRequest]) (*types.MessageResponse, error) {
	request, err := f.Body()

	if err != nil {
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	err = c.service.CreateDirectory(f.Request().Context(), request.Path)
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
		Message: "Directory created successfully",
	}, nil
}
