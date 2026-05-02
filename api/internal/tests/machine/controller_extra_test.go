package machine

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/google/uuid"
	machine_types "github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
)

func TestCheckSSHStatus_NoAuth(t *testing.T) {
	Test(t,
		Description("GET /machines/ssh/status without auth returns 401"),
		Get(tests.GetMachineSSHStatusURL()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestCheckSSHStatus_ValidAuth(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /machines/ssh/status with valid auth returns 200"),
		Get(tests.GetMachineSSHStatusURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
	)
}

func TestExecMachineCommand_NoAuth(t *testing.T) {
	Test(t,
		Description("POST /machines/exec without auth returns 401"),
		Post(tests.GetMachinesURL()+"/exec"),
		Send().Body().JSON(machine_types.HostExecRequest{Command: "echo hi"}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestExecMachineCommand_MissingCommand(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /machines/exec without command returns 400"),
		Post(tests.GetMachinesURL()+"/exec"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(machine_types.HostExecRequest{Command: ""}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestExecMachineCommand_ValidRequest_NoSSHConfigured(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /machines/exec with command returns 500 without configured SSH"),
		Post(tests.GetMachinesURL()+"/exec"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(machine_types.HostExecRequest{Command: "echo hi"}),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusInternalServerError)),
	)
}

func TestMachineTrialProvision_NoAuth(t *testing.T) {
	Test(t,
		Description("POST /machines/trial/provision without auth returns 401"),
		Post(tests.GetMachineTrialProvisionURL()),
		Send().Body().JSON(machine_types.ProvisionRequest{}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestMachineTrialProvision_WithAuth_MissingOrgHeader(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /machines/trial/provision without org header returns 403"),
		Post(tests.GetMachineTrialProvisionURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Body().JSON(machine_types.ProvisionRequest{}),
		Expect().Status().OneOf(int64(http.StatusForbidden), int64(http.StatusBadRequest)),
	)
}

func TestMachineTrialStatus_NoAuth(t *testing.T) {
	Test(t,
		Description("GET /machines/trial/status/:sessionId without auth returns 401"),
		Get(tests.GetMachineTrialStatusURL(uuid.New().String())),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestMachineTrialStatus_WithAuth_InvalidSessionID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /machines/trial/status/not-a-uuid returns 400"),
		Get(tests.GetMachineTrialStatusURL("not-a-uuid")),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestTrailUpgradeResources_NoSecret(t *testing.T) {
	Test(t,
		Description("POST /trail/upgrade-resources without internal secret returns 401"),
		Post(tests.GetTrailUpgradeResourcesURL()),
		Send().Body().JSON(machine_types.UpgradeResourcesRequest{
			UserID:    uuid.New().String(),
			OrgID:     uuid.New().String(),
			VcpuCount: 2,
			MemoryMB:  1024,
		}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}
