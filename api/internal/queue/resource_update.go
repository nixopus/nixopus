package queue

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

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
}

var (
	onceResourceUpdateQueues sync.Once
	ResourceUpdateQueue      taskq.Queue
	TaskResourceUpdate       *taskq.Task
)

func SetupResourceUpdateQueue() {
	onceResourceUpdateQueues.Do(func() {
		ResourceUpdateQueue = RegisterQueue(&taskq.QueueOptions{
			Name:                queueResourceUpdate,
			ConsumerIdleTimeout: 10 * time.Minute,
			MinNumWorker:        1,
			MaxNumWorker:        1,
			ReservationSize:     1,
			ReservationTimeout:  10 * time.Minute,
			WaitTimeout:         5 * time.Second,
			BufferSize:          16,
		})

		TaskResourceUpdate = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       taskResourceUpdate,
			RetryLimit: 1,
			Handler: func(ctx context.Context, payload ResourceUpdatePayload) error {
				fmt.Printf("[%s] task enqueued: vm_name=%s, vcpu=%d, mem=%d, disk=%d\n",
					taskResourceUpdate, payload.VMName, payload.VcpuCount, payload.MemoryMB, payload.DiskGB)
				return nil
			},
		})

		log.Printf("Resource update queue registered: %s", queueResourceUpdate)
		ensureConsumerGroupReady(context.Background(), queueResourceUpdate)
	})
}

func EnqueueResourceUpdateTask(ctx context.Context, payload ResourceUpdatePayload) error {
	if ResourceUpdateQueue == nil {
		return fmt.Errorf("resource update queue not initialized - call SetupResourceUpdateQueue first")
	}
	return ResourceUpdateQueue.Add(TaskResourceUpdate.WithArgs(ctx, payload))
}
