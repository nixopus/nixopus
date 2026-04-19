package queue

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
	"github.com/vmihailenco/taskq/v3"
	cryptossh "golang.org/x/crypto/ssh"
)

const (
	queueMachineVerify = "machine-verify"
	taskMachineVerify  = "task_machine_verify"
)

type MachineVerifyPayload struct {
	MachineID string `json:"machine_id"`
	OrgID     string `json:"org_id"`
	ServerID  string `json:"server_id,omitempty"`
}

var (
	onceMachineVerify     sync.Once
	machineVerifyQueue    taskq.Queue
	taskMachineVerifyTask *taskq.Task
)

func SetupMachineVerifyQueue(ctx context.Context, db *bun.DB) {
	onceMachineVerify.Do(func() {
		machineVerifyQueue = registerProducerQueue(&taskq.QueueOptions{
			Name: queueMachineVerify,
		})
		taskMachineVerifyTask = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       taskMachineVerify,
			RetryLimit: 1,
			Handler: func(ctx context.Context, payload MachineVerifyPayload) error {
				return handleMachineVerify(ctx, db, payload)
			},
		})
		log.Printf("Machine verify queue initialized")
	})
}

func handleMachineVerify(ctx context.Context, db *bun.DB, payload MachineVerifyPayload) error {
	machineUUID, err := uuid.Parse(payload.MachineID)
	if err != nil {
		return fmt.Errorf("invalid machine_id: %w", err)
	}

	var sshKey shared_types.SSHKey
	err = db.NewSelect().
		Model(&sshKey).
		Where("id = ?", machineUUID).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return fmt.Errorf("failed to load ssh key: %w", err)
	}

	if sshKey.PrivateKeyEncrypted == nil || sshKey.Host == nil {
		markMachineInactive(ctx, db, payload.MachineID)
		return fmt.Errorf("ssh key missing private key or host")
	}

	signer, err := cryptossh.ParsePrivateKey([]byte(*sshKey.PrivateKeyEncrypted))
	if err != nil {
		markMachineInactive(ctx, db, payload.MachineID)
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	port := 22
	if sshKey.Port != nil {
		port = *sshKey.Port
	}
	user := "root"
	if sshKey.User != nil {
		user = *sshKey.User
	}

	config := &cryptossh.ClientConfig{
		User:            user,
		Auth:            []cryptossh.AuthMethod{cryptossh.PublicKeys(signer)},
		HostKeyCallback: cryptossh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", *sshKey.Host, port)
	client, err := cryptossh.Dial("tcp", addr, config)
	if err != nil {
		markMachineInactive(ctx, db, payload.MachineID)
		return fmt.Errorf("SSH dial failed: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		markMachineInactive(ctx, db, payload.MachineID)
		return fmt.Errorf("SSH session failed: %w", err)
	}
	defer session.Close()

	runErr := make(chan error, 1)
	go func() { runErr <- session.Run("echo ok") }()

	select {
	case err := <-runErr:
		if err != nil {
			markMachineInactive(ctx, db, payload.MachineID)
			return fmt.Errorf("SSH command failed: %w", err)
		}
	case <-ctx.Done():
		session.Close()
		markMachineInactive(ctx, db, payload.MachineID)
		return ctx.Err()
	}

	now := time.Now()
	_, err = db.NewUpdate().
		Model((*shared_types.SSHKey)(nil)).
		Set("is_active = ?", true).
		Set("last_used_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?::uuid", payload.MachineID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update ssh key status: %w", err)
	}

	log.Printf("[machine-verify] success: machine_id=%s", payload.MachineID)
	return nil
}

func markMachineInactive(ctx context.Context, db *bun.DB, machineID string) {
	_, err := db.NewUpdate().
		Model((*shared_types.SSHKey)(nil)).
		Set("is_active = ?", false).
		Set("updated_at = ?", time.Now()).
		Where("id = ?::uuid", machineID).
		Exec(ctx)
	if err != nil {
		log.Printf("[machine-verify] failed to mark inactive: machine_id=%s err=%v", machineID, err)
	}
}

func EnqueueMachineVerifyTask(ctx context.Context, payload MachineVerifyPayload) error {
	if taskMachineVerifyTask == nil {
		return fmt.Errorf("machine verify queue not initialized")
	}
	q := machineVerifyQueue
	if payload.ServerID != "" {
		q = getOrCreateProducerQueue(queueMachineVerify + "-" + payload.ServerID)
	}
	return q.Add(taskMachineVerifyTask.WithArgs(ctx, payload))
}
