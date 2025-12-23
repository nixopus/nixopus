package deploy

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/raghavyuva/nixopus-api/internal/tests"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
)

func TestGetDeploymentLogs(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()
	testDeploymentID := "123e4567-e89b-12d3-a456-426614174000"

	testCases := []struct {
		name           string
		token          string
		organizationID string
		deploymentID   string
		expectedStatus int
		description    string
	}{
		{
			name:           "Get deployment logs without authentication",
			token:          "",
			organizationID: orgID,
			deploymentID:   testDeploymentID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no authentication token is provided",
		},
		{
			name:           "Get deployment logs with invalid token",
			token:          "invalid-token",
			organizationID: orgID,
			deploymentID:   testDeploymentID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when invalid authentication token is provided",
		},
		{
			name:           "Get deployment logs without organization header",
			token:          user.AccessToken,
			organizationID: "",
			deploymentID:   testDeploymentID,
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when organization ID is not provided",
		},
		{
			name:           "Get deployment logs with invalid deployment ID format",
			token:          user.AccessToken,
			organizationID: orgID,
			deploymentID:   "invalid-uuid",
			expectedStatus: http.StatusInternalServerError,
			description:    "Should return 500 when deployment ID format is invalid",
		},
		{
			name:           "Get deployment logs for non-existent deployment",
			token:          user.AccessToken,
			organizationID: orgID,
			deploymentID:   testDeploymentID,
			expectedStatus: http.StatusOK,
			description:    "Should return 200 with empty logs when deployment doesn't exist",
		},
		{
			name:           "Get deployment logs with empty deployment ID",
			token:          user.AccessToken,
			organizationID: orgID,
			deploymentID:   "",
			expectedStatus: http.StatusInternalServerError,
			description:    "Should return 500 when deployment ID is empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var url string
			if tc.deploymentID != "" {
				url = tests.GetDeployApplicationDeploymentLogsURL(tc.deploymentID)
			} else {
				url = tests.GetDeployApplicationDeploymentLogsURL("") // This will result in malformed URL
			}

			testSteps := []IStep{
				Description(tc.description),
				Get(url),
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

func TestGetDeploymentLogsSuccess(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()
	testDeploymentID := "123e4567-e89b-12d3-a456-426614174000"

	t.Run("Get deployment logs with valid format", func(t *testing.T) {
		Test(t,
			Description("Should attempt to fetch deployment logs with valid UUID format"),
			Get(tests.GetDeployApplicationDeploymentLogsURL(testDeploymentID)),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-ID").Add(orgID),
			Expect().Status().OneOf(http.StatusOK, http.StatusNotFound), // Either OK with logs or 404 if deployment doesn't exist
		)
	})
}
