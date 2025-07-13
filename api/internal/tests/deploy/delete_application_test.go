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

func TestDeleteApplication(t *testing.T) {
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
		request        types.DeleteDeploymentRequest
		expectedStatus int
		description    string
	}{
		{
			name:           "Delete application without authentication",
			token:          "",
			organizationID: orgID,
			request: types.DeleteDeploymentRequest{
				ID: testApplicationID,
			},
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no authentication token is provided",
		},
		{
			name:           "Delete application with invalid token",
			token:          "invalid-token",
			organizationID: orgID,
			request: types.DeleteDeploymentRequest{
				ID: testApplicationID,
			},
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when invalid authentication token is provided",
		},
		{
			name:           "Delete application without organization header",
			token:          user.AccessToken,
			organizationID: "",
			request: types.DeleteDeploymentRequest{
				ID: testApplicationID,
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when organization ID is not provided",
		},
		{
			name:           "Delete application with missing ID",
			token:          user.AccessToken,
			organizationID: orgID,
			request:        types.DeleteDeploymentRequest{},
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when application ID is missing",
		},
		{
			name:           "Delete application that doesn't exist",
			token:          user.AccessToken,
			organizationID: orgID,
			request: types.DeleteDeploymentRequest{
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
				Delete(tests.GetDeployApplicationURL()),
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
