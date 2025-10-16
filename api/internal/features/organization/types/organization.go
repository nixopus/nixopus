package types

import (
	"time"

	"github.com/google/uuid"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type ResourceType string

const (
	ResourceTypeUser            ResourceType = "user"
	ResourceTypeOrganization    ResourceType = "organization"
	ResourceTypeRole            ResourceType = "role"
	ResourceTypePermission      ResourceType = "permission"
	ResourceTypeDomain          ResourceType = "domain"
	ResourceTypeGithubConnector ResourceType = "github-connector"
	ResourceTypeNotification    ResourceType = "notification"
	ResourceTypeFileManager     ResourceType = "file-manager"
	ResourceTypeDeploy          ResourceType = "deploy"
	ResourceTypeAudit           ResourceType = "audit"
)

type CreateOrganizationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateOrganizationRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type DeleteOrganizationRequest struct {
	ID string `json:"id"`
}

type AddUserToOrganizationRequest struct {
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
}

type RemoveUserFromOrganizationRequest struct {
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
}

type InviteSendRequest struct {
	Email          string `json:"email"`
	OrganizationID string `json:"organization_id"`
	Role           string `json:"role"`
}

type InviteResendRequest struct {
	Email          string `json:"email"`
	OrganizationID string `json:"organization_id"`
	Role           string `json:"role"`
}

type UpdateUserRoleRequest struct {
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
	Role           string `json:"role"`
}

func NewOrganization(name string, description string) shared_types.Organization {
	return shared_types.Organization{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		DeletedAt:   nil,
	}
}
