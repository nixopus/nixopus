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

	uploadedCount := 0

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
		uploadedCount = 1
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
			var loopErr error
			func() {
				file, err := fileHeader.Open()
				if err != nil {
					c.logger.Log(logger.Error, "Failed to open uploaded file: "+err.Error(), "")
					loopErr = err
					return
				}
				defer func() {
					closeErr := file.Close()
					if closeErr != nil {
						c.logger.Log(logger.Error, "Failed to close uploaded file: "+closeErr.Error(), "")
					}
				}()
				relativePath := relativePaths[i]
				if relativePath == "" {
					relativePath = fileHeader.Filename
				}
				err = c.service.UploadFile(file, basePath, relativePath)
				if err != nil {
					c.logger.Log(logger.Error, "Failed to upload file "+relativePath+": "+err.Error(), "")
					loopErr = err
				}
			}()
			if loopErr != nil {
				return nil, fuego.HTTPError{
					Err:    loopErr,
					Status: http.StatusInternalServerError,
				}
			}
		}
		uploadedCount = len(files)
	}

	message := "Files uploaded successfully"
	if uploadedCount == 1 {
		message = "File uploaded successfully"
	}

	return &shared_types.Response{
		Status:  "success",
		Message: message,
	}, nil
}
