package storage_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/machine/storage"
	"github.com/nixopus/nixopus/api/internal/testutils"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMachineIsActive_ActiveMachine(t *testing.T) {
	setup := testutils.NewTestSetup()
	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	keyID := uuid.New()
	key := &types.SSHKey{
		ID:             keyID,
		OrganizationID: org.ID,
		Name:           "active-machine",
		AuthMethod:     "key",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err = setup.DB.NewInsert().Model(key).Exec(setup.Ctx)
	require.NoError(t, err)

	regStorage := storage.NewRegistrationStorage(setup.DB, setup.Ctx)
	isActive, err := regStorage.GetMachineIsActive(keyID)
	require.NoError(t, err)
	assert.True(t, isActive, "machine seeded as active should report is_active=true")
}

func TestGetMachineIsActive_InactiveMachine(t *testing.T) {
	setup := testutils.NewTestSetup()
	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	keyID := uuid.New()
	key := &types.SSHKey{
		ID:             keyID,
		OrganizationID: org.ID,
		Name:           "inactive-machine",
		AuthMethod:     "key",
		IsActive:       false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err = setup.DB.NewInsert().Model(key).Exec(setup.Ctx)
	require.NoError(t, err)

	regStorage := storage.NewRegistrationStorage(setup.DB, setup.Ctx)
	isActive, err := regStorage.GetMachineIsActive(keyID)
	require.NoError(t, err)
	assert.False(t, isActive, "machine seeded as inactive should report is_active=false")
}

func TestGetMachineIsActive_NotFound(t *testing.T) {
	setup := testutils.NewTestSetup()
	_, _, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	regStorage := storage.NewRegistrationStorage(setup.DB, setup.Ctx)
	_, err = regStorage.GetMachineIsActive(uuid.New())
	assert.Error(t, err, "querying non-existent machine should return an error")
}

func TestSetMachineActive_SetFalse(t *testing.T) {
	setup := testutils.NewTestSetup()
	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	keyID := uuid.New()
	key := &types.SSHKey{
		ID:             keyID,
		OrganizationID: org.ID,
		Name:           "pause-me",
		AuthMethod:     "key",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err = setup.DB.NewInsert().Model(key).Exec(setup.Ctx)
	require.NoError(t, err)

	regStorage := storage.NewRegistrationStorage(setup.DB, setup.Ctx)
	err = regStorage.SetMachineActive(keyID, false)
	require.NoError(t, err)

	isActive, err := regStorage.GetMachineIsActive(keyID)
	require.NoError(t, err)
	assert.False(t, isActive, "machine should be inactive after SetMachineActive(false)")
}

func TestSetMachineActive_SetTrue(t *testing.T) {
	setup := testutils.NewTestSetup()
	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	keyID := uuid.New()
	key := &types.SSHKey{
		ID:             keyID,
		OrganizationID: org.ID,
		Name:           "resume-me",
		AuthMethod:     "key",
		IsActive:       false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err = setup.DB.NewInsert().Model(key).Exec(setup.Ctx)
	require.NoError(t, err)

	regStorage := storage.NewRegistrationStorage(setup.DB, setup.Ctx)
	err = regStorage.SetMachineActive(keyID, true)
	require.NoError(t, err)

	isActive, err := regStorage.GetMachineIsActive(keyID)
	require.NoError(t, err)
	assert.True(t, isActive, "machine should be active after SetMachineActive(true)")
}

func TestSetMachineActive_ToggleTwice(t *testing.T) {
	setup := testutils.NewTestSetup()
	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	keyID := uuid.New()
	key := &types.SSHKey{
		ID:             keyID,
		OrganizationID: org.ID,
		Name:           "toggle-me",
		AuthMethod:     "key",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err = setup.DB.NewInsert().Model(key).Exec(setup.Ctx)
	require.NoError(t, err)

	regStorage := storage.NewRegistrationStorage(setup.DB, setup.Ctx)

	require.NoError(t, regStorage.SetMachineActive(keyID, false))
	isActive, err := regStorage.GetMachineIsActive(keyID)
	require.NoError(t, err)
	assert.False(t, isActive)

	require.NoError(t, regStorage.SetMachineActive(keyID, true))
	isActive, err = regStorage.GetMachineIsActive(keyID)
	require.NoError(t, err)
	assert.True(t, isActive)
}

func TestInsertProvisionDetails_UserOwned(t *testing.T) {
	setup := testutils.NewTestSetup()
	user, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)

	keyID := uuid.New()
	key := &types.SSHKey{
		ID:             keyID,
		OrganizationID: org.ID,
		Name:           "byos-key",
		AuthMethod:     "key",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err = setup.DB.NewInsert().Model(key).Exec(setup.Ctx)
	require.NoError(t, err)

	regStorage := storage.NewRegistrationStorage(setup.DB, setup.Ctx)
	err = regStorage.InsertProvisionDetails(user.ID, org.ID, keyID, "user_owned", types.ProvisionStepCompleted)
	require.NoError(t, err)

	exists, err := setup.DB.NewSelect().
		TableExpr("user_provision_details").
		Where("ssh_key_id = ?", keyID).
		Where("type = 'user_owned'").
		Exists(setup.Ctx)
	require.NoError(t, err)
	assert.True(t, exists, "user_provision_details row for user_owned should exist")
}
