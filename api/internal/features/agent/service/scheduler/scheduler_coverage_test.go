package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/pkg/llm"
	"github.com/nixopus/nixopus/api/pkg/llm/memory"
	"github.com/stretchr/testify/require"
)

type waitDeadlineRunner struct{}

func (waitDeadlineRunner) RunScheduledTask(ctx context.Context, prompt, threadID, userID, orgID, model string, history []llm.Message) (*llm.RunResult, error) {
	<-ctx.Done()
	return nil, nil
}

func TestScheduler_execute_run_timed_out(t *testing.T) {
	_, store, mem := setupSchedulerTestDB(t)
	ctx := context.Background()
	uid := uuid.New()
	require.NoError(t, mem.CreateThread(ctx, &memory.Thread{ID: "to", UserID: uid.String(), Title: "x"}))
	past := time.Now().UTC().Add(-time.Minute)
	sch := Schedule{
		ID:              uuid.New(),
		UserID:          uid.String(),
		OrgID:           uuid.New().String(),
		ThreadID:        "to",
		Name:            "to",
		Prompt:          "p",
		CronExpression:  "* * * * *",
		Status:          StatusActive,
		NextRunAt:       &past,
		TimeoutSeconds:  1,
		DeliveryChannel: "slack",
		NotifyOn:        "failure_only",
	}
	require.NoError(t, store.CreateSchedule(ctx, &sch))

	s := New(store, mem, waitDeadlineRunner{}, &mockNotifier{}, logger.NewLogger())
	s.execute(ctx, sch)
	runs, err := store.GetRunsForSchedule(ctx, sch.ID, 5)
	require.NoError(t, err)
	require.Len(t, runs, 1)
	require.Equal(t, RunTimedOut, runs[0].Status)
}

func TestScheduler_execute_default_timeout_when_zero(t *testing.T) {
	_, store, mem := setupSchedulerTestDB(t)
	ctx := context.Background()
	uid := uuid.New()
	require.NoError(t, mem.CreateThread(ctx, &memory.Thread{ID: "tz0", UserID: uid.String(), Title: "x"}))
	past := time.Now().UTC().Add(-time.Minute)
	sch := Schedule{
		ID:              uuid.New(),
		UserID:          uid.String(),
		OrgID:           uuid.New().String(),
		ThreadID:        "tz0",
		Name:            "z",
		Prompt:          "p",
		CronExpression:  "* * * * *",
		Status:          StatusActive,
		NextRunAt:       &past,
		TimeoutSeconds:  0,
		DeliveryChannel: "",
	}
	require.NoError(t, store.CreateSchedule(ctx, &sch))
	// Bun inserts apply DB default (300) when 0 is omitted; execute still needs the zero branch tested.
	sch.TimeoutSeconds = 0

	s := New(store, mem, &mockRunner{result: &llm.RunResult{Content: "ok", TotalUsage: llm.Usage{TotalTokens: 1}}}, nil, logger.NewLogger())
	s.execute(ctx, sch)
}

func TestScheduler_Stop_cancels_in_flight(t *testing.T) {
	_, st, m := setupSchedulerTestDB(t)
	ctx := context.Background()
	uid := uuid.New()
	org := uuid.New()
	require.NoError(t, m.CreateThread(ctx, &memory.Thread{ID: "stop-t", UserID: uid.String(), Title: "x"}))
	past := time.Now().UTC().Add(-time.Minute)
	sch := Schedule{
		ID:             uuid.New(),
		UserID:         uid.String(),
		OrgID:          org.String(),
		ThreadID:       "stop-t",
		Name:           "stop",
		Prompt:         "p",
		CronExpression: "* * * * *",
		Status:         StatusActive,
		NextRunAt:      &past,
		TimeoutSeconds: 60,
	}
	require.NoError(t, st.CreateSchedule(ctx, &sch))

	blocked := make(chan struct{}, 1)
	unblock := make(chan struct{})
	runner := &mockRunner{
		blocked: blocked,
		unblock: unblock,
		result:  &llm.RunResult{Content: "[OK]", TotalUsage: llm.Usage{TotalTokens: 1}},
	}

	s := New(st, m, runner, nil, logger.NewLogger())
	go s.execute(ctx, sch)
	<-blocked
	s.Stop()
	close(unblock)
	require.Eventually(t, func() bool {
		s.mu.Lock()
		defer s.mu.Unlock()
		return len(s.running) == 0
	}, 2*time.Second, 5*time.Millisecond)
}

func TestScheduler_notifyFailure_emit_error_all_channels(t *testing.T) {
	_, st, m := setupSchedulerTestDB(t)
	n := &mockNotifier{err: errors.New("emit failed")}
	s := New(st, m, &mockRunner{}, n, logger.NewLogger())
	s.notifyFailure(Schedule{
		ID:              uuid.New(),
		Name:            "n",
		UserID:          uuid.New().String(),
		OrgID:           uuid.New().String(),
		DeliveryChannel: "all",
	}, "reason")
}

func TestScheduler_execute_create_run_fails(t *testing.T) {
	db, store, mem := setupSchedulerTestDB(t)
	ctx := context.Background()
	uid := uuid.New()
	require.NoError(t, mem.CreateThread(ctx, &memory.Thread{ID: "crf", UserID: uid.String(), Title: "x"}))
	past := time.Now().UTC().Add(-time.Minute)
	sch := Schedule{
		ID:             uuid.New(),
		UserID:         uid.String(),
		OrgID:          uuid.New().String(),
		ThreadID:       "crf",
		Name:           "x",
		Prompt:         "p",
		CronExpression: "* * * * *",
		Status:         StatusActive,
		NextRunAt:      &past,
	}
	require.NoError(t, store.CreateSchedule(ctx, &sch))

	_, err := db.ExecContext(ctx, `ALTER TABLE agent_schedule_runs RENAME TO agent_schedule_runs_gone`)
	require.NoError(t, err)

	s := New(store, mem, &mockRunner{result: &llm.RunResult{Content: "x"}}, nil, logger.NewLogger())
	s.execute(ctx, sch)
}

func TestScheduler_execute_get_messages_fails(t *testing.T) {
	db, store, mem := setupSchedulerTestDB(t)
	ctx := context.Background()
	uid := uuid.New()
	require.NoError(t, mem.CreateThread(ctx, &memory.Thread{ID: "gmf", UserID: uid.String(), Title: "x"}))
	past := time.Now().UTC().Add(-time.Minute)
	sch := Schedule{
		ID:              uuid.New(),
		UserID:          uid.String(),
		OrgID:           uuid.New().String(),
		ThreadID:        "gmf",
		Name:            "x",
		Prompt:          "p",
		CronExpression:  "* * * * *",
		Status:          StatusActive,
		NextRunAt:       &past,
		DeliveryChannel: "",
	}
	require.NoError(t, store.CreateSchedule(ctx, &sch))

	_, err := db.ExecContext(ctx, `ALTER TABLE agent_messages RENAME TO agent_messages_gone`)
	require.NoError(t, err)

	s := New(store, mem, &mockRunner{result: &llm.RunResult{Content: "[OK]", TotalUsage: llm.Usage{TotalTokens: 1}}}, nil, logger.NewLogger())
	s.execute(ctx, sch)
}

func TestScheduler_execute_complete_run_fails(t *testing.T) {
	_, store, mem := setupSchedulerTestDB(t)
	ctx := context.Background()
	uid := uuid.New()
	require.NoError(t, mem.CreateThread(ctx, &memory.Thread{ID: "crun", UserID: uid.String(), Title: "x"}))
	past := time.Now().UTC().Add(-time.Minute)
	sch := Schedule{
		ID:              uuid.New(),
		UserID:          uid.String(),
		OrgID:           uuid.New().String(),
		ThreadID:        "crun",
		Name:            "x",
		Prompt:          "p",
		CronExpression:  "* * * * *",
		Status:          StatusActive,
		NextRunAt:       &past,
		DeliveryChannel: "",
	}
	require.NoError(t, store.CreateSchedule(ctx, &sch))

	runner := &mockRunner{
		result: &llm.RunResult{Content: "[OK]", TotalUsage: llm.Usage{TotalTokens: 1}},
		beforeReturn: func() {
			_, _ = store.DB.ExecContext(context.Background(), `ALTER TABLE agent_schedule_runs RENAME TO agent_schedule_runs_gone2`)
		},
	}
	s := New(store, mem, runner, nil, logger.NewLogger())
	s.execute(ctx, sch)
}

func TestScheduler_execute_update_after_run_fails(t *testing.T) {
	_, store, mem := setupSchedulerTestDB(t)
	ctx := context.Background()
	uid := uuid.New()
	require.NoError(t, mem.CreateThread(ctx, &memory.Thread{ID: "uarf", UserID: uid.String(), Title: "x"}))
	past := time.Now().UTC().Add(-time.Minute)
	sch := Schedule{
		ID:              uuid.New(),
		UserID:          uid.String(),
		OrgID:           uuid.New().String(),
		ThreadID:        "uarf",
		Name:            "x",
		Prompt:          "p",
		CronExpression:  "* * * * *",
		Status:          StatusActive,
		NextRunAt:       &past,
		DeliveryChannel: "",
	}
	require.NoError(t, store.CreateSchedule(ctx, &sch))

	runner := &mockRunner{
		result: &llm.RunResult{Content: "[OK]", TotalUsage: llm.Usage{TotalTokens: 1}},
		beforeReturn: func() {
			_, _ = store.DB.ExecContext(context.Background(), `ALTER TABLE agent_schedules RENAME TO agent_schedules_gone`)
		},
	}
	s := New(store, mem, runner, nil, logger.NewLogger())
	s.execute(ctx, sch)
}

func TestScheduler_execute_append_messages_fails(t *testing.T) {
	_, store, mem := setupSchedulerTestDB(t)
	ctx := context.Background()
	uid := uuid.New()
	require.NoError(t, mem.CreateThread(ctx, &memory.Thread{ID: "amf", UserID: uid.String(), Title: "x"}))
	past := time.Now().UTC().Add(-time.Minute)
	sch := Schedule{
		ID:              uuid.New(),
		UserID:          uid.String(),
		OrgID:           uuid.New().String(),
		ThreadID:        "amf",
		Name:            "x",
		Prompt:          "p",
		CronExpression:  "* * * * *",
		Status:          StatusActive,
		NextRunAt:       &past,
		DeliveryChannel: "",
	}
	require.NoError(t, store.CreateSchedule(ctx, &sch))

	runner := &mockRunner{
		result: &llm.RunResult{Content: "[OK]", TotalUsage: llm.Usage{TotalTokens: 1}},
		beforeReturn: func() {
			_, _ = store.DB.ExecContext(context.Background(), `DROP TABLE agent_messages`)
		},
	}
	s := New(store, mem, runner, nil, logger.NewLogger())
	s.execute(ctx, sch)
}

func TestStore_CreateSchedule_duplicate_id(t *testing.T) {
	_, store, _ := setupSchedulerTestDB(t)
	ctx := context.Background()
	id := uuid.New()
	past := time.Now().UTC().Add(-time.Hour)
	s1 := &Schedule{
		ID:              id,
		UserID:          uuid.New().String(),
		OrgID:           uuid.New().String(),
		ThreadID:        "t-dup",
		Name:            "a",
		Prompt:          "p",
		CronExpression:  "* * * * *",
		Timezone:        "UTC",
		Status:          StatusActive,
		NextRunAt:       &past,
		DeliveryChannel: "slack",
		NotifyOn:        "smart",
	}
	require.NoError(t, store.CreateSchedule(ctx, s1))
	s2 := &Schedule{
		ID:              id,
		UserID:          s1.UserID,
		OrgID:           s1.OrgID,
		ThreadID:        "t-dup2",
		Name:            "b",
		Prompt:          "p",
		CronExpression:  "* * * * *",
		Timezone:        "UTC",
		Status:          StatusActive,
		NextRunAt:       &past,
		DeliveryChannel: "slack",
		NotifyOn:        "smart",
	}
	require.Error(t, store.CreateSchedule(ctx, s2))
}
