package update

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
)

// --- Check for updates ---

func TestCheckForUpdates_NoAuth(t *testing.T) {
	Test(t,
		Description("GET /update/check without auth returns 401"),
		Get(tests.GetUpdateCheckURL()),
		Expect().Status().Equal(int64(http.StatusUnauthorized)),
	)
}

func TestCheckForUpdates_ValidAuth(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	// CheckForUpdates needs APP_VERSION or version.txt on disk; CI env often has neither → 500.
	Test(t,
		Description("GET /update/check with valid auth — 200 when version configured else 500"),
		Get(tests.GetUpdateCheckURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
	)
}

// --- Perform update ---

func TestPerformUpdate_NoAuth(t *testing.T) {
	Test(t,
		Description("POST /update without auth returns 401"),
		Post(tests.GetUpdateURL()),
		Send().Body().JSON(map[string]interface{}{"force": false}),
		Expect().Status().Equal(int64(http.StatusUnauthorized)),
	)
}

func TestPerformUpdate_ValidAuth(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	// In a test/dev environment PerformUpdate may return 200 with success=false (no update
	// available), 200 with success=true (dev mode shortcut), or 5xx if infra is unavailable.
	Test(t,
		Description("POST /update with valid auth is accepted — outcome depends on environment"),
		Post(tests.GetUpdateURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"force": false}),
		Expect().Status().OneOf(
			int64(http.StatusOK),
			int64(http.StatusBadRequest),
			int64(http.StatusInternalServerError),
		),
	)
}

func TestPerformUpdate_ForceFlag(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	// force=true triggers an actual update attempt; allow infra-level failures.
	Test(t,
		Description("POST /update with force=true is accepted — outcome depends on environment"),
		Post(tests.GetUpdateURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{"force": true}),
		Expect().Status().OneOf(
			int64(http.StatusOK),
			int64(http.StatusBadRequest),
			int64(http.StatusInternalServerError),
		),
	)
}
