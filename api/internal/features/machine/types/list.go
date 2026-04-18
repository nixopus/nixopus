package types

import (
	"errors"

	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

var (
	ErrMachineNotFound = errors.New("server not found")
	ErrMachineInactive = errors.New("cannot set an inactive server as default")
)

type MachineListParams struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
	Search    string `json:"search"`
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
	Status    string `json:"status"`
	IsActive  *bool  `json:"is_active"`
}

type MachineResponse struct {
	shared_types.SSHKey
	Provision   *shared_types.UserProvisionDetails `json:"provision,omitempty"`
	TotalVcpu   int                                `json:"total_vcpu"`
	TotalRamMB  int                                `json:"total_ram_mb"`
	TotalDiskGB int                                `json:"total_disk_gb"`
}

type ListMachinesResponseData struct {
	Servers    []MachineResponse `json:"servers"`
	TotalCount int               `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	SortBy     string            `json:"sort_by"`
	SortOrder  string            `json:"sort_order"`
	Search     string            `json:"search"`
	Status     string            `json:"status"`
	IsActive   *bool             `json:"is_active,omitempty"`
}

type ListMachinesResponse struct {
	Status  string                   `json:"status"`
	Message string                   `json:"message"`
	Data    ListMachinesResponseData `json:"data"`
}

type SetDefaultMachineResponse struct {
	Status  string              `json:"status"`
	Message string              `json:"message"`
	Data    shared_types.SSHKey `json:"data"`
}

type SSHConnectionStatusResponse struct {
	Status       string `json:"status"`
	Connected    bool   `json:"connected"`
	Message      string `json:"message"`
	IsConfigured bool   `json:"is_configured"`
}
