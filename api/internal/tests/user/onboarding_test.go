package user

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
)

// — GET /user/onboarded —

func TestGetIsOnboarded_success(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /user/onboarded returns onboarding status for authenticated user"),
		Get(tests.GetUserOnboardedURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
		Expect().Body().JSON().JQ(".data.is_onboarded").NotEqual(nil),
	)
}

func TestGetIsOnboarded_noAuth(t *testing.T) {
	Test(t,
		Description("GET /user/onboarded without auth returns 401"),
		Get(tests.GetUserOnboardedURL()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestGetIsOnboarded_newUser_notOnboarded(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /user/onboarded for a freshly created user returns false"),
		Get(tests.GetUserOnboardedURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".data.is_onboarded").Equal(false),
	)
}

func TestGetIsOnboarded_responseShape(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /user/onboarded response has expected fields"),
		Get(tests.GetUserOnboardedURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").NotEqual(nil),
		Expect().Body().JSON().JQ(".message").NotEqual(nil),
		Expect().Body().JSON().JQ(".data").NotEqual(nil),
	)
}

// — POST /user/onboarded —

func TestMarkOnboardingComplete_success(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /user/onboarded marks user as onboarded"),
		Post(tests.GetUserOnboardedURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".data.is_onboarded").Equal(true),
	)
}

func TestMarkOnboardingComplete_noAuth(t *testing.T) {
	Test(t,
		Description("POST /user/onboarded without auth returns 401"),
		Post(tests.GetUserOnboardedURL()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestMarkOnboardingComplete_idempotent(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	cookies := auth.GetAuthCookiesHeader()
	orgID := auth.OrganizationID

	Test(t,
		Description("POST /user/onboarded first call succeeds"),
		Post(tests.GetUserOnboardedURL()),
		Send().Headers("Cookie").Add(cookies),
		Send().Headers("X-Organization-ID").Add(orgID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".data.is_onboarded").Equal(true),
	)

	Test(t,
		Description("POST /user/onboarded second call is idempotent"),
		Post(tests.GetUserOnboardedURL()),
		Send().Headers("Cookie").Add(cookies),
		Send().Headers("X-Organization-ID").Add(orgID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".data.is_onboarded").Equal(true),
	)
}

func TestMarkOnboardingComplete_thenGet(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	cookies := auth.GetAuthCookiesHeader()
	orgID := auth.OrganizationID

	Test(t,
		Description("POST /user/onboarded marks user; subsequent GET reflects the change"),
		Post(tests.GetUserOnboardedURL()),
		Send().Headers("Cookie").Add(cookies),
		Send().Headers("X-Organization-ID").Add(orgID),
		Expect().Status().Equal(http.StatusOK),
	)

	Test(t,
		Description("GET /user/onboarded after marking complete returns true"),
		Get(tests.GetUserOnboardedURL()),
		Send().Headers("Cookie").Add(cookies),
		Send().Headers("X-Organization-ID").Add(orgID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".data.is_onboarded").Equal(true),
	)
}
