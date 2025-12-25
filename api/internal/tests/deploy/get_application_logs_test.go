package deploy

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/raghavyuva/nixopus-api/internal/tests"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
)

func TestGetApplicationLogs(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetSupertokensAuthResponse()
	if err != nil {
		t.Fatalf("failed to get supertokens auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()
	testApplicationID := "123e4567-e89b-12d3-a456-426614174000"

	testCases := []struct {
		name           string
		cookies        string
		organizationID string
		applicationID  string
		expectedStatus int
		description    string
	}{
		{
			name:           "Get application logs without authentication",
			cookies:        "",
			organizationID: orgID,
			applicationID:  testApplicationID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no authentication cookies are provided",
		},
		{
			name:           "Get application logs with invalid cookies",
			cookies:        "invalid-cookies",
			organizationID: orgID,
			applicationID:  testApplicationID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when invalid authentication cookies are provided",
		},
		{
			name:           "Get application logs without organization header",
			cookies:        cookies,
			organizationID: "",
			applicationID:  testApplicationID,
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when organization ID is not provided",
		},
		{
			name:           "Get application logs with invalid application ID format",
			cookies:        cookies,
			organizationID: orgID,
			applicationID:  "invalid-uuid",
			expectedStatus: http.StatusInternalServerError,
			description:    "Should return 500 when application ID format is invalid",
		},
		{
			name:           "Get application logs for non-existent application",
			cookies:        cookies,
			organizationID: orgID,
			applicationID:  testApplicationID,
			expectedStatus: http.StatusOK,
			description:    "Should return 200 with empty logs when application doesn't exist",
		},
		{
			name:           "Get application logs with empty application ID",
			cookies:        cookies,
			organizationID: orgID,
			applicationID:  "",
			expectedStatus: http.StatusNotFound,
			description:    "Should return 404 when application ID is empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var url string
			if tc.applicationID != "" {
				url = tests.GetDeployApplicationLogsURL(tc.applicationID)
			} else {
				url = tests.GetDeployApplicationLogsURL("") // This will result in malformed URL
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

func TestGetApplicationLogsSuccess(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetSupertokensAuthResponse()
	if err != nil {
		t.Fatalf("failed to get supertokens auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()
	testApplicationID := "123e4567-e89b-12d3-a456-426614174000"

	t.Run("Get application logs with valid format", func(t *testing.T) {
		Test(t,
			Description("Should attempt to fetch application logs with valid UUID format"),
			Get(tests.GetDeployApplicationLogsURL(testApplicationID)),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Expect().Status().OneOf(http.StatusOK, http.StatusNotFound), // Either OK with logs or 404 if application doesn't exist
		)
	})
}
