package tasks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/raghavyuva/caddygo"
	"github.com/raghavyuva/nixopus-api/internal/config"
	types "github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/queue"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/vmihailenco/taskq/v3"
)

var (
	onceQueues            sync.Once
	CreateDeploymentQueue taskq.Queue
	TaskCreateDeployment  *taskq.Task
	UpdateDeploymentQueue taskq.Queue
	TaskUpdateDeployment  *taskq.Task
	ReDeployQueue         taskq.Queue
	TaskReDeploy          *taskq.Task
	RollbackQueue         taskq.Queue
	TaskRollback          *taskq.Task
	RestartQueue          taskq.Queue
	TaskRestart           *taskq.Task
)

var (
	TASK_CREATE_DEPLOYMENT  = "task_create_deployment"
	QUEUE_CREATE_DEPLOYMENT = "create-deployment"
	QUEUE_UPDATE_DEPLOYMENT = "update-deployment"
	TASK_UPDATE_DEPLOYMENT  = "task_update_deployment"
	QUEUE_REDEPLOYMENT      = "redeploy-deployment"
	TASK_REDEPLOYMENT       = "task_redeploy_deployment"
	QUEUE_ROLLBACK          = "rollback-deployment"
	TASK_ROLLBACK           = "task_rollback_deployment"
	QUEUE_RESTART           = "restart-deployment"
	TASK_RESTART            = "task_restart_deployment"
)

var caddyClient *caddygo.Client

func GetCaddyClient() *caddygo.Client {
	if caddyClient == nil {
		caddyClient = caddygo.NewClient(config.AppConfig.Proxy.CaddyEndpoint)
	}
	return caddyClient
}

func (t *TaskService) SetupCreateDeploymentQueue() {
	onceQueues.Do(func() {
		CreateDeploymentQueue = queue.RegisterQueue(&taskq.QueueOptions{
			Name:                QUEUE_CREATE_DEPLOYMENT,
			ConsumerIdleTimeout: 10 * time.Minute,
			MinNumWorker:        1,
			MaxNumWorker:        4,
			ReservationSize:     1,
			ReservationTimeout:  15 * time.Minute,
			WaitTimeout:         5 * time.Second,
			BufferSize:          100,
		})

		TaskCreateDeployment = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       TASK_CREATE_DEPLOYMENT,
			RetryLimit: 5,
			Handler: func(ctx context.Context, data shared_types.TaskPayload) error {
				fmt.Printf("[%s] start: correlation_id=%s\n", TASK_CREATE_DEPLOYMENT, data.CorrelationID)
				err := t.BuildPack(ctx, data)
				if err != nil {
					fmt.Print("error handling create deployment: ", err)
					return err
				}
				fmt.Printf("[%s] done: correlation_id=%s\n", TASK_CREATE_DEPLOYMENT, data.CorrelationID)
				return nil
			},
		})

		UpdateDeploymentQueue = queue.RegisterQueue(&taskq.QueueOptions{
			Name:                QUEUE_UPDATE_DEPLOYMENT,
			ConsumerIdleTimeout: 10 * time.Minute,
			MinNumWorker:        1,
			MaxNumWorker:        4,
			ReservationSize:     1,
			ReservationTimeout:  15 * time.Minute,
			WaitTimeout:         5 * time.Second,
			BufferSize:          100,
		})

		TaskUpdateDeployment = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       TASK_UPDATE_DEPLOYMENT,
			RetryLimit: 5,
			Handler: func(ctx context.Context, data shared_types.TaskPayload) error {
				fmt.Println("Updating deployment")
				err := t.HandleUpdateDeployment(ctx, data)
				if err != nil {
					return err
				}
				return nil
			},
		})

		// Redeploy queue and task registration
		ReDeployQueue = queue.RegisterQueue(&taskq.QueueOptions{
			Name:                QUEUE_REDEPLOYMENT,
			ConsumerIdleTimeout: 10 * time.Minute,
			MinNumWorker:        1,
			MaxNumWorker:        4,
			ReservationSize:     1,
			ReservationTimeout:  15 * time.Minute,
			WaitTimeout:         5 * time.Second,
			BufferSize:          100,
		})

		TaskReDeploy = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       TASK_REDEPLOYMENT,
			RetryLimit: 5,
			Handler: func(ctx context.Context, data shared_types.TaskPayload) error {
				fmt.Println("Redeploying application")
				err := t.HandleReDeploy(ctx, data)
				if err != nil {
					return err
				}
				return nil
			},
		})

		// Rollback queue and task registration
		RollbackQueue = queue.RegisterQueue(&taskq.QueueOptions{
			Name:                QUEUE_ROLLBACK,
			ConsumerIdleTimeout: 10 * time.Minute,
			MinNumWorker:        1,
			MaxNumWorker:        4,
			ReservationSize:     1,
			ReservationTimeout:  15 * time.Minute,
			WaitTimeout:         5 * time.Second,
			BufferSize:          100,
		})

		TaskRollback = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       TASK_ROLLBACK,
			RetryLimit: 10,
			Handler: func(ctx context.Context, data shared_types.TaskPayload) error {
				fmt.Println("Rolling back deployment")
				err := t.HandleRollback(ctx, data)
				if err != nil {
					return err
				}
				return nil
			},
		})

		// Restart queue and task registration
		RestartQueue = queue.RegisterQueue(&taskq.QueueOptions{
			Name:                QUEUE_RESTART,
			ConsumerIdleTimeout: 10 * time.Minute,
			MinNumWorker:        1,
			MaxNumWorker:        4,
			ReservationSize:     1,
			ReservationTimeout:  15 * time.Minute,
			WaitTimeout:         5 * time.Second,
			BufferSize:          100,
		})

		TaskRestart = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       TASK_RESTART,
			RetryLimit: 5,
			Handler: func(ctx context.Context, data shared_types.TaskPayload) error {
				fmt.Println("Restarting deployment")
				err := t.HandleRestart(ctx, data)
				if err != nil {
					return err
				}
				return nil
			},
		})
	})
}

func (t *TaskService) StartConsumers(ctx context.Context) error {
	return queue.StartConsumers(ctx)
}

func (t *TaskService) BuildPack(ctx context.Context, d shared_types.TaskPayload) error {
	var err error
	switch d.Application.BuildPack {
	case shared_types.DockerFile:
		err = t.PrerunCommands(d)
		if err != nil {
			return err
		}
		err = t.HandleCreateDockerfileDeployment(ctx, d)
		if err != nil {
			return err
		}
		err = t.PostRunCommands(d)
		if err != nil {
			return err
		}
	case shared_types.DockerCompose:
		err = t.HandleCreateDockerComposeDeployment(ctx, d)
	case shared_types.Static:
		err = t.HandleCreateStaticDeployment(ctx, d)
	default:
		return types.ErrInvalidBuildPack
	}

	if err != nil {
		return err
	}
	return nil
}
