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
	auth, err := setup.GetSupertokensAuthResponse()
	if err != nil {
		t.Fatalf("failed to get supertokens auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()
	testDeploymentID := "123e4567-e89b-12d3-a456-426614174000"

	testCases := []struct {
		name           string
		cookies        string
		organizationID string
		deploymentID   string
		expectedStatus int
		description    string
	}{
		{
			name:           "Get deployment logs without authentication",
			cookies:        "",
			organizationID: orgID,
			deploymentID:   testDeploymentID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no authentication cookies are provided",
		},
		{
			name:           "Get deployment logs with invalid cookies",
			cookies:        "invalid-cookies",
			organizationID: orgID,
			deploymentID:   testDeploymentID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when invalid authentication cookies are provided",
		},
		{
			name:           "Get deployment logs without organization header",
			cookies:        cookies,
			organizationID: "",
			deploymentID:   testDeploymentID,
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when organization ID is not provided",
		},
		{
			name:           "Get deployment logs with invalid deployment ID format",
			cookies:        cookies,
			organizationID: orgID,
			deploymentID:   "invalid-uuid",
			expectedStatus: http.StatusInternalServerError,
			description:    "Should return 500 when deployment ID format is invalid",
		},
		{
			name:           "Get deployment logs for non-existent deployment",
			cookies:        cookies,
			organizationID: orgID,
			deploymentID:   testDeploymentID,
			expectedStatus: http.StatusOK,
			description:    "Should return 200 with empty logs when deployment doesn't exist",
		},
		{
			name:           "Get deployment logs with empty deployment ID",
			cookies:        cookies,
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

func TestGetDeploymentLogsSuccess(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetSupertokensAuthResponse()
	if err != nil {
		t.Fatalf("failed to get supertokens auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()
	testDeploymentID := "123e4567-e89b-12d3-a456-426614174000"

	t.Run("Get deployment logs with valid format", func(t *testing.T) {
		Test(t,
			Description("Should attempt to fetch deployment logs with valid UUID format"),
			Get(tests.GetDeployApplicationDeploymentLogsURL(testDeploymentID)),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Expect().Status().OneOf(http.StatusOK, http.StatusNotFound), // Either OK with logs or 404 if deployment doesn't exist
		)
	})
}
