package githubconnector

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/nixopus/nixopus/api/internal/features/github-connector/types"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
)

func TestCreateGithubConnector(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()

	validReq := types.CreateGithubConnectorRequest{
		AppID:         "12345",
		Slug:          "my-github-app",
		Pem:           "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEA...\n-----END RSA PRIVATE KEY-----",
		ClientID:      "Iv1.abc123def456",
		ClientSecret:  "secret",
		WebhookSecret: "webhook-secret",
	}

	testCases := []struct {
		name           string
		cookies        string
		organizationID string
		request        types.CreateGithubConnectorRequest
		expectedStatus int
		description    string
	}{
		{
			name:           "No authentication",
			cookies:        "",
			organizationID: orgID,
			request:        validReq,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no auth cookies provided",
		},
		{
			name:           "Invalid cookies",
			cookies:        "bad-cookie=bad-value",
			organizationID: orgID,
			request:        validReq,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when cookies are invalid",
		},
		{
			name:           "No organization header",
			cookies:        cookies,
			organizationID: "",
			request:        validReq,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 without organization header",
		},
		{
			name:           "Missing app_id and no global config",
			cookies:        cookies,
			organizationID: orgID,
			request:        types.CreateGithubConnectorRequest{Slug: "slug", Pem: "pem"},
			expectedStatus: http.StatusInternalServerError,
			description:    "Should return 500 when app_id is missing and no global config is set",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			steps := []IStep{
				Description(tc.description),
				Post(tests.GetGithubConnectorURL()),
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
