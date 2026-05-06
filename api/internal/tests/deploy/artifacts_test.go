package deploy

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
)

func TestListArtifacts(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()
	testAppID := "123e4567-e89b-12d3-a456-426614174000"

	testCases := []struct {
		name           string
		cookies        string
		organizationID string
		applicationID  string
		expectedStatus int
		description    string
	}{
		{
			name:           "List artifacts without authentication",
			cookies:        "",
			organizationID: orgID,
			applicationID:  testAppID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no auth cookies are provided",
		},
		{
			name:           "List artifacts with invalid cookies",
			cookies:        "invalid-cookies",
			organizationID: orgID,
			applicationID:  testAppID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when invalid auth cookies are provided",
		},
		{
			name:           "List artifacts without application_id",
			cookies:        cookies,
			organizationID: orgID,
			applicationID:  "",
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when application_id query param is missing",
		},
		{
			name:           "List artifacts with invalid application_id",
			cookies:        cookies,
			organizationID: orgID,
			applicationID:  "not-a-uuid",
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 when application_id format is invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSteps := []IStep{
				Description(tc.description),
				Get(tests.GetDeployArtifactsURL(tc.applicationID)),
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

func TestListArtifactsWithSeededData(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()

	appID, err := setup.SeedApplication(auth.User.ID.String(), orgID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	t.Run("List artifacts for app with no artifacts", func(t *testing.T) {
		Test(t,
			Description("Should return empty list when app has no artifacts"),
			Get(tests.GetDeployArtifactsURL(appID)),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".status").Equal("success"),
			Expect().Body().JSON().JQ(".data").Equal([]interface{}{}),
		)
	})

	_, err = setup.SeedDeploymentWithArtifact(appID, orgID+"/"+appID+"/test.tar.gz", 1024)
	if err != nil {
		t.Fatalf("failed to seed deployment with artifact: %v", err)
	}

	t.Run("List artifacts for app with one artifact", func(t *testing.T) {
		Test(t,
			Description("Should return artifact list with correct fields"),
			Get(tests.GetDeployArtifactsURL(appID)),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".status").Equal("success"),
			Expect().Body().JSON().JQ(".data | length").Equal(1.0),
			Expect().Body().JSON().JQ(".data[0].application_id").Equal(appID),
			Expect().Body().JSON().JQ(".data[0].size").Equal(1024.0),
			Expect().Body().JSON().JQ(".data[0].s3_key").NotEqual(""),
		)
	})
}

func TestGetArtifactDownloadURL(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()
	fakeDeploymentID := "123e4567-e89b-12d3-a456-426614174000"

	testCases := []struct {
		name           string
		cookies        string
		organizationID string
		deploymentID   string
		expectedStatus int
		description    string
	}{
		{
			name:           "Download artifact without authentication",
			cookies:        "",
			organizationID: orgID,
			deploymentID:   fakeDeploymentID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no auth cookies are provided",
		},
		{
			name:           "Download artifact with invalid cookies",
			cookies:        "invalid-cookies",
			organizationID: orgID,
			deploymentID:   fakeDeploymentID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when invalid auth cookies are provided",
		},
		{
			name:           "Download artifact for non-existent deployment",
			cookies:        cookies,
			organizationID: orgID,
			deploymentID:   fakeDeploymentID,
			expectedStatus: http.StatusNotFound,
			description:    "Should return 404 when deployment doesn't exist",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSteps := []IStep{
				Description(tc.description),
				Get(tests.GetDeployArtifactDownloadURL(tc.deploymentID)),
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

func TestGetArtifactDownloadURLNoS3Key(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()

	appID, err := setup.SeedApplication(auth.User.ID.String(), orgID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	deployID, err := setup.SeedDeploymentWithoutArtifact(appID)
	if err != nil {
		t.Fatalf("failed to seed deployment: %v", err)
	}

	t.Run("Download artifact when deployment has no S3 key", func(t *testing.T) {
		Test(t,
			Description("Should return 400 when deployment has no artifact uploaded"),
			Get(tests.GetDeployArtifactDownloadURL(deployID)),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Expect().Status().Equal(http.StatusBadRequest),
		)
	})
}

func TestDeleteArtifact(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()
	fakeDeploymentID := "123e4567-e89b-12d3-a456-426614174000"

	testCases := []struct {
		name           string
		cookies        string
		organizationID string
		deploymentID   string
		expectedStatus int
		description    string
	}{
		{
			name:           "Delete artifact without authentication",
			cookies:        "",
			organizationID: orgID,
			deploymentID:   fakeDeploymentID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when no auth cookies are provided",
		},
		{
			name:           "Delete artifact with invalid cookies",
			cookies:        "invalid-cookies",
			organizationID: orgID,
			deploymentID:   fakeDeploymentID,
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 when invalid auth cookies are provided",
		},
		{
			name:           "Delete artifact for non-existent deployment",
			cookies:        cookies,
			organizationID: orgID,
			deploymentID:   fakeDeploymentID,
			expectedStatus: http.StatusNotFound,
			description:    "Should return 404 when deployment doesn't exist",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSteps := []IStep{
				Description(tc.description),
				Delete(tests.GetDeployArtifactDeleteURL(tc.deploymentID)),
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

func TestDeleteArtifactNoS3Key(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()

	appID, err := setup.SeedApplication(auth.User.ID.String(), orgID)
	if err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}

	deployID, err := setup.SeedDeploymentWithoutArtifact(appID)
	if err != nil {
		t.Fatalf("failed to seed deployment: %v", err)
	}

	t.Run("Delete artifact when deployment has no artifact", func(t *testing.T) {
		Test(t,
			Description("Should return 400 when deployment has no artifact to delete"),
			Delete(tests.GetDeployArtifactDeleteURL(deployID)),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Expect().Status().Equal(http.StatusBadRequest),
		)
	})
}

// Security: Cross-org isolation tests

func TestArtifactsCrossOrgIsolation(t *testing.T) {
	setup := testutils.NewTestSetup()

	// Create user A (org A)
	authA, err := setup.CreateTestUserViaAuth("user-a@example.com", "User A")
	if err != nil {
		t.Fatalf("failed to create user A: %v", err)
	}
	orgA := authA.OrganizationID
	cookiesA := authA.GetAuthCookiesHeader()

	// Seed an application and artifact owned by org A
	appID, err := setup.SeedApplication(authA.User.ID.String(), orgA)
	if err != nil {
		t.Fatalf("failed to seed application for org A: %v", err)
	}

	deployID, err := setup.SeedDeploymentWithArtifact(appID, orgA+"/"+appID+"/test.tar.gz", 2048)
	if err != nil {
		t.Fatalf("failed to seed deployment with artifact: %v", err)
	}

	// Create user B (org B)
	authB, err := setup.CreateTestUserViaAuth("user-b@example.com", "User B")
	if err != nil {
		t.Fatalf("failed to create user B: %v", err)
	}
	orgB := authB.OrganizationID
	cookiesB := authB.GetAuthCookiesHeader()

	// Verify org A can see its own artifacts
	t.Run("Org A can list its own artifacts", func(t *testing.T) {
		Test(t,
			Description("Owner org should see its artifacts"),
			Get(tests.GetDeployArtifactsURL(appID)),
			Send().Headers("Cookie").Add(cookiesA),
			Send().Headers("X-Organization-ID").Add(orgA),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".data | length").Equal(1.0),
		)
	})

	// Cross-org: user B tries to list artifacts for org A's app
	t.Run("Org B cannot list org A artifacts", func(t *testing.T) {
		Test(t,
			Description("Cross-org list should fail with 403"),
			Get(tests.GetDeployArtifactsURL(appID)),
			Send().Headers("Cookie").Add(cookiesB),
			Send().Headers("X-Organization-ID").Add(orgB),
			Expect().Status().Equal(http.StatusForbidden),
		)
	})

	// Cross-org: user B tries to download org A's artifact
	t.Run("Org B cannot download org A artifact", func(t *testing.T) {
		Test(t,
			Description("Cross-org download should fail with 403"),
			Get(tests.GetDeployArtifactDownloadURL(deployID)),
			Send().Headers("Cookie").Add(cookiesB),
			Send().Headers("X-Organization-ID").Add(orgB),
			Expect().Status().Equal(http.StatusForbidden),
		)
	})

	// Cross-org: user B tries to delete org A's artifact
	t.Run("Org B cannot delete org A artifact", func(t *testing.T) {
		Test(t,
			Description("Cross-org delete should fail with 403"),
			Delete(tests.GetDeployArtifactDeleteURL(deployID)),
			Send().Headers("Cookie").Add(cookiesB),
			Send().Headers("X-Organization-ID").Add(orgB),
			Expect().Status().Equal(http.StatusForbidden),
		)
	})

	// Verify org A's artifact still exists after org B's failed delete attempt
	t.Run("Org A artifact still intact after cross-org delete attempt", func(t *testing.T) {
		Test(t,
			Description("Artifact should still be visible to owner after failed cross-org delete"),
			Get(tests.GetDeployArtifactsURL(appID)),
			Send().Headers("Cookie").Add(cookiesA),
			Send().Headers("X-Organization-ID").Add(orgA),
			Expect().Status().Equal(http.StatusOK),
			Expect().Body().JSON().JQ(".data | length").Equal(1.0),
		)
	})
}
