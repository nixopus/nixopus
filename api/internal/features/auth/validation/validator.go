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

// ValidateRequest validates request body - only API key requests are supported
// Better Auth handles authentication requests
func (v *Validator) ValidateRequest(req interface{}) error {
	switch r := req.(type) {
	case *types.CreateAPIKeyRequest:
		return v.validateCreateAPIKeyRequest(*r)
	default:
		fmt.Printf("invalid request type: %T\n", req)
		return types.ErrInvalidRequestType
	}
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
