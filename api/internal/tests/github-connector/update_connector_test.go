package githubconnector

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/nixopus/nixopus/api/internal/features/github-connector/types"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
)

func TestUpdateGithubConnector(t *testing.T) {
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
		request        types.UpdateGithubConnectorRequest
		expectedStatus int
		description    string
	}{
		{
			name:           "No authentication",
			cookies:        "",
			organizationID: orgID,
			request:        types.UpdateGithubConnectorRequest{InstallationID: "12345"},
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no auth cookies provided",
		},
		{
			name:           "Invalid cookies",
			cookies:        "bad=val",
			organizationID: orgID,
			request:        types.UpdateGithubConnectorRequest{InstallationID: "12345"},
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when cookies are invalid",
		},
		{
			name:           "Missing installation_id",
			cookies:        cookies,
			organizationID: orgID,
			request:        types.UpdateGithubConnectorRequest{InstallationID: ""},
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when installation_id is empty",
		},
		{
			name:           "No connectors exist",
			cookies:        cookies,
			organizationID: orgID,
			request:        types.UpdateGithubConnectorRequest{InstallationID: "99999"},
			expectedStatus: http.StatusInternalServerError,
			description:    "Should return 500 when no connectors are found for the user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			steps := []IStep{
				Description(tc.description),
				Put(tests.GetGithubConnectorURL()),
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
