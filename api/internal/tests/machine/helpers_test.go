package machine

import (
	"testing"

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
