package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/config"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testLogger() logger.Logger {
	return logger.NewLogger()
}

type usageMock struct {
	debitCalled  bool
	debitOrgID   uuid.UUID
	debitAmount  int
	debitOK      bool
	debitErr     error
	balanceCents int
	balanceErr   error
	insertCalled bool
	insertedLog  *AIUsageLog
	insertErr    error
}

func (m *usageMock) deps() UsageDeps {
	return UsageDeps{
		DebitWallet: func(orgID uuid.UUID, amountCents int, reason string, referenceID string) (bool, error) {
			m.debitCalled = true
			m.debitOrgID = orgID
			m.debitAmount = amountCents
			return m.debitOK, m.debitErr
		},
		GetWalletBalance: func(orgID uuid.UUID) (int, error) {
			return m.balanceCents, m.balanceErr
		},
		InsertLog: func(ctx context.Context, log *AIUsageLog) error {
			m.insertCalled = true
			m.insertedLog = log
			return m.insertErr
		},
	}
}

func TestTrackUsage_Success(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	orgID := uuid.New()
	mock := &usageMock{debitOK: true, balanceCents: 4200}
	tracker := newUsageTrackerWithDeps(mock.deps(), testLogger())

	balance := tracker.TrackUsage(context.Background(), UsageTrackingParams{
		OrgID:            orgID.String(),
		ModelID:          "openai/gpt-4o",
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
		CostUsd:          0.05,
		RequestType:      "chat",
		SessionID:        "thread-1",
	})

	assert.Equal(t, 4200, balance)
	require.True(t, mock.debitCalled)
	assert.Equal(t, orgID, mock.debitOrgID)
	assert.Equal(t, 5, mock.debitAmount) // ceil(0.05 * 100) = 5

	require.True(t, mock.insertCalled)
	assert.Equal(t, orgID, mock.insertedLog.OrganizationID)
	assert.Equal(t, "openai/gpt-4o", mock.insertedLog.ModelID)
	assert.Equal(t, 100, mock.insertedLog.PromptTokens)
	assert.Equal(t, 50, mock.insertedLog.CompletionTokens)
	assert.Equal(t, 150, mock.insertedLog.TotalTokens)
	assert.Equal(t, 0.05, mock.insertedLog.CostUsd)
	assert.Equal(t, "chat", mock.insertedLog.RequestType)
	assert.Equal(t, "thread-1", mock.insertedLog.SessionID)
	assert.Equal(t, "success", mock.insertedLog.Status)
}

func TestTrackUsage_ZeroCost_NoDebit(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	orgID := uuid.New()
	mock := &usageMock{debitOK: true, balanceCents: 9999}
	tracker := newUsageTrackerWithDeps(mock.deps(), testLogger())

	balance := tracker.TrackUsage(context.Background(), UsageTrackingParams{
		OrgID:            orgID.String(),
		ModelID:          "openai/gpt-4o-mini",
		PromptTokens:     10,
		CompletionTokens: 5,
		TotalTokens:      15,
		CostUsd:          0,
		RequestType:      "stream",
	})

	assert.Equal(t, 9999, balance)
	assert.False(t, mock.debitCalled, "debit should not be called for zero cost")
	require.True(t, mock.insertCalled, "usage log should still be inserted")
	assert.Equal(t, 15, mock.insertedLog.TotalTokens)
}

func TestTrackUsage_SkippedForSelfHosted(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = true
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	mock := &usageMock{}
	tracker := newUsageTrackerWithDeps(mock.deps(), testLogger())

	balance := tracker.TrackUsage(context.Background(), UsageTrackingParams{
		OrgID:   uuid.New().String(),
		ModelID: "openai/gpt-4o",
		CostUsd: 1.0,
	})

	assert.Equal(t, -1, balance)
	assert.False(t, mock.debitCalled)
	assert.False(t, mock.insertCalled)
}

func TestTrackUsage_SkippedForEmptyOrgID(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	mock := &usageMock{}
	tracker := newUsageTrackerWithDeps(mock.deps(), testLogger())

	balance := tracker.TrackUsage(context.Background(), UsageTrackingParams{
		OrgID:   "",
		ModelID: "openai/gpt-4o",
		CostUsd: 0.50,
	})

	assert.Equal(t, -1, balance)
	assert.False(t, mock.debitCalled)
	assert.False(t, mock.insertCalled)
}

func TestTrackUsage_DebitFailure_StillLogs(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	orgID := uuid.New()
	mock := &usageMock{
		debitOK:      false,
		debitErr:     errors.New("db connection lost"),
		balanceCents: 1000,
	}
	tracker := newUsageTrackerWithDeps(mock.deps(), testLogger())

	balance := tracker.TrackUsage(context.Background(), UsageTrackingParams{
		OrgID:            orgID.String(),
		ModelID:          "openai/gpt-4o",
		PromptTokens:     200,
		CompletionTokens: 100,
		TotalTokens:      300,
		CostUsd:          0.10,
		RequestType:      "chat",
	})

	assert.Equal(t, 1000, balance, "balance should still be returned")
	assert.True(t, mock.debitCalled, "debit should be attempted")
	assert.True(t, mock.insertCalled, "usage log should still be inserted despite debit failure")
}

func TestTrackUsage_InsertFailure_ReturnsBalance(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	orgID := uuid.New()
	mock := &usageMock{
		debitOK:      true,
		balanceCents: 500,
		insertErr:    errors.New("insert failed: unique constraint"),
	}
	tracker := newUsageTrackerWithDeps(mock.deps(), testLogger())

	balance := tracker.TrackUsage(context.Background(), UsageTrackingParams{
		OrgID:   orgID.String(),
		ModelID: "openai/gpt-4o",
		CostUsd: 0.01,
	})

	assert.Equal(t, 500, balance, "balance should still be returned even if log insert fails")
	assert.True(t, mock.debitCalled)
	assert.True(t, mock.insertCalled)
}

func TestTrackUsage_BalanceError_ReturnsNegativeOne(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	orgID := uuid.New()
	mock := &usageMock{
		debitOK:    true,
		balanceErr: errors.New("connection refused"),
	}
	tracker := newUsageTrackerWithDeps(mock.deps(), testLogger())

	balance := tracker.TrackUsage(context.Background(), UsageTrackingParams{
		OrgID:   orgID.String(),
		ModelID: "openai/gpt-4o",
		CostUsd: 0.02,
	})

	assert.Equal(t, -1, balance, "should return -1 when balance lookup fails")
	assert.True(t, mock.insertCalled, "usage log should still be inserted")
}

func TestTrackUsage_CostCentsRoundsUp(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	mock := &usageMock{debitOK: true, balanceCents: 100}
	tracker := newUsageTrackerWithDeps(mock.deps(), testLogger())

	tracker.TrackUsage(context.Background(), UsageTrackingParams{
		OrgID:   uuid.New().String(),
		ModelID: "test",
		CostUsd: 0.001, // 0.1 cents → rounds up to 1 cent
	})

	require.True(t, mock.debitCalled)
	assert.Equal(t, 1, mock.debitAmount, "0.001 USD should round up to 1 cent")
}

func TestTrackUsage_DefaultStatusIsSuccess(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	mock := &usageMock{balanceCents: 100}
	tracker := newUsageTrackerWithDeps(mock.deps(), testLogger())

	tracker.TrackUsage(context.Background(), UsageTrackingParams{
		OrgID:   uuid.New().String(),
		ModelID: "test",
	})

	require.True(t, mock.insertCalled)
	assert.Equal(t, "success", mock.insertedLog.Status)
}

func TestTrackUsage_CustomStatus(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	mock := &usageMock{balanceCents: 100}
	tracker := newUsageTrackerWithDeps(mock.deps(), testLogger())

	tracker.TrackUsage(context.Background(), UsageTrackingParams{
		OrgID:        uuid.New().String(),
		ModelID:      "test",
		Status:       "error",
		ErrorMessage: "timeout",
	})

	require.True(t, mock.insertCalled)
	assert.Equal(t, "error", mock.insertedLog.Status)
	assert.Equal(t, "timeout", mock.insertedLog.ErrorMessage)
}

func TestResolveModelID(t *testing.T) {
	assert.Equal(t, "custom/model", resolveModelID("custom/model"))

	t.Setenv("AGENT_MODEL", "env/model")
	assert.Equal(t, "env/model", resolveModelID(""))

	t.Setenv("AGENT_MODEL", "")
	assert.Equal(t, "anthropic/claude-sonnet-4", resolveModelID(""))
}
