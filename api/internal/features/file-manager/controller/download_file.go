package controller

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *FileManagerController) DownloadFile(f fuego.ContextNoBody) (*shared_types.Response, error) {
	path := f.QueryParam("path")
	if path == "" {
		return nil, fuego.HTTPError{
			Err:    fuego.NewError(http.StatusBadRequest, "path query parameter is required"),
			Status: http.StatusBadRequest,
		}
	}

	fileDownload, err := c.service.DownloadFile(path)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}
	defer fileDownload.Reader.Close()

	// Set headers for file download
	f.Response().Header().Set("Content-Disposition", "attachment; filename=\""+fileDownload.Name+"\"")
	f.Response().Header().Set("Content-Type", fileDownload.ContentType)
	f.Response().Header().Set("Content-Length", fmt.Sprintf("%d", fileDownload.Size))

	// Stream the file content directly to response
	_, err = io.Copy(f.Response(), fileDownload.Reader)
	if err != nil {
		c.logger.Log(logger.Error, "Failed to stream file content: "+err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	// Return nil to indicate we handled the response directly
	return nil, nil
}
