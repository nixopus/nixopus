package validation

import (
	"encoding/json"
	"io"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type Validator struct {
	storage storage.DomainStorageInterface
}

func NewValidator(storage storage.DomainStorageInterface) *Validator {
	return &Validator{
		storage: storage,
	}
}

// ValidateID validates the domain ID.
//
// It returns an error if the domain ID is empty or not a valid UUID.
func (v *Validator) ValidateID(id string) error {
	if id == "" {
		return types.ErrMissingDomainID
	}
	if _, err := uuid.Parse(id); err != nil {
		return types.ErrInvalidDomainID
	}
	return nil
}

// ValidateName validates the domain name.
//
// It returns an error if the domain name is empty.
func (v *Validator) ValidateName(domain string) error {
	if domain == "" {
		return types.ErrMissingDomainName
	}

	return nil
}

func (v *Validator) ParseRequestBody(req interface{}, body io.ReadCloser, decoded interface{}) error {
	return json.NewDecoder(body).Decode(decoded)
}

// ValidateRequest validates a request object against a set of predefined rules.
// It returns an error if the request object is invalid.
//
// The supported request types are:
// - types.CreateDomainRequest
// - types.UpdateDomainRequest
// - types.DeleteDomainRequest
//
// If the request object is not of one of the above types, it returns
// types.ErrInvalidRequestType.
func (v *Validator) ValidateRequest(req interface{}, user shared_types.User) error {
	switch r := req.(type) {
	case *types.CreateDomainRequest:
		return v.validateCreateDomainRequest(*r)
	case *types.UpdateDomainRequest:
		return v.validateUpdateDomainRequest(*r, user)
	case *types.DeleteDomainRequest:
		return v.validateDeleteDomainRequest(*r, user)
	default:
		return types.ErrInvalidRequestType
	}
}

// validateCreateDomainRequest validates a request object against a set of predefined rules.
// It returns an error if the request object is invalid.
//
// The rules are:
// - The name should not be empty
// - The name should be between 3 and 255 characters long
func (v *Validator) validateCreateDomainRequest(req types.CreateDomainRequest) error {
	if err := v.ValidateName(req.Name); err != nil {
		return err
	}

	if len(req.Name) > 255 {
		return types.ErrDomainNameTooLong
	}

	if len(req.Name) < 3 {
		return types.ErrDomainNameTooShort
	}

	return nil
}

// validateUpdateDomainRequest validates a request object against a set of predefined rules.
// It returns an error if the request object is invalid.
//
// The rules are:
// - The name should not be empty
// - The ID should be a valid UUID
// - The ID should correspond to an existing domain
// - The user should be the owner of the domain or an admin
func (v *Validator) validateUpdateDomainRequest(req types.UpdateDomainRequest, user shared_types.User) error {
	domain, err := v.storage.GetDomain(req.ID)
	if err != nil {
		return err
	}
	if domain == nil {
		return types.ErrDomainNotFound
	}

	if user.Type != "admin" && domain.UserID.String() != user.ID.String() {
		return types.ErrNotAllowed
	}

	if err := v.ValidateName(req.Name); err != nil {
		return err
	}

	return nil
}

// validateDeleteDomainRequest validates the delete domain request.
//
// It checks if the domain with the given ID exists in the storage.
// If the domain does not exist, it returns ErrDomainNotFound.
// It also checks if the requesting user is either an admin or the owner of the domain.
// If the user is not authorized to delete the domain, it returns ErrNotAllowed.
//
// It returns an error if any check fails, otherwise it returns nil.
func (v *Validator) validateDeleteDomainRequest(req types.DeleteDomainRequest, user shared_types.User) error {
	domain, err := v.storage.GetDomain(req.ID)
	if err != nil {
		return err
	}
	if domain == nil {
		return types.ErrDomainNotFound
	}

	if user.Type != "admin" && domain.UserID.String() != user.ID.String() {
		return types.ErrNotAllowed
	}

	return nil
}
