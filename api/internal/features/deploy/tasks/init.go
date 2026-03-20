package tasks

import (
	"context"
	"sync"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"

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
	LiveDevQueue          taskq.Queue
	TaskLiveDev           *taskq.Task
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
	QUEUE_LIVE_DEV          = "live-dev"
	TASK_LIVE_DEV           = "task_live_dev"
)

func (t *TaskService) SetupCreateDeploymentQueue() {
	onceQueues.Do(func() {
		CreateDeploymentQueue = queue.RegisterQueue(&taskq.QueueOptions{
			Name:                QUEUE_CREATE_DEPLOYMENT,
			ConsumerIdleTimeout: 10 * time.Minute,
			MinNumWorker:        4,
			MaxNumWorker:        16,
			ReservationSize:     1,
			ReservationTimeout:  15 * time.Minute,
			WaitTimeout:         5 * time.Second,
			BufferSize:          64,
		})

		TaskCreateDeployment = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       TASK_CREATE_DEPLOYMENT,
			RetryLimit: 1,
			Handler: func(ctx context.Context, data shared_types.TaskPayload) error {
				deploymentID := data.ApplicationDeployment.ID.String()
				ctx, cancel := context.WithCancel(ctx)
				defer cancel()
				t.RegisterCancellation(deploymentID, cancel)
				defer t.DeregisterCancellation(deploymentID)

				t.Logger.Log(logger.Info, "starting create deployment", data.CorrelationID)
				if err := t.BuildPack(ctx, data); err != nil {
					t.Logger.Log(logger.Error, "create deployment failed: "+err.Error(), data.CorrelationID)
					return err
				}
				t.Logger.Log(logger.Info, "create deployment completed", data.CorrelationID)
				return nil
			},
		})

		UpdateDeploymentQueue = queue.RegisterQueue(&taskq.QueueOptions{
			Name:                QUEUE_UPDATE_DEPLOYMENT,
			ConsumerIdleTimeout: 10 * time.Minute,
			MinNumWorker:        4,
			MaxNumWorker:        16,
			ReservationSize:     1,
			ReservationTimeout:  15 * time.Minute,
			WaitTimeout:         5 * time.Second,
			BufferSize:          64,
		})

		TaskUpdateDeployment = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       TASK_UPDATE_DEPLOYMENT,
			RetryLimit: 1,
			Handler: func(ctx context.Context, data shared_types.TaskPayload) error {
				deploymentID := data.ApplicationDeployment.ID.String()
				ctx, cancel := context.WithCancel(ctx)
				defer cancel()
				t.RegisterCancellation(deploymentID, cancel)
				defer t.DeregisterCancellation(deploymentID)

				t.Logger.Log(logger.Info, "starting update deployment", data.CorrelationID)
				return t.HandleUpdateDeployment(ctx, data)
			},
		})

		ReDeployQueue = queue.RegisterQueue(&taskq.QueueOptions{
			Name:                QUEUE_REDEPLOYMENT,
			ConsumerIdleTimeout: 10 * time.Minute,
			MinNumWorker:        4,
			MaxNumWorker:        16,
			ReservationSize:     1,
			ReservationTimeout:  15 * time.Minute,
			WaitTimeout:         5 * time.Second,
			BufferSize:          64,
		})

		TaskReDeploy = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       TASK_REDEPLOYMENT,
			RetryLimit: 1,
			Handler: func(ctx context.Context, data shared_types.TaskPayload) error {
				deploymentID := data.ApplicationDeployment.ID.String()
				ctx, cancel := context.WithCancel(ctx)
				defer cancel()
				t.RegisterCancellation(deploymentID, cancel)
				defer t.DeregisterCancellation(deploymentID)

				t.Logger.Log(logger.Info, "starting redeploy", data.CorrelationID)
				return t.HandleReDeploy(ctx, data)
			},
		})

		RollbackQueue = queue.RegisterQueue(&taskq.QueueOptions{
			Name:                QUEUE_ROLLBACK,
			ConsumerIdleTimeout: 10 * time.Minute,
			MinNumWorker:        4,
			MaxNumWorker:        16,
			ReservationSize:     1,
			ReservationTimeout:  15 * time.Minute,
			WaitTimeout:         5 * time.Second,
			BufferSize:          64,
		})

		TaskRollback = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       TASK_ROLLBACK,
			RetryLimit: 1,
			Handler: func(ctx context.Context, data shared_types.TaskPayload) error {
				t.Logger.Log(logger.Info, "starting rollback", data.CorrelationID)
				return t.HandleRollback(ctx, data)
			},
		})

		RestartQueue = queue.RegisterQueue(&taskq.QueueOptions{
			Name:                QUEUE_RESTART,
			ConsumerIdleTimeout: 10 * time.Minute,
			MinNumWorker:        4,
			MaxNumWorker:        16,
			ReservationSize:     1,
			ReservationTimeout:  15 * time.Minute,
			WaitTimeout:         5 * time.Second,
			BufferSize:          64,
		})

		TaskRestart = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       TASK_RESTART,
			RetryLimit: 1,
			Handler: func(ctx context.Context, data shared_types.TaskPayload) error {
				t.Logger.Log(logger.Info, "starting restart", data.CorrelationID)
				return t.HandleRestart(ctx, data)
			},
		})

		LiveDevQueue = queue.RegisterQueue(&taskq.QueueOptions{
			Name:                QUEUE_LIVE_DEV,
			ConsumerIdleTimeout: 10 * time.Minute,
			MinNumWorker:        4,
			MaxNumWorker:        16,
			ReservationSize:     1,
			ReservationTimeout:  15 * time.Minute,
			WaitTimeout:         5 * time.Second,
			BufferSize:          64,
		})

		TaskLiveDev = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       TASK_LIVE_DEV,
			RetryLimit: 1,
			Handler: func(ctx context.Context, config LiveDevConfig) error {
				if err := t.HandleBuildFirstLiveDev(ctx, config); err != nil {
					t.Logger.Log(logger.Error, "live dev deployment failed: "+err.Error(), config.ApplicationID.String())
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
		err = t.PrerunCommands(ctx, d)
		if err != nil {
			return err
		}
		if err := checkCancelled(ctx); err != nil {
			return err
		}
		err = t.HandleCreateDockerfileDeployment(ctx, d)
		if err != nil {
			return err
		}
		if err := checkCancelled(ctx); err != nil {
			return err
		}
		err = t.PostRunCommands(ctx, d)
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
