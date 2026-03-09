package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/file-manager/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
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
