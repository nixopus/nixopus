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

// ensureConsumerGroupReady ensures the consumer group exists and is positioned to read all messages
// (both existing and new). Call this at startup and before each enqueue so that any messages,
// old or new, are always eligible for the consumer to pick up.
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
		// Group already exists - check if it needs to be reset so all messages (old + new) are readable
		if errMsg == "BUSYGROUP Consumer Group name already exists" || errMsg == "BUSYGROUP" {
			// Group exists, check if it's positioned correctly to read all messages
			groupInfoCmd := redisClient.XInfoGroups(ctx, streamKey)
			if groupInfoCmd.Err() == nil {
				groups, _ := groupInfoCmd.Result()
				for _, group := range groups {
					if group.Name != groupName {
						continue
					}
					streamLenCmd := redisClient.XLen(ctx, streamKey)
					streamLen := int64(0)
					if streamLenCmd.Err() == nil {
						streamLen, _ = streamLenCmd.Result()
					}
					if streamLen == 0 {
						break
					}

					// Reset to "0" when group would miss existing messages:
					// "$" = only new messages; "0-0" or "" = initial/empty; or last-delivered-id behind first message
					needsReset := false
					switch group.LastDeliveredID {
					case "$", "0-0", "":
						needsReset = true
					default:
						rangeCmd := redisClient.XRangeN(ctx, streamKey, "-", "+", 1)
						if rangeCmd.Err() == nil {
							messages, err := rangeCmd.Result()
							if err == nil && len(messages) > 0 {
								if compareStreamID(group.LastDeliveredID, messages[0].ID) < 0 {
									needsReset = true
								}
							}
						}
					}

					// Only reset if no pending messages (avoid re-delivering in-flight work)
					if needsReset && group.Pending == 0 {
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
		} else {
			log.Printf("Warning: Failed to create consumer group '%s' for queue '%s': %v", groupName, queueName, groupCreateCmd.Err())
		}
	} else {
		log.Printf("Created consumer group '%s' for queue '%s' (starting from beginning)", groupName, queueName)
	}
}

// compareStreamID compares two Redis stream message IDs (timestamp-sequence).
// Returns: -1 if id1 < id2, 0 if id1 == id2, 1 if id1 > id2
func compareStreamID(id1, id2 string) int {
	if id1 < id2 {
		return -1
	}
	if id1 > id2 {
		return 1
	}
	return 0
}

// EnqueueProvisionTask enqueues a provision task to the Redis queue.
// Ensures the consumer group is ready to read all messages (old and new) before adding,
// so that abyss reliably picks up every message regardless of timing.
func EnqueueProvisionTask(ctx context.Context, payload trail_types.ProvisionPayload) error {
	if ProvisionQueue == nil {
		return fmt.Errorf("provision queue not initialized - call SetupProvisionQueue first")
	}

	// Ensure consumer group is positioned to read all messages before we add this one.
	// This makes delivery consistent even when abyss starts after messages exist.
	ensureConsumerGroupReady(ctx, queueProvision)

	return ProvisionQueue.Add(TaskProvision.WithArgs(ctx, payload))
}
