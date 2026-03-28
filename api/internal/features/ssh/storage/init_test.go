package storage_test

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/ssh/storage"
	"github.com/nixopus/nixopus/api/internal/testutils"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func insertSSHKey(setup *testutils.TestSetup, key *types.SSHKey) error {
	_, err := setup.DB.NewInsert().Model(key).Exec(setup.Ctx)
	return err
}

func TestGetDefaultSSHKeyByOrganizationID_ActiveDefault(t *testing.T) {
	setup := testutils.NewTestSetup()
	sshStorage := &storage.SSHKeyStorage{DB: setup.DB, Ctx: setup.Ctx}

	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)
	require.NotNil(t, org)

	key := &types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "default-active-key",
		AuthMethod:     "key",
		IsActive:       true,
		IsDefault:      true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, insertSSHKey(setup, key))

	result, err := sshStorage.GetDefaultSSHKeyByOrganizationID(org.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, key.ID, result.ID)
	assert.True(t, result.IsDefault)
	assert.True(t, result.IsActive)
}

func TestGetDefaultSSHKeyByOrganizationID_InactiveDefault(t *testing.T) {
	setup := testutils.NewTestSetup()
	sshStorage := &storage.SSHKeyStorage{DB: setup.DB, Ctx: setup.Ctx}

	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)
	require.NotNil(t, org)

	key := &types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "default-inactive-key",
		AuthMethod:     "key",
		IsActive:       false,
		IsDefault:      true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, insertSSHKey(setup, key))

	result, err := sshStorage.GetDefaultSSHKeyByOrganizationID(org.ID)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, sql.ErrNoRows), "expected sql.ErrNoRows, got: %v", err)
}

func TestGetDefaultSSHKeyByOrganizationID_NoDefault(t *testing.T) {
	setup := testutils.NewTestSetup()
	sshStorage := &storage.SSHKeyStorage{DB: setup.DB, Ctx: setup.Ctx}

	_, org, err := setup.CreateTestUserAndOrg()
	require.NoError(t, err)
	require.NotNil(t, org)

	key := &types.SSHKey{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "non-default-key",
		AuthMethod:     "key",
		IsActive:       true,
		IsDefault:      false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, insertSSHKey(setup, key))

	result, err := sshStorage.GetDefaultSSHKeyByOrganizationID(org.ID)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, sql.ErrNoRows), "expected sql.ErrNoRows, got: %v", err)
}
