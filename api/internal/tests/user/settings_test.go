package user

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
)

// — GET /user/settings —

func TestGetSettings_success(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /user/settings returns settings for authenticated user"),
		Get(tests.GetUserSettingsURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
		Expect().Body().JSON().JQ(".data").NotEqual(nil),
	)
}

func TestGetSettings_noAuth(t *testing.T) {
	Test(t,
		Description("GET /user/settings without auth returns 401"),
		Get(tests.GetUserSettingsURL()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestGetSettings_responseShape(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /user/settings response includes font_family and theme fields"),
		Get(tests.GetUserSettingsURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".data.font_family").NotEqual(nil),
		Expect().Body().JSON().JQ(".data.theme").NotEqual(nil),
	)
}

// — PATCH /user/settings/font —

func TestUpdateFont_success(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PATCH /user/settings/font returns updated settings"),
		Method("PATCH", tests.GetUserSettingsFontURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"font_family": "JetBrains Mono",
			"font_size":   14,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
		Expect().Body().JSON().JQ(".data.font_family").Equal("JetBrains Mono"),
		Expect().Body().JSON().JQ(".data.font_size").Equal(float64(14)),
	)
}

func TestUpdateFont_noAuth(t *testing.T) {
	Test(t,
		Description("PATCH /user/settings/font without auth returns 401"),
		Method("PATCH", tests.GetUserSettingsFontURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(map[string]interface{}{
			"font_family": "monospace",
			"font_size":   14,
		}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestUpdateFont_missingBody(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PATCH /user/settings/font with empty body returns 400 or 422"),
		Method("PATCH", tests.GetUserSettingsFontURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{}),
		Expect().Status().Between(http.StatusBadRequest, http.StatusUnprocessableEntity),
	)
}

// — PATCH /user/settings/theme —

func TestUpdateTheme_success(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PATCH /user/settings/theme returns updated settings"),
		Method("PATCH", tests.GetUserSettingsThemeURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"theme": "dark",
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
		Expect().Body().JSON().JQ(".data.theme").Equal("dark"),
	)
}

func TestUpdateTheme_noAuth(t *testing.T) {
	Test(t,
		Description("PATCH /user/settings/theme without auth returns 401"),
		Method("PATCH", tests.GetUserSettingsThemeURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(map[string]interface{}{"theme": "dark"}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

// — PATCH /user/settings/language —

func TestUpdateLanguage_success(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PATCH /user/settings/language returns updated settings"),
		Method("PATCH", tests.GetUserSettingsLanguageURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"language": "fr",
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
		Expect().Body().JSON().JQ(".data.language").Equal("fr"),
	)
}

func TestUpdateLanguage_noAuth(t *testing.T) {
	Test(t,
		Description("PATCH /user/settings/language without auth returns 401"),
		Method("PATCH", tests.GetUserSettingsLanguageURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(map[string]interface{}{"language": "fr"}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

// — PATCH /user/settings/auto-update —

func TestUpdateAutoUpdate_enable(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PATCH /user/settings/auto-update enables auto-update"),
		Method("PATCH", tests.GetUserSettingsAutoUpdateURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"auto_update": true,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
		Expect().Body().JSON().JQ(".data.auto_update").Equal(true),
	)
}

func TestUpdateAutoUpdate_disable(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PATCH /user/settings/auto-update disables auto-update"),
		Method("PATCH", tests.GetUserSettingsAutoUpdateURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"auto_update": false,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestUpdateAutoUpdate_noAuth(t *testing.T) {
	Test(t,
		Description("PATCH /user/settings/auto-update without auth returns 401"),
		Method("PATCH", tests.GetUserSettingsAutoUpdateURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(map[string]interface{}{"auto_update": true}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}
