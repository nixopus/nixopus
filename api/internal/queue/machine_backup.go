package queue

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	apilog "github.com/nixopus/nixopus/api/internal/log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/config"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	machine_storage "github.com/nixopus/nixopus/api/internal/features/machine/storage"
	machine_types "github.com/nixopus/nixopus/api/internal/features/machine/types"
	sshpkg "github.com/nixopus/nixopus/api/internal/features/ssh"
	"github.com/uptrace/bun"
	"github.com/vmihailenco/taskq/v3"
)

const (
	queueMachineBackup = "machine-backup"
	taskMachineBackup  = "task_machine_backup"
	backupReplyPrefix  = "backup:reply:"
)

type MachineBackupPayload struct {
	RequestID      string   `json:"request_id"`
	MachineName    string   `json:"machine_name"`
	UserID         string   `json:"user_id"`
	OrgID          string   `json:"org_id"`
	ServerID       string   `json:"server_id,omitempty"`
	BackupRowID    string   `json:"backup_row_id"`
	BackupPaths    []string `json:"backup_paths,omitempty"`
	RetentionCount int      `json:"retention_count,omitempty"`
	Trigger        string   `json:"trigger"`
	ExpiresAt      int64    `json:"expires_at"`
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
	backupWorker          *BackupWorker
)

func machineBackupTaskHandler(ctx context.Context, payload MachineBackupPayload) error {
	return backupWorker.Handle(ctx, payload)
}

// backupSSHRunner is satisfied by *sshpkg.SSHManager; swappable in tests.
type backupSSHRunner interface {
	RunCommand(cmd string) (string, error)
}

// getSSHManagerForServerForBackup is the hook used by BackupWorker; tests may replace it.
var getSSHManagerForServerForBackup = func(ctx context.Context, orgID, serverID uuid.UUID) (backupSSHRunner, error) {
	return sshpkg.GetSSHManagerForServer(ctx, orgID, serverID)
}

type BackupWorker struct {
	db          *bun.DB
	backupStore *machine_storage.BackupStorage
}

func newBackupWorker(db *bun.DB, ctx context.Context) *BackupWorker {
	return &BackupWorker{
		db:          db,
		backupStore: machine_storage.NewBackupStorage(db, ctx),
	}
}

var defaultBackupPaths = []string{"/home", "/etc", "/var/lib/docker/volumes"}

func (w *BackupWorker) Handle(ctx context.Context, payload MachineBackupPayload) error {
	log := logger.NewLogger()

	if time.Now().Unix() > payload.ExpiresAt {
		log.Log(logger.Warning, "backup task expired, dropping", payload.OrgID)
		return nil
	}

	rowID, err := uuid.Parse(payload.BackupRowID)
	if err != nil {
		log.Log(logger.Error, fmt.Sprintf("invalid backup row id: %v", err), payload.OrgID)
		return nil
	}
	orgID, err := uuid.Parse(payload.OrgID)
	if err != nil {
		return nil
	}
	serverID, err := uuid.Parse(payload.ServerID)
	if err != nil {
		log.Log(logger.Error, fmt.Sprintf("invalid server id: %v", err), payload.OrgID)
		return w.failBackup(ctx, rowID, "invalid server id in payload")
	}

	now := time.Now()
	if err := w.backupStore.UpdateBackupStatus(ctx, rowID, machine_types.BackupStatusInProgress, map[string]interface{}{
		"started_at": now,
	}); err != nil {
		log.Log(logger.Error, fmt.Sprintf("failed to mark backup in_progress: %v", err), payload.OrgID)
		return nil
	}

	sshMgr, err := getSSHManagerForServerForBackup(ctx, orgID, serverID)
	if err != nil {
		log.Log(logger.Error, fmt.Sprintf("failed to get SSH manager: %v", err), payload.OrgID)
		return w.failBackup(ctx, rowID, fmt.Sprintf("SSH connection failed: %v", err))
	}

	mac := hmac.New(sha256.New, []byte(config.AppConfig.BetterAuth.Secret))
	mac.Write([]byte(serverID.String()))
	resticPassword := hex.EncodeToString(mac.Sum(nil))

	s3cfg := config.AppConfig.S3
	scheme := "https"
	if !s3cfg.UseSSL {
		scheme = "http"
	}
	repoURL := fmt.Sprintf("s3:%s://%s/%s/backups/%s/%s",
		scheme, s3cfg.Endpoint, s3cfg.Bucket, orgID.String(), serverID.String())

	envPrefix := fmt.Sprintf(
		"RESTIC_REPOSITORY=%s RESTIC_PASSWORD=%s AWS_ACCESS_KEY_ID=%s AWS_SECRET_ACCESS_KEY=%s AWS_DEFAULT_REGION=%s",
		repoURL, resticPassword,
		s3cfg.AccessKey, s3cfg.SecretKey,
		regionOrDefault(s3cfg.Region),
	)

	installCmd := `which restic >/dev/null 2>&1 || (mkdir -p ~/.local/bin && curl -fsSL https://github.com/restic/restic/releases/latest/download/restic_linux_amd64.bz2 | bzcat > ~/.local/bin/restic && chmod +x ~/.local/bin/restic)`
	if _, err := sshMgr.RunCommand(installCmd); err != nil {
		return w.failBackup(ctx, rowID, fmt.Sprintf("failed to install restic: %v", err))
	}

	resticBin := "restic"
	if out, err := sshMgr.RunCommand("which restic 2>/dev/null || echo ~/.local/bin/restic"); err == nil {
		resticBin = strings.TrimSpace(out)
	}

	initCmd := fmt.Sprintf(`%s %s init 2>&1 || true`, envPrefix, resticBin)
	if _, err := sshMgr.RunCommand(initCmd); err != nil {
		return w.failBackup(ctx, rowID, fmt.Sprintf("restic init failed: %v", err))
	}

	paths := payload.BackupPaths
	if len(paths) == 0 {
		paths = defaultBackupPaths
	}
	backupCmd := fmt.Sprintf(`%s %s backup --json %s 2>&1`,
		envPrefix, resticBin, strings.Join(paths, " "))
	backupOut, backupErr := sshMgr.RunCommand(backupCmd)

	if backupErr != nil {
		errMsg := backupOut
		if len(errMsg) > 500 {
			errMsg = errMsg[:500]
		}
		return w.failBackup(ctx, rowID, fmt.Sprintf("restic backup failed: %s", errMsg))
	}

	snapshotID, sizeBytes := parseResticSummary(backupOut)
	s3Path := fmt.Sprintf("backups/%s/%s/snapshots/%s", orgID.String(), serverID.String(), snapshotID)

	retention := payload.RetentionCount
	if retention <= 0 {
		retention = 7
	}
	pruneCmd := fmt.Sprintf(`%s %s forget --keep-last %d --prune --json 2>&1`,
		envPrefix, resticBin, retention)
	if _, err := sshMgr.RunCommand(pruneCmd); err != nil {
		log.Log(logger.Warning, fmt.Sprintf("restic prune warning (non-fatal): %v", err), payload.OrgID)
	}

	completedAt := time.Now()
	if err := w.backupStore.UpdateBackupStatus(ctx, rowID, machine_types.BackupStatusCompleted, map[string]interface{}{
		"snapshot_path": snapshotID,
		"s3_path":       s3Path,
		"size_bytes":    sizeBytes,
		"completed_at":  completedAt,
	}); err != nil {
		log.Log(logger.Error, fmt.Sprintf("failed to mark backup completed: %v", err), payload.OrgID)
	}

	log.Log(logger.Info, fmt.Sprintf("backup completed: snapshot=%s size=%d org=%s", snapshotID, sizeBytes, payload.OrgID), "")
	return nil
}

func (w *BackupWorker) failBackup(ctx context.Context, rowID uuid.UUID, errMsg string) error {
	_ = w.backupStore.UpdateBackupStatus(ctx, rowID, machine_types.BackupStatusFailed, map[string]interface{}{
		"error":        errMsg,
		"completed_at": time.Now(),
	})
	return nil
}

func regionOrDefault(r string) string {
	if r == "" {
		return "us-east-1"
	}
	return r
}

// parseResticSummary scans restic --json output for the summary line and returns
// snapshot ID and total bytes processed. Returns empty string and 0 on parse failure.
func parseResticSummary(output string) (snapshotID string, sizeBytes int64) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, `"message_type":"summary"`) {
			if idx := strings.Index(line, `"snapshot_id":"`); idx >= 0 {
				rest := line[idx+len(`"snapshot_id":"`):]
				if end := strings.Index(rest, `"`); end >= 0 {
					snapshotID = rest[:end]
				}
			}
			var n int64
			if _, err := fmt.Sscanf(extractJSONField(line, "total_bytes_processed"), "%d", &n); err == nil {
				sizeBytes = n
			}
			return
		}
	}
	return
}

func extractJSONField(line, key string) string {
	needle := fmt.Sprintf(`"%s":`, key)
	idx := strings.Index(line, needle)
	if idx < 0 {
		return ""
	}
	rest := line[idx+len(needle):]
	end := strings.IndexAny(rest, ",}")
	if end < 0 {
		return rest
	}
	return strings.TrimSpace(rest[:end])
}

func SetupMachineBackupQueue(ctx context.Context, db *bun.DB) {
	onceMachineBackup.Do(func() {
		machineBackupQueue = registerProducerQueue(&taskq.QueueOptions{
			Name: queueMachineBackup,
		})

		backupWorker = newBackupWorker(db, ctx)

		taskMachineBackupTask = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       taskMachineBackup,
			RetryLimit: 1,
			Handler:    machineBackupTaskHandler,
		})

		backupReplyMux = NewReplyMultiplexerWithPrefix(backupReplyPrefix)
		backupReplyMux.Start(ctx)

		apilog.Printf("Machine backup queue and reply multiplexer initialized")
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
