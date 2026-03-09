package types

import (
	"errors"

	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

type MessageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ListDomainsResponse struct {
	Status  string                `json:"status"`
	Message string                `json:"message"`
	Data    []shared_types.Domain `json:"data"`
}

type RandomSubdomainResponseWrapper struct {
	Status  string                  `json:"status"`
	Message string                  `json:"message"`
	Data    RandomSubdomainResponse `json:"data"`
}

type RandomSubdomainResponse struct {
	Subdomain string `json:"subdomain"`
	Domain    string `json:"domain"`
}

var (
	ErrDomainNotFound                  = errors.New("domain not found")
	ErrDomainAlreadyExists             = errors.New("domain already exists")
	ErrMissingID                       = errors.New("id is required")
	ErrAccessDenied                    = errors.New("access denied")
	ErrDomainNameInvalid               = errors.New("invalid domain name")
	ErrDomainNameTooLong               = errors.New("domain name too long")
	ErrDomainNameTooShort              = errors.New("domain name too short")
	ErrMissingDomainName               = errors.New("domain name is required")
	ErrInvalidDomainID                 = errors.New("invalid domain id")
	ErrMissingDomainID                 = errors.New("domain id is required")
	ErrPermissionDenied                = errors.New("permission denied")
	ErrUserDoesNotBelongToOrganization = errors.New("user does not belong to organization")
	ErrDomainDoesNotBelongToServer     = errors.New("domain does not belong to current server")
)
