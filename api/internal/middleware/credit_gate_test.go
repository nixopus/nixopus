package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/config"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldSkipCreditCheck(t *testing.T) {
	tests := []struct {
		path   string
		expect bool
	}{
		{"/api/v1/credits/balance", true},
		{"/api/internal/credits/topup", true},
		{"/health", true},
		{"/healthz", true},
		{"/readyz", true},
		{"/metrics", true},
		{"/api/v1/agent/threads/123/messages", true},
		{"/api/v1/memory/recall", true},
		{"/api/v1/agent/chat", false},
		{"/api/v1/deploy", false},
		{"/api/v1/agent/chat/stream", false},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			assert.Equal(t, tc.expect, ShouldSkipCreditCheck(tc.path))
		})
	}
}

func TestCreditGateMiddleware_SelfHostedBypasses(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = true
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	called := false
	handler := CreditGateMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }),
		CreditGateDeps{},
		testLogger(),
	)

	req := httptest.NewRequest("POST", "/api/v1/agent/chat", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestCreditGateMiddleware_SkippedPaths(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	called := false
	handler := CreditGateMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }),
		CreditGateDeps{},
		testLogger(),
	)

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.True(t, called)
}

func TestCreditGateMiddleware_NoOrgIDPassesThrough(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	called := false
	handler := CreditGateMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }),
		CreditGateDeps{},
		testLogger(),
	)

	req := httptest.NewRequest("POST", "/api/v1/agent/chat", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.True(t, called)
}

func TestCreditGateMiddleware_InvalidOrgIDPassesThrough(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	called := false
	handler := CreditGateMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }),
		CreditGateDeps{},
		testLogger(),
	)

	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, "not-a-uuid")
	req := httptest.NewRequest("POST", "/api/v1/agent/chat", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.True(t, called)
}

func TestCreditGateMiddleware_ZeroBalanceBlocks(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	orgID := uuid.New()
	deps := CreditGateDeps{
		GetWalletBalance: func(ctx context.Context, id uuid.UUID) (int, error) {
			return 0, nil
		},
	}

	called := false
	handler := CreditGateMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }),
		deps,
		testLogger(),
	)

	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, orgID.String())
	req := httptest.NewRequest("POST", "/api/v1/agent/chat", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusPaymentRequired, rr.Code)

	var body map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&body))
	assert.Equal(t, "Payment Required", body["title"])
	assert.Equal(t, "/billing", body["upgrade_url"])
}

func TestCreditGateMiddleware_NegativeBalanceBlocks(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	orgID := uuid.New()
	deps := CreditGateDeps{
		GetWalletBalance: func(ctx context.Context, id uuid.UUID) (int, error) {
			return -50, nil
		},
	}

	called := false
	handler := CreditGateMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }),
		deps,
		testLogger(),
	)

	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, orgID.String())
	req := httptest.NewRequest("POST", "/api/v1/agent/chat", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusPaymentRequired, rr.Code)
}

func TestCreditGateMiddleware_BalanceErrorBlocks(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	orgID := uuid.New()
	deps := CreditGateDeps{
		GetWalletBalance: func(ctx context.Context, id uuid.UUID) (int, error) {
			return 0, errors.New("db error")
		},
	}

	called := false
	handler := CreditGateMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }),
		deps,
		testLogger(),
	)

	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, orgID.String())
	req := httptest.NewRequest("POST", "/api/v1/agent/chat", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusPaymentRequired, rr.Code)
}

func TestCreditGateMiddleware_PositiveBalanceAllows(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	orgID := uuid.New()
	deps := CreditGateDeps{
		GetWalletBalance: func(ctx context.Context, id uuid.UUID) (int, error) {
			return 500, nil
		},
	}

	var capturedResult *CreditCheckResult
	handler := CreditGateMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedResult = GetCreditCheckResult(r.Context())
			w.WriteHeader(http.StatusOK)
		}),
		deps,
		testLogger(),
	)

	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, orgID.String())
	req := httptest.NewRequest("POST", "/api/v1/agent/chat", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, capturedResult)
	assert.True(t, capturedResult.Allowed)
	assert.Equal(t, 500, capturedResult.BalanceCents)
	assert.Nil(t, capturedResult.MachineWarning)
}

func TestCreditGateMiddleware_MachineWarningSuspended(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	orgID := uuid.New()
	deps := CreditGateDeps{
		GetWalletBalance: func(ctx context.Context, id uuid.UUID) (int, error) {
			return 100, nil
		},
		GetMachineStatus: func(ctx context.Context, id uuid.UUID) (*MachineStatusResult, error) {
			return &MachineStatusResult{
				HasMachine: true,
				Status:     "suspended",
			}, nil
		},
	}

	var capturedResult *CreditCheckResult
	handler := CreditGateMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedResult = GetCreditCheckResult(r.Context())
		}),
		deps,
		testLogger(),
	)

	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, orgID.String())
	req := httptest.NewRequest("POST", "/api/v1/agent/chat", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.NotNil(t, capturedResult)
	require.NotNil(t, capturedResult.MachineWarning)
	assert.Equal(t, "suspended", capturedResult.MachineWarning.Status)
	assert.Contains(t, capturedResult.MachineWarning.Message, "reset due to insufficient")
	assert.NotEmpty(t, rr.Header().Get("X-Credit-Warning"))
}

func TestCreditGateMiddleware_MachineWarningGracePeriod(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	orgID := uuid.New()
	days := 3
	deadline := "2026-05-10T00:00:00Z"
	deps := CreditGateDeps{
		GetWalletBalance: func(ctx context.Context, id uuid.UUID) (int, error) {
			return 100, nil
		},
		GetMachineStatus: func(ctx context.Context, id uuid.UUID) (*MachineStatusResult, error) {
			return &MachineStatusResult{
				HasMachine:    true,
				Status:        "grace_period",
				GraceDeadline: &deadline,
				DaysRemaining: &days,
				PlanCostCents: 999,
			}, nil
		},
	}

	var capturedResult *CreditCheckResult
	handler := CreditGateMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedResult = GetCreditCheckResult(r.Context())
		}),
		deps,
		testLogger(),
	)

	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, orgID.String())
	req := httptest.NewRequest("POST", "/api/v1/agent/chat", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.NotNil(t, capturedResult)
	require.NotNil(t, capturedResult.MachineWarning)
	assert.Equal(t, "grace_period", capturedResult.MachineWarning.Status)
	assert.Contains(t, capturedResult.MachineWarning.Message, "3 days")
	assert.Contains(t, capturedResult.MachineWarning.Message, "$9.99")
	assert.Equal(t, &deadline, capturedResult.MachineWarning.GraceDeadline)
	assert.Equal(t, &days, capturedResult.MachineWarning.DaysRemaining)
	assert.NotEmpty(t, rr.Header().Get("X-Credit-Warning"))
}

func TestCreditGateMiddleware_MachineWarningGracePeriodSingularDay(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	orgID := uuid.New()
	days := 1
	deps := CreditGateDeps{
		GetWalletBalance: func(ctx context.Context, id uuid.UUID) (int, error) {
			return 100, nil
		},
		GetMachineStatus: func(ctx context.Context, id uuid.UUID) (*MachineStatusResult, error) {
			return &MachineStatusResult{
				HasMachine:    true,
				Status:        "grace_period",
				DaysRemaining: &days,
				PlanCostCents: 0,
			}, nil
		},
	}

	var capturedResult *CreditCheckResult
	handler := CreditGateMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedResult = GetCreditCheckResult(r.Context())
		}),
		deps,
		testLogger(),
	)

	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, orgID.String())
	req := httptest.NewRequest("POST", "/api/v1/agent/chat", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.NotNil(t, capturedResult)
	require.NotNil(t, capturedResult.MachineWarning)
	assert.Contains(t, capturedResult.MachineWarning.Message, "1 day.")
	assert.Contains(t, capturedResult.MachineWarning.Message, "the monthly machine cost")
	_ = rr
}

func TestCreditGateMiddleware_MachineWarningUnbilled(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	orgID := uuid.New()
	deps := CreditGateDeps{
		GetWalletBalance: func(ctx context.Context, id uuid.UUID) (int, error) {
			return 100, nil
		},
		GetMachineStatus: func(ctx context.Context, id uuid.UUID) (*MachineStatusResult, error) {
			return &MachineStatusResult{
				HasMachine: true,
				Status:     "unbilled",
			}, nil
		},
	}

	var capturedResult *CreditCheckResult
	handler := CreditGateMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedResult = GetCreditCheckResult(r.Context())
		}),
		deps,
		testLogger(),
	)

	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, orgID.String())
	req := httptest.NewRequest("POST", "/api/v1/agent/chat", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.NotNil(t, capturedResult)
	require.NotNil(t, capturedResult.MachineWarning)
	assert.Equal(t, "upgrade_required", capturedResult.MachineWarning.Status)
	assert.Contains(t, capturedResult.MachineWarning.Message, "trial machine without a billing plan")
	_ = rr
}

func TestCreditGateMiddleware_MachineStatusError(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	orgID := uuid.New()
	deps := CreditGateDeps{
		GetWalletBalance: func(ctx context.Context, id uuid.UUID) (int, error) {
			return 100, nil
		},
		GetMachineStatus: func(ctx context.Context, id uuid.UUID) (*MachineStatusResult, error) {
			return nil, errors.New("machine service down")
		},
	}

	var capturedResult *CreditCheckResult
	handler := CreditGateMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedResult = GetCreditCheckResult(r.Context())
		}),
		deps,
		testLogger(),
	)

	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, orgID.String())
	req := httptest.NewRequest("POST", "/api/v1/agent/chat", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, capturedResult)
	assert.True(t, capturedResult.Allowed)
	assert.Nil(t, capturedResult.MachineWarning)
}

func TestCreditGateMiddleware_NoMachineStatusFunc(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	orgID := uuid.New()
	deps := CreditGateDeps{
		GetWalletBalance: func(ctx context.Context, id uuid.UUID) (int, error) {
			return 100, nil
		},
		GetMachineStatus: nil,
	}

	var capturedResult *CreditCheckResult
	handler := CreditGateMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedResult = GetCreditCheckResult(r.Context())
		}),
		deps,
		testLogger(),
	)

	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, orgID.String())
	req := httptest.NewRequest("POST", "/api/v1/agent/chat", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, capturedResult)
	assert.Nil(t, capturedResult.MachineWarning)
}

func TestCreditGateMiddleware_MachineNoMachine(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	orgID := uuid.New()
	deps := CreditGateDeps{
		GetWalletBalance: func(ctx context.Context, id uuid.UUID) (int, error) {
			return 100, nil
		},
		GetMachineStatus: func(ctx context.Context, id uuid.UUID) (*MachineStatusResult, error) {
			return &MachineStatusResult{HasMachine: false}, nil
		},
	}

	var capturedResult *CreditCheckResult
	handler := CreditGateMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedResult = GetCreditCheckResult(r.Context())
		}),
		deps,
		testLogger(),
	)

	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, orgID.String())
	req := httptest.NewRequest("POST", "/api/v1/agent/chat", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.NotNil(t, capturedResult)
	assert.Nil(t, capturedResult.MachineWarning)
	_ = rr
}

func TestCreditGateMiddleware_MachineUnknownStatus(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	orgID := uuid.New()
	deps := CreditGateDeps{
		GetWalletBalance: func(ctx context.Context, id uuid.UUID) (int, error) {
			return 100, nil
		},
		GetMachineStatus: func(ctx context.Context, id uuid.UUID) (*MachineStatusResult, error) {
			return &MachineStatusResult{HasMachine: true, Status: "active"}, nil
		},
	}

	var capturedResult *CreditCheckResult
	handler := CreditGateMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedResult = GetCreditCheckResult(r.Context())
		}),
		deps,
		testLogger(),
	)

	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, orgID.String())
	req := httptest.NewRequest("POST", "/api/v1/agent/chat", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.NotNil(t, capturedResult)
	assert.Nil(t, capturedResult.MachineWarning)
	_ = rr
}

func TestCreditGateMiddleware_EmptyOrgIDPassesThrough(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	called := false
	handler := CreditGateMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }),
		CreditGateDeps{},
		testLogger(),
	)

	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, "")
	req := httptest.NewRequest("POST", "/api/v1/agent/chat", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.True(t, called)
}

func TestGetCreditCheckResult_NilContext(t *testing.T) {
	result := GetCreditCheckResult(context.Background())
	assert.Nil(t, result)
}

func TestBuildMachineWarning_NilInput(t *testing.T) {
	assert.Nil(t, buildMachineWarning(nil))
}

func TestBuildMachineWarning_NoMachine(t *testing.T) {
	assert.Nil(t, buildMachineWarning(&MachineStatusResult{HasMachine: false}))
}

func TestFormatCents(t *testing.T) {
	tests := []struct {
		cents  int
		expect string
	}{
		{999, "$9.99"},
		{100, "$1.00"},
		{1050, "$10.50"},
		{5, "$0.05"},
		{0, "$0.00"},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expect, formatCents(tc.cents))
	}
}

func TestItoa(t *testing.T) {
	assert.Equal(t, "0", itoa(0))
	assert.Equal(t, "1", itoa(1))
	assert.Equal(t, "42", itoa(42))
	assert.Equal(t, "100", itoa(100))
	assert.Equal(t, "-5", itoa(-5))
}

func TestCreditGateMiddleware_GracePeriodNilDays(t *testing.T) {
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = false
	defer func() { config.AppConfig.App.SelfHosted = orig }()

	orgID := uuid.New()
	deps := CreditGateDeps{
		GetWalletBalance: func(ctx context.Context, id uuid.UUID) (int, error) {
			return 100, nil
		},
		GetMachineStatus: func(ctx context.Context, id uuid.UUID) (*MachineStatusResult, error) {
			return &MachineStatusResult{
				HasMachine:    true,
				Status:        "grace_period",
				DaysRemaining: nil,
				PlanCostCents: 500,
			}, nil
		},
	}

	var capturedResult *CreditCheckResult
	handler := CreditGateMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedResult = GetCreditCheckResult(r.Context())
		}),
		deps,
		testLogger(),
	)

	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, orgID.String())
	req := httptest.NewRequest("POST", "/api/v1/agent/chat", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.NotNil(t, capturedResult)
	require.NotNil(t, capturedResult.MachineWarning)
	assert.Contains(t, capturedResult.MachineWarning.Message, "0 days")
	assert.Contains(t, capturedResult.MachineWarning.Message, "$5.00")
	_ = rr
}

func testLogger() logger.Logger {
	return logger.NewLogger()
}
