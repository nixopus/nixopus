package machine

import (
	"net/http"
	"testing"
	"time"

	. "github.com/Eun/go-hit"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
	api_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/require"
)

// --- Create machine ---

func TestCreateMachine_NoAuth(t *testing.T) {
	Test(t,
		Description("POST /machines without auth returns 401"),
		Post(tests.GetMachinesURL()),
		Send().Body().JSON(types.CreateMachineRequest{
			Name: "test-server",
			Host: "192.0.2.10",
			Port: 22,
			User: "root",
		}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestCreateMachine_MissingName(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /machines without name returns 400"),
		Post(tests.GetMachinesURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(types.CreateMachineRequest{
			Host: "192.0.2.10",
			Port: 22,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestCreateMachine_MissingHost(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /machines without host returns 400"),
		Post(tests.GetMachinesURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(types.CreateMachineRequest{
			Name: "test-server",
			Port: 22,
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestCreateMachine_DuplicateHost(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID, err := uuid.Parse(auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to parse org ID: %v", err)
	}

	existingKey := &api_types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           "existing-machine",
		Host:           strPtr("10.0.0.99"),
		Port:           intPtr(22),
		AuthMethod:     "key",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, insertErr := setup.DB.NewInsert().Model(existingKey).Exec(setup.Ctx)
	require.NoError(t, insertErr)

	Test(t,
		Description("POST /machines with duplicate host:port returns 400"),
		Post(tests.GetMachinesURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(types.CreateMachineRequest{
			Name: "another-machine",
			Host: "10.0.0.99",
			Port: 22,
			User: "ubuntu",
		}),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

// --- Verify machine ---

func TestVerifyMachine_NoAuth(t *testing.T) {
	machineID := uuid.New().String()
	Test(t,
		Description("POST /machines/:id/verify without auth returns 401"),
		Post(tests.GetMachineVerifyURL(machineID)),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestVerifyMachine_InvalidID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /machines/not-a-uuid/verify returns 400"),
		Post(tests.GetMachineVerifyURL("not-a-uuid")),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestVerifyMachine_NotFound(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /machines/:id/verify for non-existent machine returns 500"),
		Post(tests.GetMachineVerifyURL(uuid.New().String())),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusInternalServerError),
	)
}

func TestVerifyMachine_MissingKeyData(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID, err := uuid.Parse(auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to parse org ID: %v", err)
	}

	keyID := uuid.New()
	key := &api_types.SSHKey{
		ID:             keyID,
		OrganizationID: orgID,
		Name:           "no-key-machine",
		Host:           strPtr("192.0.2.50"),
		Port:           intPtr(22),
		AuthMethod:     "key",
		IsActive:       false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	insertSSHKeyHelper(t, setup, key)

	Test(t,
		Description("POST /machines/:id/verify with missing private key returns failed status"),
		Post(tests.GetMachineVerifyURL(keyID.String())),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("failed"),
		Expect().Body().JSON().JQ(".is_active").Equal(false),
	)
}

func TestVerifyMachine_UnreachableHost(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID, err := uuid.Parse(auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to parse org ID: %v", err)
	}

	created := createMachineHelper(t, auth, "unreachable-machine", "192.0.2.1", 22, "root")

	keyID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	// Confirm the key exists in the DB with a private key
	var sshKey api_types.SSHKey
	require.NoError(t, setup.DB.NewSelect().
		Model(&sshKey).
		Where("id = ?", keyID).
		Where("organization_id = ?", orgID).
		Scan(setup.Ctx))
	require.NotNil(t, sshKey.PrivateKeyEncrypted, "key should have private key after creation")

	Test(t,
		Description("POST /machines/:id/verify with unreachable host returns failed status"),
		Post(tests.GetMachineVerifyURL(keyID.String())),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("failed"),
		Expect().Body().JSON().JQ(".is_active").Equal(false),
	)
}

// --- Delete machine ---

func TestDeleteMachine_NoAuth(t *testing.T) {
	machineID := uuid.New().String()
	Test(t,
		Description("DELETE /machines/:id without auth returns 401"),
		Delete(tests.GetMachineByIDURL(machineID)),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestDeleteMachine_InvalidID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("DELETE /machines/not-a-uuid returns 400"),
		Delete(tests.GetMachineByIDURL("not-a-uuid")),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestDeleteMachine_NotFound(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("DELETE /machines/:id for non-existent machine returns 500 (no active apps check fails gracefully)"),
		Delete(tests.GetMachineByIDURL(uuid.New().String())),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusNotFound), int64(http.StatusInternalServerError)),
	)
}

func TestDeleteMachine_ValidMachine(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID, err := uuid.Parse(auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to parse org ID: %v", err)
	}

	keyID := uuid.New()
	key := &api_types.SSHKey{
		ID:             keyID,
		OrganizationID: orgID,
		Name:           "deletable-machine",
		Host:           strPtr("10.1.2.3"),
		Port:           intPtr(2222),
		AuthMethod:     "key",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, insertErr := setup.DB.NewInsert().Model(key).Exec(setup.Ctx)
	require.NoError(t, insertErr)

	Test(t,
		Description("DELETE /machines/:id for existing machine with no apps returns 200"),
		Delete(tests.GetMachineByIDURL(keyID.String())),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("deleted"),
	)

	// Verify it's soft-deleted (deleted_at set, not hard deleted)
	var softDeleted api_types.SSHKey
	queryErr := setup.DB.NewSelect().
		Model(&softDeleted).
		Where("id = ?", keyID).
		WhereAllWithDeleted().
		Scan(setup.Ctx)
	require.NoError(t, queryErr)
	require.NotNil(t, softDeleted.DeletedAt, "machine should be soft-deleted")
}

// --- SSH key status ---

func TestGetSSHKeyStatus_NoAuth(t *testing.T) {
	machineID := uuid.New().String()
	Test(t,
		Description("GET /machines/:id/ssh/status without auth returns 401"),
		Get(tests.GetMachineSSHKeyStatusURL(machineID)),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestGetSSHKeyStatus_InvalidID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /machines/not-a-uuid/ssh/status returns 400"),
		Get(tests.GetMachineSSHKeyStatusURL("not-a-uuid")),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestGetSSHKeyStatus_ValidMachine(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID, err := uuid.Parse(auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to parse org ID: %v", err)
	}

	keyID := uuid.New()
	key := &api_types.SSHKey{
		ID:             keyID,
		OrganizationID: orgID,
		Name:           "status-check-machine",
		Host:           strPtr("10.2.3.4"),
		Port:           intPtr(22),
		AuthMethod:     "key",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, insertErr := setup.DB.NewInsert().Model(key).Exec(setup.Ctx)
	require.NoError(t, insertErr)

	Test(t,
		Description("GET /machines/:id/ssh/status returns active status"),
		Get(tests.GetMachineSSHKeyStatusURL(keyID.String())),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".is_active").Equal(true),
	)
}
