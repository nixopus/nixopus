package deploy

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/tests"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func TestCreateApplication(t *testing.T) {
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
		request        types.CreateDeploymentRequest
		expectedStatus int
		description    string
	}{
		{
			name:           "Successfully create application with valid data",
			cookies:        cookies,
			organizationID: orgID,
			request: types.CreateDeploymentRequest{
				Name:        "test-app",
				Domains:     []string{"test-app.example.com"},
				Environment: shared_types.Development,
				BuildPack:   shared_types.DockerFile,
				Repository:  "https://github.com/test/test-app.git",
				Branch:      "main",
				Port:        3000,
				BuildVariables: map[string]string{
					"NODE_ENV": "development",
				},
				EnvironmentVariables: map[string]string{
					"PORT": "3000",
				},
			},
			expectedStatus: http.StatusOK, // API returns 200 not 201
			description:    "Should create application successfully with valid data",
		},
		{
			name:           "Create application without authentication",
			cookies:        "",
			organizationID: orgID,
			request: types.CreateDeploymentRequest{
				Name:        "test-app",
				Domains:     []string{"test-app.example.com"},
				Environment: shared_types.Development,
				BuildPack:   shared_types.DockerFile,
				Repository:  "https://github.com/test/test-app.git",
				Branch:      "main",
				Port:        3000,
			},
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no authentication token is provided",
		},
		{
			name:           "Create application with invalid cookies",
			cookies:        "invalid-cookies",
			organizationID: orgID,
			request: types.CreateDeploymentRequest{
				Name:        "test-app",
				Domains:     []string{"test-app.example.com"},
				Environment: shared_types.Development,
				BuildPack:   shared_types.DockerFile,
				Repository:  "https://github.com/test/test-app.git",
				Branch:      "main",
				Port:        3000,
			},
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when invalid authentication token is provided",
		},
		{
			name:           "Create application without organization header",
			cookies:        cookies,
			organizationID: "",
			request: types.CreateDeploymentRequest{
				Name:        "test-app",
				Domains:     []string{"test-app.example.com"},
				Environment: shared_types.Development,
				BuildPack:   shared_types.DockerFile,
				Repository:  "https://github.com/test/test-app.git",
				Branch:      "main",
				Port:        3000,
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when organization ID is not provided",
		},
		{
			name:           "Create application with missing name",
			cookies:        cookies,
			organizationID: orgID,
			request: types.CreateDeploymentRequest{
				Domains:     []string{"test-app.example.com"},
				Environment: shared_types.Development,
				BuildPack:   shared_types.DockerFile,
				Repository:  "https://github.com/test/test-app.git",
				Branch:      "main",
				Port:        3000,
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when name is missing",
		},
		{
			name:           "Create application with missing domain",
			cookies:        cookies,
			organizationID: orgID,
			request: types.CreateDeploymentRequest{
				Name:        "test-app",
				Domains:     []string{},
				Environment: shared_types.Development,
				BuildPack:   shared_types.DockerFile,
				Repository:  "https://github.com/test/test-app.git",
				Branch:      "main",
				Port:        3000,
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when domain is missing",
		},
		{
			name:           "Create application with missing repository",
			cookies:        cookies,
			organizationID: orgID,
			request: types.CreateDeploymentRequest{
				Name:        "test-app",
				Domains:     []string{"test-app.example.com"},
				Environment: shared_types.Development,
				BuildPack:   shared_types.DockerFile,
				Branch:      "main",
				Port:        3000,
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when repository is missing",
		},
		{
			name:           "Create application with missing port",
			cookies:        cookies,
			organizationID: orgID,
			request: types.CreateDeploymentRequest{
				Name:        "test-app",
				Domains:     []string{"test-app.example.com"},
				Environment: shared_types.Development,
				BuildPack:   shared_types.DockerFile,
				Repository:  "https://github.com/test/test-app.git",
				Branch:      "main",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when port is missing",
		},
		{
			name:           "Create application with invalid environment",
			cookies:        cookies,
			organizationID: orgID,
			request: types.CreateDeploymentRequest{
				Name:        "test-app",
				Domains:     []string{"test-app.example.com"},
				Environment: "invalid",
				BuildPack:   shared_types.DockerFile,
				Repository:  "https://github.com/test/test-app.git",
				Branch:      "main",
				Port:        3000,
			},
			expectedStatus: http.StatusInternalServerError, // API returns 500 for invalid enum values
			description:    "Should return 500 when environment is invalid",
		},
		{
			name:           "Create application with invalid build pack",
			cookies:        cookies,
			organizationID: orgID,
			request: types.CreateDeploymentRequest{
				Name:        "test-app",
				Domains:     []string{"test-app.example.com"},
				Environment: shared_types.Development,
				BuildPack:   "invalid",
				Repository:  "https://github.com/test/test-app.git",
				Branch:      "main",
				Port:        3000,
			},
			expectedStatus: http.StatusInternalServerError, // API returns 500 for invalid enum values
			description:    "Should return 500 when build pack is invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSteps := []IStep{
				Description(tc.description),
				Post(tests.GetDeployApplicationURL()),
				Send().Body().JSON(tc.request),
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
