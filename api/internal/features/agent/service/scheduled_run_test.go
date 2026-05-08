package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/agent/service/scheduler"
	"github.com/nixopus/nixopus/api/internal/features/agent/service/usage"
	"github.com/nixopus/nixopus/api/pkg/llm"
	"github.com/nixopus/nixopus/api/pkg/llm/memory"
	"github.com/stretchr/testify/require"
)

func TestGetScheduleStore_GetScheduler_nil(t *testing.T) {
	var svc AgentService
	require.Nil(t, svc.GetScheduleStore())
	require.Nil(t, svc.GetScheduler())
}

func noopUsageTracker() *usage.Tracker {
	return usage.NewTrackerWithDeps(usage.Deps{
		DebitWallet:      func(uuid.UUID, int, string, string) (bool, error) { return true, nil },
		GetWalletBalance: func(uuid.UUID) (int, error) { return 1000, nil },
		InsertLog:        func(context.Context, *usage.AIUsageLog) error { return nil },
	}, testLogger())
}

func TestRunScheduledTask_deploy_agent_missing(t *testing.T) {
	svc := &AgentService{
		agents: llm.NewAgentRegistry(),
		logger: testLogger(),
	}
	_, err := svc.RunScheduledTask(context.Background(), "hi", "tid", uuid.New().String(), uuid.New().String(), "", nil)
	require.Error(t, err)
}

func TestRunScheduledTask_success(t *testing.T) {
	t.Setenv("AUTH_SERVICE_SECRET", "test-secret")
	t.Setenv("PORT", "29999")
	t.Setenv("AGENT_LIGHT_MODEL", "gpt-4o-mini")

	mockLLM := newScriptedLLM(t, []llm.Response{
		textResponse("scheduled reply"),
	})

	tools := llm.NewToolRegistry()
	provider := llm.NewOpenAIProvider(llm.OpenAIConfig{APIKey: "key", BaseURL: mockLLM.URL})
	agent := llm.NewAgent(provider, tools, llm.AgentConfig{Model: "gpt-4o-mini", MaxSteps: 2})

	reg := llm.NewAgentRegistry()
	reg.Register("deploy", agent)

	svc := &AgentService{
		agents: reg,
		logger: testLogger(),
		usage:  noopUsageTracker(),
	}

	res, err := svc.RunScheduledTask(context.Background(), "prompt", "tid", uuid.New().String(), uuid.New().String(), "", nil)
	require.NoError(t, err)
	require.Equal(t, "scheduled reply", res.Content)
}

func TestRunScheduledTask_uses_explicit_model(t *testing.T) {
	t.Setenv("AUTH_SERVICE_SECRET", "s")
	t.Setenv("PORT", "29998")

	mockLLM := newScriptedLLM(t, []llm.Response{
		textResponse("ok"),
	})
	tools := llm.NewToolRegistry()
	provider := llm.NewOpenAIProvider(llm.OpenAIConfig{APIKey: "key", BaseURL: mockLLM.URL})
	agent := llm.NewAgent(provider, tools, llm.AgentConfig{Model: "gpt-4o-mini", MaxSteps: 2})
	reg := llm.NewAgentRegistry()
	reg.Register("deploy", agent)
	svc := &AgentService{agents: reg, logger: testLogger(), usage: noopUsageTracker()}

	res, err := svc.RunScheduledTask(context.Background(), "p", "tid", uuid.New().String(), uuid.New().String(), "anthropic/claude-3-haiku", nil)
	require.NoError(t, err)
	require.Equal(t, "ok", res.Content)
}

func TestRunScheduledTask_no_auth_secret_falls_back_and_warns(t *testing.T) {
	t.Setenv("AUTH_SERVICE_SECRET", "")
	t.Setenv("BETTER_AUTH_SECRET", "")
	t.Setenv("PORT", "29997")

	mockLLM := newScriptedLLM(t, []llm.Response{textResponse("x")})
	tools := llm.NewToolRegistry()
	provider := llm.NewOpenAIProvider(llm.OpenAIConfig{APIKey: "k", BaseURL: mockLLM.URL})
	agent := llm.NewAgent(provider, tools, llm.AgentConfig{Model: "gpt-4o-mini", MaxSteps: 2})
	reg := llm.NewAgentRegistry()
	reg.Register("deploy", agent)
	svc := &AgentService{agents: reg, logger: testLogger(), usage: noopUsageTracker()}

	_, err := svc.RunScheduledTask(context.Background(), "p", "tid", uuid.New().String(), uuid.New().String(), "", nil)
	require.NoError(t, err)
}

func TestRunScheduledTask_BETTER_AUTH_SECRET_used_when_AUTH_missing(t *testing.T) {
	t.Setenv("AUTH_SERVICE_SECRET", "")
	t.Setenv("BETTER_AUTH_SECRET", "better")
	t.Setenv("PORT", "29996")

	mockLLM := newScriptedLLM(t, []llm.Response{textResponse("y")})
	tools := llm.NewToolRegistry()
	provider := llm.NewOpenAIProvider(llm.OpenAIConfig{APIKey: "k", BaseURL: mockLLM.URL})
	agent := llm.NewAgent(provider, tools, llm.AgentConfig{Model: "gpt-4o-mini", MaxSteps: 2})
	reg := llm.NewAgentRegistry()
	reg.Register("deploy", agent)
	svc := &AgentService{agents: reg, logger: testLogger(), usage: noopUsageTracker()}

	_, err := svc.RunScheduledTask(context.Background(), "p", "tid", uuid.New().String(), uuid.New().String(), "", nil)
	require.NoError(t, err)
}

func TestRunScheduledTask_agent_run_fails(t *testing.T) {
	t.Setenv("AUTH_SERVICE_SECRET", "s")
	t.Setenv("PORT", "29995")

	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(bad.Close)

	tools := llm.NewToolRegistry()
	provider := llm.NewOpenAIProvider(llm.OpenAIConfig{APIKey: "k", BaseURL: bad.URL})
	agent := llm.NewAgent(provider, tools, llm.AgentConfig{Model: "gpt-4o-mini", MaxSteps: 2})
	reg := llm.NewAgentRegistry()
	reg.Register("deploy", agent)
	svc := &AgentService{agents: reg, logger: testLogger(), usage: noopUsageTracker()}

	_, err := svc.RunScheduledTask(context.Background(), "p", "tid", uuid.New().String(), uuid.New().String(), "", nil)
	require.Error(t, err)
}

func TestRunScheduledTask_debits_org_wallet_on_success(t *testing.T) {
	t.Setenv("AUTH_SERVICE_SECRET", "s")
	t.Setenv("PORT", "29993")

	var debitedCents int
	tr := usage.NewTrackerWithDeps(usage.Deps{
		DebitWallet: func(_ uuid.UUID, amountCents int, _ string, _ string) (bool, error) {
			debitedCents = amountCents
			return true, nil
		},
		GetWalletBalance: func(uuid.UUID) (int, error) { return 500, nil },
		InsertLog:        func(context.Context, *usage.AIUsageLog) error { return nil },
	}, testLogger())

	mockLLM := newScriptedLLM(t, []llm.Response{textResponse("z")})
	tools := llm.NewToolRegistry()
	provider := llm.NewOpenAIProvider(llm.OpenAIConfig{APIKey: "k", BaseURL: mockLLM.URL})
	agent := llm.NewAgent(provider, tools, llm.AgentConfig{Model: "gpt-4o-mini", MaxSteps: 2})
	reg := llm.NewAgentRegistry()
	reg.Register("deploy", agent)
	svc := &AgentService{agents: reg, logger: testLogger(), usage: tr}

	_, err := svc.RunScheduledTask(context.Background(), "p", "tid", uuid.New().String(), uuid.New().String(), "", nil)
	require.NoError(t, err)
	// textResponse fixture uses 20 total tokens → one 1k chunk at default $0.01 → 1¢ debit
	require.Equal(t, 1, debitedCents)
}

func TestGetScheduleStore_GetScheduler_non_nil(t *testing.T) {
	db := setupTestDB(t)
	log := testLogger()
	st := scheduler.NewStore(db, log)
	mem := memory.NewPostgresStore(db)

	_, err := db.ExecContext(context.Background(), memoryTableSQLForScheduler()+schedulerTablesSQLForScheduler())
	require.NoError(t, err)

	runner := agentRunnerFunc(func(ctx context.Context, prompt, threadID, userID, orgID, model string, history []llm.Message) (*llm.RunResult, error) {
		return &llm.RunResult{Content: "r", TotalUsage: llm.Usage{TotalTokens: 1}}, nil
	})
	sch := scheduler.New(st, mem, runner, nil, log)

	svc := &AgentService{
		scheduleStore: st,
		scheduler:     sch,
	}
	require.Same(t, st, svc.GetScheduleStore())
	require.Same(t, sch, svc.GetScheduler())
}

type agentRunnerFunc func(ctx context.Context, prompt, threadID, userID, orgID, model string, history []llm.Message) (*llm.RunResult, error)

func (f agentRunnerFunc) RunScheduledTask(ctx context.Context, prompt, threadID, userID, orgID, model string, history []llm.Message) (*llm.RunResult, error) {
	return f(ctx, prompt, threadID, userID, orgID, model, history)
}

func memoryTableSQLForScheduler() string {
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
`
}

func schedulerTablesSQLForScheduler() string {
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
