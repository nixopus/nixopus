package validation

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
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

	userType := req.Type
	if userType == "" {
		userType = "app_user"
	}

	if userType != "app_user" && userType != "admin" {
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

func (v *Validator) validateResetPasswordRequest(resetPasswordRequest types.ChangePasswordRequest) error {
	if resetPasswordRequest.NewPassword == "" || resetPasswordRequest.OldPassword == "" {
		return types.ErrEmptyPassword
	}

	if resetPasswordRequest.NewPassword == resetPasswordRequest.OldPassword {
		return types.ErrSamePassword
	}

	return v.IsValidPassword(resetPasswordRequest.NewPassword)
}
