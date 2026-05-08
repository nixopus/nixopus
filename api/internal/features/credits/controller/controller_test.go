package controller

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/config"
	credits_storage "github.com/nixopus/nixopus/api/internal/features/credits/storage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	_ "modernc.org/sqlite"
)

// ---------- mock repository ----------

type mockRepo struct {
	balance    int
	balanceErr error
	ledger     *credits_storage.WalletLedger
	ledgerErr  error
	history    *credits_storage.UsageHistory
	historyErr error
	logs       *credits_storage.UsageLogList
	logsErr    error
	// introspection
	lastLedgerLimit   int
	lastLedgerOffset  int
	lastLogsLimit     int
	lastLogsOffset    int
	lastHistoryPeriod string
	lastHistoryGroup  string
}

func (m *mockRepo) GetWalletBalance(_ uuid.UUID) (int, error) {
	return m.balance, m.balanceErr
}

func (m *mockRepo) GetWalletLedger(_ uuid.UUID, limit, offset int) (*credits_storage.WalletLedger, error) {
	m.lastLedgerLimit = limit
	m.lastLedgerOffset = offset
	return m.ledger, m.ledgerErr
}

func (m *mockRepo) GetUsageHistory(_ uuid.UUID, period, groupBy string) (*credits_storage.UsageHistory, error) {
	m.lastHistoryPeriod = period
	m.lastHistoryGroup = groupBy
	return m.history, m.historyErr
}

func (m *mockRepo) GetUsageLogs(_ uuid.UUID, limit, offset int) (*credits_storage.UsageLogList, error) {
	m.lastLogsLimit = limit
	m.lastLogsOffset = offset
	return m.logs, m.logsErr
}

// ---------- helpers ----------

func testLogger() logger.Logger { return logger.NewLogger() }

func ctxWithUser(req *http.Request, orgID uuid.UUID) *http.Request {
	user := &types.User{ID: uuid.New(), Email: "test@example.com", Name: "Tester"}
	ctx := context.WithValue(req.Context(), types.UserContextKey, user)
	ctx = context.WithValue(ctx, types.OrganizationIDKey, orgID.String())
	return req.WithContext(ctx)
}

func makeCtx(method, url string, orgID *uuid.UUID) fuego.ContextNoBody {
	req := httptest.NewRequest(method, url, nil)
	if orgID != nil {
		req.Header.Set("X-ORGANIZATION-ID", orgID.String())
		req = ctxWithUser(req, *orgID)
	}
	return fuego.NewNetHTTPContext[any, any](fuego.BaseRoute{}, httptest.NewRecorder(), req, fuego.ReadOptions)
}

func setSelfHosted(t *testing.T, v bool) {
	t.Helper()
	orig := config.AppConfig.App.SelfHosted
	config.AppConfig.App.SelfHosted = v
	t.Cleanup(func() { config.AppConfig.App.SelfHosted = orig })
}

// ---------- constructors ----------

func TestNewCreditsControllerWithRepository(t *testing.T) {
	c := NewCreditsControllerWithRepository(&mockRepo{}, testLogger())
	require.NotNil(t, c)
}

func TestNewCreditsController(t *testing.T) {
	sqldb, err := sql.Open("sqlite", "file:creditsctl?mode=memory&cache=shared")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	t.Cleanup(func() { _ = db.Close() })
	c := NewCreditsController(db, context.Background(), testLogger())
	require.NotNil(t, c)
}

// =====================================================================
// GetBalance
// =====================================================================

func TestGetBalance_SelfHosted(t *testing.T) {
	setSelfHosted(t, true)
	c := NewCreditsControllerWithRepository(&mockRepo{}, testLogger())
	resp, err := c.GetBalance(makeCtx("GET", "/", nil))
	require.NoError(t, err)
	assert.False(t, resp.CreditsEnabled)
	assert.Equal(t, 0, resp.BalanceUSDCents)
}

func TestGetBalance_Unauthorized(t *testing.T) {
	setSelfHosted(t, false)
	c := NewCreditsControllerWithRepository(&mockRepo{}, testLogger())
	req := httptest.NewRequest("GET", "/", nil)
	fctx := fuego.NewNetHTTPContext[any, any](fuego.BaseRoute{}, httptest.NewRecorder(), req, fuego.ReadOptions)
	_, err := c.GetBalance(fctx)
	var u fuego.UnauthorizedError
	require.ErrorAs(t, err, &u)
}

func TestGetBalance_NoOrgID(t *testing.T) {
	setSelfHosted(t, false)
	c := NewCreditsControllerWithRepository(&mockRepo{}, testLogger())
	user := &types.User{ID: uuid.New()}
	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), types.UserContextKey, user))
	fctx := fuego.NewNetHTTPContext[any, any](fuego.BaseRoute{}, httptest.NewRecorder(), req, fuego.ReadOptions)
	_, err := c.GetBalance(fctx)
	var b fuego.BadRequestError
	require.ErrorAs(t, err, &b)
}

func TestGetBalance_StorageError(t *testing.T) {
	setSelfHosted(t, false)
	repo := &mockRepo{balanceErr: errors.New("db error")}
	c := NewCreditsControllerWithRepository(repo, testLogger())
	orgID := uuid.New()
	_, err := c.GetBalance(makeCtx("GET", "/", &orgID))
	var h fuego.HTTPError
	require.ErrorAs(t, err, &h)
	assert.Equal(t, http.StatusInternalServerError, h.Status)
}

func TestGetBalance_Success_CentsLessThan10(t *testing.T) {
	setSelfHosted(t, false)
	// e.g. 105 cents → $1.05, but cents%100 = 5 < 10 → "1.05"
	repo := &mockRepo{balance: 105}
	c := NewCreditsControllerWithRepository(repo, testLogger())
	orgID := uuid.New()
	resp, err := c.GetBalance(makeCtx("GET", "/", &orgID))
	require.NoError(t, err)
	assert.True(t, resp.CreditsEnabled)
	assert.Equal(t, 105, resp.BalanceUSDCents)
	assert.Equal(t, "1.05", resp.BalanceUSD)
}

func TestGetBalance_Success_CentsGTE10(t *testing.T) {
	setSelfHosted(t, false)
	// e.g. 150 cents → $1.50, cents%100 = 50 >= 10 → "1.50"
	repo := &mockRepo{balance: 150}
	c := NewCreditsControllerWithRepository(repo, testLogger())
	orgID := uuid.New()
	resp, err := c.GetBalance(makeCtx("GET", "/", &orgID))
	require.NoError(t, err)
	assert.Equal(t, "1.50", resp.BalanceUSD)
}

func TestGetBalance_Success_ZeroBalance(t *testing.T) {
	setSelfHosted(t, false)
	repo := &mockRepo{balance: 0}
	c := NewCreditsControllerWithRepository(repo, testLogger())
	orgID := uuid.New()
	resp, err := c.GetBalance(makeCtx("GET", "/", &orgID))
	require.NoError(t, err)
	assert.Equal(t, 0, resp.BalanceUSDCents)
	assert.Equal(t, "0.00", resp.BalanceUSD)
}

// =====================================================================
// GetUsage
// =====================================================================

func TestGetUsage_SelfHosted(t *testing.T) {
	setSelfHosted(t, true)
	c := NewCreditsControllerWithRepository(&mockRepo{}, testLogger())
	resp, err := c.GetUsage(makeCtx("GET", "/", nil))
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Status)
	assert.NotNil(t, resp.Data)
}

func TestGetUsage_Unauthorized(t *testing.T) {
	setSelfHosted(t, false)
	c := NewCreditsControllerWithRepository(&mockRepo{}, testLogger())
	req := httptest.NewRequest("GET", "/", nil)
	fctx := fuego.NewNetHTTPContext[any, any](fuego.BaseRoute{}, httptest.NewRecorder(), req, fuego.ReadOptions)
	_, err := c.GetUsage(fctx)
	var u fuego.UnauthorizedError
	require.ErrorAs(t, err, &u)
}

func TestGetUsage_NoOrgID(t *testing.T) {
	setSelfHosted(t, false)
	c := NewCreditsControllerWithRepository(&mockRepo{}, testLogger())
	user := &types.User{ID: uuid.New()}
	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), types.UserContextKey, user))
	fctx := fuego.NewNetHTTPContext[any, any](fuego.BaseRoute{}, httptest.NewRecorder(), req, fuego.ReadOptions)
	_, err := c.GetUsage(fctx)
	var b fuego.BadRequestError
	require.ErrorAs(t, err, &b)
}

func TestGetUsage_StorageError(t *testing.T) {
	setSelfHosted(t, false)
	repo := &mockRepo{historyErr: errors.New("db error")}
	c := NewCreditsControllerWithRepository(repo, testLogger())
	orgID := uuid.New()
	_, err := c.GetUsage(makeCtx("GET", "/?period=7d", &orgID))
	var h fuego.HTTPError
	require.ErrorAs(t, err, &h)
	assert.Equal(t, http.StatusInternalServerError, h.Status)
}

func TestGetUsage_DefaultPeriodAndGroupBy(t *testing.T) {
	setSelfHosted(t, false)
	repo := &mockRepo{history: &credits_storage.UsageHistory{}}
	c := NewCreditsControllerWithRepository(repo, testLogger())
	orgID := uuid.New()
	_, err := c.GetUsage(makeCtx("GET", "/", &orgID))
	require.NoError(t, err)
	assert.Equal(t, "30d", repo.lastHistoryPeriod)
	assert.Equal(t, "day", repo.lastHistoryGroup)
}

func TestGetUsage_ValidPeriodAndGroupBy(t *testing.T) {
	setSelfHosted(t, false)
	repo := &mockRepo{history: &credits_storage.UsageHistory{}}
	c := NewCreditsControllerWithRepository(repo, testLogger())
	orgID := uuid.New()
	_, err := c.GetUsage(makeCtx("GET", "/?period=7d&groupBy=model", &orgID))
	require.NoError(t, err)
	assert.Equal(t, "7d", repo.lastHistoryPeriod)
	assert.Equal(t, "model", repo.lastHistoryGroup)
}

func TestGetUsage_90dPeriod_UserGroup(t *testing.T) {
	setSelfHosted(t, false)
	repo := &mockRepo{history: &credits_storage.UsageHistory{}}
	c := NewCreditsControllerWithRepository(repo, testLogger())
	orgID := uuid.New()
	_, err := c.GetUsage(makeCtx("GET", "/?period=90d&groupBy=user", &orgID))
	require.NoError(t, err)
	assert.Equal(t, "90d", repo.lastHistoryPeriod)
	assert.Equal(t, "user", repo.lastHistoryGroup)
}

func TestGetUsage_InvalidParams_Defaults(t *testing.T) {
	setSelfHosted(t, false)
	repo := &mockRepo{history: &credits_storage.UsageHistory{}}
	c := NewCreditsControllerWithRepository(repo, testLogger())
	orgID := uuid.New()
	_, err := c.GetUsage(makeCtx("GET", "/?period=invalid&groupBy=bad", &orgID))
	require.NoError(t, err)
	assert.Equal(t, "30d", repo.lastHistoryPeriod)
	assert.Equal(t, "day", repo.lastHistoryGroup)
}

func TestGetUsage_Success(t *testing.T) {
	setSelfHosted(t, false)
	history := &credits_storage.UsageHistory{
		TotalCostUSD: 1.5,
		TotalTokens:  10000,
	}
	repo := &mockRepo{history: history}
	c := NewCreditsControllerWithRepository(repo, testLogger())
	orgID := uuid.New()
	resp, err := c.GetUsage(makeCtx("GET", "/", &orgID))
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Status)
	assert.Equal(t, 1.5, resp.Data.TotalCostUSD)
}

// =====================================================================
// GetTransactions
// =====================================================================

func TestGetTransactions_SelfHosted(t *testing.T) {
	setSelfHosted(t, true)
	c := NewCreditsControllerWithRepository(&mockRepo{}, testLogger())
	resp, err := c.GetTransactions(makeCtx("GET", "/", nil))
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Status)
	assert.NotNil(t, resp.Data)
}

func TestGetTransactions_Unauthorized(t *testing.T) {
	setSelfHosted(t, false)
	c := NewCreditsControllerWithRepository(&mockRepo{}, testLogger())
	req := httptest.NewRequest("GET", "/", nil)
	fctx := fuego.NewNetHTTPContext[any, any](fuego.BaseRoute{}, httptest.NewRecorder(), req, fuego.ReadOptions)
	_, err := c.GetTransactions(fctx)
	var u fuego.UnauthorizedError
	require.ErrorAs(t, err, &u)
}

func TestGetTransactions_NoOrgID(t *testing.T) {
	setSelfHosted(t, false)
	c := NewCreditsControllerWithRepository(&mockRepo{}, testLogger())
	user := &types.User{ID: uuid.New()}
	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), types.UserContextKey, user))
	fctx := fuego.NewNetHTTPContext[any, any](fuego.BaseRoute{}, httptest.NewRecorder(), req, fuego.ReadOptions)
	_, err := c.GetTransactions(fctx)
	var b fuego.BadRequestError
	require.ErrorAs(t, err, &b)
}

func TestGetTransactions_StorageError(t *testing.T) {
	setSelfHosted(t, false)
	repo := &mockRepo{ledgerErr: errors.New("db error")}
	c := NewCreditsControllerWithRepository(repo, testLogger())
	orgID := uuid.New()
	_, err := c.GetTransactions(makeCtx("GET", "/", &orgID))
	var h fuego.HTTPError
	require.ErrorAs(t, err, &h)
	assert.Equal(t, http.StatusInternalServerError, h.Status)
}

func TestGetTransactions_DefaultLimitOffset(t *testing.T) {
	setSelfHosted(t, false)
	repo := &mockRepo{ledger: &credits_storage.WalletLedger{}}
	c := NewCreditsControllerWithRepository(repo, testLogger())
	orgID := uuid.New()
	_, err := c.GetTransactions(makeCtx("GET", "/", &orgID))
	require.NoError(t, err)
	assert.Equal(t, 0, repo.lastLedgerLimit)
	assert.Equal(t, 0, repo.lastLedgerOffset)
}

func TestGetTransactions_CustomLimitOffset(t *testing.T) {
	setSelfHosted(t, false)
	repo := &mockRepo{ledger: &credits_storage.WalletLedger{}}
	c := NewCreditsControllerWithRepository(repo, testLogger())
	orgID := uuid.New()
	_, err := c.GetTransactions(makeCtx("GET", "/?limit=50&offset=10", &orgID))
	require.NoError(t, err)
	assert.Equal(t, 50, repo.lastLedgerLimit)
	assert.Equal(t, 10, repo.lastLedgerOffset)
}

func TestGetTransactions_Success(t *testing.T) {
	setSelfHosted(t, false)
	reason := "ai_usage"
	ledger := &credits_storage.WalletLedger{
		Items:      []credits_storage.WalletLedgerItem{{ID: uuid.New().String(), AmountCents: 100, EntryType: "debit", Reason: &reason}},
		TotalCount: 1,
	}
	repo := &mockRepo{ledger: ledger}
	c := NewCreditsControllerWithRepository(repo, testLogger())
	orgID := uuid.New()
	resp, err := c.GetTransactions(makeCtx("GET", "/", &orgID))
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Status)
	assert.Equal(t, 1, resp.Data.TotalCount)
}

// =====================================================================
// GetUsageLogs
// =====================================================================

func TestGetUsageLogs_SelfHosted(t *testing.T) {
	setSelfHosted(t, true)
	c := NewCreditsControllerWithRepository(&mockRepo{}, testLogger())
	resp, err := c.GetUsageLogs(makeCtx("GET", "/", nil))
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Status)
	assert.NotNil(t, resp.Data)
}

func TestGetUsageLogs_Unauthorized(t *testing.T) {
	setSelfHosted(t, false)
	c := NewCreditsControllerWithRepository(&mockRepo{}, testLogger())
	req := httptest.NewRequest("GET", "/", nil)
	fctx := fuego.NewNetHTTPContext[any, any](fuego.BaseRoute{}, httptest.NewRecorder(), req, fuego.ReadOptions)
	_, err := c.GetUsageLogs(fctx)
	var u fuego.UnauthorizedError
	require.ErrorAs(t, err, &u)
}

func TestGetUsageLogs_NoOrgID(t *testing.T) {
	setSelfHosted(t, false)
	c := NewCreditsControllerWithRepository(&mockRepo{}, testLogger())
	user := &types.User{ID: uuid.New()}
	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), types.UserContextKey, user))
	fctx := fuego.NewNetHTTPContext[any, any](fuego.BaseRoute{}, httptest.NewRecorder(), req, fuego.ReadOptions)
	_, err := c.GetUsageLogs(fctx)
	var b fuego.BadRequestError
	require.ErrorAs(t, err, &b)
}

func TestGetUsageLogs_StorageError(t *testing.T) {
	setSelfHosted(t, false)
	repo := &mockRepo{logsErr: errors.New("db error")}
	c := NewCreditsControllerWithRepository(repo, testLogger())
	orgID := uuid.New()
	_, err := c.GetUsageLogs(makeCtx("GET", "/", &orgID))
	var h fuego.HTTPError
	require.ErrorAs(t, err, &h)
	assert.Equal(t, http.StatusInternalServerError, h.Status)
}

func TestGetUsageLogs_DefaultLimitOffset(t *testing.T) {
	setSelfHosted(t, false)
	repo := &mockRepo{logs: &credits_storage.UsageLogList{}}
	c := NewCreditsControllerWithRepository(repo, testLogger())
	orgID := uuid.New()
	_, err := c.GetUsageLogs(makeCtx("GET", "/", &orgID))
	require.NoError(t, err)
	assert.Equal(t, 0, repo.lastLogsLimit)
	assert.Equal(t, 0, repo.lastLogsOffset)
}

func TestGetUsageLogs_CustomLimitOffset(t *testing.T) {
	setSelfHosted(t, false)
	repo := &mockRepo{logs: &credits_storage.UsageLogList{}}
	c := NewCreditsControllerWithRepository(repo, testLogger())
	orgID := uuid.New()
	_, err := c.GetUsageLogs(makeCtx("GET", "/?limit=30&offset=5", &orgID))
	require.NoError(t, err)
	assert.Equal(t, 30, repo.lastLogsLimit)
	assert.Equal(t, 5, repo.lastLogsOffset)
}

func TestGetUsageLogs_Success(t *testing.T) {
	setSelfHosted(t, false)
	latency := 250
	logList := &credits_storage.UsageLogList{
		Items: []credits_storage.UsageLogItem{
			{ID: uuid.New().String(), ModelID: "gpt-4o", TotalTokens: 500, CostUSD: 0.01, LatencyMs: &latency},
		},
		TotalCount: 1,
	}
	repo := &mockRepo{logs: logList}
	c := NewCreditsControllerWithRepository(repo, testLogger())
	orgID := uuid.New()
	resp, err := c.GetUsageLogs(makeCtx("GET", "/", &orgID))
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Status)
	assert.Equal(t, 1, resp.Data.TotalCount)
	assert.Equal(t, "gpt-4o", resp.Data.Items[0].ModelID)
}
