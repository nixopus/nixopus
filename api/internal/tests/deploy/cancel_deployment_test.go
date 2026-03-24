package deploy

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/tests"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
)

func TestCancelDeployment(t *testing.T) {
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
		request        types.CancelDeploymentRequest
		expectedStatus int
		description    string
	}{
		{
			name:           "Cancel deployment without authentication",
			cookies:        "",
			organizationID: orgID,
			request: types.CancelDeploymentRequest{
				DeploymentID: uuid.New(),
			},
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no authentication cookies are provided",
		},
		{
			name:           "Cancel deployment with missing deployment ID",
			cookies:        cookies,
			organizationID: orgID,
			request:        types.CancelDeploymentRequest{},
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when deployment ID is missing",
		},
		{
			name:           "Cancel deployment that doesn't exist",
			cookies:        cookies,
			organizationID: orgID,
			request: types.CancelDeploymentRequest{
				DeploymentID: uuid.New(),
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when deployment is not running on this instance",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSteps := []IStep{
				Description(tc.description),
				Post(tests.GetDeployApplicationCancelURL()),
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
