package deploy

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/tests"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
)

func TestRestartApplication(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetSupertokensAuthResponse()
	if err != nil {
		t.Fatalf("failed to get supertokens auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()
	testApplicationID := uuid.New()

	testCases := []struct {
		name           string
		cookies        string
		organizationID string
		request        types.RestartDeploymentRequest
		expectedStatus int
		description    string
	}{
		{
			name:           "Restart application without authentication",
			cookies:        "",
			organizationID: orgID,
			request: types.RestartDeploymentRequest{
				ID: testApplicationID,
			},
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no authentication cookies are provided",
		},
		{
			name:           "Restart application with invalid cookies",
			cookies:        "invalid-cookies",
			organizationID: orgID,
			request: types.RestartDeploymentRequest{
				ID: testApplicationID,
			},
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when invalid authentication cookies are provided",
		},
		{
			name:           "Restart application without organization header",
			cookies:        cookies,
			organizationID: "",
			request: types.RestartDeploymentRequest{
				ID: testApplicationID,
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when organization ID is not provided",
		},
		{
			name:           "Restart application with missing ID",
			cookies:        cookies,
			organizationID: orgID,
			request:        types.RestartDeploymentRequest{},
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when application ID is missing",
		},
		{
			name:           "Restart application that doesn't exist",
			cookies:        cookies,
			organizationID: orgID,
			request: types.RestartDeploymentRequest{
				ID: testApplicationID,
			},
			expectedStatus: http.StatusInternalServerError,
			description:    "Should return 500 when application doesn't exist",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSteps := []IStep{
				Description(tc.description),
				Post(tests.GetDeployApplicationRestartURL()),
				Send().Body().JSON(tc.request),
			}

			if tc.cookies != "" {
				testSteps = append(testSteps, Send().Headers("Cookie").Add(tc.cookies))
			}

			if tc.organizationID != "" {
				testSteps = append(testSteps, Send().Headers("X-Organization-ID").Add(tc.organizationID))
			}

			testSteps = append(testSteps, Expect().Status().Equal(int64(tc.expectedStatus)))

			Test(t, testSteps...)
		})
	}
}
