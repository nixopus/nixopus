package queue

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vmihailenco/taskq/v3"
)

const (
	queueMachineBackup = "machine-backup"
	taskMachineBackup  = "task_machine_backup"
	backupReplyPrefix  = "backup:reply:"
)

type MachineBackupPayload struct {
	RequestID   string `json:"request_id"`
	MachineName string `json:"machine_name"`
	UserID      string `json:"user_id"`
	OrgID       string `json:"org_id"`
	ServerID    string `json:"server_id,omitempty"`
	Trigger     string `json:"trigger"`
	ExpiresAt   int64  `json:"expires_at"`
}

type MachineBackupResult struct {
	RequestID    string `json:"request_id"`
	Success      bool   `json:"success"`
	BackupID     string `json:"backup_id,omitempty"`
	SnapshotPath string `json:"snapshot_path,omitempty"`
	S3Path       string `json:"s3_path,omitempty"`
	SizeBytes    int64  `json:"size_bytes,omitempty"`
	Error        string `json:"error,omitempty"`
}

var (
	onceMachineBackup     sync.Once
	machineBackupQueue    taskq.Queue
	taskMachineBackupTask *taskq.Task
	backupReplyMux        *ReplyMultiplexer
)

func SetupMachineBackupQueue(ctx context.Context) {
	onceMachineBackup.Do(func() {
		machineBackupQueue = registerProducerQueue(&taskq.QueueOptions{
			Name: queueMachineBackup,
		})

		taskMachineBackupTask = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       taskMachineBackup,
			RetryLimit: 0,
			Handler: func(ctx context.Context, payload MachineBackupPayload) error {
				return nil
			},
		})

		backupReplyMux = NewReplyMultiplexerWithPrefix(backupReplyPrefix)
		backupReplyMux.Start(ctx)

		log.Printf("Machine backup queue and reply multiplexer initialized")
	})
}

// EnqueueMachineBackup enqueues a backup task and returns the request ID immediately.
// The caller should poll the machine_backups table for completion status.
func EnqueueMachineBackup(ctx context.Context, payload MachineBackupPayload) (string, error) {
	if taskMachineBackupTask == nil {
		return "", fmt.Errorf("machine backup queue not initialized - call SetupMachineBackupQueue first")
	}

	requestID := uuid.New().String()
	payload.RequestID = requestID
	payload.ExpiresAt = time.Now().Add(30 * time.Minute).Unix()

	q := machineBackupQueue
	if payload.ServerID != "" {
		q = getOrCreateProducerQueue(queueMachineBackup + "-" + payload.ServerID)
	}

	if err := q.Add(taskMachineBackupTask.WithArgs(ctx, payload)); err != nil {
		return "", fmt.Errorf("failed to enqueue backup task: %w", err)
	}

	return requestID, nil
}
