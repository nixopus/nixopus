package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type CreateDirectoryRequest struct {
	Path string `json:"path"`
}

func (c *FileManagerController) CreateDirectory(f fuego.ContextWithBody[CreateDirectoryRequest]) (*shared_types.Response, error) {
	request, err := f.Body()

	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	err = c.service.CreateDirectory(f.Request().Context(), request.Path)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Directory created successfully",
	}, nil
}
