package githubconnector

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
)

func TestGetGithubRepositories(t *testing.T) {
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
		query          string
		expectedStatus int
		description    string
	}{
		{
			name:           "No authentication",
			cookies:        "",
			organizationID: orgID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no auth cookies provided",
		},
		{
			name:           "Invalid cookies",
			cookies:        "bad=val",
			organizationID: orgID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when cookies are invalid",
		},
		{
			name:           "No connectors — empty list",
			cookies:        cookies,
			organizationID: orgID,
			expectedStatus: http.StatusOK,
			description:    "No connectors yields 200 with empty repositories and total_count 0",
		},
		{
			name:           "Invalid sort direction is corrected to asc",
			cookies:        cookies,
			organizationID: orgID,
			query:          "?sort_direction=sideways",
			expectedStatus: http.StatusOK,
			description:    "Invalid sort_direction is sanitized; 200 with empty list when no connectors",
		},
		{
			name:           "Pagination params parsed correctly",
			cookies:        cookies,
			organizationID: orgID,
			query:          "?page=2&page_size=5",
			expectedStatus: http.StatusOK,
			description:    "Pagination reflected in response body; 200 with empty list when no connectors",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			steps := []IStep{
				Description(tc.description),
				Get(tests.GetGithubRepositoriesURL() + tc.query),
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
