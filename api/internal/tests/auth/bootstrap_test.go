package auth

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
)

func TestBootstrapHTTP_NoAuth(t *testing.T) {
	Test(t,
		Description("GET /auth/bootstrap without auth returns 401"),
		Get(tests.GetAuthBootstrapURL()),
		Expect().Status().Equal(int64(http.StatusUnauthorized)),
	)
}

func TestBootstrapHTTP_ValidAuth(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /auth/bootstrap with valid auth returns 200 with user and organizations"),
		Get(tests.GetAuthBootstrapURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(int64(http.StatusOK)),
		Expect().Body().JSON().JQ(".user").NotEqual(nil),
		Expect().Body().JSON().JQ(".organizations").NotEqual(nil),
	)
}
