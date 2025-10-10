package types

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrDomainNotFound                          = errors.New("domain not found")
	ErrInvalidRequestType                      = errors.New("invalid request type")
	ErrMissingDomainName                       = errors.New("domain name is required")
	ErrInvalidDomainID                         = errors.New("invalid domain id")
	ErrMissingDomainID                         = errors.New("domain id is required")
	ErrDomainAlreadyExists                     = errors.New("domain already exists")
	ErrNotAllowed                              = errors.New("request not allowed")
	ErrDomainNameTooLong                       = errors.New("domain name too long")
	ErrDomainNameTooShort                      = errors.New("domain name too short")
	ErrInvalidUserID                           = errors.New("invalid user id")
	ErrInvalidAccess                           = errors.New("invalid access")
	ErrUserDoesNotBelongToOrganization         = errors.New("user does not belong to organization")
	ErrUserDoesNotHavePermissionForTheResource = errors.New("user does not have permission for the resource")
	ErrInvalidResource                         = errors.New("invalid resource")
	ErrMissingID                               = errors.New("id is required")
	ErrPermissionDenied                        = errors.New("permission denied")
	ErrAccessDenied                            = errors.New("access denied")
	ErrDomainNameInvalid                       = errors.New("invalid domain name")
	ErrFailedToCreateDomain                    = errors.New("failed to create domain")
	ErrFailedToDeleteDomain                    = errors.New("failed to delete domain")
	ErrFailedToUpdateDomain                    = errors.New("failed to update domain")
	ErrDomainDoesNotBelongToServer             = errors.New("domain does not belong to current server")
)

type CreateDomainRequest struct {
	Name           string    `json:"name"`
	OrganizationID uuid.UUID `json:"organization_id"`
}

type UpdateDomainRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type DeleteDomainRequest struct {
	ID string `json:"id"`
}

type CreateDomainResponse struct {
	ID string `json:"id"`
}

type RandomSubdomainResponse struct {
	Subdomain string `json:"subdomain"`
	Domain    string `json:"domain"`
}
