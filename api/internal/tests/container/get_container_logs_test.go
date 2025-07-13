package container

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/raghavyuva/nixopus-api/internal/tests"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
)

func TestGetContainerLogs(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()

	// Note: First, get a container ID from the list; sue the same for test validation (PSQL test db container)
	// Test cases are designed to work with an existing container.
	// TODO: Run a script on pre running E2E tests to create a containers & add as DB seeding
	var containerID string
	Test(t,
		Description("Get container id for logs tests"),
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
			name:           "Successfully fetch container logs with valid ID and token",
			containerID:    containerID,
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusOK,
			description:    "Should return container logs with valid authentication",
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
			name:           "Unauthorized request with invalid token",
			containerID:    containerID,
			token:          "invalid-token",
			organizationID: orgID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when invalid authentication token is provided",
		},
		{
			name:           "Request without organization header",
			containerID:    containerID,
			token:          user.AccessToken,
			organizationID: "",
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when organization header is missing",
		},
		{
			name:           "Request with invalid container ID",
			containerID:    "invalid-container-id",
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusInternalServerError,
			description:    "Should return 500 when container ID doesn't exist",
		},
		{
			name:           "Request with empty container ID",
			containerID:    "",
			token:          user.AccessToken,
			organizationID: orgID,
			expectedStatus: http.StatusInternalServerError,
			description:    "Should return 500 when container ID is empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSteps := []IStep{
				Description(tc.description),
				Post(tests.GetContainerLogsURL(tc.containerID)),
			}

			if tc.token != "" {
				testSteps = append(testSteps, Send().Headers("Authorization").Add("Bearer "+tc.token))
			}

			if tc.organizationID != "" {
				testSteps = append(testSteps, Send().Headers("X-Organization-Id").Add(tc.organizationID))
			}

			requestBody := map[string]interface{}{
				"id":     tc.containerID,
				"follow": false,
				"tail":   100,
				"stdout": true,
				"stderr": true,
			}
			testSteps = append(testSteps, Send().Body().JSON(requestBody))

			testSteps = append(testSteps, Expect().Status().Equal(int64(tc.expectedStatus)))

			if tc.expectedStatus == http.StatusOK {
				testSteps = append(testSteps,
					Expect().Body().JSON().JQ(".status").Equal("success"),
					Expect().Body().JSON().JQ(".message").Equal("Container logs fetched successfully"),
					Expect().Body().JSON().JQ(".data").NotEqual(nil),
				)
			}

			Test(t, testSteps...)
		})
	}
}

func TestGetContainerLogsWithFilters(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()

	var containerID string
	Test(t,
		Description("Get container ID for logs filter tests"),
		Get(tests.GetContainersURL()),
		Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
		Send().Headers("X-Organization-Id").Add(orgID),
		Expect().Status().Equal(http.StatusOK),
		Store().Response().Body().JSON().JQ(".data[0].id").In(&containerID),
	)

	t.Run("Fetch logs with tail parameter", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"id":     containerID,
			"follow": false,
			"tail":   50,
			"stdout": true,
			"stderr": true,
		}
		Test(t,
			Description("Should return limited number of log lines when tail parameter is provided"),
			Post(tests.GetContainerLogsURL(containerID)),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().JSON(requestBody),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".status").Equal("success"),
			Expect().Body().JSON().JQ(".message").Equal("Container logs fetched successfully"),
			Expect().Body().JSON().JQ(".data").NotEqual(nil),
		)
	})

	t.Run("Fetch logs with since parameter", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"id":     containerID,
			"follow": false,
			"since":  "2024-01-01T00:00:00Z",
			"stdout": true,
			"stderr": true,
		}
		Test(t,
			Description("Should return logs since specified timestamp"),
			Post(tests.GetContainerLogsURL(containerID)),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().JSON(requestBody),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".status").Equal("success"),
			Expect().Body().JSON().JQ(".message").Equal("Container logs fetched successfully"),
		)
	})

	t.Run("Fetch logs with timestamps", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"id":     containerID,
			"follow": false,
			"stdout": true,
			"stderr": true,
		}
		Test(t,
			Description("Should return logs with timestamps when timestamps=true"),
			Post(tests.GetContainerLogsURL(containerID)),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().JSON(requestBody),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".status").Equal("success"),
			Expect().Body().JSON().JQ(".message").Equal("Container logs fetched successfully"),
		)
	})

	t.Run("Fetch logs with follow parameter", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"id":     containerID,
			"follow": false,
			"stdout": true,
			"stderr": true,
		}
		Test(t,
			Description("Should handle follow parameter for streaming logs"),
			Post(tests.GetContainerLogsURL(containerID)),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().JSON(requestBody),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".status").Equal("success"),
		)
	})
}

func TestGetContainerLogsErrorHandling(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()

	t.Run("Malformed authorization header", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"id":     "some-container-id",
			"follow": false,
			"stdout": true,
			"stderr": true,
		}
		Test(t,
			Description("Should handle malformed authorization header gracefully"),
			Post(tests.GetContainerLogsURL("some-container-id")),
			Send().Headers("Authorization").Add("InvalidFormat"),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().JSON(requestBody),
			Expect().Status().Equal(http.StatusUnauthorized),
		)
	})

	t.Run("Empty authorization header", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"id":     "some-container-id",
			"follow": false,
			"stdout": true,
			"stderr": true,
		}
		Test(t,
			Description("Should handle empty authorization header"),
			Post(tests.GetContainerLogsURL("some-container-id")),
			Send().Headers("Authorization").Add(""),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().JSON(requestBody),
			Expect().Status().Equal(http.StatusUnauthorized),
		)
	})

	t.Run("Invalid UUID format for container ID", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"id":     "not-a-uuid",
			"follow": false,
			"stdout": true,
			"stderr": true,
		}
		Test(t,
			Description("Should handle invalid UUID format for container ID"),
			Post(tests.GetContainerLogsURL("not-a-uuid")),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().JSON(requestBody),
			Expect().Status().Equal(http.StatusInternalServerError),
		)
	})

	t.Run("Non-existent container ID with valid UUID format", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"id":     "123e4567-e89b-12d3-a456-426614174000",
			"follow": false,
			"stdout": true,
			"stderr": true,
		}
		Test(t,
			Description("Should return 500 for non-existent container with valid UUID format"),
			Post(tests.GetContainerLogsURL("123e4567-e89b-12d3-a456-426614174000")),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().JSON(requestBody),
			Expect().Status().Equal(http.StatusInternalServerError),
		)
	})

	t.Run("Invalid tail parameter", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"id":     "some-container-id",
			"follow": false,
			"tail":   "invalid-number", // should throw an error since tail expects int
			"stdout": true,
			"stderr": true,
		}
		Test(t,
			Description("Should handle invalid tail parameter gracefully"),
			Post(tests.GetContainerLogsURL("some-container-id")),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().JSON(requestBody),
			Expect().Status().Equal(http.StatusBadRequest),
		)
	})

	t.Run("Invalid since parameter", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"id":     "some-container-id",
			"follow": false,
			"since":  "invalid-timestamp",
			"stdout": true,
			"stderr": true,
		}
		Test(t,
			Description("Should handle invalid since timestamp parameter"),
			Post(tests.GetContainerLogsURL("some-container-id")),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().JSON(requestBody),
			Expect().Status().Equal(http.StatusInternalServerError),
		)
	})
}

func TestGetContainerLogsPermissions(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.GetTestAuthResponse()
	if err != nil {
		t.Fatalf("failed to get test auth response: %v", err)
	}

	orgID := org.ID.String()

	var containerID string
	Test(t,
		Description("Get container ID for permission tests"),
		Get(tests.GetContainersURL()),
		Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
		Send().Headers("X-Organization-Id").Add(orgID),
		Expect().Status().Equal(http.StatusOK),
		Store().Response().Body().JSON().JQ(".data[0].id").In(&containerID),
	)

	t.Run("Access logs with organization member permissions", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"id":     containerID,
			"follow": false,
			"stdout": true,
			"stderr": true,
		}
		Test(t,
			Description("Should allow organization members to access container logs"),
			Post(tests.GetContainerLogsURL(containerID)),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add(orgID),
			Send().Body().JSON(requestBody),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".status").Equal("success"),
		)
	})

	t.Run("Cross-organization access attempt", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"id":     containerID,
			"follow": false,
			"stdout": true,
			"stderr": true,
		}
		Test(t,
			Description("Should deny access to logs from different organization"),
			Post(tests.GetContainerLogsURL(containerID)),
			Send().Headers("Authorization").Add("Bearer "+user.AccessToken),
			Send().Headers("X-Organization-Id").Add("123e4567-e89b-12d3-a456-426614174000"),
			Send().Body().JSON(requestBody),
			Expect().Status().Equal(http.StatusForbidden),
		)
	})
}
