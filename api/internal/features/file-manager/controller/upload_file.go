package controller

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

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

	filename := filepath.Base(header.Filename)

	if err := validateUploadPath(path); err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}
	if err := validateUploadPath(filename); err != nil {
		return nil, fuego.HTTPError{
			Err:    fmt.Errorf("invalid filename: %w", err),
			Status: http.StatusBadRequest,
		}
	}

	err = c.service.UploadFile(f.Request().Context(), file, path, filename)
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

// validateUploadPath rejects path traversal and null bytes.
func validateUploadPath(p string) error {
	if strings.Contains(p, "\x00") {
		return fmt.Errorf("path contains null bytes")
	}
	cleaned := filepath.Clean(p)
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("path traversal is not allowed")
	}
	return nil
}
