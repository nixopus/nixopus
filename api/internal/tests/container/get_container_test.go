package container

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/raghavyuva/nixopus-api/internal/tests"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
)

func TestGetContainer(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()

	// First, get a container ID from the list
	var containerID string
	Test(t,
		Description("Get container ID for individual container tests"),
		Get(tests.GetContainersURL()),
		Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
		Send().Headers("X-Organization-Id").Add(orgID),
		Expect().Status().Equal(http.StatusOK),
		Store().Response().Body().JSON().JQ(".data[0].id").In(&containerID),
	)

	testCases := []struct {
		name           string
		containerID    string
		token          string
		organizationID string
		expectedStatus int
		description    string
	}{
		{
			name:           "Successfully fetch container with valid ID and token",
			containerID:    containerID,
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK,
			description:    "Should return container details",
		},
		{
			name:           "Unauthorized request without token",
			containerID:    containerID,
			token:          "",
			organizationID: orgID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no authentication token is provided",
		},
		{
			name:           "Request with invalid container ID",
			containerID:    "invalid-container-id",
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusInternalServerError,
			description:    "Should return 500 when container ID is invalid/doesnt exist",
		},
		{
			name:           "Request with container ID doesnt exist",
			containerID:    "1234567890123456789012345678901234567890123456789012345678901234",
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusInternalServerError,
			description:    "Should return 500 when container doesnt exist",
		},
		{
			name:           "Request without organization header",
			containerID:    containerID,
			token:          user.AccessToken,
			organizationID: "",
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when organization header is missing",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Skip tests that depend on valid container ID if we couldn't get one
			if tc.containerID == containerID && containerID == "" {
				t.Skip("No container ID available for testing")
			}

			testSteps := []IStep{
				Description(tc.description),
				Get(tests.GetContainerURL(tc.containerID)),
			}

			// Add authentication header if token is provided
			if tc.token != "" {
				testSteps = append(testSteps, Send().Headers("Authorization").Add("Bearer "+tc.token))
			}

			// Add organization header if provided
			if tc.organizationID != "" {
				testSteps = append(testSteps, Send().Headers("X-Organization-Id").Add(tc.organizationID))
			}

			testSteps = append(testSteps, Expect().Status().Equal(int64(tc.expectedStatus)))

			// Additional validations for successful response
			if tc.expectedStatus == http.StatusOK {
				testSteps = append(testSteps,
					Expect().Body().JSON().JQ(".status").Equal("success"),
					Expect().Body().JSON().JQ(".message").Equal("Container fetched successfully"),
					Expect().Body().JSON().JQ(".data").NotEqual(nil),
					Expect().Body().JSON().JQ(".data.id").Equal(tc.containerID),
				)
			}

			Test(t, testSteps...)
		})
	}
}

func TestGetContainerDetailedValidation(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()

	// Get the test container ID specifically
	var containerID string
	Test(t,
		Description("Get test container ID for detailed validation"),
		Get(tests.GetContainersURL()),
		Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
		Send().Headers("X-Organization-Id").Add(orgID),
		Expect().Status().Equal(http.StatusOK),
		Store().Response().Body().JSON().JQ(`.data[] | select(.name == "nixopus-test-db-container") | .id`).In(&containerID),
	)

	if containerID == "" {
		t.Skip("nixopus-test-db-container not found, skipping detailed validation")
	}

	t.Run("Validate complete container structure for test container", func(t *testing.T) {
		Test(t,
			Description("Should return complete container structure with all expected fields"),
			Get(tests.GetContainerURL(containerID)),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".status").Equal("success"),
			Expect().Body().JSON().JQ(".message").Equal("Container fetched successfully"),
			Expect().Body().JSON().JQ(".data.name").Equal("nixopus-test-db-container"),
			Expect().Body().JSON().JQ(".data.image").Equal("postgres:14-alpine"),
			Expect().Body().JSON().JQ(".data.command").NotEqual(""),
			Expect().Body().JSON().JQ(".data.status").NotEqual(""),
			Expect().Body().JSON().JQ(".data.state").NotEqual(""),
			Expect().Body().JSON().JQ(".data.created").NotEqual(""),
			Expect().Body().JSON().JQ(".data.labels").NotEqual(nil),
			Expect().Body().JSON().JQ(".data.ports").NotEqual(nil),
			Expect().Body().JSON().JQ(".data.mounts").NotEqual(nil),
			Expect().Body().JSON().JQ(".data.networks").NotEqual(nil),
			Expect().Body().JSON().JQ(".data.host_config").NotEqual(nil),

			Expect().Body().JSON().JQ(".data.ports[0].private_port").Equal(float64(5432)),
			Expect().Body().JSON().JQ(".data.ports[0].public_port").Equal(float64(5433)),
			Expect().Body().JSON().JQ(".data.ports[0].type").Equal("tcp"),

			Expect().Body().JSON().JQ(".data.host_config.memory").NotEqual(nil),
			Expect().Body().JSON().JQ(".data.host_config.memory_swap").NotEqual(nil),
			Expect().Body().JSON().JQ(".data.host_config.cpu_shares").NotEqual(nil),
		)
	})
}

func TestGetContainerErrorScenarios(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()

	t.Run("Container ID with special characters", func(t *testing.T) {
		Test(t,
			Description("Should handle container ID with special characters"),
			Get(tests.GetContainerURL("container-special")),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Expect().Status().Equal(http.StatusInternalServerError),
		)
	})

	t.Run("Empty container ID", func(t *testing.T) {
		Test(t,
			Description("Should handle empty container ID"),
			Get(tests.GetContainerURL("")),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Expect().Status().Equal(http.StatusNotFound),
		)
	})

	t.Run("Very long container ID", func(t *testing.T) {
		longID := "abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890"
		Test(t,
			Description("Should handle very long container ID"),
			Get(tests.GetContainerURL(longID)),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Expect().Status().Equal(http.StatusInternalServerError),
		)
	})
}
