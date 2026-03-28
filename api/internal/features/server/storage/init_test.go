package storage_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/server/storage"
	server_types "github.com/nixopus/nixopus/api/internal/features/server/types"
	"github.com/nixopus/nixopus/api/internal/testutils"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func insertKey(setup *testutils.TestSetup, key *types.SSHKey) error {
	_, err := setup.DB.NewInsert().Model(key).Exec(setup.Ctx)
	return err
}

func TestSetDefaultServer_HappyPath(t *testing.T) {
	setup := testutils.NewTestSetup()
	serverStorage := &storage.ServerStorage{DB: setup.DB, Ctx: setup.Ctx}

	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)
	require.NotNil(t, org)

	keyA := &types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "key-a",
		AuthMethod:     "key",
		IsActive:       true,
		IsDefault:      true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	keyB := &types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "key-b",
		AuthMethod:     "key",
		IsActive:       true,
		IsDefault:      false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, insertKey(setup, keyA))
	require.NoError(t, insertKey(setup, keyB))

	oldDefaultID, err := serverStorage.SetDefaultServer(org.ID, keyB.ID)
	require.NoError(t, err)
	require.NotNil(t, oldDefaultID)
	assert.Equal(t, keyA.ID, *oldDefaultID)

	var updatedA types.SSHKey
	require.NoError(t, setup.DB.NewSelect().Model(&updatedA).Where("id = ?", keyA.ID).Scan(setup.Ctx))
	assert.False(t, updatedA.IsDefault)

	var updatedB types.SSHKey
	require.NoError(t, setup.DB.NewSelect().Model(&updatedB).Where("id = ?", keyB.ID).Scan(setup.Ctx))
	assert.True(t, updatedB.IsDefault)
}

func TestSetDefaultServer_TargetNotFound(t *testing.T) {
	setup := testutils.NewTestSetup()
	serverStorage := &storage.ServerStorage{DB: setup.DB, Ctx: setup.Ctx}

	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)
	require.NotNil(t, org)

	nonExistentID := uuid.New()
	_, err = serverStorage.SetDefaultServer(org.ID, nonExistentID)
	assert.ErrorIs(t, err, server_types.ErrServerNotFound)
}

func TestSetDefaultServer_TargetInactive(t *testing.T) {
	setup := testutils.NewTestSetup()
	serverStorage := &storage.ServerStorage{DB: setup.DB, Ctx: setup.Ctx}

	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)
	require.NotNil(t, org)

	inactiveKey := &types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "inactive-key",
		AuthMethod:     "key",
		IsActive:       false,
		IsDefault:      false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, insertKey(setup, inactiveKey))

	_, err = serverStorage.SetDefaultServer(org.ID, inactiveKey.ID)
	assert.ErrorIs(t, err, server_types.ErrServerInactive)
}

func TestSetDefaultServer_Idempotent(t *testing.T) {
	setup := testutils.NewTestSetup()
	serverStorage := &storage.ServerStorage{DB: setup.DB, Ctx: setup.Ctx}

	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)
	require.NotNil(t, org)

	activeDefault := &types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "already-default",
		AuthMethod:     "key",
		IsActive:       true,
		IsDefault:      true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, insertKey(setup, activeDefault))

	oldDefaultID, err := serverStorage.SetDefaultServer(org.ID, activeDefault.ID)
	require.NoError(t, err)
	require.NotNil(t, oldDefaultID)
	assert.Equal(t, activeDefault.ID, *oldDefaultID)

	var updated types.SSHKey
	require.NoError(t, setup.DB.NewSelect().Model(&updated).Where("id = ?", activeDefault.ID).Scan(setup.Ctx))
	assert.True(t, updated.IsDefault)
}
