package controller

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

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

	// Path validation to prevent directory traversal
	if filepath.IsAbs(path) {
		return nil, fuego.HTTPError{
			Err:    fuego.NewError(http.StatusBadRequest, "absolute paths are not allowed"),
			Status: http.StatusBadRequest,
		}
	}

	cleanedPath := filepath.Clean(path)
	baseDir := "." // Base directory for relative path validation
	rel, err := filepath.Rel(baseDir, filepath.Join(baseDir, cleanedPath))
	if err != nil || strings.HasPrefix(rel, "..") {
		return nil, fuego.HTTPError{
			Err:    fuego.NewError(http.StatusBadRequest, "invalid path: directory traversal detected"),
			Status: http.StatusBadRequest,
		}
	}

	folderDownload, err := c.service.DownloadFolder(cleanedPath)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}
	defer folderDownload.Reader.Close()

	// Set headers for folder download (ZIP)
	f.Response().Header().Set("Content-Disposition", SanitizeFilenameForHeader(folderDownload.Name))
	f.Response().Header().Set("Content-Type", folderDownload.ContentType)
	f.Response().Header().Set("Content-Length", fmt.Sprintf("%d", folderDownload.Size))

	// Stream the ZIP content directly to response
	// NOTE: Once streaming begins, we cannot change HTTP status codes on errors.
	// Clients should verify Content-Length matches bytes received to detect truncated transfers.
	bytesCopied, err := io.Copy(f.Response(), folderDownload.Reader)
	if err != nil {
		// Log the error with details, but cannot return HTTP error since response has started
		c.logger.Log(logger.Error, fmt.Sprintf("Failed to stream folder content for %s: copied %d/%d bytes, error: %v",
			path, bytesCopied, folderDownload.Size, err), "")
		// Response is already partially sent, so we can't change status code
		// Clients can detect issues by comparing received bytes to Content-Length
		return nil, nil
	}

	// Log successful transfer
	c.logger.Log(logger.Info, fmt.Sprintf("Successfully streamed folder %s: %d bytes", path, bytesCopied), "")

	// Return nil to indicate we handled the response directly
	return nil, nil
}