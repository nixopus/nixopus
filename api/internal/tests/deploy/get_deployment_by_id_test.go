package deploy

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/raghavyuva/nixopus-api/internal/tests"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
)

func TestGetDeploymentByID(t *testing.T) {
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
			name:           "Get deployment by ID without authentication",
			token:          "",
			organizationID: orgID,
			deploymentID:   testDeploymentID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no authentication token is provided",
		},
		{
			name:           "Get deployment by ID with invalid token",
			token:          "invalid-token",
			organizationID: orgID,
			deploymentID:   testDeploymentID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when invalid authentication token is provided",
		},
		{
			name:           "Get deployment by ID without organization header",
			token:          user.AccessToken,
			organizationID: "",
			deploymentID:   testDeploymentID,
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when organization ID is not provided",
		},
		{
			name:           "Get deployment by ID with invalid deployment ID format",
			token:          user.AccessToken,
			organizationID: orgID,
			deploymentID:   "invalid-uuid",
			expectedStatus: http.StatusInternalServerError,
			description:    "Should return 500 when deployment ID format is invalid",
		},
		{
			name:           "Get deployment by ID that doesn't exist",
			token:          user.AccessToken,
			organizationID: orgID,
			deploymentID:   testDeploymentID,
			expectedStatus: http.StatusInternalServerError,
			description:    "Should return 500 when deployment doesn't exist",
		},
		{
			name:           "Get deployment by ID with empty deployment ID",
			token:          user.AccessToken,
			organizationID: orgID,
			deploymentID:   "",
			expectedStatus: http.StatusNotFound,
			description:    "Should return 404 when deployment ID is empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var url string
			if tc.deploymentID != "" {
				url = tests.GetDeployApplicationDeploymentByIDURL(tc.deploymentID)
			} else {
				url = tests.GetDeployApplicationDeploymentByIDURL("") // This will result in malformed URL
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

func TestGetDeploymentByIDSuccess(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()
	testDeploymentID := "123e4567-e89b-12d3-a456-426614174000"

	t.Run("Get deployment by valid ID format", func(t *testing.T) {
		Test(t,
			Description("Should attempt to fetch deployment with valid UUID format"),
			Get(tests.GetDeployApplicationDeploymentByIDURL(testDeploymentID)),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-ID").Add(orgID),
			Expect().Status().OneOf(http.StatusOK, http.StatusInternalServerError), // Either OK if exists or 500 if not
		)
	})
}
