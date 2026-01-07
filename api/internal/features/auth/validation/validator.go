package validation

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

const (
	MaxUserNameLength = 50
)

var (
	NumberPattern      = regexp.MustCompile(`[0-9]`)
	SpecialCharPattern = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`)
	UppercasePattern   = regexp.MustCompile(`[A-Z]`)
	LowercasePattern   = regexp.MustCompile(`[a-z]`)
	EmailPattern       = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateName(name string) error {
	switch {
	case name == "":
		return types.ErrMissingRequiredFields
	case utf8.RuneCountInString(name) > MaxUserNameLength:
		return types.ErrUserNameTooLong
	case strings.Contains(name, " "):
		return types.ErrUserNameContainsSpaces
	default:
		return nil
	}
}

func (v *Validator) IsValidPassword(password string) error {
	if password == "" {
		return types.ErrEmptyPassword
	}
	if len(password) < 8 {
		return types.ErrPasswordMustHaveAtLeast8Chars
	}
	if !NumberPattern.MatchString(password) {
		return types.ErrPasswordMustHaveAtLeast1Number
	}
	if !SpecialCharPattern.MatchString(password) {
		return types.ErrPasswordMustHaveAtLeast1SpecialChar
	}
	if !UppercasePattern.MatchString(password) {
		return types.ErrPasswordMustHaveAtLeast1UppercaseLetter
	}
	if !LowercasePattern.MatchString(password) {
		return types.ErrPasswordMustHaveAtLeast1LowercaseLetter
	}
	return nil
}

func (v *Validator) ValidateEmail(email string) error {
	if email == "" {
		return types.ErrMissingRequiredFields
	}

	if !EmailPattern.MatchString(email) {
		return types.ErrInvalidEmail
	}

	return nil
}

func (v *Validator) ParseRequestBody(req interface{}, body io.ReadCloser, decoded interface{}) error {
	return json.NewDecoder(body).Decode(decoded)
}

func (v *Validator) ValidateRequest(req interface{}) error {
	switch r := req.(type) {
	case *types.LoginRequest:
		return v.validateLoginRequest(*r)
	case *types.LogoutRequest:
		return v.validateLogoutRequest(*r)
	case *types.RefreshTokenRequest:
		return v.validateRefreshTokenRequest(*r)
	case *types.ResetPasswordRequest:
		return v.validateResetPasswordRequest(*r)
	case *types.RegisterRequest:
		return v.validateCreateUserRequest(*r)
	case *types.CreateAPIKeyRequest:
		return v.validateCreateAPIKeyRequest(*r)
	default:
		fmt.Printf("invalid request type: %T\n", req)
		return types.ErrInvalidRequestType
	}
}

func (v *Validator) validateLoginRequest(req types.LoginRequest) error {
	if err := v.ValidateEmail(req.Email); err != nil {
		return err
	}
	if err := v.IsValidPassword(req.Password); err != nil {
		return err
	}
	return nil
}

func (v *Validator) validateLogoutRequest(req types.LogoutRequest) error {
	if req.RefreshToken == "" {
		return types.ErrMissingRefreshToken
	}

	if _, err := uuid.Parse(req.RefreshToken); err != nil {
		return types.ErrInvalidRefreshToken
	}
	return nil
}

func (v *Validator) validateRefreshTokenRequest(req types.RefreshTokenRequest) error {
	if req.RefreshToken == "" {
		return types.ErrMissingRefreshToken
	}

	if _, err := uuid.Parse(req.RefreshToken); err != nil {
		return types.ErrInvalidRefreshToken
	}

	return nil
}

func (v *Validator) validateResetPasswordRequest(resetPasswordRequest types.ResetPasswordRequest) error {
	if resetPasswordRequest.Password == "" {
		return types.ErrMissingRequiredFields
	}

	if err := v.IsValidPassword(resetPasswordRequest.Password); err != nil {
		return err
	}

	return nil
}

func (v *Validator) validateCreateUserRequest(createUserRequest types.RegisterRequest) error {
	if err := v.ValidateEmail(createUserRequest.Email); err != nil {
		return err
	}
	if err := v.IsValidPassword(createUserRequest.Password); err != nil {
		return err
	}
	if err := v.ValidateName(createUserRequest.Username); err != nil {
		return err
	}

	if createUserRequest.Type != shared_types.RoleAdmin && createUserRequest.Type != shared_types.RoleMember && createUserRequest.Type != shared_types.RoleViewer {
		return types.ErrInvalidUserType
	}

	return nil
}

func (v *Validator) validateCreateAPIKeyRequest(req types.CreateAPIKeyRequest) error {
	if req.Name == "" {
		return types.ErrMissingRequiredFields
	}
	if utf8.RuneCountInString(req.Name) > 255 {
		return fmt.Errorf("name must be at most 255 characters")
	}
	if req.ExpiresInDays != nil {
		if *req.ExpiresInDays < 1 || *req.ExpiresInDays > 365 {
			return fmt.Errorf("expires_in_days must be between 1 and 365")
		}
	}
	return nil
}
