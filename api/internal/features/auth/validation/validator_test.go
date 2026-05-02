package validation

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nixopus/nixopus/api/internal/features/auth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator()
	require.NotNil(t, v)
}

func TestValidator_ValidateName(t *testing.T) {
	v := NewValidator()

	assert.NoError(t, v.ValidateName("alice"))
	require.ErrorIs(t, v.ValidateName(""), types.ErrMissingRequiredFields)
	long := strings.Repeat("a", MaxUserNameLength+1)
	require.ErrorIs(t, v.ValidateName(long), types.ErrUserNameTooLong)
	require.ErrorIs(t, v.ValidateName("x y"), types.ErrUserNameContainsSpaces)
}

func TestValidator_IsValidPassword(t *testing.T) {
	v := NewValidator()

	require.ErrorIs(t, v.IsValidPassword(""), types.ErrEmptyPassword)
	require.ErrorIs(t, v.IsValidPassword("short1!"), types.ErrPasswordMustHaveAtLeast8Chars)
	require.ErrorIs(t, v.IsValidPassword("NoNumber!abcd"), types.ErrPasswordMustHaveAtLeast1Number)
	require.ErrorIs(t, v.IsValidPassword("NoSpecial1Aa"), types.ErrPasswordMustHaveAtLeast1SpecialChar)
	require.ErrorIs(t, v.IsValidPassword("noupper1!aabcdef"), types.ErrPasswordMustHaveAtLeast1UppercaseLetter)
	require.ErrorIs(t, v.IsValidPassword("NOLOWER123!ABCDEF"), types.ErrPasswordMustHaveAtLeast1LowercaseLetter)
	require.NoError(t, v.IsValidPassword("Good1Pass!"))
}

func TestValidator_ValidateEmail(t *testing.T) {
	v := NewValidator()

	require.ErrorIs(t, v.ValidateEmail(""), types.ErrMissingRequiredFields)
	require.ErrorIs(t, v.ValidateEmail("bad"), types.ErrInvalidEmail)
	require.NoError(t, v.ValidateEmail("ok@domain.co"))
}

func TestValidator_ParseRequestBody(t *testing.T) {
	v := NewValidator()
	var out struct {
		K string `json:"k"`
	}
	err := v.ParseRequestBody(nil, ioNopCloser(`{"k":"v"}`), &out)
	require.NoError(t, err)
	require.Equal(t, "v", out.K)
}

type readClose struct {
	*bytes.Reader
}

func (readClose) Close() error { return nil }

func ioNopCloser(s string) readClose {
	return readClose{Reader: bytes.NewReader([]byte(s))}
}

func TestValidator_ValidateRequest(t *testing.T) {
	v := NewValidator()
	require.ErrorIs(t, v.ValidateRequest(struct{}{}), types.ErrInvalidRequestType)
}
