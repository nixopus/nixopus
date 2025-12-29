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
	auth, err := setup.GetSupertokensAuthResponse()
	if err != nil {
		t.Fatalf("failed to get supertokens auth response: %v", err)
	}

	userID := auth.User.ID.String()
	cookies := auth.GetAuthCookiesHeader()

	testCases := []struct {
		name           string
		userID         string
		expectedStatus int
		cookies        string
		description    string
	}{
		{
			name:           "Get user details",
			userID:         userID,
			expectedStatus: http.StatusOK,
			cookies:        cookies,
			description:    "should return user details with valid cookies",
		},
		{
			name:           "Get user details with invalid cookies",
			userID:         userID,
			expectedStatus: http.StatusUnauthorized,
			cookies:        "invalid-cookies",
			description:    "should return 401 when invalid cookies are provided",
		},
		{
			name:           "Get user details without cookies",
			userID:         userID,
			expectedStatus: http.StatusUnauthorized,
			cookies:        "",
			description:    "should return 401 when no cookies are provided",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSteps := []IStep{
				Description(tc.description),
				Get(tests.GetUserDetailsURL()),
			}

			if tc.cookies != "" {
				testSteps = append(testSteps, Send().Headers("Cookie").Add(tc.cookies))
			}

			testSteps = append(testSteps, Expect().Status().Equal(int64(tc.expectedStatus)))

			if tc.expectedStatus == http.StatusOK {
				testSteps = append(testSteps,
					Expect().Body().JSON().JQ(".status").Equal("success"),
					Expect().Body().JSON().JQ(".data").NotEqual(nil),
				)
			}

			Test(t, testSteps...)
		})
	}
}
