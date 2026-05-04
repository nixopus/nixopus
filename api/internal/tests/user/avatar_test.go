package user

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
)

const validAvatarData = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="

func TestUpdateAvatar_success(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PATCH /user/avatar with valid base64 image returns 200"),
		Method("PATCH", tests.GetUserAvatarURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"avatarData": validAvatarData,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestUpdateAvatar_noAuth(t *testing.T) {
	Test(t,
		Description("PATCH /user/avatar without auth returns 401"),
		Method("PATCH", tests.GetUserAvatarURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(map[string]interface{}{
			"avatarData": validAvatarData,
		}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestUpdateAvatar_emptyAvatarData(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PATCH /user/avatar with empty avatarData returns 400"),
		Method("PATCH", tests.GetUserAvatarURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"avatarData": "",
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestUpdateAvatar_invalidFormat(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PATCH /user/avatar with non-data-URI returns 400"),
		Method("PATCH", tests.GetUserAvatarURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"avatarData": "not-a-valid-data-uri",
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestUpdateAvatar_unsupportedImageType(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PATCH /user/avatar with unsupported image format returns 400"),
		Method("PATCH", tests.GetUserAvatarURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"avatarData": "data:image/bmp;base64,abc",
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestUpdateAvatar_responseShape(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PATCH /user/avatar success response has status and message"),
		Method("PATCH", tests.GetUserAvatarURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"avatarData": validAvatarData,
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").NotEqual(nil),
		Expect().Body().JSON().JQ(".message").NotEqual(nil),
	)
}
