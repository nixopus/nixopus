package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	flagstorage "github.com/nixopus/nixopus/api/internal/features/feature-flags/storage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	_ "modernc.org/sqlite"
)

// mockFeatureRepo implements flagstorage.FeatureFlagRepository for unit tests.
type mockFeatureRepo struct {
	getFlags   func(uuid.UUID) ([]types.FeatureFlag, error)
	updateFlag func(uuid.UUID, string, bool) error
	isEnabled  func(uuid.UUID, string) (bool, error)

	beginTxErr error
	beginTxFn  func() (bun.Tx, error)
	withTxFn   func(bun.Tx) flagstorage.FeatureFlagRepository
}

func (m *mockFeatureRepo) GetFeatureFlags(org uuid.UUID) ([]types.FeatureFlag, error) {
	if m.getFlags != nil {
		return m.getFlags(org)
	}
	return nil, nil
}

func (m *mockFeatureRepo) UpdateFeatureFlag(org uuid.UUID, featureName string, isEnabled bool) error {
	if m.updateFlag != nil {
		return m.updateFlag(org, featureName, isEnabled)
	}
	return errors.New("mockFeatureRepo: UpdateFeatureFlag not configured")
}

func (m *mockFeatureRepo) IsFeatureEnabled(org uuid.UUID, featureName string) (bool, error) {
	if m.isEnabled != nil {
		return m.isEnabled(org, featureName)
	}
	return false, errors.New("mockFeatureRepo: IsFeatureEnabled not configured")
}

func (m *mockFeatureRepo) BeginTx() (bun.Tx, error) {
	if m.beginTxErr != nil {
		return bun.Tx{}, m.beginTxErr
	}
	if m.beginTxFn != nil {
		return m.beginTxFn()
	}
	return bun.Tx{}, errors.New("mockFeatureRepo: BeginTx not configured")
}

func (m *mockFeatureRepo) WithTx(tx bun.Tx) flagstorage.FeatureFlagRepository {
	if m.withTxFn != nil {
		return m.withTxFn(tx)
	}
	return &mockTxRepo{}
}

// mockTxRepo is returned from WithTx; only UpdateFeatureFlag is used when seeding defaults.
type mockTxRepo struct {
	updateFlag func(uuid.UUID, string, bool) error
}

func (m *mockTxRepo) GetFeatureFlags(uuid.UUID) ([]types.FeatureFlag, error) {
	return nil, errors.New("unexpected GetFeatureFlags on tx repo")
}

func (m *mockTxRepo) UpdateFeatureFlag(org uuid.UUID, featureName string, isEnabled bool) error {
	if m.updateFlag != nil {
		return m.updateFlag(org, featureName, isEnabled)
	}
	return nil
}

func (m *mockTxRepo) IsFeatureEnabled(uuid.UUID, string) (bool, error) {
	return false, errors.New("unexpected IsFeatureEnabled on tx repo")
}

func (m *mockTxRepo) BeginTx() (bun.Tx, error) {
	return bun.Tx{}, errors.New("unexpected BeginTx on tx repo")
}

func (m *mockTxRepo) WithTx(bun.Tx) flagstorage.FeatureFlagRepository {
	return m
}

func TestFeatureFlagService_UpdateFeatureFlag_delegatesToStorage(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New()
	var seen types.UpdateFeatureFlagRequest
	repo := &mockFeatureRepo{
		updateFlag: func(org uuid.UUID, featureName string, isEnabled bool) error {
			assert.Equal(t, orgID, org)
			seen = types.UpdateFeatureFlagRequest{FeatureName: featureName, IsEnabled: isEnabled}
			return errors.New("storage error")
		},
	}
	svc := NewFeatureFlagService(repo, logger.NewLogger(), ctx)
	req := types.UpdateFeatureFlagRequest{FeatureName: "domain", IsEnabled: false}
	err := svc.UpdateFeatureFlag(orgID, req)
	require.Error(t, err)
	assert.Equal(t, req, seen)
}

func TestFeatureFlagService_IsFeatureEnabled_delegatesToStorage(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New()
	repo := &mockFeatureRepo{
		isEnabled: func(org uuid.UUID, featureName string) (bool, error) {
			assert.Equal(t, orgID, org)
			assert.Equal(t, "terminal", featureName)
			return true, nil
		},
	}
	svc := NewFeatureFlagService(repo, logger.NewLogger(), ctx)
	ok, err := svc.IsFeatureEnabled(orgID, "terminal")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestFeatureFlagService_GetFeatureFlags_storageError(t *testing.T) {
	ctx := context.Background()
	repo := &mockFeatureRepo{
		getFlags: func(uuid.UUID) ([]types.FeatureFlag, error) {
			return nil, errors.New("db read failed")
		},
	}
	svc := NewFeatureFlagService(repo, logger.NewLogger(), ctx)
	_, err := svc.GetFeatureFlags(uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get feature flags")
}

func TestFeatureFlagService_GetFeatureFlags_beginTxError(t *testing.T) {
	ctx := context.Background()
	repo := &mockFeatureRepo{
		getFlags: func(uuid.UUID) ([]types.FeatureFlag, error) {
			return nil, nil
		},
		beginTxErr: errors.New("tx unavailable"),
	}
	svc := NewFeatureFlagService(repo, logger.NewLogger(), ctx)
	_, err := svc.GetFeatureFlags(uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start transaction")
}

func TestFeatureFlagService_GetFeatureFlags_defaultSeedUpdateError(t *testing.T) {
	ctx := context.Background()
	sqldb, err := sql.Open("sqlite", "file:seederr?mode=memory")
	require.NoError(t, err)
	bunDB := bun.NewDB(sqldb, sqlitedialect.New())
	t.Cleanup(func() { _ = bunDB.Close() })

	tx, err := bunDB.BeginTx(ctx, nil)
	require.NoError(t, err)

	calls := 0
	repo := &mockFeatureRepo{
		getFlags: func(uuid.UUID) ([]types.FeatureFlag, error) {
			return nil, nil
		},
		beginTxFn: func() (bun.Tx, error) {
			return tx, nil
		},
		withTxFn: func(bun.Tx) flagstorage.FeatureFlagRepository {
			return &mockTxRepo{
				updateFlag: func(uuid.UUID, string, bool) error {
					calls++
					if calls == 2 {
						return errors.New("insert failed")
					}
					return nil
				},
			}
		},
	}

	svc := NewFeatureFlagService(repo, logger.NewLogger(), ctx)
	_, err = svc.GetFeatureFlags(uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create default feature flag")
}

func TestFeatureFlagService_GetFeatureFlags_seedsDefaultsWhenEmpty(t *testing.T) {
	ctx := context.Background()
	sqldb, err := sql.Open("sqlite", "file:seedok?mode=memory")
	require.NoError(t, err)
	bunDB := bun.NewDB(sqldb, sqlitedialect.New())
	t.Cleanup(func() { _ = bunDB.Close() })

	tx, err := bunDB.BeginTx(ctx, nil)
	require.NoError(t, err)

	orgID := uuid.New()
	repo := &mockFeatureRepo{
		getFlags: func(uuid.UUID) ([]types.FeatureFlag, error) {
			return nil, nil
		},
		beginTxFn: func() (bun.Tx, error) {
			return tx, nil
		},
		withTxFn: func(bun.Tx) flagstorage.FeatureFlagRepository {
			return &mockTxRepo{updateFlag: func(uuid.UUID, string, bool) error { return nil }}
		},
	}

	svc := NewFeatureFlagService(repo, logger.NewLogger(), ctx)
	flags, err := svc.GetFeatureFlags(orgID)
	require.NoError(t, err)
	require.Len(t, flags, 9)
	names := make(map[string]bool)
	for _, f := range flags {
		names[f.FeatureName] = true
		assert.Equal(t, orgID, f.OrganizationID)
		assert.True(t, f.IsEnabled)
	}
	for _, feature := range []types.FeatureName{
		types.FeatureDomain,
		types.FeatureTerminal,
		types.FeatureNotifications,
		types.FeatureSelfHosted,
		types.FeatureAudit,
		types.FeatureGithubConnector,
		types.FeatureMonitoring,
		types.FeatureContainer,
		types.FeatureMCP,
	} {
		assert.True(t, names[string(feature)], "missing %s", feature)
	}
}

func TestFeatureFlagService_GetFeatureFlags_commitError(t *testing.T) {
	ctx := context.Background()
	sqldb, err := sql.Open("sqlite", "file:commiterr?mode=memory")
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqldb.Close() })
	bunDB := bun.NewDB(sqldb, sqlitedialect.New())
	t.Cleanup(func() { _ = bunDB.Close() })

	txCtx, cancelTx := context.WithCancel(ctx)
	tx, err := bunDB.BeginTx(txCtx, nil)
	require.NoError(t, err)

	defaultCount := 9
	calls := 0
	repo := &mockFeatureRepo{
		getFlags: func(uuid.UUID) ([]types.FeatureFlag, error) {
			return nil, nil
		},
		beginTxFn: func() (bun.Tx, error) {
			return tx, nil
		},
		withTxFn: func(bun.Tx) flagstorage.FeatureFlagRepository {
			return &mockTxRepo{
				updateFlag: func(uuid.UUID, string, bool) error {
					calls++
					if calls == defaultCount {
						cancelTx()
					}
					return nil
				},
			}
		},
	}

	svc := NewFeatureFlagService(repo, logger.NewLogger(), ctx)
	_, err = svc.GetFeatureFlags(uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to commit transaction")
}

func TestFeatureFlagService_GetFeatureFlags_returnsExistingRows(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New()
	existing := []types.FeatureFlag{
		{OrganizationID: orgID, FeatureName: "domain", IsEnabled: true},
	}
	repo := &mockFeatureRepo{
		getFlags: func(id uuid.UUID) ([]types.FeatureFlag, error) {
			if id != orgID {
				return nil, fmt.Errorf("unexpected org")
			}
			return existing, nil
		},
	}
	svc := NewFeatureFlagService(repo, logger.NewLogger(), ctx)
	flags, err := svc.GetFeatureFlags(orgID)
	require.NoError(t, err)
	require.Len(t, flags, 1)
	assert.Equal(t, "domain", flags[0].FeatureName)
}
