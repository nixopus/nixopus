package machine

import (
	"encoding/json"
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
	api_types "github.com/nixopus/nixopus/api/internal/types"
)

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

func insertSSHKeyHelper(t *testing.T, setup *testutils.TestSetup, key *api_types.SSHKey) {
	t.Helper()
	_, err := setup.DB.NewInsert().Model(key).Exec(setup.Ctx)
	if err != nil {
		t.Fatalf("failed to insert SSH key: %v", err)
	}
}

func createMachineHelper(t *testing.T, auth *testutils.TestAuthResponse, name, host string, port int, user string) types.CreateMachineResponse {
	t.Helper()
	var created types.CreateMachineResponse
	Test(t,
		Post(tests.GetMachinesURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(types.CreateMachineRequest{
			Name: name,
			Host: host,
			Port: port,
			User: user,
		}),
		Expect().Status().Equal(http.StatusCreated),
		Expect().Custom(func(hit Hit) error {
			return json.Unmarshal(hit.Response().Body().MustBytes(), &created)
		}),
	)
	return created
}
