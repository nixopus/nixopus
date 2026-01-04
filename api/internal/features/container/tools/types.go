package tools

import container_types "github.com/raghavyuva/nixopus-api/internal/features/container/types"

// GetContainerLogsInput is the input structure for the MCP tool
type GetContainerLogsInput struct {
	ID     string  `json:"id" jsonschema:"required"`
	Follow bool    `json:"follow,omitempty"`
	Tail   *int    `json:"tail,omitempty"`
	Since  *string `json:"since,omitempty"`
	Until  *string `json:"until,omitempty"`
	Stdout bool    `json:"stdout,omitempty"`
	Stderr bool    `json:"stderr,omitempty"`
}

// GetContainerLogsOutput is the output structure for the MCP tool
type GetContainerLogsOutput struct {
	Logs string `json:"logs"`
}

// GetContainerInput is the input structure for the MCP tool
type GetContainerInput struct {
	ID string `json:"id" jsonschema:"required"`
}

// GetContainerOutput is the output structure for the MCP tool
type GetContainerOutput struct {
	Container container_types.Container `json:"container"`
}

// ListContainersInput is the input structure for the MCP tool
type ListContainersInput struct {
	Page      *int   `json:"page,omitempty"`
	PageSize  *int   `json:"page_size,omitempty"`
	Search    string `json:"search,omitempty"`
	SortBy    string `json:"sort_by,omitempty"`
	SortOrder string `json:"sort_order,omitempty"`
	Status    string `json:"status,omitempty"`
	Name      string `json:"name,omitempty"`
	Image     string `json:"image,omitempty"`
}

// ListContainersOutput is the output structure for the MCP tool
type ListContainersOutput struct {
	Response container_types.ListContainersResponse `json:"response"`
}

// ListImagesInput is the input structure for the MCP tool
type ListImagesInput struct {
	All         *bool  `json:"all,omitempty"`
	ContainerID string `json:"container_id,omitempty"`
	ImagePrefix string `json:"image_prefix,omitempty"`
}

// ListImagesOutput is the output structure for the MCP tool
type ListImagesOutput struct {
	Response container_types.ListImagesResponse `json:"response"`
}

// PruneImagesInput is the input structure for the MCP tool
type PruneImagesInput struct {
	Until    string `json:"until,omitempty"`
	Label    string `json:"label,omitempty"`
	Dangling *bool  `json:"dangling,omitempty"`
}

// PruneImagesOutput is the output structure for the MCP tool
type PruneImagesOutput struct {
	Response container_types.PruneImagesResponse `json:"response"`
}

// PruneBuildCacheInput is the input structure for the MCP tool
type PruneBuildCacheInput struct {
	All *bool `json:"all,omitempty"`
}

// PruneBuildCacheOutput is the output structure for the MCP tool
type PruneBuildCacheOutput struct {
	Response container_types.MessageResponse `json:"response"`
}

// RemoveContainerInput is the input structure for the MCP tool
type RemoveContainerInput struct {
	ID    string `json:"id" jsonschema:"required"`
	Force *bool  `json:"force,omitempty"`
}

// RemoveContainerOutput is the output structure for the MCP tool
type RemoveContainerOutput struct {
	Response container_types.ContainerActionResponse `json:"response"`
}

// RestartContainerInput is the input structure for the MCP tool
type RestartContainerInput struct {
	ID      string `json:"id" jsonschema:"required"`
	Timeout *int   `json:"timeout,omitempty"`
}

// RestartContainerOutput is the output structure for the MCP tool
type RestartContainerOutput struct {
	Response container_types.ContainerActionResponse `json:"response"`
}

// StartContainerInput is the input structure for the MCP tool
type StartContainerInput struct {
	ID string `json:"id" jsonschema:"required"`
}

// StartContainerOutput is the output structure for the MCP tool
type StartContainerOutput struct {
	Response container_types.ContainerActionResponse `json:"response"`
}

// StopContainerInput is the input structure for the MCP tool
type StopContainerInput struct {
	ID      string `json:"id" jsonschema:"required"`
	Timeout *int   `json:"timeout,omitempty"`
}

// StopContainerOutput is the output structure for the MCP tool
type StopContainerOutput struct {
	Response container_types.ContainerActionResponse `json:"response"`
}

// UpdateContainerResourcesInput is the input structure for the MCP tool
type UpdateContainerResourcesInput struct {
	ID         string `json:"id" jsonschema:"required"`
	Memory     *int64 `json:"memory,omitempty"`      // Memory limit in bytes (0 = unlimited)
	MemorySwap *int64 `json:"memory_swap,omitempty"` // Total memory limit in bytes (0 = unlimited, -1 = unlimited swap)
	CPUShares  *int64 `json:"cpu_shares,omitempty"`  // CPU shares (relative weight)
}

// UpdateContainerResourcesOutput is the output structure for the MCP tool
type UpdateContainerResourcesOutput struct {
	Response container_types.UpdateContainerResourcesResponse `json:"response"`
}
