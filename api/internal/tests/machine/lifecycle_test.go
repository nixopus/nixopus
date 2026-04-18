package machine

import (
	"net/http"
	"testing"
	"time"

	. "github.com/Eun/go-hit"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
	api_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/require"
)

// --- Auth/validation guards ---

func TestGetMachineStatus_NoAuth(t *testing.T) {
	serverID := uuid.New().String()
	Test(t,
		Description("GET /machines/status without auth returns 401"),
		Get(tests.GetMachineStatusURL(serverID)),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestGetMachineStatus_MissingServerID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /machines/status without server_id returns 400"),
		Get(tests.GetMachinesURL()+"/status"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestPauseMachine_NoAuth(t *testing.T) {
	serverID := uuid.New().String()
	Test(t,
		Description("POST /machines/pause without auth returns 401"),
		Post(tests.GetMachinePauseURL(serverID)),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestPauseMachine_MissingServerID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /machines/pause without server_id returns 400"),
		Post(tests.GetMachinesURL()+"/pause"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestResumeMachine_NoAuth(t *testing.T) {
	serverID := uuid.New().String()
	Test(t,
		Description("POST /machines/resume without auth returns 401"),
		Post(tests.GetMachineResumeURL(serverID)),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestResumeMachine_MissingServerID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /machines/resume without server_id returns 400"),
		Post(tests.GetMachinesURL()+"/resume"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

func TestRestartMachine_NoAuth(t *testing.T) {
	serverID := uuid.New().String()
	Test(t,
		Description("POST /machines/restart without auth returns 401"),
		Post(tests.GetMachineRestartURL(serverID)),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestRestartMachine_MissingServerID(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("POST /machines/restart without server_id returns 400"),
		Post(tests.GetMachinesURL()+"/restart"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusBadRequest),
	)
}

// --- BYOS lifecycle tests (user_owned machines, DB-only operations) ---

// seedBYOSMachine inserts a user_owned SSH key and provision record, returning the SSH key ID.
// Pause and resume are pure DB flag flips, safe to run in CI without a real SSH host.
func seedBYOSMachine(t *testing.T, setup *testutils.TestSetup, orgID uuid.UUID, userID uuid.UUID, isActive bool) uuid.UUID {
	t.Helper()

	keyID := uuid.New()
	key := &api_types.SSHKey{
		ID:             keyID,
		OrganizationID: orgID,
		Name:           "byos-test-machine",
		Host:           strPtr("192.0.2.1"),
		Port:           intPtr(22),
		AuthMethod:     "key",
		IsActive:       isActive,
		IsDefault:      true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err := setup.DB.NewInsert().Model(key).Exec(setup.Ctx)
	require.NoError(t, err, "insert SSH key")

	step := api_types.ProvisionStepCompleted
	provision := &api_types.UserProvisionDetails{
		ID:             uuid.New(),
		UserID:         userID,
		OrganizationID: orgID,
		SSHKeyID:       &keyID,
		Type:           "user_owned",
		Step:           &step,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err = setup.DB.NewInsert().Model(provision).Exec(setup.Ctx)
	require.NoError(t, err, "insert provision details")

	return keyID
}

// TestBYOSPauseMachine verifies that pausing a user-owned machine sets is_active=false in the DB.
// This is a pure DB operation — no SSH required — so it passes in CI.
func TestBYOSPauseMachine_SetsMachineInactive(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID, err := uuid.Parse(auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to parse org ID: %v", err)
	}
	userID := auth.User.ID

	keyID := seedBYOSMachine(t, setup, orgID, userID, true)

	Test(t,
		Description("POST /machines/pause on user-owned machine returns 200"),
		Post(tests.GetMachinePauseURL(keyID.String())),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)

	var updated api_types.SSHKey
	err = setup.DB.NewSelect().Model(&updated).Where("id = ?", keyID).Scan(setup.Ctx)
	require.NoError(t, err)
	require.False(t, updated.IsActive, "machine should be inactive after pause")
}

// TestBYOSResumeMachine verifies that resuming a paused user-owned machine sets is_active=true.
// Pure DB flip — CI-safe.
func TestBYOSResumeMachine_SetsMachineActive(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID, err := uuid.Parse(auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to parse org ID: %v", err)
	}
	userID := auth.User.ID

	keyID := seedBYOSMachine(t, setup, orgID, userID, false)

	Test(t,
		Description("POST /machines/resume on paused user-owned machine returns 200"),
		Post(tests.GetMachineResumeURL(keyID.String())),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)

	var updated api_types.SSHKey
	err = setup.DB.NewSelect().Model(&updated).Where("id = ?", keyID).Scan(setup.Ctx)
	require.NoError(t, err)
	require.True(t, updated.IsActive, "machine should be active after resume")
}

// TestBYOSPauseResumeCycle verifies that repeated pause/resume operations are consistent.
func TestBYOSPauseResumeCycle(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID, err := uuid.Parse(auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to parse org ID: %v", err)
	}
	userID := auth.User.ID

	keyID := seedBYOSMachine(t, setup, orgID, userID, true)

	Test(t,
		Description("Pause → is_active becomes false"),
		Post(tests.GetMachinePauseURL(keyID.String())),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
	)

	var afterPause api_types.SSHKey
	require.NoError(t, setup.DB.NewSelect().Model(&afterPause).Where("id = ?", keyID).Scan(setup.Ctx))
	require.False(t, afterPause.IsActive)

	Test(t,
		Description("Resume → is_active becomes true"),
		Post(tests.GetMachineResumeURL(keyID.String())),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
	)

	var afterResume api_types.SSHKey
	require.NoError(t, setup.DB.NewSelect().Model(&afterResume).Where("id = ?", keyID).Scan(setup.Ctx))
	require.True(t, afterResume.IsActive)
}

// TestBYOSGetStatus_PausedMachine verifies GET /machines/status returns "Paused"
// for a user-owned machine with is_active=false, without needing SSH.
func TestBYOSGetStatus_PausedMachine(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID, err := uuid.Parse(auth.OrganizationID)
	if err != nil {
		t.Fatalf("failed to parse org ID: %v", err)
	}
	userID := auth.User.ID

	// Seed a paused (is_active=false) BYOS machine
	keyID := seedBYOSMachine(t, setup, orgID, userID, false)

	Test(t,
		Description("GET /machines/status for paused BYOS machine returns 200 with Paused state"),
		Get(tests.GetMachineStatusURL(keyID.String())),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
		Expect().Body().JSON().JQ(".data.state").Equal("Paused"),
	)
}
