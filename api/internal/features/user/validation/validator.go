package validation

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/features/user/types"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ParseRequestBody(req interface{}, body io.ReadCloser, decoded interface{}) error {
	return json.NewDecoder(body).Decode(decoded)
}

// ValidateRequest checks if the request is valid
func (v *Validator) ValidateRequest(req interface{}, user shared_types.User) error {
	switch r := req.(type) {
	case *types.UpdateUserNameRequest:
		return v.ValidateUpdateUserNameRequest(*r, user)
	default:
		return types.ErrInvalidRequestType
	}
}

// validateUpdateUserNameRequest checks if the username satisfies the requirements
func (v *Validator) ValidateUpdateUserNameRequest(req types.UpdateUserNameRequest, user shared_types.User) error {
	if req.Name == "" {
		return types.ErrUserNameIsEmpty
	}
	if req.Name == user.Username {
		return types.ErrSameUserName
	}

	if len(req.Name) > 50 {
		return types.ErrUserNameTooLong
	}

	if strings.Contains(req.Name, " ") {
		return types.ErrUserNameContainsSpaces
	}

	if len(req.Name) < 3 {
		return types.ErrUsernameTooShort
	}

	return nil
}
