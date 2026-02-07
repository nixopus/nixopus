package types

import (
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// ServerListParams represents query parameters for listing servers
type ServerListParams struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
	Search    string `json:"search"`
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
	Status    string `json:"status"`    // Filter by provision status
	IsActive  *bool  `json:"is_active"` // Filter by SSH key active status
}

// ServerResponse represents a server (SSH key) with optional provision details
type ServerResponse struct {
	shared_types.SSHKey
	Provision *shared_types.UserProvisionDetails `json:"provision,omitempty"`
}

// ListServersResponseData contains the data for list servers response
type ListServersResponseData struct {
	Servers    []ServerResponse `json:"servers"`
	TotalCount int              `json:"total_count"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	SortBy     string           `json:"sort_by"`
	SortOrder  string           `json:"sort_order"`
	Search     string           `json:"search"`
	Status     string           `json:"status"`
	IsActive   *bool            `json:"is_active,omitempty"`
}

// ListServersResponse is the typed response for listing servers
type ListServersResponse struct {
	Status  string                  `json:"status"`
	Message string                  `json:"message"`
	Data    ListServersResponseData `json:"data"`
}
