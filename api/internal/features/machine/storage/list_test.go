package storage_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/machine/storage"
	machine_types "github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/nixopus/nixopus/api/internal/testutils"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func insertSSHKey(setup *testutils.TestSetup, key *types.SSHKey) error {
	_, err := setup.DB.NewInsert().Model(key).Exec(setup.Ctx)
	return err
}

func TestSetDefaultMachine_HappyPath(t *testing.T) {
	setup := testutils.NewTestSetup()
	listStorage := storage.NewListStorage(setup.DB, setup.Ctx)

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
	require.NoError(t, insertSSHKey(setup, keyA))
	require.NoError(t, insertSSHKey(setup, keyB))

	oldDefaultID, err := listStorage.SetDefaultMachine(org.ID, keyB.ID)
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

func TestSetDefaultMachine_TargetNotFound(t *testing.T) {
	setup := testutils.NewTestSetup()
	listStorage := storage.NewListStorage(setup.DB, setup.Ctx)

	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)
	require.NotNil(t, org)

	_, err = listStorage.SetDefaultMachine(org.ID, uuid.New())
	assert.ErrorIs(t, err, machine_types.ErrMachineNotFound)
}

func TestSetDefaultMachine_TargetInactive(t *testing.T) {
	setup := testutils.NewTestSetup()
	listStorage := storage.NewListStorage(setup.DB, setup.Ctx)

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
	require.NoError(t, insertSSHKey(setup, inactiveKey))

	_, err = listStorage.SetDefaultMachine(org.ID, inactiveKey.ID)
	assert.ErrorIs(t, err, machine_types.ErrMachineInactive)
}

func TestSetDefaultMachine_Idempotent(t *testing.T) {
	setup := testutils.NewTestSetup()
	listStorage := storage.NewListStorage(setup.DB, setup.Ctx)

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
	require.NoError(t, insertSSHKey(setup, activeDefault))

	oldDefaultID, err := listStorage.SetDefaultMachine(org.ID, activeDefault.ID)
	require.NoError(t, err)
	require.NotNil(t, oldDefaultID)
	assert.Equal(t, activeDefault.ID, *oldDefaultID)

	var updated types.SSHKey
	require.NoError(t, setup.DB.NewSelect().Model(&updated).Where("id = ?", activeDefault.ID).Scan(setup.Ctx))
	assert.True(t, updated.IsDefault)
}
