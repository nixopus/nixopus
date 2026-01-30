package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// ProvisionStatus represents the status of user provision details
type ProvisionStatus string

const (
	ProvisionStatusPending            ProvisionStatus = "pending"
	ProvisionStatusInitializing       ProvisionStatus = "initializing"
	ProvisionStatusCreatingContainer  ProvisionStatus = "creating_container"
	ProvisionStatusConfiguringSSH     ProvisionStatus = "configuring_ssh"
	ProvisionStatusSettingUpSubdomain ProvisionStatus = "setting_up_subdomain"
	ProvisionStatusCompleted          ProvisionStatus = "completed"
	ProvisionStatusFailed             ProvisionStatus = "failed"
)

// UserProvisionDetails represents user provision details for an organization
type UserProvisionDetails struct {
	bun.BaseModel `bun:"table:user_provision_details,alias:upd" swaggerignore:"true"`

	ID               uuid.UUID       `json:"id" bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	UserID           uuid.UUID       `json:"user_id" bun:"user_id,notnull,type:uuid"`
	OrganizationID   uuid.UUID       `json:"organization_id" bun:"organization_id,notnull,type:uuid"`
	LXDContainerName *string         `json:"lxd_container_name,omitempty" bun:"lxd_container_name"`
	SSHKeyID         *uuid.UUID      `json:"ssh_key_id,omitempty" bun:"ssh_key_id,type:uuid"`
	Subdomain        *string         `json:"subdomain,omitempty" bun:"subdomain"`
	Domain           *string         `json:"domain,omitempty" bun:"domain"`
	Status           ProvisionStatus `json:"status" bun:"status,notnull,type:provision_status,default:'pending'"`
	Error            *string         `json:"error,omitempty" bun:"error,type:text"`
	CreatedAt        time.Time       `json:"created_at" bun:"created_at,notnull,default:now()"`
	UpdatedAt        time.Time       `json:"updated_at" bun:"updated_at,notnull,default:now()"`

	User         *User         `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
	Organization *Organization `json:"organization,omitempty" bun:"rel:belongs-to,join:organization_id=id"`
	SSHKey       *SSHKey       `json:"ssh_key,omitempty" bun:"rel:belongs-to,join:ssh_key_id=id"`
}
