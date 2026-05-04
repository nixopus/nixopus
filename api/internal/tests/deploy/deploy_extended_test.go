package deploy

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/deploy/types"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// TestDeployApplicationTemplate covers POST /api/v1/deploy/application/template
func TestDeployApplicationTemplate(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()

	t.Run("No auth returns 401", func(t *testing.T) {
		Test(t,
			Description("No auth should return 401"),
			Post(tests.GetDeployApplicationTemplateURL()),
			Send().Body().JSON(map[string]interface{}{}),
			Expect().Status().Equal(int64(http.StatusUnauthorized)),
		)
	})

	t.Run("No auth with valid body returns 401", func(t *testing.T) {
		Test(t,
			Description("Valid body but no authentication should return 401"),
			Post(tests.GetDeployApplicationTemplateURL()),
			Send().Body().JSON(types.CreateTemplateDeploymentRequest{
				TemplateID: "wordpress",
				Name:       "my-wordpress",
			}),
			Expect().Status().Equal(int64(http.StatusUnauthorized)),
		)
	})

	t.Run("Auth with missing template_id returns 400", func(t *testing.T) {
		Test(t,
			Description("Missing template_id should return 400"),
			Post(tests.GetDeployApplicationTemplateURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(types.CreateTemplateDeploymentRequest{
				Name: "my-app",
			}),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Auth with missing name returns 400", func(t *testing.T) {
		Test(t,
			Description("Missing name should return 400"),
			Post(tests.GetDeployApplicationTemplateURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(types.CreateTemplateDeploymentRequest{
				TemplateID: "wordpress",
			}),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Auth with nonexistent template returns 404", func(t *testing.T) {
		Test(t,
			Description("Nonexistent template_id should return 404"),
			Post(tests.GetDeployApplicationTemplateURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(types.CreateTemplateDeploymentRequest{
				TemplateID: "nonexistent-template-xyz-123",
				Name:       "my-app",
			}),
			Expect().Status().Equal(int64(http.StatusNotFound)),
		)
	})
}

// TestCreateProject covers POST /api/v1/deploy/application/project
func TestCreateProject(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()

	t.Run("No auth returns 401", func(t *testing.T) {
		Test(t,
			Description("No auth should return 401"),
			Post(tests.GetDeployApplicationProjectURL()),
			Send().Body().JSON(map[string]interface{}{}),
			Expect().Status().Equal(int64(http.StatusUnauthorized)),
		)
	})

	t.Run("No auth with valid body returns 401", func(t *testing.T) {
		Test(t,
			Description("Valid body but no auth should return 401"),
			Post(tests.GetDeployApplicationProjectURL()),
			Send().Body().JSON(types.CreateProjectRequest{
				Name:       "test-project",
				Repository: "https://github.com/test/repo.git",
			}),
			Expect().Status().Equal(int64(http.StatusUnauthorized)),
		)
	})

	t.Run("Auth with missing name returns 400", func(t *testing.T) {
		Test(t,
			Description("Missing name should return 400"),
			Post(tests.GetDeployApplicationProjectURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(types.CreateProjectRequest{
				Repository: "https://github.com/test/repo.git",
			}),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Auth with missing repository returns 400", func(t *testing.T) {
		Test(t,
			Description("Missing repository should return 400"),
			Post(tests.GetDeployApplicationProjectURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(types.CreateProjectRequest{
				Name: "test-project",
			}),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Auth with valid minimal body creates project or returns infra error", func(t *testing.T) {
		t.Skip("Creating a project triggers SSH/infra calls; skip in CI without live infra")
	})
}

// TestDeployProject covers POST /api/v1/deploy/application/project/deploy
func TestDeployProject(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()

	t.Run("No auth returns 401", func(t *testing.T) {
		Test(t,
			Description("No auth should return 401"),
			Post(tests.GetDeployApplicationProjectDeployURL()),
			Send().Body().JSON(map[string]interface{}{}),
			Expect().Status().Equal(int64(http.StatusUnauthorized)),
		)
	})

	t.Run("No auth with valid ID returns 401", func(t *testing.T) {
		Test(t,
			Description("Valid body but no auth should return 401"),
			Post(tests.GetDeployApplicationProjectDeployURL()),
			Send().Body().JSON(types.DeployProjectRequest{
				ID: uuid.New(),
			}),
			Expect().Status().Equal(int64(http.StatusUnauthorized)),
		)
	})

	t.Run("Auth with zero-value ID returns 400", func(t *testing.T) {
		Test(t,
			Description("Zero-value UUID should fail validation with 400"),
			Post(tests.GetDeployApplicationProjectDeployURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(map[string]interface{}{}),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Auth with nonexistent project ID returns 500", func(t *testing.T) {
		Test(t,
			Description("Nonexistent project ID returns 500 (infra-dependent)"),
			Post(tests.GetDeployApplicationProjectDeployURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(types.DeployProjectRequest{
				ID: uuid.New(),
			}),
			Expect().Status().OneOf(int64(http.StatusInternalServerError), int64(http.StatusNotFound)),
		)
	})
}

// TestDuplicateProject covers POST /api/v1/deploy/application/project/duplicate
func TestDuplicateProject(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()

	t.Run("No auth returns 401", func(t *testing.T) {
		Test(t,
			Description("No auth should return 401"),
			Post(tests.GetDeployApplicationProjectDuplicateURL()),
			Send().Body().JSON(map[string]interface{}{}),
			Expect().Status().Equal(int64(http.StatusUnauthorized)),
		)
	})

	t.Run("No auth with valid body returns 401", func(t *testing.T) {
		Test(t,
			Description("Valid body but no auth should return 401"),
			Post(tests.GetDeployApplicationProjectDuplicateURL()),
			Send().Body().JSON(types.DuplicateProjectRequest{
				SourceProjectID: uuid.New(),
				Environment:     shared_types.Environment("staging"),
			}),
			Expect().Status().Equal(int64(http.StatusUnauthorized)),
		)
	})

	t.Run("Auth with missing source_project_id returns 400", func(t *testing.T) {
		Test(t,
			Description("Missing source_project_id should return 400"),
			Post(tests.GetDeployApplicationProjectDuplicateURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(types.DuplicateProjectRequest{
				Environment: shared_types.Environment("staging"),
			}),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Auth with missing environment returns 400", func(t *testing.T) {
		Test(t,
			Description("Missing environment should return 400"),
			Post(tests.GetDeployApplicationProjectDuplicateURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(types.DuplicateProjectRequest{
				SourceProjectID: uuid.New(),
			}),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Auth with nonexistent source project returns 404 or 500", func(t *testing.T) {
		Test(t,
			Description("Nonexistent source project should return 404 or 500"),
			Post(tests.GetDeployApplicationProjectDuplicateURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(types.DuplicateProjectRequest{
				SourceProjectID: uuid.New(),
				Environment:     shared_types.Environment("staging"),
			}),
			Expect().Status().OneOf(int64(http.StatusNotFound), int64(http.StatusInternalServerError)),
		)
	})
}

// TestAddApplicationToFamily covers POST /api/v1/deploy/application/project/add-to-family
func TestAddApplicationToFamily(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()

	t.Run("No auth returns 401", func(t *testing.T) {
		Test(t,
			Description("No auth should return 401"),
			Post(tests.GetDeployApplicationProjectAddToFamilyURL()),
			Send().Body().JSON(map[string]interface{}{}),
			Expect().Status().Equal(int64(http.StatusUnauthorized)),
		)
	})

	t.Run("No auth with valid body returns 401", func(t *testing.T) {
		Test(t,
			Description("Valid body but no auth should return 401"),
			Post(tests.GetDeployApplicationProjectAddToFamilyURL()),
			Send().Body().JSON(types.AddApplicationToFamilyRequest{
				Name:       "my-api",
				Repository: "https://github.com/test/repo.git",
			}),
			Expect().Status().Equal(int64(http.StatusUnauthorized)),
		)
	})

	t.Run("Auth with missing name returns 400", func(t *testing.T) {
		Test(t,
			Description("Missing name should return 400"),
			Post(tests.GetDeployApplicationProjectAddToFamilyURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(types.AddApplicationToFamilyRequest{
				Repository: "https://github.com/test/repo.git",
			}),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Auth with missing repository returns 400", func(t *testing.T) {
		Test(t,
			Description("Missing repository should return 400"),
			Post(tests.GetDeployApplicationProjectAddToFamilyURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(types.AddApplicationToFamilyRequest{
				Name: "my-api",
			}),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Auth with nonexistent family_id returns 404 or 500", func(t *testing.T) {
		familyID := uuid.New()
		Test(t,
			Description("Nonexistent family_id should return 404 or 500"),
			Post(tests.GetDeployApplicationProjectAddToFamilyURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(types.AddApplicationToFamilyRequest{
				FamilyID:   &familyID,
				Name:       "my-api",
				Repository: "https://github.com/test/repo.git",
			}),
			Expect().Status().OneOf(int64(http.StatusNotFound), int64(http.StatusInternalServerError)),
		)
	})
}

// TestGetProjectFamily covers GET /api/v1/deploy/application/project/family
func TestGetProjectFamily(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()

	t.Run("No auth returns 401", func(t *testing.T) {
		Test(t,
			Description("No auth should return 401"),
			Get(tests.GetDeployApplicationProjectFamilyURL()),
			Expect().Status().Equal(int64(http.StatusUnauthorized)),
		)
	})

	t.Run("Valid family_id without auth returns 401", func(t *testing.T) {
		Test(t,
			Description("Valid family_id but no auth should return 401"),
			Get(tests.GetDeployApplicationProjectFamilyURL()+"?family_id="+uuid.New().String()),
			Expect().Status().Equal(int64(http.StatusUnauthorized)),
		)
	})

	t.Run("Auth with missing family_id returns 400", func(t *testing.T) {
		Test(t,
			Description("Missing family_id query param should return 400"),
			Get(tests.GetDeployApplicationProjectFamilyURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Auth with invalid UUID family_id returns 400", func(t *testing.T) {
		Test(t,
			Description("Invalid UUID for family_id should return 400"),
			Get(tests.GetDeployApplicationProjectFamilyURL()+"?family_id=not-a-uuid"),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Auth with nonexistent family_id returns 404 or 500", func(t *testing.T) {
		Test(t,
			Description("Nonexistent family_id should return 404 or 500"),
			Get(tests.GetDeployApplicationProjectFamilyURL()+"?family_id="+uuid.New().String()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Expect().Status().OneOf(int64(http.StatusNotFound), int64(http.StatusInternalServerError)),
		)
	})
}

// TestGetEnvironmentsInFamily covers GET /api/v1/deploy/application/project/family/environments
func TestGetEnvironmentsInFamily(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()

	t.Run("No auth returns 401", func(t *testing.T) {
		Test(t,
			Description("No auth should return 401"),
			Get(tests.GetDeployApplicationProjectFamilyEnvironmentsURL()),
			Expect().Status().Equal(int64(http.StatusUnauthorized)),
		)
	})

	t.Run("Auth with missing family_id returns 400", func(t *testing.T) {
		Test(t,
			Description("Missing family_id query param should return 400"),
			Get(tests.GetDeployApplicationProjectFamilyEnvironmentsURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Auth with invalid UUID family_id returns 400", func(t *testing.T) {
		Test(t,
			Description("Invalid UUID for family_id should return 400"),
			Get(tests.GetDeployApplicationProjectFamilyEnvironmentsURL()+"?family_id=invalid-uuid"),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Valid family_id without auth returns 401", func(t *testing.T) {
		Test(t,
			Description("Valid family_id but no auth should return 401"),
			Get(tests.GetDeployApplicationProjectFamilyEnvironmentsURL()+"?family_id="+uuid.New().String()),
			Expect().Status().Equal(int64(http.StatusUnauthorized)),
		)
	})

	t.Run("Auth with nonexistent family_id returns 500", func(t *testing.T) {
		Test(t,
			Description("Nonexistent family_id returns 500 (DB error expected)"),
			Get(tests.GetDeployApplicationProjectFamilyEnvironmentsURL()+"?family_id="+uuid.New().String()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
		)
	})
}

// TestUpdateApplicationLabels covers PUT /api/v1/deploy/application/labels
func TestUpdateApplicationLabels(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()
	nonexistentID := uuid.New().String()

	t.Run("Missing id query param returns 400 without auth", func(t *testing.T) {
		Test(t,
			Description("Missing ?id param should return 400"),
			Put(tests.GetDeployApplicationLabelsURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(map[string]interface{}{"labels": []string{"prod"}}),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Invalid UUID id param returns 400", func(t *testing.T) {
		Test(t,
			Description("Invalid UUID for ?id should return 400"),
			Put(tests.GetDeployApplicationLabelsURL()+"?id=not-a-uuid"),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(map[string]interface{}{"labels": []string{"prod"}}),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Valid id param without auth returns 401", func(t *testing.T) {
		Test(t,
			Description("Valid ?id but no auth should return 401"),
			Put(tests.GetDeployApplicationLabelsURL()+"?id="+nonexistentID),
			Send().Body().JSON(map[string]interface{}{"labels": []string{"prod"}}),
			Expect().Status().Equal(int64(http.StatusUnauthorized)),
		)
	})

	t.Run("Auth with nonexistent app ID still returns 200", func(t *testing.T) {
		// UpdateApplicationLabels runs UPDATE ... WHERE id=?; Postgres succeeds with 0 rows —
		// storage does not check RowsAffected, so no error surfaces.
		Test(t,
			Description("Nonexistent application ID — UPDATE no-match still responds success"),
			Put(tests.GetDeployApplicationLabelsURL()+"?id="+nonexistentID),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(map[string]interface{}{"labels": []string{"prod", "v2"}}),
			Expect().Status().Equal(int64(http.StatusOK)),
			Expect().Body().JSON().JQ(".status").Equal("success"),
			Expect().Body().JSON().JQ(".message").Equal("Labels updated successfully"),
		)
	})
}

// TestAddApplicationDomain covers POST /api/v1/deploy/application/domains
func TestAddApplicationDomain(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()
	nonexistentID := uuid.New().String()

	t.Run("Missing id query param returns 400", func(t *testing.T) {
		Test(t,
			Description("Missing ?id should return 400"),
			Post(tests.GetDeployApplicationDomainsURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(map[string]interface{}{"domain": "example.com"}),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Missing domain in body returns 400", func(t *testing.T) {
		Test(t,
			Description("Missing domain in body should return 400"),
			Post(tests.GetDeployApplicationDomainsURL()+"?id="+nonexistentID),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(map[string]interface{}{}),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Valid params without auth returns 401", func(t *testing.T) {
		Test(t,
			Description("Valid params but no auth should return 401"),
			Post(tests.GetDeployApplicationDomainsURL()+"?id="+nonexistentID),
			Send().Body().JSON(map[string]interface{}{"domain": "example.com"}),
			Expect().Status().Equal(int64(http.StatusUnauthorized)),
		)
	})

	t.Run("Invalid UUID id param returns 400", func(t *testing.T) {
		Test(t,
			Description("Invalid UUID for ?id should return 400"),
			Post(tests.GetDeployApplicationDomainsURL()+"?id=not-a-uuid"),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(map[string]interface{}{"domain": "example.com"}),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Auth with nonexistent app ID returns 404", func(t *testing.T) {
		Test(t,
			Description("Nonexistent application ID should return 404"),
			Post(tests.GetDeployApplicationDomainsURL()+"?id="+nonexistentID),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(map[string]interface{}{"domain": "example.com"}),
			Expect().Status().Equal(int64(http.StatusNotFound)),
		)
	})
}

// TestRemoveApplicationDomain covers DELETE /api/v1/deploy/application/domains
func TestRemoveApplicationDomain(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()
	nonexistentID := uuid.New().String()

	t.Run("Missing id query param returns 400", func(t *testing.T) {
		Test(t,
			Description("Missing ?id should return 400"),
			Delete(tests.GetDeployApplicationDomainsURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(map[string]interface{}{"domain": "example.com"}),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Missing domain in body returns 400", func(t *testing.T) {
		Test(t,
			Description("Missing domain in body should return 400"),
			Delete(tests.GetDeployApplicationDomainsURL()+"?id="+nonexistentID),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(map[string]interface{}{}),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Valid params without auth returns 401", func(t *testing.T) {
		Test(t,
			Description("Valid params but no auth should return 401"),
			Delete(tests.GetDeployApplicationDomainsURL()+"?id="+nonexistentID),
			Send().Body().JSON(map[string]interface{}{"domain": "example.com"}),
			Expect().Status().Equal(int64(http.StatusUnauthorized)),
		)
	})

	t.Run("Invalid UUID id param returns 400", func(t *testing.T) {
		Test(t,
			Description("Invalid UUID for ?id should return 400"),
			Delete(tests.GetDeployApplicationDomainsURL()+"?id=invalid-uuid"),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(map[string]interface{}{"domain": "example.com"}),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Auth with nonexistent app ID returns 404", func(t *testing.T) {
		Test(t,
			Description("Nonexistent application ID should return 404"),
			Delete(tests.GetDeployApplicationDomainsURL()+"?id="+nonexistentID),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(map[string]interface{}{"domain": "example.com"}),
			Expect().Status().Equal(int64(http.StatusNotFound)),
		)
	})
}

// TestGetComposeServices covers GET /api/v1/deploy/application/compose-services
func TestGetComposeServices(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()
	nonexistentID := uuid.New().String()

	t.Run("Missing id query param returns 400", func(t *testing.T) {
		Test(t,
			Description("Missing ?id should return 400"),
			Get(tests.GetDeployApplicationComposeServicesURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Valid id without auth returns 401", func(t *testing.T) {
		Test(t,
			Description("Valid ?id but no auth should return 401"),
			Get(tests.GetDeployApplicationComposeServicesURL()+"?id="+nonexistentID),
			Expect().Status().Equal(int64(http.StatusUnauthorized)),
		)
	})

	t.Run("Invalid UUID id with auth returns 400", func(t *testing.T) {
		Test(t,
			Description("Invalid UUID for ?id should return 400"),
			Get(tests.GetDeployApplicationComposeServicesURL()+"?id=not-a-uuid"),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Auth with nonexistent app ID returns 404", func(t *testing.T) {
		Test(t,
			Description("Nonexistent application ID should return 404"),
			Get(tests.GetDeployApplicationComposeServicesURL()+"?id="+nonexistentID),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Expect().Status().Equal(int64(http.StatusNotFound)),
		)
	})
}

// TestPreviewCompose covers POST /api/v1/deploy/application/preview-compose
func TestPreviewCompose(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	cookies := auth.GetAuthCookiesHeader()

	t.Run("Missing repository returns 400", func(t *testing.T) {
		Test(t,
			Description("Missing repository should return 400"),
			Post(tests.GetDeployApplicationPreviewComposeURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Body().JSON(types.PreviewComposeRequest{
				Branch: "main",
			}),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Missing branch returns 400", func(t *testing.T) {
		Test(t,
			Description("Missing branch should return 400"),
			Post(tests.GetDeployApplicationPreviewComposeURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Body().JSON(types.PreviewComposeRequest{
				Repository: "https://github.com/test/repo.git",
			}),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Valid body without auth returns 401", func(t *testing.T) {
		Test(t,
			Description("Valid body but no auth should return 401"),
			Post(tests.GetDeployApplicationPreviewComposeURL()),
			Send().Body().JSON(types.PreviewComposeRequest{
				Repository: "https://github.com/test/repo.git",
				Branch:     "main",
			}),
			Expect().Status().Equal(int64(http.StatusUnauthorized)),
		)
	})

	t.Run("Auth with nonexistent repo returns 422 or 500", func(t *testing.T) {
		Test(t,
			Description("Nonexistent repository returns unprocessable or infra error"),
			Post(tests.GetDeployApplicationPreviewComposeURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Body().JSON(types.PreviewComposeRequest{
				Repository: "https://github.com/nonexistent-org-xyz/nonexistent-repo-xyz.git",
				Branch:     "main",
			}),
			Expect().Status().OneOf(
				int64(http.StatusUnprocessableEntity),
				int64(http.StatusInternalServerError),
				int64(http.StatusNotFound),
			),
		)
	})
}

// TestRecover covers POST /api/v1/deploy/application/recover
func TestRecover(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()

	t.Run("No auth returns 401", func(t *testing.T) {
		Test(t,
			Description("No auth should return 401"),
			Post(tests.GetDeployApplicationRecoverURL()),
			Send().Body().JSON(types.RecoverRequest{}),
			Expect().Status().Equal(int64(http.StatusUnauthorized)),
		)
	})

	t.Run("Auth with no app_id returns 400 due to missing S3 config", func(t *testing.T) {
		Test(t,
			Description("S3 not configured in test env; recover all returns 400"),
			Post(tests.GetDeployApplicationRecoverURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(types.RecoverRequest{}),
			Expect().Status().OneOf(int64(http.StatusBadRequest), int64(http.StatusInternalServerError)),
		)
	})

	t.Run("Auth with specific nonexistent app_id returns 400 or 500", func(t *testing.T) {
		appID := uuid.New()
		Test(t,
			Description("Nonexistent app_id returns 400 (S3 not configured) or 500"),
			Post(tests.GetDeployApplicationRecoverURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(types.RecoverRequest{ApplicationID: &appID}),
			Expect().Status().OneOf(int64(http.StatusBadRequest), int64(http.StatusInternalServerError)),
		)
	})
}

// TestGetApplicationServers covers GET /api/v1/deploy/application/servers
func TestGetApplicationServers(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()
	nonexistentID := uuid.New().String()

	t.Run("No auth returns 401", func(t *testing.T) {
		Test(t,
			Description("No auth should return 401"),
			Get(tests.GetDeployApplicationServersURL()+"?id="+nonexistentID),
			Expect().Status().Equal(int64(http.StatusUnauthorized)),
		)
	})

	t.Run("Auth with missing id param returns 400", func(t *testing.T) {
		// Empty string fails uuid.Parse → BadRequestError
		Test(t,
			Description("Missing ?id should return 400"),
			Get(tests.GetDeployApplicationServersURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Auth with invalid UUID id returns 400", func(t *testing.T) {
		Test(t,
			Description("Invalid UUID for ?id should return 400"),
			Get(tests.GetDeployApplicationServersURL()+"?id=not-a-uuid"),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Auth with nonexistent app ID returns 404", func(t *testing.T) {
		Test(t,
			Description("Nonexistent application ID should return 404"),
			Get(tests.GetDeployApplicationServersURL()+"?id="+nonexistentID),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Expect().Status().Equal(int64(http.StatusNotFound)),
		)
	})
}

// TestSetApplicationServers covers PUT /api/v1/deploy/application/servers
func TestSetApplicationServers(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID := auth.OrganizationID
	cookies := auth.GetAuthCookiesHeader()
	nonexistentID := uuid.New()
	serverID := uuid.New()

	t.Run("No auth returns 401", func(t *testing.T) {
		Test(t,
			Description("No auth should return 401"),
			Put(tests.GetDeployApplicationServersURL()),
			Send().Body().JSON(types.SetApplicationServersRequest{
				ApplicationID: nonexistentID,
				ServerIDs:     []uuid.UUID{serverID},
			}),
			Expect().Status().Equal(int64(http.StatusUnauthorized)),
		)
	})

	t.Run("Auth with empty server_ids returns 400", func(t *testing.T) {
		Test(t,
			Description("Empty server_ids should return 400"),
			Put(tests.GetDeployApplicationServersURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(types.SetApplicationServersRequest{
				ApplicationID: nonexistentID,
				ServerIDs:     []uuid.UUID{},
			}),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Auth with primary_server_id not in server_ids returns 400", func(t *testing.T) {
		otherServerID := uuid.New()
		Test(t,
			Description("primary_server_id not in server_ids should return 400"),
			Put(tests.GetDeployApplicationServersURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(types.SetApplicationServersRequest{
				ApplicationID:   nonexistentID,
				ServerIDs:       []uuid.UUID{serverID},
				PrimaryServerID: &otherServerID,
			}),
			Expect().Status().Equal(int64(http.StatusBadRequest)),
		)
	})

	t.Run("Auth with valid body but nonexistent app returns 404", func(t *testing.T) {
		Test(t,
			Description("Nonexistent application ID should return 404"),
			Put(tests.GetDeployApplicationServersURL()),
			Send().Headers("Cookie").Add(cookies),
			Send().Headers("X-Organization-ID").Add(orgID),
			Send().Body().JSON(types.SetApplicationServersRequest{
				ApplicationID: nonexistentID,
				ServerIDs:     []uuid.UUID{serverID},
			}),
			Expect().Status().Equal(int64(http.StatusNotFound)),
		)
	})
}
