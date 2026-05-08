package scheduler

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/uptrace/bun"
)

type ScheduleStatus string

const (
	StatusActive    ScheduleStatus = "active"
	StatusPaused    ScheduleStatus = "paused"
	StatusCompleted ScheduleStatus = "completed"
	StatusFailed    ScheduleStatus = "failed"
)

type RunStatus string

const (
	RunRunning   RunStatus = "running"
	RunSucceeded RunStatus = "succeeded"
	RunFailed    RunStatus = "failed"
	RunTimedOut  RunStatus = "timed_out"
)

type JSONMetadata map[string]string

var _ driver.Valuer = JSONMetadata{}
var _ sql.Scanner = &JSONMetadata{}

func (m JSONMetadata) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}
	b, err := json.Marshal(m)
	return string(b), err
}

func (m *JSONMetadata) Scan(src interface{}) error {
	if src == nil {
		*m = make(map[string]string)
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("scheduler: unsupported type for JSONMetadata: %T", src)
	}
	return json.Unmarshal(data, m)
}

type Schedule struct {
	bun.BaseModel `bun:"table:agent_schedules"`

	ID                    uuid.UUID      `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	UserID                string         `bun:"user_id,notnull" json:"user_id"`
	OrgID                 string         `bun:"org_id,notnull" json:"org_id"`
	ThreadID              string         `bun:"thread_id,notnull" json:"thread_id"`
	Name                  string         `bun:"name,notnull" json:"name"`
	Description           string         `bun:"description" json:"description"`
	Prompt                string         `bun:"prompt,notnull" json:"prompt"`
	CronExpression        string         `bun:"cron_expression,notnull" json:"cron_expression"`
	Timezone              string         `bun:"timezone,notnull,default:'UTC'" json:"timezone"`
	Status                ScheduleStatus `bun:"status,notnull,default:'active'" json:"status"`
	DeliveryChannel       string         `bun:"delivery_channel" json:"delivery_channel"`
	NotifyOn              string         `bun:"notify_on,notnull,default:'smart'" json:"notify_on"`
	LastRunAt             *time.Time     `bun:"last_run_at" json:"last_run_at"`
	NextRunAt             *time.Time     `bun:"next_run_at" json:"next_run_at"`
	RunCount              int            `bun:"run_count,notnull,default:0" json:"run_count"`
	MaxRuns               *int           `bun:"max_runs" json:"max_runs"`
	ErrorCount            int            `bun:"error_count,notnull,default:0" json:"error_count"`
	MaxConsecutiveErrors  int            `bun:"max_consecutive_errors,notnull,default:3" json:"max_consecutive_errors"`
	ConsecutiveErrorCount int            `bun:"consecutive_error_count,notnull,default:0" json:"consecutive_error_count"`
	TimeoutSeconds        int            `bun:"timeout_seconds,notnull,default:300" json:"timeout_seconds"`
	Metadata              JSONMetadata   `bun:"metadata,type:jsonb" json:"metadata"`
	CreatedAt             time.Time      `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt             time.Time      `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt             *time.Time     `bun:"deleted_at,soft_delete" json:"deleted_at,omitempty"`
}

type ScheduleRun struct {
	bun.BaseModel `bun:"table:agent_schedule_runs"`

	ID          uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	ScheduleID  uuid.UUID  `bun:"schedule_id,notnull,type:uuid" json:"schedule_id"`
	ThreadID    string     `bun:"thread_id,notnull" json:"thread_id"`
	Status      RunStatus  `bun:"status,notnull,default:'running'" json:"status"`
	Result      string     `bun:"result" json:"result"`
	Error       string     `bun:"error" json:"error"`
	TokensUsed  int        `bun:"tokens_used,notnull,default:0" json:"tokens_used"`
	StartedAt   time.Time  `bun:"started_at,notnull,default:current_timestamp" json:"started_at"`
	CompletedAt *time.Time `bun:"completed_at" json:"completed_at"`
}

type Store struct {
	DB     *bun.DB
	Logger logger.Logger
}

func NewStore(db *bun.DB, l logger.Logger) *Store {
	return &Store{DB: db, Logger: l}
}

func (s *Store) CreateTables(ctx context.Context) {
	if _, err := s.DB.NewCreateTable().
		Model((*Schedule)(nil)).
		IfNotExists().
		Exec(ctx); err != nil {
		s.Logger.Log(logger.Error, "scheduler: failed to create agent_schedules table", err.Error())
	}
	if _, err := s.DB.NewCreateTable().
		Model((*ScheduleRun)(nil)).
		IfNotExists().
		Exec(ctx); err != nil {
		s.Logger.Log(logger.Error, "scheduler: failed to create agent_schedule_runs table", err.Error())
	}
}

func (s *Store) CreateSchedule(ctx context.Context, schedule *Schedule) error {
	if schedule.ID == uuid.Nil {
		schedule.ID = uuid.New()
	}
	now := time.Now()
	schedule.CreatedAt = now
	schedule.UpdatedAt = now

	_, err := s.DB.NewInsert().Model(schedule).Exec(ctx)
	if err != nil {
		return fmt.Errorf("scheduler: insert schedule: %w", err)
	}
	return nil
}

func (s *Store) GetSchedule(ctx context.Context, id uuid.UUID) (*Schedule, error) {
	schedule := &Schedule{}
	err := s.DB.NewSelect().
		Model(schedule).
		Where("id = ? AND deleted_at IS NULL", id).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("scheduler: get schedule %s: %w", id, err)
	}
	return schedule, nil
}

func (s *Store) GetScheduleForUser(ctx context.Context, id uuid.UUID, userID string) (*Schedule, error) {
	schedule := &Schedule{}
	err := s.DB.NewSelect().
		Model(schedule).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("scheduler: get schedule %s for user %s: %w", id, userID, err)
	}
	return schedule, nil
}

func (s *Store) ListSchedulesForUser(ctx context.Context, userID, orgID string) ([]Schedule, error) {
	var schedules []Schedule
	err := s.DB.NewSelect().
		Model(&schedules).
		Where("user_id = ? AND org_id = ? AND deleted_at IS NULL", userID, orgID).
		OrderExpr("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("scheduler: list schedules: %w", err)
	}
	return schedules, nil
}

func (s *Store) GetDueSchedules(ctx context.Context, now time.Time) ([]Schedule, error) {
	var schedules []Schedule
	err := s.DB.NewSelect().
		Model(&schedules).
		Where("status = ?", StatusActive).
		Where("next_run_at <= ?", now).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("scheduler: get due schedules: %w", err)
	}
	return schedules, nil
}

func (s *Store) UpdateScheduleStatus(ctx context.Context, id uuid.UUID, status ScheduleStatus) error {
	_, err := s.DB.NewUpdate().
		Model((*Schedule)(nil)).
		Set("status = ?", status).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (s *Store) UpdateScheduleAfterRun(ctx context.Context, id uuid.UUID, nextRunAt *time.Time, success bool) error {
	now := time.Now()
	q := s.DB.NewUpdate().
		Model((*Schedule)(nil)).
		Set("last_run_at = ?", now).
		Set("next_run_at = ?", nextRunAt).
		Set("run_count = run_count + 1").
		Set("updated_at = ?", now).
		Where("id = ?", id)

	if success {
		q = q.Set("consecutive_error_count = 0")
	} else {
		q = q.Set("error_count = error_count + 1").
			Set("consecutive_error_count = consecutive_error_count + 1")
	}

	_, err := q.Exec(ctx)
	return err
}

func (s *Store) SoftDeleteSchedule(ctx context.Context, id uuid.UUID, userID string) error {
	now := time.Now()
	_, err := s.DB.NewUpdate().
		Model((*Schedule)(nil)).
		Set("deleted_at = ?", now).
		Set("updated_at = ?", now).
		Set("status = ?", StatusCompleted).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).
		Exec(ctx)
	return err
}

func (s *Store) CreateRun(ctx context.Context, run *ScheduleRun) error {
	if run.ID == uuid.Nil {
		run.ID = uuid.New()
	}
	run.StartedAt = time.Now()

	_, err := s.DB.NewInsert().Model(run).Exec(ctx)
	return err
}

func (s *Store) CompleteRun(ctx context.Context, id uuid.UUID, status RunStatus, result, errMsg string, tokensUsed int) error {
	now := time.Now()
	_, err := s.DB.NewUpdate().
		Model((*ScheduleRun)(nil)).
		Set("status = ?", status).
		Set("result = ?", result).
		Set("error = ?", errMsg).
		Set("tokens_used = ?", tokensUsed).
		Set("completed_at = ?", now).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (s *Store) GetRunsForSchedule(ctx context.Context, scheduleID uuid.UUID, limit int) ([]ScheduleRun, error) {
	var runs []ScheduleRun
	q := s.DB.NewSelect().
		Model(&runs).
		Where("schedule_id = ?", scheduleID).
		OrderExpr("started_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Scan(ctx)
	return runs, err
}
