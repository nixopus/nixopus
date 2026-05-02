package githubconnector

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
)

type getBranchesRequest struct {
	RepositoryName string `json:"repository_name"`
}

func TestGetGithubRepositoryBranches(t *testing.T) {
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
		request        getBranchesRequest
		expectedStatus int
		description    string
	}{
		{
			name:           "No authentication",
			cookies:        "",
			organizationID: orgID,
			request:        getBranchesRequest{RepositoryName: "owner/repo"},
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no auth cookies provided",
		},
		{
			name:           "Invalid cookies",
			cookies:        "bad=val",
			organizationID: orgID,
			request:        getBranchesRequest{RepositoryName: "owner/repo"},
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when cookies are invalid",
		},
		{
			name:           "Empty repository_name",
			cookies:        cookies,
			organizationID: orgID,
			request:        getBranchesRequest{RepositoryName: ""},
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when repository_name is empty",
		},
		{
			name:           "No connectors — service error",
			cookies:        cookies,
			organizationID: orgID,
			request:        getBranchesRequest{RepositoryName: "owner/some-repo"},
			expectedStatus: http.StatusInternalServerError,
			description:    "Should return 500 when no GitHub connectors are configured for the user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			steps := []IStep{
				Description(tc.description),
				Post(tests.GetGithubRepositoryBranchesURL()),
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
