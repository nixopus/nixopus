package scheduler

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/nixopus/nixopus/api/pkg/llm"
	"github.com/nixopus/nixopus/api/pkg/llm/memory"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	_ "modernc.org/sqlite"
)

func setupSchedulerTestDB(t *testing.T) (*bun.DB, *Store, *memory.PostgresStore) {
	t.Helper()
	sqldb, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqldb.Close() })
	db := bun.NewDB(sqldb, sqlitedialect.New())

	stmts := []string{
		memoryTableSQL(),
		schedulerTableSQL(),
	}
	for _, s := range stmts {
		_, err := db.ExecContext(context.Background(), s)
		require.NoError(t, err, s)
	}

	log := logger.NewLogger()
	return db, NewStore(db, log), memory.NewPostgresStore(db)
}

func memoryTableSQL() string {
	return `
CREATE TABLE IF NOT EXISTS agent_threads (
	id TEXT PRIMARY KEY,
	user_id TEXT NOT NULL,
	title TEXT NOT NULL DEFAULT '',
	metadata TEXT DEFAULT '{}',
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS agent_messages (
	id TEXT PRIMARY KEY,
	thread_id TEXT NOT NULL REFERENCES agent_threads(id) ON DELETE CASCADE,
	role TEXT NOT NULL,
	content TEXT NOT NULL DEFAULT '',
	tool_calls TEXT DEFAULT '[]',
	tool_call_id TEXT DEFAULT '',
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	seq INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_agent_messages_thread_seq ON agent_messages(thread_id, seq);
`
}

func schedulerTableSQL() string {
	return `
CREATE TABLE IF NOT EXISTS agent_schedules (
	id TEXT NOT NULL PRIMARY KEY,
	user_id TEXT NOT NULL,
	org_id TEXT NOT NULL,
	thread_id TEXT NOT NULL,
	name TEXT NOT NULL,
	description TEXT,
	prompt TEXT NOT NULL,
	cron_expression TEXT NOT NULL,
	timezone TEXT NOT NULL DEFAULT 'UTC',
	status TEXT NOT NULL DEFAULT 'active',
	delivery_channel TEXT,
	notify_on TEXT NOT NULL DEFAULT 'smart',
	last_run_at TIMESTAMP,
	next_run_at TIMESTAMP,
	run_count INTEGER NOT NULL DEFAULT 0,
	max_runs INTEGER,
	error_count INTEGER NOT NULL DEFAULT 0,
	max_consecutive_errors INTEGER NOT NULL DEFAULT 3,
	consecutive_error_count INTEGER NOT NULL DEFAULT 0,
	timeout_seconds INTEGER NOT NULL DEFAULT 300,
	metadata TEXT,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	deleted_at TIMESTAMP
);
CREATE TABLE IF NOT EXISTS agent_schedule_runs (
	id TEXT NOT NULL PRIMARY KEY,
	schedule_id TEXT NOT NULL,
	thread_id TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'running',
	result TEXT,
	error TEXT,
	tokens_used INTEGER NOT NULL DEFAULT 0,
	started_at TIMESTAMP NOT NULL,
	completed_at TIMESTAMP
);
`
}

func TestParseSmartDecision(t *testing.T) {
	alert, body := parseSmartDecision("[ALERT] something bad")
	require.True(t, alert)
	require.Equal(t, "something bad", body)

	ok, body := parseSmartDecision("[OK] all fine")
	require.False(t, ok)
	require.Equal(t, "all fine", body)

	ok2, body2 := parseSmartDecision("  [OK]  trimmed  ")
	require.False(t, ok2)
	require.Equal(t, "trimmed", body2)

	// fail-open: no tag
	a3, b3 := parseSmartDecision("no prefix")
	require.True(t, a3)
	require.Equal(t, "no prefix", b3)
}

func TestStoredToMessages(t *testing.T) {
	stored := []memory.StoredMessage{
		{Role: "user", Content: "hi", ToolCalls: nil},
		{Role: "assistant", Content: "yo", ToolCalls: []llm.ToolCall{{ID: "1", Type: "function", Function: llm.FunctionCall{Name: "x", Arguments: "{}"}}}},
	}
	msgs := storedToMessages(stored)
	require.Len(t, msgs, 2)
	require.Equal(t, llm.RoleUser, msgs[0].Role)
	require.Len(t, msgs[1].ToolCalls, 1)
}

func TestJSONMetadata_ValueScan(t *testing.T) {
	meta := JSONMetadata{"k": "v"}
	dv, err := meta.Value()
	require.NoError(t, err)
	require.Contains(t, dv.(string), "k")

	var nilMeta JSONMetadata
	dv2, err := nilMeta.Value()
	require.NoError(t, err)
	require.Equal(t, "{}", dv2)

	var m JSONMetadata
	require.NoError(t, m.Scan(`{"a":"b"}`))
	require.Equal(t, "b", m["a"])

	var m2 JSONMetadata
	require.NoError(t, m2.Scan([]byte(`{}`)))
	require.NotNil(t, m2)

	var m3 JSONMetadata
	require.NoError(t, m3.Scan(nil))
	require.NotNil(t, m3)

	require.Error(t, m3.Scan(42))
}

func TestStore_CRUD_and_GetDue(t *testing.T) {
	_, store, _ := setupSchedulerTestDB(t)
	ctx := context.Background()

	uid := uuid.New()
	org := uuid.New()
	threadID := "th-1"
	past := time.Now().UTC().Add(-time.Hour)
	sch := &Schedule{
		ID:              uid,
		UserID:          uid.String(),
		OrgID:           org.String(),
		ThreadID:        threadID,
		Name:            "n1",
		Prompt:          "p",
		CronExpression:  "* * * * *",
		Timezone:        "UTC",
		Status:          StatusActive,
		DeliveryChannel: "slack",
		NotifyOn:        "smart",
		NextRunAt:       &past,
	}
	require.NoError(t, store.CreateSchedule(ctx, sch))

	got, err := store.GetSchedule(ctx, uid)
	require.NoError(t, err)
	require.Equal(t, sch.Name, got.Name)

	gotUser, err := store.GetScheduleForUser(ctx, uid, uid.String())
	require.NoError(t, err)
	require.Equal(t, sch.Name, gotUser.Name)

	list, err := store.ListSchedulesForUser(ctx, uid.String(), org.String())
	require.NoError(t, err)
	require.Len(t, list, 1)

	due, err := store.GetDueSchedules(ctx, time.Now().UTC())
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(due), 1)

	require.NoError(t, store.UpdateScheduleStatus(ctx, uid, StatusPaused))

	future := time.Now().UTC().Add(time.Hour)
	require.NoError(t, store.UpdateScheduleAfterRun(ctx, uid, &future, true))

	runs, err := store.GetRunsForSchedule(ctx, uid, 10)
	require.NoError(t, err)
	require.Len(t, runs, 0)

	run := &ScheduleRun{ScheduleID: uid, ThreadID: threadID, Status: RunRunning}
	require.NoError(t, store.CreateRun(ctx, run))
	require.NoError(t, store.CompleteRun(ctx, run.ID, RunSucceeded, "ok", "", 3))

	runs2, err := store.GetRunsForSchedule(ctx, uid, 10)
	require.NoError(t, err)
	require.Len(t, runs2, 1)

	require.NoError(t, store.SoftDeleteSchedule(ctx, uid, uid.String()))
	_, err = store.GetSchedule(ctx, uid)
	require.Error(t, err)
}

type mockRunner struct {
	mu           sync.Mutex
	result       *llm.RunResult
	err          error
	blocked      chan struct{}
	unblock      chan struct{}
	beforeReturn func()
	afterRun     func()
	calls        int
	lastArgs     []llm.Message
}

func (m *mockRunner) RunScheduledTask(ctx context.Context, prompt string, threadID, userID, orgID, model string, history []llm.Message) (*llm.RunResult, error) {
	defer func() {
		if m.afterRun != nil {
			m.afterRun()
		}
	}()
	m.mu.Lock()
	m.calls++
	m.lastArgs = history
	m.mu.Unlock()
	if m.blocked != nil {
		select {
		case m.blocked <- struct{}{}:
		default:
		}
		if m.unblock != nil {
			<-m.unblock
		}
	}
	if m.beforeReturn != nil {
		m.beforeReturn()
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

type mockNotifier struct {
	mu     sync.Mutex
	events []types.NotificationEvent
	err    error
}

func (m *mockNotifier) Emit(e types.NotificationEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, e)
	return m.err
}

func (m *mockNotifier) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.events)
}

func TestScheduler_execute_notifyModes(t *testing.T) {
	_, store, mem := setupSchedulerTestDB(t)
	ctx := context.Background()
	log := logger.NewLogger()

	uid := uuid.New()
	org := uuid.New()
	threadID := "thread-xyz"
	require.NoError(t, mem.CreateThread(ctx, &memory.Thread{ID: threadID, UserID: uid.String(), Title: "t", Metadata: memory.JSONMap{"org_id": org.String()}}))

	baseSched := func(notifyOn, channel string) Schedule {
		past := time.Now().UTC().Add(-time.Minute)
		return Schedule{
			ID:                   uuid.New(),
			UserID:               uid.String(),
			OrgID:                org.String(),
			ThreadID:             threadID,
			Name:                 "job",
			Prompt:               "check",
			CronExpression:       "* * * * *",
			Timezone:             "UTC",
			Status:               StatusActive,
			DeliveryChannel:      channel,
			NotifyOn:             notifyOn,
			NextRunAt:            &past,
			TimeoutSeconds:       300,
			MaxConsecutiveErrors: 3,
		}
	}

	t.Run("smart OK skips notify", func(t *testing.T) {
		runner := &mockRunner{result: &llm.RunResult{Content: "[OK] quiet", TotalUsage: llm.Usage{TotalTokens: 1}}}
		notif := &mockNotifier{}
		s := New(store, mem, runner, notif, log)
		sch := baseSched("smart", "slack")
		require.NoError(t, store.CreateSchedule(ctx, &sch))
		s.execute(ctx, sch)
		require.Equal(t, 0, notif.count())
	})

	t.Run("smart ALERT notifies", func(t *testing.T) {
		runner := &mockRunner{result: &llm.RunResult{Content: "[ALERT] new issue", TotalUsage: llm.Usage{TotalTokens: 2}}}
		notif := &mockNotifier{}
		s := New(store, mem, runner, notif, log)
		sch := baseSched("smart", "slack")
		require.NoError(t, store.CreateSchedule(ctx, &sch))
		s.execute(ctx, sch)
		require.Equal(t, 1, notif.count())
	})

	t.Run("smart fail-open when no tag", func(t *testing.T) {
		runner := &mockRunner{result: &llm.RunResult{Content: "plain text", TotalUsage: llm.Usage{TotalTokens: 1}}}
		notif := &mockNotifier{}
		s := New(store, mem, runner, notif, log)
		sch := baseSched("smart", "slack")
		require.NoError(t, store.CreateSchedule(ctx, &sch))
		s.execute(ctx, sch)
		require.Equal(t, 1, notif.count())
	})

	t.Run("smart failure always notifies", func(t *testing.T) {
		runner := &mockRunner{err: context.DeadlineExceeded}
		notif := &mockNotifier{}
		s := New(store, mem, runner, notif, log)
		sch := baseSched("smart", "slack")
		require.NoError(t, store.CreateSchedule(ctx, &sch))
		s.execute(ctx, sch)
		require.Equal(t, 1, notif.count())
	})

	t.Run("failure_only on success skips", func(t *testing.T) {
		runner := &mockRunner{result: &llm.RunResult{Content: "ok", TotalUsage: llm.Usage{TotalTokens: 1}}}
		notif := &mockNotifier{}
		s := New(store, mem, runner, notif, log)
		sch := baseSched("failure_only", "slack")
		require.NoError(t, store.CreateSchedule(ctx, &sch))
		s.execute(ctx, sch)
		require.Equal(t, 0, notif.count())
	})

	t.Run("success_only on failure skips", func(t *testing.T) {
		runner := &mockRunner{err: context.Canceled}
		notif := &mockNotifier{}
		s := New(store, mem, runner, notif, log)
		sch := baseSched("success_only", "slack")
		require.NoError(t, store.CreateSchedule(ctx, &sch))
		s.execute(ctx, sch)
		require.Equal(t, 0, notif.count())
	})

	t.Run("never skips", func(t *testing.T) {
		runner := &mockRunner{result: &llm.RunResult{Content: "[ALERT] x", TotalUsage: llm.Usage{TotalTokens: 1}}}
		notif := &mockNotifier{}
		s := New(store, mem, runner, notif, log)
		sch := baseSched("never", "slack")
		require.NoError(t, store.CreateSchedule(ctx, &sch))
		s.execute(ctx, sch)
		require.Equal(t, 0, notif.count())
	})

	t.Run("always with default branch", func(t *testing.T) {
		runner := &mockRunner{result: &llm.RunResult{Content: "hi", TotalUsage: llm.Usage{TotalTokens: 1}}}
		notif := &mockNotifier{}
		s := New(store, mem, runner, notif, log)
		sch := baseSched("always", "slack")
		require.NoError(t, store.CreateSchedule(ctx, &sch))
		s.execute(ctx, sch)
		require.Equal(t, 1, notif.count())
	})

	t.Run("truncation long result", func(t *testing.T) {
		long := strings.Repeat("a", maxResultLength+100)
		runner := &mockRunner{result: &llm.RunResult{Content: long, TotalUsage: llm.Usage{TotalTokens: 1}}}
		notif := &mockNotifier{}
		s := New(store, mem, runner, notif, log)
		sch := baseSched("always", "slack")
		require.NoError(t, store.CreateSchedule(ctx, &sch))
		s.execute(ctx, sch)
		require.Equal(t, 1, notif.count())
	})

	t.Run("empty body uses fallback subject line path", func(t *testing.T) {
		runner := &mockRunner{result: &llm.RunResult{Content: "", TotalUsage: llm.Usage{TotalTokens: 0}}}
		notif := &mockNotifier{}
		s := New(store, mem, runner, notif, log)
		sch := baseSched("always", "slack")
		require.NoError(t, store.CreateSchedule(ctx, &sch))
		s.execute(ctx, sch)
		require.Equal(t, 1, notif.count())
	})

	t.Run("nil notifier deliver path", func(t *testing.T) {
		runner := &mockRunner{result: &llm.RunResult{Content: "x", TotalUsage: llm.Usage{TotalTokens: 1}}}
		s := New(store, mem, runner, nil, log)
		sch := baseSched("always", "slack")
		require.NoError(t, store.CreateSchedule(ctx, &sch))
		s.execute(ctx, sch) // should not panic
	})

	t.Run("notify emits error logged", func(t *testing.T) {
		runner := &mockRunner{result: &llm.RunResult{Content: "x", TotalUsage: llm.Usage{TotalTokens: 1}}}
		notif := &mockNotifier{err: context.Canceled}
		s := New(store, mem, runner, notif, log)
		sch := baseSched("always", "slack")
		require.NoError(t, store.CreateSchedule(ctx, &sch))
		s.execute(ctx, sch)
	})

	t.Run("smart instruction appended to prompt", func(t *testing.T) {
		runner2 := &mockRunner{}
		runner2.result = &llm.RunResult{Content: "[OK]", TotalUsage: llm.Usage{TotalTokens: 1}}
		wrapped := &promptCaptureRunner{inner: runner2}
		notif := &mockNotifier{}
		s := New(store, mem, wrapped, notif, log)
		sch := baseSched("smart", "")
		sch.NotifyOn = "smart"
		sch.DeliveryChannel = "" // skip notify
		require.NoError(t, store.CreateSchedule(ctx, &sch))
		s.execute(ctx, sch)
		require.Contains(t, wrapped.lastPrompt, smartNotifyInstruction)
	})
}

type promptCaptureRunner struct {
	inner      *mockRunner
	lastPrompt string
}

func (p *promptCaptureRunner) RunScheduledTask(ctx context.Context, prompt string, threadID, userID, orgID, model string, history []llm.Message) (*llm.RunResult, error) {
	p.lastPrompt = prompt
	return p.inner.RunScheduledTask(ctx, prompt, threadID, userID, orgID, model, history)
}

func TestScheduler_tick_branches(t *testing.T) {
	ctx := context.Background()
	log := logger.NewLogger()
	uid := uuid.New()
	org := uuid.New()

	t.Run("tick get due schedules db error", func(t *testing.T) {
		sqldb, err := sql.Open("sqlite", ":memory:")
		require.NoError(t, err)
		t.Cleanup(func() { _ = sqldb.Close() })
		db := bun.NewDB(sqldb, sqlitedialect.New())
		badStore := NewStore(db, log)
		memEmpty := memory.NewPostgresStore(db)
		s := New(badStore, memEmpty, &mockRunner{}, &mockNotifier{}, log)
		s.tick(context.Background())
	})

	t.Run("max runs completes", func(t *testing.T) {
		db, st, m := setupSchedulerTestDB(t)
		_ = db
		max := 1
		past := time.Now().UTC().Add(-time.Minute)
		sch := &Schedule{
			ID:             uuid.New(),
			UserID:         uid.String(),
			OrgID:          org.String(),
			ThreadID:       "t1",
			Name:           "n",
			Prompt:         "p",
			CronExpression: "* * * * *",
			Status:         StatusActive,
			NextRunAt:      &past,
			MaxRuns:        &max,
			RunCount:       1,
		}
		require.NoError(t, st.CreateSchedule(ctx, sch))
		s := New(st, m, &mockRunner{}, &mockNotifier{}, log)
		s.tick(ctx)
		got, _ := st.GetSchedule(ctx, sch.ID)
		require.Equal(t, StatusCompleted, got.Status)
	})

	t.Run("consecutive errors fails schedule", func(t *testing.T) {
		db, st, m := setupSchedulerTestDB(t)
		_ = db
		past := time.Now().UTC().Add(-time.Minute)
		sch := &Schedule{
			ID:                    uuid.New(),
			UserID:                uid.String(),
			OrgID:                 org.String(),
			ThreadID:              "t2",
			Name:                  "n2",
			Prompt:                "p",
			CronExpression:        "* * * * *",
			Status:                StatusActive,
			NextRunAt:             &past,
			DeliveryChannel:       "slack",
			ConsecutiveErrorCount: 3,
			MaxConsecutiveErrors:  3,
		}
		require.NoError(t, st.CreateSchedule(ctx, sch))
		notif := &mockNotifier{}
		s := New(st, m, &mockRunner{}, notif, log)
		s.tick(ctx)
		got, _ := st.GetSchedule(ctx, sch.ID)
		require.Equal(t, StatusFailed, got.Status)
		require.GreaterOrEqual(t, notif.count(), 1)
	})

	t.Run("already running skips second dispatch", func(t *testing.T) {
		db, st, m := setupSchedulerTestDB(t)
		_ = db
		past := time.Now().UTC().Add(-time.Minute)
		sch := &Schedule{
			ID:             uuid.New(),
			UserID:         uid.String(),
			OrgID:          org.String(),
			ThreadID:       "t3",
			Name:           "n3",
			Prompt:         "p",
			CronExpression: "* * * * *",
			Status:         StatusActive,
			NextRunAt:      &past,
			TimeoutSeconds: 60,
		}
		require.NoError(t, st.CreateSchedule(ctx, sch))
		require.NoError(t, m.CreateThread(ctx, &memory.Thread{ID: sch.ThreadID, UserID: uid.String(), Title: "x"}))

		blocked := make(chan struct{}, 1)
		unblock := make(chan struct{})
		runner := &mockRunner{blocked: blocked, unblock: unblock, result: &llm.RunResult{Content: "[OK]", TotalUsage: llm.Usage{TotalTokens: 1}}}
		execDone := make(chan struct{})
		runner.afterRun = func() { close(execDone) }
		s := New(st, m, runner, &mockNotifier{}, log)

		go s.tick(ctx)
		<-blocked
		s.tick(ctx) // should skip already running
		close(unblock)
		<-execDone
	})

	t.Run("notifyFailure no notifier", func(t *testing.T) {
		_, st, m := setupSchedulerTestDB(t)
		s := New(st, m, &mockRunner{}, nil, log)
		s.notifyFailure(Schedule{ID: uuid.New(), DeliveryChannel: "slack"}, "reason")
		s.notifyFailure(Schedule{ID: uuid.New(), DeliveryChannel: ""}, "r")
	})

	t.Run("deliver all channels", func(t *testing.T) {
		_, st, m := setupSchedulerTestDB(t)
		n := &mockNotifier{}
		s := New(st, m, &mockRunner{}, n, log)
		s.deliver(Schedule{Name: "x", DeliveryChannel: "all", UserID: uid.String(), OrgID: org.String()}, "body", "", RunSucceeded)
		require.Equal(t, 1, n.count())
	})
}

func TestScheduler_computeNextRun_and_ValidateCron(t *testing.T) {
	_, store, mem := setupSchedulerTestDB(t)
	s := New(store, mem, &mockRunner{}, nil, logger.NewLogger())

	require.NoError(t, s.ValidateCron("* * * * *"))
	require.Error(t, s.ValidateCron("not-valid"))

	sch := Schedule{CronExpression: "* * * * *", Timezone: "America/New_York"}
	nr := s.computeNextRun(sch)
	require.NotNil(t, nr)

	schBad := Schedule{ID: uuid.New(), CronExpression: "@@@invalid"}
	require.Nil(t, s.computeNextRun(schBad))

	schTZ := Schedule{CronExpression: "@every 1m", Timezone: "Invalid/Zone_XYZ"}
	_ = s.computeNextRun(schTZ) // falls back to UTC branch partially
}

func TestScheduler_StartStop_loop(t *testing.T) {
	_, store, mem := setupSchedulerTestDB(t)
	old := tickInterval
	tickInterval = 15 * time.Millisecond
	t.Cleanup(func() { tickInterval = old })

	ctx, cancel := context.WithCancel(context.Background())
	s := New(store, mem, &mockRunner{}, nil, logger.NewLogger())
	s.Start(ctx)
	time.Sleep(40 * time.Millisecond)
	cancel()
	time.Sleep(40 * time.Millisecond)
	s2 := New(store, mem, &mockRunner{}, nil, logger.NewLogger())
	s2.Start(context.Background())
	s2.Stop()
}

func TestStore_CreateTables_sqlite(t *testing.T) {
	sqldb, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer sqldb.Close()
	db := bun.NewDB(sqldb, sqlitedialect.New())
	log := logger.NewLogger()
	st := NewStore(db, log)
	st.CreateTables(context.Background())
}
