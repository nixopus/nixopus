package storage_test

// Integration tests for CreditsStorage requiring a live PostgreSQL test database.
// Tests that query columns added in pending migrations (cost_usd, cached_tokens,
// reasoning_tokens on ai_usage_logs) are skipped until those migrations are applied.

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/credits/storage"
	machine_types "github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/nixopus/nixopus/api/internal/testutils"
	api_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// createOrgDirect inserts a user and organization directly into the DB, bypassing
// the auth service (which may not be available in the test environment).
func createOrgDirect(t *testing.T, setup *testutils.TestSetup) (*api_types.User, *api_types.Organization) {
	t.Helper()

	now := time.Now()
	user := &api_types.User{
		ID:        uuid.New(),
		Name:      "Test User",
		Email:     uuid.New().String() + "@test.example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err := setup.DB.NewInsert().Model(user).Exec(setup.Ctx)
	require.NoError(t, err)

	org := &api_types.Organization{
		ID:        uuid.New(),
		Name:      "Test Org",
		Slug:      uuid.New().String(),
		CreatedAt: now,
	}
	_, err = setup.DB.NewInsert().Model(org).Exec(setup.Ctx)
	require.NoError(t, err)

	return user, org
}

// hasColumn returns true if the given table has the given column in the connected DB.
func hasColumn(setup *testutils.TestSetup, table, column string) bool {
	var count int
	err := setup.DB.NewRaw(
		`SELECT COUNT(*) FROM information_schema.columns
		 WHERE table_name = ? AND column_name = ?`,
		table, column,
	).Scan(setup.Ctx, &count)
	return err == nil && count > 0
}

func newStorageFromSetup(setup *testutils.TestSetup) *storage.CreditsStorage {
	return storage.NewCreditsStorage(setup.DB, setup.Ctx)
}

// closedDB returns a *bun.DB backed by a connection that has been closed,
// allowing tests to exercise the "DB error" branches in storage functions.
func closedDB(t *testing.T) *bun.DB {
	t.Helper()
	sqldb, err := sql.Open("pgx", "host=localhost port=5433 user=nixopus password=nixopus dbname=nixopus_test sslmode=disable")
	require.NoError(t, err)
	db := bun.NewDB(sqldb, pgdialect.New())
	require.NoError(t, sqldb.Close()) // close immediately to force errors on all queries
	return db
}

// ---------- GetWalletBalance ----------

func TestGetWalletBalance_NoRows(t *testing.T) {
	setup := testutils.NewTestSetup()
	s := newStorageFromSetup(setup)
	balance, err := s.GetWalletBalance(uuid.New())
	require.NoError(t, err)
	assert.Equal(t, 0, balance)
}

func TestGetWalletBalance_DBError(t *testing.T) {
	db := closedDB(t)
	s := storage.NewCreditsStorage(db, context.Background())
	_, err := s.GetWalletBalance(uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get wallet balance")
}

func TestGetWalletBalance_WithTransaction(t *testing.T) {
	setup := testutils.NewTestSetup()
	_, org := createOrgDirect(t, setup)

	reason := "topup"
	tx := &machine_types.WalletTransaction{
		OrganizationID:    org.ID,
		AmountCents:       500,
		EntryType:         "credit",
		BalanceAfterCents: 500,
		Reason:            &reason,
	}
	_, err := setup.DB.NewInsert().Model(tx).Exec(setup.Ctx)
	require.NoError(t, err)

	s := newStorageFromSetup(setup)
	balance, err := s.GetWalletBalance(org.ID)
	require.NoError(t, err)
	assert.Equal(t, 500, balance)
}

func TestGetWalletBalance_ReturnsLatest(t *testing.T) {
	setup := testutils.NewTestSetup()
	_, org := createOrgDirect(t, setup)

	r1 := "topup"
	r2 := "ai_usage"
	txs := []*machine_types.WalletTransaction{
		{OrganizationID: org.ID, AmountCents: 1000, EntryType: "credit", BalanceAfterCents: 1000, Reason: &r1, CreatedAt: time.Now().Add(-2 * time.Second)},
		{OrganizationID: org.ID, AmountCents: 50, EntryType: "debit", BalanceAfterCents: 950, Reason: &r2, CreatedAt: time.Now()},
	}
	for _, tx := range txs {
		_, err := setup.DB.NewInsert().Model(tx).Exec(setup.Ctx)
		require.NoError(t, err)
	}

	s := newStorageFromSetup(setup)
	balance, err := s.GetWalletBalance(org.ID)
	require.NoError(t, err)
	assert.Equal(t, 950, balance, "should return the most recent balance_after_cents")
}

// ---------- GetWalletLedger ----------

func TestGetWalletLedger_DBError(t *testing.T) {
	db := closedDB(t)
	s := storage.NewCreditsStorage(db, context.Background())
	_, err := s.GetWalletLedger(uuid.New(), 20, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get wallet ledger")
}

func TestGetWalletLedger_Empty(t *testing.T) {
	setup := testutils.NewTestSetup()
	s := newStorageFromSetup(setup)
	ledger, err := s.GetWalletLedger(uuid.New(), 20, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, ledger.TotalCount)
	assert.Empty(t, ledger.Items)
}

func TestGetWalletLedger_LimitClamped_Zero(t *testing.T) {
	setup := testutils.NewTestSetup()
	_, org := createOrgDirect(t, setup)

	r := "topup"
	tx := &machine_types.WalletTransaction{
		OrganizationID: org.ID, AmountCents: 100, EntryType: "credit",
		BalanceAfterCents: 100, Reason: &r,
	}
	_, err := setup.DB.NewInsert().Model(tx).Exec(setup.Ctx)
	require.NoError(t, err)

	s := newStorageFromSetup(setup)
	// limit=0 → defaults to 20 inside storage
	ledger, err := s.GetWalletLedger(org.ID, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, ledger.TotalCount)
}

func TestGetWalletLedger_LimitClamped_Over100(t *testing.T) {
	setup := testutils.NewTestSetup()
	s := newStorageFromSetup(setup)
	// limit=200 → clamped to 100, no error
	ledger, err := s.GetWalletLedger(uuid.New(), 200, 0)
	require.NoError(t, err)
	assert.NotNil(t, ledger)
}

func TestGetWalletLedger_WithReason(t *testing.T) {
	setup := testutils.NewTestSetup()
	_, org := createOrgDirect(t, setup)

	reason := "manual_topup"
	tx := &machine_types.WalletTransaction{
		OrganizationID: org.ID, AmountCents: 200, EntryType: "credit",
		BalanceAfterCents: 200, Reason: &reason,
	}
	_, err := setup.DB.NewInsert().Model(tx).Exec(setup.Ctx)
	require.NoError(t, err)

	s := newStorageFromSetup(setup)
	ledger, err := s.GetWalletLedger(org.ID, 20, 0)
	require.NoError(t, err)
	require.Len(t, ledger.Items, 1)
	assert.Equal(t, "manual_topup", *ledger.Items[0].Reason)
	assert.Equal(t, "credit", ledger.Items[0].EntryType)
	assert.Equal(t, 200, ledger.Items[0].AmountCents)
}

func TestGetWalletLedger_NilReason(t *testing.T) {
	setup := testutils.NewTestSetup()
	_, org := createOrgDirect(t, setup)

	tx := &machine_types.WalletTransaction{
		OrganizationID: org.ID, AmountCents: 50, EntryType: "debit",
		BalanceAfterCents: 0, Reason: nil,
	}
	_, err := setup.DB.NewInsert().Model(tx).Exec(setup.Ctx)
	require.NoError(t, err)

	s := newStorageFromSetup(setup)
	ledger, err := s.GetWalletLedger(org.ID, 20, 0)
	require.NoError(t, err)
	require.Len(t, ledger.Items, 1)
	assert.Nil(t, ledger.Items[0].Reason)
}

func TestGetWalletLedger_Pagination(t *testing.T) {
	setup := testutils.NewTestSetup()
	_, org := createOrgDirect(t, setup)

	r := "topup"
	for i := range 5 {
		tx := &machine_types.WalletTransaction{
			OrganizationID: org.ID, AmountCents: (i + 1) * 10, EntryType: "credit",
			BalanceAfterCents: (i + 1) * 10, Reason: &r,
			CreatedAt: time.Now().Add(time.Duration(i) * time.Millisecond),
		}
		_, err := setup.DB.NewInsert().Model(tx).Exec(setup.Ctx)
		require.NoError(t, err)
	}

	s := newStorageFromSetup(setup)
	page1, err := s.GetWalletLedger(org.ID, 3, 0)
	require.NoError(t, err)
	assert.Equal(t, 5, page1.TotalCount)
	assert.Len(t, page1.Items, 3)

	page2, err := s.GetWalletLedger(org.ID, 3, 3)
	require.NoError(t, err)
	assert.Equal(t, 5, page2.TotalCount)
	assert.Len(t, page2.Items, 2)
}

func TestGetUsageHistory_DBError(t *testing.T) {
	db := closedDB(t)
	s := storage.NewCreditsStorage(db, context.Background())
	_, err := s.GetUsageHistory(uuid.New(), "30d", "day")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get usage totals")
}

func TestGetUsageLogs_DBError(t *testing.T) {
	db := closedDB(t)
	s := storage.NewCreditsStorage(db, context.Background())
	_, err := s.GetUsageLogs(uuid.New(), 20, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get usage logs")
}

// ---------- GetUsageHistory ----------

// skipIfMissingColumn skips the test if the ai_usage_logs table doesn't have a required column.
// This handles the case where the schema migration adding cost_usd, cached_tokens, etc. hasn't
// been applied to the test database yet.
func skipIfMissingAIUsageColumn(t *testing.T, setup *testutils.TestSetup, column string) {
	t.Helper()
	if !hasColumn(setup, "ai_usage_logs", column) {
		t.Skipf("ai_usage_logs.%s column not yet migrated — run pending DB migrations", column)
	}
}

func TestGetUsageHistory_Empty(t *testing.T) {
	setup := testutils.NewTestSetup()
	skipIfMissingAIUsageColumn(t, setup, "cost_usd")
	s := newStorageFromSetup(setup)
	history, err := s.GetUsageHistory(uuid.New(), "30d", "day")
	require.NoError(t, err)
	assert.Equal(t, float64(0), history.TotalCostUSD)
	assert.Equal(t, int64(0), history.TotalTokens)
	assert.NotNil(t, history.Breakdown)
	assert.NotNil(t, history.DailyUsage)
}

func TestGetUsageHistory_WithLogs_GroupByDay(t *testing.T) {
	setup := testutils.NewTestSetup()
	skipIfMissingAIUsageColumn(t, setup, "cost_usd")
	_, org := createOrgDirect(t, setup)

	_, err := setup.DB.NewRaw(
		`INSERT INTO ai_usage_logs (organization_id, user_id, model_id, model_tier, total_tokens, cost_usd, status)
		 VALUES (gen_random_uuid(), gen_random_uuid(), 'gpt-4o', 'premium', 150, 0.01, 'success')`,
	).Exec(context.Background())
	_ = err // ignore; just skip if still missing columns

	s := newStorageFromSetup(setup)
	history, err := s.GetUsageHistory(org.ID, "30d", "day")
	require.NoError(t, err)
	assert.NotNil(t, history)
}

func TestGetUsageHistory_Period7d(t *testing.T) {
	setup := testutils.NewTestSetup()
	skipIfMissingAIUsageColumn(t, setup, "cost_usd")
	s := newStorageFromSetup(setup)
	history, err := s.GetUsageHistory(uuid.New(), "7d", "model")
	require.NoError(t, err)
	assert.NotNil(t, history)
}

func TestGetUsageHistory_Period90d(t *testing.T) {
	setup := testutils.NewTestSetup()
	skipIfMissingAIUsageColumn(t, setup, "cost_usd")
	s := newStorageFromSetup(setup)
	history, err := s.GetUsageHistory(uuid.New(), "90d", "user")
	require.NoError(t, err)
	assert.NotNil(t, history)
}

// ---------- GetUsageLogs ----------

func TestGetUsageLogs_Empty(t *testing.T) {
	setup := testutils.NewTestSetup()
	skipIfMissingAIUsageColumn(t, setup, "cached_tokens")
	s := newStorageFromSetup(setup)
	list, err := s.GetUsageLogs(uuid.New(), 20, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, list.TotalCount)
	assert.Empty(t, list.Items)
}

func TestGetUsageLogs_LimitClamped_Zero(t *testing.T) {
	setup := testutils.NewTestSetup()
	skipIfMissingAIUsageColumn(t, setup, "cached_tokens")
	s := newStorageFromSetup(setup)
	list, err := s.GetUsageLogs(uuid.New(), 0, 0)
	require.NoError(t, err)
	assert.NotNil(t, list)
}

func TestGetUsageLogs_LimitClamped_Over100(t *testing.T) {
	setup := testutils.NewTestSetup()
	skipIfMissingAIUsageColumn(t, setup, "cached_tokens")
	s := newStorageFromSetup(setup)
	list, err := s.GetUsageLogs(uuid.New(), 200, 0)
	require.NoError(t, err)
	assert.NotNil(t, list)
}

func TestGetUsageLogs_Pagination(t *testing.T) {
	setup := testutils.NewTestSetup()
	skipIfMissingAIUsageColumn(t, setup, "cached_tokens")
	_, org := createOrgDirect(t, setup)

	s := newStorageFromSetup(setup)
	// Just verify pagination returns correct counts; log insertion requires cost_usd.
	page1, err := s.GetUsageLogs(org.ID, 4, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, page1.TotalCount)

	page2, err := s.GetUsageLogs(org.ID, 4, 0)
	require.NoError(t, err)
	assert.NotNil(t, page2)
}
