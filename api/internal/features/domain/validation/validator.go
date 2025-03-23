package validation

import (
	"encoding/json"
	"io"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
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

	return nil
}

// ParseRequestBody decodes JSON request body
func (v *Validator) ParseRequestBody(req interface{}, body io.ReadCloser, decoded interface{}) error {
	return json.NewDecoder(body).Decode(decoded)
}

// ValidateRequest validates different domain request types
func (v *Validator) ValidateRequest(req interface{}, user shared_types.User) error {
	switch r := req.(type) {
	case *types.CreateDomainRequest:
		return v.ValidateCreateDomainRequest(*r)
	case *types.UpdateDomainRequest:
		return v.ValidateUpdateDomainRequest(*r, user)
	case *types.DeleteDomainRequest:
		return v.ValidateDeleteDomainRequest(*r, user)
	default:
		return types.ErrInvalidRequestType
	}
}

// validateCreateDomainRequest validates domain creation requests
func (v *Validator) ValidateCreateDomainRequest(req types.CreateDomainRequest) error {
	return v.ValidateName(req.Name)
}

// validateUpdateDomainRequest validates domain update requests
func (v *Validator) ValidateUpdateDomainRequest(req types.UpdateDomainRequest, user shared_types.User) error {
	// Validate ID first
	if err := v.ValidateID(req.ID); err != nil {
		return err
	}

	// Validate name
	if err := v.ValidateName(req.Name); err != nil {
		return err
	}

	// Check if domain exists
	domain, err := v.storage.GetDomain(req.ID)
	if err != nil {
		return err
	}
	if domain == nil {
		return types.ErrDomainNotFound
	}

	// Check permissions
	if user.Type != "admin" && domain.UserID.String() != user.ID.String() {
		return types.ErrNotAllowed
	}

	return nil
}

// validateDeleteDomainRequest validates domain deletion requests
func (v *Validator) ValidateDeleteDomainRequest(req types.DeleteDomainRequest, user shared_types.User) error {
	// Validate ID first
	if err := v.ValidateID(req.ID); err != nil {
		return err
	}

	// Check if domain exists
	domain, err := v.storage.GetDomain(req.ID)
	if err != nil {
		return err
	}
	if domain == nil {
		return types.ErrDomainNotFound
	}

	// Check permissions
	if user.Type != "admin" && domain.UserID.String() != user.ID.String() {
		return types.ErrNotAllowed
	}

	return nil
}
