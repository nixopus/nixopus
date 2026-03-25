package types

import (
	"errors"

	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

type AddCustomDomainRequest struct {
	Name string `json:"name"`
}

type VerifyCustomDomainRequest struct {
	ID string `json:"id"`
}

type RemoveCustomDomainRequest struct {
	ID string `json:"id"`
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
	ErrCustomDomainNotFound    = errors.New("custom domain not found")
	ErrDNSNotVerified          = errors.New("DNS records not verified")
	ErrSubscriptionRequired    = errors.New("active subscription required for custom domains")
	ErrMaxCustomDomainsReached = errors.New("maximum custom domains limit reached")
	ErrInvalidCustomDomain     = errors.New("invalid custom domain")
)
