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
// Also ensures the consumer group exists and is positioned correctly to read all messages.
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

		// Ensure consumer group exists and is positioned correctly
		// This makes startup order independent - whether niixopus-api or abyss starts first,
		// the consumer group will be ready to read all messages
		ensureConsumerGroupReady(context.Background(), queueProvision)
	})
}

// ensureConsumerGroupReady ensures the consumer group exists and is positioned to read all messages.
// This is called by niixopus-api to make startup order independent.
func ensureConsumerGroupReady(ctx context.Context, queueName string) {
	redisClient := RedisClient()
	if redisClient == nil {
		log.Printf("Warning: Redis client not available, skipping consumer group setup for queue '%s'", queueName)
		return
	}

	streamKey := fmt.Sprintf("taskq:{%s}", queueName)
	groupName := "taskq"

	// Create consumer group starting from "0" (beginning) if it doesn't exist
	// This ensures all messages are readable regardless of when they were added
	groupCreateCmd := redisClient.XGroupCreateMkStream(ctx, streamKey, groupName, "0")
	if groupCreateCmd.Err() != nil {
		errMsg := groupCreateCmd.Err().Error()
		// Group already exists - check if it needs to be reset
		if errMsg == "BUSYGROUP Consumer Group name already exists" || errMsg == "BUSYGROUP" {
			// Group exists, check if it's positioned correctly
			groupInfoCmd := redisClient.XInfoGroups(ctx, streamKey)
			if groupInfoCmd.Err() == nil {
				groups, _ := groupInfoCmd.Result()
				for _, group := range groups {
					if group.Name == groupName {
						streamLenCmd := redisClient.XLen(ctx, streamKey)
						streamLen := int64(0)
						if streamLenCmd.Err() == nil {
							streamLen, _ = streamLenCmd.Result()
						}

						// If group is at "$" (only new messages) or "0-0" (initial) and there are messages,
						// reset it to "0" to read all messages (only if no pending messages)
						if (group.LastDeliveredID == "$" || group.LastDeliveredID == "0-0") &&
							streamLen > 0 && group.Pending == 0 {
							log.Printf("Queue '%s': Consumer group at '%s' with %d messages - resetting to '0' to read all",
								queueName, group.LastDeliveredID, streamLen)
							setIdCmd := redisClient.XGroupSetID(ctx, streamKey, groupName, "0")
							if setIdCmd.Err() == nil {
								log.Printf("Queue '%s': Consumer group reset to '0' - will read all messages", queueName)
							}
						}
						break
					}
				}
			}
		} else {
			log.Printf("Warning: Failed to create consumer group '%s' for queue '%s': %v", groupName, queueName, groupCreateCmd.Err())
		}
	} else {
		log.Printf("Created consumer group '%s' for queue '%s' (starting from beginning)", groupName, queueName)
	}
}

// EnqueueProvisionTask enqueues a provision task to the Redis queue.
func EnqueueProvisionTask(ctx context.Context, payload trail_types.ProvisionPayload) error {
	if ProvisionQueue == nil {
		return fmt.Errorf("provision queue not initialized - call SetupProvisionQueue first")
	}

	return ProvisionQueue.Add(TaskProvision.WithArgs(ctx, payload))
}
