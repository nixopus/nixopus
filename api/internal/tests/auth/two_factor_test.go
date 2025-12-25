package auth

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/tests"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
	"github.com/xlzd/gotp"
)

func TestSetupTwoFactor(t *testing.T) {
	setup := testutils.NewTestSetup()
	authResponse, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	testCases := []struct {
		name           string
		token          string
		expectedStatus int
		orgID          string
		expectSecret   bool
	}{
		{
			name:           "Successfully setup 2FA",
			token:          authResponse.AccessToken,
			expectedStatus: http.StatusOK,
			orgID:          org.ID.String(),
			expectSecret:   true,
		},
		{
			name:           "Setup 2FA without token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			orgID:          org.ID.String(),
			expectSecret:   false,
		},
		{
			name:           "Setup 2FA with invalid token",
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
			orgID:          org.ID.String(),
			expectSecret:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			steps := []IStep{
				Description(tc.name),
				Post(tests.GetSetup2FAURL()),
				Send().Headers("Authorization").Add("Bearer " + tc.token),
				Send().Headers("X-Organization-Id").Add(tc.orgID),
				Expect().Status().Equal(int64(tc.expectedStatus)),
			}

			if tc.expectSecret {
				steps = append(steps,
					Expect().Body().JSON().JQ(".data.secret").NotEqual(""),
					Expect().Body().JSON().JQ(".data.qr_code").NotEqual(""),
					Expect().Body().JSON().JQ(".status").Equal("success"),
					Expect().Body().JSON().JQ(".message").NotEqual(""),
				)
			} else {
				steps = append(steps,
					Expect().Body().JSON().JQ(".error").NotEqual(""),
					Expect().Body().JSON().JQ(".status").Equal("error"),
				)
			}

			Test(t, steps...)
		})
	}
}

func TestVerifyTwoFactor(t *testing.T) {
	setup := testutils.NewTestSetup()
	authResponse, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	Test(t,
		Post(tests.GetSetup2FAURL()),
		Send().Headers("Authorization").Add("Bearer "+authResponse.AccessToken),
		Send().Headers("X-Organization-Id").Add(org.ID.String()),
		Expect().Status().Equal(http.StatusOK),
	)

	testCases := []struct {
		name           string
		token          string
		orgID          string
		request        types.TwoFactorVerifyRequest
		expectedStatus int
	}{
		{
			name:  "Invalid verification code",
			token: authResponse.AccessToken,
			orgID: org.ID.String(),
			request: types.TwoFactorVerifyRequest{
				Code: "123456",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "Empty verification code",
			token: authResponse.AccessToken,
			orgID: org.ID.String(),
			request: types.TwoFactorVerifyRequest{
				Code: "",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "Invalid code format",
			token: authResponse.AccessToken,
			orgID: org.ID.String(),
			request: types.TwoFactorVerifyRequest{
				Code: "12345",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "Non-numeric code",
			token: authResponse.AccessToken,
			orgID: org.ID.String(),
			request: types.TwoFactorVerifyRequest{
				Code: "abcdef",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "Verify without token",
			token: "",
			orgID: org.ID.String(),
			request: types.TwoFactorVerifyRequest{
				Code: "123456",
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			Test(t,
				Description(tc.name),
				Post(tests.GetVerify2FAURL()),
				Send().Headers("Authorization").Add("Bearer "+tc.token),
				Send().Headers("X-Organization-Id").Add(tc.orgID),
				Send().Body().JSON(tc.request),
				Expect().Status().Equal(int64(tc.expectedStatus)),
			)
		})
	}
}

func TestDisableTwoFactor(t *testing.T) {
	setup := testutils.NewTestSetup()
	authResponse, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	testCases := []struct {
		name           string
		token          string
		orgID          string
		expectedStatus int
	}{
		{
			name:           "Successfully disable 2FA",
			token:          authResponse.AccessToken,
			orgID:          org.ID.String(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Disable without token",
			token:          "",
			orgID:          org.ID.String(),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Disable with invalid token",
			token:          "invalid-token",
			orgID:          org.ID.String(),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			Test(t,
				Description(tc.name),
				Post(tests.GetDisable2FAURL()),
				Send().Headers("Authorization").Add("Bearer "+tc.token),
				Send().Headers("X-Organization-Id").Add(tc.orgID),
				Expect().Status().Equal(int64(tc.expectedStatus)),
			)
		})
	}
}

func TestTwoFactorLogin(t *testing.T) {
	setup := testutils.NewTestSetup()
	authResponse, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	// Get the secret from the response
	var secret string
	Test(t,
		Post(tests.GetSetup2FAURL()),
		Send().Headers("Authorization").Add("Bearer "+authResponse.AccessToken),
		Send().Headers("X-Organization-Id").Add(org.ID.String()),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".data.secret").NotEqual(""),
		Store().Response().Body().JSON().JQ(".data.secret").In(&secret),
	)

	// Verify the 2FA setup
	totp := gotp.NewDefaultTOTP(secret)
	verifyCode := totp.Now()

	Test(t,
		Post(tests.GetVerify2FAURL()),
		Send().Headers("Authorization").Add("Bearer "+authResponse.AccessToken),
		Send().Headers("X-Organization-Id").Add(org.ID.String()),
		Send().Body().JSON(types.TwoFactorVerifyRequest{
			Code: verifyCode,
		}),
		Expect().Status().Equal(http.StatusOK),
	)

	// Try to login with normal login and store the temporary tokens
	var tempToken string
	Test(t,
		Post(tests.GetLoginURL()),
		Send().Body().JSON(types.LoginRequest{
			Email:    authResponse.User.Email,
			Password: "Password123@",
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".data.temp_token").NotEqual(""),
		Store().Response().Body().JSON().JQ(".data.temp_token").In(&tempToken),
	)

	// Generate TOTP code using the secret
	loginCode := totp.Now()

	Test(t,
		Post(tests.Get2FALoginURL()),
		Send().Headers("Authorization").Add("Bearer "+tempToken),
		Send().Headers("X-Organization-Id").Add(org.ID.String()),
		Send().Body().JSON(types.TwoFactorLoginRequest{
			Email:    authResponse.User.Email,
			Password: "Password123@",
			Code:     loginCode,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".data.access_token").NotEqual(""),
		Expect().Body().JSON().JQ(".data.refresh_token").NotEqual(""),
		Expect().Body().JSON().JQ(".data.user.email").Equal(authResponse.User.Email),
	)
}

func TestTwoFactorLoginEdgeCases(t *testing.T) {
	setup := testutils.NewTestSetup()
	authResponse, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	// Setup and verify 2FA first
	var secret string
	Test(t,
		Post(tests.GetSetup2FAURL()),
		Send().Headers("Authorization").Add("Bearer "+authResponse.AccessToken),
		Send().Headers("X-Organization-Id").Add(org.ID.String()),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".data.secret").NotEqual(""),
		Store().Response().Body().JSON().JQ(".data.secret").In(&secret),
	)

	totp := gotp.NewDefaultTOTP(secret)
	verifyCode := totp.Now()

	Test(t,
		Post(tests.GetVerify2FAURL()),
		Send().Headers("Authorization").Add("Bearer "+authResponse.AccessToken),
		Send().Headers("X-Organization-Id").Add(org.ID.String()),
		Send().Body().JSON(types.TwoFactorVerifyRequest{
			Code: verifyCode,
		}),
		Expect().Status().Equal(http.StatusOK),
	)

	// Get temp token for all test cases
	var tempToken string
	Test(t,
		Post(tests.GetLoginURL()),
		Send().Body().JSON(types.LoginRequest{
			Email:    authResponse.User.Email,
			Password: "Password123@",
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".data.temp_token").NotEqual(""),
		Store().Response().Body().JSON().JQ(".data.temp_token").In(&tempToken),
	)

	testCases := []struct {
		name           string
		token          string
		orgID          string
		request        types.TwoFactorLoginRequest
		expectedStatus int
	}{
		{
			name:  "Invalid TOTP code",
			token: tempToken,
			orgID: org.ID.String(),
			request: types.TwoFactorLoginRequest{
				Email:    authResponse.User.Email,
				Password: "Password123@",
				Code:     "123456",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:  "Empty TOTP code",
			token: tempToken,
			orgID: org.ID.String(),
			request: types.TwoFactorLoginRequest{
				Email:    authResponse.User.Email,
				Password: "Password123@",
				Code:     "",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:  "Invalid password",
			token: tempToken,
			orgID: org.ID.String(),
			request: types.TwoFactorLoginRequest{
				Email:    authResponse.User.Email,
				Password: "WrongPassword123@",
				Code:     totp.Now(),
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:  "Invalid email",
			token: tempToken,
			orgID: org.ID.String(),
			request: types.TwoFactorLoginRequest{
				Email:    "wrong@example.com",
				Password: "Password123@",
				Code:     totp.Now(),
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:  "Missing authorization token",
			token: "",
			orgID: org.ID.String(),
			request: types.TwoFactorLoginRequest{
				Email:    authResponse.User.Email,
				Password: "Password123@",
				Code:     totp.Now(),
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:  "Invalid authorization token",
			token: "invalid-token",
			orgID: org.ID.String(),
			request: types.TwoFactorLoginRequest{
				Email:    authResponse.User.Email,
				Password: "Password123@",
				Code:     totp.Now(),
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:  "Expired TOTP code",
			token: tempToken,
			orgID: org.ID.String(),
			request: types.TwoFactorLoginRequest{
				Email:    authResponse.User.Email,
				Password: "Password123@",
				Code:     "000000",
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			Test(t,
				Description(tc.name),
				Post(tests.Get2FALoginURL()),
				Send().Headers("Authorization").Add("Bearer "+tc.token),
				Send().Headers("X-Organization-Id").Add(tc.orgID),
				Send().Body().JSON(tc.request),
				Expect().Status().Equal(int64(tc.expectedStatus)),
			)
		})
	}
}
