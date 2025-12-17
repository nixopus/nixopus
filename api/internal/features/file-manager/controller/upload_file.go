package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *FileManagerController) UploadFile(f fuego.ContextNoBody) (*shared_types.Response, error) {
	// Parse multipart form
	err := f.Request().ParseMultipartForm(32 << 20) // 32 MB max memory
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	basePath := f.Request().FormValue("path")
	if basePath == "" {
		basePath = "."
	}

	// Handle multiple files
	files := f.Request().MultipartForm.File["files"]
	if len(files) == 0 {
		// Fallback to single file for backward compatibility
		file, header, err := f.Request().FormFile("file")
		if err != nil {
			c.logger.Log(logger.Error, err.Error(), "")
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusBadRequest,
			}
		}
		defer file.Close()

		relativePath := f.Request().FormValue("relativePath")
		if relativePath == "" {
			relativePath = header.Filename
		}

		err = c.service.UploadFile(file, basePath, relativePath)
		if err != nil {
			c.logger.Log(logger.Error, err.Error(), "")
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusInternalServerError,
			}
		}
	} else {
		// Handle multiple files
		relativePaths := f.Request().Form["relativePaths"]
		if len(relativePaths) != len(files) {
			return nil, fuego.HTTPError{
				Err:    fuego.NewError(http.StatusBadRequest, "number of relativePaths must match number of files"),
				Status: http.StatusBadRequest,
			}
		}

		for i, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				c.logger.Log(logger.Error, "Failed to open uploaded file: "+err.Error(), "")
				continue
			}

			relativePath := relativePaths[i]
			if relativePath == "" {
				relativePath = fileHeader.Filename
			}

			err = c.service.UploadFile(file, basePath, relativePath)
			file.Close()

			if err != nil {
				c.logger.Log(logger.Error, "Failed to upload file "+relativePath+": "+err.Error(), "")
				return nil, fuego.HTTPError{
					Err:    err,
					Status: http.StatusInternalServerError,
				}
			}
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Files uploaded successfully",
	}, nil
}
