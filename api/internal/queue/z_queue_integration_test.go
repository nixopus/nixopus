package queue

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	trail_types "github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/taskq/v3"
)

func TestZ_QueueIntegration_miniredis(t *testing.T) {
	ctx := context.Background()

	t.Run("cleanup_nil_redis_client", func(t *testing.T) {
		require.NoError(t, cleanupDeadConsumers(ctx))
	})

	t.Run("not_initialized_enqueue_errors", func(t *testing.T) {
		require.Error(t, EnqueueProvisionTask(ctx, trail_types.ProvisionPayload{}))
		require.Error(t, EnqueueRegisterCustomDomain(ctx, CustomDomainPayload{}))
		require.Error(t, EnqueueRemoveCustomDomain(ctx, RemoveCustomDomainPayload{}))
		require.Error(t, EnqueueResourceUpdateTask(ctx, ResourceUpdatePayload{}))
		require.Error(t, EnqueueVMDeleteTask(ctx, VMDeletePayload{}))
		require.Error(t, EnqueueMachineVerifyTask(ctx, MachineVerifyPayload{}))
	})

	mr := miniredis.RunT(t)
	opt, err := redis.ParseURL("redis://" + mr.Addr())
	require.NoError(t, err)
	rdb := redis.NewClient(opt)
	t.Cleanup(func() { _ = rdb.Close() })

	Init(rdb)
	t.Cleanup(func() {
		_ = Close()
	})

	require.Same(t, rdb, RedisClient())

	t.Run("register_queue", func(t *testing.T) {
		_ = RegisterQueue(&taskq.QueueOptions{
			Name:         "integration-q",
			MinNumWorker: 0,
			MaxNumWorker: 1,
		})
		_ = RegisterQueue(&taskq.QueueOptions{
			Name:         "integration-q2",
			Redis:        rdb,
			MinNumWorker: 0,
			MaxNumWorker: 1,
		})
	})

	t.Run("dead_consumer_cleanup_threshold", func(t *testing.T) {
		prev := deadConsumerIdleThreshold
		deadConsumerIdleThreshold = -time.Second
		t.Cleanup(func() { deadConsumerIdleThreshold = prev })
		require.NoError(t, cleanupDeadConsumers(ctx))
	})

	require.NoError(t, cleanupDeadConsumers(ctx))

	t.Run("cleanup_removes_idle_consumer", func(t *testing.T) {
		prev := deadConsumerIdleThreshold
		deadConsumerIdleThreshold = 15 * time.Minute
		t.Cleanup(func() { deadConsumerIdleThreshold = prev })
		qname := "idle-consumer-q"
		_ = RegisterQueue(&taskq.QueueOptions{Name: qname, MinNumWorker: 0, MaxNumWorker: 1})
		stream := fmt.Sprintf("taskq:{%s}", qname)
		require.NoError(t, rdb.XGroupCreateMkStream(ctx, stream, "taskq", "0").Err())
		msgID, err := rdb.XAdd(ctx, &redis.XAddArgs{Stream: stream, Values: map[string]interface{}{"t": "1"}}).Result()
		require.NoError(t, err)
		xres, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    "taskq",
			Consumer: "idle-c",
			Streams:  []string{stream, ">"},
			Count:    1,
			Block:    time.Millisecond,
		}).Result()
		require.NoError(t, err)
		require.NotEmpty(t, xres)
		require.NoError(t, rdb.XAck(ctx, stream, "taskq", msgID).Err())
		mr.FastForward(20 * time.Minute)
		require.NoError(t, cleanupDeadConsumers(ctx))
	})

	t.Run("cleanup_skips_consumer_with_pending", func(t *testing.T) {
		qname := "pending-consumer-q"
		_ = RegisterQueue(&taskq.QueueOptions{Name: qname, MinNumWorker: 0, MaxNumWorker: 1})
		stream := fmt.Sprintf("taskq:{%s}", qname)
		require.NoError(t, rdb.XGroupCreateMkStream(ctx, stream, "taskq", "0").Err())
		_, err := rdb.XAdd(ctx, &redis.XAddArgs{Stream: stream, Values: map[string]interface{}{"task": "1"}}).Result()
		require.NoError(t, err)
		_, err = rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    "taskq",
			Consumer: "c1",
			Streams:  []string{stream, ">"},
			Count:    1,
			Block:    time.Millisecond,
		}).Result()
		require.NoError(t, err)
		require.NoError(t, cleanupDeadConsumers(ctx))
	})

	require.NoError(t, StartConsumers(ctx))
	assert.True(t, IsConsumersStarted())

	t.Run("setup_enqueue_chains", func(t *testing.T) {
		SetupProvisionQueue()
		require.NoError(t, EnqueueProvisionTask(ctx, trail_types.ProvisionPayload{}))
		require.NoError(t, EnqueueProvisionTask(ctx, trail_types.ProvisionPayload{ServerID: "srv1"}))

		SetupCustomDomainQueue()
		require.NoError(t, EnqueueRegisterCustomDomain(ctx, CustomDomainPayload{DomainID: "d"}))
		require.NoError(t, EnqueueRegisterCustomDomain(ctx, CustomDomainPayload{DomainID: "d", ServerID: "s"}))
		require.NoError(t, EnqueueRemoveCustomDomain(ctx, RemoveCustomDomainPayload{DomainID: "d"}))
		require.NoError(t, EnqueueRemoveCustomDomain(ctx, RemoveCustomDomainPayload{DomainID: "d", ServerID: "s"}))

		SetupResourceUpdateQueue()
		require.NoError(t, EnqueueResourceUpdateTask(ctx, ResourceUpdatePayload{VMName: "v"}))
		require.NoError(t, EnqueueResourceUpdateTask(ctx, ResourceUpdatePayload{VMName: "v", ServerID: "s"}))

		SetupVMDeleteQueue()
		require.NoError(t, EnqueueVMDeleteTask(ctx, VMDeletePayload{VMName: "v"}))
		require.NoError(t, EnqueueVMDeleteTask(ctx, VMDeletePayload{VMName: "v", ServerID: "s"}))
	})

	t.Run("machine_verify_enqueue", func(t *testing.T) {
		db := testVerifySQLite(t)
		SetupMachineVerifyQueue(ctx, db)
		require.Error(t, machineVerifyTaskHandler(ctx, MachineVerifyPayload{MachineID: "bad"}))
		require.NoError(t, EnqueueMachineVerifyTask(ctx, MachineVerifyPayload{
			MachineID: "00000000-0000-0000-0000-000000000001",
			OrgID:     "00000000-0000-0000-0000-000000000002",
		}))
		require.NoError(t, EnqueueMachineVerifyTask(ctx, MachineVerifyPayload{
			MachineID: "00000000-0000-0000-0000-000000000001",
			OrgID:     "00000000-0000-0000-0000-000000000002",
			ServerID:  "edge",
		}))
	})

	t.Run("machine_backup_enqueue", func(t *testing.T) {
		db := testBackupSQLite(t)
		SetupMachineBackupQueue(ctx, db)
		require.NoError(t, machineBackupTaskHandler(ctx, MachineBackupPayload{
			ExpiresAt: time.Now().Add(-time.Hour).Unix(),
		}))
		_, err := EnqueueMachineBackup(ctx, MachineBackupPayload{
			MachineName: "m",
			UserID:      "00000000-0000-0000-0000-000000000003",
			OrgID:       "00000000-0000-0000-0000-000000000004",
			ServerID:    "00000000-0000-0000-0000-000000000005",
			BackupRowID: "00000000-0000-0000-0000-000000000006",
			Trigger:     "manual",
		})
		require.NoError(t, err)
	})

	t.Run("machine_lifecycle_execute", func(t *testing.T) {
		SetupMachineLifecycleQueue(ctx)
		require.NotNil(t, replyMux)

		shortCtx, cancel := context.WithTimeout(ctx, 150*time.Millisecond)
		defer cancel()
		_, err := ExecuteMachineLifecycle(shortCtx, MachineLifecyclePayload{
			InstanceName: "vm",
			Action:       "status",
		})
		require.Error(t, err)

		rid := "11111111-1111-1111-1111-111111111111"
		go func() {
			time.Sleep(40 * time.Millisecond)
			replyMux.Dispatch(rid, `{"request_id":"`+rid+`","success":true,"action":"status"}`)
		}()
		res, err := ExecuteMachineLifecycle(ctx, MachineLifecyclePayload{
			RequestID:    rid,
			InstanceName: "vm",
			Action:       "status",
		})
		require.NoError(t, err)
		require.True(t, res.Success)

		rid2 := "22222222-2222-2222-2222-222222222222"
		go func() {
			time.Sleep(40 * time.Millisecond)
			replyMux.Dispatch(rid2, `not-json`)
		}()
		_, err = ExecuteMachineLifecycle(ctx, MachineLifecyclePayload{
			RequestID:    rid2,
			InstanceName: "vm2",
			Action:       "pause",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse")
	})

	time.Sleep(600 * time.Millisecond)
}
