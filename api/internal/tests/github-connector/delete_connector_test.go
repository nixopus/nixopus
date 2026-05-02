package githubconnector

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/github-connector/types"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
)

func TestDeleteGithubConnector(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()
	randomID := uuid.New().String()

	testCases := []struct {
		name           string
		cookies        string
		organizationID string
		request        types.DeleteGithubConnectorRequest
		expectedStatus int
		description    string
	}{
		{
			name:           "No authentication",
			cookies:        "",
			organizationID: orgID,
			request:        types.DeleteGithubConnectorRequest{ID: randomID},
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no auth cookies provided",
		},
		{
			name:           "Invalid cookies",
			cookies:        "bad=val",
			organizationID: orgID,
			request:        types.DeleteGithubConnectorRequest{ID: randomID},
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when cookies are invalid",
		},
		{
			name:           "Missing ID field",
			cookies:        cookies,
			organizationID: orgID,
			request:        types.DeleteGithubConnectorRequest{ID: ""},
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when id is empty",
		},
		{
			name:           "Non-existent connector",
			cookies:        cookies,
			organizationID: orgID,
			request:        types.DeleteGithubConnectorRequest{ID: randomID},
			expectedStatus: http.StatusNotFound,
			description:    "Should return 404 when connector does not exist",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			steps := []IStep{
				Description(tc.description),
				Delete(tests.GetGithubConnectorURL()),
				Send().Body().JSON(tc.request),
			}
			if tc.cookies != "" {
				steps = append(steps, Send().Headers("Cookie").Add(tc.cookies))
			}
			if tc.organizationID != "" {
				steps = append(steps, Send().Headers("X-Organization-ID").Add(tc.organizationID))
			}
			steps = append(steps, Expect().Status().Equal(int64(tc.expectedStatus)))
			Test(t, steps...)
		})
	}
}
