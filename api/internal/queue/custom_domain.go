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
	queueCustomDomain        = "custom-domain"
	taskRegisterCustomDomain = "task_register_custom_domain"
	taskRemoveCustomDomain   = "task_remove_custom_domain"
)

type CustomDomainPayload struct {
	DomainID  string `json:"domain_id"`
	Domain    string `json:"domain"`
	Subdomain string `json:"subdomain"`
	GuestIP   string `json:"guest_ip"`
}

type RemoveCustomDomainPayload struct {
	DomainID string `json:"domain_id"`
	Domain   string `json:"domain"`
}

var (
	onceCustomDomainQueues   sync.Once
	CustomDomainQueue        taskq.Queue
	TaskRegisterCustomDomain *taskq.Task
	TaskRemoveCustomDomain   *taskq.Task
)

func SetupCustomDomainQueue() {
	onceCustomDomainQueues.Do(func() {
		CustomDomainQueue = RegisterQueue(&taskq.QueueOptions{
			Name:                queueCustomDomain,
			ConsumerIdleTimeout: 10 * time.Minute,
			MinNumWorker:        1,
			MaxNumWorker:        1,
			ReservationSize:     1,
			ReservationTimeout:  15 * time.Minute,
			WaitTimeout:         5 * time.Second,
			BufferSize:          16,
		})

		TaskRegisterCustomDomain = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       taskRegisterCustomDomain,
			RetryLimit: 1,
			Handler: func(ctx context.Context, payload CustomDomainPayload) error {
				fmt.Printf("[%s] task enqueued: domain_id=%s, domain=%s\n",
					taskRegisterCustomDomain, payload.DomainID, payload.Domain)
				return nil
			},
		})

		TaskRemoveCustomDomain = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       taskRemoveCustomDomain,
			RetryLimit: 1,
			Handler: func(ctx context.Context, payload RemoveCustomDomainPayload) error {
				fmt.Printf("[%s] task enqueued: domain_id=%s, domain=%s\n",
					taskRemoveCustomDomain, payload.DomainID, payload.Domain)
				return nil
			},
		})

		log.Printf("Custom domain queue registered: %s", queueCustomDomain)
		ensureConsumerGroupReady(context.Background(), queueCustomDomain)
	})
}

func EnqueueRegisterCustomDomain(ctx context.Context, payload CustomDomainPayload) error {
	if CustomDomainQueue == nil {
		return fmt.Errorf("custom domain queue not initialized - call SetupCustomDomainQueue first")
	}
	return CustomDomainQueue.Add(TaskRegisterCustomDomain.WithArgs(ctx, payload))
}

func EnqueueRemoveCustomDomain(ctx context.Context, payload RemoveCustomDomainPayload) error {
	if CustomDomainQueue == nil {
		return fmt.Errorf("custom domain queue not initialized - call SetupCustomDomainQueue first")
	}
	return CustomDomainQueue.Add(TaskRemoveCustomDomain.WithArgs(ctx, payload))
}
