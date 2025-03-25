package tests

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/validation"
)

func TestValidateName(t *testing.T) {
	validator := validation.NewValidator()

	tests := []struct {
		name     string
		input    string
		expected error
	}{
		{"Empty name", "", types.ErrMissingRequiredFields},
		{"Valid name", "johndoe", nil},
		{"Name with spaces", "john doe", types.ErrUserNameContainsSpaces},
		{"Name too long", strings.Repeat("a", validation.MaxUserNameLength+1), types.ErrUserNameTooLong},
		{"Name exactly max length", strings.Repeat("a", validation.MaxUserNameLength), nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validator.ValidateName(test.input)
			if err != test.expected {
				t.Errorf("Expected error %v, got %v", test.expected, err)
			}
		})
	}
}

func TestIsValidPassword(t *testing.T) {
	validator := validation.NewValidator()

	tests := []struct {
		name     string
		input    string
		expected error
	}{
		{"Empty password", "", types.ErrEmptyPassword},
		{"Too short", "Aa1!", types.ErrPasswordMustHaveAtLeast8Chars},
		{"No numbers", "Abcdefg!", types.ErrPasswordMustHaveAtLeast1Number},
		{"No special chars", "Abcdefg1", types.ErrPasswordMustHaveAtLeast1SpecialChar},
		{"No uppercase", "abcdefg1!", types.ErrPasswordMustHaveAtLeast1UppercaseLetter},
		{"No lowercase", "ABCDEFG1!", types.ErrPasswordMustHaveAtLeast1LowercaseLetter},
		{"Valid password", "Abcdefg1!", nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validator.IsValidPassword(test.input)
			if err != test.expected {
				t.Errorf("Expected error %v, got %v", test.expected, err)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	validator := validation.NewValidator()

	tests := []struct {
		name     string
		input    string
		expected error
	}{
		{"Empty email", "", types.ErrMissingRequiredFields},
		{"No @ symbol", "johndoeexample.com", types.ErrInvalidEmail},
		{"No domain", "johndoe@", types.ErrInvalidEmail},
		{"No dot", "johndoe@examplecom", types.ErrInvalidEmail},
		{"Valid email", "johndoe@example.com", nil},
		{"Valid email with subdomain", "john.doe@sub.example.com", nil},
		{"Valid email with plus", "john+doe@example.com", nil},
		{"Valid email with numbers", "john123@example.com", nil},
		{"Invalid TLD length", "johndoe@example.c", types.ErrInvalidEmail},
		{"Invalid characters", "john@doe@example.com", types.ErrInvalidEmail},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validator.ValidateEmail(test.input)
			if err != test.expected {
				t.Errorf("Expected error %v, got %v", test.expected, err)
			}
		})
	}
}

func TestRegexPatterns(t *testing.T) {
	tests := []struct {
		name                   string
		input                  string
		hasNumber              bool
		hasSpecialChar         bool
		hasUppercaseLetter     bool
		hasLowercaseLetter     bool
	}{
		{"Empty string", "", false, false, false, false},
		{"Only numbers", "12345", true, false, false, false},
		{"Only special chars", "!@#$%", false, true, false, false},
		{"Only uppercase", "ABCDE", false, false, true, false},
		{"Only lowercase", "abcde", false, false, false, true},
		{"Mixed content", "Abc123!", true, true, true, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := validation.NumberPattern.MatchString(test.input); got != test.hasNumber {
				t.Errorf("numberPattern.MatchString() = %v, want %v", got, test.hasNumber)
			}
			if got := validation.SpecialCharPattern.MatchString(test.input); got != test.hasSpecialChar {
				t.Errorf("specialCharPattern.MatchString() = %v, want %v", got, test.hasSpecialChar)
			}
			if got := validation.UppercasePattern.MatchString(test.input); got != test.hasUppercaseLetter {
				t.Errorf("uppercasePattern.MatchString() = %v, want %v", got, test.hasUppercaseLetter)
			}
			if got := validation.LowercasePattern.MatchString(test.input); got != test.hasLowercaseLetter {
				t.Errorf("lowercasePattern.MatchString() = %v, want %v", got, test.hasLowercaseLetter)
			}
		})
	}
}

func TestParseRequestBody(t *testing.T) {
	validator := validation.NewValidator()
	
	// Test valid JSON
	validJSON := `{"email":"test@example.com","password":"Password1!"}`
	validReader := io.NopCloser(bytes.NewReader([]byte(validJSON)))
	
	var decoded types.LoginRequest
	err := validator.ParseRequestBody(nil, validReader, &decoded)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if decoded.Email != "test@example.com" || decoded.Password != "Password1!" {
		t.Errorf("Failed to decode JSON correctly")
	}
	
	// Test invalid JSON
	invalidJSON := `{"email":"test@example.com","password":"Password1!`
	invalidReader := io.NopCloser(bytes.NewReader([]byte(invalidJSON)))
	
	err = validator.ParseRequestBody(nil, invalidReader, &decoded)
	
	if err == nil {
		t.Errorf("Expected JSON decode error, got nil")
	}
}

func TestValidateRequest(t *testing.T) {
	validator := validation.NewValidator()
	
	// Test LoginRequest
	loginReq := &types.LoginRequest{
		Email:    "test@example.com",
		Password: "Password1!",
	}
	if err := validator.ValidateRequest(loginReq); err != nil {
		t.Errorf("Expected no error for valid login request, got %v", err)
	}
	
	// Test RegisterRequest
	registerReq := &types.RegisterRequest{
		Email:    "test@example.com",
		Password: "Password1!",
		Username: "testuser",
		Type:     "app_user",
	}
	if err := validator.ValidateRequest(registerReq); err != nil {
		t.Errorf("Expected no error for valid register request, got %v", err)
	}
	
	// Test LogoutRequest
	logoutReq := &types.LogoutRequest{
		RefreshToken: "validtoken",
	}
	if err := validator.ValidateRequest(logoutReq); err != nil {
		t.Errorf("Expected no error for valid logout request, got %v", err)
	}
	
	// Test RefreshTokenRequest
	refreshReq := &types.RefreshTokenRequest{
		RefreshToken: "validtoken",
	}
	if err := validator.ValidateRequest(refreshReq); err != nil {
		t.Errorf("Expected no error for valid refresh token request, got %v", err)
	}
	
	// Test ChangePasswordRequest
	changePassReq := &types.ChangePasswordRequest{
		OldPassword: "OldPass1!",
		NewPassword: "NewPass1!",
	}
	if err := validator.ValidateRequest(changePassReq); err != nil {
		t.Errorf("Expected no error for valid change password request, got %v", err)
	}
	
	// Test invalid request type
	invalidReq := &struct{}{}
	if err := validator.ValidateRequest(invalidReq); err != types.ErrInvalidRequestType {
		t.Errorf("Expected invalid request type error, got %v", err)
	}
}

func TestValidateLoginRequestWithInvalidData(t *testing.T) {
	validator := validation.NewValidator()
	
	// Missing email
	loginReq := &types.LoginRequest{
		Email:    "",
		Password: "Password1!",
	}
	if err := validator.ValidateRequest(loginReq); err != types.ErrMissingRequiredFields {
		t.Errorf("Expected missing email error, got %v", err)
	}
	
	// Invalid password
	loginReq = &types.LoginRequest{
		Email:    "test@example.com",
		Password: "pass", // Too short
	}
	if err := validator.ValidateRequest(loginReq); err != types.ErrPasswordMustHaveAtLeast8Chars {
		t.Errorf("Expected password too short error, got %v", err)
	}
}

func TestValidateRegisterRequestWithInvalidData(t *testing.T) {
	validator := validation.NewValidator()
	
	// Invalid user type
	registerReq := &types.RegisterRequest{
		Email:    "test@example.com",
		Password: "Password1!",
		Username: "testuser",
		Type:     "invalid_type",
	}
	if err := validator.ValidateRequest(registerReq); err != types.ErrInvalidUserType {
		t.Errorf("Expected invalid user type error, got %v", err)
	}
	
	// Default type
	registerReq = &types.RegisterRequest{
		Email:    "test@example.com",
		Password: "Password1!",
		Username: "testuser",
		Type:     "", // Empty should default to "app_user"
	}
	if err := validator.ValidateRequest(registerReq); err != nil {
		t.Errorf("Expected no error after defaulting type, got %v", err)
	}
	
	// Valid admin type
	registerReq = &types.RegisterRequest{
		Email:    "test@example.com",
		Password: "Password1!",
		Username: "testuser",
		Type:     "admin",
	}
	if err := validator.ValidateRequest(registerReq); err != nil {
		t.Errorf("Expected no error for admin type, got %v", err)
	}
}

func TestChangePasswordRequestValidation(t *testing.T) {
	validator := validation.NewValidator()
	
	// Same password
	changePassReq := &types.ChangePasswordRequest{
		OldPassword: "Password1!",
		NewPassword: "Password1!",
	}
	if err := validator.ValidateRequest(changePassReq); err != types.ErrSamePassword {
		t.Errorf("Expected same password error, got %v", err)
	}
	
	// Empty passwords
	changePassReq = &types.ChangePasswordRequest{
		OldPassword: "",
		NewPassword: "",
	}
	if err := validator.ValidateRequest(changePassReq); err != types.ErrEmptyPassword {
		t.Errorf("Expected empty password error, got %v", err)
	}
	
	// Invalid new password
	changePassReq = &types.ChangePasswordRequest{
		OldPassword: "OldPassword1!",
		NewPassword: "weak", // Too short, missing uppercase, number, special char
	}
	if err := validator.ValidateRequest(changePassReq); err != types.ErrPasswordMustHaveAtLeast8Chars {
		t.Errorf("Expected password validation error, got %v", err)
	}
	
	// Valid passwords
	changePassReq = &types.ChangePasswordRequest{
		OldPassword: "OldPassword1!",
		NewPassword: "NewPassword2@",
	}
	if err := validator.ValidateRequest(changePassReq); err != nil {
		t.Errorf("Expected no error for valid passwords, got %v", err)
	}
}