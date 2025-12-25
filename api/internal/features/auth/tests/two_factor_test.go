package tests

import (
	"testing"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/xlzd/gotp"
)

func TestSetupTwoFactor(t *testing.T) {
	setup := testutils.NewTestSetup()

	registerRequest := types.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Username: "testuser",
		Type:     "viewer",
	}

	registerResponse, err := setup.AuthService.Register(registerRequest, "app_user")
	assert.NoError(t, err)

	response, err := setup.AuthService.SetupTwoFactor(&registerResponse.User)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.Secret)
	assert.NotEmpty(t, response.QRCode)

	user, err := setup.AuthService.GetUserByID(registerResponse.User.ID.String())
	assert.NoError(t, err)
	assert.NotEmpty(t, user.TwoFactorSecret)
	assert.False(t, user.TwoFactorEnabled)
}

func TestVerifyTwoFactor(t *testing.T) {
	setup := testutils.NewTestSetup()

	registerRequest := types.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Username: "testuser",
		Type:     "viewer",
	}

	registerResponse, err := setup.AuthService.Register(registerRequest, "app_user")
	assert.NoError(t, err)

	setupResponse, err := setup.AuthService.SetupTwoFactor(&registerResponse.User)
	assert.NoError(t, err)

	totp := gotp.NewDefaultTOTP(setupResponse.Secret)
	code := totp.Now()

	err = setup.AuthService.VerifyTwoFactor(&registerResponse.User, code)
	assert.NoError(t, err)

	user, err := setup.AuthService.GetUserByID(registerResponse.User.ID.String())
	assert.NoError(t, err)
	assert.True(t, user.TwoFactorEnabled)
}

func TestVerifyTwoFactorInvalidCode(t *testing.T) {
	setup := testutils.NewTestSetup()

	registerRequest := types.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Username: "testuser",
		Type:     "viewer",
	}

	registerResponse, err := setup.AuthService.Register(registerRequest, "app_user")
	assert.NoError(t, err)

	_, err = setup.AuthService.SetupTwoFactor(&registerResponse.User)
	assert.NoError(t, err)

	err = setup.AuthService.VerifyTwoFactor(&registerResponse.User, "invalid_code")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), types.ErrInvalid2FACode.Error())
}

func TestDisableTwoFactor(t *testing.T) {
	setup := testutils.NewTestSetup()

	registerRequest := types.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Username: "testuser",
		Type:     "viewer",
	}

	registerResponse, err := setup.AuthService.Register(registerRequest, "app_user")
	assert.NoError(t, err)

	setupResponse, err := setup.AuthService.SetupTwoFactor(&registerResponse.User)
	assert.NoError(t, err)

	totp := gotp.NewDefaultTOTP(setupResponse.Secret)
	code := totp.Now()

	err = setup.AuthService.VerifyTwoFactor(&registerResponse.User, code)
	assert.NoError(t, err)

	err = setup.AuthService.DisableTwoFactor(&registerResponse.User)
	assert.NoError(t, err)

	user, err := setup.AuthService.GetUserByID(registerResponse.User.ID.String())
	assert.NoError(t, err)
	assert.False(t, user.TwoFactorEnabled)
	assert.Empty(t, user.TwoFactorSecret)
}

func TestVerifyTwoFactorCode(t *testing.T) {
	setup := testutils.NewTestSetup()

	registerRequest := types.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Username: "testuser",
		Type:     "viewer",
	}

	registerResponse, err := setup.AuthService.Register(registerRequest, "app_user")
	assert.NoError(t, err)

	setupResponse, err := setup.AuthService.SetupTwoFactor(&registerResponse.User)
	assert.NoError(t, err)

	totp := gotp.NewDefaultTOTP(setupResponse.Secret)
	code := totp.Now()

	err = setup.AuthService.VerifyTwoFactor(&registerResponse.User, code)
	assert.NoError(t, err)

	err = setup.AuthService.VerifyTwoFactorCode(&registerResponse.User, code)
	assert.NoError(t, err)
}

func TestVerifyTwoFactorCodeInvalid(t *testing.T) {
	setup := testutils.NewTestSetup()

	registerRequest := types.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Username: "testuser",
		Type:     "viewer",
	}

	registerResponse, err := setup.AuthService.Register(registerRequest, "app_user")
	assert.NoError(t, err)

	setupResponse, err := setup.AuthService.SetupTwoFactor(&registerResponse.User)
	assert.NoError(t, err)

	totp := gotp.NewDefaultTOTP(setupResponse.Secret)
	code := totp.Now()

	err = setup.AuthService.VerifyTwoFactor(&registerResponse.User, code)
	assert.NoError(t, err)

	err = setup.AuthService.VerifyTwoFactorCode(&registerResponse.User, "invalid_code")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), types.ErrInvalid2FACode.Error())
}
