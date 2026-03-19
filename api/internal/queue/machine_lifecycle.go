package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vmihailenco/taskq/v3"
)

const (
	queueMachineLifecycle = "machine-lifecycle"
	taskMachineLifecycle  = "task_machine_lifecycle"
)

type MachineLifecyclePayload struct {
	RequestID    string `json:"request_id"`
	InstanceName string `json:"instance_name"`
	Action       string `json:"action"`
	ServerID     string `json:"server_id,omitempty"`
	ExpiresAt    int64  `json:"expires_at"`
}

type MachineLifecycleResult struct {
	RequestID string          `json:"request_id"`
	Success   bool            `json:"success"`
	Action    string          `json:"action"`
	Data      json.RawMessage `json:"data,omitempty"`
	Error     string          `json:"error,omitempty"`
}

var (
	onceMachineLifecycle     sync.Once
	machineLifecycleQueue    taskq.Queue
	taskMachineLifecycleTask *taskq.Task
	replyMux                 *ReplyMultiplexer
)

func SetupMachineLifecycleQueue(ctx context.Context) {
	onceMachineLifecycle.Do(func() {
		machineLifecycleQueue = registerProducerQueue(&taskq.QueueOptions{
			Name: queueMachineLifecycle,
		})

		taskMachineLifecycleTask = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       taskMachineLifecycle,
			RetryLimit: 0,
			Handler: func(ctx context.Context, payload MachineLifecyclePayload) error {
				return nil
			},
		})

		replyMux = NewReplyMultiplexer()
		replyMux.Start(ctx)

		log.Printf("Machine lifecycle queue and reply multiplexer initialized")
	})
}

func ExecuteMachineLifecycle(ctx context.Context, payload MachineLifecyclePayload) (*MachineLifecycleResult, error) {
	if taskMachineLifecycleTask == nil || replyMux == nil {
		return nil, fmt.Errorf("machine lifecycle queue not initialized - call SetupMachineLifecycleQueue first")
	}

	requestID := uuid.New().String()
	payload.RequestID = requestID
	payload.ExpiresAt = time.Now().Add(15 * time.Second).Unix()

	waiterCh := replyMux.RegisterWaiter(requestID)
	defer replyMux.RemoveWaiter(requestID)

	q := machineLifecycleQueue
	if payload.ServerID != "" {
		q = getOrCreateProducerQueue(queueMachineLifecycle + "-" + payload.ServerID)
	}

	if err := q.Add(taskMachineLifecycleTask.WithArgs(ctx, payload)); err != nil {
		return nil, fmt.Errorf("failed to enqueue lifecycle task: %w", err)
	}

	rpcCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	select {
	case data := <-waiterCh:
		var result MachineLifecycleResult
		if err := json.Unmarshal([]byte(data), &result); err != nil {
			return nil, fmt.Errorf("failed to parse lifecycle result: %w", err)
		}
		return &result, nil
	case <-rpcCtx.Done():
		return nil, fmt.Errorf("machine operation timed out")
	}
}
