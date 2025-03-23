package validation

import (
	"encoding/json"
	"io"
	"net/http"
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

// AccessValidator checks if the user has access to the requested resource
func (v *Validator) AccessValidator(w http.ResponseWriter, r *http.Request, user *shared_types.User) error {
	path := r.URL.Path
	// Allow access to /api/v1/user, /api/v1/user/name, and /api/v1/user/organizations (endpoints for updating user name and getting user organizations, and accessing user's own details)
	if path == "/api/v1/user" || path == "/api/v1/user/name" || path == "/api/v1/user/organizations" {
		return nil
	}
	return types.ErrInvalidAccess
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
