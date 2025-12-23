package auth

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/tests"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
)

func TestLogout(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, _, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	testCases := []struct {
		name           string
		request        types.LogoutRequest
		expectedStatus int
		token          string
	}{
		{
			name: "Successfully logout",
			request: types.LogoutRequest{
				RefreshToken: user.RefreshToken,
			},
			expectedStatus: http.StatusOK,
			token:          user.AccessToken,
		},
		{
			name: "Logout with invalid refresh token",
			request: types.LogoutRequest{
				RefreshToken: "invalid-refresh-token",
			},
			expectedStatus: http.StatusBadRequest,
			token:          user.AccessToken,
		},
		{
			name: "Logout with expired refresh token",
			request: types.LogoutRequest{
				RefreshToken: "6092f3b4-9b16-433f-9bd8-fd289347ac87",
			},
			expectedStatus: http.StatusInternalServerError,
			token:          user.AccessToken,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			Test(t,
				Description(tc.name),
				Post(tests.GetLogoutURL()),
				Send().Headers("Authorization").Add("Bearer "+tc.token),
				Send().Body().JSON(tc.request),
				Expect().Status().Equal(int64(tc.expectedStatus)),
			)
		})
	}
}
