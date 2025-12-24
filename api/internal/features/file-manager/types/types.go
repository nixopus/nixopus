package types

import (
	"github.com/raghavyuva/nixopus-api/internal/features/file-manager/service"
)

// MessageResponse is a generic response with just status and message
type MessageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ListFilesResponse is the typed response for listing files
type ListFilesResponse struct {
	Status  string             `json:"status"`
	Message string             `json:"message"`
	Data    []service.FileData `json:"data"`
}
