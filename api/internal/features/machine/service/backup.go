package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/machine/storage"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/nixopus/nixopus/api/internal/queue"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/nixopus/nixopus/api/internal/utils"
	"github.com/uptrace/bun"
)

type BackupService struct {
	provisionInfo ProvisionInfoProvider
	backupStore   *storage.BackupStorage
	db            *bun.DB
}

func NewBackupService(p ProvisionInfoProvider, bs *storage.BackupStorage, db *bun.DB) *BackupService {
	return &BackupService{provisionInfo: p, backupStore: bs, db: db}
}

func (s *BackupService) TriggerBackup(ctx context.Context, userID, orgID uuid.UUID, serverID *uuid.UUID) (*types.TriggerBackupResponse, error) {
	info, err := s.provisionInfo.GetProvisionInfo(ctx, orgID, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve machine: %w", err)
	}
	if info == nil || info.ContainerName == "" {
		return nil, types.ErrMachineNotProvisioned
	}

	hasRunning, err := s.backupStore.HasInProgressBackup(ctx, orgID)
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

	requestID, err := queue.EnqueueMachineBackup(ctx, payload)
	if err != nil {
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

	backups, totalCount, err := s.backupStore.ListByOrg(ctx, orgID, params)
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

func (s *BackupService) GetBackupSchedule(ctx context.Context, orgID uuid.UUID) (*types.BackupScheduleResponse, error) {
	settings, err := utils.GetOrganizationSettings(ctx, s.db, orgID)
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

func (s *BackupService) UpdateBackupSchedule(ctx context.Context, orgID uuid.UUID, req types.BackupScheduleData) (*types.BackupScheduleResponse, error) {
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

	var orgSettings shared_types.OrganizationSettings
	err := s.db.NewSelect().
		Model(&orgSettings).
		Where("organization_id = ?", orgID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load organization settings: %w", err)
	}

	orgSettings.Settings.BackupScheduleEnabled = &req.Enabled
	orgSettings.Settings.BackupScheduleFrequency = &req.Frequency
	orgSettings.Settings.BackupScheduleHourUTC = &req.HourUTC
	orgSettings.Settings.BackupScheduleDayOfWeek = &req.DayOfWeek
	orgSettings.Settings.BackupRetentionCount = &req.RetentionCount
	orgSettings.UpdatedAt = time.Now()

	_, err = s.db.NewUpdate().
		Model(&orgSettings).
		Column("settings", "updated_at").
		Where("id = ?", orgSettings.ID).
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update backup schedule: %w", err)
	}

	return &types.BackupScheduleResponse{
		Status:  "success",
		Message: "Backup schedule updated",
		Data:    req,
	}, nil
}
