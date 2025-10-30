package types

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrServerNotFound                          = errors.New("server not found")
	ErrInvalidRequestType                      = errors.New("invalid request type")
	ErrMissingServerName                       = errors.New("server name is required")
	ErrInvalidServerID                         = errors.New("invalid server id")
	ErrMissingServerID                         = errors.New("server id is required")
	ErrServerAlreadyExists                     = errors.New("server already exists")
	ErrServerHostAlreadyExists                 = errors.New("server with this host and port already exists")
	ErrNotAllowed                              = errors.New("request not allowed")
	ErrServerNameTooLong                       = errors.New("server name too long")
	ErrServerNameTooShort                      = errors.New("server name too short")
	ErrInvalidUserID                           = errors.New("invalid user id")
	ErrInvalidAccess                           = errors.New("invalid access")
	ErrUserDoesNotBelongToOrganization         = errors.New("user does not belong to organization")
	ErrUserDoesNotHavePermissionForTheResource = errors.New("user does not have permission for the resource")
	ErrInvalidResource                         = errors.New("invalid resource")
	ErrMissingID                               = errors.New("id is required")
	ErrPermissionDenied                        = errors.New("permission denied")
	ErrAccessDenied                            = errors.New("access denied")
	ErrServerNameInvalid                       = errors.New("invalid server name")
	ErrNoRoleAssigned                          = errors.New("no role assigned")
	ErrFailedToCreateServer                    = errors.New("failed to create server")
	ErrFailedToDeleteServer                    = errors.New("failed to delete server")
	ErrFailedToUpdateServer                    = errors.New("failed to update server")
	ErrMissingHost                             = errors.New("host is required")
	ErrInvalidHost                             = errors.New("invalid host")
	ErrMissingPort                             = errors.New("port is required")
	ErrInvalidPort                             = errors.New("invalid port")
	ErrMissingUsername                         = errors.New("username is required")
	ErrMissingSSHAuth                          = errors.New("either ssh_password or ssh_private_key_path is required")
	ErrBothSSHAuthProvided                     = errors.New("provide either ssh_password or ssh_private_key_path, not both")
	ErrInvalidSSHPrivateKeyPath                = errors.New("invalid ssh private key path")
	ErrSSHConnectionFailed                     = errors.New("failed to establish SSH connection")
	ErrSSHAuthenticationFailed                 = errors.New("SSH authentication failed")
	ErrMissingStatus                           = errors.New("status is required")
	ErrInvalidStatus                           = errors.New("invalid status")
)

type CreateServerRequest struct {
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Host              string    `json:"host"`
	Port              int       `json:"port"`
	Username          string    `json:"username"`
	SSHPassword       *string   `json:"ssh_password,omitempty"`
	SSHPrivateKeyPath *string   `json:"ssh_private_key_path,omitempty"`
	Status            *string   `json:"status,omitempty"`
	OrganizationID    uuid.UUID `json:"organization_id"`
}

type UpdateServerRequest struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	Host              string  `json:"host"`
	Port              int     `json:"port"`
	Username          string  `json:"username"`
	SSHPassword       *string `json:"ssh_password,omitempty"`
	SSHPrivateKeyPath *string `json:"ssh_private_key_path,omitempty"`
	Status            *string `json:"status,omitempty"`
}

type UpdateServerStatusRequest struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type DeleteServerRequest struct {
	ID string `json:"id"`
}

type CreateServerResponse struct {
	ID string `json:"id"`
}

type ServerResponseWithoutSecrets struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Host           string    `json:"host"`
	Port           int       `json:"port"`
	Username       string    `json:"username"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	UserID         uuid.UUID `json:"user_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
}
