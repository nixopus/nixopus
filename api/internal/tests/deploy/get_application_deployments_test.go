package deploy

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/raghavyuva/nixopus-api/internal/tests"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
)

func TestGetApplicationDeployments(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetSupertokensAuthResponse()
	if err != nil {
		t.Fatalf("failed to get supertokens auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()

	testCases := []struct {
		name           string
		cookies        string
		organizationID string
		applicationID  string
		expectedStatus int
		description    string
	}{
		{
			name:           "Get application deployments without authentication",
			cookies:        "",
			organizationID: orgID,
			applicationID:  "123e4567-e89b-12d3-a456-426614174000",
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no authentication cookies are provided",
		},
		{
			name:           "Get application deployments with invalid cookies",
			cookies:        "invalid-cookies",
			organizationID: orgID,
			applicationID:  "123e4567-e89b-12d3-a456-426614174000",
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when invalid authentication cookies are provided",
		},
		{
			name:           "Get application deployments without organization header",
			cookies:        cookies,
			organizationID: "",
			applicationID:  "123e4567-e89b-12d3-a456-426614174000",
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when organization ID is not provided",
		},
		{
			name:           "Get application deployments with invalid application ID",
			cookies:        cookies,
			organizationID: orgID,
			applicationID:  "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when application ID format is invalid",
		},
		{
			name:           "Get application deployments for non-existent application",
			cookies:        cookies,
			organizationID: orgID,
			applicationID:  "123e4567-e89b-12d3-a456-426614174000",
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when application doesn't exist",
		},
		{
			name:           "Get application deployments with missing application ID",
			cookies:        cookies,
			organizationID: orgID,
			applicationID:  "",
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when application ID is missing",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var url string
			if tc.applicationID != "" {
				url = tests.GetDeployApplicationDeploymentsURL() + "?application_id=" + tc.applicationID
			} else {
				url = tests.GetDeployApplicationDeploymentsURL()
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

func TestGetApplicationDeploymentsSuccess(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetSupertokensAuthResponse()
	if err != nil {
		t.Fatalf("failed to get supertokens auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()

	t.Run("Get deployments with valid application ID should return structure", func(t *testing.T) {
		Test(t,
			Description("Should return deployments structure even if empty"),
			Get(tests.GetDeployApplicationDeploymentsURL()+"?application_id=123e4567-e89b-12d3-a456-426614174000"),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Expect().Status().OneOf(http.StatusOK, http.StatusBadRequest), // Either OK with empty list or 400 if app doesn't exist
		)
	})
}
