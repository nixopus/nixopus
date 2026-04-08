package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// ProvisionStep represents the granular step in a provisioning operation
// Maps to provision_step enum in database
type ProvisionStep string

const (
	ProvisionStepInitializing       ProvisionStep = "INITIALIZING"
	ProvisionStepCreatingContainer  ProvisionStep = "CREATING_CONTAINER"
	ProvisionStepSetupNetworking    ProvisionStep = "SETUP_NETWORKING"
	ProvisionStepInstallingDeps     ProvisionStep = "INSTALLING_DEPENDENCIES"
	ProvisionStepConfiguringSSH     ProvisionStep = "CONFIGURING_SSH"
	ProvisionStepSetupSSHForwarding ProvisionStep = "SETUP_SSH_FORWARDING"
	ProvisionStepVerifyingSSH       ProvisionStep = "VERIFYING_SSH"
	ProvisionStepCompleted          ProvisionStep = "COMPLETED"
)

// UserProvisionDetails represents user provision details for an organization
type UserProvisionDetails struct {
	bun.BaseModel `bun:"table:user_provision_details,alias:upd" swaggerignore:"true"`

	ID               uuid.UUID      `json:"id" bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	UserID           uuid.UUID      `json:"user_id" bun:"user_id,notnull,type:uuid"`
	OrganizationID   uuid.UUID      `json:"organization_id" bun:"organization_id,notnull,type:uuid"`
	ServerID         *uuid.UUID     `json:"server_id,omitempty" bun:"server_id,type:uuid"`
	GuestIP          *string        `json:"guest_ip,omitempty" bun:"guest_ip"`
	LXDContainerName *string        `json:"lxd_container_name,omitempty" bun:"lxd_container_name"`
	SSHKeyID         *uuid.UUID     `json:"ssh_key_id,omitempty" bun:"ssh_key_id,type:uuid"`
	Subdomain        *string        `json:"subdomain,omitempty" bun:"subdomain"`
	Domain           *string        `json:"domain,omitempty" bun:"domain"`
	VcpuCount        int            `json:"vcpu_count" bun:"vcpu_count"`
	MemoryMB         int            `json:"memory_mb" bun:"memory_mb"`
	DiskSizeGB       int            `json:"disk_size_gb" bun:"disk_size_gb"`
	Step             *ProvisionStep `json:"step,omitempty" bun:"step,type:provision_step"`
	Error            *string        `json:"error,omitempty" bun:"error,type:text"`
	CreatedAt        time.Time      `json:"created_at" bun:"created_at,notnull,default:now()"`
	UpdatedAt        time.Time      `json:"updated_at" bun:"updated_at,notnull,default:now()"`

	User         *User         `json:"-" bun:"rel:belongs-to,join:user_id=id"`
	Organization *Organization `json:"-" bun:"rel:belongs-to,join:organization_id=id"`
	SSHKey       *SSHKey       `json:"-" bun:"rel:belongs-to,join:ssh_key_id=id"`
}
