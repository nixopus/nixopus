package tasks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/raghavyuva/caddygo"
	"github.com/raghavyuva/nixopus-api/internal/config"
	types "github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
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

// retryTracker tracks retry attempts per correlation ID
var retryTracker = sync.Map{}

func GetCaddyClient() *caddygo.Client {
	if caddyClient == nil {
		caddyClient = caddygo.NewClient(config.AppConfig.Proxy.CaddyEndpoint)
	}
	return caddyClient
}

func (t *TaskService) wrapHandlerWithRetryLogging(
	taskName string,
	retryLimit int,
	handler func(context.Context, shared_types.TaskPayload) error,
) func(context.Context, shared_types.TaskPayload) error {
	return func(ctx context.Context, data shared_types.TaskPayload) error {
		correlationID := data.CorrelationID
		if correlationID == "" {
			correlationID = fmt.Sprintf("%s-%d", taskName, time.Now().UnixNano())
		}

		// Get current attempt count (starts at 0 for first attempt)
		retryCountInterface, _ := retryTracker.LoadOrStore(correlationID, 0)
		attemptNumber := retryCountInterface.(int)
		totalAttempts := retryLimit + 1

		// Log attempt
		if attemptNumber > 0 {
			t.Logger.Log(logger.Warning, fmt.Sprintf("[%s] Retry attempt %d/%d for correlation_id=%s", taskName, attemptNumber+1, totalAttempts, correlationID), "")
		} else {
			t.Logger.Log(logger.Info, fmt.Sprintf("[%s] Start: correlation_id=%s", taskName, correlationID), "")
		}

		// Execute handler
		err := handler(ctx, data)

		if err != nil {
			// Increment attempt count for next retry
			attemptNumber++
			retryTracker.Store(correlationID, attemptNumber)

			// Log failure with retry count
			t.Logger.Log(logger.Error, fmt.Sprintf("[%s] Failed (attempt %d/%d): correlation_id=%s, error=%v", taskName, attemptNumber, totalAttempts, correlationID, err), "")

			// Clean up retry tracker after max retries
			if attemptNumber >= retryLimit {
				retryTracker.Delete(correlationID)
				t.Logger.Log(logger.Error, fmt.Sprintf("[%s] Max retries (%d) exceeded for correlation_id=%s", taskName, totalAttempts, correlationID), "")
			}

			return err
		}

		if attemptNumber > 0 {
			t.Logger.Log(logger.Info, fmt.Sprintf("[%s] Succeeded after %d retries: correlation_id=%s", taskName, attemptNumber, correlationID), "")
		} else {
			t.Logger.Log(logger.Info, fmt.Sprintf("[%s] Done: correlation_id=%s", taskName, correlationID), "")
		}
		retryTracker.Delete(correlationID)

		return nil
	}
}

// Configuration:
// - ConsumerIdleTimeout: 10 minutes
// - MinNumWorker: 2 => Minimum number of workers to start
// - MaxNumWorker: 4 => Maximum number of workers to start
// - ReservationSize: 1 => Number of tasks to reserve for each worker
// - ReservationTimeout: 15 minutes => Timeout for reserving tasks
// - WaitTimeout: 5 seconds => Timeout for waiting for a task
// - BufferSize: 100 => Buffer size for the queue
// - RetryLimit: 3 => Number of times to retry the task if it fails
// - Handler: function that handles the task
// - Name: name of the task
// - Queue: queue to enqueue the task
// - Task: task to enqueue the task

func (t *TaskService) SetupCreateDeploymentQueue() {
	onceQueues.Do(func() {
		CreateDeploymentQueue = queue.RegisterQueue(&taskq.QueueOptions{
			Name:                QUEUE_CREATE_DEPLOYMENT,
			ConsumerIdleTimeout: 10 * time.Minute,
			MinNumWorker:        2,
			MaxNumWorker:        4,
			ReservationSize:     1,
			ReservationTimeout:  15 * time.Minute,
			WaitTimeout:         5 * time.Second,
			BufferSize:          100,
		})

		TaskCreateDeployment = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       TASK_CREATE_DEPLOYMENT,
			RetryLimit: 3,
			Handler: t.wrapHandlerWithRetryLogging(TASK_CREATE_DEPLOYMENT, 3, func(ctx context.Context, data shared_types.TaskPayload) error {
				return t.BuildPack(ctx, data)
			}),
		})

		UpdateDeploymentQueue = queue.RegisterQueue(&taskq.QueueOptions{
			Name:                QUEUE_UPDATE_DEPLOYMENT,
			ConsumerIdleTimeout: 10 * time.Minute,
			MinNumWorker:        2,
			MaxNumWorker:        4,
			ReservationSize:     1,
			ReservationTimeout:  15 * time.Minute,
			WaitTimeout:         5 * time.Second,
			BufferSize:          100,
		})

		TaskUpdateDeployment = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       TASK_UPDATE_DEPLOYMENT,
			RetryLimit: 3,
			Handler: t.wrapHandlerWithRetryLogging(TASK_UPDATE_DEPLOYMENT, 3, func(ctx context.Context, data shared_types.TaskPayload) error {
				return t.HandleUpdateDeployment(ctx, data)
			}),
		})

		// Redeploy queue and task registration
		ReDeployQueue = queue.RegisterQueue(&taskq.QueueOptions{
			Name:                QUEUE_REDEPLOYMENT,
			ConsumerIdleTimeout: 10 * time.Minute,
			MinNumWorker:        2,
			MaxNumWorker:        4,
			ReservationSize:     1,
			ReservationTimeout:  15 * time.Minute,
			WaitTimeout:         5 * time.Second,
			BufferSize:          100,
		})

		TaskReDeploy = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       TASK_REDEPLOYMENT,
			RetryLimit: 3,
			Handler: t.wrapHandlerWithRetryLogging(TASK_REDEPLOYMENT, 3, func(ctx context.Context, data shared_types.TaskPayload) error {
				return t.HandleReDeploy(ctx, data)
			}),
		})

		// Rollback queue and task registration
		RollbackQueue = queue.RegisterQueue(&taskq.QueueOptions{
			Name:                QUEUE_ROLLBACK,
			ConsumerIdleTimeout: 10 * time.Minute,
			MinNumWorker:        2,
			MaxNumWorker:        4,
			ReservationSize:     1,
			ReservationTimeout:  15 * time.Minute,
			WaitTimeout:         5 * time.Second,
			BufferSize:          100,
		})

		TaskRollback = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       TASK_ROLLBACK,
			RetryLimit: 3,
			Handler: t.wrapHandlerWithRetryLogging(TASK_ROLLBACK, 3, func(ctx context.Context, data shared_types.TaskPayload) error {
				return t.HandleRollback(ctx, data)
			}),
		})

		// Restart queue and task registration
		RestartQueue = queue.RegisterQueue(&taskq.QueueOptions{
			Name:                QUEUE_RESTART,
			ConsumerIdleTimeout: 10 * time.Minute,
			MinNumWorker:        2,
			MaxNumWorker:        4,
			ReservationSize:     1,
			ReservationTimeout:  15 * time.Minute,
			WaitTimeout:         5 * time.Second,
			BufferSize:          100,
		})

		TaskRestart = taskq.RegisterTask(&taskq.TaskOptions{
			Name:       TASK_RESTART,
			RetryLimit: 3,
			Handler: t.wrapHandlerWithRetryLogging(TASK_RESTART, 3, func(ctx context.Context, data shared_types.TaskPayload) error {
				return t.HandleRestart(ctx, data)
			}),
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
