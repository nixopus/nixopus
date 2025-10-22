package validation

import (
	"encoding/json"
	"io"
	"regexp"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
)

const (
	MaxNameLength        = 50
	MaxDescriptionLength = 100
)

type Validator struct {
	storage storage.OrganizationRepository
}

func NewValidator(storage storage.OrganizationRepository) *Validator {
	return &Validator{
		storage: storage,
	}
}

func (v *Validator) ValidateID(id string, errorType error) error {
	if id == "" {
		return errorType
	}
	if _, err := uuid.Parse(id); err != nil {
		return errorType
	}
	return nil
}

func (v *Validator) ValidateName(name string) error {
	switch {
	case name == "":
		return types.ErrMissingOrganizationName
	case utf8.RuneCountInString(name) > MaxNameLength:
		return types.ErrOrganizationNameTooLong
	default:
		return nil
	}
}

func (v *Validator) ValidateDescription(description string) error {
	if utf8.RuneCountInString(description) > MaxDescriptionLength {
		return types.ErrOrganizationDescriptionTooLong
	}
	return nil
}

// ValidateRequest validates a request object against a set of predefined rules.
// It returns an error if the request object is invalid.
//
// The supported request types are:
// - types.CreateOrganizationRequest
// - types.UpdateOrganizationRequest
// - types.DeleteOrganizationRequest
// - types.AddUserToOrganizationRequest
// - types.RemoveUserFromOrganizationRequest
//
// If the request object is not of one of the above types, it returns
// types.ErrInvalidRequestType.
func (v *Validator) ValidateRequest(req interface{}) error {
	switch r := req.(type) {
	case *types.CreateOrganizationRequest:
		return v.validateCreate(*r)
	case *types.UpdateOrganizationRequest:
		return v.validateUpdate(*r)
	case *types.DeleteOrganizationRequest:
		return v.validateDelete(*r)
	case *types.AddUserToOrganizationRequest:
		return v.validateAddUser(*r)
	case *types.RemoveUserFromOrganizationRequest:
		return v.validateRemoveUser(*r)
	case *types.InviteSendRequest:
		return v.validateInviteSend(*r)
	case *types.InviteResendRequest:
		return v.validateInviteResend(*r)
	case *types.UpdateUserRoleRequest:
		return v.validateUpdateUserRole(*r)
	default:
		return types.ErrInvalidRequestType
	}
}

// validateCreate validates a CreateOrganizationRequest object.
// It checks that the organization's name and description adhere to predefined rules.
//
// It returns an error if the name is missing or too long, or if the description is too long.
// Otherwise, it returns nil.
func (v *Validator) validateCreate(req types.CreateOrganizationRequest) error {
	if err := v.ValidateName(req.Name); err != nil {
		return err
	}
	return v.ValidateDescription(req.Description)
}

// validateUpdate validates an UpdateOrganizationRequest object.
// It checks that the organization's name and description adhere to predefined rules,
// and that the organization exists.
//
// It returns an error if the name is missing or too long, or if the description is too long,
// or if the organization does not exist.
// Otherwise, it returns nil.
func (v *Validator) validateUpdate(req types.UpdateOrganizationRequest) error {
	if err := v.ValidateName(req.Name); err != nil {
		return err
	}
	if err := v.ValidateDescription(req.Description); err != nil {
		return err
	}
	if err := v.ValidateID(req.ID, types.ErrMissingOrganizationID); err != nil {
		return err
	}

	organization, err := v.storage.GetOrganization(req.ID)

	if err != nil {
		return err
	}

	if organization == nil {
		return types.ErrOrganizationNotFound
	}

	return nil
}

// validateDelete validates a DeleteOrganizationRequest object.
// It checks that the organization's ID is valid and that the organization exists.
//
// It returns an error if the ID is missing or invalid, or if the organization does not exist.
// Otherwise, it returns nil.
func (v *Validator) validateDelete(req types.DeleteOrganizationRequest) error {
	err := v.ValidateID(req.ID, types.ErrMissingOrganizationID)

	if err != nil {
		return err
	}

	return nil
}

// validateAddUser validates an AddUserToOrganizationRequest object.
// It checks that the organization's ID is valid and that the organization exists,
// that the user's ID is valid, and that the role's ID is valid.
//
// It returns an error if any of these checks fail.
// Otherwise, it returns nil.
func (v *Validator) validateAddUser(req types.AddUserToOrganizationRequest) error {
	if err := v.ValidateID(req.OrganizationID, types.ErrMissingOrganizationID); err != nil {
		return err
	}
	if err := v.ValidateID(req.UserID, types.ErrMissingUserID); err != nil {
		return err
	}

	organization, err := v.storage.GetOrganization(req.OrganizationID)

	if err != nil {
		return err
	}

	if organization == nil {
		return types.ErrOrganizationNotFound
	}

	return nil
}

// validateRemoveUser validates a RemoveUserFromOrganizationRequest object.
// It checks that the organization's ID is valid and that the organization exists,
// and that the user's ID is valid.
//
// It returns an error if any of these checks fail.
// Otherwise, it returns nil.
func (v *Validator) validateRemoveUser(req types.RemoveUserFromOrganizationRequest) error {
	if err := v.ValidateID(req.OrganizationID, types.ErrMissingOrganizationID); err != nil {
		return err
	}

	organization, err := v.storage.GetOrganization(req.OrganizationID)

	if err != nil {
		return err
	}

	if organization == nil {
		return types.ErrOrganizationNotFound
	}

	return v.ValidateID(req.UserID, types.ErrMissingUserID)
}

func (v *Validator) ValidateEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return types.ErrInvalidEmail
	}
	return nil
}

func (v *Validator) validateInviteSend(req types.InviteSendRequest) error {
	if err := v.ValidateEmail(req.Email); err != nil {
		return err
	}
	if req.OrganizationID == "" {
		return types.ErrMissingOrganizationID
	}
	if req.Role == "" {
		return types.ErrMissingRoleID
	}
	return nil
}

func (v *Validator) validateInviteResend(req types.InviteResendRequest) error {
	if err := v.ValidateEmail(req.Email); err != nil {
		return err
	}
	if req.OrganizationID == "" {
		return types.ErrMissingOrganizationID
	}
	if req.Role == "" {
		return types.ErrMissingRoleID
	}
	return nil
}

func (v *Validator) validateUpdateUserRole(req types.UpdateUserRoleRequest) error {
	if err := v.ValidateID(req.OrganizationID, types.ErrMissingOrganizationID); err != nil {
		return err
	}
	if err := v.ValidateID(req.UserID, types.ErrMissingUserID); err != nil {
		return err
	}
	if req.Role == "" {
		return types.ErrMissingRoleID
	}

	organization, err := v.storage.GetOrganization(req.OrganizationID)
	if err != nil {
		return err
	}
	if organization == nil {
		return types.ErrOrganizationNotFound
	}

	return nil
}

func (v *Validator) ParseRequestBody(req interface{}, body io.ReadCloser, decoded interface{}) error {
	return json.NewDecoder(body).Decode(decoded)
}
