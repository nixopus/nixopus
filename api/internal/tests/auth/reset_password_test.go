package auth

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/tests"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
)

func TestRequestResetPassword(t *testing.T) {
	setup := testutils.NewTestSetup()
	authResponse, _, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}
	testCases := []struct {
		name           string
		expectedStatus int
	}{
		{
			name:           "Successfully send password reset link",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			Test(t,
				Description(tc.name),
				Post(tests.GetRequestPasswordResetURL()),
				Send().Headers("Authorization").Add("Bearer "+authResponse.AccessToken),
				Expect().Status().Equal(int64(tc.expectedStatus)),
			)
		})
	}
}

func TestResetPassword(t *testing.T) {
	setup := testutils.NewTestSetup()
	authResponse, organization, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	user, err := setup.AuthService.GetUserByEmail(authResponse.User.Email)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}

	_, resetToken, err := setup.AuthService.GeneratePasswordResetLink(user)
	if err != nil {
		t.Fatalf("failed to generate reset token: %v", err)
	}

	if resetToken == "" {
		t.Fatal("Failed to get reset token")
	}

	testCases := []struct {
		name           string
		request        types.ResetPasswordRequest
		token          string
		expectedStatus int
	}{
		{
			name: "Successfully reset password",
			request: types.ResetPasswordRequest{
				Password: "Password123@",
			},
			token:          resetToken,
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid password format",
			request: types.ResetPasswordRequest{
				Password: "weakpassword",
			},
			token:          resetToken,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid reset token",
			request: types.ResetPasswordRequest{
				Password: "Password123@",
			},
			token:          "invalid-token",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Expired reset token",
			request: types.ResetPasswordRequest{
				Password: "Password123@",
			},
			token:          "6092f3b4-9b16-433f-9bd8-fd289347ac87",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			Test(t,
				Description(tc.name),
				Post(tests.GetResetPasswordURL()+"?token="+tc.token),
				Send().Headers("Authorization").Add("Bearer "+authResponse.AccessToken),
				Send().Headers("X-Organization-ID").Add(organization.ID.String()),
				Send().Body().JSON(tc.request),
				Expect().Status().Equal(int64(tc.expectedStatus)),
			)
		})
	}
}
