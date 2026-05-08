package scheduler

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/pkg/llm"
	"github.com/stretchr/testify/require"
)

func testToolDeps(t *testing.T) (context.Context, ToolDeps, *Store, func()) {
	t.Helper()
	_, store, mem := setupSchedulerTestDB(t)
	ctx := context.Background()
	uid := uuid.New()
	org := uuid.New()
	ctx = context.WithValue(ctx, ctxKeyTestUser, uid.String())
	ctx = context.WithValue(ctx, ctxKeyTestOrg, org.String())

	sch := New(store, mem, &mockRunner{result: &llm.RunResult{Content: "x", TotalUsage: llm.Usage{TotalTokens: 1}}}, nil, logger.NewLogger())

	deps := ToolDeps{
		Store:     store,
		Scheduler: func() *Scheduler { return sch },
		Memory:    mem,
		Logger:    logger.NewLogger(),
		UserID: func(c context.Context) string {
			v, _ := c.Value(ctxKeyTestUser).(string)
			return v
		},
		OrgID: func(c context.Context) string {
			v, _ := c.Value(ctxKeyTestOrg).(string)
			return v
		},
	}
	return ctx, deps, store, func() {}
}

type testCtxKey string

const (
	ctxKeyTestUser testCtxKey = "test_user"
	ctxKeyTestOrg  testCtxKey = "test_org"
)

func TestCreateScheduleTool_callsHandler(t *testing.T) {
	ctx, deps, _, _ := testToolDeps(t)
	td := CreateScheduleTool(deps)
	require.Equal(t, "create_schedule", td.Name)
	_, err := td.Handler(ctx, json.RawMessage(`not-json`))
	require.Error(t, err)
}

func TestCreateScheduleHandler_paths(t *testing.T) {
	ctx, deps, store, _ := testToolDeps(t)
	uidStr, _ := ctx.Value(ctxKeyTestUser).(string)
	orgStr, _ := ctx.Value(ctxKeyTestOrg).(string)

	t.Run("invalid json", func(t *testing.T) {
		_, err := createScheduleHandler(ctx, json.RawMessage(`{`), deps)
		require.Error(t, err)
	})

	t.Run("missing fields", func(t *testing.T) {
		_, err := createScheduleHandler(ctx, json.RawMessage(`{"name":"a"}`), deps)
		require.Error(t, err)
	})

	t.Run("nil scheduler", func(t *testing.T) {
		d := deps
		d.Scheduler = func() *Scheduler { return nil }
		_, err := createScheduleHandler(ctx, json.RawMessage(`{"name":"a","prompt":"p","cron_expression":"* * * * *","delivery_channel":"slack"}`), d)
		require.Error(t, err)
	})

	t.Run("invalid cron returns json", func(t *testing.T) {
		raw, err := createScheduleHandler(ctx, json.RawMessage(`{"name":"a","prompt":"p","cron_expression":"@@@bad","delivery_channel":"slack"}`), deps)
		require.NoError(t, err)
		require.Contains(t, string(raw), "invalid cron")
	})

	t.Run("missing user", func(t *testing.T) {
		d := deps
		d.UserID = func(context.Context) string { return "" }
		_, err := createScheduleHandler(ctx, json.RawMessage(`{"name":"a","prompt":"p","cron_expression":"* * * * *","delivery_channel":"slack"}`), d)
		require.Error(t, err)
	})

	t.Run("missing org", func(t *testing.T) {
		d := deps
		d.OrgID = func(context.Context) string { return "" }
		_, err := createScheduleHandler(ctx, json.RawMessage(`{"name":"a","prompt":"p","cron_expression":"* * * * *","delivery_channel":"slack"}`), d)
		require.Error(t, err)
	})

	t.Run("bad user uuid", func(t *testing.T) {
		d := deps
		d.UserID = func(context.Context) string { return "nope" }
		_, err := createScheduleHandler(ctx, json.RawMessage(`{"name":"a","prompt":"p","cron_expression":"* * * * *","delivery_channel":"slack"}`), d)
		require.Error(t, err)
	})

	t.Run("bad org uuid", func(t *testing.T) {
		d := deps
		d.OrgID = func(context.Context) string { return "nope" }
		_, err := createScheduleHandler(ctx, json.RawMessage(`{"name":"a","prompt":"p","cron_expression":"* * * * *","delivery_channel":"slack"}`), d)
		require.Error(t, err)
	})

	t.Run("default notify_on invalid becomes smart", func(t *testing.T) {
		raw, err := createScheduleHandler(ctx, json.RawMessage(`{"name":"a","prompt":"p","cron_expression":"* * * * *","delivery_channel":"slack","notify_on":"bogus"}`), deps)
		require.NoError(t, err)
		var out map[string]interface{}
		require.NoError(t, json.Unmarshal(raw, &out))
		require.Equal(t, "smart", out["notify_on"])
	})

	t.Run("explicit notify_on always", func(t *testing.T) {
		raw, err := createScheduleHandler(ctx, json.RawMessage(`{"name":"a2","prompt":"p","cron_expression":"* * * * *","delivery_channel":"slack","notify_on":"always"}`), deps)
		require.NoError(t, err)
		require.Contains(t, string(raw), `"notify_on":"always"`)
	})

	t.Run("create thread fails on closed db", func(t *testing.T) {
		_, dstore, mem := setupSchedulerTestDB(t)
		d := deps
		d.Store, d.Memory = dstore, mem
		_ = dstore.DB.Close()
		_, err := createScheduleHandler(ctx, json.RawMessage(`{"name":"a","prompt":"p","cron_expression":"* * * * *","delivery_channel":"slack"}`), d)
		require.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		raw, err := createScheduleHandler(ctx, json.RawMessage(`{"name":"okjob","prompt":"p","cron_expression":"* * * * *","delivery_channel":"slack","notify_on":"failure_only"}`), deps)
		require.NoError(t, err)
		require.Contains(t, string(raw), "schedule_id")
	})

	t.Run("CreateSchedule insert fails after thread", func(t *testing.T) {
		ctx2, d2, st2, _ := testToolDeps(t)
		_, err := st2.DB.ExecContext(ctx2, `DROP TABLE agent_schedules`)
		require.NoError(t, err)
		_, err = createScheduleHandler(ctx2, json.RawMessage(`{"name":"insfail","prompt":"p","cron_expression":"* * * * *","delivery_channel":"slack"}`), d2)
		require.Error(t, err)
	})

	t.Run("CreateSchedule with preset ID", func(t *testing.T) {
		fixed := uuid.New()
		sch := &Schedule{
			ID:              fixed,
			UserID:          uidStr,
			OrgID:           orgStr,
			ThreadID:        "thr-preset",
			Name:            "preset",
			Prompt:          "p",
			CronExpression:  "* * * * *",
			Status:          StatusActive,
			DeliveryChannel: "slack",
			NotifyOn:        "smart",
		}
		past := time.Now().UTC().Add(-time.Minute)
		sch.NextRunAt = &past
		require.NoError(t, store.CreateSchedule(context.Background(), sch))
		require.Equal(t, fixed, sch.ID)
	})
}

func TestListSchedulesTool(t *testing.T) {
	ctx, deps, store, _ := testToolDeps(t)
	td := ListSchedulesTool(deps)
	require.Equal(t, "list_schedules", td.Name)

	uidStr, _ := ctx.Value(ctxKeyTestUser).(string)
	orgStr, _ := ctx.Value(ctxKeyTestOrg).(string)
	past := time.Now().UTC().Add(-time.Minute)
	last := time.Now().UTC().Add(-30 * time.Minute)
	require.NoError(t, store.CreateSchedule(context.Background(), &Schedule{
		ID: uuid.New(), UserID: uidStr, OrgID: orgStr, ThreadID: "tL",
		Name: "L", Prompt: "p", CronExpression: "* * * * *", Status: StatusActive,
		NextRunAt: &past, LastRunAt: &last, DeliveryChannel: "slack",
	}))

	raw, err := td.Handler(ctx, json.RawMessage(`{}`))
	require.NoError(t, err)
	require.Contains(t, string(raw), "schedules")
	require.Contains(t, string(raw), "last_run_at")

	_ = store.DB.Close()
	_, err = listSchedulesHandler(ctx, deps)
	require.Error(t, err)
}

func TestDeleteScheduleTool(t *testing.T) {
	ctx, deps, store, _ := testToolDeps(t)
	td := DeleteScheduleTool(deps)
	_, err := td.Handler(ctx, json.RawMessage(`{`))
	require.Error(t, err)

	_, err = deleteScheduleHandler(ctx, json.RawMessage(`{"schedule_id":"bad"}`), deps)
	require.Error(t, err)

	uidStr, _ := ctx.Value(ctxKeyTestUser).(string)
	orgStr, _ := ctx.Value(ctxKeyTestOrg).(string)
	sid := uuid.New()
	past := time.Now().UTC().Add(-time.Minute)
	require.NoError(t, store.CreateSchedule(context.Background(), &Schedule{
		ID: sid, UserID: uidStr, OrgID: orgStr, ThreadID: "tD",
		Name: "D", Prompt: "p", CronExpression: "* * * * *", Status: StatusActive,
		NextRunAt: &past,
	}))

	raw, err := td.Handler(ctx, json.RawMessage(`{"schedule_id":"`+sid.String()+`"}`))
	require.NoError(t, err)
	require.Contains(t, string(raw), "deleted")

	_ = store.DB.Close()
	_, err = deleteScheduleHandler(ctx, json.RawMessage(`{"schedule_id":"`+uuid.New().String()+`"}`), deps)
	require.Error(t, err)
}

func TestPauseResumeScheduleTool(t *testing.T) {
	ctx, deps, store, _ := testToolDeps(t)
	pause := PauseScheduleTool(deps)
	resume := ResumeScheduleTool(deps)

	_, err := pause.Handler(ctx, json.RawMessage(`{`))
	require.Error(t, err)

	uidStr, _ := ctx.Value(ctxKeyTestUser).(string)
	orgStr, _ := ctx.Value(ctxKeyTestOrg).(string)
	sid := uuid.New()
	past := time.Now().UTC().Add(-time.Minute)
	require.NoError(t, store.CreateSchedule(context.Background(), &Schedule{
		ID: sid, UserID: uidStr, OrgID: orgStr, ThreadID: "tP",
		Name: "P", Prompt: "p", CronExpression: "* * * * *", Status: StatusActive,
		NextRunAt: &past,
	}))

	_, err = pause.Handler(ctx, json.RawMessage(`{"schedule_id":"`+sid.String()+`"}`))
	require.NoError(t, err)

	_, err = toggleScheduleHandler(ctx, json.RawMessage(`{"schedule_id":"bad"}`), deps, StatusPaused)
	require.Error(t, err)

	d := deps
	d.Scheduler = func() *Scheduler { return nil }
	_, err = toggleScheduleHandler(ctx, json.RawMessage(`{"schedule_id":"`+sid.String()+`"}`), d, StatusActive)
	require.Error(t, err)

	_, err = resume.Handler(ctx, json.RawMessage(`{"schedule_id":"`+sid.String()+`"}`))
	require.NoError(t, err)

	// schedule with invalid cron so computeNextRun is nil — covers resume branch without DB update
	sid2 := uuid.New()
	require.NoError(t, store.CreateSchedule(context.Background(), &Schedule{
		ID: sid2, UserID: uidStr, OrgID: orgStr, ThreadID: "tP2",
		Name: "P2", Prompt: "p", CronExpression: "* * * * *", Status: StatusPaused,
		NextRunAt: &past,
	}))
	require.NoError(t, store.UpdateScheduleStatus(context.Background(), sid2, StatusPaused))
	// patch cron to invalid in DB
	_, err = store.DB.ExecContext(context.Background(), `UPDATE agent_schedules SET cron_expression = ? WHERE id = ?`, "@@@bad", sid2.String())
	require.NoError(t, err)
	_, err = resume.Handler(ctx, json.RawMessage(`{"schedule_id":"`+sid2.String()+`"}`))
	require.NoError(t, err)
}

func TestToggleSchedule_schedule_not_found_for_user(t *testing.T) {
	ctx, deps, store, _ := testToolDeps(t)
	uidStr, _ := ctx.Value(ctxKeyTestUser).(string)
	orgStr, _ := ctx.Value(ctxKeyTestOrg).(string)
	sid := uuid.New()
	past := time.Now().UTC().Add(-time.Minute)
	require.NoError(t, store.CreateSchedule(ctx, &Schedule{
		ID: sid, UserID: uidStr, OrgID: orgStr, ThreadID: "tNF",
		Name: "NF", Prompt: "p", CronExpression: "* * * * *", Status: StatusActive,
		NextRunAt: &past,
	}))

	wrongUser := uuid.New().String()
	d := deps
	d.UserID = func(context.Context) string { return wrongUser }
	_, err := toggleScheduleHandler(ctx, json.RawMessage(`{"schedule_id":"`+sid.String()+`"}`), d, StatusPaused)
	require.Error(t, err)
}

func TestToggleSchedule_UpdateStatus_fails(t *testing.T) {
	ctx, deps, store, _ := testToolDeps(t)
	uidStr, _ := ctx.Value(ctxKeyTestUser).(string)
	orgStr, _ := ctx.Value(ctxKeyTestOrg).(string)
	sid := uuid.New()
	past := time.Now().UTC().Add(-time.Minute)
	require.NoError(t, store.CreateSchedule(ctx, &Schedule{
		ID: sid, UserID: uidStr, OrgID: orgStr, ThreadID: "tTrig",
		Name: "T", Prompt: "p", CronExpression: "* * * * *", Status: StatusActive,
		NextRunAt: &past,
	}))
	_, err := store.DB.ExecContext(ctx,
		`CREATE TRIGGER block_schedule_update BEFORE UPDATE ON agent_schedules BEGIN SELECT RAISE(ABORT, 'blocked'); END`)
	require.NoError(t, err)
	_, err = toggleScheduleHandler(ctx, json.RawMessage(`{"schedule_id":"`+sid.String()+`"}`), deps, StatusPaused)
	require.Error(t, err)
}

func TestStore_errors(t *testing.T) {
	_, store, _ := setupSchedulerTestDB(t)
	ctx := context.Background()

	_, err := store.GetSchedule(ctx, uuid.New())
	require.Error(t, err)

	_, err = store.GetScheduleForUser(ctx, uuid.New(), uuid.New().String())
	require.Error(t, err)

	_, err = store.ListSchedulesForUser(ctx, uuid.New().String(), uuid.New().String())
	require.NoError(t, err)

	_ = store.DB.Close()
	_, err = store.GetDueSchedules(context.Background(), time.Now())
	require.Error(t, err)
}
