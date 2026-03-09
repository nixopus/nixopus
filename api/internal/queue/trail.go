package queue

import (
	"context"
	"fmt"
	"sync"

	trail_types "github.com/nixopus/nixopus/api/internal/features/trail/types"
	"github.com/vmihailenco/taskq/v3"
)

var (
	onceTrailQueues sync.Once
	ProvisionQueue  taskq.Queue
	TaskProvision   *taskq.Task
)

const (
	queueProvision = "provision-trail"
	taskProvision  = "task_provision_trail"
)

func SetupProvisionQueue() {
	onceTrailQueues.Do(func() {
		ProvisionQueue = registerProducerQueue(&taskq.QueueOptions{
			Name: queueProvision,
		})

		TaskProvision = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       taskProvision,
			RetryLimit: 1,
			Handler: func(ctx context.Context, payload trail_types.ProvisionPayload) error {
				return nil
			},
		})
	})
}

// EnqueueProvisionTask enqueues a provision task to the Redis queue.
// When payload.ServerID is set, the task is routed to a per-server queue
// (provision-trail-{server_id}). Otherwise it falls back to the legacy
// "provision-trail" queue for backward compatibility.
func EnqueueProvisionTask(ctx context.Context, payload trail_types.ProvisionPayload) error {
	if TaskProvision == nil {
		return fmt.Errorf("provision queue not initialized - call SetupProvisionQueue first")
	}

	q := ProvisionQueue
	if payload.ServerID != "" {
		q = getOrCreateProducerQueue(queueProvision + "-" + payload.ServerID)
	}

	return q.Add(TaskProvision.WithArgs(ctx, payload))
}
