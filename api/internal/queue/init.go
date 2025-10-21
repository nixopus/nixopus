package queue

import (
	"context"

	"github.com/go-redis/redis/v8"

	"github.com/vmihailenco/taskq/v3"
	"github.com/vmihailenco/taskq/v3/redisq"
)

var (
	redisClient *redis.Client
	factory     taskq.Factory
)

// Init initializes the queue factory with a shared Redis v8 client.
func Init(client *redis.Client) {
	redisClient = client
	factory = redisq.NewFactory()
}

// RegisterQueue registers a new queue with the shared redis client.
func RegisterQueue(opts *taskq.QueueOptions) taskq.Queue {
	if opts.Redis == nil {
		opts.Redis = redisClient
	}
	return factory.RegisterQueue(opts)
}

func StartConsumers(ctx context.Context) error {
	return factory.StartConsumers(ctx)
}

func Close() error {
	return factory.Close()
}
