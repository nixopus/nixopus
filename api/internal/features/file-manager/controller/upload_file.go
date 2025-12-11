package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *FileManagerController) UploadFile(f fuego.ContextNoBody) (*shared_types.Response, error) {
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

	err = c.service.UploadFile(f.Request().Context(), file, path, header.Filename)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "File uploaded successfully",
	}, nil
}
