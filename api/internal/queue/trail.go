package queue

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	trail_types "github.com/raghavyuva/nixopus-api/internal/features/trail/types"
	"github.com/vmihailenco/taskq/v3"
)

var (
	onceTrailQueues sync.Once
	ProvisionQueue  taskq.Queue
	TaskProvision   *taskq.Task
)

// Queue and task name constants (must match abyss consumer).
const (
	queueProvision = "provision-trail"
	taskProvision  = "task_provision_trail"
)

// SetupProvisionQueue initializes the provision queue and task.
// The handler is a no-op since actual processing happens in the abyss consumer.
func SetupProvisionQueue() {
	onceTrailQueues.Do(func() {
		ProvisionQueue = RegisterQueue(&taskq.QueueOptions{
			Name:                queueProvision,
			ConsumerIdleTimeout: 10 * time.Minute,
			MinNumWorker:        1,
			MaxNumWorker:        1,
			ReservationSize:     1,
			ReservationTimeout:  15 * time.Minute,
			WaitTimeout:         5 * time.Second,
			BufferSize:          16,
		})

		TaskProvision = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       taskProvision,
			RetryLimit: 1,
			Handler: func(ctx context.Context, payload trail_types.ProvisionPayload) error {
				fmt.Printf("[%s] task enqueued: session_id=%s, provision_details_id=%s\n",
					taskProvision, payload.SessionID, payload.ProvisionDetailsID)
				return nil
			},
		})

		log.Printf("Trail provision queue registered: %s", queueProvision)
	})
}

// EnqueueProvisionTask enqueues a provision task to the Redis queue.
func EnqueueProvisionTask(ctx context.Context, payload trail_types.ProvisionPayload) error {
	if ProvisionQueue == nil {
		return fmt.Errorf("provision queue not initialized - call SetupProvisionQueue first")
	}

	return ProvisionQueue.Add(TaskProvision.WithArgs(ctx, payload))
}
