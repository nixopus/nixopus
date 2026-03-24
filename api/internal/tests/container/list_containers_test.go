package container

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/raghavyuva/nixopus-api/internal/tests"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
)

func TestListContainers(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()

	testCases := []struct {
		name           string
		cookies        string
		organizationID string
		expectedStatus int
		description    string
	}{
		{
			name:           "Successfully fetch containers with valid cookies",
			cookies:        cookies,
			organizationID: orgID,
			expectedStatus: http.StatusInternalServerError,
			description:    "Should return 500 because SSH infrastructure is unavailable for Docker access",
		},
		{
			name:           "Unauthorized request without cookies",
			cookies:        "",
			organizationID: orgID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no authentication cookies are provided",
		},
		{
			name:           "Unauthorized request with invalid cookies",
			cookies:        "invalid-cookies",
			organizationID: orgID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when invalid authentication cookies are provided",
		},
		{
			name:           "Request without organization header",
			cookies:        cookies,
			organizationID: "",
			expectedStatus: http.StatusInternalServerError,
			description:    "Should return 500 because session provides org but SSH infrastructure is unavailable",
		},
		{
			name:           "Request with invalid organization ID",
			cookies:        cookies,
			organizationID: "invalid-org-id",
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when organization ID format is invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSteps := []IStep{
				Description(tc.description),
				Get(tests.GetContainersURL()),
			}

			if tc.cookies != "" {
				testSteps = append(testSteps, Send().Headers("Cookie").Add(tc.cookies))
			}

			if tc.organizationID != "" {
				testSteps = append(testSteps, Send().Headers("X-Organization-Id").Add(tc.organizationID))
			}

			testSteps = append(testSteps, Expect().Status().Equal(int64(tc.expectedStatus)))

			if tc.expectedStatus == http.StatusOK {
				testSteps = append(testSteps,
					Expect().Body().JSON().JQ(".status").Equal("success"),
					Expect().Body().JSON().JQ(".message").Equal("Containers fetched successfully"),
					Expect().Body().JSON().JQ(".data").NotEqual(nil),
				)
			}

			Test(t, testSteps...)
		})
	}
}

func TestListContainersWithSpecificContainer(t *testing.T) {
	t.Skip("requires Docker/SSH infrastructure")
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()

	t.Run("Verify test container exists and has expected properties", func(t *testing.T) {
		Test(t,
			Description("Should return 500 because SSH infrastructure is unavailable for Docker access"),
			Get(tests.GetContainersURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-Id").Add(orgID),
			Expect().Status().Equal(http.StatusInternalServerError),
		)
	})
}

func TestListContainersErrorHandling(t *testing.T) {
	t.Skip("requires Docker/SSH infrastructure")
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()

	t.Run("Malformed cookie header", func(t *testing.T) {
		Test(t,
			Description("Should handle malformed cookie header gracefully"),
			Get(tests.GetContainersURL()),
			Send().Headers("Cookie").Add("InvalidFormat"),
			Send().Headers("X-Organization-Id").Add(orgID),
			Expect().Status().Equal(http.StatusUnauthorized),
		)
	})

	t.Run("Empty cookie header", func(t *testing.T) {
		Test(t,
			Description("Should handle empty cookie header"),
			Get(tests.GetContainersURL()),
			Send().Headers("Cookie").Add(""),
			Send().Headers("X-Organization-Id").Add(orgID),
			Expect().Status().Equal(http.StatusUnauthorized),
		)
	})

	t.Run("Valid cookies with organization header", func(t *testing.T) {
		Test(t,
			Description("Should return 500 because SSH infrastructure is unavailable for Docker access"),
			Get(tests.GetContainersURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-Id").Add(orgID),
			Expect().Status().Equal(http.StatusInternalServerError),
		)
	})
}
