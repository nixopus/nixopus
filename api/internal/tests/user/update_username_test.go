package user

import (
	"net/http"
	"strings"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
)

func TestUpdateUsername_success(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PATCH /user/name with valid username returns 200"),
		Method("PATCH", tests.GetUserNameURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"name": "newusername",
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
		Expect().Body().JSON().JQ(".data.name").Equal("newusername"),
	)
}

func TestUpdateUsername_noAuth(t *testing.T) {
	Test(t,
		Description("PATCH /user/name without auth returns 401"),
		Method("PATCH", tests.GetUserNameURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(map[string]interface{}{
			"name": "newusername",
		}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestUpdateUsername_invalidCookie(t *testing.T) {
	Test(t,
		Description("PATCH /user/name with invalid cookie returns 401"),
		Method("PATCH", tests.GetUserNameURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Cookie").Add("better-auth.session_token=invalid"),
		Send().Body().JSON(map[string]interface{}{
			"name": "newusername",
		}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestUpdateUsername_emptyName(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PATCH /user/name with empty name returns 400"),
		Method("PATCH", tests.GetUserNameURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"name": "",
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestUpdateUsername_nameTooShort(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PATCH /user/name with name shorter than 3 chars returns 400"),
		Method("PATCH", tests.GetUserNameURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"name": "ab",
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestUpdateUsername_nameTooLong(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PATCH /user/name with name longer than 50 chars returns 400"),
		Method("PATCH", tests.GetUserNameURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"name": strings.Repeat("a", 51),
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestUpdateUsername_nameWithSpaces(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PATCH /user/name with spaces in name returns 400"),
		Method("PATCH", tests.GetUserNameURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"name": "user name",
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestUpdateUsername_responseShape(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PATCH /user/name success response has expected fields"),
		Method("PATCH", tests.GetUserNameURL()),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(map[string]interface{}{
			"name": "validname",
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").NotEqual(nil),
		Expect().Body().JSON().JQ(".message").NotEqual(nil),
		Expect().Body().JSON().JQ(".data").NotEqual(nil),
	)
}
