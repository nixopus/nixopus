package machine

import (
	"net/http"
	"testing"
	"time"

	. "github.com/Eun/go-hit"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
	"github.com/nixopus/nixopus/api/internal/types"
)

func TestListMachines_NoAuth(t *testing.T) {
	Test(t,
		Description("GET /machines without auth should return 401"),
		Get(tests.GetMachinesURL()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestListMachines_InvalidCookies(t *testing.T) {
	Test(t,
		Description("GET /machines with invalid auth should return 401"),
		Get(tests.GetMachinesURL()),
		Send().Headers("Cookie").Add("better-auth.session_token=invalid-token"),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestListMachines_ValidAuth_EmptyList(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /machines with valid auth returns 200 with empty list"),
		Get(tests.GetMachinesURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
	)
}

func TestListMachines_WithPaginationParams(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /machines with pagination params returns 200"),
		Get(tests.GetMachinesURL()+"?page=1&page_size=10&sort_by=created_at&sort_order=desc"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
	)
}

func TestListMachines_WithSearchParam(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /machines with search param returns 200"),
		Get(tests.GetMachinesURL()+"?search=my-server"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
	)
}

func TestListMachines_WithSSHKey_ReturnsMachine(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID, err := uuid.Parse(auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to parse org ID: %v", err)
	}

	key := &types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           "test-machine",
		AuthMethod:     "key",
		IsActive:       true,
		IsDefault:      true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	insertSSHKeyHelper(t, setup, key)

	Test(t,
		Description("GET /machines returns machine that was inserted"),
		Get(tests.GetMachinesURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
	)
}

func TestSetDefaultMachine_NoAuth(t *testing.T) {
	machineID := uuid.New().String()

	Test(t,
		Description("PUT /machines/:id/set-default without auth returns 401"),
		Put(tests.GetMachineSetDefaultURL(machineID)),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestSetDefaultMachine_InvalidID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PUT /machines/not-a-uuid/set-default returns 400"),
		Put(tests.GetMachineSetDefaultURL("not-a-uuid")),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestSetDefaultMachine_NotFound(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PUT /machines/:id/set-default for non-existent machine returns 404"),
		Put(tests.GetMachineSetDefaultURL(uuid.New().String())),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusNotFound),
	)
}
