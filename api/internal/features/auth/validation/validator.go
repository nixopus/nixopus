package validation

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
)

const (
	MaxUserNameLength = 50
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
	if !containsNumber(password) {
		return types.ErrPasswordMustHaveAtLeast1Number
	}
	if !containsSpecialChar(password) {
		return types.ErrPasswordMustHaveAtLeast1SpecialChar
	}
	if !containsUppercaseLetter(password) {
		return types.ErrPasswordMustHaveAtLeast1UppercaseLetter
	}
	if !containsLowercaseLetter(password) {
		return types.ErrPasswordMustHaveAtLeast1LowercaseLetter
	}
	return nil
}

func (v *Validator) ValidateEmail(email string) error {
	if email == "" {
		return types.ErrMissingRequiredFields
	}

	if !strings.Contains(email, "@") {
		return types.ErrInvalidEmail
	}

	if !strings.Contains(email, ".") {
		return types.ErrInvalidEmail
	}

	return nil
}

func containsNumber(password string) bool {
	for _, char := range password {
		if char >= '0' && char <= '9' {
			return true
		}
	}
	return false
}

func containsSpecialChar(password string) bool {
	for _, char := range password {
		if char >= '!' && char <= '/' || char >= ':' && char <= '@' || char >= '[' && char <= '`' || char >= '{' && char <= '~' {
			return true
		}
	}
	return false
}

func containsUppercaseLetter(password string) bool {
	for _, char := range password {
		if char >= 'A' && char <= 'Z' {
			return true
		}
	}
	return false
}

func containsLowercaseLetter(password string) bool {
	for _, char := range password {
		if char >= 'a' && char <= 'z' {
			return true
		}
	}
	return false
}

func (v *Validator) ParseRequestBody(req interface{}, body io.ReadCloser, decoded interface{}) error {
	return json.NewDecoder(body).Decode(decoded)
}

func (v *Validator) ValidateRequest(req interface{}) error {
	switch r := req.(type) {
	case *types.LoginRequest:
		return v.validateLoginRequest(*r)
	case *types.RegisterRequest:
		return v.validateRegisterRequest(*r)
	case *types.LogoutRequest:
		return v.validateLogoutRequest(*r)
	case *types.RefreshTokenRequest:
		return v.validateRefreshTokenRequest(*r)
	case *types.ChangePasswordRequest:
		return v.validateResetPasswordRequest(*r)
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

func (v *Validator) validateRegisterRequest(req types.RegisterRequest) error {
	if err := v.ValidateEmail(req.Email); err != nil {
		return err
	}
	if err := v.IsValidPassword(req.Password); err != nil {
		return err
	}
	if err := v.ValidateName(req.Username); err != nil {
		return err
	}
	if req.Type == "" {
		req.Type = "app_user"
	}

	if req.Type != "app_user" && req.Type != "admin" {
		return types.ErrInvalidUserType
	}

	return nil
}

func (v *Validator) validateLogoutRequest(req types.LogoutRequest) error {
	if req.RefreshToken == "" {
		return types.ErrMissingRefreshToken
	}
	return nil
}

func (v *Validator) validateRefreshTokenRequest(req types.RefreshTokenRequest) error {
	if req.RefreshToken == "" {
		return types.ErrMissingRefreshToken
	}
	return nil
}

func (v *Validator) validateResetPasswordRequest(reset_password_request types.ChangePasswordRequest) error {
	if reset_password_request.NewPassword == "" || reset_password_request.OldPassword == "" {
		return types.ErrEmptyPassword
	}

	if reset_password_request.NewPassword == reset_password_request.OldPassword {
		return types.ErrSamePassword
	}

	return nil
}
