package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/file-manager/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

func (c *FileManagerController) UploadFile(f fuego.ContextNoBody) (*types.MessageResponse, error) {
	file, header, err := f.Request().FormFile("file")
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}
	defer file.Close()

	path := f.Request().FormValue("path")
	if path == "" {
		path = "."
	}

	err = c.service.UploadFile(file, path, header.Filename)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.MessageResponse{
		Status:  "success",
		Message: "File uploaded successfully",
	}, nil
}
