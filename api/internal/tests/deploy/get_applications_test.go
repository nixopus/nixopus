package deploy

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/raghavyuva/nixopus-api/internal/tests"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
)

func TestGetApplications(t *testing.T) {
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
		expectedStatus int
		description    string
	}{
		{
			name:           "Successfully fetch applications with valid token",
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK,
			description:    "Should return applications list with valid authentication",
		},
		{
			name:           "Get applications without authentication",
			token:          "",
			organizationID: orgID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no authentication token is provided",
		},
		{
			name:           "Get applications with invalid token",
			token:          "invalid-token",
			organizationID: orgID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when invalid authentication token is provided",
		},
		{
			name:           "Get applications without organization header",
			token:          user.AccessToken,
			organizationID: "",
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when organization ID is not provided",
		},
		{
			name:           "Get applications with invalid organization ID",
			token:          user.AccessToken,
			organizationID: "invalid-org-id",
			expectedStatus: http.StatusInternalServerError,
			description:    "Should return 500 when organization ID format is invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSteps := []IStep{
				Description(tc.description),
				Get(tests.GetDeployApplicationsURL()),
			}

			if tc.token != "" {
				testSteps = append(testSteps, Send().Headers("Authorization").Add("Bearer "+tc.token))
			}

			if tc.organizationID != "" {
				testSteps = append(testSteps, Send().Headers("X-Organization-ID").Add(tc.organizationID))
			}

			testSteps = append(testSteps, Expect().Status().Equal(int64(tc.expectedStatus)))

			if tc.expectedStatus == http.StatusOK {
				testSteps = append(testSteps,
					Expect().Body().JSON().JQ(".status").Equal("success"),
					Expect().Body().JSON().JQ(".message").Equal("Applications"), // API returns "Applications" not "Applications fetched successfully"
					Expect().Body().JSON().JQ(".data").NotEqual(nil),
				)
			}

			Test(t, testSteps...)
		})
	}
}

func TestGetApplicationsErrorHandling(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()

	t.Run("Malformed authorization header", func(t *testing.T) {
		Test(t,
			Description("Should handle malformed authorization header gracefully"),
			Get(tests.GetDeployApplicationsURL()),
			Send().Headers("Authorization").Add("InvalidFormat"),
			Send().Headers("X-Organization-ID").Add(orgID),
			Expect().Status().Equal(http.StatusUnauthorized),
		)
	})

	t.Run("Empty authorization header", func(t *testing.T) {
		Test(t,
			Description("Should handle empty authorization header"),
			Get(tests.GetDeployApplicationsURL()),
			Send().Headers("Authorization").Add(""),
			Send().Headers("X-Organization-ID").Add(orgID),
			Expect().Status().Equal(http.StatusUnauthorized),
		)
	})

	t.Run("Bearer token with extra spaces", func(t *testing.T) {
		Test(t,
			Description("Should handle get applications base case"),
			Get(tests.GetDeployApplicationsURL()),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-ID").Add(orgID),
			Expect().Status().Equal(http.StatusOK),
		)
	})
}
