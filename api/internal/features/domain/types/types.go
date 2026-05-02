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

type AddCustomDomainRequest struct {
	Name string `json:"name" validate:"required,min=3,max=255" description:"Fully qualified domain name to add" example:"app.example.com"`
}

type VerifyCustomDomainRequest struct {
	ID string `json:"id" validate:"required" description:"Custom domain ID to verify DNS records for" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type RemoveCustomDomainRequest struct {
	ID string `json:"id" validate:"required" description:"Custom domain ID to remove" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type CustomDomainResponse struct {
	Status  string               `json:"status"`
	Message string               `json:"message"`
	Data    *shared_types.Domain `json:"data"`
}

type CustomDomainListResponse struct {
	Status  string                `json:"status"`
	Message string                `json:"message"`
	Data    []shared_types.Domain `json:"data"`
}

type DNSInstruction struct {
	RecordType  string `json:"record_type"`
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

type DNSSetupResponse struct {
	Status       string               `json:"status"`
	Message      string               `json:"message"`
	Data         *shared_types.Domain `json:"data"`
	Instructions []DNSInstruction     `json:"instructions"`
	DNSProvider  string               `json:"dns_provider"`
}

type DNSCheckResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Verified  bool   `json:"verified"`
	DNSStatus string `json:"dns_status"`
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

	ErrCustomDomainNotFound    = errors.New("custom domain not found")
	ErrDNSNotVerified          = errors.New("DNS records not verified")
	ErrSubscriptionRequired    = errors.New("active subscription required for custom domains")
	ErrMaxCustomDomainsReached = errors.New("maximum custom domains limit reached")
	ErrInvalidCustomDomain     = errors.New("invalid custom domain")
)
