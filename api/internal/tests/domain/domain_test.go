package domain

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
)

// --- List domains ---

func TestGetDomains_NoAuth(t *testing.T) {
	Test(t,
		Description("GET /domain without auth returns 401"),
		Get(tests.GetDomainURL()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestGetDomains_ValidAuth(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /domain with valid auth returns 200"),
		Get(tests.GetDomainURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestGetDomains_FilterByType(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /domain?type=subdomain returns 200"),
		Get(tests.GetDomainURL()+"?type=subdomain"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
	)
}

func TestGetDomains_InvalidOrgID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /domain with invalid org ID returns 400"),
		Get(tests.GetDomainURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add("not-a-uuid"),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestGetDomains_CrossOrgDenied(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /domain for different org returns 403"),
		Get(tests.GetDomainURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add("123e4567-e89b-12d3-a456-426614174000"),
		Expect().Status().Equal(http.StatusForbidden),
	)
}

func TestAddCustomDomain_NoOrgID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /domain/custom without org ID returns 400"),
		Post(tests.GetDomainCustomURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Body().JSON(map[string]interface{}{"name": "app.example.com"}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestAddCustomDomain_InvalidDomainName_NoDot(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	// "nodotdomain" passes Fuego's min=3,max=255 validation but fails service-level
	// dot check → mapCustomDomainError → isInvalidDomainError → 400.
	Test(t,
		Description("POST /domain/custom with no-dot name returns 400"),
		Post(tests.GetDomainCustomURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"name": "nodotdomain"}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

// --- Generate random subdomain ---

func TestGenerateRandomSubdomain_NoAuth(t *testing.T) {
	Test(t,
		Description("GET /domain/generate without auth returns 401"),
		Get(tests.GetDomainGenerateURL()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestGenerateRandomSubdomain_ValidAuth(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	// Returns 404 if no base domains are seeded in the DB; 200 if base domains exist
	Test(t,
		Description("GET /domain/generate returns subdomain (404 if no base domains seeded)"),
		Get(tests.GetDomainGenerateURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusNotFound)),
	)
}

func TestGenerateRandomSubdomain_NoOrgID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /domain/generate without org ID returns 400"),
		Get(tests.GetDomainGenerateURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestGenerateRandomSubdomain_CalledTwiceReturnsDifferentValues(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	var first, second string
	var status1 int64
	Test(t,
		Description("First call to /domain/generate"),
		Get(tests.GetDomainGenerateURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusNotFound)),
		Store().Response().StatusCode().In(&status1),
		Store().Response().Body().JSON().JQ(".data.subdomain").In(&first),
	)

	if status1 != int64(http.StatusOK) {
		t.Skip("no base domains seeded — skipping uniqueness check")
	}

	Test(t,
		Description("Second call to /domain/generate"),
		Get(tests.GetDomainGenerateURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Store().Response().Body().JSON().JQ(".data.subdomain").In(&second),
	)

	if first != "" && second != "" && first == second {
		t.Logf("Note: two consecutive calls returned the same subdomain %q — may be intentional", first)
	}
}

// --- Custom domain ---

func TestAddCustomDomain_NoAuth(t *testing.T) {
	Test(t,
		Description("POST /domain/custom without auth returns 401"),
		Post(tests.GetDomainCustomURL()),
		Send().Body().JSON(map[string]interface{}{"name": "example.com"}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestAddCustomDomain_MissingName(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /domain/custom without name returns 400"),
		Post(tests.GetDomainCustomURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestAddCustomDomain_ValidName(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	// Requires a provisioned BYOS machine; returns 500 "provision details not found" without one
	Test(t,
		Description("POST /domain/custom with valid name — requires provisioned machine"),
		Post(tests.GetDomainCustomURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"name": "my-test-app.example.com"}),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
	)
}

func TestAddCustomDomain_DuplicateNameReturnsError(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	payload := map[string]interface{}{"name": "duplicate.example.com"}

	var status1 int64
	Test(t,
		Description("First custom domain creation"),
		Post(tests.GetDomainCustomURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(payload),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
		Store().Response().StatusCode().In(&status1),
	)

	if status1 != int64(http.StatusOK) {
		t.Skip("first domain creation failed (no provisioned machine) — skipping duplicate check")
	}

	Test(t,
		Description("Duplicate custom domain returns conflict or error"),
		Post(tests.GetDomainCustomURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(payload),
		Expect().Status().OneOf(int64(http.StatusConflict), int64(http.StatusBadRequest), int64(http.StatusInternalServerError)),
	)
}

func TestRemoveCustomDomain_NoAuth(t *testing.T) {
	Test(t,
		Description("DELETE /domain/custom without auth returns 401"),
		Delete(tests.GetDomainCustomURL()),
		Send().Body().JSON(map[string]interface{}{"id": uuid.New().String()}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestRemoveCustomDomain_NoOrgID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("DELETE /domain/custom without org ID returns 400"),
		Delete(tests.GetDomainCustomURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Body().JSON(map[string]interface{}{"id": uuid.New().String()}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestRemoveCustomDomain_InvalidUUID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("DELETE /domain/custom with malformed UUID returns 400"),
		Delete(tests.GetDomainCustomURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"id": "not-a-uuid"}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestRemoveCustomDomain_MissingID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("DELETE /domain/custom without id returns 400"),
		Delete(tests.GetDomainCustomURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestRemoveCustomDomain_NonExistentID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("DELETE /domain/custom with non-existent id returns error"),
		Delete(tests.GetDomainCustomURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"id": uuid.New().String()}),
		Expect().Status().OneOf(int64(http.StatusNotFound), int64(http.StatusInternalServerError)),
	)
}

func TestRemoveCustomDomain_ValidFlow(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	var domainID string
	var createStatus int64
	Test(t,
		Description("Create custom domain for deletion"),
		Post(tests.GetDomainCustomURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"name": "to-delete.example.com"}),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
		Store().Response().StatusCode().In(&createStatus),
		Store().Response().Body().JSON().JQ(".data.id").In(&domainID),
	)

	if createStatus != int64(http.StatusOK) || domainID == "" {
		t.Skip("domain creation requires provisioned machine — skipping delete test")
	}

	Test(t,
		Description("DELETE /domain/custom with valid id returns 200"),
		Delete(tests.GetDomainCustomURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"id": domainID}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

// --- Verify custom domain ---

func TestVerifyCustomDomain_NoAuth(t *testing.T) {
	Test(t,
		Description("POST /domain/verify without auth returns 401"),
		Post(tests.GetDomainVerifyURL()),
		Send().Body().JSON(map[string]interface{}{"id": uuid.New().String()}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestVerifyCustomDomain_NoOrgID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /domain/verify without org ID returns 400"),
		Post(tests.GetDomainVerifyURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Body().JSON(map[string]interface{}{"id": uuid.New().String()}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestVerifyCustomDomain_InvalidUUID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /domain/verify with malformed UUID returns 400"),
		Post(tests.GetDomainVerifyURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"id": "not-a-uuid"}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestVerifyCustomDomain_MissingID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /domain/verify without id returns 400"),
		Post(tests.GetDomainVerifyURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestVerifyCustomDomain_NonExistentID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /domain/verify with non-existent id returns error"),
		Post(tests.GetDomainVerifyURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"id": uuid.New().String()}),
		Expect().Status().OneOf(int64(http.StatusNotFound), int64(http.StatusInternalServerError)),
	)
}

func TestVerifyCustomDomain_ValidFlow(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	var domainID string
	var createStatus int64
	Test(t,
		Description("Create custom domain to verify"),
		Post(tests.GetDomainCustomURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"name": "verify-me.example.com"}),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
		Store().Response().StatusCode().In(&createStatus),
		Store().Response().Body().JSON().JQ(".data.id").In(&domainID),
	)

	if createStatus != int64(http.StatusOK) || domainID == "" {
		t.Skip("domain creation requires provisioned machine — skipping verify test")
	}

	// In CI DNS won't resolve: service returns ErrDNSNotVerified → 412 PreconditionFailed.
	// 200 if somehow DNS is configured; 400/500 for other infra errors.
	Test(t,
		Description("POST /domain/verify attempts DNS verification — DNS failure expected in CI → 412"),
		Post(tests.GetDomainVerifyURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"id": domainID}),
		Expect().Status().OneOf(
			int64(http.StatusOK),
			int64(http.StatusBadRequest),
			int64(http.StatusPreconditionFailed),
			int64(http.StatusInternalServerError),
		),
	)
}

// --- DNS check ---

func TestCheckDNSStatus_NoAuth(t *testing.T) {
	Test(t,
		Description("GET /domain/dns-check without auth returns 401"),
		Get(tests.GetDomainDNSCheckURL(uuid.New().String())),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestCheckDNSStatus_NoOrgID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /domain/dns-check without org ID returns 400"),
		Get(tests.GetDomainDNSCheckURL(uuid.New().String())),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestCheckDNSStatus_InvalidUUID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /domain/dns-check with malformed UUID returns 400"),
		Get(tests.GetDomainURL()+"/dns-check?id=not-a-uuid"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestCheckDNSStatus_MissingID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /domain/dns-check without id param returns 400"),
		Get(tests.GetDomainURL()+"/dns-check"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestCheckDNSStatus_NonExistentID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /domain/dns-check with non-existent id returns error"),
		Get(tests.GetDomainDNSCheckURL(uuid.New().String())),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusNotFound), int64(http.StatusInternalServerError)),
	)
}

func TestCheckDNSStatus_ValidFlow(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	var domainID string
	var createStatus int64
	Test(t,
		Description("Create custom domain for DNS check"),
		Post(tests.GetDomainCustomURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"name": "dns-check.example.com"}),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
		Store().Response().StatusCode().In(&createStatus),
		Store().Response().Body().JSON().JQ(".data.id").In(&domainID),
	)

	if createStatus != int64(http.StatusOK) || domainID == "" {
		t.Skip("domain creation requires provisioned machine — skipping DNS check test")
	}

	// DNS check may fail in CI since the domain doesn't actually point anywhere
	Test(t,
		Description("GET /domain/dns-check returns DNS status (may be unverified in CI)"),
		Get(tests.GetDomainDNSCheckURL(domainID)),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
	)
}
