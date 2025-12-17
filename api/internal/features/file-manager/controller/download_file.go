package controller

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// SanitizeFilenameForHeader sanitizes a filename for safe use in HTTP Content-Disposition header
// and returns a properly formatted header value with both filename and filename* parameters
func SanitizeFilenameForHeader(originalName string) string {
	const maxFilenameLength = 255

	// Step 1: Remove control characters (including CR/LF) and other dangerous chars
	sanitized := strings.Map(func(r rune) rune {
		// Remove control characters, quotes, semicolons, and other header injection chars
		if r < 32 || r == 127 || r == '"' || r == ';' || r == '\\' || r == '\r' || r == '\n' {
			return -1
		}
		return r
	}, originalName)

	// Step 2: Trim whitespace and limit length
	sanitized = strings.TrimSpace(sanitized)
	if len(sanitized) > maxFilenameLength {
 		// Truncate by runes to avoid splitting multi-byte characters
 		runes := []rune(sanitized)
 		if len(runes) > maxFilenameLength {
 			sanitized = string(runes[:maxFilenameLength])
 		}
 		// Ensure we didn't break a UTF-8 sequence
 		for !utf8.ValidString(sanitized) && len(sanitized) > 0 {
 			sanitized = sanitized[:len(sanitized)-1]
 		}
 	}

	// Step 3: Ensure we have a non-empty filename
	if sanitized == "" {
		sanitized = "download"
	}

	// Step 4: Create ASCII-safe version for filename parameter
	asciiSafe := strings.Map(func(r rune) rune {
		if r >= 32 && r < 127 && r != '"' && r != ';' && r != '\\' {
			return r
		}
		return '_'
	}, sanitized)

	// Step 5: Create RFC 5987 encoded version for filename*
 	var encoded string
 	if utf8.ValidString(sanitized) {
 		// Use proper percent-encoding for RFC 5987 (not QueryEscape which uses ' ' for spaces)
 		encoded = strings.ReplaceAll(url.PathEscape(sanitized), " ", "%20")
 	} else {
 		// If not valid UTF-8, fall back to ASCII version
 		encoded = strings.ReplaceAll(url.PathEscape(asciiSafe), " ", "%20")
 	}

	// Step 6: Format the Content-Disposition header
	// Use filename for ASCII-safe names, filename* for UTF-8
	if asciiSafe == sanitized {
		// ASCII-safe, use simple filename
		return fmt.Sprintf("attachment; filename=\"%s\"", asciiSafe)
	} else {
		// UTF-8, use both filename (ASCII fallback) and filename*
		return fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", asciiSafe, encoded)
	}
}

func (c *FileManagerController) DownloadFile(f fuego.ContextNoBody) (*shared_types.Response, error) {
	path := f.QueryParam("path")
	if path == "" {
		return nil, fuego.HTTPError{
			Err:    fuego.NewError(http.StatusBadRequest, "path query parameter is required"),
			Status: http.StatusBadRequest,
		}
	}
	
	cleanPath := filepath.Clean(path) // Ensure the cleaned path doesn't escape the intended directory
	if strings.HasPrefix(cleanPath, "..") || strings.HasPrefix(cleanPath, "/") {
		return nil, fuego.HTTPError{
			Err:    fuego.NewError(http.StatusBadRequest, "invalid path"),
			Status: http.StatusBadRequest,
		}
	}
	//pass the cleanPath to the service instead of raw `path`
	fileDownload, err := c.service.DownloadFile(cleanPath)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    fuego.NewError(http.StatusInternalServerError, "failed to download file"),
			Status: http.StatusInternalServerError,
		}
	}
	defer fileDownload.Reader.Close()

	// Set headers for file download
	f.Response().Header().Set("Content-Disposition", SanitizeFilenameForHeader(fileDownload.Name))
	f.Response().Header().Set("Content-Type", fileDownload.ContentType)
	f.Response().Header().Set("Content-Length", fmt.Sprintf("%d", fileDownload.Size))

	// Stream the file content directly to response
	// NOTE: Once streaming begins, we cannot change HTTP status codes on errors.
	// Clients should verify Content-Length matches bytes received to detect truncated transfers.
	bytesCopied, err := io.Copy(f.Response(), fileDownload.Reader)
	if err != nil {
		// Log the error with details, but cannot return HTTP error since response has started
		c.logger.Log(logger.Error, fmt.Sprintf("Failed to stream file content for %s: copied %d/%d bytes, error: %v",
			path, bytesCopied, fileDownload.Size, err), "")
		// Response is already partially sent, so we can't change status code
		// Clients can detect issues by comparing received bytes to Content-Length
		return nil, nil
	}

	// Log successful transfer
	c.logger.Log(logger.Info, fmt.Sprintf("Successfully streamed file %s: %d bytes", path, bytesCopied), "")

	// Return nil to indicate we handled the response directly
	return nil, nil
}
