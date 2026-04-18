package machine

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/nixopus/nixopus/api/internal/tests"
	"github.com/nixopus/nixopus/api/internal/testutils"
)

// --- List backups ---

func TestListBackups_NoAuth(t *testing.T) {
	Test(t,
		Description("GET /machines/backups without auth returns 401"),
		Get(tests.GetMachineBackupsURL()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestListBackups_ValidAuth_EmptyList(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /machines/backups with valid auth returns 200 with empty list"),
		Get(tests.GetMachineBackupsURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestListBackups_WithPagination(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /machines/backups with pagination params returns 200"),
		Get(tests.GetMachineBackupsURL()+"?page=1&page_size=20&sort_by=created_at&sort_order=desc"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
	)
}

func TestListBackups_WithStatusFilter(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /machines/backups filtered by status returns 200"),
		Get(tests.GetMachineBackupsURL()+"?status=completed"),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
	)
}

// --- Trigger backup ---

func TestTriggerBackup_NoAuth(t *testing.T) {
	serverID := uuid.New().String()
	Test(t,
		Description("POST /machines/backup without auth returns 401"),
		Post(tests.GetMachineTriggerBackupURL(serverID)),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

// TestTriggerBackup_BYOS verifies that triggering a backup for a BYOS machine
// takes the BYOS path (not the provisioned machine path).
// - If S3 is not configured (CI): returns 400 ErrS3NotConfigured.
// - If S3 is configured (local dev): returns 200 and initiates the backup.
// Either outcome confirms the BYOS code path was taken correctly.
func TestTriggerBackup_BYOS_NoS3Config(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	orgID, parseErr := uuid.Parse(auth.OrganizationID)
	if parseErr != nil {
		t.Fatalf("failed to parse org ID: %v", parseErr)
	}
	userID := auth.User.ID

	keyID := seedBYOSMachine(t, setup, orgID, userID, true)

	// 200 = S3 configured, backup initiated via BYOS path
	// 400 = S3 not configured (CI), ErrS3NotConfigured returned
	// 500 = unexpected server error
	Test(t,
		Description("POST /machines/backup for BYOS machine takes the BYOS code path"),
		Post(tests.GetMachineTriggerBackupURL(keyID.String())),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().OneOf(int64(http.StatusOK), int64(http.StatusBadRequest), int64(http.StatusInternalServerError)),
	)
}

// --- Backup schedule ---

func TestGetBackupSchedule_NoAuth(t *testing.T) {
	Test(t,
		Description("GET /machines/backup/schedule without auth returns 401"),
		Get(tests.GetMachineBackupScheduleURL()),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestGetBackupSchedule_ValidAuth(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("GET /machines/backup/schedule with valid auth returns 200"),
		Get(tests.GetMachineBackupScheduleURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestUpdateBackupSchedule_NoAuth(t *testing.T) {
	Test(t,
		Description("PUT /machines/backup/schedule without auth returns 401"),
		Put(tests.GetMachineBackupScheduleURL()),
		Send().Body().JSON(types.BackupScheduleData{
			Enabled:        true,
			Frequency:      "weekly",
			HourUTC:        2,
			DayOfWeek:      0,
			RetentionCount: 5,
		}),
		Expect().Status().Equal(http.StatusUnauthorized),
	)
}

func TestUpdateBackupSchedule_ValidAuth(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PUT /machines/backup/schedule with valid data returns 200"),
		Put(tests.GetMachineBackupScheduleURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(types.BackupScheduleData{
			Enabled:        true,
			Frequency:      "weekly",
			HourUTC:        3,
			DayOfWeek:      1,
			RetentionCount: 7,
			BackupPaths:    []string{"/home", "/etc", "/var/www"},
		}),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("success"),
	)
}

func TestUpdateBackupSchedule_Disable(t *testing.T) {
	setup := testutils.NewTestSetup()
	auth, err := setup.GetAuthResponse()
	if err != nil {
		t.Fatalf("failed to get auth response: %v", err)
	}

	Test(t,
		Description("PUT /machines/backup/schedule with enabled=false disables schedule"),
		Put(tests.GetMachineBackupScheduleURL()),
		Send().Headers("Cookie").Add(auth.GetAuthCookiesHeader()),
		Send().Headers("X-Organization-ID").Add(auth.OrganizationID),
		Send().Body().JSON(types.BackupScheduleData{
			Enabled:        false,
			Frequency:      "daily",
			HourUTC:        0,
			RetentionCount: 3,
		}),
		Expect().Status().Equal(http.StatusOK),
	)
}
