package audit

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
)

func TestGetAuditLogs_NoAuth(t *testing.T) {
	Test(t,
		Description("GET /audit/logs without auth returns 401"),
		Get(tests.GetAuditLogsURL()),
		Expect().Status().Equal(int64(http.StatusUnauthorized)),
	)
}

func TestGetAuditLogs_ValidAuth(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /audit/logs with valid auth returns 200 and success status"),
		Get(tests.GetAuditLogsURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(int64(http.StatusOK)),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestGetAuditLogs_WithPagination(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /audit/logs with pagination params returns 200"),
		Get(tests.GetAuditLogsURL()+"?page=1&page_size=5"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(int64(http.StatusOK)),
		Expect().Body().JSON().JQ(".status").Equal("success"),
		Expect().Body().JSON().JQ(".data.pagination.current_page").Equal(float64(1)),
	)
}

func TestGetAuditLogs_InvalidOrgID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /audit/logs with invalid org ID returns 400"),
		Get(tests.GetAuditLogsURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add("not-a-uuid"),
		Expect().Status().Equal(int64(http.StatusBadRequest)),
	)
}

func TestGetAuditLogs_MissingOrgID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /audit/logs without org ID header returns 400"),
		Get(tests.GetAuditLogsURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Expect().Status().Equal(int64(http.StatusBadRequest)),
	)
}
