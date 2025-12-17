package controller

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *FileManagerController) DownloadFolder(f fuego.ContextNoBody) (*shared_types.Response, error) {
	path := f.QueryParam("path")
	if path == "" {
		return nil, fuego.HTTPError{
			Err:    fuego.NewError(http.StatusBadRequest, "path query parameter is required"),
			Status: http.StatusBadRequest,
		}
	}

	folderDownload, err := c.service.DownloadFolder(path)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}
	defer folderDownload.Reader.Close()

	// Set headers for folder download (ZIP)
	f.Response().Header().Set("Content-Disposition", "attachment; filename=\""+folderDownload.Name+"\"")
	f.Response().Header().Set("Content-Type", folderDownload.ContentType)
	f.Response().Header().Set("Content-Length", fmt.Sprintf("%d", folderDownload.Size))

	// Stream the ZIP content directly to response
	_, err = io.Copy(f.Response(), folderDownload.Reader)
	if err != nil {
		c.logger.Log(logger.Error, "Failed to stream folder content: "+err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	// Return nil to indicate we handled the response directly
	return nil, nil
}