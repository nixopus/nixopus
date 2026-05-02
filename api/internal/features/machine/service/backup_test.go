package service_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/machine/service"
	"github.com/nixopus/nixopus/api/internal/features/machine/storage"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/nixopus/nixopus/api/internal/queue"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	_ "modernc.org/sqlite"
)

// mockBackupProvisionInfo implements service.ProvisionInfoProvider for backup tests.
type mockBackupProvisionInfo struct {
	info *storage.ProvisionInfo
	err  error
}

func (m *mockBackupProvisionInfo) GetProvisionInfo(_ context.Context, _ uuid.UUID, _ *uuid.UUID) (*storage.ProvisionInfo, error) {
	return m.info, m.err
}

// mockBackupStore implements service.BackupStoreInterface.
type mockBackupStore struct {
	hasInProgressFn       func(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID) (bool, error)
	listByOrgFn           func(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID, params types.BackupListParams) ([]types.MachineBackup, int, error)
	getProvisionIDBySSHFn func(ctx context.Context, orgID, serverID uuid.UUID) (*uuid.UUID, error)
	insertBackupFn        func(ctx context.Context, backup *types.MachineBackup) error
	updateBackupStatusFn  func(ctx context.Context, id uuid.UUID, status types.MachineBackupStatus, updates map[string]interface{}) error
}

func (m *mockBackupStore) HasInProgressBackup(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID) (bool, error) {
	if m.hasInProgressFn != nil {
		return m.hasInProgressFn(ctx, orgID, serverID)
	}
	return false, nil
}

func (m *mockBackupStore) ListByOrg(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID, params types.BackupListParams) ([]types.MachineBackup, int, error) {
	if m.listByOrgFn != nil {
		return m.listByOrgFn(ctx, orgID, serverID, params)
	}
	return []types.MachineBackup{}, 0, nil
}

func (m *mockBackupStore) GetProvisionIDBySSHKey(ctx context.Context, orgID, serverID uuid.UUID) (*uuid.UUID, error) {
	if m.getProvisionIDBySSHFn != nil {
		return m.getProvisionIDBySSHFn(ctx, orgID, serverID)
	}
	id := uuid.New()
	return &id, nil
}

func (m *mockBackupStore) InsertBackup(ctx context.Context, backup *types.MachineBackup) error {
	if m.insertBackupFn != nil {
		return m.insertBackupFn(ctx, backup)
	}
	return nil
}

func (m *mockBackupStore) UpdateBackupStatus(ctx context.Context, id uuid.UUID, status types.MachineBackupStatus, updates map[string]interface{}) error {
	if m.updateBackupStatusFn != nil {
		return m.updateBackupStatusFn(ctx, id, status, updates)
	}
	return nil
}

// helpers
func defaultSettings() shared_types.OrganizationSettingsData {
	return shared_types.OrganizationSettingsData{}
}

func settingsFn(data shared_types.OrganizationSettingsData, err error) func(context.Context, uuid.UUID) (shared_types.OrganizationSettingsData, error) {
	return func(_ context.Context, _ uuid.UUID) (shared_types.OrganizationSettingsData, error) {
		return data, err
	}
}

func successEnqueue(_ context.Context, _ queue.MachineBackupPayload) (string, error) {
	return "req-123", nil
}

func errorEnqueue(_ context.Context, _ queue.MachineBackupPayload) (string, error) {
	return "", fmt.Errorf("queue error")
}

func newSQLiteBunDB(t *testing.T) *bun.DB {
	t.Helper()
	sqldb, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqldb.Close() })
	return bun.NewDB(sqldb, sqlitedialect.New())
}

// ---------- constructor ----------

func TestNewBackupService(t *testing.T) {
	svc := service.NewBackupService(nil, nil, nil, shared_types.S3Config{})
	assert.NotNil(t, svc)
}

func TestNewBackupServiceWith(t *testing.T) {
	svc := service.NewBackupServiceWith(nil, nil, nil, shared_types.S3Config{})
	assert.NotNil(t, svc)
}

// ---------- UpdateBackupSchedule validation ----------

func TestBackupService_UpdateBackupSchedule_InvalidFrequency(t *testing.T) {
	svc := service.NewBackupService(nil, nil, nil, shared_types.S3Config{})
	_, err := svc.UpdateBackupSchedule(context.Background(), uuid.New(), nil, types.BackupScheduleData{
		Frequency:      "monthly",
		HourUTC:        2,
		DayOfWeek:      0,
		RetentionCount: 7,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid frequency")
}

func TestBackupService_UpdateBackupSchedule_InvalidHour_Negative(t *testing.T) {
	svc := service.NewBackupService(nil, nil, nil, shared_types.S3Config{})
	_, err := svc.UpdateBackupSchedule(context.Background(), uuid.New(), nil, types.BackupScheduleData{
		Frequency:      "daily",
		HourUTC:        -1,
		DayOfWeek:      0,
		RetentionCount: 7,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid hour_utc")
}

func TestBackupService_UpdateBackupSchedule_InvalidHour_TooLarge(t *testing.T) {
	svc := service.NewBackupService(nil, nil, nil, shared_types.S3Config{})
	_, err := svc.UpdateBackupSchedule(context.Background(), uuid.New(), nil, types.BackupScheduleData{
		Frequency:      "weekly",
		HourUTC:        24,
		DayOfWeek:      0,
		RetentionCount: 7,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid hour_utc")
}

func TestBackupService_UpdateBackupSchedule_InvalidDayOfWeek_Negative(t *testing.T) {
	svc := service.NewBackupService(nil, nil, nil, shared_types.S3Config{})
	_, err := svc.UpdateBackupSchedule(context.Background(), uuid.New(), nil, types.BackupScheduleData{
		Frequency:      "daily",
		HourUTC:        2,
		DayOfWeek:      -1,
		RetentionCount: 7,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid day_of_week")
}

func TestBackupService_UpdateBackupSchedule_InvalidDayOfWeek_TooLarge(t *testing.T) {
	svc := service.NewBackupService(nil, nil, nil, shared_types.S3Config{})
	_, err := svc.UpdateBackupSchedule(context.Background(), uuid.New(), nil, types.BackupScheduleData{
		Frequency:      "weekly",
		HourUTC:        12,
		DayOfWeek:      7,
		RetentionCount: 7,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid day_of_week")
}

func TestBackupService_UpdateBackupSchedule_InvalidRetention_Zero(t *testing.T) {
	svc := service.NewBackupService(nil, nil, nil, shared_types.S3Config{})
	_, err := svc.UpdateBackupSchedule(context.Background(), uuid.New(), nil, types.BackupScheduleData{
		Frequency:      "daily",
		HourUTC:        2,
		DayOfWeek:      0,
		RetentionCount: 0,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid retention_count")
}

func TestBackupService_UpdateBackupSchedule_InvalidRetention_TooLarge(t *testing.T) {
	svc := service.NewBackupService(nil, nil, nil, shared_types.S3Config{})
	_, err := svc.UpdateBackupSchedule(context.Background(), uuid.New(), nil, types.BackupScheduleData{
		Frequency:      "daily",
		HourUTC:        2,
		DayOfWeek:      0,
		RetentionCount: 366,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid retention_count")
}

func TestBackupService_UpdateBackupSchedule_GetSettingsError(t *testing.T) {
	svc := service.NewBackupServiceWith(nil, nil, settingsFn(defaultSettings(), fmt.Errorf("db error")), shared_types.S3Config{})
	_, err := svc.UpdateBackupSchedule(context.Background(), uuid.New(), nil, types.BackupScheduleData{
		Frequency:      "daily",
		HourUTC:        2,
		DayOfWeek:      0,
		RetentionCount: 7,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load organization settings")
}

func TestBackupService_UpdateBackupSchedule_Success(t *testing.T) {
	svc := service.NewBackupServiceWith(nil, nil, settingsFn(defaultSettings(), nil), shared_types.S3Config{})
	resp, err := svc.UpdateBackupSchedule(context.Background(), uuid.New(), nil, types.BackupScheduleData{
		Frequency:      "weekly",
		HourUTC:        3,
		DayOfWeek:      1,
		RetentionCount: 14,
	})
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
}

func TestBackupService_UpdateBackupSchedule_DBUpdateError(t *testing.T) {
	db := newSQLiteBunDB(t)
	svc := service.NewBackupServiceWith(nil, nil, settingsFn(defaultSettings(), nil), shared_types.S3Config{})
	svc.SetDBForTest(db)

	_, err := svc.UpdateBackupSchedule(context.Background(), uuid.New(), nil, types.BackupScheduleData{
		Frequency:      "weekly",
		HourUTC:        3,
		DayOfWeek:      1,
		RetentionCount: 14,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update backup schedule")
}

func TestBackupService_UpdateBackupSchedule_DBUpdateSuccess(t *testing.T) {
	db := newSQLiteBunDB(t)
	_, err := db.Exec(`CREATE TABLE organization_settings (
		organization_id TEXT PRIMARY KEY,
		settings TEXT,
		updated_at TEXT
	)`)
	require.NoError(t, err)

	svc := service.NewBackupServiceWith(nil, nil, settingsFn(defaultSettings(), nil), shared_types.S3Config{})
	svc.SetDBForTest(db)

	resp, err := svc.UpdateBackupSchedule(context.Background(), uuid.New(), nil, types.BackupScheduleData{
		Frequency:      "weekly",
		HourUTC:        3,
		DayOfWeek:      1,
		RetentionCount: 14,
	})
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
}

// ---------- GetBackupSchedule ----------

func TestBackupService_GetBackupSchedule_Error(t *testing.T) {
	svc := service.NewBackupServiceWith(nil, nil, settingsFn(defaultSettings(), fmt.Errorf("db error")), shared_types.S3Config{})
	_, err := svc.GetBackupSchedule(context.Background(), uuid.New(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get organization settings")
}

func TestBackupService_GetBackupSchedule_Defaults(t *testing.T) {
	svc := service.NewBackupServiceWith(nil, nil, settingsFn(defaultSettings(), nil), shared_types.S3Config{})
	resp, err := svc.GetBackupSchedule(context.Background(), uuid.New(), nil)
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, "daily", resp.Data.Frequency)
	assert.Equal(t, 2, resp.Data.HourUTC)
}

func TestBackupService_GetBackupSchedule_WithSettings(t *testing.T) {
	freq := "weekly"
	hour := 6
	day := 3
	retention := 30
	enabled := true
	settings := shared_types.OrganizationSettingsData{
		BackupScheduleEnabled:   &enabled,
		BackupScheduleFrequency: &freq,
		BackupScheduleHourUTC:   &hour,
		BackupScheduleDayOfWeek: &day,
		BackupRetentionCount:    &retention,
	}
	svc := service.NewBackupServiceWith(nil, nil, settingsFn(settings, nil), shared_types.S3Config{})
	resp, err := svc.GetBackupSchedule(context.Background(), uuid.New(), nil)
	require.NoError(t, err)
	assert.Equal(t, "weekly", resp.Data.Frequency)
	assert.Equal(t, 6, resp.Data.HourUTC)
	assert.Equal(t, 3, resp.Data.DayOfWeek)
	assert.Equal(t, 30, resp.Data.RetentionCount)
	assert.True(t, resp.Data.Enabled)
}

// ---------- ListBackups ----------

func TestBackupService_ListBackups_InvalidServerID(t *testing.T) {
	svc := service.NewBackupService(nil, nil, nil, shared_types.S3Config{})
	_, err := svc.ListBackups(context.Background(), uuid.New(), types.BackupListParams{
		ServerID: "not-a-uuid",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid server_id")
}

func TestBackupService_ListBackups_StoreError(t *testing.T) {
	bs := &mockBackupStore{
		listByOrgFn: func(_ context.Context, _ uuid.UUID, _ *uuid.UUID, _ types.BackupListParams) ([]types.MachineBackup, int, error) {
			return nil, 0, fmt.Errorf("db error")
		},
	}
	svc := service.NewBackupServiceWith(nil, bs, nil, shared_types.S3Config{})
	_, err := svc.ListBackups(context.Background(), uuid.New(), types.BackupListParams{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list backups")
}

func TestBackupService_ListBackups_Success(t *testing.T) {
	backups := []types.MachineBackup{{MachineName: "container-1"}}
	bs := &mockBackupStore{
		listByOrgFn: func(_ context.Context, _ uuid.UUID, _ *uuid.UUID, _ types.BackupListParams) ([]types.MachineBackup, int, error) {
			return backups, 1, nil
		},
	}
	svc := service.NewBackupServiceWith(nil, bs, nil, shared_types.S3Config{})
	resp, err := svc.ListBackups(context.Background(), uuid.New(), types.BackupListParams{})
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.Len(t, resp.Data.Backups, 1)
	assert.Equal(t, 1, resp.Data.TotalCount)
}

func TestBackupService_ListBackups_NilBackups(t *testing.T) {
	bs := &mockBackupStore{
		listByOrgFn: func(_ context.Context, _ uuid.UUID, _ *uuid.UUID, _ types.BackupListParams) ([]types.MachineBackup, int, error) {
			return nil, 0, nil
		},
	}
	svc := service.NewBackupServiceWith(nil, bs, nil, shared_types.S3Config{})
	resp, err := svc.ListBackups(context.Background(), uuid.New(), types.BackupListParams{})
	require.NoError(t, err)
	assert.NotNil(t, resp.Data.Backups)
	assert.Len(t, resp.Data.Backups, 0)
}

func TestBackupService_ListBackups_WithServerID(t *testing.T) {
	serverID := uuid.New()
	bs := &mockBackupStore{
		listByOrgFn: func(_ context.Context, _ uuid.UUID, sid *uuid.UUID, _ types.BackupListParams) ([]types.MachineBackup, int, error) {
			if sid == nil {
				return nil, 0, fmt.Errorf("expected serverID")
			}
			return []types.MachineBackup{}, 0, nil
		},
	}
	svc := service.NewBackupServiceWith(nil, bs, nil, shared_types.S3Config{})
	resp, err := svc.ListBackups(context.Background(), uuid.New(), types.BackupListParams{ServerID: serverID.String()})
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
}

// ---------- TriggerBackup ----------

func TestBackupService_TriggerBackup_NoProvisionInfo(t *testing.T) {
	provider := &mockBackupProvisionInfo{info: nil, err: nil}
	svc := service.NewBackupService(provider, nil, nil, shared_types.S3Config{})
	_, err := svc.TriggerBackup(context.Background(), uuid.New(), uuid.New(), nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, types.ErrMachineNotProvisioned)
}

func TestBackupService_TriggerBackup_GetProvisionError(t *testing.T) {
	provider := &mockBackupProvisionInfo{info: nil, err: fmt.Errorf("db error")}
	svc := service.NewBackupService(provider, nil, nil, shared_types.S3Config{})
	_, err := svc.TriggerBackup(context.Background(), uuid.New(), uuid.New(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve machine")
}

func TestBackupService_TriggerBackup_EmptyContainerName(t *testing.T) {
	provider := &mockBackupProvisionInfo{
		info: &storage.ProvisionInfo{ContainerName: ""},
		err:  nil,
	}
	svc := service.NewBackupService(provider, nil, nil, shared_types.S3Config{})
	_, err := svc.TriggerBackup(context.Background(), uuid.New(), uuid.New(), nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, types.ErrMachineNotProvisioned)
}

func TestBackupService_TriggerBackup_HasInProgressError(t *testing.T) {
	provider := &mockBackupProvisionInfo{info: &storage.ProvisionInfo{ContainerName: "c1"}}
	bs := &mockBackupStore{
		hasInProgressFn: func(_ context.Context, _ uuid.UUID, _ *uuid.UUID) (bool, error) {
			return false, fmt.Errorf("db error")
		},
	}
	svc := service.NewBackupServiceWith(provider, bs, nil, shared_types.S3Config{})
	_, err := svc.TriggerBackup(context.Background(), uuid.New(), uuid.New(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check backup status")
}

func TestBackupService_TriggerBackup_AlreadyRunning(t *testing.T) {
	provider := &mockBackupProvisionInfo{info: &storage.ProvisionInfo{ContainerName: "c1"}}
	bs := &mockBackupStore{
		hasInProgressFn: func(_ context.Context, _ uuid.UUID, _ *uuid.UUID) (bool, error) {
			return true, nil
		},
	}
	svc := service.NewBackupServiceWith(provider, bs, nil, shared_types.S3Config{})
	_, err := svc.TriggerBackup(context.Background(), uuid.New(), uuid.New(), nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, types.ErrBackupAlreadyRunning)
}

func TestBackupService_TriggerBackup_EnqueueError(t *testing.T) {
	provider := &mockBackupProvisionInfo{info: &storage.ProvisionInfo{ContainerName: "c1"}}
	bs := &mockBackupStore{}
	svc := service.NewBackupServiceWith(provider, bs, nil, shared_types.S3Config{})
	svc.EnqueueBackupFnForTest(errorEnqueue)
	_, err := svc.TriggerBackup(context.Background(), uuid.New(), uuid.New(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to enqueue backup")
}

func TestBackupService_TriggerBackup_Success(t *testing.T) {
	provider := &mockBackupProvisionInfo{info: &storage.ProvisionInfo{ContainerName: "c1", ServerID: "srv-1"}}
	bs := &mockBackupStore{}
	svc := service.NewBackupServiceWith(provider, bs, nil, shared_types.S3Config{})
	svc.EnqueueBackupFnForTest(successEnqueue)
	resp, err := svc.TriggerBackup(context.Background(), uuid.New(), uuid.New(), nil)
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, "req-123", resp.RequestID)
}

func TestBackupService_TriggerBackup_UserOwned_S3NotConfigured(t *testing.T) {
	serverID := uuid.New()
	provider := &mockBackupProvisionInfo{}
	bs := &mockBackupStore{}
	svc := service.NewBackupServiceWith(provider, bs, nil, shared_types.S3Config{})
	svc.CheckUserOwnedFnForTest(func(_ context.Context, _, _ uuid.UUID) (bool, error) { return true, nil })
	_, err := svc.TriggerBackup(context.Background(), uuid.New(), uuid.New(), &serverID)
	require.Error(t, err)
	assert.ErrorIs(t, err, types.ErrS3NotConfigured)
}

func TestBackupService_TriggerBackup_UserOwned_CheckError_FallsThrough(t *testing.T) {
	// When checkUserOwned returns error, fall through to normal provision info path
	serverID := uuid.New()
	provider := &mockBackupProvisionInfo{info: nil, err: nil}
	bs := &mockBackupStore{}
	svc := service.NewBackupServiceWith(provider, bs, nil, shared_types.S3Config{})
	svc.CheckUserOwnedFnForTest(func(_ context.Context, _, _ uuid.UUID) (bool, error) {
		return false, fmt.Errorf("billing error")
	})
	_, err := svc.TriggerBackup(context.Background(), uuid.New(), uuid.New(), &serverID)
	require.Error(t, err)
	assert.ErrorIs(t, err, types.ErrMachineNotProvisioned)
}

func TestBackupService_TriggerBackup_ServerIDNotUserOwned_FallsThroughToProvisionPath(t *testing.T) {
	serverID := uuid.New()
	provider := &mockBackupProvisionInfo{info: &storage.ProvisionInfo{ContainerName: "container-1", ServerID: serverID.String()}}
	bs := &mockBackupStore{}
	svc := service.NewBackupServiceWith(provider, bs, nil, shared_types.S3Config{})
	svc.CheckUserOwnedFnForTest(func(_ context.Context, _, _ uuid.UUID) (bool, error) { return false, nil })
	svc.EnqueueBackupFnForTest(successEnqueue)

	resp, err := svc.TriggerBackup(context.Background(), uuid.New(), uuid.New(), &serverID)
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
}

// ---------- triggerBYOSBackup (via TriggerBackup with user-owned server) ----------

func TestBackupService_TriggerBYOS_HasInProgressError(t *testing.T) {
	serverID := uuid.New()
	bs := &mockBackupStore{
		hasInProgressFn: func(_ context.Context, _ uuid.UUID, _ *uuid.UUID) (bool, error) {
			return false, fmt.Errorf("db error")
		},
	}
	s3Cfg := shared_types.S3Config{Bucket: "b", Endpoint: "e", AccessKey: "k", SecretKey: "s"}
	svc := service.NewBackupServiceWith(nil, bs, nil, s3Cfg)
	svc.CheckUserOwnedFnForTest(func(_ context.Context, _, _ uuid.UUID) (bool, error) { return true, nil })
	_, err := svc.TriggerBackup(context.Background(), uuid.New(), uuid.New(), &serverID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check backup status")
}

func TestBackupService_TriggerBYOS_AlreadyRunning(t *testing.T) {
	serverID := uuid.New()
	bs := &mockBackupStore{
		hasInProgressFn: func(_ context.Context, _ uuid.UUID, _ *uuid.UUID) (bool, error) {
			return true, nil
		},
	}
	s3Cfg := shared_types.S3Config{Bucket: "b", Endpoint: "e", AccessKey: "k", SecretKey: "s"}
	svc := service.NewBackupServiceWith(nil, bs, nil, s3Cfg)
	svc.CheckUserOwnedFnForTest(func(_ context.Context, _, _ uuid.UUID) (bool, error) { return true, nil })
	_, err := svc.TriggerBackup(context.Background(), uuid.New(), uuid.New(), &serverID)
	require.Error(t, err)
	assert.ErrorIs(t, err, types.ErrBackupAlreadyRunning)
}

func TestBackupService_TriggerBYOS_GetProvisionIDError(t *testing.T) {
	serverID := uuid.New()
	bs := &mockBackupStore{
		getProvisionIDBySSHFn: func(_ context.Context, _, _ uuid.UUID) (*uuid.UUID, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	s3Cfg := shared_types.S3Config{Bucket: "b", Endpoint: "e", AccessKey: "k", SecretKey: "s"}
	svc := service.NewBackupServiceWith(nil, bs, nil, s3Cfg)
	svc.CheckUserOwnedFnForTest(func(_ context.Context, _, _ uuid.UUID) (bool, error) { return true, nil })
	_, err := svc.TriggerBackup(context.Background(), uuid.New(), uuid.New(), &serverID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve provision")
}

func TestBackupService_TriggerBYOS_GetSettingsError(t *testing.T) {
	serverID := uuid.New()
	bs := &mockBackupStore{}
	s3Cfg := shared_types.S3Config{Bucket: "b", Endpoint: "e", AccessKey: "k", SecretKey: "s"}
	svc := service.NewBackupServiceWith(nil, bs, settingsFn(defaultSettings(), fmt.Errorf("db error")), s3Cfg)
	svc.CheckUserOwnedFnForTest(func(_ context.Context, _, _ uuid.UUID) (bool, error) { return true, nil })
	_, err := svc.TriggerBackup(context.Background(), uuid.New(), uuid.New(), &serverID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get org settings")
}

func TestBackupService_TriggerBYOS_InsertBackupError(t *testing.T) {
	serverID := uuid.New()
	bs := &mockBackupStore{
		insertBackupFn: func(_ context.Context, _ *types.MachineBackup) error {
			return fmt.Errorf("insert failed")
		},
	}
	s3Cfg := shared_types.S3Config{Bucket: "b", Endpoint: "e", AccessKey: "k", SecretKey: "s"}
	svc := service.NewBackupServiceWith(nil, bs, settingsFn(defaultSettings(), nil), s3Cfg)
	svc.CheckUserOwnedFnForTest(func(_ context.Context, _, _ uuid.UUID) (bool, error) { return true, nil })
	_, err := svc.TriggerBackup(context.Background(), uuid.New(), uuid.New(), &serverID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create backup record")
}

func TestBackupService_TriggerBYOS_EnqueueError(t *testing.T) {
	serverID := uuid.New()
	bs := &mockBackupStore{}
	s3Cfg := shared_types.S3Config{Bucket: "b", Endpoint: "e", AccessKey: "k", SecretKey: "s"}
	svc := service.NewBackupServiceWith(nil, bs, settingsFn(defaultSettings(), nil), s3Cfg)
	svc.CheckUserOwnedFnForTest(func(_ context.Context, _, _ uuid.UUID) (bool, error) { return true, nil })
	svc.EnqueueBackupFnForTest(errorEnqueue)
	_, err := svc.TriggerBackup(context.Background(), uuid.New(), uuid.New(), &serverID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to enqueue backup")
}

func TestBackupService_TriggerBYOS_Success(t *testing.T) {
	serverID := uuid.New()
	bs := &mockBackupStore{}
	retention := 14
	customPaths := []string{"/data", "/var/backups"}
	settings := shared_types.OrganizationSettingsData{
		BackupRetentionCount: &retention,
		BackupPaths:          customPaths,
	}
	s3Cfg := shared_types.S3Config{Bucket: "b", Endpoint: "e", AccessKey: "k", SecretKey: "s"}
	svc := service.NewBackupServiceWith(nil, bs, settingsFn(settings, nil), s3Cfg)
	svc.CheckUserOwnedFnForTest(func(_ context.Context, _, _ uuid.UUID) (bool, error) { return true, nil })
	svc.EnqueueBackupFnForTest(successEnqueue)
	resp, err := svc.TriggerBackup(context.Background(), uuid.New(), uuid.New(), &serverID)
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, "req-123", resp.RequestID)
}
