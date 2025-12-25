package deploy

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/raghavyuva/nixopus-api/internal/tests"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
)

func TestGetApplicationByID(t *testing.T) {
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
			name:           "Get application by ID without authentication",
			cookies:        "",
			organizationID: orgID,
			applicationID:  "123e4567-e89b-12d3-a456-426614174000",
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no authentication cookies are provided",
		},
		{
			name:           "Get application by ID with invalid cookies",
			cookies:        "invalid-cookies",
			organizationID: orgID,
			applicationID:  "123e4567-e89b-12d3-a456-426614174000",
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when invalid authentication cookies are provided",
		},
		{
			name:           "Get application by ID without organization header",
			cookies:        cookies,
			organizationID: "",
			applicationID:  "123e4567-e89b-12d3-a456-426614174000",
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when organization ID is not provided",
		},
		{
			name:           "Get application by ID with invalid application ID",
			cookies:        cookies,
			organizationID: orgID,
			applicationID:  "invalid-uuid",
			expectedStatus: http.StatusInternalServerError,
			description:    "Should return 500 when application ID format is invalid",
		},
		{
			name:           "Get application by ID that doesn't exist",
			cookies:        cookies,
			organizationID: orgID,
			applicationID:  "123e4567-e89b-12d3-a456-426614174000",
			expectedStatus: http.StatusInternalServerError,
			description:    "Should return 500 when application doesn't exist",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var url string
			if tc.applicationID != "" {
				url = tests.GetDeployApplicationURL() + "?id=" + tc.applicationID
			} else {
				url = tests.GetDeployApplicationURL()
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
