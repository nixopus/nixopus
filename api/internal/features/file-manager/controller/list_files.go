package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/file-manager/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

func (c *FileManagerController) ListFiles(f fuego.ContextNoBody) (*types.ListFilesResponse, error) {
	_, r := f.Response(), f.Request()
	path := r.URL.Query().Get("path")

	if path == "" {
		c.logger.Log(logger.Error, "path is required", "")
		return nil, fuego.BadRequestError{
			Detail: "path is required",
		}
	}

	files, err := c.service.ListFiles(f.Request().Context(), path)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.ListFilesResponse{
		Status:  "success",
		Message: "Files fetched successfully",
		Data:    files,
	}, nil
}
