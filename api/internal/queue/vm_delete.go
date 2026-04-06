package queue

import (
	"context"
	"fmt"
	"sync"

	"github.com/vmihailenco/taskq/v3"
)

var (
	onceVMDeleteQueue sync.Once
	VMDeleteQueue     taskq.Queue
	TaskVMDelete      *taskq.Task
)

const (
	queueVMDelete = "vm-delete"
	taskVMDelete  = "task_vm_delete"
)

type VMDeletePayload struct {
	VMName   string `json:"vm_name"`
	UserID   string `json:"user_id,omitempty"`
	OrgID    string `json:"org_id,omitempty"`
	ServerID string `json:"server_id,omitempty"`
}

func SetupVMDeleteQueue() {
	onceVMDeleteQueue.Do(func() {
		VMDeleteQueue = registerProducerQueue(&taskq.QueueOptions{
			Name: queueVMDelete,
		})

		TaskVMDelete = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       taskVMDelete,
			RetryLimit: 1,
			Handler: func(ctx context.Context, payload VMDeletePayload) error {
				return nil
			},
		})
	})
}

func EnqueueVMDeleteTask(ctx context.Context, payload VMDeletePayload) error {
	if TaskVMDelete == nil {
		return fmt.Errorf("vm-delete queue not initialized - call SetupVMDeleteQueue first")
	}

	q := VMDeleteQueue
	if payload.ServerID != "" {
		q = getOrCreateProducerQueue(queueVMDelete + "-" + payload.ServerID)
	}

	return q.Add(TaskVMDelete.WithArgs(ctx, payload))
}
