package auth

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/raghavyuva/nixopus-api/internal/tests"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
)

func TestSendVerificationEmail(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, _, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	testCases := []struct {
		name           string
		expectedStatus int
		token          string
	}{
		{
			name:           "Successfully send verification email",
			expectedStatus: http.StatusOK,
			token:          user.AccessToken,
		},
		{
			name:           "Send verification email without token",
			expectedStatus: http.StatusUnauthorized,
			token:          "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			Test(t,
				Description(tc.name),
				Post(tests.GetSendVerificationEmailURL()),
				Send().Headers("Authorization").Add("Bearer "+tc.token),
				Expect().Status().Equal(int64(tc.expectedStatus)),
			)
		})
	}
}

func TestVerifyEmail(t *testing.T) {
	setup := testutils.NewTestSetup()
	authResponse, _, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	Test(t,
		Post(tests.GetSendVerificationEmailURL()),
		Send().Headers("Authorization").Add("Bearer "+authResponse.AccessToken),
		Expect().Status().Equal(http.StatusOK),
	)

	testCases := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "Invalid verification token",
			token:          "invalid-token",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty verification token",
			token:          "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			Test(t,
				Description(tc.name),
				Get(tests.GetVerifyEmailURL()+"?token="+tc.token),
				Send().Headers("Authorization").Add("Bearer "+authResponse.AccessToken),
				Expect().Status().Equal(int64(tc.expectedStatus)),
			)
		})
	}
}
