package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type ListFilesRequest struct {
	Path string `json:"path"`
}

func (c *FileManagerController) ListFiles(f fuego.ContextWithBody[ListFilesRequest]) (*shared_types.Response, error) {
	_, r := f.Response(), f.Request()
	path := r.URL.Query().Get("path")

	if path == "" {
		c.logger.Log(logger.Error, "path is required", "")
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	files, err := c.service.ListFiles(r.Context(), path)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Files fetched successfully",
		Data:    files,
	}, nil
}
