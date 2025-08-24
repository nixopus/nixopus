package validation

import (
	"net"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
)

// Validator handles domain validation logic
type Validator struct {
	storage storage.DomainStorageInterface
}

// NewValidator creates a new validator instance
func NewValidator(storage storage.DomainStorageInterface) *Validator {
	return &Validator{
		storage: storage,
	}
}

// ValidateID validates the domain ID is a valid UUID
func (v *Validator) ValidateID(id string) error {
	if id == "" {
		return types.ErrMissingDomainID
	}
	if _, err := uuid.Parse(id); err != nil {
		return types.ErrInvalidDomainID
	}
	return nil
}

// ValidateName validates domain name meets requirements
func (v *Validator) ValidateName(name string) error {
	if name == "" {
		return types.ErrMissingDomainName
	}

	if len(name) < 3 {
		return types.ErrDomainNameTooShort
	}

	if len(name) > 255 {
		return types.ErrDomainNameTooLong
	}

	if !strings.Contains(name, ".") {
		return types.ErrDomainNameInvalid
	}

	tld := strings.Split(name, ".")[1]
	if len(tld) < 2 || len(tld) > 63 {
		return types.ErrDomainNameInvalid
	}

	return nil
}

// ValidateRequest validates different domain request types
func (v *Validator) ValidateRequest(req interface{}) error {
	switch r := req.(type) {
	case *types.CreateDomainRequest:
		return v.ValidateCreateDomainRequest(*r)
	case *types.UpdateDomainRequest:
		return v.ValidateUpdateDomainRequest(*r)
	case *types.DeleteDomainRequest:
		return v.ValidateDeleteDomainRequest(*r)
	default:
		return types.ErrInvalidRequestType
	}
}

// validateCreateDomainRequest validates domain creation requests
func (v *Validator) ValidateCreateDomainRequest(req types.CreateDomainRequest) error {
	err := v.ValidateName(req.Name)
	if err != nil {
		return err
	}
	err = v.ValidateID(req.OrganizationID.String())
	if err != nil {
		return err
	}
	return nil
}

// validateUpdateDomainRequest validates domain update requests
func (v *Validator) ValidateUpdateDomainRequest(req types.UpdateDomainRequest) error {
	// Validate ID first
	if err := v.ValidateID(req.ID); err != nil {
		return err
	}

	// Validate name
	if err := v.ValidateName(req.Name); err != nil {
		return err
	}

	return nil
}

// validateDeleteDomainRequest validates domain deletion requests
func (v *Validator) ValidateDeleteDomainRequest(req types.DeleteDomainRequest) error {
	// Validate ID first
	if err := v.ValidateID(req.ID); err != nil {
		return err
	}

	return nil
}

// ValidateDomainBelongsToServer checks if the domain belongs to the current server by resolving its IP
func (v *Validator) ValidateDomainBelongsToServer(domainName string) error {
	development := config.AppConfig.App.Environment == "development"
	if development {
		return nil
	}

	serverHost := config.AppConfig.SSH.Host
	if serverHost == "" {
		var err error
		serverHost, err = os.Hostname()
		if err != nil {
			return types.ErrDomainDoesNotBelongToServer
		}
	}

	// Handle wildcard domains by extracting main domain
	mainDomain := domainName
	if strings.HasPrefix(domainName, "*.") {
		mainDomain = strings.TrimPrefix(domainName, "*.")
	}

	domainIPs, err := net.LookupIP(mainDomain)
	if err != nil {
		return types.ErrDomainDoesNotBelongToServer
	}

	serverIPs, err := net.LookupIP(serverHost)
	if err != nil {
		return types.ErrDomainDoesNotBelongToServer
	}

	for _, domainIP := range domainIPs {
		for _, serverIP := range serverIPs {
			if domainIP.Equal(serverIP) {
				return nil
			}
		}
	}

	return types.ErrDomainDoesNotBelongToServer
}
