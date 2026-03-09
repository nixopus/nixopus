package queue

import (
	"context"
	"fmt"
	"sync"

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
	ServerID  string `json:"server_id,omitempty"`
}

type RemoveCustomDomainPayload struct {
	DomainID string `json:"domain_id"`
	Domain   string `json:"domain"`
	ServerID string `json:"server_id,omitempty"`
}

var (
	onceCustomDomainQueues   sync.Once
	CustomDomainQueue        taskq.Queue
	TaskRegisterCustomDomain *taskq.Task
	TaskRemoveCustomDomain   *taskq.Task
)

func SetupCustomDomainQueue() {
	onceCustomDomainQueues.Do(func() {
		CustomDomainQueue = registerProducerQueue(&taskq.QueueOptions{
			Name: queueCustomDomain,
		})

		TaskRegisterCustomDomain = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       taskRegisterCustomDomain,
			RetryLimit: 1,
			Handler: func(ctx context.Context, payload CustomDomainPayload) error {
				return nil
			},
		})

		TaskRemoveCustomDomain = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       taskRemoveCustomDomain,
			RetryLimit: 1,
			Handler: func(ctx context.Context, payload RemoveCustomDomainPayload) error {
				return nil
			},
		})
	})
}

// EnqueueRegisterCustomDomain enqueues a custom domain registration task.
// When payload.ServerID is set, the task is routed to a per-server queue
// (custom-domain-{server_id}). Otherwise it falls back to the legacy
// "custom-domain" queue for backward compatibility.
func EnqueueRegisterCustomDomain(ctx context.Context, payload CustomDomainPayload) error {
	if TaskRegisterCustomDomain == nil {
		return fmt.Errorf("custom domain queue not initialized - call SetupCustomDomainQueue first")
	}

	q := CustomDomainQueue
	if payload.ServerID != "" {
		q = getOrCreateProducerQueue(queueCustomDomain + "-" + payload.ServerID)
	}

	return q.Add(TaskRegisterCustomDomain.WithArgs(ctx, payload))
}

// EnqueueRemoveCustomDomain enqueues a custom domain removal task.
// Uses the same per-server routing as EnqueueRegisterCustomDomain.
func EnqueueRemoveCustomDomain(ctx context.Context, payload RemoveCustomDomainPayload) error {
	if TaskRemoveCustomDomain == nil {
		return fmt.Errorf("custom domain queue not initialized - call SetupCustomDomainQueue first")
	}

	q := CustomDomainQueue
	if payload.ServerID != "" {
		q = getOrCreateProducerQueue(queueCustomDomain + "-" + payload.ServerID)
	}

	return q.Add(TaskRemoveCustomDomain.WithArgs(ctx, payload))
}
