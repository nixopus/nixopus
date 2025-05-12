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
	case *types.UpdateAvatarRequest:
		return v.ValidateUpdateAvatarRequest(*r)
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

// validateUpdateAvatarRequest checks if the avatar data is valid
func (v *Validator) ValidateUpdateAvatarRequest(req types.UpdateAvatarRequest) error {
	if req.AvatarData == "" {
		return types.ErrInvalidAvatarData
	}

	if !strings.HasPrefix(req.AvatarData, "data:image/") {
		return types.ErrInvalidAvatarData
	}

	parts := strings.Split(req.AvatarData, ";base64,")
	if len(parts) != 2 {
		return types.ErrInvalidAvatarData
	}

	imageType := strings.TrimPrefix(parts[0], "data:image/")
	if !isValidImageType(imageType) {
		return types.ErrUnsupportedImageFormat
	}

	return nil
}

func isValidImageType(imageType string) bool {
	validTypes := map[string]bool{
		"jpeg": true,
		"jpg":  true,
		"png":  true,
		"gif":  true,
	}
	return validTypes[imageType]
}
