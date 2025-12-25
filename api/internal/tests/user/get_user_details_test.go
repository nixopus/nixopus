package user

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/raghavyuva/nixopus-api/internal/tests"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
)

func TestGetUserDetails(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, _, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("Failed to create user and organization: %v", err)
	}

	testCases := []struct {
		name           string
		userID         string
		expectedStatus int
		token          string
	}{
		{
			name:           "Get user details",
			userID:         user.User.ID.String(),
			expectedStatus: http.StatusOK,
			token:          user.AccessToken,
		},
		{
			name:           "Get user details with invalid token format",
			userID:         user.User.ID.String(),
			expectedStatus: http.StatusUnauthorized,
			token:          "invalid-token",
		},
		{
			name:           "Get user details with expired token",
			userID:         user.User.ID.String(),
			expectedStatus: http.StatusUnauthorized,
			token:          "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			Test(t,
				Description(tc.name),
				Get(tests.GetUserDetailsURL()),
				Send().Headers("Authorization").Add("Bearer "+tc.token),
				Expect().Status().Equal(int64(tc.expectedStatus)),
			)
		})
	}
}
