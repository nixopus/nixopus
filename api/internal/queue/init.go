package queue

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/vmihailenco/taskq/v3"
	"github.com/vmihailenco/taskq/v3/redisq"
)

var (
	redisClient      *redis.Client
	factory          taskq.Factory
	onceConsumers    sync.Once
	consumersStarted bool
	registeredQueues []string
	queuesMutex      sync.RWMutex
)

// Init initializes the queue factory with a shared Redis v8 client.
func Init(client *redis.Client) {
	redisClient = client
	factory = redisq.NewFactory()
	registeredQueues = make([]string, 0)
}

// RegisterQueue registers a new queue with the shared redis client.
func RegisterQueue(opts *taskq.QueueOptions) taskq.Queue {
	if opts.Redis == nil {
		opts.Redis = redisClient
	}
	queue := factory.RegisterQueue(opts)

	// Track registered queue names for cleanup
	queuesMutex.Lock()
	registeredQueues = append(registeredQueues, opts.Name)
	queuesMutex.Unlock()

	log.Printf("Registered queue: %s (MinNumWorker: %d, MaxNumWorker: %d)", opts.Name, opts.MinNumWorker, opts.MaxNumWorker)
	return queue
}

// cleanupDeadConsumers removes dead consumers from Redis consumer groups.
// Dead consumers are those that haven't been active for more than ConsumerIdleTimeout.
// This helps prevent accumulation of dead consumer entries after restarts.
func cleanupDeadConsumers(ctx context.Context) error {
	if redisClient == nil {
		return nil
	}

	queuesMutex.RLock()
	queueNames := make([]string, len(registeredQueues))
	copy(queueNames, registeredQueues)
	queuesMutex.RUnlock()

	if len(queueNames) == 0 {
		return nil
	}

	log.Println("Cleaning up dead consumers from Redis consumer groups...")
	cleanedCount := 0

	// taskq uses "taskq" as the default consumer group name
	groupName := "taskq"

	for _, queueName := range queueNames {
		// Get stream key for this queue (taskq format: taskq:{queueName})
		streamKey := fmt.Sprintf("taskq:{%s}", queueName)

		// Get consumer information
		cmd := redisClient.XInfoConsumers(ctx, streamKey, groupName)
		if cmd.Err() != nil {
			// Consumer group might not exist yet, skip
			continue
		}

		consumers, err := cmd.Result()
		if err != nil {
			log.Printf("Warning: Failed to get consumers for queue %s: %v", queueName, err)
			continue
		}

		// Check each consumer and remove if idle for too long
		for _, consumer := range consumers {
			idleTime := time.Duration(consumer.Idle) * time.Millisecond
			// Consider consumers idle for more than 15 minutes as dead
			// (longer than ConsumerIdleTimeout to account for processing time)
			// Only remove if they have no pending messages
			if idleTime > 15*time.Minute && consumer.Pending == 0 {
				delCmd := redisClient.XGroupDelConsumer(ctx, streamKey, groupName, consumer.Name)
				if delCmd.Err() == nil {
					log.Printf("Removed dead consumer '%s' from queue '%s' (idle for %v)", consumer.Name, queueName, idleTime)
					cleanedCount++
				} else {
					log.Printf("Warning: Failed to remove dead consumer '%s' from queue '%s': %v", consumer.Name, queueName, delCmd.Err())
				}
			} else if consumer.Pending > 0 {
				log.Printf("Skipping consumer '%s' from queue '%s' (has %d pending messages, idle for %v)", consumer.Name, queueName, consumer.Pending, idleTime)
			}
		}
	}

	if cleanedCount > 0 {
		log.Printf("Cleaned up %d dead consumer(s)", cleanedCount)
	} else {
		log.Println("No dead consumers found to clean up")
	}

	return nil
}

// StartConsumers starts consumers for all registered queues. This function is idempotent
// and will only start consumers once, even if called multiple times.
// It also cleans up dead consumers from previous restarts before starting new ones.
func StartConsumers(ctx context.Context) error {
	var err error
	onceConsumers.Do(func() {
		// Clean up dead consumers before starting new ones
		if cleanupErr := cleanupDeadConsumers(ctx); cleanupErr != nil {
			log.Printf("Warning: Failed to cleanup dead consumers: %v", cleanupErr)
		}

		log.Println("Starting task queue consumers...")
		err = factory.StartConsumers(ctx)
		if err != nil {
			log.Printf("Error starting consumers: %v", err)
		} else {
			consumersStarted = true
			log.Println("Task queue consumers started successfully")
		}
	})
	return err
}

// IsConsumersStarted returns whether consumers have been started
func IsConsumersStarted() bool {
	return consumersStarted
}

// Close gracefully closes all consumers and cleans up resources
func Close() error {
	log.Println("Closing task queue consumers...")
	consumersStarted = false
	return factory.Close()
}
