package types

type Container struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Image      string            `json:"image"`
	Status     string            `json:"status"`
	State      string            `json:"state"`
	Created    string            `json:"created"`
	Labels     map[string]string `json:"labels"`
	Ports      []Port            `json:"ports"`
	Mounts     []Mount           `json:"mounts"`
	Networks   []Network         `json:"networks"`
	Command    string            `json:"command"`
	IPAddress  string            `json:"ip_address"`
	HostConfig HostConfig        `json:"host_config"`
}

type Port struct {
	PrivatePort int    `json:"private_port"`
	PublicPort  int    `json:"public_port"`
	Type        string `json:"type"`
}

type Mount struct {
	Type        string `json:"type"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Mode        string `json:"mode"`
}

type Network struct {
	Name       string   `json:"name"`
	IPAddress  string   `json:"ip_address"`
	Gateway    string   `json:"gateway"`
	MacAddress string   `json:"mac_address"`
	Aliases    []string `json:"aliases"`
}

type HostConfig struct {
	Memory     int64 `json:"memory"`
	MemorySwap int64 `json:"memory_swap"`
	CPUShares  int64 `json:"cpu_shares"`
}

type ContainerStats struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage int64   `json:"memory_usage"`
	NetworkIO   struct {
		RxBytes int64 `json:"rx_bytes"`
		TxBytes int64 `json:"tx_bytes"`
	} `json:"network_io"`
	BlockIO struct {
		Read  int64 `json:"read"`
		Write int64 `json:"write"`
	} `json:"block_io"`
}

type ContainerListOptions struct {
	All     bool   `json:"all"`
	Limit   int    `json:"limit"`
	Since   string `json:"since"`
	Before  string `json:"before"`
	Size    bool   `json:"size"`
	Filters map[string][]string
}

type ContainerIDRequest struct {
	ID string `json:"id"`
}

type ContainerLogsRequest struct {
	ID     string `json:"id"`
	Follow bool   `json:"follow"`
	Tail   int    `json:"tail"`
	Since  string `json:"since"`
	Until  string `json:"until"`
	Stdout bool   `json:"stdout"`
	Stderr bool   `json:"stderr"`
}

type ContainerExecRequest struct {
	ID      string   `json:"id"`
	Command []string `json:"command"`
	User    string   `json:"user"`
	WorkDir string   `json:"work_dir"`
}

type ContainerExecResponse struct {
	ID string `json:"id"`
}

type Volume struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Mountpoint string            `json:"mountpoint"`
	Labels     map[string]string `json:"labels"`
	Options    map[string]string `json:"options"`
}

type VolumeCreateRequest struct {
	Name    string            `json:"name"`
	Driver  string            `json:"driver"`
	Labels  map[string]string `json:"labels"`
	Options map[string]string `json:"options"`
}

type VolumeListOptions struct {
	Filters map[string][]string `json:"filters"`
}

type ContainerListParams struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
	Search    string `json:"search"`
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
	Status    string `json:"status"`
	Name      string `json:"name"`
	Image     string `json:"image"`
}

type ContainerListRow struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Status  string            `json:"status"`
	State   string            `json:"state"`
	Created int64             `json:"created"`
	Labels  map[string]string `json:"labels"`
}

// ContainerGroup represents a group of containers belonging to the same application
type ContainerGroup struct {
	ApplicationID   string      `json:"application_id"`
	ApplicationName string      `json:"application_name"`
	Containers      []Container `json:"containers"`
}

// ListContainersResponseData contains the data for list containers response
type ListContainersResponseData struct {
	Containers []Container      `json:"containers"` // Deprecated: use Groups instead. Kept for backward compatibility
	Groups     []ContainerGroup `json:"groups,omitempty"`
	Ungrouped  []Container      `json:"ungrouped,omitempty"`
	TotalCount int              `json:"total_count"` // Total number of containers (not groups)
	GroupCount int              `json:"group_count"` // Total number of groups
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	SortBy     string           `json:"sort_by"`
	SortOrder  string           `json:"sort_order"`
	Search     string           `json:"search"`
	Status     string           `json:"status"`
	Name       string           `json:"name"`
	Image      string           `json:"image"`
}

// ListContainersResponse is the typed response for listing containers
type ListContainersResponse struct {
	Status  string                     `json:"status"`
	Message string                     `json:"message"`
	Data    ListContainersResponseData `json:"data"`
}

// GetContainerResponse is the typed response for getting a single container
type GetContainerResponse struct {
	Status  string    `json:"status"`
	Message string    `json:"message"`
	Data    Container `json:"data"`
}

// ContainerLogsResponse is the typed response for container logs
type ContainerLogsResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

// ContainerStatusData contains the status of a container operation
type ContainerStatusData struct {
	Status string `json:"status"`
}

// ContainerActionResponse is the typed response for container actions (start/stop/restart/remove)
type ContainerActionResponse struct {
	Status  string              `json:"status"`
	Message string              `json:"message"`
	Data    ContainerStatusData `json:"data"`
}

// MessageResponse is a generic response with just status and message
type MessageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ListImagesResponse is the typed response for listing images
type ListImagesResponse struct {
	Status  string  `json:"status"`
	Message string  `json:"message"`
	Data    []Image `json:"data"`
}

// ImageDeleteResponse represents a deleted image
type ImageDeleteResponse struct {
	Untagged string `json:"untagged,omitempty"`
	Deleted  string `json:"deleted,omitempty"`
}

// PruneImagesResponseData contains prune operation results
type PruneImagesResponseData struct {
	ImagesDeleted  []ImageDeleteResponse `json:"images_deleted,omitempty"`
	SpaceReclaimed uint64                `json:"space_reclaimed"`
}

// PruneImagesResponse is the typed response for pruning images
type PruneImagesResponse struct {
	Status  string                  `json:"status"`
	Message string                  `json:"message"`
	Data    PruneImagesResponseData `json:"data"`
}

// UpdateContainerResourcesRequest is the request body for updating container resource limits
type UpdateContainerResourcesRequest struct {
	Memory     int64 `json:"memory"`      // Memory limit in bytes (0 = unlimited)
	MemorySwap int64 `json:"memory_swap"` // Total memory limit (memory + swap) in bytes (0 = unlimited, -1 = unlimited swap)
	CPUShares  int64 `json:"cpu_shares"`  // CPU shares (relative weight)
}

// UpdateContainerResourcesResponseData contains the updated resource limits
type UpdateContainerResourcesResponseData struct {
	ContainerID string   `json:"container_id"`
	Memory      int64    `json:"memory"`
	MemorySwap  int64    `json:"memory_swap"`
	CPUShares   int64    `json:"cpu_shares"`
	Warnings    []string `json:"warnings,omitempty"`
}

// UpdateContainerResourcesResponse is the typed response for updating container resources
type UpdateContainerResourcesResponse struct {
	Status  string                               `json:"status"`
	Message string                               `json:"message"`
	Data    UpdateContainerResourcesResponseData `json:"data"`
}
