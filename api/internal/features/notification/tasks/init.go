package tasks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/channel"
	"github.com/raghavyuva/nixopus-api/internal/queue"
	"github.com/vmihailenco/taskq/v3"
)

var (
	onceNotificationQueue sync.Once
	NotificationQueue     taskq.Queue
	TaskSendNotification  *taskq.Task
)

const (
	QUEUE_NOTIFICATION     = "notification-delivery"
	TASK_SEND_NOTIFICATION = "task_send_notification"
)

// SetupNotificationQueue registers the notification delivery queue and task
// with the shared Redis taskq factory. It follows the same pattern as the
// deploy task queues but with higher retry limits since notifications are
// non-critical and can safely be retried.
func SetupNotificationQueue(channels map[string]channel.Channel, l logger.Logger) {
	onceNotificationQueue.Do(func() {
		NotificationQueue = queue.RegisterQueue(&taskq.QueueOptions{
			Name:                QUEUE_NOTIFICATION,
			ConsumerIdleTimeout: 10 * time.Minute,
			MinNumWorker:        2,
			MaxNumWorker:        8,
			ReservationSize:     1,
			ReservationTimeout:  2 * time.Minute,
			WaitTimeout:         5 * time.Second,
			BufferSize:          64,
		})

		TaskSendNotification = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       TASK_SEND_NOTIFICATION,
			RetryLimit: 3,
			Handler: func(ctx context.Context, payload channel.DeliveryPayload) error {
				ch, ok := channels[payload.Channel]
				if !ok {
					l.Log(logger.Error, fmt.Sprintf("unknown notification channel: %s", payload.Channel), "")
					return fmt.Errorf("unknown notification channel: %s", payload.Channel)
				}

				if err := ch.Send(ctx, payload.Message); err != nil {
					l.Log(logger.Error, fmt.Sprintf("notification delivery failed on %s: %s", payload.Channel, err.Error()), "")
					return err
				}

				l.Log(logger.Info, fmt.Sprintf("notification delivered via %s to %s", payload.Channel, payload.Message.To), "")
				return nil
			},
		})
	})
}

// Enqueue adds a delivery payload to the notification queue.
func Enqueue(payload channel.DeliveryPayload) error {
	if NotificationQueue == nil || TaskSendNotification == nil {
		return fmt.Errorf("notification queue not initialized")
	}
	return NotificationQueue.Add(TaskSendNotification.WithArgs(context.Background(), payload))
}
