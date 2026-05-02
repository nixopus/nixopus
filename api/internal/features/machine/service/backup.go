package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	deploy_s3 "github.com/nixopus/nixopus/api/internal/features/deploy/s3"
	"github.com/nixopus/nixopus/api/internal/features/machine/storage"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/nixopus/nixopus/api/internal/queue"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/nixopus/nixopus/api/internal/utils"
	"github.com/uptrace/bun"
)

var defaultBackupPaths = []string{"/home", "/etc", "/var/lib/docker/volumes"}

// BackupStoreInterface abstracts BackupStorage for testability.
type BackupStoreInterface interface {
	HasInProgressBackup(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID) (bool, error)
	ListByOrg(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID, params types.BackupListParams) ([]types.MachineBackup, int, error)
	GetProvisionIDBySSHKey(ctx context.Context, orgID, serverID uuid.UUID) (*uuid.UUID, error)
	InsertBackup(ctx context.Context, backup *types.MachineBackup) error
	UpdateBackupStatus(ctx context.Context, id uuid.UUID, status types.MachineBackupStatus, updates map[string]interface{}) error
}

type BackupService struct {
	provisionInfo    ProvisionInfoProvider
	backupStore      BackupStoreInterface
	db               *bun.DB
	s3Cfg            shared_types.S3Config
	getSettingsFn    func(ctx context.Context, orgID uuid.UUID) (shared_types.OrganizationSettingsData, error) // nil → production
	checkUserOwnedFn func(ctx context.Context, orgID, serverID uuid.UUID) (bool, error)                        // nil → production billing storage
	enqueueBackupFn  func(ctx context.Context, payload queue.MachineBackupPayload) (string, error)             // nil → real queue
}

func NewBackupService(p ProvisionInfoProvider, bs *storage.BackupStorage, db *bun.DB, s3Cfg shared_types.S3Config) *BackupService {
	return &BackupService{provisionInfo: p, backupStore: bs, db: db, s3Cfg: s3Cfg}
}

// NewBackupServiceWith creates a BackupService with injectable dependencies for tests.
func NewBackupServiceWith(p ProvisionInfoProvider, bs BackupStoreInterface, getSettingsFn func(ctx context.Context, orgID uuid.UUID) (shared_types.OrganizationSettingsData, error), s3Cfg shared_types.S3Config) *BackupService {
	return &BackupService{provisionInfo: p, backupStore: bs, getSettingsFn: getSettingsFn, s3Cfg: s3Cfg}
}

// EnqueueBackupFnForTest injects a mock enqueue function for tests.
func (s *BackupService) EnqueueBackupFnForTest(fn func(ctx context.Context, payload queue.MachineBackupPayload) (string, error)) {
	s.enqueueBackupFn = fn
}

// CheckUserOwnedFnForTest injects a mock user-ownership check for tests.
func (s *BackupService) CheckUserOwnedFnForTest(fn func(ctx context.Context, orgID, serverID uuid.UUID) (bool, error)) {
	s.checkUserOwnedFn = fn
}

// SetDBForTest injects a DB handle for tests.
func (s *BackupService) SetDBForTest(db *bun.DB) {
	s.db = db
}

func (s *BackupService) getOrganizationSettings(ctx context.Context, orgID uuid.UUID) (shared_types.OrganizationSettingsData, error) {
	if s.getSettingsFn != nil {
		return s.getSettingsFn(ctx, orgID)
	}
	return utils.GetOrganizationSettings(ctx, s.db, orgID)
}

func (s *BackupService) TriggerBackup(ctx context.Context, userID, orgID uuid.UUID, serverID *uuid.UUID) (*types.TriggerBackupResponse, error) {
	if serverID != nil {
		var isUserOwned bool
		var checkErr error
		if s.checkUserOwnedFn != nil {
			isUserOwned, checkErr = s.checkUserOwnedFn(ctx, orgID, *serverID)
		} else {
			billingStore := storage.NewBillingStorage(s.db, ctx)
			isUserOwned, checkErr = billingStore.IsServerUserOwned(orgID, *serverID)
		}
		if checkErr == nil && isUserOwned {
			return s.triggerBYOSBackup(ctx, userID, orgID, *serverID)
		}
	}

	info, err := s.provisionInfo.GetProvisionInfo(ctx, orgID, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve machine: %w", err)
	}
	if info == nil || info.ContainerName == "" {
		return nil, types.ErrMachineNotProvisioned
	}

	hasRunning, err := s.backupStore.HasInProgressBackup(ctx, orgID, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to check backup status: %w", err)
	}
	if hasRunning {
		return nil, types.ErrBackupAlreadyRunning
	}

	payload := queue.MachineBackupPayload{
		MachineName: info.ContainerName,
		UserID:      userID.String(),
		OrgID:       orgID.String(),
		ServerID:    info.ServerID,
		Trigger:     "api",
	}

	enqueue := queue.EnqueueMachineBackup
	if s.enqueueBackupFn != nil {
		enqueue = s.enqueueBackupFn
	}
	requestID, err := enqueue(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue backup: %w", err)
	}

	return &types.TriggerBackupResponse{
		Status:    "success",
		Message:   "Backup initiated. Check status via GET /machine/backups.",
		RequestID: requestID,
	}, nil
}

func (s *BackupService) triggerBYOSBackup(ctx context.Context, userID, orgID, serverID uuid.UUID) (*types.TriggerBackupResponse, error) {
	if !deploy_s3.IsConfigured(s.s3Cfg) {
		return nil, types.ErrS3NotConfigured
	}

	hasRunning, err := s.backupStore.HasInProgressBackup(ctx, orgID, &serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to check backup status: %w", err)
	}
	if hasRunning {
		return nil, types.ErrBackupAlreadyRunning
	}

	provisionID, err := s.backupStore.GetProvisionIDBySSHKey(ctx, orgID, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve provision: %w", err)
	}

	settings, err := s.getOrganizationSettings(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get org settings: %w", err)
	}
	paths := defaultBackupPaths
	if len(settings.BackupPaths) > 0 {
		paths = settings.BackupPaths
	}
	retention := 7
	if settings.BackupRetentionCount != nil {
		retention = *settings.BackupRetentionCount
	}

	backup := &types.MachineBackup{
		UserID:         userID,
		OrganizationID: orgID,
		ProvisionID:    provisionID,
		MachineName:    serverID.String(),
		Status:         types.BackupStatusPending,
		Trigger:        "api",
	}
	if err := s.backupStore.InsertBackup(ctx, backup); err != nil {
		return nil, fmt.Errorf("failed to create backup record: %w", err)
	}

	enqueue := queue.EnqueueMachineBackup
	if s.enqueueBackupFn != nil {
		enqueue = s.enqueueBackupFn
	}

	bkPayload := queue.MachineBackupPayload{
		MachineName:    serverID.String(),
		UserID:         userID.String(),
		OrgID:          orgID.String(),
		ServerID:       serverID.String(),
		BackupRowID:    backup.ID.String(),
		BackupPaths:    paths,
		RetentionCount: retention,
		Trigger:        "api",
	}

	requestID, err := enqueue(ctx, bkPayload)
	if err != nil {
		_ = s.backupStore.UpdateBackupStatus(ctx, backup.ID, types.BackupStatusFailed, map[string]interface{}{
			"error": "failed to enqueue backup task",
		})
		return nil, fmt.Errorf("failed to enqueue backup: %w", err)
	}

	return &types.TriggerBackupResponse{
		Status:    "success",
		Message:   "Backup initiated. Check status via GET /machine/backups.",
		RequestID: requestID,
	}, nil
}

func (s *BackupService) ListBackups(ctx context.Context, orgID uuid.UUID, params types.BackupListParams) (*types.BackupListResponse, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.SortBy == "" {
		params.SortBy = "created_at"
	}
	if params.SortOrder == "" {
		params.SortOrder = "desc"
	}

	var serverUUID *uuid.UUID
	if params.ServerID != "" {
		parsed, err := uuid.Parse(params.ServerID)
		if err != nil {
			return nil, fmt.Errorf("invalid server_id: %w", err)
		}
		serverUUID = &parsed
	}

	backups, totalCount, err := s.backupStore.ListByOrg(ctx, orgID, serverUUID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}
	if backups == nil {
		backups = []types.MachineBackup{}
	}

	return &types.BackupListResponse{
		Status:  "success",
		Message: "Backups retrieved",
		Data: types.BackupListResponseData{
			Backups:    backups,
			TotalCount: totalCount,
			Page:       params.Page,
			PageSize:   params.PageSize,
		},
	}, nil
}

func (s *BackupService) GetBackupSchedule(ctx context.Context, orgID uuid.UUID, _ *uuid.UUID) (*types.BackupScheduleResponse, error) {
	settings, err := s.getOrganizationSettings(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization settings: %w", err)
	}

	data := types.BackupScheduleData{
		Frequency:      "daily",
		HourUTC:        2,
		DayOfWeek:      0,
		RetentionCount: 7,
	}
	if settings.BackupScheduleEnabled != nil {
		data.Enabled = *settings.BackupScheduleEnabled
	}
	if settings.BackupScheduleFrequency != nil {
		data.Frequency = *settings.BackupScheduleFrequency
	}
	if settings.BackupScheduleHourUTC != nil {
		data.HourUTC = *settings.BackupScheduleHourUTC
	}
	if settings.BackupScheduleDayOfWeek != nil {
		data.DayOfWeek = *settings.BackupScheduleDayOfWeek
	}
	if settings.BackupRetentionCount != nil {
		data.RetentionCount = *settings.BackupRetentionCount
	}

	return &types.BackupScheduleResponse{
		Status:  "success",
		Message: "Backup schedule retrieved",
		Data:    data,
	}, nil
}

func (s *BackupService) UpdateBackupSchedule(ctx context.Context, orgID uuid.UUID, _ *uuid.UUID, req types.BackupScheduleData) (*types.BackupScheduleResponse, error) {
	if req.Frequency != "daily" && req.Frequency != "weekly" {
		return nil, fmt.Errorf("invalid frequency: must be 'daily' or 'weekly'")
	}
	if req.HourUTC < 0 || req.HourUTC > 23 {
		return nil, fmt.Errorf("invalid hour_utc: must be 0-23")
	}
	if req.DayOfWeek < 0 || req.DayOfWeek > 6 {
		return nil, fmt.Errorf("invalid day_of_week: must be 0 (Sun) - 6 (Sat)")
	}
	if req.RetentionCount < 1 || req.RetentionCount > 365 {
		return nil, fmt.Errorf("invalid retention_count: must be 1-365")
	}

	// GetOrganizationSettings upserts default settings if the row is missing,
	// so we always have a valid row to update.
	currentSettings, err := s.getOrganizationSettings(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to load organization settings: %w", err)
	}

	currentSettings.BackupScheduleEnabled = &req.Enabled
	currentSettings.BackupScheduleFrequency = &req.Frequency
	currentSettings.BackupScheduleHourUTC = &req.HourUTC
	currentSettings.BackupScheduleDayOfWeek = &req.DayOfWeek
	currentSettings.BackupRetentionCount = &req.RetentionCount

	if s.db != nil {
		_, err = s.db.NewUpdate().
			TableExpr("organization_settings").
			Set("settings = ?", currentSettings).
			Set("updated_at = CURRENT_TIMESTAMP").
			Where("organization_id = ?", orgID).
			Exec(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to update backup schedule: %w", err)
		}
	}

	return &types.BackupScheduleResponse{
		Status:  "success",
		Message: "Backup schedule updated",
		Data:    req,
	}, nil
}
