package queue

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/vmihailenco/taskq/v3"
)

const (
	queueMachineVerify = "machine-verify"
	taskMachineVerify  = "task_machine_verify"
)

type MachineVerifyPayload struct {
	MachineID string `json:"machine_id"`
	OrgID     string `json:"org_id"`
	ServerID  string `json:"server_id,omitempty"`
}

var (
	onceMachineVerify     sync.Once
	machineVerifyQueue    taskq.Queue
	taskMachineVerifyTask *taskq.Task
)

func SetupMachineVerifyQueue(ctx context.Context) {
	onceMachineVerify.Do(func() {
		machineVerifyQueue = registerProducerQueue(&taskq.QueueOptions{
			Name: queueMachineVerify,
		})
		taskMachineVerifyTask = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       taskMachineVerify,
			RetryLimit: 1,
			Handler: func(ctx context.Context, payload MachineVerifyPayload) error {
				return nil
			},
		})
		log.Printf("Machine verify queue initialized")
	})
}

func EnqueueMachineVerifyTask(ctx context.Context, payload MachineVerifyPayload) error {
	if taskMachineVerifyTask == nil {
		return fmt.Errorf("machine verify queue not initialized")
	}
	q := machineVerifyQueue
	if payload.ServerID != "" {
		q = getOrCreateProducerQueue(queueMachineVerify + "-" + payload.ServerID)
	}
	return q.Add(taskMachineVerifyTask.WithArgs(ctx, payload))
}
