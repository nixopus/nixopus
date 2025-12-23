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

func TestRedeployApplication(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()
	testApplicationID := uuid.New()

	testCases := []struct {
		name           string
		token          string
		organizationID string
		request        types.ReDeployApplicationRequest
		expectedStatus int
		description    string
	}{
		{
			name:           "Redeploy application without authentication",
			token:          "",
			organizationID: orgID,
			request: types.ReDeployApplicationRequest{
				ID:    testApplicationID,
				Force: false,
			},
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no authentication token is provided",
		},
		{
			name:           "Redeploy application with invalid token",
			token:          "invalid-token",
			organizationID: orgID,
			request: types.ReDeployApplicationRequest{
				ID:    testApplicationID,
				Force: false,
			},
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when invalid authentication token is provided",
		},
		{
			name:           "Redeploy application without organization header",
			token:          user.AccessToken,
			organizationID: "",
			request: types.ReDeployApplicationRequest{
				ID:    testApplicationID,
				Force: false,
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when organization ID is not provided",
		},
		{
			name:           "Redeploy application with missing ID",
			token:          user.AccessToken,
			organizationID: orgID,
			request:        types.ReDeployApplicationRequest{},
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when application ID is missing",
		},
		{
			name:           "Redeploy application that doesn't exist",
			token:          user.AccessToken,
			organizationID: orgID,
			request: types.ReDeployApplicationRequest{
				ID:    testApplicationID,
				Force: false,
			},
			expectedStatus: http.StatusInternalServerError,
			description:    "Should return 500 when application doesn't exist",
		},
		{
			name:           "Redeploy application with force flag",
			token:          user.AccessToken,
			organizationID: orgID,
			request: types.ReDeployApplicationRequest{
				ID:    testApplicationID,
				Force: true,
			},
			expectedStatus: http.StatusInternalServerError, // API returns 500 since app doesn't exist
			description:    "Should redeploy application with force flag",
		},
		{
			name:           "Redeploy application with force without cache",
			token:          user.AccessToken,
			organizationID: orgID,
			request: types.ReDeployApplicationRequest{
				ID:                testApplicationID,
				Force:             true,
				ForceWithoutCache: true,
			},
			expectedStatus: http.StatusInternalServerError, // API returns 500 since app doesn't exist
			description:    "Should redeploy application with force without cache",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSteps := []IStep{
				Description(tc.description),
				Post(tests.GetDeployApplicationRedeployURL()),
				Send().Body().JSON(tc.request),
			}

			if tc.token != "" {
				testSteps = append(testSteps, Send().Headers("Authorization").Add("Bearer "+tc.token))
			}

			if tc.organizationID != "" {
				testSteps = append(testSteps, Send().Headers("X-Organization-ID").Add(tc.organizationID))
			}

			testSteps = append(testSteps, Expect().Status().Equal(int64(tc.expectedStatus)))

			Test(t, testSteps...)
		})
	}
}
