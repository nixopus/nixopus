package queue

import (
	"context"
	"fmt"
	"sync"

	"github.com/vmihailenco/taskq/v3"
)

const (
	queueResourceUpdate = "resource-update"
	taskResourceUpdate  = "task_resource_update"
)

type ResourceUpdatePayload struct {
	VMName    string `json:"vm_name"`
	VcpuCount int    `json:"vcpu_count,omitempty"`
	MemoryMB  int    `json:"memory_mb,omitempty"`
	DiskGB    int    `json:"disk_gb,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	OrgID     string `json:"org_id,omitempty"`
	ServerID  string `json:"server_id,omitempty"`
}

var (
	onceResourceUpdateQueues sync.Once
	ResourceUpdateQueue      taskq.Queue
	TaskResourceUpdate       *taskq.Task
)

func SetupResourceUpdateQueue() {
	onceResourceUpdateQueues.Do(func() {
		ResourceUpdateQueue = registerProducerQueue(&taskq.QueueOptions{
			Name: queueResourceUpdate,
		})

		TaskResourceUpdate = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       taskResourceUpdate,
			RetryLimit: 1,
			Handler: func(ctx context.Context, payload ResourceUpdatePayload) error {
				return nil
			},
		})
	})
}

// EnqueueResourceUpdateTask enqueues a resource update task. When
// payload.ServerID is set, the task is routed to a per-server queue
// (resource-update-{server_id}). Otherwise it falls back to the legacy
// "resource-update" queue for backward compatibility.
func EnqueueResourceUpdateTask(ctx context.Context, payload ResourceUpdatePayload) error {
	if TaskResourceUpdate == nil {
		return fmt.Errorf("resource update queue not initialized - call SetupResourceUpdateQueue first")
	}

	q := ResourceUpdateQueue
	if payload.ServerID != "" {
		q = getOrCreateProducerQueue(queueResourceUpdate + "-" + payload.ServerID)
	}

	return q.Add(TaskResourceUpdate.WithArgs(ctx, payload))
}
