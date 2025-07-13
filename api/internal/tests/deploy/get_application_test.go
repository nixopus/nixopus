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
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()

	testCases := []struct {
		name           string
		token          string
		organizationID string
		applicationID  string
		expectedStatus int
		description    string
	}{
		{
			name:           "Get application by ID without authentication",
			token:          "",
			organizationID: orgID,
			applicationID:  "123e4567-e89b-12d3-a456-426614174000",
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no authentication token is provided",
		},
		{
			name:           "Get application by ID with invalid token",
			token:          "invalid-token",
			organizationID: orgID,
			applicationID:  "123e4567-e89b-12d3-a456-426614174000",
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when invalid authentication token is provided",
		},
		{
			name:           "Get application by ID without organization header",
			token:          user.AccessToken,
			organizationID: "",
			applicationID:  "123e4567-e89b-12d3-a456-426614174000",
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when organization ID is not provided",
		},
		{
			name:           "Get application by ID with invalid application ID",
			token:          user.AccessToken,
			organizationID: orgID,
			applicationID:  "invalid-uuid",
			expectedStatus: http.StatusInternalServerError,
			description:    "Should return 500 when application ID format is invalid",
		},
		{
			name:           "Get application by ID that doesn't exist",
			token:          user.AccessToken,
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
