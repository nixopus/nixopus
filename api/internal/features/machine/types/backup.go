package types

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type MachineBackupStatus string

const (
	BackupStatusPending    MachineBackupStatus = "pending"
	BackupStatusInProgress MachineBackupStatus = "in_progress"
	BackupStatusCompleted  MachineBackupStatus = "completed"
	BackupStatusFailed     MachineBackupStatus = "failed"
)

type MachineBackup struct {
	bun.BaseModel `bun:"table:machine_backups,alias:mb" swaggerignore:"true"`

	ID             uuid.UUID           `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	UserID         uuid.UUID           `json:"user_id" bun:"user_id,type:uuid,notnull"`
	OrganizationID uuid.UUID           `json:"organization_id" bun:"organization_id,type:uuid,notnull"`
	ProvisionID    *uuid.UUID          `json:"provision_id" bun:"provision_id,type:uuid"`
	MachineName    string              `json:"machine_name" bun:"machine_name,type:varchar(255),notnull"`
	Status         MachineBackupStatus `json:"status" bun:"status,type:machine_backup_status,notnull,default:'pending'"`
	Trigger        string              `json:"trigger" bun:"trigger,type:varchar(50),notnull"`
	SnapshotPath   *string             `json:"snapshot_path" bun:"snapshot_path,type:text"`
	S3Path         *string             `json:"s3_path" bun:"s3_path,type:text"`
	SizeBytes      int64               `json:"size_bytes" bun:"size_bytes,type:bigint,default:0"`
	Error          *string             `json:"error" bun:"error,type:text"`
	StartedAt      *time.Time          `json:"started_at" bun:"started_at,type:timestamptz"`
	CompletedAt    *time.Time          `json:"completed_at" bun:"completed_at,type:timestamptz"`
	CreatedAt      time.Time           `json:"created_at" bun:"created_at,type:timestamptz,notnull,default:now()"`
	UpdatedAt      time.Time           `json:"updated_at" bun:"updated_at,type:timestamptz,notnull,default:now()"`
}

type TriggerBackupResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

type BackupListResponse struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Data    []MachineBackup `json:"data"`
}

type BackupScheduleData struct {
	Enabled        bool   `json:"enabled"`
	Frequency      string `json:"frequency"`
	HourUTC        int    `json:"hour_utc"`
	DayOfWeek      int    `json:"day_of_week"`
	RetentionCount int    `json:"retention_count"`
}

type BackupScheduleResponse struct {
	Status  string             `json:"status"`
	Message string             `json:"message"`
	Data    BackupScheduleData `json:"data"`
}

var (
	ErrBackupAlreadyRunning = errors.New("a backup is already in progress for this machine")
)
