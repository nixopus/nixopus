package user

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
)

// — GET /user/preferences —

func TestGetPreferences_success(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /user/preferences returns preferences for authenticated user"),
		Get(tests.GetUserPreferencesURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
		Expect().Body().JSON().JQ(".data").NotEqual(nil),
	)
}

func TestGetPreferences_noAuth(t *testing.T) {
	Test(t,
		Description("GET /user/preferences without auth returns 401"),
		Get(tests.GetUserPreferencesURL()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestGetPreferences_responseShape(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /user/preferences response contains preferences data"),
		Get(tests.GetUserPreferencesURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").NotEqual(nil),
		Expect().Body().JSON().JQ(".message").NotEqual(nil),
		Expect().Body().JSON().JQ(".data.preferences").NotEqual(nil),
	)
}

// — PUT /user/preferences —

func TestUpdatePreferences_success(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PUT /user/preferences updates user preferences"),
		Put(tests.GetUserPreferencesURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"debug_mode":             true,
			"show_api_error_details": false,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
		Expect().Body().JSON().JQ(".data").NotEqual(nil),
	)
}

func TestUpdatePreferences_noAuth(t *testing.T) {
	Test(t,
		Description("PUT /user/preferences without auth returns 401"),
		Put(tests.GetUserPreferencesURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(map[string]interface{}{
			"debug_mode": true,
		}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestUpdatePreferences_terminalSettings(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PUT /user/preferences updates terminal-related preferences"),
		Put(tests.GetUserPreferencesURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"debug_mode":             false,
			"show_api_error_details": false,
			"terminal_font_size":     16,
			"terminal_cursor_style":  "block",
			"terminal_cursor_blink":  true,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestUpdatePreferences_emptyBody(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PUT /user/preferences with empty body still returns 200 (all fields have defaults)"),
		Put(tests.GetUserPreferencesURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{}),
		Expect().Status().Equal(http.StatusOK),
	)
}
